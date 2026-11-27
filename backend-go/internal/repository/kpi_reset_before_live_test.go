package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (11-27) التصفير يرجّع بس خصومات الفترة الي خلصت — خصم يوم ٢٦ يبقى.
func TestKpiResetBeforeLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	var emp string
	if err := db.Get(&emp, `SELECT id FROM "Employee" LIMIT 1`); err != nil {
		t.Skip("ماكو موظف")
	}
	db.MustExec(`DELETE FROM "KpiEvaluation" WHERE reason = 'rb-test'`)
	defer db.Exec(`DELETE FROM "KpiEvaluation" WHERE reason = 'rb-test'`)
	cut := time.Now().Add(-time.Hour)
	db.MustExec(`INSERT INTO "KpiEvaluation" (id, "employeeId", "evaluatorId", points, reason, "deductionAmount", "createdAt") VALUES
		('rb-old', $1, $1, 3, 'rb-test', 0, $2), ('rb-new', $1, $1, 2, 'rb-test', 0, $3)`, emp, cut.Add(-48*time.Hour), time.Now())
	if _, _, err := NewKpiRepository(db).ResetBefore(cut); err != nil {
		t.Fatal(err)
	}
	var oldC, newC bool
	db.Get(&oldC, `SELECT cancelled FROM "KpiEvaluation" WHERE id = 'rb-old'`)
	db.Get(&newC, `SELECT cancelled FROM "KpiEvaluation" WHERE id = 'rb-new'`)
	if !oldC || newC {
		t.Fatalf("old=%v new=%v", oldC, newC)
	}
}
