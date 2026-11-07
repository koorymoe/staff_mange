package repository

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
)

// (11-07) التجهيز: بند من طلب ثاني ينرفض، والمجهّز ما ينعاد تجهيزه ولا ترجع حالته.
func TestProcurementFulfillGuardsLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	repo := NewProcurementRepository(db)
	emp := uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'طالب مواد', $2, 'TECHNICIAN', 'ACTIVE')`, emp, "pr_"+sfx)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "ProcurementRequest" WHERE "requestedById" = $1`, emp) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)                       //nolint:errcheck
	})
	mk := func() *model.ProcurementRequest {
		r, err := repo.Create(model.CreateProcurementRequestRequest{RequestedByID: emp, RequestType: model.RequestTypePersonalSupply,
			Items: []model.CreateProcurementItemRequest{{ProductName: "كيبل " + sfx, Quantity: 1}}})
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	a, b := mk(), mk()
	price := 5000.0
	// بند من الطلب b بتجهيز الطلب a
	if _, err := repo.Fulfill(a.ID, model.FulfillProcurementRequestRequest{FulfilledByID: emp, Items: []model.FulfillProcurementItemRequest{{ID: b.Items[0].ID, UnitPrice: &price}}}); err == nil {
		t.Fatal("بند من طلب ثاني انقبل")
	}
	var bp *float64
	_ = db.Get(&bp, `SELECT "unitPrice" FROM "ProcurementItem" WHERE id = $1`, b.Items[0].ID)
	if bp != nil && *bp == price {
		t.Fatal("سعر بند الطلب الثاني تغيّر")
	}
	cost := 5000.0
	if _, err := repo.Fulfill(a.ID, model.FulfillProcurementRequestRequest{FulfilledByID: emp, TotalCost: &cost, Items: []model.FulfillProcurementItemRequest{{ID: a.Items[0].ID, UnitPrice: &price}}}); err != nil {
		t.Fatal(err)
	}
	other := 9.0
	if _, err := repo.Fulfill(a.ID, model.FulfillProcurementRequestRequest{FulfilledByID: emp, TotalCost: &other}); err == nil {
		t.Error("تجهيز ثاني انقبل")
	}
	if _, err := repo.UpdateStatus(a.ID, "PENDING", emp); err == nil {
		t.Error("المجهّز رجع معلّق")
	}
	var st string
	var tc float64
	_ = db.QueryRow(`SELECT status, "totalCost" FROM "ProcurementRequest" WHERE id = $1`, a.ID).Scan(&st, &tc)
	if st != "FULFILLED" || tc != cost {
		t.Fatalf("الحالة %s والكلفة %v", st, tc)
	}
	// المعلّق يتحدّث عادي — چان يفشل دائماً بخطأ نوع المعامل (من 10-10)
	for _, st := range []string{"IN_PROGRESS", "REJECTED"} {
		got, err := repo.UpdateStatus(b.ID, st, emp)
		if err != nil {
			t.Fatalf("%s: %v", st, err)
		}
		if got.Status != st {
			t.Fatalf("الحالة %s، المفروض %s", got.Status, st)
		}
	}
}
