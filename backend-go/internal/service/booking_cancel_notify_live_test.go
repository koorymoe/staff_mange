package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (11-15) إلغاء حجز عليه كادر يوصل إشعار لكل واحد بالكادر (عدا الي لغاه).
func TestBookingCancelNotifiesCrewLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" LIKE 'bc-%'`)
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" = 'bc-b'`)
		db.Exec(`DELETE FROM "Booking" WHERE id = 'bc-b'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'bc-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'bc-%'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('bc-t1', 'bc فني', 'TECHNICIAN'), ('bc-t2', 'bc ليدر', 'TECHNICIAN'), ('bc-m', 'bc منسق', 'HR_COORDINATOR')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('bc-cu', 'x', '07700001212')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status) VALUES ('bc-b', 'Bbc', 'bc-cu', 'IN_PROGRESS')`)
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES ('bc-a1', 'bc-b', 'bc-t1', 'TECH_1'), ('bc-a2', 'bc-b', 'bc-t2', 'TECH_2')`)
	s := NewBookingService(repository.NewBookingRepository(db), repository.NewEmployeeRepository(db), repository.NewCustomerRepository(db),
		repository.NewQualityFollowUpRepository(db), repository.NewNotificationRepository(db), repository.NewInventoryRepository(db))
	if _, err := s.Cancel("bc-b", "الزبون ألغى", "bc-m"); err != nil {
		t.Fatal(err)
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE type = 'booking_cancelled' AND "employeeId" IN ('bc-t1', 'bc-t2')`)
	if n != 2 {
		t.Fatalf("الكادر لازم يوصله إشعار: %d", n)
	}
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'bc-m'`)
	if n != 0 {
		t.Fatal("الي لغاه ما يوصله")
	}
}
