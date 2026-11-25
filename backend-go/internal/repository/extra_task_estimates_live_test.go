package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
)

// (ل) 11-25 دازاي: مهمة سريعة بوقت لازم وتكلفة تقديرية.
func TestExtraTaskEstimatesLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" = 'xe-e'`)
	db.Exec(`DELETE FROM "Employee" WHERE id = 'xe-e'`)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" = 'xe-e'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'xe-e'`)
	})
	db.MustExec(`INSERT INTO "Employee"(id, name, role, status) VALUES ('xe-e', 'xe', 'TECHNICIAN', 'ACTIVE')`)
	r := NewExtraTaskRepository(db)
	m, c := 30, 5000.0
	due := "2099-01-01T10:00:00Z"
	row, err := r.Create(model.CreateExtraTaskRequest{Title: "شد يدة باب", AssignedToID: "xe-e", Priority: "NORMAL", DueAt: &due, EstimatedMinutes: &m, EstimatedCost: &c}, "xe-e")
	if err != nil {
		t.Fatal(err)
	}
	if row.EstimatedMinutes == nil || *row.EstimatedMinutes != 30 || row.EstimatedCost == nil || *row.EstimatedCost != 5000 {
		t.Fatalf("التقديرات ما انحفظت: %+v", row)
	}
	bad := -1.0
	if _, err := r.Create(model.CreateExtraTaskRequest{Title: "x", AssignedToID: "xe-e", Priority: "NORMAL", DueAt: &due, EstimatedCost: &bad}, "xe-e"); err == nil {
		t.Fatal("تكلفة سالبة لازم تنرفض")
	}
}
