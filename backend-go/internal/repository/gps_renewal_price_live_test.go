package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
)

// (دازاي 11-17) «أسعار اشتراك GPS تدخل بالتجديد»: الطلب ياخذ سعر نوعه، والموافقة تثبّته.
func TestGpsRenewalPrice_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const emp, cu, dev = "gprice-emp", "gprice-cu", "gprice-dev"
	var oldPrice *float64
	_ = db.Get(&oldPrice, `SELECT price FROM "GpsSubscriptionPrice" WHERE "subscriptionType" = 'SIX_MONTHS'`)
	clean := func() {
		db.Exec(`DELETE FROM "GpsRenewalRequest" WHERE "deviceRequestId" = $1`, dev)
		db.Exec(`DELETE FROM "GpsDeviceRequest" WHERE id = $1`, dev)
		db.Exec(`DELETE FROM "GpsCustomer" WHERE id = $1`, cu)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	clean()
	t.Cleanup(func() {
		clean()
		if oldPrice == nil {
			db.Exec(`DELETE FROM "GpsSubscriptionPrice" WHERE "subscriptionType" = 'SIX_MONTHS'`)
		} else {
			db.Exec(`UPDATE "GpsSubscriptionPrice" SET price = $1 WHERE "subscriptionType" = 'SIX_MONTHS'`, *oldPrice)
		}
	})
	db.MustExec(`INSERT INTO "Employee"(id, name) VALUES ($1, 'موظف GPS')`, emp)
	db.MustExec(`INSERT INTO "GpsCustomer"(id, "fullName", phone) VALUES ($1, 'زبون GPS', '07700009999')`, cu)
	db.MustExec(`INSERT INTO "GpsDeviceRequest"(id, "customerId", "employeeId", "purchaseType", "subscriptionType") VALUES ($1, $2, $3, 'DEVICE_SIM', 'SIX_MONTHS')`, dev, cu, emp)

	r := NewGpsRepository(db)
	if _, err := r.UpsertSetting("SIX_MONTHS", "SIX_MONTHS", 45000); err != nil {
		t.Fatal(err)
	}
	if _, err := r.UpsertSetting("SIX_MONTHS", "SIX_MONTHS", 50000); err != nil {
		t.Fatalf("تعديل السعر لازم يمشي: %v", err)
	}
	c, d, e, typ := cu, dev, emp, "SIX_MONTHS"
	rn, err := r.CreateRenewal(model.UpsertGpsRenewalRequest{CustomerID: &c, DeviceRequestID: &d, EmployeeID: &e, SubscriptionType: &typ})
	if err != nil {
		t.Fatal(err)
	}
	if rn.Price == nil || *rn.Price != 50000 {
		t.Fatalf("الطلب لازم ياخذ سعر نوعه: %v", rn.Price)
	}
	ap, err := r.ApproveRenewal(rn.ID, nil)
	if err != nil || ap.Price == nil || *ap.Price != 50000 {
		t.Fatalf("الموافقة تثبّت السعر: %v %v", ap, err)
	}
}
