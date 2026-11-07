package service

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

func TestMaintenanceSummaryNoNames(t *testing.T) {
	cam := "كاميرات"
	out := maintenanceSummary([]MaintenanceDue{{Name: "زبون سري جداً", Service: &cam, Reason: "فات ٦ أشهر على آخر شغل"}, {Name: "ثاني", Reason: "شكاوى متكررة"}})
	if strings.Contains(out, "سري") || strings.Contains(out, "ثاني") || !strings.Contains(out, "كاميرات: 1") || !strings.HasPrefix(out, "2 زبون") {
		t.Fatalf("%s", out)
	}
}

func TestMaintenanceDueLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	var bookings []string
	sfx := uuid.NewString()[:6]
	cust := func(key string, doneAgo time.Duration, open bool, complaints int) {
		c := uuid.NewString()
		ids[key] = c
		db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, $2, $3)`, c, "صيانة "+key+sfx, fmt.Sprintf("077%08d", rand.Intn(1e8)))
		b := uuid.NewString()
		bookings = append(bookings, b)
		db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "scheduledAt", "completedAt") VALUES ($1, $2, $3, 'COMPLETED', now() - $4::interval, now() - $4::interval)`, b, "MD"+key+sfx, c, doneAgo.String())
		if open {
			o := uuid.NewString()
			bookings = append(bookings, o)
			db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "scheduledAt") VALUES ($1, $2, $3, 'CONFIRMED', now())`, o, "MDo"+key+sfx, c)
		}
		for i := 0; i < complaints; i++ {
			db.MustExec(`INSERT INTO "Complaint" (id, "customerId", description, status, "createdByEmployeeId") VALUES (gen_random_uuid()::text, $1, 'فحص', 'NEW', (SELECT id FROM "Employee" LIMIT 1))`, c)
		}
	}
	day := 24 * time.Hour
	cust("old", 200*day, false, 0)      // فات ٦ أشهر = يطلع
	cust("recent", 30*day, false, 0)    // جديد = لا
	cust("busy", 200*day, true, 0)      // عنده حجز مفتوح = لا
	cust("angry", 30*day, false, 2)     // شكوتين = يطلع
	cust("ancient", 1500*day, false, 0) // أكثر من ٣ سنين = لا
	cust("stale", 200*day, true, 0)     // حجزه المفتوح مؤرشف = يطلع (11-07)
	db.MustExec(`UPDATE "Booking" SET "archivedAt" = now() WHERE code = $1`, "MDostale"+sfx)
	t.Cleanup(func() {
		for _, c := range ids {
			db.Exec(`DELETE FROM "Complaint" WHERE "customerId" = $1`, c) //nolint:errcheck
		}
		for _, b := range bookings {
			db.Exec(`DELETE FROM "Booking" WHERE id = $1`, b) //nolint:errcheck
		}
		for _, c := range ids {
			db.Exec(`DELETE FROM "Customer" WHERE id = $1`, c) //nolint:errcheck
		}
	})
	rows, err := MaintenanceDueList(db, 100000)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, r := range rows {
		got[r.CustomerID] = r.Reason
	}
	for key, want := range map[string]bool{"old": true, "recent": false, "busy": false, "angry": true, "ancient": false, "stale": true} {
		if _, ok := got[ids[key]]; ok != want {
			t.Errorf("%s: يطلع=%v، المتوقع %v", key, ok, want)
		}
	}
	if got[ids["angry"]] != "شكاوى متكررة" {
		t.Errorf("السبب: %q", got[ids["angry"]])
	}
}
