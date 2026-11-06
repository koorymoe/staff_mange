package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// TestLateAuditRule (دازاي 11-06): يوم واحد مو تأخير، يومين فما فوق تأخير، والي بالانتظار
// («زبون ما رد») ينفصلون بقسم وحدهم.
func TestLateAuditRule(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	cust := uuid.NewString()
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون فحص التأخير', '07700000001')`, cust)
	loc := baghdadLoc()
	now := time.Now().In(loc)
	mk := func(code, status string, sched time.Time, done *time.Time, waiting bool) {
		id := uuid.NewString()
		db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "scheduledAt", "completedAt") VALUES ($1, $2, $3, $4::"BookingStatus", $5, $6)`,
			id, code, cust, status, sched, done)
		if waiting {
			db.MustExec(`UPDATE "Booking" SET "waitingSince" = now(), "waitingKind" = 'NO_ANSWER', "contactAttempts" = 2 WHERE id = $1`, id)
		}
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = $1`, cust) //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)          //nolint:errcheck
	})
	p := "LT" + uuid.NewString()[:5]
	d := func(days int, hour int) time.Time {
		y, m, dd := now.AddDate(0, 0, -days).Date()
		return time.Date(y, m, dd, hour, 0, 0, 0, loc)
	}
	next := d(5, 0).Add(32 * time.Hour)                // اليوم الثاني ٨ الصبح
	mk(p+"-1day", "COMPLETED", d(5, 16), &next, false) // موعد ٤ العصر وخلص اليوم الثاني = مو تأخير
	far := d(1, 12)
	mk(p+"-4days", "COMPLETED", d(5, 16), &far, false) // ٤ أيام = متأخر
	mk(p+"-open1", "CONFIRMED", d(1, 16), nil, false)  // موعده البارحة وبعده مفتوح = مو تأخير بعد
	mk(p+"-open3", "CONFIRMED", d(3, 16), nil, false)  // ٣ أيام مفتوح = متأخر
	mk(p+"-wait", "CONFIRMED", d(4, 16), nil, true)    // زبون ما رد = بالانتظار

	r := httptest.NewRequest(http.MethodGet, "/api/late-audit?from="+d(7, 0).Format("2006-01-02")+"&to="+now.Format("2006-01-02"), nil)
	w := httptest.NewRecorder()
	NewLateAuditHandler(db).List(w, r)
	var out struct {
		Rows    []lateRow `json:"rows"`
		Waiting []lateRow `json:"waiting"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	late, wait := map[string]bool{}, map[string]bool{}
	for _, x := range out.Rows {
		late[x.Code] = true
	}
	for _, x := range out.Waiting {
		wait[x.Code] = true
	}
	for code, want := range map[string]bool{p + "-1day": false, p + "-4days": true, p + "-open1": false, p + "-open3": true, p + "-wait": false} {
		if late[code] != want {
			t.Errorf("%s: متأخر=%v، المتوقع %v", code, late[code], want)
		}
	}
	if !wait[p+"-wait"] {
		t.Error("زبون ما رد لازم يطلع بقسم الانتظار")
	}
}
