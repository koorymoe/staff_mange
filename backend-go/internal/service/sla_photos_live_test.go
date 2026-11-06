package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// TestSlaAndPhotosLive (دازاي 11-06): الالتزام يحسب من التسجيل + ساعات الخدمة، الانتظار والمنتهي برا،
// التنبيه مرة وحدة، و«تم» ينرفض قبل أقل عدد صور.
func TestSlaAndPhotosLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	svc, cust := uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Service" (id, name, "slaHours", "minPhotos") VALUES ($1, $2, 10, 2)`, svc, "خدمة فحص الالتزام "+sfx)
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون فحص الالتزام', '07700000005')`, cust)
	ids := map[string]string{}
	mk := func(key string, createdAgo time.Duration, status string, waiting bool) {
		id := uuid.NewString()
		ids[key] = id
		db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", "serviceId", status, "scheduledAt", "createdAt") VALUES ($1, $2, $3, $4, $5::"BookingStatus", now(), now() - $6::interval)`,
			id, "SL"+key+sfx, cust, svc, status, createdAgo.String())
		if waiting {
			db.MustExec(`UPDATE "Booking" SET "waitingSince" = now(), "waitingKind" = 'NO_ANSWER' WHERE id = $1`, id)
		}
	}
	mk("soon", 6*time.Hour, "CONFIRMED", false)  // باقي ٤ ساعات
	mk("over", 12*time.Hour, "CONFIRMED", false) // فات بساعتين
	mk("fresh", 1*time.Hour, "CONFIRMED", false) // باقي ٩ — ما يطلع بالتنبيه (٦)
	mk("wait", 12*time.Hour, "CONFIRMED", true)  // بالانتظار — برا
	mk("done", 12*time.Hour, "COMPLETED", false) // منتهي — برا
	mk("old", 100*time.Hour, "CONFIRMED", false) // فات من زمان — باللوحة بلا تنبيه
	t.Cleanup(func() {
		for _, id := range ids {
			db.Exec(`DELETE FROM "Booking" WHERE id = $1`, id) //nolint:errcheck
		}
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust) //nolint:errcheck
		db.Exec(`DELETE FROM "Service" WHERE id = $1`, svc)   //nolint:errcheck
	})
	board, err := SlaBoard(db, 24)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]float64{}
	for _, it := range board {
		got[it.ID] = it.HoursLeft
	}
	if _, ok := got[ids["wait"]]; ok {
		t.Error("الانتظار لازم برا")
	}
	if _, ok := got[ids["done"]]; ok {
		t.Error("المنتهي لازم برا")
	}
	if h := got[ids["over"]]; h > -1.5 || h < -2.5 {
		t.Errorf("فات بساعتين: %v", h)
	}
	kinds := map[string]int{}
	al := NewSlaAlerter(db, func(kind, msg string) {
		if strings.Contains(msg, sfx) {
			kinds[kind]++
		}
	})
	_ = al.Tick()
	_ = al.Tick() // مرة وحدة بس
	if _, ok := got[ids["old"]]; !ok {
		t.Errorf("القديم لازم يبين باللوحة (%d صف، over=%v)", len(board), got[ids["over"]])
	}
	if kinds["SOON"] != 1 || kinds["OVERDUE"] != 1 { // القديم ما ينبّه
		t.Fatalf("التنبيهات: %v", kinds)
	}

	repo := repository.NewBookingRepository(db)
	s := &BookingService{repo: repo}
	if err := s.requirePhotos(ids["soon"]); err == nil || !strings.Contains(err.Error(), "2") {
		t.Fatalf("لازم يرفض بلا صور: %v", err)
	}
	for i := 0; i < 2; i++ {
		db.MustExec(`INSERT INTO "WorkPhoto" ("bookingId", phase, "fileKey", "contentType") VALUES ($1, 'AFTER', 'k', 'image/jpeg')`, ids["soon"])
	}
	if err := s.requirePhotos(ids["soon"]); err != nil {
		t.Fatalf("صورتين تكفي: %v", err)
	}
}
