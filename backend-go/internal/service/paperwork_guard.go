package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ═══ منو يسوي ورق الحجز؟ ═══
//
// «اكو حجوزات ما تستدعي فريقاً كاملاً — فقط فني. والفني من يكمل مطلوب
// منه تقرير وفاتورة، واحنه ما نريده يسوي التقرير والفاتورة».
//
// بخدمات مؤشّرة (`managerHandlesPaperwork`)، الورق على **مسؤول الخدمة**
// مو على الفني. وبقية الخدمات ما تنلمس أبداً.
//
// ⚠️⚠️ **الفحص على الخدمة مو على الدور.** دور `SERVICE_MANAGER` لحاله
// يقول «هذا مسؤول خدمة» وما يقول **أي خدمة** — فمسؤول الجي بي اس چان
// يگدر يسوي ورق الداش كام وبالعكس. ولهذا نسأل جدول `ServiceManager`:
// هل هذا الموظف مسؤول عن خدمة **هذا الحجز** بالذات؟
//
// ⚠️⚠️ **وحارس واحد للتقرير والفاتورة.** لو انكتب فحصان منفصلان،
// أول تعديل على القاعدة يوصل واحد وينسى الثاني — فيصير مسؤول الخدمة
// يگدر يسوي التقرير وما يگدر يسوي الفاتورة (أو العكس)، والحجز يعلگ
// نص مكتمل بلا تفسير.

// PaperworkGuard يقرّر منو يحق يسوي ورق حجز معيّن.
type PaperworkGuard struct {
	db *sqlx.DB
}

func NewPaperworkGuard(db *sqlx.DB) *PaperworkGuard {
	return &PaperworkGuard{db: db}
}

// ErrPaperworkNotYours الفني حاول يسوي ورق خدمة ورقها على مسؤولها.
var ErrPaperworkNotYours = errors.New("ورق هذي الخدمة على مسؤول الخدمة — مو عليك")

type bookingPaperworkRow struct {
	ServiceID   sql.NullString `db:"serviceId"`
	ServiceName sql.NullString `db:"serviceName"`
	Managed     bool           `db:"managed"`
}

// Check يرجّع nil لو مسموح، وإلا خطأ برسالة واضحة للموظف.
//
// ⚠️ الحجز الي ما ينلگه أو خدمته غير مؤشّرة **يمر** — لأن هالحارس
// يضيف قاعدة جديدة، ما يشدّد القديمة. أي تشديد عام يكسر موظفاً
// يشتغل هسه بطريقة ما نعرفها.
func (g *PaperworkGuard) Check(bookingID, employeeID, role string) error {
	if bookingID == "" {
		return nil
	}
	// المالك والإداري يمرّون دائماً — هما الي يصلّحون العالق.
	if role == "OWNER" || role == "ADMIN" {
		return nil
	}

	var row bookingPaperworkRow
	err := g.db.Get(&row, `
		SELECT b."serviceId" AS "serviceId",
		       COALESCE(s.name, '') AS "serviceName",
		       COALESCE(s."managerHandlesPaperwork", false) AS managed
		FROM "Booking" b
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.id = $1`, bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // مو شغلة هالحارس — التحقق من وجود الحجز يصير بمكانه
		}
		return err
	}
	if !row.Managed || !row.ServiceID.Valid {
		return nil // ⚠️ خدمة غير مؤشّرة: سلوك اليوم حرفياً
	}

	var count int
	if err := g.db.Get(&count,
		`SELECT COUNT(*) FROM "ServiceManager" WHERE "employeeId" = $1 AND "serviceId" = $2`,
		employeeID, row.ServiceID.String); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	// ⚠️ الرسالة تقول **منو** يسويه مو بس «ممنوع»: الفني الي يقرا
	// «ما عندك صلاحية» يتصل بالإدارة ويضيّع وقتاً — والي يقرا «الورق
	// على مسؤول الخدمة» يعرف شنو يسوي بالضبط.
	return fmt.Errorf("ورق خدمة «%s» على مسؤول الخدمة — مو على الفني", row.ServiceName.String)
}

// IsBookingManager هل هذا الموظف **مسؤول خدمة هذا الحجز**، والخدمة
// مؤشّرة إن ورقها عليه؟
//
// ⚠️⚠️ **الشرطان لازم الاثنان.** لو رجّعنا `true` لمجرد إن الخدمة غير
// مؤشّرة (يعني «ماكو قاعدة تمنعه»)، أي موظف مو ليدر چان صار يگدر يسوي
// فواتير كل الحجوزات العادية — يعني **نفتح ثغرة** بدل ما نضيف قاعدة.
// الفرق بين «ماكو منع» و«اكو سماح» هو كل الفرق هنا.
func (g *PaperworkGuard) IsBookingManager(employeeID, bookingID string) (bool, error) {
	if bookingID == "" || employeeID == "" {
		return false, nil
	}
	var count int
	err := g.db.Get(&count, `
		SELECT COUNT(*)
		FROM "Booking" b
		JOIN "Service" s ON s.id = b."serviceId"
		JOIN "ServiceManager" sm ON sm."serviceId" = s.id
		WHERE b.id = $1
		  AND sm."employeeId" = $2
		  AND s."managerHandlesPaperwork" = true`, bookingID, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return count > 0, err
}

// IsManagerPaperwork هل ورق هالحجز على مسؤول الخدمة؟ (للواجهة والسلال)
func (g *PaperworkGuard) IsManagerPaperwork(bookingID string) (bool, error) {
	var managed bool
	err := g.db.Get(&managed, `
		SELECT COALESCE(s."managerHandlesPaperwork", false)
		FROM "Booking" b LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.id = $1`, bookingID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return managed, err
}
