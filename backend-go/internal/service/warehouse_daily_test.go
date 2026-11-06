package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

func TestWarehouseDigestMessage(t *testing.T) {
	if (WarehouseDigest{ActiveCars: 3}).Message() != "" {
		t.Fatal("ماكو شي = بلا تنبيه")
	}
	m := WarehouseDigest{StaleRequests: 2, OldestDays: 6, ActiveCars: 5, Unrated: 4, Overdue: []string{"بيك اب (12345)"}}.Message()
	for _, want := range []string{"2 طلب", "6 يوم", "4 سيارة من 5", "بيك اب"} {
		if !strings.Contains(m, want) {
			t.Errorf("ناقص %q: %s", want, m)
		}
	}
	if got := carWashText(WarehouseDigest{ActiveCars: 5, WashedToday: 2, Unrated: 1}); !strings.Contains(got, "2 سيارة من 5") || !strings.Contains(got, "باقي 3") {
		t.Fatal(got)
	}
}

// مرة وحدة باليوم، وبعد ١١ الصبح بس.
func TestWarehouseDailyOncePerDayLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	veh := uuid.NewString()
	db.MustExec(`INSERT INTO "Vehicle" (id, name, "plateNumber", "isActive") VALUES ($1, 'سيارة فحص', $2, true)`, veh, "T"+veh[:6])
	t.Cleanup(func() { db.Exec(`DELETE FROM "Vehicle" WHERE id = $1`, veh) }) //nolint:errcheck
	claimed := map[string]bool{}
	claim := func(k, d string) (bool, error) {
		if claimed[k+d] {
			return false, nil
		}
		claimed[k+d] = true
		return true, nil
	}
	sent := 0
	w := NewWarehouseDaily(db, claim, nil, func(string) { sent++ })
	morning := time.Date(2026, 11, 7, 9, 0, 0, 0, debriefLoc)
	noon := time.Date(2026, 11, 7, 12, 0, 0, 0, debriefLoc)
	_ = w.RunIfDue(morning)
	if sent != 0 {
		t.Fatal("قبل ١١ ما ينبّه")
	}
	_ = w.RunIfDue(noon)
	_ = w.RunIfDue(noon.Add(time.Hour))
	if sent != 1 { // سيارة فحص فعّالة وما انقيّمت = أكو شي
		t.Fatalf("لازم مرة وحدة: %d", sent)
	}
}
