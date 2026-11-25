package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/timeutil"
)

// ═══ 💼 تكاليف الشركة وأهداف الموظفين (ل) 11-24 دازاي ═══
// «المفروض تطلع رواتب كل الموظفين من الحجوزات وزيادة، بالإضافة لمصاريف العدد والكاز والسيارات…
//  نجمع كل رواتبهم ومصاريف الشهر ونقسمه حسب ساعات كل موظف وراتبه — كل موظف شكد لازم يدخّل باليوم وبالساعة».
// «رواتب الإداريين من المبلغ الإجمالي؛ إداري حجوزات الشفت الصباحي من فلوس الصباحي، والمسائي من المسائي».
// «فترات التوقف ما يتحاسب عليها الموظف». كله إدخال يدوي من الإعدادات (المالك بس).
//
// الحساب (لكل موظف «يدخّل فلوس»):
//   أيام شغله = أيامه بالشهر − إجازاته − أيام التوقف (بنسبة أيام شغله من الشهر)
//   وزنه = راتبه × ساعاته بالشهر (ساعات اليوم × أيام شغله)
//   هدفه = (راتبه + حصته بالوزن من [رواتب الإداريين + مصاريف الشهر] + حصته بالوزن من رواتب إداريي حجوزات شفته)
//          × (1 + نسبة ربح الشركة)
//   باليوم = هدفه ÷ أيام شغله · بالساعة = باليوم ÷ ساعات يومه

type CostEmployee struct {
	EmployeeID   string  `db:"employeeId" json:"employeeId"`
	Name         string  `db:"name" json:"name"`
	Kind         string  `db:"kind" json:"kind"`
	Shift        *string `db:"shift" json:"shift"`
	Salary       float64 `db:"salary" json:"salary"`
	BaseSalary   float64 `db:"baseSalary" json:"baseSalary"`
	HoursPerDay  float64 `db:"hoursPerDay" json:"hoursPerDay"`
	DaysPerMonth float64 `db:"daysPerMonth" json:"daysPerMonth"`
	LeaveDays    float64 `db:"leaveDays" json:"leaveDays"`
}

type CostExpense struct {
	ID     string  `db:"id" json:"id"`
	Month  string  `db:"month" json:"month"`
	Kind   string  `db:"kind" json:"kind"`
	Amount float64 `db:"amount" json:"amount"`
	Note   *string `db:"note" json:"note"`
}

type DowntimePeriod struct {
	ID        string `db:"id" json:"id"`
	StartDate string `db:"startDate" json:"startDate"`
	EndDate   string `db:"endDate" json:"endDate"`
	Reason    string `db:"reason" json:"reason"`
}

type EmployeeTarget struct {
	EmployeeID string  `json:"employeeId"`
	Name       string  `json:"name"`
	Shift      *string `json:"shift"`
	Salary     float64 `json:"salary"`
	WorkDays   float64 `json:"workDays"`
	Overhead   float64 `json:"overhead"` // حصته من رواتب الإداريين والمصاريف
	ShiftShare float64 `json:"shiftShare"`
	Target     float64 `json:"target"` // بالشهر
	Daily      float64 `json:"daily"`
	Hourly     float64 `json:"hourly"`
	Brought    float64 `json:"brought"` // الي دخّله لحد هسه (WORK)
	Remaining  float64 `json:"remaining"`
	Forecast   float64 `json:"forecast"` // على هالمنوال لآخر الشهر
	Pct        float64 `json:"pct"`      // من الهدف
	// (ل) 11-24 دازاي «الأساسي 350 والباقي حسب الهدف»
	BaseSalary     float64 `json:"baseSalary"`
	ExpectedNow    float64 `json:"expectedNow"`    // حسب الي دخّله لحد هسه
	ExpectedSalary float64 `json:"expectedSalary"` // حسب التوقع لآخر الشهر
	DailyNeeded    float64 `json:"dailyNeeded"`    // شكد لازم يدخّل باليوم بباقي الشهر حتى يوصل
	Errors         int     `json:"errors"`         // أخطاء فنية انوافق عليها (هو الكادر الأول)
	Fixes          int     `json:"fixes"`          // صلّح أخطاء غيره
	Rank           int     `json:"rank"`
}

// ExpectedSalary الأساسي + (الراتب − الأساسي) × نسبة الوصول (سقفها ١٠٠٪، وما ينزل عن الأساسي).
func ExpectedSalary(salary, base, brought, target float64) float64 {
	if base > salary {
		base = salary
	}
	if target <= 0 {
		return salary
	}
	r := math.Max(0, math.Min(1, brought/target))
	return math.Round(base + (salary-base)*r)
}

type CostsReport struct {
	Month         string             `json:"month"`
	MarginPct     float64            `json:"marginPct"`
	Employees     []CostEmployee     `json:"employees"`
	Expenses      []CostExpense      `json:"expenses"`
	Downtime      []DowntimePeriod   `json:"downtime"`
	DowntimeDays  int                `json:"downtimeDays"`
	SalariesEarn  float64            `json:"salariesEarn"`
	SalariesAdmin float64            `json:"salariesAdmin"`
	SalariesShift map[string]float64 `json:"salariesShift"`
	ExpensesTotal float64            `json:"expensesTotal"`
	TotalCost     float64            `json:"totalCost"`
	TotalTarget   float64            `json:"totalTarget"`
	Targets       []EmployeeTarget   `json:"targets"`
}

// ComputeTargets الحساب الصافي (بلا قاعدة بيانات) — للاختبار.
func ComputeTargets(emps []CostEmployee, expenses, marginPct float64, downtimeDays, daysInMonth int) ([]EmployeeTarget, float64, float64, map[string]float64) {
	adminPool, shiftPool := expenses, map[string]float64{}
	earners := []CostEmployee{}
	var salAdmin float64
	for _, e := range emps {
		switch e.Kind {
		case "EARNER":
			earners = append(earners, e)
		case "BOOKING_ADMIN":
			if e.Shift != nil {
				shiftPool[*e.Shift] += e.Salary
				continue
			}
			adminPool += e.Salary
			salAdmin += e.Salary
		default:
			adminPool += e.Salary
			salAdmin += e.Salary
		}
	}
	workDays := func(e CostEmployee) float64 {
		dt := 0.0
		if daysInMonth > 0 {
			dt = float64(downtimeDays) * e.DaysPerMonth / float64(daysInMonth)
		}
		return math.Max(1, e.DaysPerMonth-e.LeaveDays-dt)
	}
	weight := func(e CostEmployee) float64 { return e.Salary * e.HoursPerDay * workDays(e) }
	var wAll float64
	wShift := map[string]float64{}
	for _, e := range earners {
		w := weight(e)
		wAll += w
		if e.Shift != nil {
			wShift[*e.Shift] += w
		}
	}
	// شفت بيه إداري حجوزات وماكو بيه فنيين ← راتبه يتوزع على الكل
	for s, v := range shiftPool {
		if wShift[s] == 0 {
			adminPool += v
			delete(shiftPool, s)
		}
	}
	out := []EmployeeTarget{}
	for _, e := range earners {
		w, d := weight(e), workDays(e)
		t := EmployeeTarget{EmployeeID: e.EmployeeID, Name: e.Name, Shift: e.Shift, Salary: e.Salary, BaseSalary: e.BaseSalary, WorkDays: math.Round(d*10) / 10}
		if wAll > 0 {
			t.Overhead = adminPool * w / wAll
		}
		if e.Shift != nil && wShift[*e.Shift] > 0 {
			t.ShiftShare = shiftPool[*e.Shift] * w / wShift[*e.Shift]
		}
		t.Target = math.Round((e.Salary + t.Overhead + t.ShiftShare) * (1 + marginPct/100))
		t.Overhead, t.ShiftShare = math.Round(t.Overhead), math.Round(t.ShiftShare)
		t.Daily = math.Round(t.Target / d)
		t.Hourly = math.Round(t.Daily / e.HoursPerDay)
		out = append(out, t)
	}
	return out, salAdmin, adminPool - expenses - salAdmin, shiftPool
}

type CompanyCosts struct{ db *sqlx.DB }

func NewCompanyCosts(db *sqlx.DB) *CompanyCosts { return &CompanyCosts{db: db} }

// DB للتقارير الجانبية (مؤشر الهشاشة)
func (c *CompanyCosts) DB() *sqlx.DB { return c.db }

// downtimeDaysIn أيام التوقف (اتحاد الفترات) الي تطيح جوه الشهر.
func (c *CompanyCosts) downtimeDaysIn(from, to time.Time) int {
	var n int
	_ = c.db.Get(&n, `SELECT count(DISTINCT d)::int FROM "DowntimePeriod" p,
		generate_series(GREATEST(p."startDate", $1::date), LEAST(p."endDate", $2::date - 1), interval '1 day') d
		WHERE p."endDate" >= $1::date AND p."startDate" < $2::date`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	return n
}

// Report كل شي لشهر: الإعدادات والأهداف ووين وصل كل موظف.
func (c *CompanyCosts) Report(month string, now time.Time) (*CostsReport, error) {
	from, to, ok := timeutil.MonthStart(month)
	if !ok {
		return nil, errors.New("الشهر غلط")
	}
	r := &CostsReport{Month: month, Employees: []CostEmployee{}, Expenses: []CostExpense{}, Downtime: []DowntimePeriod{}}
	_ = c.db.Get(&r.MarginPct, `SELECT "marginPct"::float8 FROM "CostConfig" WHERE id = 1`)
	if err := c.db.Select(&r.Employees, `SELECT ce."employeeId", e.name, ce.kind, ce.shift, ce.salary::float8 AS salary, ce."baseSalary"::float8 AS "baseSalary",
		ce."hoursPerDay"::float8 AS "hoursPerDay", ce."daysPerMonth"::float8 AS "daysPerMonth", ce."leaveDays"::float8 AS "leaveDays"
		FROM "CostEmployee" ce JOIN "Employee" e ON e.id = ce."employeeId" ORDER BY ce.kind, e.name`); err != nil {
		return nil, err
	}
	if err := c.db.Select(&r.Expenses, `SELECT id, month, kind, amount::float8 AS amount, note FROM "CostExpense" WHERE month = $1 ORDER BY "createdAt"`, month); err != nil {
		return nil, err
	}
	if err := c.db.Select(&r.Downtime, `SELECT id, "startDate"::text AS "startDate", "endDate"::text AS "endDate", reason
		FROM "DowntimePeriod" WHERE "endDate" >= $1::date - 60 ORDER BY "startDate" DESC`, from.Format("2006-01-02")); err != nil {
		return nil, err
	}
	for _, x := range r.Expenses {
		r.ExpensesTotal += x.Amount
	}
	r.DowntimeDays = c.downtimeDaysIn(from, to)
	daysInMonth := int(to.Sub(from).Hours()/24 + 0.5)
	var bookingAdmins float64
	r.Targets, r.SalariesAdmin, bookingAdmins, r.SalariesShift = ComputeTargets(r.Employees, r.ExpensesTotal, r.MarginPct, r.DowntimeDays, daysInMonth)
	r.SalariesAdmin += bookingAdmins
	for _, e := range r.Employees {
		if e.Kind == "EARNER" {
			r.SalariesEarn += e.Salary
		}
	}
	var shiftSal float64
	for _, v := range r.SalariesShift {
		shiftSal += v
	}
	r.TotalCost = r.SalariesEarn + r.SalariesAdmin + shiftSal + r.ExpensesTotal
	// وين وصل كل واحد — نفس «الفلوس الي دخّلها» بالتقييم (توزيع الفلوس من الإعدادات)
	brought := map[string]float64{}
	if rev, err := repository.NewMatrixScoreRepository(c.db).Revenue(from.UTC(), to.UTC(), ""); err == nil {
		for _, x := range rev {
			if x.Kind == "WORK" {
				brought[x.EmployeeID] += x.Amount
			}
		}
	}
	frac := 1.0 // كم مضى من الشهر
	if now.Before(to) {
		frac = math.Max(0.02, now.Sub(from).Hours()/to.Sub(from).Hours())
	}
	type cnt struct {
		ID string `db:"id"`
		N  int    `db:"n"`
	}
	errs, fixes := []cnt{}, []cnt{}
	_ = c.db.Select(&errs, `SELECT ra."employeeId" AS id, count(DISTINCT ra."claimId")::int AS n FROM "ReworkAdjustment" ra
		WHERE ra.amount < 0 AND ra."createdAt" >= $1 AND ra."createdAt" < $2 GROUP BY 1`, from.UTC(), to.UTC())
	_ = c.db.Select(&fixes, `SELECT ra."employeeId" AS id, count(DISTINCT ra."claimId")::int AS n FROM "ReworkAdjustment" ra
		WHERE ra.amount > 0 AND ra."createdAt" >= $1 AND ra."createdAt" < $2 GROUP BY 1`, from.UTC(), to.UTC())
	em, fm := map[string]int{}, map[string]int{}
	for _, x := range errs {
		em[x.ID] = x.N
	}
	for _, x := range fixes {
		fm[x.ID] = x.N
	}
	for i := range r.Targets {
		t := &r.Targets[i]
		r.TotalTarget += t.Target
		t.Brought = math.Round(brought[t.EmployeeID])
		t.Remaining = math.Max(0, t.Target-t.Brought)
		t.Forecast = math.Round(t.Brought / frac)
		if t.Target > 0 {
			t.Pct = math.Round(t.Brought / t.Target * 100)
		}
		t.ExpectedNow = ExpectedSalary(t.Salary, t.BaseSalary, t.Brought, t.Target)
		t.ExpectedSalary = ExpectedSalary(t.Salary, t.BaseSalary, t.Forecast, t.Target)
		if left := t.WorkDays * (1 - frac); left >= 0.5 {
			t.DailyNeeded = math.Round(t.Remaining / left)
		}
		t.Errors, t.Fixes = em[t.EmployeeID], fm[t.EmployeeID]
	}
	// الترتيب: الأعلى وصولاً لهدفه أول
	order := make([]int, len(r.Targets))
	for i := range order {
		order[i] = i
	}
	for i := 1; i < len(order); i++ {
		for j := i; j > 0 && r.Targets[order[j]].Pct > r.Targets[order[j-1]].Pct; j-- {
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	sorted := make([]EmployeeTarget, len(order))
	for k, i := range order {
		sorted[k] = r.Targets[i]
		sorted[k].Rank = k + 1
	}
	r.Targets = sorted
	return r, nil
}

// MyTarget هدف موظف واحد لشاشته (nil إذا مو «يدخّل فلوس» بالتكاليف).
func (c *CompanyCosts) MyTarget(emp string, now time.Time) (*EmployeeTarget, error) {
	rep, err := c.Report(timeutil.ThisMonth(), now)
	if err != nil {
		return nil, err
	}
	for i := range rep.Targets {
		if rep.Targets[i].EmployeeID == emp {
			t := rep.Targets[i]
			return &t, nil
		}
	}
	return nil, nil
}

// ─── الحفظ ───

func (c *CompanyCosts) SaveEmployee(e CostEmployee) error {
	if e.Kind != "EARNER" && e.Kind != "ADMIN" && e.Kind != "BOOKING_ADMIN" {
		return errors.New("النوع غلط")
	}
	if e.Shift != nil && *e.Shift != "MORNING" && *e.Shift != "EVENING" {
		e.Shift = nil
	}
	if e.BaseSalary <= 0 {
		e.BaseSalary = 350000
	}
	if e.Salary < 0 || e.BaseSalary > e.Salary && e.Salary > 0 || e.HoursPerDay <= 0 || e.HoursPerDay > 24 || e.DaysPerMonth <= 0 || e.DaysPerMonth > 31 || e.LeaveDays < 0 || e.LeaveDays >= e.DaysPerMonth {
		return errors.New("تأكد من الأرقام: الراتب (والأساسي ما يعبره)، ساعات اليوم (١–٢٤)، أيام الشهر (١–٣١)، والإجازات أقل من أيام الشهر")
	}
	_, err := c.db.Exec(`INSERT INTO "CostEmployee" ("employeeId", kind, shift, salary, "hoursPerDay", "daysPerMonth", "leaveDays", "baseSalary", "updatedAt")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		ON CONFLICT ("employeeId") DO UPDATE SET kind = $2, shift = $3, salary = $4, "hoursPerDay" = $5, "daysPerMonth" = $6, "leaveDays" = $7, "baseSalary" = $8, "updatedAt" = now()`,
		e.EmployeeID, e.Kind, e.Shift, e.Salary, e.HoursPerDay, e.DaysPerMonth, e.LeaveDays, e.BaseSalary)
	if err != nil && strings.Contains(err.Error(), "foreign key") {
		return errors.New("الموظف مو موجود")
	}
	return err
}

func (c *CompanyCosts) RemoveEmployee(id string) error {
	_, err := c.db.Exec(`DELETE FROM "CostEmployee" WHERE "employeeId" = $1`, id)
	return err
}

func (c *CompanyCosts) AddExpense(x CostExpense, by string) error {
	if _, _, ok := timeutil.MonthStart(x.Month); !ok {
		return errors.New("الشهر غلط")
	}
	if x.Amount < 0 {
		return errors.New("المبلغ ما يصير سالب")
	}
	_, err := c.db.Exec(`INSERT INTO "CostExpense" (month, kind, amount, note, "createdById") VALUES ($1, $2, $3, NULLIF(btrim(COALESCE($4, '')), ''), NULLIF($5, ''))`,
		x.Month, x.Kind, x.Amount, x.Note, by)
	if err != nil && strings.Contains(err.Error(), "cost_expense_kind_ok") {
		return errors.New("نوع المصروف غلط")
	}
	return err
}

func (c *CompanyCosts) RemoveExpense(id string) error {
	_, err := c.db.Exec(`DELETE FROM "CostExpense" WHERE id = $1`, id)
	return err
}

func (c *CompanyCosts) AddDowntime(d DowntimePeriod, by string) error {
	d.Reason = strings.TrimSpace(d.Reason)
	if d.Reason == "" {
		return errors.New("اكتب السبب (مثلاً: عيد، زيارة الأربعين)")
	}
	_, err := c.db.Exec(`INSERT INTO "DowntimePeriod" ("startDate", "endDate", reason, "createdById") VALUES ($1::date, $2::date, $3, NULLIF($4, ''))`,
		d.StartDate, d.EndDate, d.Reason, by)
	if err != nil {
		return errors.New("التواريخ غلط (لازم النهاية بعد البداية)")
	}
	return nil
}

func (c *CompanyCosts) RemoveDowntime(id string) error {
	_, err := c.db.Exec(`DELETE FROM "DowntimePeriod" WHERE id = $1`, id)
	return err
}

func (c *CompanyCosts) SetMargin(p float64) error {
	if p < 0 || p > 500 {
		return errors.New("نسبة الربح بين ٠ و٥٠٠")
	}
	_, err := c.db.Exec(`INSERT INTO "CostConfig" (id, "marginPct", "updatedAt") VALUES (1, $1, now())
		ON CONFLICT (id) DO UPDATE SET "marginPct" = $1, "updatedAt" = now()`, p)
	return err
}

// ═══ ماتركس يتابع الهدف (ل) 11-24 دازاي ═══
// «ماتركس يتابع كل موظف شكد معدل إدخاله اليومي، وعلى هالمنوال شكد راح يدخّل لنهاية الشهر، ويتوقع وينبّه».
// مرة بالليل (بعد ٨ ببغداد): كل موظف «يدخّل فلوس» يوصله وين وصل، والمالك يوصله منو متأخر عن هدفه.

// TargetMessage رسالة الموظف.
func TargetMessage(t EmployeeTarget) string {
	msg := fmt.Sprintf("🎯 ماتركس: دخّلت هالشهر %s من هدفك %s (%d٪).", FmtIQD(t.Brought), FmtIQD(t.Target), int(t.Pct))
	if t.Remaining > 0 && t.DailyNeeded > 0 {
		msg += fmt.Sprintf(" حتى توصل تحتاج %s باليوم بباقي الشهر.", FmtIQD(t.DailyNeeded))
	} else if t.Remaining <= 0 {
		msg += " وصلت هدفك، عاشت إيدك 👏"
	}
	msg += fmt.Sprintf(" على هالمنوال توصل %s، وراتبك المتوقع %s.", FmtIQD(t.Forecast), FmtIQD(t.ExpectedSalary))
	return msg
}

// NotifyTargets مرة باليوم بعد ٨ المسا ببغداد.
func (c *CompanyCosts) NotifyTargets(now time.Time, claim func(key, day string) (bool, error), notify func(employeeID, typ, msg string) error, owner func(msg string)) error {
	local := now.In(timeutil.CompanyLocation())
	if local.Hour() < 20 {
		return nil
	}
	ok, err := claim("MONEY_TARGET", local.Format("2006-01-02"))
	if err != nil || !ok {
		return err
	}
	rep, err := c.Report(timeutil.ThisMonth(), now)
	if err != nil {
		return err
	}
	behind := []string{}
	for _, t := range rep.Targets {
		_ = notify(t.EmployeeID, "MONEY_TARGET", TargetMessage(t))
		if t.Forecast < t.Target {
			behind = append(behind, fmt.Sprintf("%s (%d٪، متوقع %s من %s)", t.Name, int(t.Pct), FmtIQD(t.Forecast), FmtIQD(t.Target)))
		}
	}
	if len(behind) > 0 && owner != nil {
		owner("🎯 ماتركس — متأخرين عن هدف الشهر على هالمنوال: " + strings.Join(behind, "، "))
	}
	return nil
}
