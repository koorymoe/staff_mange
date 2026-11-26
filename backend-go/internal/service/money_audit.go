package service

import (
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ 🔎 فحص الأرقام (دازاي 11-26) ═══
// «هل كل الفواتير بليدر؟ هل كل المبالغ حسبها النظام؟ الصافي صحيح؟ شكد إدخال الشركة؟
//  هل تقييم الموظفين صحيح؟ … لازم النتيجه صحيحه». الفترة من 26 لـ25.
// قراءة بس — ما يغيّر ولا رقم. كل فحص: ✅ أو ❌ ويا العدد وقائمة الي سببته.

type AuditItem struct {
	Ref    string  `db:"ref" json:"ref"`       // كود الحجز/الفاتورة أو اسم الموظف
	Label  string  `db:"label" json:"label"`   // شرح قصير
	Amount float64 `db:"amount" json:"amount"` // مبلغ إذا يخص
	Date   string  `db:"date" json:"date"`
}

type AuditCheck struct {
	Key   string      `json:"key"`
	Title string      `json:"title"`
	Why   string      `json:"why"` // ليش مهم
	OK    bool        `json:"ok"`
	Warn  bool        `json:"warn,omitempty"` // مو خطأ أكيد، بس يستاهل نظرة
	Count int         `json:"count"`
	Value string      `json:"value,omitempty"`
	Items []AuditItem `json:"items,omitempty"`
}

type AuditSection struct {
	Title  string       `json:"title"`
	Checks []AuditCheck `json:"checks"`
}

type AuditMoneyLine struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
	Note   string  `json:"note,omitempty"`
	Bold   bool    `json:"bold,omitempty"`
}

type MoneyAudit struct {
	Period   string           `json:"period"`
	From     string           `json:"from"`
	To       string           `json:"to"`
	Problems int              `json:"problems"`
	Money    []AuditMoneyLine `json:"money"`
	Sections []AuditSection   `json:"sections"`
}

const auditListLimit = 60

type MoneyAuditService struct {
	db     *sqlx.DB
	scores *MatrixScoreService
}

func NewMoneyAuditService(db *sqlx.DB, scores *MatrixScoreService) *MoneyAuditService {
	return &MoneyAuditService{db: db, scores: scores}
}

// auditPeriod حدود الفترة (26 → 25) بتوقيت بغداد؛ to حصري (أول يوم بعد النهاية).
func auditPeriod(label string, now time.Time) (string, time.Time, time.Time) {
	if _, err := time.Parse("2006-01", label); err != nil {
		label = repository.PeriodOf(now.In(debriefLoc), 26)
	}
	a, b := repository.PeriodRange(label, 26)
	from := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, debriefLoc)
	to := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, debriefLoc).AddDate(0, 0, 1)
	return label, from, to
}

// الفواتير المحسوبة: مو مسودة ومو ملغية.
const auditInv = `WITH inv AS (
	SELECT li.id, li."bookingId", li."employeeId", li."netTotal"::float8 AS n, li."createdAt", li."isFree", li."accountingCode"
	FROM "LeaderInvoice" li
	WHERE li."revokedAt" IS NULL AND li.status <> 'DRAFT' AND li."createdAt" >= $1 AND li."createdAt" < $2
)`

func (s *MoneyAuditService) list(q string, args ...any) []AuditItem {
	out := []AuditItem{}
	if err := s.db.Select(&out, q, args...); err != nil {
		return []AuditItem{{Ref: "⚠️", Label: "تعذر الفحص: " + err.Error()}}
	}
	return out
}

func check(key, title, why string, items []AuditItem) AuditCheck {
	c := AuditCheck{Key: key, Title: title, Why: why, Count: len(items), OK: len(items) == 0}
	if len(items) > auditListLimit {
		items = items[:auditListLimit]
	}
	c.Items = items
	return c
}

func iqd(v float64) string {
	n := int64(math.Round(v))
	s := fmt.Sprintf("%d", n)
	if n < 0 {
		s = s[1:]
	}
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	if n < 0 {
		out = "-" + out
	}
	return out + " د.ع"
}

func (s *MoneyAuditService) Run(label string) (*MoneyAudit, error) {
	period, from, to := auditPeriod(label, time.Now())
	a := &MoneyAudit{Period: period, From: from.Format("2006-01-02"), To: to.AddDate(0, 0, -1).Format("2006-01-02")}

	// ── ١) الفواتير والليدر ──
	inv := AuditSection{Title: "🧾 الفواتير والليدر"}
	inv.Checks = append(inv.Checks, check("no_leader", "فواتير بلا ليدر",
		"الفاتورة البلا ليدر حصة الليدر ما تروح لأحد وتنحسب للشركة، والفنيين ياخذون نسبة «بلا ليدر».",
		s.list(auditInv+`,`+auditCrew+`
			SELECT COALESCE(b.code, inv."accountingCode") AS ref, 'الكادر بلا ليدر' AS label, inv.n AS amount, to_char(inv."createdAt", 'YYYY-MM-DD') AS date
			FROM inv LEFT JOIN cnt ON cnt.inv = inv.id LEFT JOIN "Booking" b ON b.id = inv."bookingId"
			WHERE NOT COALESCE(cnt.has_lead, false) AND COALESCE(cnt.techs, 0) > 0 ORDER BY inv."createdAt"`, from, to)))
	inv.Checks = append(inv.Checks, check("no_crew", "فواتير بلا كادر أبد",
		"فلوسها ما تنحسب لأي فني بالإحصائيات (إلا إذا الي فوترها ليدر).",
		s.list(auditInv+`,`+auditCrew+`
			SELECT COALESCE(b.code, inv."accountingCode") AS ref, 'ماكو عمولات ولا تكليف' AS label, inv.n AS amount, to_char(inv."createdAt", 'YYYY-MM-DD') AS date
			FROM inv LEFT JOIN cnt ON cnt.inv = inv.id LEFT JOIN "Booking" b ON b.id = inv."bookingId"
			WHERE cnt.inv IS NULL ORDER BY inv."createdAt"`, from, to)))
	inv.Checks = append(inv.Checks, check("dup_invoice", "أكثر من فاتورة لنفس الحجز",
		"ممكن تنحسب الفلوس مرتين.",
		s.list(auditInv+`
			SELECT COALESCE(b.code, '?') AS ref, count(*)::text || ' فواتير' AS label, sum(inv.n) AS amount, to_char(min(inv."createdAt"), 'YYYY-MM-DD') AS date
			FROM inv LEFT JOIN "Booking" b ON b.id = inv."bookingId" WHERE inv."bookingId" IS NOT NULL
			GROUP BY b.code, inv."bookingId" HAVING count(*) > 1 ORDER BY 4`, from, to)))
	inv.Checks = append(inv.Checks, check("done_no_invoice", "حجوزات مكتملة بلا فاتورة",
		"شغل خلص بس فلوسه ما داخلة بأي حساب.",
		s.list(`SELECT b.code AS ref, COALESCE(s.name, 'بلا خدمة') || CASE WHEN b."bookingType"::text = 'INTERNAL' THEN ' (داخلي)' ELSE '' END AS label,
				COALESCE(b."amountCollected", 0) AS amount, to_char(b."completedAt", 'YYYY-MM-DD') AS date
			FROM "Booking" b LEFT JOIN "Service" s ON s.id = b."serviceId"
			WHERE b.status = 'COMPLETED' AND b."completedAt" >= $1 AND b."completedAt" < $2 AND b."bookingType"::text <> 'INTERNAL'
			  AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id AND li."revokedAt" IS NULL AND li.status <> 'DRAFT')
			ORDER BY b."completedAt"`, from, to)))
	inv.Checks = append(inv.Checks, check("drafts", "فواتير مسودة (ما منحسبة)",
		"المسودة ما تدخل بأي رقم لحد ما تنعتمد.",
		s.list(`SELECT COALESCE(b.code, li."accountingCode") AS ref, 'مسودة' AS label, li."netTotal"::float8 AS amount, to_char(li."createdAt", 'YYYY-MM-DD') AS date
			FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId"
			WHERE li.status = 'DRAFT' AND li."revokedAt" IS NULL AND li."createdAt" >= $1 AND li."createdAt" < $2 ORDER BY li."createdAt"`, from, to)))
	inv.Checks = append(inv.Checks, check("zero_invoice", "فواتير مبلغها صفر (مو مجانية)",
		"فاتورة بلا مبلغ وهي مو مأشّرة مجانية = نسيان إدخال.",
		s.list(auditInv+`
			SELECT COALESCE(b.code, inv."accountingCode") AS ref, 'المبلغ صفر' AS label, inv.n AS amount, to_char(inv."createdAt", 'YYYY-MM-DD') AS date
			FROM inv LEFT JOIN "Booking" b ON b.id = inv."bookingId" WHERE inv.n <= 0 AND NOT inv."isFree" ORDER BY inv."createdAt"`, from, to)))
	a.Sections = append(a.Sections, inv)

	// ── ٢) هل النظام حسب كل المبالغ؟ ──
	calc := AuditSection{Title: "🧮 هل المبالغ محسوبة صح؟"}
	calc.Checks = append(calc.Checks, check("collected_mismatch", "المستلم بالحجز يختلف عن مبلغ الفاتورة",
		"المبلغ الي انكتب «انستلم» بالحجز مو نفس صافي الفاتورة (فرق أكثر من 1,000 د.ع).",
		s.list(auditInv+`
			SELECT b.code AS ref, 'المستلم ' || to_char(b."amountCollected", 'FM999,999,999,990') || ' · الفاتورة ' || to_char(inv.n, 'FM999,999,999,990') AS label,
				(b."amountCollected" - inv.n) AS amount, to_char(inv."createdAt", 'YYYY-MM-DD') AS date
			FROM inv JOIN "Booking" b ON b.id = inv."bookingId"
			WHERE b."amountCollected" IS NOT NULL AND abs(b."amountCollected" - inv.n) > 1000 ORDER BY abs(b."amountCollected" - inv.n) DESC`, from, to)))
	calc.Checks = append(calc.Checks, check("not_collected", "فواتير بلا مبلغ مستلم بالحجز",
		"الفاتورة موجودة بس ما مكتوب بالحجز شكد انستلم فعلاً.",
		s.list(auditInv+`
			SELECT b.code AS ref, 'ما مسجّل مستلم' AS label, inv.n AS amount, to_char(inv."createdAt", 'YYYY-MM-DD') AS date
			FROM inv JOIN "Booking" b ON b.id = inv."bookingId"
			WHERE b."amountCollected" IS NULL AND inv.n > 0 ORDER BY inv."createdAt"`, from, to)))

	// مطابقة: مجموع حصص الإحصائيات = المتوقع من كل فاتورة حسب النسب
	ms := repository.LoadMoneySplit(s.db)
	var tot struct {
		Invoices int     `db:"invoices"`
		Total    float64 `db:"total"`
		Expected float64 `db:"expected"`
	}
	_ = s.db.Get(&tot, auditInv+`,`+auditCrew+`
		SELECT count(*)::int AS invoices, COALESCE(sum(inv.n), 0)::float8 AS total,
			COALESCE(sum(CASE WHEN cnt.inv IS NULL THEN 0
				ELSE inv.n * ((CASE WHEN cnt.has_lead THEN $3::float8 ELSE 0 END) + (CASE WHEN cnt.techs > 0 THEN (CASE WHEN cnt.has_lead THEN $4::float8 ELSE $5::float8 END) ELSE 0 END)) / 100.0 END), 0)::float8 AS expected
		FROM inv LEFT JOIN cnt ON cnt.inv = inv.id`, from, to, ms.StatsLeaderPct, ms.StatsTechPct, ms.StatsNoLeaderTechPct)
	rev, _ := s.scores.Repo().Revenue(from, to, "")
	var work, booked, proj float64
	for _, r := range rev {
		switch r.Kind {
		case "WORK":
			work += r.Amount
		case "BOOKED":
			booked += r.Amount
		case "PROJECT":
			proj += r.Amount
		}
	}
	var rework float64
	_ = s.db.Get(&rework, `SELECT COALESCE(sum(amount), 0)::float8 FROM "ReworkAdjustment" WHERE "createdAt" >= $1 AND "createdAt" < $2`, from, to)
	var leaderOnly float64 // فواتير بلا كادر والي فوترها ليدر: ياخذ نسبة الليدر
	_ = s.db.Get(&leaderOnly, auditInv+`,`+auditCrew+`
		SELECT COALESCE(sum(inv.n * $3::float8 / 100.0), 0)::float8 FROM inv JOIN "Employee" e ON e.id = inv."employeeId"
		WHERE e."isLeader" AND e.role::text <> 'SERVICE_MANAGER' AND NOT EXISTS (SELECT 1 FROM crew2 WHERE crew2.inv = inv.id)`, from, to, ms.StatsLeaderPct)
	diff := work - (tot.Expected + leaderOnly + rework)
	c := AuditCheck{Key: "split_match", Title: "حصص الموظفين بالإحصائيات = النسب × الفواتير",
		Why:   fmt.Sprintf("النسب الحالية: ليدر %.0f٪ · فنيين %.0f٪ (يتقاسمونها) · فني بلا ليدر %.0f٪.", ms.StatsLeaderPct, ms.StatsTechPct, ms.StatsNoLeaderTechPct),
		Value: fmt.Sprintf("المحسوب %s · المتوقع %s · الفرق %s", iqd(work), iqd(tot.Expected+leaderOnly+rework), iqd(diff))}
	c.OK = math.Abs(diff) < 1
	if !c.OK {
		c.Count = 1
	}
	calc.Checks = append(calc.Checks, c)
	calc.Checks = append(calc.Checks, AuditCheck{Key: "booked_match", Title: "«الحجوزات الي سجّلها» = مجموع الفواتير",
		Why: "كل فاتورة تنحسب مرة وحدة للي سجّل الحجز.", OK: math.Abs(booked-tot.Total) < 1 || booked <= tot.Total,
		Warn: booked < tot.Total-1, Value: fmt.Sprintf("المحسوب %s من %s (الفرق = فواتير حجوزها بلا منشئ)", iqd(booked), iqd(tot.Total))})
	a.Sections = append(a.Sections, calc)

	// ── ٣) إدخال الشركة والصافي ──
	var cp repository.CompanyProfit
	cp, _ = s.scores.Repo().CompanyProfit(from, to)
	var salaries, expenses float64
	var costPeople int
	_ = s.db.Get(&salaries, `SELECT COALESCE(sum(salary), 0)::float8 FROM "CostEmployee"`)
	_ = s.db.Get(&costPeople, `SELECT count(*)::int FROM "CostEmployee"`)
	_ = s.db.Get(&expenses, `SELECT COALESCE(sum(amount), 0)::float8 FROM "CostExpense" WHERE month = $1`, period)
	a.Money = []AuditMoneyLine{
		{Label: "فواتير الحجوزات (المعتمدة)", Amount: cp.Total, Note: fmt.Sprintf("%d فاتورة", cp.Invoices)},
		{Label: "دفعات المشاريع", Amount: proj},
		{Label: "إدخال الشركة الكلي", Amount: cp.Total + proj, Bold: true},
		{Label: fmt.Sprintf("− حصة الليادر (%.0f٪)", ms.RevenueLeaderPct), Amount: -cp.Leaders},
		{Label: fmt.Sprintf("− حصة الفنيين (%.0f٪)", ms.RevenueTechPct), Amount: -cp.Techs},
		{Label: "= حصة الشركة من الحجوزات", Amount: cp.Company, Bold: true},
		{Label: "− الرواتب (من التكاليف)", Amount: -salaries, Note: fmt.Sprintf("%d موظف مدخّل", costPeople)},
		{Label: "− مصاريف الشهر (من التكاليف)", Amount: -expenses},
		{Label: "= الصافي (حصة الشركة + المشاريع − الرواتب − المصاريف)", Amount: cp.Company + proj - salaries - expenses, Bold: true},
	}
	money := AuditSection{Title: "💰 الصافي"}
	money.Checks = append(money.Checks, AuditCheck{Key: "split_sum", Title: "الليادر + الفنيين + الشركة = مجموع الفواتير",
		Why: "إذا ما يتساوون اكو فلوس ضايعة بالحسبة.", OK: math.Abs(cp.Leaders+cp.Techs+cp.Company-cp.Total) < 1,
		Value: fmt.Sprintf("%s + %s + %s = %s", iqd(cp.Leaders), iqd(cp.Techs), iqd(cp.Company), iqd(cp.Total))})
	money.Checks = append(money.Checks, AuditCheck{Key: "company_negative", Title: "حصة الشركة مو بالسالب",
		Why: "إذا نسب الفنيين «لكل فني» والكادر كبير، ممكن حصتهم تعبر الفاتورة.", OK: cp.Company >= 0, Value: iqd(cp.Company)})
	money.Checks = append(money.Checks, AuditCheck{Key: "costs_entered", Title: "الرواتب والمصاريف مدخّلة",
		Why: "بدونها الصافي والأهداف يطلعن غلط.", OK: costPeople > 0 && expenses > 0, Warn: costPeople > 0 && expenses == 0,
		Value: fmt.Sprintf("%d موظف · مصاريف %s", costPeople, iqd(expenses))})
	a.Sections = append(a.Sections, money)

	// ── ٤) التقييم ──
	a.Sections = append(a.Sections, s.scoreChecks())

	for _, sec := range a.Sections {
		for _, c := range sec.Checks {
			if !c.OK {
				a.Problems++
			}
		}
	}
	return a, nil
}

// auditCrew نفس كادر الإحصائيات (workCrewSQL) — مكرر هنا حتى الفحص يقرا نفس الشي.
const auditCrew = `
	crew AS (
		(SELECT DISTINCT ON (ec."leaderInvoiceId", ec."employeeId") ec."leaderInvoiceId" AS inv, ec."employeeId", (ec.role = 'LEADER') AS lead
		FROM "EmployeeCommission" ec WHERE ec."leaderInvoiceId" IN (SELECT id FROM inv)
		ORDER BY ec."leaderInvoiceId", ec."employeeId", (ec.role = 'LEADER') DESC)
		UNION ALL
		SELECT inv.id, a."employeeId", (a."employeeId" = booking_leader_id(inv."bookingId"))
		FROM inv JOIN "BookingAssignment" a ON a."bookingId" = inv."bookingId"
		WHERE NOT EXISTS (SELECT 1 FROM "EmployeeCommission" ec WHERE ec."leaderInvoiceId" = inv.id)
	), crew2 AS (
		SELECT c.* FROM crew c JOIN "Employee" e ON e.id = c."employeeId" WHERE e.role::text <> 'SERVICE_MANAGER'
	), cnt AS (
		SELECT inv, bool_or(lead) AS has_lead, count(*) FILTER (WHERE NOT lead) AS techs FROM crew2 GROUP BY 1
	)`

func (s *MoneyAuditService) scoreChecks() AuditSection {
	sec := AuditSection{Title: "⭐ التقييم"}
	b, err := s.scores.Board("")
	if err != nil {
		sec.Checks = append(sec.Checks, AuditCheck{Key: "board", Title: "لوحة التقييم", Why: "تعذر حسابها: " + err.Error()})
		return sec
	}
	var noScore, noRole, out, noMax, bigDrop []AuditItem
	for _, p := range b.Staff {
		switch {
		case p.Final == nil:
			noScore = append(noScore, AuditItem{Ref: p.Name, Label: p.GroupLabel})
		case *p.Final < 0 || *p.Final > 100:
			out = append(out, AuditItem{Ref: p.Name, Label: fmt.Sprintf("النتيجة %.1f", *p.Final), Amount: *p.Final})
		}
		if p.Max == 0 && p.Final != nil {
			noMax = append(noMax, AuditItem{Ref: p.Name, Label: "ماكو ولا نقطة ماتركس"})
		}
		if !p.RoleWork && p.Final != nil {
			noRole = append(noRole, AuditItem{Ref: p.Name, Label: p.GroupLabel + " — نقاطه كلها دوام وتقارير"})
		}
		if p.Final != nil && p.PrevFinal != nil && *p.PrevFinal-*p.Final >= 25 {
			bigDrop = append(bigDrop, AuditItem{Ref: p.Name, Label: fmt.Sprintf("من %.0f إلى %.0f", *p.PrevFinal, *p.Final), Amount: *p.Final - *p.PrevFinal})
		}
	}
	sec.Checks = append(sec.Checks,
		AuditCheck{Key: "score_period", Title: "فترة التقييم", Warn: true, OK: true,
			Why:   "التقييم ينحسب بالشهر الميلادي (" + b.Month + ")، والتصفير يوم 27، والفلوس وساعات البيت من 26 لـ25 — ثلث فترات مختلفة.",
			Value: "يحتاج قرارك نوحدها لو لا"},
		check("no_score", "موظفين فعّالين بلا تقييم", "ماكو بيانات كافية تطلعله رقم.", noScore),
		check("no_role_work", "تقييمه ما بيه شغل دوره", "كل نقاطه من الدوام والتقارير — ماتركس ما يقرا شغل دوره الحقيقي.", noRole),
		check("no_matrix", "ماكو نقاط ماتركس أبد", "نتيجته من تقييم البشر بس.", noMax),
		check("out_of_range", "نتيجة برّا 0–100", "خطأ حسبة أكيد.", out),
		check("big_drop", "نزل 25 نقطة أو أكثر عن الشهر الي فات", "مو خطأ أكيد — بس يستاهل تشوف السبب.", bigDrop),
	)
	for i := range sec.Checks {
		if sec.Checks[i].Key == "big_drop" || sec.Checks[i].Key == "no_role_work" {
			sec.Checks[i].Warn = !sec.Checks[i].OK
			sec.Checks[i].OK = true
		}
	}
	return sec
}
