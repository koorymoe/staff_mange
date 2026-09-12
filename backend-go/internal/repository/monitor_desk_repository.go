package repository

import "github.com/jmoiron/sqlx"

// MonitorDeskRepository عدّادات طوابير مكتب المراقب الأربعة (غير
// صندوق المراقب — عدّه جاهز أصلاً بـ`MonitorReviewRepository.Counts`).
//
// ⚠️ كل تعريف هنا **منسوخ من نفس فلترة الشاشة المستقلة** لا مخترع:
// «عند المراقب» بالفواتير نفس `atMonitor` بـ`LeaderInvoicesListPage.tsx`،
// و«تنسيق الحجوزات» نفس مصدر `bookingHandler.PendingAudit`. رقم
// واحد بتعريف واحد، بدل رقمين يفترقان بأول تعديل.
type MonitorDeskRepository struct {
	db *sqlx.DB
}

func NewMonitorDeskRepository(db *sqlx.DB) *MonitorDeskRepository {
	return &MonitorDeskRepository{db: db}
}

func (r *MonitorDeskRepository) Counts() (issues, invoices, quality, crew int, err error) {
	if err = r.db.Get(&issues, `SELECT COUNT(*) FROM "BookingAuditIssue" WHERE status = 'OPEN'`); err != nil {
		return
	}
	if err = r.db.Get(&invoices, `
		SELECT COUNT(*) FROM "LeaderInvoice"
		WHERE status <> 'APPROVED' AND "monitorRequestedAt" IS NOT NULL AND "monitorDecidedAt" IS NULL`); err != nil {
		return
	}
	if err = r.db.Get(&quality, `SELECT COUNT(*) FROM "QualityFollowUp" WHERE "inspectionStatus" = 'PENDING'`); err != nil {
		return
	}
	if err = r.db.Get(&crew, `SELECT COUNT(*) FROM "Booking" WHERE status = 'PENDING'`); err != nil {
		return
	}
	return
}
