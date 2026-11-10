package repository

import (
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (11-07) تدقيق/إلغاء/تعديل دفعة ما تغيّرت (ملغية أو مدققة من قبل) چان يرجّع «تم».
func TestProjectPaymentUnchangedLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	repo := NewProjectPaymentRepository(db)
	emp, proj, pay := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'محاسب دفعات', $2, 'FINANCE', 'ACTIVE')`, emp, "pp_"+sfx)
	db.MustExec(`INSERT INTO "Project" (id, code, name) VALUES ($1, $2, 'مشروع دفعات')`, proj, "PP"+sfx)
	db.MustExec(`INSERT INTO "ProjectPayment" (id, "projectId", amount, "paidAt") VALUES ($1, $2, 1000000, CURRENT_DATE)`, pay, proj)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Project" WHERE id = $1`, proj) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp) //nolint:errcheck
	})
	if err := repo.Verify(pay, emp); err != nil {
		t.Fatal(err)
	}
	if err := repo.Verify(pay, emp); !errors.Is(err, ErrPaymentUnchanged) {
		t.Errorf("تدقيق ثاني: %v", err)
	}
	if err := repo.Cancel(pay, emp, "تكرار"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Cancel(pay, emp, "مرة ثانية"); !errors.Is(err, ErrPaymentUnchanged) {
		t.Errorf("إلغاء ثاني: %v", err)
	}
	// التعديل نفسه چان يفشل دائماً (خطأ نوع) — نجرّبه على دفعة سليمة أول
	ok := uuid.NewString()
	db.MustExec(`INSERT INTO "ProjectPayment" (id, "projectId", amount, "paidAt") VALUES ($1, $2, 500000, CURRENT_DATE)`, ok, proj)
	if err := repo.SetAmount(ok, emp, 450000, "تصحيح تكرار"); err != nil {
		t.Fatalf("تعديل المبلغ: %v", err)
	}
	var amt float64
	_ = db.Get(&amt, `SELECT amount FROM "ProjectPayment" WHERE id = $1`, ok)
	if amt != 450000 {
		t.Fatalf("المبلغ %v", amt)
	}
	if err := repo.SetAmount(pay, emp, 3000000, "تصحيح"); !errors.Is(err, ErrPaymentUnchanged) {
		t.Errorf("تعديل دفعة ملغية: %v", err)
	}
	if err := repo.Verify(uuid.NewString(), emp); !errors.Is(err, ErrPaymentUnchanged) {
		t.Errorf("دفعة مو موجودة: %v", err)
	}
}
