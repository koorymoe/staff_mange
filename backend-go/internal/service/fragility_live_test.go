package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// (ل) 11-25 دازاي: «موظف واحد بس يشتغل على الأقفال» ← حرج، ويطلع اسمه «إذا غاب ماكو غيره».
func TestFragilityLive(t *testing.T) {
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
		db.Exec(`DELETE FROM "BookingService" WHERE "bookingId" LIKE 'fr-%'`)
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" LIKE 'fr-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'fr-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'fr-cu'`)
		db.Exec(`DELETE FROM "Service" WHERE id LIKE 'fr-%'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'fr-%'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role, status) VALUES ('fr-a', 'fr أقفال', 'TECHNICIAN', 'ACTIVE'), ('fr-b', 'fr ب', 'TECHNICIAN', 'ACTIVE'), ('fr-c', 'fr ج', 'TECHNICIAN', 'ACTIVE')`)
	db.MustExec(`INSERT INTO "Service"(id, name) VALUES ('fr-lock', 'أقفال fr'), ('fr-cam', 'كاميرات fr')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('fr-cu', 'x', '07700009191')`)
	for i, x := range []struct{ id, svc, emp string }{{"fr-1", "fr-lock", "fr-a"}, {"fr-2", "fr-lock", "fr-a"}, {"fr-3", "fr-cam", "fr-b"}, {"fr-4", "fr-cam", "fr-c"}, {"fr-5", "fr-cam", "fr-a"}} {
		db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "completedAt") VALUES ($1, upper($1), 'fr-cu', 'COMPLETED', now() - interval '2 days')`, x.id)
		db.MustExec(`INSERT INTO "BookingService"("bookingId", "serviceId") VALUES ($1, $2) ON CONFLICT DO NOTHING`, x.id, x.svc)
		db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES ($1, $2, $3, 'TECH_1')`, x.id+"-a", x.id, x.emp)
		_ = i
	}
	rep, err := Fragility(db, 30, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	risk := map[string]string{}
	for _, s := range rep.Services {
		risk[s.ServiceID] = s.Risk
	}
	if risk["fr-lock"] != "CRITICAL" || risk["fr-cam"] != "OK" {
		t.Fatalf("الخطر غلط: %v", risk)
	}
	found := false
	for _, p := range rep.People {
		if p.EmployeeID == "fr-a" && len(p.SoleFor) == 1 && p.SoleFor[0] == "أقفال fr" {
			found = true
		}
	}
	if !found {
		t.Fatalf("لازم يطلع «إذا غاب fr-a ماكو غيره للأقفال»: %+v", rep.People)
	}
}
