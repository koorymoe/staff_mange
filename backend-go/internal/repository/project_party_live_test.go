package repository

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (ون بيس 11-10) طرف المشروع: المسؤول أو مشرف حجزه — والغريب لا.
func TestProjectIsPartyLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	sfx := uuid.NewString()[:6]
	resp, sup, other, cust, bk, proj := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	for i, e := range []string{resp, sup, other} {
		db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'طرف', $2, 'TECHNICIAN', 'ACTIVE')`, e, "pp"+sfx+string(rune('a'+i)))
	}
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون طرف', $2)`, cust, "0775"+sfx)
	db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", "projectSupervisorId") VALUES ($1, $2, $3, $4)`, bk, "PB"+sfx, cust, sup)
	db.MustExec(`INSERT INTO "Project" (id, code, name, "responsibleEmployeeId", "bookingId") VALUES ($1, $2, 'مشروع طرف', $3, $4)`, proj, "PX"+sfx, resp, bk)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Project" WHERE id = $1`, proj)                         //nolint:errcheck
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bk)                           //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)                        //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2, $3)`, resp, sup, other) //nolint:errcheck
	})
	r := NewProjectRepository(db)
	if !r.IsParty(proj, resp) || !r.IsParty(proj, sup) {
		t.Fatal("المسؤول ومشرف الحجز لازم يكونون أطراف")
	}
	if r.IsParty(proj, other) || r.IsParty("nope", resp) {
		t.Fatal("الغريب أو مشروع مو موجود — مو طرف")
	}
}
