package service

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (11-07) موعد حجز الديكور الي ينكتب باليد (بتوقيت بغداد) چان ينخزن ٣ ساعات غلط.
func TestCreateDecorScheduleBaghdadLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	svc := NewBookingService(repository.NewBookingRepository(db), repository.NewEmployeeRepository(db), repository.NewCustomerRepository(db),
		repository.NewQualityFollowUpRepository(db), repository.NewNotificationRepository(db), repository.NewInventoryRepository(db))
	cust, parent, serv := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون ديكور', $2)`, cust, "0772"+sfx)
	db.MustExec(`INSERT INTO "Service" (id, name) VALUES ($1, $2)`, serv, "ديكور فحص "+sfx)
	db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status) VALUES ($1, $2, $3, 'PENDING')`, parent, "DP"+sfx, cust)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = $1`, cust) //nolint:errcheck
		db.Exec(`DELETE FROM "Service" WHERE id = $1`, serv)           //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)          //nolint:errcheck
	})
	at := "2026-11-10T10:00"
	b, err := svc.CreateDecor(parent, serv, "WITH", &at, nil)
	if err != nil {
		t.Fatal(err)
	}
	if b.ScheduledAt == nil {
		t.Fatal("ما انحفظ موعد")
	}
	// ١٠ الصبح بغداد = ٧ الصبح UTC
	if got := b.ScheduledAt.UTC().Format("15:04"); got != "07:00" {
		t.Fatalf("الموعد انخزن %s UTC، المفروض 07:00", got)
	}
}
