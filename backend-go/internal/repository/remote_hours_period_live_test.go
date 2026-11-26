package repository

import (
	"os"
	"testing"

	"staffmange-api/internal/database"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// الفترة بالقاعدة: سجل يوم 26 يروح للشهر الجاي، والرفع يسكّر لحد 25 بس.
func TestRemotePeriodLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("no DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	r := NewRemoteHoursRepository(db)
	var emp string
	if err := db.Get(&emp, `SELECT id FROM "Employee" LIMIT 1`); err != nil {
		t.Skip("no employee")
	}
	tx := func(q string, a ...any) {
		t.Helper()
		if _, err := db.Exec(q, a...); err != nil {
			t.Fatal(err)
		}
	}
	tx(`DELETE FROM "RemoteWorkEntry" WHERE note = 'period-test'`)
	defer tx(`DELETE FROM "RemoteWorkEntry" WHERE note = 'period-test'`)
	defer tx(`DELETE FROM "RemoteHoursConfig"`)
	if err := r.SaveConfig(RemoteHoursConfig{PeriodStartDay: 26, CloseDay: 25, CloseHour: 23}, ""); err != nil {
		t.Fatal(err)
	}
	n := "period-test"
	for _, d := range []string{"2031-11-25", "2031-11-26"} {
		if _, err := r.CreateFrom(emp, d, 3600, &n, "", "MANUAL"); err != nil {
			t.Fatal(err)
		}
	}
	dec, _ := r.List(emp, "2031-12")
	nov, _ := r.List(emp, "2031-11")
	cnt := func(es []RemoteEntry) (k int) {
		for _, e := range es {
			if e.Note != nil && *e.Note == n {
				k++
			}
		}
		return
	}
	if cnt(dec) != 1 || cnt(nov) != 1 {
		t.Fatalf("dec=%d nov=%d", cnt(dec), cnt(nov))
	}
	ms, _ := r.Months()
	found := false
	for _, m := range ms {
		found = found || m.Month == "2031-12"
	}
	if !found {
		t.Fatal("2031-12 not in months")
	}
	_, to := PeriodRange("2031-11", 26)
	if _, err := r.CloseUpTo(to, ""); err != nil {
		t.Fatal(err)
	}
	var open int
	db.Get(&open, `SELECT COUNT(*) FROM "RemoteWorkEntry" WHERE note = $1 AND "closedAt" IS NULL`, n)
	if open != 1 {
		t.Fatalf("open=%d want 1 (26th stays open)", open)
	}
	_ = time.Now
}
