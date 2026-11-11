package service

import (
	"fmt"
	"math"
	"sort"
	"staffmange-api/internal/errlog"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ 📂 ملف الموظف الكامل لماتركس (دازاي 11-02 — تعديلات ٩.٣) ═══
// «ماتركس لازم يعرف كل موظف: شهادته ومعلوماته وملفه ومشاكله وشغله، وبيمن قوي
// وبيمن ضعيف، ومنو صارت عنده مشاكل، وسرعته — ويتابع: يتقدم لو يتراجع».
// يجمع من مصادر النظام نفسه (أرقام حقيقية بس). ⚠️ بلا راتب — الملف يروح لماتركس.

type DossierMonth struct {
	Month     string   `db:"month" json:"month"`
	MatrixPct *float64 `db:"matrixPct" json:"matrixPct"`
	HumanAvg  *float64 `db:"humanAvg" json:"humanAvg"` // من ٥
	Ratings   int      `db:"ratings" json:"ratings"`
}

type DossierRule struct {
	Rule  string  `db:"rule" json:"rule"`
	Title string  `db:"-" json:"title"`
	Pct   float64 `db:"pct" json:"pct"`
	N     int     `db:"n" json:"n"`
}

type Dossier struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Effective   string `json:"effectiveRole"`
	Position    string `json:"position"` // منصبه بالهيكل الجديد
	JobTitle    string `json:"jobTitle"`
	HireDate    string `json:"hireDate"`
	ContractEnd string `json:"contractEnd"`
	IsLeader    bool   `json:"isLeader"`

	Major         string `json:"major"`
	Degree        string `json:"degree"`
	Certificate   string `json:"certificate"`
	MonthlyHours  *int   `json:"monthlyHours"`
	CompanyRating *int   `json:"companyRating"`
	CompanyNote   string `json:"companyNote"`

	Skills    []string `json:"skills"`
	Training  []string `json:"training"`
	LeaderSkl []string `json:"leaderSkills"`

	Months    []DossierMonth `json:"months"` // آخر ٦ أشهر — الأقدم أول
	Trend     string         `json:"trend"`  // UP | DOWN | STEADY | UNKNOWN
	TrendNote string         `json:"trendNote"`
	Strong    []DossierRule  `json:"strong"`
	Weak      []DossierRule  `json:"weak"`

	SpeedRatio *float64 `json:"speedRatio"` // مدته ÷ المعتاد (١ = طبيعي، ١.٥ = أبطأ بنص)
	SpeedN     int      `json:"speedN"`

	Attendance30   int `json:"attendance30"`
	Reports30      int `json:"reports30"`
	ReportsLate30  int `json:"reportsLate30"`
	Negligence90   int `json:"negligence90"`
	Discipline90   int `json:"discipline90"` // مجموع النقاط المخصومة
	Justif90       int `json:"justifications90"`
	JustifRejected int `json:"justificationsRejected90"`
	Complaints90   int `json:"complaints90"`
	Issues90       int `json:"workplaceIssues90"`
	PeerNotes90    int `json:"peerNotesAbout90"`

	ReportMatchPct int `json:"reportMatchPct"` // ٩.٤ تقارير تطابق الشغل المسجّل آخر ٣٠ يوم
	ReportMatchN   int `json:"reportMatchN"`

	Flags []string `json:"flags"` // ملاحظات ماتركس الجاهزة
}

func round1(f float64) float64 { return math.Round(f*10) / 10 }

// BuildDossier يجمع الملف. who = معرّف الموظف.
func BuildDossier(db *sqlx.DB, id string) (*Dossier, error) {
	d := &Dossier{ID: id, Skills: []string{}, Training: []string{}, LeaderSkl: []string{}, Strong: []DossierRule{}, Weak: []DossierRule{}, Flags: []string{}}
	var e struct {
		Name     string     `db:"name"`
		Role     string     `db:"role"`
		Eff      string     `db:"eff"`
		Job      *string    `db:"jobTitle"`
		Pos      *string    `db:"position"`
		Cert     *string    `db:"certificate"`
		Hire     *time.Time `db:"hireDate"`
		Contract *time.Time `db:"contractEndDate"`
		Leader   bool       `db:"isLeader"`
	}
	if err := db.Get(&e, `SELECT name, role::text AS role, `+model.EffectiveRoleSQL(`id`, `role`)+` AS eff, "jobTitle", position, certificate,
			"hireDate", "contractEndDate", COALESCE("isLeader", false) AS "isLeader" FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil, fmt.Errorf("الموظف مو موجود")
	}
	d.Name, d.Role, d.Effective, d.IsLeader = e.Name, e.Role, e.Eff, e.Leader
	d.JobTitle = deref(e.Job, deref(e.Pos, ""))
	d.Certificate = deref(e.Cert, "")
	if e.Hire != nil {
		d.HireDate = e.Hire.Format("2006-01-02")
	}
	if e.Contract != nil {
		d.ContractEnd = e.Contract.Format("2006-01-02")
	}
	errlog.Err(db.Get(&d.Position, `SELECT COALESCE(string_agg(CASE oa.position WHEN 'DIRECTOR' THEN 'مدير ' WHEN 'DEPUTY' THEN 'نائب ' ELSE 'عضو ' END
			|| COALESCE(u.name, dv.name) || CASE WHEN oa.acting THEN ' (بالوكالة)' ELSE '' END, '، '), '')
		FROM "OrgAssignment" oa JOIN "OrgDivision" dv ON dv.id = oa."divisionId" LEFT JOIN "OrgUnit" u ON u.id = oa."unitId"
		WHERE oa."employeeId" = $1 AND oa."endedAt" IS NULL`, id))

	// الاستمارة (٩.١)
	var sv struct {
		Major, Degree, Note string
		Hours, Rating       *int
		Contract            *time.Time
	}
	if err := db.QueryRow(`SELECT major, degree, "companyNote", "monthlyHours", "companyRating", "contractEnd" FROM "EmployeeProfileSurvey" WHERE "employeeId" = $1`, id).
		Scan(&sv.Major, &sv.Degree, &sv.Note, &sv.Hours, &sv.Rating, &sv.Contract); err == nil {
		d.Major, d.Degree, d.CompanyNote, d.MonthlyHours, d.CompanyRating = sv.Major, sv.Degree, sv.Note, sv.Hours, sv.Rating
		if d.ContractEnd == "" && sv.Contract != nil {
			d.ContractEnd = sv.Contract.Format("2006-01-02")
		}
	}
	errlog.

		// المهارات والتدريب
		Err(db.Select(&d.Skills, `SELECT s.name FROM "EmployeeSkill" es JOIN "Skill" s ON s.id = es."skillId" WHERE es."employeeId" = $1 AND es."canPerform" ORDER BY s.name LIMIT 40`, id))
	errlog.Err(db.Select(&d.Training, `SELECT p.title || CASE WHEN tp.passed THEN ' ✓' ELSE ' (ما نجح بعد)' END FROM "TrainingProgramParticipant" tp
		JOIN "TrainingProgram" p ON p.id = tp."programId" WHERE tp."employeeId" = $1 ORDER BY tp."createdAt" DESC LIMIT 15`, id))
	errlog.Err(db.Select(&d.LeaderSkl, `SELECT DISTINCT ON (skill) skill || ': ' || score::text || '/5' FROM "LeaderSkillRating" WHERE "employeeId" = $1 ORDER BY skill, "ratedAt" DESC`, id))

	// المنحنى: آخر ٦ أشهر — ماتركس ٪ وتقييم البشر
	months := []DossierMonth{}
	if err := db.Select(&months, `
		WITH m AS (SELECT to_char(g, 'YYYY-MM') AS month, g AS f, g + interval '1 month' AS t
		           FROM generate_series(date_trunc('month', now() AT TIME ZONE 'Asia/Baghdad') - interval '5 months',
		                                date_trunc('month', now() AT TIME ZONE 'Asia/Baghdad'), interval '1 month') g)
		SELECT m.month,
		  (SELECT round(100.0 * sum(points) / NULLIF(sum("maxPoints"), 0), 1)::float8 FROM "MatrixScore"
		    WHERE "employeeId" = $1 AND "cancelledAt" IS NULL AND rule NOT IN ('ATTENDANCE','INVENTORY')
		      AND "at" >= m.f AT TIME ZONE 'Asia/Baghdad' AND "at" < m.t AT TIME ZONE 'Asia/Baghdad') AS "matrixPct",
		  (SELECT round(avg(score), 2)::float8 FROM (
		      SELECT score, "createdAt" FROM "StaffRating" WHERE "rateeId" = $1
		      UNION ALL SELECT score, "createdAt" FROM "CrewRating" WHERE "technicianId" = $1) h
		    WHERE h."createdAt" >= m.f AT TIME ZONE 'Asia/Baghdad' AND h."createdAt" < m.t AT TIME ZONE 'Asia/Baghdad') AS "humanAvg",
		  (SELECT count(*) FROM (
		      SELECT "createdAt" FROM "StaffRating" WHERE "rateeId" = $1
		      UNION ALL SELECT "createdAt" FROM "CrewRating" WHERE "technicianId" = $1) h
		    WHERE h."createdAt" >= m.f AT TIME ZONE 'Asia/Baghdad' AND h."createdAt" < m.t AT TIME ZONE 'Asia/Baghdad')::int AS ratings
		FROM m ORDER BY m.month`, id); err != nil {
		return nil, fmt.Errorf("المنحنى: %w", err)
	}
	d.Months = months
	d.Trend, d.TrendNote = trendOf(months)

	// قوي/ضعيف: نسبة كل قاعدة ماتركس (آخر ٦٠ يوم، ٣ مرات أو أكثر)
	rules := []DossierRule{}
	if err := db.Select(&rules, `SELECT rule, round(100.0 * sum(points) / NULLIF(sum("maxPoints"), 0), 1)::float8 AS pct, count(*)::int AS n
		FROM "MatrixScore" WHERE "employeeId" = $1 AND "cancelledAt" IS NULL AND "at" > now() - interval '60 days' AND rule NOT IN ('ATTENDANCE', 'INVENTORY')
		GROUP BY rule HAVING count(*) >= 3`, id); err != nil {
		return nil, fmt.Errorf("القواعد: %w", err)
	}
	for i := range rules {
		rules[i].Title = ruleTitle(rules[i].Rule)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Pct > rules[j].Pct })
	for _, r := range rules {
		if r.Pct >= 80 && len(d.Strong) < 3 {
			d.Strong = append(d.Strong, r)
		}
	}
	for i := len(rules) - 1; i >= 0 && len(d.Weak) < 3; i-- {
		if rules[i].Pct < 60 {
			d.Weak = append(d.Weak, rules[i])
		}
	}

	// السرعة: مدته على الحجز ÷ وسيط نفس الخدمة (آخر ١٨٠ يوم)، بلا الي فوگ ١٤ ساعة
	var sp struct {
		Ratio *float64
		N     int
	}
	_ = db.QueryRow(`
		WITH med AS (SELECT "serviceId", percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM ("completedAt" - "startedAt"))) AS m
		             FROM "Booking" WHERE "startedAt" IS NOT NULL AND "completedAt" > "startedAt" AND "completedAt" - "startedAt" < interval '14 hours'
		               AND "completedAt" > now() - interval '180 days' GROUP BY "serviceId" HAVING count(*) >= 5),
		     mine AS (SELECT DISTINCT b.id, b."serviceId", EXTRACT(EPOCH FROM (b."completedAt" - b."startedAt")) AS s FROM "Booking" b
		             LEFT JOIN "BookingAssignment" a ON a."bookingId" = b.id
		             WHERE (a."employeeId" = $1 OR b."projectSupervisorId" = $1) AND b."startedAt" IS NOT NULL AND b."completedAt" > b."startedAt"
		               AND b."completedAt" - b."startedAt" < interval '14 hours' AND b."completedAt" > now() - interval '30 days')
		SELECT round(avg(mine.s / NULLIF(med.m, 0))::numeric, 2)::float8, count(*)::int FROM mine JOIN med ON med."serviceId" = mine."serviceId"`, id).Scan(&sp.Ratio, &sp.N)
	if sp.N >= 3 {
		d.SpeedRatio, d.SpeedN = sp.Ratio, sp.N
	}
	errlog.

		// الالتزام والمشاكل
		Err(db.Get(&d.Attendance30, `SELECT count(DISTINCT date::date) FROM "Attendance" WHERE "employeeId" = $1 AND date::date > current_date - 30`, id))
	_ = db.QueryRow(`SELECT count(*), count(*) FILTER (WHERE late) FROM "DailyWorkReport" WHERE "employeeId" = $1 AND day > current_date - 30`, id).Scan(&d.Reports30, &d.ReportsLate30)
	errlog.Err(db.Get(&d.Negligence90, `SELECT count(*) FROM "EmployeeNegligence" WHERE "employeeId" = $1 AND "voidedAt" IS NULL AND day > current_date - 90`, id))
	errlog.Err(db.Get(&d.Discipline90, `SELECT COALESCE(-sum(delta) FILTER (WHERE delta < 0), 0)::int FROM "DisciplineEvent" WHERE "employeeId" = $1 AND "createdAt" > now() - interval '90 days'`, id))
	_ = db.QueryRow(`SELECT count(*), count(*) FILTER (WHERE verdict = 'REJECTED') FROM "MonitorJustification" WHERE "employeeId" = $1 AND "createdAt" > now() - interval '90 days'`, id).Scan(&d.Justif90, &d.JustifRejected)
	errlog.Err(db.Get(&d.Complaints90, `SELECT count(*) FROM "Complaint" WHERE "relatedEmployeeId" = $1 AND "createdAt" > now() - interval '90 days'`, id))
	errlog.Err(db.Get(&d.Issues90, `SELECT count(*) FROM "WorkplaceIssue" WHERE ("partyAId" = $1 OR "partyBId" = $1) AND "createdAt" > now() - interval '90 days'`, id))
	errlog.Err(db.Get(&d.PeerNotes90, `SELECT count(*) FROM "PeerNote" WHERE "targetId" = $1 AND "createdAt" > now() - interval '90 days'`, id))

	d.ReportMatchPct, d.ReportMatchN = MatchRate(db, id, 30)

	d.Flags = dossierFlags(d)
	return d, nil
}

// trendOf آخر شهر بيه بيانات مقابل معدل الشهرين الي قبله (ماتركس أول، وإلا البشر).
func trendOf(ms []DossierMonth) (string, string) {
	pick := func(m DossierMonth) *float64 {
		if m.MatrixPct != nil {
			return m.MatrixPct
		}
		if m.HumanAvg != nil {
			v := *m.HumanAvg * 20 // ٥ نجوم ← ١٠٠
			return &v
		}
		return nil
	}
	vals := []float64{}
	labels := []string{}
	for _, m := range ms {
		if v := pick(m); v != nil {
			vals = append(vals, *v)
			labels = append(labels, m.Month)
		}
	}
	if len(vals) < 2 {
		return "UNKNOWN", "بعد ماكو أشهر كافية للمقارنة."
	}
	last := vals[len(vals)-1]
	prev := vals[max(0, len(vals)-3) : len(vals)-1]
	sum := 0.0
	for _, v := range prev {
		sum += v
	}
	avg := sum / float64(len(prev))
	delta := round1(last - avg)
	switch {
	case delta >= 5:
		return "UP", fmt.Sprintf("يتقدم: %s صار %.0f٪ مقابل %.0f٪ قبلها (+%.0f).", labels[len(labels)-1], last, avg, delta)
	case delta <= -5:
		return "DOWN", fmt.Sprintf("يتراجع: %s صار %.0f٪ مقابل %.0f٪ قبلها (%.0f).", labels[len(labels)-1], last, avg, delta)
	}
	return "STEADY", fmt.Sprintf("ثابت حول %.0f٪.", last)
}

func dossierFlags(d *Dossier) []string {
	f := []string{}
	if d.ReportMatchN >= 3 && d.ReportMatchPct < 60 {
		f = append(f, fmt.Sprintf("🧾 تقارير إنجازه ما تطابق شغله المسجّل إلا %d٪ (%d تقرير) — يكتب غير الي يسويه؟", d.ReportMatchPct, d.ReportMatchN))
	}
	if d.Trend == "DOWN" {
		f = append(f, "📉 "+d.TrendNote)
	}
	if d.SpeedRatio != nil && *d.SpeedRatio >= 1.4 {
		f = append(f, fmt.Sprintf("🐢 أبطأ من المعتاد %.1f مرة على %d حجز — يستاهل نشوف ليش (تدريب؟ أدوات؟).", *d.SpeedRatio, d.SpeedN))
	}
	if d.SpeedRatio != nil && *d.SpeedRatio <= 0.7 {
		f = append(f, fmt.Sprintf("⚡ أسرع من المعتاد (%.1f) — تأكد الجودة ما نازلة.", *d.SpeedRatio))
	}
	if d.Negligence90 >= 3 {
		f = append(f, fmt.Sprintf("📒 فاته واجب التقييم %d مرات بـ٩٠ يوم.", d.Negligence90))
	}
	if d.ReportsLate30 >= 3 {
		f = append(f, fmt.Sprintf("📝 %d تقرير إنجاز متأخر بـ٣٠ يوم.", d.ReportsLate30))
	}
	if d.JustifRejected >= 2 {
		f = append(f, fmt.Sprintf("⚠️ %d تبرير انرفض من المراقب.", d.JustifRejected))
	}
	if d.Complaints90 > 0 {
		f = append(f, fmt.Sprintf("😠 %d شكوى زبون مرتبطة بيه.", d.Complaints90))
	}
	if d.Issues90 > 0 {
		f = append(f, fmt.Sprintf("🤝 طرف بـ%d مشكلة وظيفية.", d.Issues90))
	}
	if d.Major == "" {
		f = append(f, "🧾 ما عبّى استمارة معلوماته بعد.")
	}
	if d.ContractEnd != "" {
		if t, err := time.Parse("2006-01-02", d.ContractEnd); err == nil && time.Until(t) < 31*24*time.Hour && time.Until(t) > 0 {
			f = append(f, "📅 عقده يخلص "+d.ContractEnd+" — قرّب.")
		}
	}
	if len(d.Weak) > 0 {
		w := []string{}
		for _, r := range d.Weak {
			w = append(w, fmt.Sprintf("%s %.0f٪", r.Title, r.Pct))
		}
		f = append(f, "🔻 أضعف شي: "+strings.Join(w, "، "))
	}
	return f
}

// Text نص الملف لماتركس (بلا راتب).
func (d *Dossier) Text() string {
	b := &strings.Builder{}
	w := func(f string, a ...any) { fmt.Fprintf(b, f+"\n", a...) }
	w("الموظف: %s · الدور %s (الفعلي %s)%s", d.Name, d.Role, d.Effective, map[bool]string{true: " · ليدر", false: ""}[d.IsLeader])
	if d.Position != "" {
		w("منصبه بالهيكل: %s", d.Position)
	}
	w("الوظيفة: %s · التعيين %s · نهاية العقد %s", or(d.JobTitle), or(d.HireDate), or(d.ContractEnd))
	w("الدراسة: %s / %s · الشهادة بالملف: %s · ساعاته المطلوبة: %s", or(d.Major), or(d.Degree), or(d.Certificate), intOr(d.MonthlyHours))
	if d.CompanyRating != nil {
		w("رأيه بالشركة: %d/5 %s", *d.CompanyRating, d.CompanyNote)
	}
	w("المهارات: %s", orList(d.Skills))
	w("التدريب: %s", orList(d.Training))
	if len(d.LeaderSkl) > 0 {
		w("مهارات الليدرية: %s", strings.Join(d.LeaderSkl, "، "))
	}
	mm := []string{}
	for _, m := range d.Months {
		s := m.Month + ": "
		if m.MatrixPct != nil {
			s += fmt.Sprintf("ماتركس %.0f٪", *m.MatrixPct)
		} else {
			s += "ماتركس —"
		}
		if m.HumanAvg != nil {
			s += fmt.Sprintf("، بشر %.1f/5 (%d)", *m.HumanAvg, m.Ratings)
		}
		mm = append(mm, s)
	}
	w("المنحنى (٦ أشهر): %s", strings.Join(mm, " | "))
	w("الاتجاه: %s", d.TrendNote)
	sr := func(rs []DossierRule) string {
		x := []string{}
		for _, r := range rs {
			x = append(x, fmt.Sprintf("%s %.0f٪ (%d)", r.Title, r.Pct, r.N))
		}
		return orList(x)
	}
	w("أقوى شي: %s", sr(d.Strong))
	w("أضعف شي: %s", sr(d.Weak))
	if d.SpeedRatio != nil {
		w("السرعة: مدته %.2f من المعتاد على %d حجز (١ = طبيعي)", *d.SpeedRatio, d.SpeedN)
	} else {
		w("السرعة: ماكو حجوزات كافية")
	}
	w("آخر ٣٠ يوم: دوام %d يوم، تقارير إنجاز %d (متأخرة %d)", d.Attendance30, d.Reports30, d.ReportsLate30)
	if d.ReportMatchN > 0 {
		w("صدق التقارير: %d٪ منها تطابق الشغل المسجّل بالنظام (%d مطابقة)", d.ReportMatchPct, d.ReportMatchN)
	}
	w("آخر ٩٠ يوم: إهمال تقييم %d، خصم انضباط %d نقطة، تبريرات %d (مرفوضة %d)، شكاوي %d، مشاكل وظيفية %d، ملاحظات زملاء عنه %d",
		d.Negligence90, d.Discipline90, d.Justif90, d.JustifRejected, d.Complaints90, d.Issues90, d.PeerNotes90)
	if len(d.Flags) > 0 {
		w("ملاحظات: %s", strings.Join(d.Flags, " · "))
	}
	return b.String()
}

func or(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
func orList(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, "، ")
}
func intOr(p *int) string {
	if p == nil {
		return "—"
	}
	return fmt.Sprint(*p)
}
