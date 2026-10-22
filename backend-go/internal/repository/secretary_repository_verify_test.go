package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestSecretaryRepository_Live يشغّل كل استعلامات السكرتير على قاعدة حية
// (إذا DATABASE_URL موجود): البحث بالاسم والهاتف، مراحل الحجز، الحضور،
// «ينتظر قرارك»، أرقام اليوم، والمقترح الي ينقفل مرة وحدة.
func TestSecretaryRepository_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	r := NewSecretaryRepository(db)

	const emp, cust, bk = "__sec_emp", "__sec_cust", "__sec_bk"
	cleanup := func() {
		db.Exec(`DELETE FROM "BookingDeleteRequest" WHERE "bookingId" = $1`, bk)
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bk)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "SecretaryAction" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "SecretaryMessage" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status) VALUES ($1, 'سكرتير اختبار', 'ACTIVE')`, emp)
	db.MustExec(`INSERT INTO "Customer"(id, "customerCode", name, phone, "createdAt") VALUES ($1, 987654, 'زبون السكرتير', '07701234567', now())`, cust)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "createdAt", "updatedAt") VALUES ($1, 'SEC-1', $2, 'CONFIRMED', now(), now())`, bk, cust)
	db.MustExec(`INSERT INTO "BookingDeleteRequest"("bookingId", "requestedById", reason) VALUES ($1, $2, 'مكرر')`, bk, emp)

	for _, q := range []string{"زبون السكرتير", "SEC-1", "0770 123 4567"} {
		f, err := r.Find(q)
		if err != nil {
			t.Fatalf("Find(%q): %v", q, err)
		}
		if len(f.Bookings) == 0 || f.Bookings[0].Code != "SEC-1" {
			t.Fatalf("Find(%q) ما لگى الحجز: %+v", q, f.Bookings)
		}
	}
	if f, _ := r.Find("سكرتير اختبار"); len(f.Employees) == 0 {
		t.Fatal("Find ما لگى الموظف")
	}

	b, steps, err := r.BookingTimeline("SEC-1")
	if err != nil || b == nil || len(steps) == 0 || steps[0].Label != "انسجّل" {
		t.Fatalf("BookingTimeline: %v %+v %+v", err, b, steps)
	}
	if b, _, err := r.BookingTimeline("ماكو-هيج-حجز"); err != nil || b != nil {
		t.Fatalf("حجز مو موجود لازم يرجّع nil: %v %+v", err, b)
	}
	if _, err := r.AttendanceToday(); err != nil {
		t.Fatalf("AttendanceToday: %v", err)
	}
	p, err := r.Pending()
	if err != nil {
		t.Fatalf("Pending: %v", err)
	}
	found := false
	for _, x := range p {
		if x.Kind == "delete_request" && strings.Contains(x.Question, "SEC-1") {
			found = true
		}
	}
	if !found {
		t.Fatalf("Pending ما بيه طلب الحذف: %+v", p)
	}
	if _, err := r.Today(); err != nil {
		t.Fatalf("Today: %v", err)
	}

	id, err := r.CreateAction(emp, "task", "طلب", "ملخص", map[string]string{"title": "x"})
	if err != nil {
		t.Fatalf("CreateAction: %v", err)
	}
	if _, err := r.Action(id, "__someone_else"); err == nil {
		t.Fatal("مقترح غيره لازم ما ينقرا")
	}
	if err := r.Decide(id, emp, "SENT", "انرسلت", nil); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if err := r.Decide(id, emp, "SENT", "مرة ثانية", nil); err != ErrSecretaryActionDone {
		t.Fatalf("الإرسال الثاني لازم يرفض: %v", err)
	}
	if _, err := r.SaveMessage(emp, "assistant", "هلا", nil, nil, &id); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	h, err := r.History(emp, 10)
	if err != nil || len(h) != 1 || h[0].ActionStatus == nil || *h[0].ActionStatus != "SENT" {
		t.Fatalf("History: %v %+v", err, h)
	}
}
