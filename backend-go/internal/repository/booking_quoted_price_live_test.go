package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
)

// (11-07) حفظ حقل ثاني بتفاصيل الحجز (السيارة مثلاً) چان يمسح المبلغ المقدّر.
// غياب المفتاح = لا تلمسه، و null صريح = امسحه (المنسّق يمسحه بقصد).
func TestUpdateDetailsKeepsQuotedPriceLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	repo := NewBookingRepository(db)
	cust, b := uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون سعر', $2)`, cust, "0771"+sfx)
	db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "quotedPrice") VALUES ($1, $2, $3, 'PENDING', 250000)`, b, "QP"+sfx, cust)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, b)     //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust) //nolint:errcheck
	})
	price := func() *float64 {
		var p *float64
		if err := db.Get(&p, `SELECT "quotedPrice" FROM "Booking" WHERE id = $1`, b); err != nil {
			t.Fatal(err)
		}
		return p
	}
	upd := func(body string) {
		var req model.UpdateBookingDetailsRequest
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			t.Fatal(err)
		}
		if err := repo.UpdateDetails(b, req); err != nil {
			t.Fatal(err)
		}
	}
	upd(`{"assignedVehicle":"كيا ٢"}`)
	if p := price(); p == nil || *p != 250000 {
		t.Fatalf("حفظ السيارة مسح المبلغ: %v", p)
	}
	upd(`{"quotedPrice":300000}`)
	if p := price(); p == nil || *p != 300000 {
		t.Fatalf("تعديل المبلغ ما اشتغل: %v", p)
	}
	upd(`{"quotedPrice":null}`)
	if p := price(); p != nil {
		t.Fatalf("null صريح ما مسح المبلغ: %v", *p)
	}
}
