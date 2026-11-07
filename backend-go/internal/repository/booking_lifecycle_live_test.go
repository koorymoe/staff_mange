package repository

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
)

// (11-07) حرّاس دورة الحجز — بلا اختبار مباشر چانت. كل حالة هنا قاعدة موجودة بالكود
// (تعليق «فحص» أو شكوى سابقة)، فالاختبار يمسكها إذا تعديل بعدين شالها.
func TestBookingLifecycleGuardsLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	repo := NewBookingRepository(db)
	cust, emp := uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون دورة', $2)`, cust, "0773"+sfx)
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'فني دورة', $2, 'TECHNICIAN', 'ACTIVE')`, emp, "lc_"+sfx)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = $1`, cust) //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)          //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)           //nolint:errcheck
	})
	n := 0
	newBooking := func(status string, crew bool) string {
		n++
		id := uuid.NewString()
		db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status) VALUES ($1, $2, $3, $4::"BookingStatus")`, id, "LC"+sfx+string(rune('a'+n)), cust, status)
		if crew {
			db.MustExec(`UPDATE "Booking" SET "expenseResponsibleId" = $2 WHERE id = $1`, id, emp)
		}
		return id
	}
	status := func(id string) string {
		var s string
		if err := db.Get(&s, `SELECT status::text FROM "Booking" WHERE id = $1`, id); err != nil {
			t.Fatal(err)
		}
		return s
	}

	t.Run("البدء يحتاج كادر", func(t *testing.T) {
		id := newBooking("CONFIRMED", false)
		if err := repo.StartWithResponseTime(id); err == nil {
			t.Fatal("حجز بلا كادر بدا")
		}
		withCrew := newBooking("CONFIRMED", true)
		if err := repo.StartWithResponseTime(withCrew); err != nil {
			t.Fatal(err)
		}
		if s := status(withCrew); s != "IN_PROGRESS" {
			t.Fatalf("الحالة %s", s)
		}
	})

	t.Run("الملغي ما يبدي ولا يكتمل ولا يرجع مثبّت", func(t *testing.T) {
		id := newBooking("CONFIRMED", true)
		if err := repo.Cancel(id, "الزبون ألغى", emp); err != nil {
			t.Fatal(err)
		}
		if err := repo.Cancel(id, "مرة ثانية", emp); err == nil {
			t.Error("إلغاء مرتين انقبل")
		}
		if err := repo.StartWithResponseTime(id); err == nil {
			t.Error("الملغي بدا")
		}
		if err := repo.Complete(id, model.CompleteBookingRequest{}); err == nil {
			t.Error("الملغي اكتمل")
		}
		_ = repo.Confirm(id, model.ConfirmBookingRequest{}, nil)
		if s := status(id); s != "CANCELLED" {
			t.Errorf("التثبيت رجّع الملغي: %s", s)
		}
	})

	t.Run("المنجز ما ينلغي ولا يكتمل مرتين", func(t *testing.T) {
		id := newBooking("IN_PROGRESS", true)
		if err := repo.Complete(id, model.CompleteBookingRequest{}); err != nil {
			t.Fatal(err)
		}
		if err := repo.Complete(id, model.CompleteBookingRequest{}); err == nil {
			t.Error("إنجاز مرتين انقبل")
		}
		if err := repo.Cancel(id, "x", emp); err == nil {
			t.Error("المنجز انلغى")
		}
	})

	t.Run("الأرشيف: مرة وحدة، والمؤرشف ما ينلغي", func(t *testing.T) {
		id := newBooking("PENDING", false)
		if err := repo.Restore(id); err == nil {
			t.Error("استرجاع حجز مو بالأرشيف انقبل")
		}
		if err := repo.Archive(id, emp, "ما رد"); err != nil {
			t.Fatal(err)
		}
		if err := repo.Archive(id, emp, "مرة ثانية"); err == nil {
			t.Error("أرشفة مرتين انقبلت")
		}
		if err := repo.Cancel(id, "x", emp); err == nil {
			t.Error("المؤرشف انلغى")
		}
		if err := repo.Restore(id); err != nil {
			t.Fatal(err)
		}
	})
}
