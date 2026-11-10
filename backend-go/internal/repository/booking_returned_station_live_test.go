package repository

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (ع) 11-10 بلاغ #29/#30: الراجع من المشاريع بخانة وحدة، وسجل «منو سوّى شنو» يلگه.
func TestReturnedStationAndHistoryLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	cust, emp, id := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون راجع', $2)`, cust, "0774"+sfx)
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'منسق راجع', $2, 'ADMIN', 'ACTIVE')`, emp, "rt_"+sfx)
	db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "confirmedAt", "transferToProjects") VALUES ($1, $2, $3, 'CONFIRMED', now(), true)`, id, "RT"+sfx, cust)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "EmployeeActivity" WHERE "employeeId" = $1`, emp) //nolint:errcheck
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, id)                     //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)                  //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)                   //nolint:errcheck
	})
	repo := NewBookingRepository(db)
	in := func(bucket string) bool {
		var ok bool
		if err := db.Get(&ok, `SELECT `+BucketCondition(bucket)+` FROM "Booking" b WHERE b.id = $1`, id); err != nil {
			t.Fatal(bucket, err)
		}
		return ok
	}
	note := "ما يحتاج مشروع"
	if err := repo.ReturnToCrew(id, &note, emp); err != nil {
		t.Fatal(err)
	}
	if !in("returned") || in("confirmed") || in("pending") || in("assigned") {
		t.Fatalf("الراجع لازم يكون بخانة «رجعت» بس: returned=%v confirmed=%v", in("returned"), in("confirmed"))
	}
	if ok, err := repo.HandleReturn(id, emp); err != nil || !ok {
		t.Fatal("الاستلام", err)
	}
	if in("returned") || !in("confirmed") {
		t.Fatal("بعد الاستلام لازم يرجع لـ«تم التثبيت»")
	}

	act := NewEmployeeActivityRepository(db)
	act.Record(emp, "PUT", "PUT /api/bookings/{id}/return-handled", "/api/bookings/"+id+"/return-handled", 200)
	act.Record(emp, "PUT", "PUT /api/bookings/{id}/confirm", "/api/bookings/"+id+"/confirm", 400)
	rows, err := act.ForBooking(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].By != "منسق راجع" {
		t.Fatalf("السجل: %+v", rows)
	}
}
