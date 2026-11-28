package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (11-28) تجديد الـGPS المعتمد ينحسب «فلوس دخّلها» للي سجّله — والمعلّق ما ينحسب.
func TestRevenueGpsLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "GpsRenewalRequest" WHERE id LIKE 'rg-%'`)
		db.Exec(`DELETE FROM "GpsDeviceRequest" WHERE id = 'rg-d'`)
		db.Exec(`DELETE FROM "GpsCustomer" WHERE id = 'rg-c'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'rg-e'`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name, username, password, role, status) VALUES ('rg-e','rg','rg-e','x','SALES','ACTIVE')`)
	db.MustExec(`INSERT INTO "GpsCustomer"(id, "fullName", phone) VALUES ('rg-c','rg زبون','0770')`)
	db.MustExec(`INSERT INTO "GpsDeviceRequest"(id, "customerId", "employeeId", "purchaseType", "subscriptionType") VALUES ('rg-d','rg-c','rg-e','DEVICE_SIM','THREE_MONTHS')`)
	db.MustExec(`INSERT INTO "GpsRenewalRequest"(id, "customerId", "deviceRequestId", "employeeId", status, price, "subscriptionType") VALUES
		('rg-1','rg-c','rg-d','rg-e','APPROVED',150000,'THREE_MONTHS'), ('rg-2','rg-c','rg-d','rg-e','PENDING',90000,'THREE_MONTHS')`)
	rows, err := NewMatrixScoreRepository(db).Revenue(time.Now().Add(-time.Hour), time.Now().Add(time.Hour), "rg-e")
	if err != nil {
		t.Fatal(err)
	}
	var gps float64
	for _, r := range rows {
		if r.Kind == "GPS" {
			gps += r.Amount
		}
	}
	if gps != 150000 {
		t.Fatalf("GPS = %v (المعتمد بس لازم 150000): %+v", gps, rows)
	}
}
