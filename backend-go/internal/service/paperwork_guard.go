package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	if ok, err := g.holdsServicePaperworkPermission(bookingID, employeeID); err != nil {
		return err
	} else if ok {
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
	// (دازاي 10-26) الليدر السابق ويا صلاحية «إكمال ورق حجوزاته القديمة»: يسوي
	// فاتورة حجز اشتغل بيه هو وبعده ناقص ورق — مو أي حجز.
	if ok, _ := g.CarryoverAllowed(employeeID, bookingID); ok {
		return true, nil
	}
	// (دازاي 10-28) الحجز المرحّل لإدارة التقنيين: التقني يكتب فاتورته، وإذا طلب
	// كادر فليدر الكادر — مو أي ليدر ومو أي تقني. الي انحل بالتلفون بلا فاتورة.
	var hv int
	_ = g.db.Get(&hv, `SELECT count(*) FROM "Booking" b WHERE b.id = $1 AND `+handoverInvoicerSQL+` = $2 AND `+handoverNeedsInvoiceSQL, bookingID, employeeID)
	if hv > 0 {
		return true, nil
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

// holdsServicePaperworkPermission صاحب «فواتير وتقارير الجي بي اس/
// الداش كام» يسوي ورق حجوزات نوع خدمته — (ع): «من انطيها للموظف يكدر
// يسوي تقرير وفاتورة». نفس المفتاح القديم، فالي عندهم الصلاحية أصلاً
// تشتغل عندهم تلقائياً بلا ما ينمنحون من جديد.
//
// ⚠️ مطابقة بالنوع: صاحب الجي بي اس ما يسوي ورق داش كام وبالعكس.
func (g *PaperworkGuard) holdsServicePaperworkPermission(bookingID, employeeID string) (bool, error) {
	var ok bool
	err := g.db.Get(&ok, `
		SELECT EXISTS (
			SELECT 1
			FROM "Service" s
			JOIN "Permission" p ON p.name = CASE s."serviceKind"
				WHEN 'GPS' THEN 'invoice_gps'
				WHEN 'DASHCAM' THEN 'invoice_dashcam'
			END
			JOIN "EmployeePermission" ep ON ep."permissionId" = p.id AND ep."employeeId" = $2
			WHERE s.id IN (
				SELECT "serviceId" FROM "Booking" WHERE id = $1 AND "serviceId" IS NOT NULL
				UNION
				SELECT "serviceId" FROM "BookingService" WHERE "bookingId" = $1
			)
		)`, bookingID, employeeID)
	return ok, err
}

// handoverInvoicerSQL كاتب فاتورة الحجز المرحّل (دازاي 10-28): ليدر الكادر إذا
// التقني طلب كادر وانعيّن إله ليدر، وإلا التقني نفسه.
const handoverInvoicerSQL = `(CASE WHEN b."techCrewRequestedAt" IS NOT NULL AND booking_leader_id(b.id) IS NOT NULL
	THEN booking_leader_id(b.id) ELSE b."handoverToId" END)`

// handoverNeedsInvoiceSQL حجز مرحّل يحتاج فاتورة: مو منحل بالتلفون.
const handoverNeedsInvoiceSQL = `b."handoverToId" IS NOT NULL AND COALESCE(b."techDecision", '') <> 'PHONE'`

// HandoverMissing حجوزات الموظف المرحّلة المنجزة الي بعدها بلا فاتورة.
func (g *PaperworkGuard) HandoverMissing(employeeID string) ([]string, error) {
	ids := []string{}
	err := g.db.Select(&ids, `SELECT b.id FROM "Booking" b
		WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL AND b."settledLegacyAt" IS NULL
		  AND `+handoverNeedsInvoiceSQL+` AND `+handoverInvoicerSQL+` = $1
		  AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)
		ORDER BY b."completedAt" DESC LIMIT 200`, employeeID)
	return ids, err
}

// carryoverMissingSQL حجوزات الموظف المنجزة الي ناقصها فاتورة أو تقرير.
const carryoverMissingSQL = `FROM "Booking" b
	WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL
	  AND (EXISTS (SELECT 1 FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id AND ba."employeeId" = $1)
	       OR b."projectSupervisorId" = $1)
	  AND (NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)
	       OR NOT EXISTS (SELECT 1 FROM "WorkReport" w WHERE w."bookingId" = b.id))`

func (g *PaperworkGuard) holdsCarryover(employeeID string) bool {
	var n int
	_ = g.db.Get(&n, `SELECT count(*) FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
		WHERE ep."employeeId" = $1 AND p.name = 'paperwork_carryover'`, employeeID)
	return n > 0
}

// CarryoverAllowed الموظف عنده الصلاحية، والحجز من حجوزاته وناقصه ورق.
func (g *PaperworkGuard) CarryoverAllowed(employeeID, bookingID string) (bool, error) {
	if !g.holdsCarryover(employeeID) {
		return false, nil
	}
	var n int
	err := g.db.Get(&n, `SELECT count(*) `+carryoverMissingSQL+` AND b.id = $2`, employeeID, bookingID)
	return n > 0, err
}

// CarryoverItem حجز قديم ناقصه ورق.
type CarryoverItem struct {
	ID          string  `db:"id" json:"id"`
	Code        string  `db:"code" json:"code"`
	CompletedAt *string `db:"completedAt" json:"completedAt"`
	HasInvoice  bool    `db:"hasInvoice" json:"hasInvoice"`
	HasReport   bool    `db:"hasReport" json:"hasReport"`
}

// CarryoverList حجوزاته الناقص ورقها — فاضية إذا ما عنده الصلاحية.
func (g *PaperworkGuard) CarryoverList(employeeID string) ([]CarryoverItem, error) {
	out := []CarryoverItem{}
	if !g.holdsCarryover(employeeID) {
		return out, nil
	}
	err := g.db.Select(&out, `SELECT b.id, b.code, to_char(baghdad_date(b."completedAt"), 'YYYY-MM-DD') AS "completedAt",
		EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id) AS "hasInvoice",
		EXISTS (SELECT 1 FROM "WorkReport" w WHERE w."bookingId" = b.id) AS "hasReport"
		`+carryoverMissingSQL+` ORDER BY b."completedAt"`, employeeID)
	return out, err
}

// SweepCarryover يشيل الصلاحية عن الي خلّص كل ورقه، ويرجّع أسماءهم للتبليغ.
func (g *PaperworkGuard) SweepCarryover() ([]struct{ ID, Name string }, error) {
	var done []struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	err := g.db.Select(&done, `SELECT e.id, e.name FROM "Employee" e
		JOIN "EmployeePermission" ep ON ep."employeeId" = e.id
		JOIN "Permission" p ON p.id = ep."permissionId" AND p.name = 'paperwork_carryover'
		WHERE NOT EXISTS (SELECT 1 `+strings.Replace(carryoverMissingSQL, "$1", "e.id", -1)+`)`)
	if err != nil {
		return nil, err
	}
	out := []struct{ ID, Name string }{}
	for _, d := range done {
		if _, err := g.db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1
			AND "permissionId" = (SELECT id FROM "Permission" WHERE name = 'paperwork_carryover')`, d.ID); err == nil {
			out = append(out, struct{ ID, Name string }{d.ID, d.Name})
		}
	}
	return out, nil
}
