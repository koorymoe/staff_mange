package repository

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ «شنو سوّى هذا الموظف» — الحصيلة بالتفصيل (ع) 10-16 ═══
// «من أدخل على عين وأختار موظف، يطلعلي شنو سوّى خلال اليوم وخلال الشهر —
// بالتفصيل: أي حجز طلع، شنو خدمته، جرد، حضور، تقييمه، فواتير، تقارير…
// والمراقب شنو دقّق». كل قسم استعلام مستقل من الجداول نفسها (أرقام حقيقية)،
// والقسم الفارغ ما يطلع. فشل قسم (جدول مو موجود) ما يكسر الباقي.

type LedgerItem struct {
	At   time.Time `db:"at" json:"at"`
	Text string    `db:"text" json:"text"`
	Link *string   `db:"link" json:"link,omitempty"`
}

type LedgerSection struct {
	Key   string       `json:"key"`
	Title string       `json:"title"`
	Count int          `json:"count"`
	Items []LedgerItem `json:"items"`
}

type EmployeeLedgerRepository struct {
	db *sqlx.DB
}

func NewEmployeeLedgerRepository(db *sqlx.DB) *EmployeeLedgerRepository {
	return &EmployeeLedgerRepository{db: db}
}

// الوصف العربي لمرحلة الحجز.
const ledgerStageAr = `CASE p.stage WHEN 'ACCEPTED' THEN 'استلم' WHEN 'MATERIALS' THEN 'جهّز المواد' WHEN 'DEPARTED' THEN 'انطلق' WHEN 'ARRIVED' THEN 'وصل'
	WHEN 'STARTED' THEN 'بدأ العمل' WHEN 'COMPLETED' THEN 'أنجز' ELSE p.stage END`

const ledgerBookingLabel = `COALESCE(b.code, '') || ' — ' || COALESCE(c.name, 'زبون') || ' — ' || COALESCE(s.name, 'بلا خدمة محددة')`

var ledgerSections = []struct{ key, title, sql string }{
	{"field", "🚗 الحجوزات الي طلع بيها", `
		SELECT MIN(p."pressedAt") AS at,
		       ` + ledgerBookingLabel + ` || ' · ' || string_agg(` + ledgerStageAr + ` || ' ' || to_char(p."pressedAt" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad', 'HH24:MI'), '، ' ORDER BY p."pressedAt") AS text,
		       '/bookings?q=' || b.code AS link
		  FROM "BookingStageProof" p JOIN "Booking" b ON b.id = p."bookingId"
		  LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
		 WHERE p."employeeId" = $1 AND p."pressedAt" >= $2 AND p."pressedAt" < $3
		 GROUP BY b.id, b.code, c.name, s.name`},
	{"materials", "📦 جهّز مواد", `
		SELECT b."materialsReadyAt" AS at, ` + ledgerBookingLabel + ` AS text, '/bookings?q=' || b.code AS link
		  FROM "Booking" b LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
		 WHERE b."materialsReadyById" = $1 AND b."materialsReadyAt" >= $2 AND b."materialsReadyAt" < $3`},
	{"booked", "📝 حجوزات سجّلها", `
		SELECT b."createdAt" AS at, ` + ledgerBookingLabel + ` AS text, '/bookings?q=' || b.code AS link
		  FROM "Booking" b LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
		 WHERE b."createdById" = $1 AND b."createdAt" >= $2 AND b."createdAt" < $3`},
	{"invoices", "🧾 فواتير سوّاها", `
		SELECT li."createdAt" AS at,
		       COALESCE(li."accountingCode", '') || ' — ' || COALESCE(li."customerName", '') || ' — الصافي ' || to_char(li."netTotal", 'FM999,999,999') || ' د.ع'
		       || CASE WHEN li."isFree" THEN ' (مجاني)' ELSE '' END AS text, '/leader-invoices' AS link
		  FROM "LeaderInvoice" li WHERE li."employeeId" = $1 AND li."createdAt" >= $2 AND li."createdAt" < $3`},
	{"reports", "📄 تقارير العمل الي رفعها", `
		SELECT w."createdAt" AS at, ` + ledgerBookingLabel + ` || CASE WHEN w."workStatus" = 'STOPPED' THEN ' · متوقف' ELSE ' · مكتمل' END AS text, '/reports-desk' AS link
		  FROM "WorkReport" w JOIN "Booking" b ON b.id = w."bookingId"
		  LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
		 WHERE w."employeeId" = $1 AND w."createdAt" >= $2 AND w."createdAt" < $3`},
	{"daily", "🏗️ تقارير مشاريع يومية", `
		SELECT r."createdAt" AS at, ` + ledgerBookingLabel + ` || ' · اليوم ' || r."dayNumber" || ' · ' || r."percentDone" || '٪' AS text, '/reports-desk' AS link
		  FROM "BookingProgressReport" r JOIN "Booking" b ON b.id = r."bookingId"
		  LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
		 WHERE r."reportedById" = $1 AND r."createdAt" >= $2 AND r."createdAt" < $3`},
	{"inventory", "📋 الجرد", `
		SELECT i."checkedAt" AS at, CASE WHEN i.complete THEN 'جرد كامل' ELSE 'جرد ناقص: ' || COALESCE(i."missingItems"::text, '') END AS text, '/inventory' AS link
		  FROM "InventoryCheck" i WHERE i."employeeId" = $1 AND i."checkedAt" >= $2 AND i."checkedAt" < $3`},
	{"crewRatings", "⭐ قيّم الكادر", `
		SELECT r."createdAt" AS at, COALESCE(e.name, '') || ' — ' || r.score || '/5' || COALESCE(' — ' || NULLIF(r.note, ''), '') AS text, NULL AS link
		  FROM "CrewRating" r LEFT JOIN "Employee" e ON e.id = r."technicianId"
		 WHERE r."leaderId" = $1 AND r."createdAt" >= $2 AND r."createdAt" < $3`},
	{"ratedBy", "🌟 تقييمه من الليدر", `
		SELECT r."createdAt" AS at, 'قيّمه ' || COALESCE(e.name, '') || ' — ' || r.score || '/5' || COALESCE(' — ' || NULLIF(r.note, ''), '') AS text, NULL AS link
		  FROM "CrewRating" r LEFT JOIN "Employee" e ON e.id = r."leaderId"
		 WHERE r."technicianId" = $1 AND r."createdAt" >= $2 AND r."createdAt" < $3`},
	{"attendance", "🕒 الحضور", `
		SELECT a."checkIn" AS at,
		       'حضور ' || to_char(a."checkIn" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad', 'MM-DD HH24:MI')
		       || COALESCE(' — انصراف ' || to_char(a."checkOut" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad', 'HH24:MI'), ' — بلا انصراف') AS text, '/attendance' AS link
		  FROM "Attendance" a WHERE a."employeeId" = $1 AND a."checkIn" >= $2 AND a."checkIn" < $3`},
	// ── شغل المراقب ──
	{"monitorInbox", "🔎 دقّق بصندوق المراقب", `
		SELECT m."reviewedAt" AS at, m.title || ' — ' || CASE m.status WHEN 'OK' THEN 'سليم' WHEN 'FLAGGED' THEN 'بيها ملاحظة' ELSE m.status END
		       || COALESCE(' — ' || NULLIF(m.note, ''), '') AS text, '/monitor-inbox' AS link
		  FROM "MonitorReview" m WHERE m."reviewedById" = $1 AND m."reviewedAt" >= $2 AND m."reviewedAt" < $3`},
	{"reviewedReports", "✅ دقّق تقارير عمل", `
		SELECT w."reviewedAt" AS at, 'تقرير ' || COALESCE(e.name, '') || ' — ' || ` + ledgerBookingLabel + ledgerReviewNote + ` AS text, '/reports-desk' AS link
		  FROM "WorkReport" w JOIN "Booking" b ON b.id = w."bookingId" LEFT JOIN "Employee" e ON e.id = w."employeeId"
		  LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
		 WHERE w."reviewedById" = $1 AND w."reviewedAt" >= $2 AND w."reviewedAt" < $3`},
	{"reviewedDaily", "✅ دقّق تقارير مشاريع", `
		SELECT r."reviewedAt" AS at, 'تقرير ' || COALESCE(e.name, '') || ' — ' || ` + ledgerBookingLabel + ` || ' · اليوم ' || r."dayNumber" AS text, '/reports-desk' AS link
		  FROM "BookingProgressReport" r JOIN "Booking" b ON b.id = r."bookingId" LEFT JOIN "Employee" e ON e.id = r."reportedById"
		  LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
		 WHERE r."reviewedById" = $1 AND r."reviewedAt" >= $2 AND r."reviewedAt" < $3`},
	{"monitorInvoices", "🧾 قرّر على فواتير (مراقبة)", `
		SELECT li."monitorDecidedAt" AS at, COALESCE(li."accountingCode", '') || ' — ' || COALESCE(li."customerName", '') || COALESCE(' — ' || li."monitorVerdict", '') AS text, '/leader-invoices' AS link
		  FROM "LeaderInvoice" li WHERE li."monitorDecidedById" = $1 AND li."monitorDecidedAt" >= $2 AND li."monitorDecidedAt" < $3`},
	{"justifications", "📝 طلب تفسير", `
		SELECT j."createdAt" AS at, 'من ' || COALESCE(e.name, '') || ': ' || j.question || CASE WHEN j.answer IS NULL THEN ' — بعده ما جاوب' ELSE '' END AS text, '/monitor-inbox' AS link
		  FROM "MonitorJustification" j LEFT JOIN "Employee" e ON e.id = j."employeeId"
		 WHERE j."monitorId" = $1 AND j."createdAt" >= $2 AND j."createdAt" < $3`},
	{"qualityVerdicts", "⭐ قيّم شغل مهندس الجودة", `
		SELECT q."monitorAt" AS at, COALESCE(b.code, '') || ' — ' || CASE q."monitorVerdict" WHEN 'OK' THEN 'تمام' WHEN 'NOT_OK' THEN 'مو تمام' ELSE 'تقصير' END || COALESCE(' — ' || q."monitorNote", '') AS text, '/quality-desk' AS link
		  FROM "QualityFollowUp" q LEFT JOIN "Booking" b ON b.id = q."bookingId"
		 WHERE q."monitorById" = $1 AND q."monitorAt" >= $2 AND q."monitorAt" < $3`},
	// ── المحاسب والجودة ──
	{"audited", "🪙 دقّق فواتير (محاسب)", `
		SELECT li."auditedAt" AS at, COALESCE(li."accountingCode", '') || ' — ' || COALESCE(li."customerName", '') || COALESCE(' — ' || li."auditVerdict", '') AS text, '/leader-invoices' AS link
		  FROM "LeaderInvoice" li WHERE li."auditedById" = $1 AND li."auditedAt" >= $2 AND li."auditedAt" < $3`},
	{"qualityContacts", "📞 تواصل متابعة جودة", `
		SELECT q."contactedAt" AS at, COALESCE(b.code, '') || ' — ' || COALESCE(c.name, '') || COALESCE(' — ' || q."reportType", '') AS text, '/quality-desk' AS link
		  FROM "QualityFollowUp" q LEFT JOIN "Booking" b ON b.id = q."bookingId" LEFT JOIN "Customer" c ON c.id = q."customerId"
		 WHERE q."contactedByEmployeeId" = $1 AND q."contactedAt" >= $2 AND q."contactedAt" < $3`},
}

// ledgerReviewNote ملاحظة تدقيق التقرير إن وجدت.
const ledgerReviewNote = ` || COALESCE(' — ' || NULLIF(w."reviewNote", ''), '')`

// Ledger كل أقسام الحصيلة بين from و to (UTC بلا منطقة، مثل أعمدة القاعدة).
func (r *EmployeeLedgerRepository) Ledger(employeeID string, from, to time.Time) []LedgerSection {
	out := []LedgerSection{}
	for _, s := range ledgerSections {
		items := []LedgerItem{}
		if err := r.db.Select(&items, s.sql+` ORDER BY at DESC LIMIT 300`, employeeID, from, to); err != nil {
			log.Printf("[ledger] %s: %v", s.key, err)
			continue
		}
		if len(items) == 0 {
			continue
		}
		out = append(out, LedgerSection{Key: s.key, Title: s.title, Count: len(items), Items: items})
	}
	return out
}
