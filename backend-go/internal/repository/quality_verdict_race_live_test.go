package repository

import (
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
)

// (11-07) حكمين «سلبي» بنفس اللحظة على نفس المتابعة چانوا يخصمون نقطتين من الليدر:
// الفحص يقرا بلا قفل. هسه FOR UPDATE — الثاني ينتظر ويلگه انحسم.
func TestQualityVerdictConcurrentDeductsOnceLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	repo := NewQualityFollowUpRepository(db)
	cust, leader, eng, b, fu := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون جودة', $2)`, cust, "0774"+sfx)
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status, "isLeader") VALUES ($1, 'ليدر جودة', $2, 'TECHNICIAN', 'ACTIVE', true), ($3, 'مهندس جودة', $4, 'QUALITY_ENGINEER', 'ACTIVE', false)`, leader, "ql_"+sfx, eng, "qe_"+sfx)
	db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "projectSupervisorId") VALUES ($1, $2, $3, 'COMPLETED', $4)`, b, "QV"+sfx, cust, leader)
	db.MustExec(`INSERT INTO "QualityFollowUp" (id, "bookingId", "customerId") VALUES ($1, $2, $3)`, fu, b, cust)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "QualityFollowUp" WHERE id = $1`, fu)                               //nolint:errcheck
		db.Exec(`DELETE FROM "KpiEvaluation" WHERE "employeeId" = $1 OR "evaluatorId" = $2`, leader, eng) //nolint:errcheck
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, b)                                        //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)                                    //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, leader, eng)                      //nolint:errcheck
	})
	var wg sync.WaitGroup
	errs := make([]error, 4)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = repo.Verdict(fu, eng, model.QualityVerdictRequest{ReportType: "NEGATIVE", Notes: "فحص تسابق"})
		}(i)
	}
	wg.Wait()
	ok := 0
	for _, e := range errs {
		if e == nil {
			ok++
		}
	}
	var deductions int
	if err := db.Get(&deductions, `SELECT count(*) FROM "KpiEvaluation" WHERE "employeeId" = $1`, leader); err != nil {
		t.Fatal(err)
	}
	if ok != 1 || deductions != 1 {
		t.Fatalf("نجح %d حكم وانخصم %d — المفروض ١ و١", ok, deductions)
	}
}
