package service

import (
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// (11-07) إنشاء عرض لمشروع عنده عرض چان يرجّع العرض الموجود لأي أحد (بنوده وأسعاره وزبونه)،
// والمنشئ چان ينوخذ من جسم الطلب.
func TestQuotationCreateAuthzLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	svc := NewQuotationService(repository.NewQuotationRepository(db), repository.NewPermissionRepository(db))
	owner, stranger, colleague := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES
		($1, 'صاحب العرض', $2, 'SALES', 'ACTIVE'), ($3, 'غريب', $4, 'SALES', 'ACTIVE'), ($5, 'زميل', $6, 'SALES', 'ACTIVE')`,
		owner, "qo_"+sfx, stranger, "qs_"+sfx, colleague, "qc_"+sfx)
	proj := uuid.NewString()
	db.MustExec(`INSERT INTO "Project" (id, code, name) VALUES ($1, $2, $3)`, proj, "PQ"+sfx, "مشروع عرض "+sfx)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Quotation" WHERE "projectId" = $1 OR "createdByEmployeeId" IN ($2, $3, $4)`, proj, owner, stranger, colleague) //nolint:errcheck
		db.Exec(`DELETE FROM "Project" WHERE id = $1`, proj)                                                                                 //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2, $3)`, owner, stranger, colleague)                                               //nolint:errcheck
	})
	items := []model.QuotationItemInput{{ProductName: "كاميرا", Quantity: 1, UnitPrice: 100000}}
	q, err := svc.Create(model.CreateQuotationRequest{CustomerName: "زبون سري", CreatedByEmployeeID: owner, ProjectID: &proj, Items: items}, owner, "SALES")
	if err != nil {
		t.Fatal(err)
	}
	// غريب بلا صلاحية يطلب نفس المشروع: لازم «مو موجود»، مو العرض
	if got, err := svc.Create(model.CreateQuotationRequest{CustomerName: "x", CreatedByEmployeeID: stranger, ProjectID: &proj}, stranger, "SALES"); !errors.Is(err, ErrQuotationHidden) {
		t.Fatalf("الغريب استلم عرض غيره: %v err=%v", got != nil && got.ID == q.ID, err)
	}
	// المنشئ من الجلسة، مو من الجسم
	q2, err := svc.Create(model.CreateQuotationRequest{CustomerName: "زبون ٢", CreatedByEmployeeID: colleague, Items: items}, stranger, "SALES")
	if err != nil {
		t.Fatal(err)
	}
	if q2.CreatedByEmployeeID != stranger {
		t.Fatalf("المنشئ انوخذ من الطلب: %s", q2.CreatedByEmployeeID)
	}
}
