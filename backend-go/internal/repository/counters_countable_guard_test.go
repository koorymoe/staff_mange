package repository

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ═══ حارس: «الحجز الي ما ينعدّ» يبقى غير معدود ═══
//
// (ع) بعد ما قرا تقرير الفحص: «لو ما تسوي غير شي واحد — اكتب ثلاثة
// اختبارات». وهذا أولها وأهمها، لأن العلّة الي يحرسها **رجعت مرتين**:
// انصلّحت بـ`0495fa9` بمصدر واحد، وانكنست بـ`848d5d5`، وبقت **١٣
// موضعاً** بعائلة عدّادات إنتاجية الموظف لحد `58172dc`.
//
// والفاحص سمّى السبب بدقة: «ماكو اختبار يمسك الجولة ما كملت».
// فهذا الملف يمسكها بطريقتين — بالبنية وبالبيانات.

// ⚠️ **الاستثناءان الوحيدان**: فحصا **صلاحية** مو عدّادان. الفني
// المكلَّف بحجز مطلوب حذفه لسه يحتاج يوصله، وترشيحهما **يقفل باب
// موظف شغّال**. أي إضافة لهاي القائمة لازم تكون قراراً مكتوباً.
var countableExempt = map[string]string{
	"IsAssignedTo":              "فحص صلاحية — الفني المكلَّف يوصل حجزه حتى لو مطلوب حذفه",
	"IsCartItemOfAssignedBooking": "فحص صلاحية — نفس السبب",
}

// TestEveryBookingCounterIsScoped يقرا المستودع نفسه ويتأكد إن كل
// دالة تعدّ حجوزات تنادي شرط النطاق.
//
// ⚠️ **اختبار بنيوي بلا قاعدة بيانات**: يشتغل بكل تشغيل للاختبارات،
// ويفشل **يوم ينكتب العدّاد** مو يوم يشتكي المالك من رقم غلط.
func TestEveryBookingCounterIsScoped(t *testing.T) {
	src, err := os.ReadFile("booking_repository.go")
	if err != nil {
		t.Fatalf("قراءة المستودع: %v", err)
	}
	// كل دالة على `*BookingRepository`، من توقيعها لتوقيع الي بعدها.
	parts := regexp.MustCompile(`\nfunc \(r \*BookingRepository\) `).Split(string(src), -1)
	var missing []string
	for _, p := range parts[1:] {
		name := p[:strings.IndexByte(p, '(')]
		if _, ok := countableExempt[name]; ok {
			continue
		}
		touchesBooking := strings.Contains(p, `FROM "Booking"`) || strings.Contains(p, `"Booking" b`)
		counts := strings.Contains(p, "COUNT(") || strings.Contains(p, "SUM(")
		if !touchesBooking || !counts {
			continue
		}
		scoped := strings.Contains(p, "BookingCountable") ||
			strings.Contains(p, "NotDeletePendingSQL") ||
			strings.Contains(p, "BookingDeletePendingSQL")
		if !scoped {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("عدّادات حجوزات بلا شرط «الي ينعدّ» — المؤرشف والمطلوب حذفه راح ينعدّون بيها:\n  · %s\n\nالعلاج: ضيف BookingCountableAndSQL(`b`) لآخر شرط WHERE. وإذا الدالة **فحص صلاحية** مو عدّاد، ضيفها لـcountableExempt وياها السبب.",
			strings.Join(missing, "\n  · "))
	}
}

// TestArchivedAndDeletePendingNotCounted_Live نفس القاعدة بالبيانات:
// حجز مؤرشف وحجز مطلوب حذفه ما ينعدّون بإنتاجية الموظف.
//
// ⚠️ والاثنان **حالتان مختلفتان**: المؤرشف مؤشَّر بعمود، والمطلوب
// حذفه بصف بجدول ثاني — وانصلّحت وحدة وبقت الثانية بجولة سابقة.
func TestArchivedAndDeletePendingNotCounted_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	repo := NewBookingRepository(db)

	const emp = "cnt-emp"
	const cust = "cnt-cust"
	ids := []string{"cnt-ok", "cnt-arch", "cnt-del"}
	clean := func() {
		db.Exec(`DELETE FROM "BookingDeleteRequest" WHERE "bookingId" = ANY($1)`, "{cnt-ok,cnt-arch,cnt-del}")
		for _, id := range ids {
			db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" = $1`, id)
			db.Exec(`DELETE FROM "Booking" WHERE id = $1`, id)
		}
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	clean()
	defer clean()

	if _, err := db.Exec(`INSERT INTO "Employee" (id, name, username, role, status)
		VALUES ($1, 'موظف حارس العدّاد', 'cnt_guard_emp', 'TECHNICIAN', 'ACTIVE')`, emp); err != nil {
		t.Fatalf("موظف: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون حارس العدّاد', '07700000000')`, cust); err != nil {
		t.Fatalf("زبون: %v", err)
	}
	var role string
	if err := db.Get(&role, `SELECT unnest(enum_range(NULL::"TechnicianRole")) LIMIT 1`); err != nil {
		t.Fatalf("دور فني: %v", err)
	}
	mk := func(id string, archived bool) {
		arch := "NULL"
		if archived {
			arch = "now()"
		}
		if _, err := db.Exec(`INSERT INTO "Booking" (id, code, status, "customerId", "createdAt", "updatedAt", "completedAt", "archivedAt")
			VALUES ($1, $1, 'COMPLETED', $2, now(), now(), now(), `+arch+`)`, id, cust); err != nil {
			t.Fatalf("حجز %s: %v", id, err)
		}
		if _, err := db.Exec(`INSERT INTO "BookingAssignment" (id, "bookingId", "employeeId", role)
			VALUES ($1, $2, $3, $4::"TechnicianRole")`, "a-"+id, id, emp, role); err != nil {
			t.Fatalf("تعيين %s: %v", id, err)
		}
	}
	mk("cnt-ok", false)
	mk("cnt-arch", true)
	mk("cnt-del", false)
	if _, err := db.Exec(`INSERT INTO "BookingDeleteRequest" (id, "bookingId", "requestedById", reason, status)
		VALUES ('dr-cnt', 'cnt-del', $1, 'حارس اختبار', 'PENDING')`, emp); err != nil {
		t.Fatalf("طلب حذف: %v", err)
	}

	var month string
	if err := db.Get(&month, `SELECT to_char(now(), 'YYYY-MM')`); err != nil {
		t.Fatalf("الشهر: %v", err)
	}
	got, err := repo.CountAssignedForEmployeeMonth(emp, month)
	if err != nil {
		t.Fatalf("العدّاد: %v", err)
	}
	if got != 1 {
		t.Fatalf("الحجوزات المكلَّفة: انتظرنا ١ (الطبيعي وحده) ولگينا %d — يعني المؤرشف أو المطلوب حذفه لسه ينعدّ بإنتاجية الموظف، وهاي العدّادات تغذّي الغرامات", got)
	}
}
