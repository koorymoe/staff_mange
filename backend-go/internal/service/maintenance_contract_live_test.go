package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (ع) 11-08: العقد المستحق خلال أسبوع يفتح حجز مرة وحدة ويتقدّم موعده؛ البعيد والموقوف لا.
func TestMaintenanceContracts_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "MaintenanceContract" WHERE "customerId" = 'mc-cust'`)                                            //nolint:errcheck
		db.Exec(`DELETE FROM "BookingService" WHERE "bookingId" IN (SELECT id FROM "Booking" WHERE "customerId" = 'mc-cust')`) //nolint:errcheck
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = 'mc-cust'`)                                                        //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = 'mc-cust'`)                                                                 //nolint:errcheck
		db.Exec(`DELETE FROM "Service" WHERE id = 'mc-svc'`)                                                                   //nolint:errcheck
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('mc-cust','زبون عقد','07700009911')`)
	db.MustExec(`INSERT INTO "Service"(id, name) VALUES ('mc-svc','صيانة كاميرات عقد')`)
	notif := repository.NewNotificationRepository(db)
	bs := NewBookingService(repository.NewBookingRepository(db), repository.NewEmployeeRepository(db), repository.NewCustomerRepository(db),
		repository.NewQualityFollowUpRepository(db), notif, repository.NewInventoryRepository(db))
	s := NewMaintenanceContractService(db, bs, notif)
	soon := time.Now().AddDate(0, 0, 3).Format("2006-01-02")
	far := time.Now().AddDate(0, 2, 0).Format("2006-01-02")
	if _, err := s.Create("", ContractInput{CustomerID: "mc-cust", ServiceID: "mc-svc", IntervalMonths: 0, NextVisitDate: soon}); err == nil {
		t.Fatal("لازم الفترة")
	}
	due, err := s.Create("", ContractInput{CustomerID: "mc-cust", ServiceID: "mc-svc", IntervalMonths: 3, NextVisitDate: soon})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("", ContractInput{CustomerID: "mc-cust", ServiceID: "mc-svc", IntervalMonths: 6, NextVisitDate: far}); err != nil {
		t.Fatal(err)
	}
	off, _ := s.Create("", ContractInput{CustomerID: "mc-cust", ServiceID: "mc-svc", IntervalMonths: 1, NextVisitDate: soon})
	f := false
	if err := s.Update(off, ContractInput{Active: &f}); err != nil {
		t.Fatal(err)
	}
	if n := s.GenerateDue(); n != 1 {
		var why *string
		db.Get(&why, `SELECT "lastError" FROM "MaintenanceContract" WHERE id = $1`, due) //nolint:errcheck
		t.Fatalf("لازم ينفتح حجز واحد: %d (%v)", n, str(why))
	}
	if n := s.GenerateDue(); n != 0 {
		t.Fatal("ما لازم ينفتح مرتين")
	}
	var next string
	var hasBooking bool
	db.QueryRow(`SELECT "nextVisitDate"::text, "lastBookingId" IS NOT NULL FROM "MaintenanceContract" WHERE id = $1`, due).Scan(&next, &hasBooking)
	want := time.Now().AddDate(0, 0, 3).AddDate(0, 3, 0).Format("2006-01-02")
	if next != want || !hasBooking {
		t.Fatalf("الموعد الجاي %s (نريد %s) والحجز %v", next, want, hasBooking)
	}
}
