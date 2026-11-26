package repository

import (
	"fmt"
	"time"
)

// ═══ (دازاي 11-26) فترة ساعات البيت مو الشهر الميلادي ═══
// «شهرنه يبدي من يوم 26 وينتهي يوم 25»: الفترة تنسمّى بالشهر الي تنتهي بيه —
// فترة 2026-12 = من 2026-11-26 إلى 2026-12-25. وموعد الرفع (التسكير) يحدده الأدمن.

type RemoteHoursConfig struct {
	PeriodStartDay int `db:"periodStartDay" json:"periodStartDay"`
	CloseDay       int `db:"closeDay" json:"closeDay"`
	CloseHour      int `db:"closeHour" json:"closeHour"`
}

// DefaultRemoteHoursConfig قبل ما الأدمن يحفظ شي.
var DefaultRemoteHoursConfig = RemoteHoursConfig{PeriodStartDay: 26, CloseDay: 25, CloseHour: 23}

func (c RemoteHoursConfig) Valid() bool {
	return c.PeriodStartDay >= 1 && c.PeriodStartDay <= 28 && c.CloseDay >= 1 && c.CloseDay <= 28 && c.CloseHour >= 0 && c.CloseHour <= 23
}

func (r *RemoteHoursRepository) Config() RemoteHoursConfig {
	c := DefaultRemoteHoursConfig
	if err := r.db.Get(&c, `SELECT "periodStartDay", "closeDay", "closeHour" FROM "RemoteHoursConfig" WHERE id = 1`); err != nil {
		return DefaultRemoteHoursConfig
	}
	return c
}

func (r *RemoteHoursRepository) SaveConfig(c RemoteHoursConfig, by string) error {
	_, err := r.db.Exec(`INSERT INTO "RemoteHoursConfig" (id, "periodStartDay", "closeDay", "closeHour", "updatedById", "updatedAt")
		VALUES (1, $1, $2, $3, NULLIF($4,''), now())
		ON CONFLICT (id) DO UPDATE SET "periodStartDay" = $1, "closeDay" = $2, "closeHour" = $3, "updatedById" = NULLIF($4,''), "updatedAt" = now()`,
		c.PeriodStartDay, c.CloseDay, c.CloseHour, by)
	return err
}

// PeriodRange أول وآخر يوم بفترة الشهر label (YYYY-MM).
func PeriodRange(label string, startDay int) (time.Time, time.Time) {
	m, err := time.Parse("2006-01", label)
	if err != nil {
		m = time.Now()
	}
	end := time.Date(m.Year(), m.Month(), 1, 0, 0, 0, 0, time.UTC)
	if startDay <= 1 {
		return end, end.AddDate(0, 1, -1)
	}
	from := time.Date(m.Year(), m.Month()-1, startDay, 0, 0, 0, 0, time.UTC)
	return from, time.Date(m.Year(), m.Month(), startDay-1, 0, 0, 0, 0, time.UTC)
}

// PeriodOf اسم الفترة الي يوقع بيها اليوم d.
func PeriodOf(d time.Time, startDay int) string {
	if startDay > 1 && d.Day() >= startDay {
		return time.Date(d.Year(), d.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
	}
	return d.Format("2006-01")
}

// periodSQL تعبير SQL يعطي اسم الفترة لعمود تاريخ (startDay رقم مضبوط 1..28).
func periodSQL(col string, startDay int) string {
	if startDay <= 1 {
		return fmt.Sprintf(`to_char(%s, 'YYYY-MM')`, col)
	}
	return fmt.Sprintf(`to_char(%s - %d + interval '1 month', 'YYYY-MM')`, col, startDay-1)
}

// CloseUpTo يرفع (يسكّر) السجلات المفتوحة لحد آخر يوم بالفترة — الي بعده يبقى مفتوح.
func (r *RemoteHoursRepository) CloseUpTo(lastDay time.Time, byID string) (int64, error) {
	res, err := r.db.Exec(`UPDATE "RemoteWorkEntry" SET "closedAt" = now(), "closedById" = NULLIF($1,'')
		WHERE "closedAt" IS NULL AND "workDate" <= $2::date`, byID, lastDay.Format("2006-01-02"))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CloseMoment موعد رفع الفترة label: أول يوم closeDay بعد (أو بنفس) آخر يوم بالفترة، بساعة closeHour.
func CloseMoment(label string, c RemoteHoursConfig, loc *time.Location) time.Time {
	_, last := PeriodRange(label, c.PeriodStartDay)
	t := time.Date(last.Year(), last.Month(), c.CloseDay, c.CloseHour, 0, 0, 0, loc)
	if t.Before(time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, loc)) {
		t = t.AddDate(0, 1, 0)
	}
	return t
}
