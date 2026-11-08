package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

func TestValidSalary(t *testing.T) {
	ok, bad := 600000.0, 600.0
	if ValidSalary(&ok) != nil || ValidSalary(nil) != nil {
		t.Fatal("الراتب الكامل لازم ينقبل")
	}
	if err := ValidSalary(&bad); err == nil {
		t.Fatal("600 لازم ينرفض (مكتوب بالآلاف)")
	}
}

// (ع) 11-08: العقد الي ينتهي خلال شهر ينبّه مرة وحدة، والي بلا تاريخ ينبّه أسبوعياً.
func TestContractWatch_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	emps := []string{"__cw_soon", "__cw_far", "__cw_none"}
	clean := func() {
		db.Exec(`DELETE FROM "ContractReminder"`)                           //nolint:errcheck
		db.Exec(`DELETE FROM "Notification" WHERE type = 'CONTRACT_WATCH'`) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, emps)          //nolint:errcheck
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role, "contractEndDate") VALUES
		('__cw_soon','عقده قريب','ACTIVE','TECHNICIAN', now() + interval '10 days'),
		('__cw_far','عقده بعيد','ACTIVE','TECHNICIAN', now() + interval '200 days'),
		('__cw_none','بلا عقد','ACTIVE','TECHNICIAN', NULL)`)
	n := repository.NewNotificationRepository(db)
	exp, miss := ContractWatchOnce(db, n, time.Now())
	if exp < 1 || miss < 1 {
		t.Fatalf("لازم ينبّه: قريب=%d بلا=%d", exp, miss)
	}
	var soon int
	db.Get(&soon, `SELECT count(*) FROM "ContractReminder" WHERE "employeeId" = '__cw_soon'`) //nolint:errcheck
	var far int
	db.Get(&far, `SELECT count(*) FROM "ContractReminder" WHERE "employeeId" = '__cw_far'`) //nolint:errcheck
	if soon != 1 || far != 0 {
		t.Fatalf("القريب بس: %d %d", soon, far)
	}
	if exp2, miss2 := ContractWatchOnce(db, n, time.Now()); exp2 != 0 || miss2 != 0 {
		t.Fatal("ما لازم ينبّه مرتين")
	}
}
