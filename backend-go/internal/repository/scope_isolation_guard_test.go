package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ═══ حارس: الموظف بلا رؤية شاملة يشوف شغله هو وبس ═══
//
// ثاني اختبار طلبه (ع) بعد تقرير الفحص. والعلّة الي يحرسها **صارت
// فعلاً**: صلاحية «فاتورة شغل داخل الشركة» چانت تعرض **كل** الحجوزات
// الداخلية للشركة لأي موظف يمنحها المالك — و(ع) شافها بعينه وكال
// «ماريد تطلعله كل الحجوزات داخل الشركة، يطلعله فقط الحجز الي توجّه
// اله».
//
// ⚠️ والخطر بالعزل إنه **ينكسر بالصمت**: ماكو رسالة خطأ ولا ٤٠٣ —
// بس قائمة أطول من اللازم، وماكو منو يلاحظ إلا لو دقّق.
func TestInternalBookingsScopedToParty_Live(t *testing.T) {
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

	const mine, other = "scope-emp-a", "scope-emp-b"
	const bkMine, bkOther = "scope-bk-a", "scope-bk-b"
	const cust = "scope-cust"
	clean := func() {
		for _, id := range []string{bkMine, bkOther} {
			db.Exec(`DELETE FROM "Booking" WHERE id = $1`, id)
		}
		for _, id := range []string{mine, other} {
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, id)
		}
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
	}
	clean()
	defer clean()

	for i, id := range []string{mine, other} {
		u := []string{"scope_guard_a", "scope_guard_b"}[i]
		if _, err := db.Exec(`INSERT INTO "Employee" (id, name, username, role, status)
			VALUES ($1, 'موظف حارس العزل', $2, 'TECHNICIAN', 'ACTIVE')`, id, u); err != nil {
			t.Fatalf("موظف %s: %v", id, err)
		}
	}
	// ⚠️ الزبون إلزامي بالجدول حتى للحجز الداخلي — القسم وصاحب الطلب
	// محلّه بالشغل، بس العمود `NOT NULL` بالسكيما.
	if _, err := db.Exec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون حارس العزل', '07700000001')`, cust); err != nil {
		t.Fatalf("زبون: %v", err)
	}

	// حجزان داخليان: واحد مسؤوله «أنا» والثاني مسؤوله غيري.
	mk := func(id, responsible string) {
		if _, err := db.Exec(`INSERT INTO "Booking" (id, code, status, "bookingType", "customerId", "expenseResponsibleId", "createdAt", "updatedAt", "completedAt")
			VALUES ($1, $1, 'COMPLETED', 'INTERNAL', $2, $3, now(), now(), now())`, id, cust, responsible); err != nil {
			t.Fatalf("حجز %s: %v", id, err)
		}
	}
	mk(bkMine, mine)
	mk(bkOther, other)

	// ① محصور بالموظف: يشوف حجزه وبس.
	rows, err := repo.ListInternal("COMPLETED", mine, 0)
	if err != nil {
		t.Fatalf("محصور: %v", err)
	}
	seen := map[string]bool{}
	for _, b := range rows {
		seen[b.ID] = true
	}
	if !seen[bkMine] {
		t.Fatalf("الموظف ما شاف حجزه هو (%s) — الحصر انكلب لمنع، والصلاحية تبقى شكلية", bkMine)
	}
	if seen[bkOther] {
		t.Fatalf("🔴 الموظف شاف حجز غيره (%s) — العزل مكسور، وصلاحية الفاتورة الداخلية صارت باب لبيانات كل الأقسام", bkOther)
	}

	// ② بلا حصر (للي عنده رؤية شاملة): يشوف الاثنين.
	all, err := repo.ListInternal("COMPLETED", "", 0)
	if err != nil {
		t.Fatalf("بلا حصر: %v", err)
	}
	seenAll := map[string]bool{}
	for _, b := range all {
		seenAll[b.ID] = true
	}
	if !seenAll[bkMine] || !seenAll[bkOther] {
		t.Fatalf("صاحب الرؤية الشاملة ما شاف الاثنين — الحصر انطبّق عليه بالغلط (أ=%v ب=%v)", seenAll[bkMine], seenAll[bkOther])
	}
}
