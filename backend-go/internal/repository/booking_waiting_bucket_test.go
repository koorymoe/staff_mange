package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ═══ الحجز الي ما رد صاحبه ينتقل — ما ينتنسخ ═══
//
// «تنتقل نسخة لما وصل للتنفيذ وتبقى نسخة بانتظار التثبيت، ويرجع
// الموظف الغافل يتصل ع الزبون من جديد وما يدري».
//
// الحارس يثبت ثلاثة:
//  ① الحجز الطبيعي يبقى بـ«بانتظار التثبيت» (ما نكسر الطبيعي).
//  ② أول ما ينتأشّر «ما رد» ← **يختفي منها**، ويطلع بسلّته بس.
//  ③ ولمن الزبون يرد ← **يرجع لمحطته لحاله**.
func TestWaitingLeavesPendingBucket_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	const cust = "wbk-cust"
	const bk = "wbk-booking"
	clean := func() {
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bk)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
	}
	clean()
	defer clean()

	if _, err := db.Exec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1,'زبون فحص السلال','07700000000')`, cust); err != nil {
		t.Fatalf("customer: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO "Booking" (id, code, "customerId", status, "scheduledAt")
		VALUES ($1, 'WBK-1', $2, 'PENDING', now())`, bk, cust); err != nil {
		t.Fatalf("booking: %v", err)
	}

	inPending := func() bool {
		var n int
		if err := db.Get(&n, `SELECT COUNT(*) FROM "Booking" b WHERE b.id = $1 AND (`+bucketCondition("pending")+`)`, bk); err != nil {
			t.Fatalf("count: %v", err)
		}
		return n == 1
	}
	// نفس شرط سلّة «ما رد — قبل التثبيت» بـListByStageBucket
	inNoAnswer := func() bool {
		var n int
		if err := db.Get(&n, `SELECT COUNT(*) FROM "Booking"
			WHERE id = $1 AND "archivedAt" IS NULL
			  AND status <> 'CANCELLED' AND "waitingSince" IS NOT NULL AND "confirmedAt" IS NULL`, bk); err != nil {
			t.Fatalf("count: %v", err)
		}
		return n == 1
	}

	// ① الطبيعي يبقى بمحطته
	if !inPending() {
		t.Fatal("الحجز الطبيعي ما طلع بـ«بانتظار التثبيت» — انكسر الطبيعي")
	}

	// ② «الزبون ما رد» ← ينتقل، ما ينتنسخ
	r := NewBookingRepository(db)
	if err := r.MarkWaiting(bk, "اتصلنا وما رد", ""); err != nil {
		t.Fatalf("MarkWaiting: %v", err)
	}
	if inPending() {
		t.Error("⚠️ الحجز بقى بـ«بانتظار التثبيت» بعد تأشير «ما رد» — النسخة الثانية رجعت")
	}
	if !inNoAnswer() {
		t.Error("الحجز ما طلع بسلّة «ما رد» — انضاع بدل ما ينتقل")
	}

	// ③ الزبون رد ← يرجع لمحطته
	if err := r.ResumeFromWaiting(bk); err != nil {
		t.Fatalf("ResumeFromWaiting: %v", err)
	}
	if !inPending() {
		t.Error("الحجز ما رجع لمحطته بعد ما رد الزبون — انحبس بسلّة الانتظار")
	}
	if inNoAnswer() {
		t.Error("الحجز بقى بسلّة «ما رد» بعد رجوعه — نسختان مرة ثانية")
	}
}
