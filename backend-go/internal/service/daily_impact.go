package service

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 📊 أثرك اليوم (دازاي 11-21 — أفكار ١: «من أمن العقاب أساء الأدب») ═══
// العقوبة تردع بيقينها وقربها مو بشدتها: كل موظف يوصله بالليل شنو كسب وشنو خسر اليوم —
// يبدي بالزين («من أمن الشكر قلّ جهده»). المالك والمدير ما ينقيّمون فما يوصلهم.

type DailyImpact struct {
	EmployeeID string  `db:"employeeId"`
	Full       int     `db:"full"`
	Partial    int     `db:"partial"`
	Lost       int     `db:"lost"`
	Earned     int     `db:"earned"`
	Max        int     `db:"max"`
	Worst      *string `db:"worst"`
}

// DailyImpacts نقاط ماتركس لكل موظف بيوم بغداد (day).
func DailyImpacts(db *sqlx.DB, day time.Time) ([]DailyImpact, error) {
	d := day.In(debriefLoc)
	from := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, debriefLoc)
	out := []DailyImpact{}
	err := db.Select(&out, `
		SELECT m."employeeId",
		       COUNT(*) FILTER (WHERE m.points = m."maxPoints")::int AS full,
		       COUNT(*) FILTER (WHERE m.points > 0 AND m.points < m."maxPoints")::int AS partial,
		       COUNT(*) FILTER (WHERE m.points = 0)::int AS lost,
		       SUM(m.points)::int AS earned, SUM(m."maxPoints")::int AS max,
		       (ARRAY_AGG(m.reason ORDER BY m.points - m."maxPoints", m."at") FILTER (WHERE m.points < m."maxPoints"))[1] AS worst
		FROM "MatrixScore" m JOIN "Employee" e ON e.id = m."employeeId"
		WHERE m."cancelledAt" IS NULL AND m."at" >= $1 AND m."at" < $2
		  AND e.status = 'ACTIVE' AND e.role NOT IN ('ADMIN','OWNER')
		GROUP BY m."employeeId"`, from, from.Add(24*time.Hour))
	return out, err
}

// DailyImpactMessage نص الإشعار — الزين أول.
func DailyImpactMessage(x DailyImpact) string {
	pct := 0
	if x.Max > 0 {
		pct = x.Earned * 100 / x.Max
	}
	msg := fmt.Sprintf("📊 أثرك اليوم: %d٪ — ✅ %d خطوة كاملة", pct, x.Full)
	if x.Partial > 0 {
		msg += fmt.Sprintf(" · ⚠️ %d ناقصة", x.Partial)
	}
	if x.Lost > 0 {
		msg += fmt.Sprintf(" · ❌ %d فاتت", x.Lost)
	}
	if x.Worst != nil {
		w := []rune(*x.Worst)
		if len(w) > 90 {
			w = append(w[:90], '…')
		}
		msg += ". أكثر شي نزّلك: " + string(w) + " — إذا عندك عذر اعتذر من «تقييمي»."
	} else {
		msg += ". يوم كامل، عاشت إيدك 👏"
	}
	return msg
}

// NotifyDailyImpact مرة باليوم بعد ٨ المسا ببغداد.
func NotifyDailyImpact(db *sqlx.DB, now time.Time, claim func(key, day string) (bool, error), notify func(employeeID, typ, msg string) error) error {
	now = now.In(debriefLoc)
	if now.Hour() < 20 {
		return nil
	}
	ok, err := claim("DAILY_IMPACT", now.Format("2006-01-02"))
	if err != nil || !ok {
		return err
	}
	rows, err := DailyImpacts(db, now)
	if err != nil {
		return err
	}
	for _, x := range rows {
		_ = notify(x.EmployeeID, "DAILY_IMPACT", DailyImpactMessage(x))
	}
	return nil
}
