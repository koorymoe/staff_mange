package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (ل) 11-25 دازاي «كل الحجوزات تطلع بلا خدمة وبلا ليدر» بصفحة الجودة: الحجز لازم ينجاب ويا خدمته وليدره.
func TestQualityFollowUpHydrateServiceLeaderLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "QualityFollowUp" WHERE "bookingId" = 'qh-b'`)
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" = 'qh-b'`)
		db.Exec(`DELETE FROM "BookingService" WHERE "bookingId" = 'qh-b'`)
		db.Exec(`DELETE FROM "Booking" WHERE id = 'qh-b'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'qh-cu'`)
		db.Exec(`DELETE FROM "Service" WHERE id = 'qh-s'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'qh-l'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role, status, "isLeader") VALUES ('qh-l', 'qh ليدر', 'TECHNICIAN', 'ACTIVE', true)`)
	db.MustExec(`INSERT INTO "Service"(id, name) VALUES ('qh-s', 'كاميرات qh')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('qh-cu', 'x', '07700008282')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "serviceId") VALUES ('qh-b', 'QH-B', 'qh-cu', 'COMPLETED', 'qh-s')`)
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES ('qh-a', 'qh-b', 'qh-l', 'TECH_1')`)
	db.MustExec(`INSERT INTO "QualityFollowUp"(id, "bookingId", "customerId") VALUES ('qh-q', 'qh-b', 'qh-cu')`)
	items, err := NewQualityFollowUpRepository(db).List()
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range items {
		if q.BookingID != "qh-b" {
			continue
		}
		if q.Booking == nil || q.Booking.Service == nil || q.Booking.Service.Name != "كاميرات qh" {
			t.Fatalf("الخدمة ما انجابت: %+v", q.Booking)
		}
		if q.Booking.LeaderName == nil || *q.Booking.LeaderName != "qh ليدر" {
			t.Fatalf("الليدر ما انجاب: %+v", q.Booking.LeaderName)
		}
		return
	}
	t.Fatal("المتابعة ما طلعت")
}
