package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestPendingPaperworkForEmployee_SurveyExemption_Live يحرس استثناء
// حجز «كشف» (زيارة معاينة) من "منجزة وناقصها ورق" — قرار عمل صريح
// (schema_survey_booking.go): الكشف ما يحتاج فاتورة ولا تقرير عمل
// أصلاً، فبقاؤه بهالطابور يعاقب موظفاً على ورق محد طلبه منه.
func TestPendingPaperworkForEmployee_SurveyExemption_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer db.Close()

	var adminID string
	if err := db.Get(&adminID, `SELECT id FROM "Employee" WHERE status = 'ACTIVE' LIMIT 1`); err != nil {
		t.Skipf("ماكو موظف بالقاعدة: %v", err)
	}
	var customerID string
	if err := db.Get(&customerID, `SELECT id FROM "Customer" LIMIT 1`); err != nil {
		t.Skipf("ماكو زبون بالقاعدة: %v", err)
	}

	const bookingID = "bk_survey_exemption_test"
	const assignID = "ba_survey_exemption_test"
	cleanup := func() {
		db.Exec(`DELETE FROM "BookingAssignment" WHERE id = $1`, assignID)
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bookingID)
	}
	cleanup()
	defer cleanup()

	if _, err := db.Exec(`
		INSERT INTO "Booking" (id, code, "customerId", status, "bookingType", "createdAt", "completedAt")
		VALUES ($1, 'T-SURVEY-TEST', $2, 'COMPLETED', 'SURVEY',
		        now() - interval '10 hours', now() - interval '9 hours')
	`, bookingID, customerID); err != nil {
		t.Fatalf("تعذر إنشاء حجز الاختبار: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO "BookingAssignment" (id, "bookingId", "employeeId", "assignedById", role, "createdAt")
		VALUES ($1, $2, $3, $3, 'TECH_1', now())
	`, assignID, bookingID, adminID); err != nil {
		t.Fatalf("تعذر إنشاء تكليف الاختبار: %v", err)
	}

	repo := NewDisciplineRepository(db)
	rows, err := repo.PendingPaperworkForEmployee(adminID)
	if err != nil {
		t.Fatalf("الاستعلام فشل: %v", err)
	}

	for _, r := range rows {
		if r.BookingID == bookingID {
			t.Errorf("حجز الكشف طلع بطابور «ناقصها ورق» — ما المفروض، الكشف ما يحتاج فاتورة ولا تقرير أصلاً")
		}
	}
}
