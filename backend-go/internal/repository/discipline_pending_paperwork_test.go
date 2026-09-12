package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestPendingPaperworkForEmployee_NoLeader_Live يحرس علّة حقيقية صارت
// فعلاً بالتطوير: الحجز الي ماكو إله **ليدر مسجّل** جان يخلّي عمود
// `asLeader` يرجع NULL (لأن `COALESCE(...) = $1` بين NULL وقيمة =
// NULL مو false)، فيفشل الاستعلام كله بـScan error.
//
// ⚠️⚠️ والأخطر إن الفشل جان **صامت**: خدمة الكيان تسجّل الخطأ
// بالسجل وترجّع صفر سطور، فالموظف يشوف كيان يكله «شغلك نظيف»
// وهو عليه ورق متأخر وغرامة جاي عليه. تحذير ناقص أسوأ من ماكو
// تحذير — لأنه يطمّن الموظف غلط.
//
// الاختبار ينشئ حجزاً منجزاً بلا فاتورة ولا تقرير وبلا أي ليدر،
// ويتأكد إن الاستعلام **ما يفشل** ويرجّع الصف بـ`asLeader=false`.
func TestPendingPaperworkForEmployee_NoLeader_Live(t *testing.T) {
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

	const bookingID = "bk_pending_paperwork_test"
	const assignID = "ba_pending_paperwork_test"
	cleanup := func() {
		db.Exec(`DELETE FROM "BookingAssignment" WHERE id = $1`, assignID)
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bookingID)
	}
	cleanup()
	defer cleanup()

	// الحجز ينجز بعد تاريخ تشغيل الغرامات حتى يدخل بالنطاق أصلاً.
	if _, err := db.Exec(`
		INSERT INTO "Booking" (id, code, "customerId", status, "createdAt", "completedAt")
		VALUES ($1, 'T-PPW-TEST', $2, 'COMPLETED',
		        now() - interval '10 hours', now() - interval '9 hours')
	`, bookingID, customerID); err != nil {
		t.Fatalf("تعذر إنشاء حجز الاختبار: %v", err)
	}
	// ⚠️ الموظف هنا **الإداري الي كلّف** بس — وماكو ولا ليدر بالحجز،
	// وهاي بالضبط الحالة الي جانت تكسر الاستعلام.
	if _, err := db.Exec(`
		INSERT INTO "BookingAssignment" (id, "bookingId", "employeeId", "assignedById", role, "createdAt")
		VALUES ($1, $2, $3, $3, 'TECH_1', now())
	`, assignID, bookingID, adminID); err != nil {
		t.Fatalf("تعذر إنشاء تكليف الاختبار: %v", err)
	}

	repo := NewDisciplineRepository(db)
	rows, err := repo.PendingPaperworkForEmployee(adminID)
	if err != nil {
		t.Fatalf("الاستعلام فشل (هاي بالضبط العلّة الي نحرسها): %v", err)
	}

	found := false
	for _, r := range rows {
		if r.BookingID != bookingID {
			continue
		}
		found = true
		if r.AsLeader {
			t.Errorf("asLeader لازم تكون false — ماكو ليدر بهذا الحجز أصلاً")
		}
		if r.HasInvoice || r.HasReport {
			t.Errorf("الحجز بلا فاتورة ولا تقرير، بس رجع hasInvoice=%v hasReport=%v", r.HasInvoice, r.HasReport)
		}
	}
	if !found {
		t.Errorf("الحجز الناقص ورقه ما رجع بالنتيجة — الموظف راح يشوف «شغلك نظيف» وهو عليه ورق")
	}
}
