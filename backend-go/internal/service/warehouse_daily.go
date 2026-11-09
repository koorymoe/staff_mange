package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ ملخص المخازن اليومي (طلبات ماتركس 11-04، وافق المالك — إداري الكميات) ═══
// تنبيه واحد مجمّع الصبح لصلاحية المخازن: طلبات الأدوات المعلّقة أكثر من ٣ أيام،
// السيارات الي ما انقيّمت اليوم، والصيانة الي فات موعدها. وماتركس يگدر يجاوب عن الغسيل.

type WarehouseDigest struct {
	StaleRequests int
	OldestDays    int
	ActiveCars    int
	Unrated       int
	WashedToday   int
	Overdue       []string
}

// BuildWarehouseDigest يجمع الأرقام (بلا أسماء موظفين؛ السيارات بأسمائها/لوحاتها — مو بيانات شخصية).
func BuildWarehouseDigest(db *sqlx.DB, overdue func() ([]string, error)) (WarehouseDigest, error) {
	var d WarehouseDigest
	err := db.QueryRow(`SELECT count(*), COALESCE(max(baghdad_today() - baghdad_date("requestedAt")), 0)
		FROM "ToolRequest" WHERE status = 'PENDING' AND "requestedAt" < now() - interval '3 days'`).Scan(&d.StaleRequests, &d.OldestDays)
	if err != nil {
		return d, err
	}
	err = db.QueryRow(`SELECT count(*),
		count(*) FILTER (WHERE NOT EXISTS (SELECT 1 FROM "VehicleDailyRating" r WHERE r."vehicleId" = v.id AND r."ratedDate" = baghdad_today())),
		count(*) FILTER (WHERE EXISTS (SELECT 1 FROM "VehicleWashLog" w WHERE w."vehicleId" = v.id AND w."washDate" = baghdad_today()))
		FROM "Vehicle" v WHERE v."isActive"`).Scan(&d.ActiveCars, &d.Unrated, &d.WashedToday)
	if err != nil {
		return d, err
	}
	if overdue != nil {
		d.Overdue, _ = overdue()
	}
	return d, nil
}

// Message نص التنبيه — فارغ إذا ماكو شي يستاهل.
func (d WarehouseDigest) Message() string {
	parts := []string{}
	if d.StaleRequests > 0 {
		parts = append(parts, fmt.Sprintf("⏳ %d طلب أدوات معلّق أكثر من ٣ أيام (الأقدم %d يوم)", d.StaleRequests, d.OldestDays))
	}
	if d.Unrated > 0 {
		parts = append(parts, fmt.Sprintf("🚗 %d سيارة من %d ما انقيّمت اليوم", d.Unrated, d.ActiveCars))
	}
	if len(d.Overdue) > 0 {
		parts = append(parts, "🔧 صيانة فات موعدها: "+strings.Join(d.Overdue, "، "))
	}
	if len(parts) == 0 {
		return ""
	}
	return "📦 ملخص المخازن اليوم — " + strings.Join(parts, " · ")
}

// WarehouseDaily ينبّه مرة باليوم (بعد ١١ الصبح بغداد — بعد ما يمدي التقييم الصباحي).
type WarehouseDaily struct {
	db      *sqlx.DB
	claim   func(key, day string) (bool, error)
	overdue func() ([]string, error)
	notify  func(msg string)
	doneDay string // يوم انحسم بهاي العملية — حتى ما نعيد البناء كل دورة بعد ما ينرسل
}

func NewWarehouseDaily(db *sqlx.DB, claim func(key, day string) (bool, error), overdue func() ([]string, error), notify func(string)) *WarehouseDaily {
	return &WarehouseDaily{db: db, claim: claim, overdue: overdue, notify: notify}
}

func (w *WarehouseDaily) RunIfDue(now time.Time) error {
	now = now.In(debriefLoc)
	if now.Hour() < 11 {
		return nil
	}
	day := now.Format("2006-01-02")
	if w.doneDay == day {
		return nil
	}
	// (11-07) نبني أول وبعدين نحجز اليوم: چان يحجز أول، فلو فشل البناء (خطأ قاعدة لحظي)
	// يبقى اليوم محجوز والملخص ما يطلع لباچر.
	d, err := BuildWarehouseDigest(w.db, w.overdue)
	if err != nil {
		return err
	}
	ok, err := w.claim("WAREHOUSE_DAILY_SENT", day)
	if err != nil {
		return err
	}
	w.doneDay = day
	if !ok {
		return nil
	}
	if msg := d.Message(); msg != "" && w.notify != nil {
		w.notify(msg)
	}
	return nil
}

// carWashText لأداة ماتركس — أرقام بس.
func carWashText(d WarehouseDigest) string {
	if d.ActiveCars == 0 {
		return "ماكو سيارات فعّالة مسجلة."
	}
	return fmt.Sprintf("اليوم انغسلت %d سيارة من %d فعّالة (باقي %d). وما انقيّمت %d سيارة اليوم.",
		d.WashedToday, d.ActiveCars, d.ActiveCars-d.WashedToday, d.Unrated)
}
