package repository

import (
	"math"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
)

// بلاغ #61: خانة توقيع الزبون تنحفظ وتتعدّل. وبلاغ #60: خصم بمبلغ = نسبة كسرية، والقيمة ترجع نفس المبلغ.
func TestQuotationSignatureLive(t *testing.T) {
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
	r := NewQuotationRepository(db)
	pct := 50000.0 / 992000.0 * 100
	q, err := r.Create(model.CreateQuotationRequest{CustomerName: "sig-test", DiscountPercent: &pct, CustomerSignature: true, CreatedByEmployeeID: emp,
		Items: []model.QuotationItemInput{{ProductName: "x", Quantity: 1, UnitPrice: 992000}}})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DELETE FROM "QuotationVersion" WHERE "quotationId" = $1`, q.ID)
	defer db.Exec(`DELETE FROM "QuotationItem" WHERE "quotationId" = $1`, q.ID)
	defer db.Exec(`DELETE FROM "Quotation" WHERE id = $1`, q.ID)
	if !q.CustomerSignature {
		t.Fatal("التوقيع ما انحفظ")
	}
	if math.Round(q.DiscountValue) != 50000 {
		t.Fatalf("الخصم %v لازم 50000", q.DiscountValue)
	}
	off := false
	u, err := r.Update(q.ID, model.UpdateQuotationRequest{CustomerSignature: &off}, "")
	if err != nil {
		t.Fatal(err)
	}
	if u.CustomerSignature {
		t.Fatal("التوقيع ما انشال")
	}
	u, err = r.Update(q.ID, model.UpdateQuotationRequest{}, "")
	if err != nil || u.CustomerSignature {
		t.Fatalf("تعديل بلا الحقل غيّره: %v %v", err, u.CustomerSignature)
	}
}
