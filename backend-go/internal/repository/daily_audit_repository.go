package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// DailyAuditRepository التدقيق اليومي للمحاسب.
//
// الفرق عن التدقيق العام: هذا يمشي *بيوم واحد* — المحاسب يحدد تاريخ،
// يشوف كل حجوزات ذاك اليوم بمبالغها وحالتها، ويعرف من البداية شكد
// المفروض يجمع اليوم قبل ما يبدي.
type DailyAuditRepository struct {
	db *sqlx.DB
}

func NewDailyAuditRepository(db *sqlx.DB) *DailyAuditRepository {
	return &DailyAuditRepository{db: db}
}

// Day يبني تقرير يوم واحد (date بصيغة YYYY-MM-DD).
//
// ⚠️ الحجوزات المنجزة **بس**. الحجز الي ما انجز ماكو منه مبلغ مستلم حتى
// ينتدقق، وظهوره بالقائمة كان يخلي المحاسب يشوف طابور شغل ما يكدر يشتغله.
//
// المبلغ المعتمد لكل حجز: فاتورة الليدر إذا رفعها، وإلا الكلفة
// التقديرية الي حطها الإداري — هذا الي انطلب بالضبط، لأن بلا فاتورة
// ولا تقدير الحجز يطلع صفر ويخرب المجموع.
func (r *DailyAuditRepository) Day(date string) (*model.DailyAuditReport, error) {
	rep := &model.DailyAuditReport{Date: date, Rows: []model.DailyAuditRow{}}

	err := r.db.Select(&rep.Rows, `
		SELECT
			b.id, b.code, b.status::text AS status,
			COALESCE(c.name, '')  AS "customerName",
			COALESCE(c.phone, '') AS "customerPhone",
			COALESCE(s.name, '—') AS "serviceName",
			b."amountVerified",
			COALESCE(b."amountCollected", 0) + COALESCE(b."advancePaid", 0) AS collected,
			COALESCE(b."quotedPrice", 0) AS "quotedPrice",
			li."netTotal" AS "invoiceTotal",
			li."accountingCode" AS "invoiceCode",
			-- المعتمد: فاتورة الليدر أولاً، وإلا تقدير الإداري
			COALESCE(li."netTotal", NULLIF(b."quotedPrice", 0), 0) AS "expectedAmount",
			(SELECT COUNT(*) FROM "BookingAuditIssue" i
			  WHERE i."bookingId" = b.id AND i.status = 'OPEN') AS "openIssues"
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service"  s ON s.id = b."serviceId"
		LEFT JOIN LATERAL (
			SELECT "netTotal", "accountingCode" FROM "LeaderInvoice"
			WHERE "bookingId" = b.id ORDER BY "createdAt" DESC LIMIT 1
		) li ON true
		WHERE b.status = 'COMPLETED'
		  AND baghdad_date(COALESCE(b."completedAt", b."scheduledAt", b."createdAt")) = $1::date
		ORDER BY b."createdAt" DESC`, date)
	if err != nil {
		return nil, err
	}

	for _, row := range rep.Rows {
		// ١) المبالغ: الي انجمعت فعلاً من الحجوزات المكتملة
		if row.Status == "COMPLETED" {
			rep.CollectedTotal += row.Collected
			rep.CompletedCount++
		} else {
			rep.PendingCount++
		}
		// ٢) ما تم تدقيقه، و٣) المدقق
		if row.AmountVerified {
			rep.VerifiedTotal += row.Collected
		} else {
			rep.NotVerifiedTotal += row.Collected
		}
		// ٤) الإجمالي المتوقع لليوم — من الفواتير والتقديرات
		rep.ExpectedTotal += row.ExpectedAmount
		if row.OpenIssues > 0 {
			rep.IssuesCount++
		}
	}
	rep.AllAmountsTotal = rep.VerifiedTotal + rep.NotVerifiedTotal
	return rep, nil
}
