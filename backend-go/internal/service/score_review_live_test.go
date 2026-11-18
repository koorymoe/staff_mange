package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type nopNotif struct{}

func (nopNotif) Create(string, string, string) error { return nil }

// (دازاي 11-18) تقييم الليدر السيء يأكده مدير إدارة الفني (أو نائبه) بس، والغريب 404.
func TestScoreReviewLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	ids := []string{"sr_tech", "sr_lead", "sr_dir", "sr_out"}
	cleanup := func() {
		db.Exec(`DELETE FROM "WorkplaceIssue" WHERE "partyAId" = ANY($1)`, ids)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = ANY($1)`, ids)
		db.Exec(`DELETE FROM "OrgDivision" WHERE id = 'sr_div'`)
		db.Exec(`DELETE FROM "Booking" WHERE id = 'sr_bk'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'sr_cus'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, ids)
	}
	cleanup()
	t.Cleanup(func() { cleanup(); db.Close() })
	for _, id := range ids {
		db.MustExec(`INSERT INTO "Employee" (id, name, username, password, role, status) VALUES ($1, $1, $1, 'x', 'TECHNICIAN', 'ACTIVE')`, id)
	}
	db.MustExec(`INSERT INTO "OrgDivision" (id, name) VALUES ('sr_div', 'إدارة اختبار')`)
	db.MustExec(`INSERT INTO "OrgAssignment" ("employeeId", "divisionId", position) VALUES ('sr_dir','sr_div','DIRECTOR'), ('sr_tech','sr_div','MEMBER')`)
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ('sr_cus', 'زبون', '07000000999')`)
	db.MustExec(`INSERT INTO "Booking" (id, "customerId", code, "updatedAt") VALUES ('sr_bk', 'sr_cus', 'SR1', now())`)
	db.MustExec(`INSERT INTO "CrewRating" (id, "bookingId", "leaderId", "technicianId", score) VALUES ('sr_cr', 'sr_bk', 'sr_lead', 'sr_tech', 1)`)
	db.MustExec(`INSERT INTO "WorkplaceIssue" (id, title, "partyAId", "partyBId", description) VALUES ('sr_wi', 'خلاف', 'sr_tech', 'sr_lead', 'تفاصيل')`)

	svc := NewScoreReviewService(db, nopNotif{})
	if q, _ := svc.Queue("sr_out", "TECHNICIAN"); len(q) != 0 {
		t.Fatalf("outsider sees %d", len(q))
	}
	q, err := svc.Queue("sr_dir", "TECHNICIAN")
	if err != nil || len(q) != 2 {
		t.Fatalf("director queue = %d, %v", len(q), err)
	}
	if err := svc.DecideRating("sr_out", "TECHNICIAN", "sr_cr", true); err != ErrReviewHidden {
		t.Fatalf("outsider decided: %v", err)
	}
	if err := svc.DecideRating("sr_dir", "TECHNICIAN", "sr_cr", false); err != nil {
		t.Fatal(err)
	}
	if err := svc.DecideIssue("sr_dir", "TECHNICIAN", "sr_wi", "sr_out"); err != ErrReviewHidden {
		t.Fatal("fault on a non-party accepted")
	}
	if err := svc.DecideIssue("sr_dir", "TECHNICIAN", "sr_wi", "sr_lead"); err != nil {
		t.Fatal(err)
	}
	var st string
	db.Get(&st, `SELECT "reviewStatus" FROM "CrewRating" WHERE id = 'sr_cr'`)
	var fault string
	db.Get(&fault, `SELECT "faultEmployeeId" FROM "WorkplaceIssue" WHERE id = 'sr_wi'`)
	if st != "REJECTED" || fault != "sr_lead" {
		t.Fatalf("st=%s fault=%s", st, fault)
	}
}
