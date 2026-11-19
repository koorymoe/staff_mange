package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// بلاغ #53 (11-19): تعديل رقم الزبون لرقم مسجّل لزبون ثاني ← رسالة واضحة، مو خطأ قاعدة.
func TestCustomerUpdateDuplicatePhoneLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	cleanup := func() { db.Exec(`DELETE FROM "Customer" WHERE id IN ('dup_a', 'dup_b')`) }
	cleanup()
	t.Cleanup(func() { cleanup(); db.Close() })
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('dup_a', 'زبون أ', '07700000881'), ('dup_b', 'زبون ب', '07700000882')`)
	svc := NewCustomerService(repository.NewCustomerRepository(db))
	_, err := svc.Update("dup_a", model.UpdateCustomerRequest{Name: "زبون أ", Phone: "07700000882"})
	if err == nil || !strings.Contains(err.Error(), "مسجّل لزبون ثاني") {
		t.Fatalf("نريد رسالة واضحة: %v", err)
	}
	if _, err := svc.Update("dup_a", model.UpdateCustomerRequest{Name: "زبون أ الجديد", Phone: "07700000881"}); err != nil {
		t.Fatalf("نفس رقمه لازم ينحفظ: %v", err)
	}
}
