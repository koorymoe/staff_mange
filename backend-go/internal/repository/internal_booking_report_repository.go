package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ تقرير حجوزات داخل الشركة الشهري ═══
//
// (ع): «نهاية الشهر احنه نريد إكسل خاص بحجوزات داخل الشركة، يكون بي
// كل الفواتير والمبالغ الي محتاجينها».
//
// ⚠️ **استعلام مكتوب بالإيد مو `SELECT *`**: هذا تقرير يجمع أربعة
// جداول (الحجز · فاتورة الليدر · القسم · الليدر) وينطي **صفاً لكل
// حجز**، ومو صفوف موديل. و`LEFT JOIN` على الفاتورة **بقصد**: الحجز
// الداخلي الي ماكو إله فاتورة لازم يطلع بالملف **بمبلغ فاضي** — هو
// بالضبط الي يريد يشوفه المالك («منو ما سوى فاتورة»). لو خلّيناها
// `JOIN` يختفي الي ناقص وينكتم العيب.
type InternalBookingReportRepository struct {
	db *sqlx.DB
}

func NewInternalBookingReportRepository(db *sqlx.DB) *InternalBookingReportRepository {
	return &InternalBookingReportRepository{db: db}
}

// InternalBookingReportRow صف واحد بالتقرير.
type InternalBookingReportRow struct {
	BookingDate    *time.Time `db:"bookingDate"`
	BookingCode    string     `db:"bookingCode"`
	Department     *string    `db:"department"`
	RequesterName  *string    `db:"requesterName"`
	RequesterPhone *string    `db:"requesterPhone"`
	Services       *string    `db:"services"`
	DeviceCount    *int       `db:"deviceCount"`
	LeaderName     *string    `db:"leaderName"`
	InvoiceCode    *string    `db:"invoiceCode"`
	Work           *string    `db:"work"`
	NetTotal       *float64   `db:"netTotal"`
	InvoiceStatus  *string    `db:"invoiceStatus"`
	ExternalNumber *string    `db:"externalNumber"`
	HrNote         *string    `db:"hrNote"`
}

// Month يرجّع صفوف شهر واحد بصيغة YYYY-MM.
func (r *InternalBookingReportRepository) Month(month string) ([]InternalBookingReportRow, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, fmt.Errorf("الشهر لازم يكون بصيغة YYYY-MM")
	}
	end := start.AddDate(0, 1, 0)

	rows := []InternalBookingReportRow{}
	err = r.db.Select(&rows, `
		SELECT
			b."createdAt"            AS "bookingDate",
			b.code                   AS "bookingCode",
			COALESCE(d.name, b."internalDepartment") AS department,
			b."internalEmployeeName" AS "requesterName",
			b."internalEmployeePhone" AS "requesterPhone",
			(SELECT string_agg(s.name, ' + ' ORDER BY s.name)
			   FROM "BookingService" bs
			   JOIN "Service" s ON s.id = bs."serviceId"
			  WHERE bs."bookingId" = b.id)  AS services,
			b."deviceCount"          AS "deviceCount",
			e.name                   AS "leaderName",
			li."accountingCode"      AS "invoiceCode",
			li."manualWork"          AS work,
			li."netTotal"            AS "netTotal",
			li.status                AS "invoiceStatus",
			li."externalInvoiceNumber" AS "externalNumber",
			b."internalHrNote"       AS "hrNote"
		FROM "Booking" b
		LEFT JOIN "Department" d ON d.id = b."internalDepartmentId"
		LEFT JOIN "LeaderInvoice" li ON li."bookingId" = b.id
		LEFT JOIN "Employee" e ON e.id = li."employeeId"
		WHERE b."bookingType" = 'INTERNAL'
		  AND b."createdAt" >= $1 AND b."createdAt" < $2
		ORDER BY b."createdAt"`, start, end)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
