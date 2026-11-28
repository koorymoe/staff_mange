package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (11-28) سجل الإجازات على شهر الشركة: إجازة يوم ٢٦ تطلع بالشهر الجاي، ويوم ٢٥ بالشهر نفسه.
func TestLeaveRegisterPeriodLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	db.Exec(`DELETE FROM "LeaveRequest" WHERE "employeeId" = 'lr-e'`)
	db.Exec(`DELETE FROM "Employee" WHERE id = 'lr-e'`)
	db.MustExec(`INSERT INTO "Employee"(id, name, username, password, role, status) VALUES ('lr-e','lr','lr-e','x','TECHNICIAN','ACTIVE')`)
	defer db.Exec(`DELETE FROM "Employee" WHERE id = 'lr-e'`)
	defer db.Exec(`DELETE FROM "LeaveRequest" WHERE "employeeId" = 'lr-e'`)
	var cols []string
	db.Select(&cols, `SELECT column_name FROM information_schema.columns WHERE table_name = 'LeaveRequest' AND is_nullable = 'NO' AND column_default IS NULL`)
	_, err := db.Exec(`INSERT INTO "LeaveRequest"(id, "employeeId", "startDate", "endDate", kind, status, reason, route) VALUES
		('lr-25','lr-e','2031-11-25','2031-11-25','REGULAR','APPROVED','x','ADMIN'), ('lr-26','lr-e','2031-11-26','2031-11-26','REGULAR','APPROVED','x','ADMIN')`)
	if err != nil {
		t.Skipf("جدول الإجازات يحتاج أعمدة ثانية %v: %v", cols, err)
	}
	r := NewLeaveRequestRepository(db)
	has := func(month, id string) bool {
		rows, err := r.Register(month)
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range rows {
			if x.ID == id {
				return true
			}
		}
		return false
	}
	if !has("2031-11", "lr-25") || has("2031-11", "lr-26") || !has("2031-12", "lr-26") || has("2031-12", "lr-25") {
		t.Fatal("حدود الفترة غلط")
	}
}
