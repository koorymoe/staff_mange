package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// TestComplaintHydrateAllSame_Live بند ٧م: hydrateAll لازم يطلع نفس hydrate حرفياً.
func TestComplaintHydrateAllSame_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cleanup := func() {
		db.Exec(`DELETE FROM "Complaint" WHERE id LIKE '__hc%'`)
		db.Exec(`DELETE FROM "VipCustomer" WHERE "customerId" LIKE '__hc%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id = '__hcb'`)
		db.Exec(`DELETE FROM "Customer" WHERE id LIKE '__hc%'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE '__hc%'`)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ('__hce1','موظف أ','ACTIVE','TECHNICIAN'),('__hce2','موظف ب','ACTIVE','TECHNICIAN')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone, position) VALUES ('__hcc1','زبون أ','07999000001',NULL),('__hcc2','زبون ب','07999000002','مدير')`)
	db.MustExec(`INSERT INTO "VipCustomer"(id, "customerId", "customerPosition", "markedByEmployeeId") VALUES ('__hcv','__hcc1','وزير','__hce1')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", "scheduledAt") VALUES ('__hcb','HC-1','__hcc1', now())`)
	db.MustExec(`INSERT INTO "Complaint"(id, "customerId", "bookingId", type, description, "createdByEmployeeId", "relatedEmployeeId", "assignedToEmployeeId", "contactedById")
		VALUES ('__hc1','__hcc1','__hcb','OTHER','أ','__hce1','__hce2','__hce2','__hce1'),
		       ('__hc2','__hcc2',NULL,'OTHER','ب','__hce2',NULL,NULL,NULL)`)
	r := NewComplaintRepository(db)
	a, b := []model.Complaint{}, []model.Complaint{}
	if err := db.Select(&a, `SELECT * FROM "Complaint" WHERE id LIKE '__hc%' ORDER BY id`); err != nil {
		t.Fatal(err)
	}
	if err := db.Select(&b, `SELECT * FROM "Complaint" WHERE id LIKE '__hc%' ORDER BY id`); err != nil {
		t.Fatal(err)
	}
	for i := range a {
		r.hydrate(&a[i])
	}
	r.hydrateAll(b)
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Fatalf("hydrateAll يختلف:\n%s\n---\n%s", ja, jb)
	}
	if b[0].Customer == nil || b[0].Booking == nil || !b[0].IsVip || b[0].ContactedByName == nil || !b[1].IsVip {
		t.Fatalf("البيانات ما انعبّت: %+v", b[0])
	}
}
