package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// TestRemoteHoursAgent (قرار 0011): بلا رمز ٤٠٤، الرمز يسجّل للموظف الثابت بس، والحدود والتكرار.
func TestRemoteHoursAgent(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	emp := "rh-" + uuid.NewString()[:8]
	db.MustExec(`INSERT INTO "Employee" (id, name, role, status) VALUES ($1, 'م', 'OWNER', 'ACTIVE')`, emp)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "RemoteWorkEntry" WHERE "employeeId" = $1`, emp) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)                  //nolint:errcheck
	})
	const tok = "agent-token-for-tests-0123456789abcdef"
	h := NewRemoteHoursHandler(repository.NewRemoteHoursRepository(db))
	today := time.Now().In(baghdad).Format("2006-01-02")
	post := func(token, key, body string) int {
		r := httptest.NewRequest(http.MethodPost, "/api/remote-hours/agent", bytes.NewBufferString(body))
		if token != "" {
			r.Header.Set("X-Remote-Hours-Agent-Token", token)
		}
		if key != "" {
			r.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		h.AgentAdd(w, r)
		return w.Code
	}
	ok := `{"date":"` + today + `","minutes":9,"note":"فحص ١٠ الصبح"}`
	if post(tok, "k-000001", ok) != 404 {
		t.Fatal("قبل التفعيل لازم ٤٠٤")
	}
	h.SetAgent(tok, emp)
	if post("", "k-000001", ok) != 404 || post("wrong-token-wrong-token-wrong-token-x", "k-000001", ok) != 404 {
		t.Fatal("بلا رمز أو رمز غلط = ٤٠٤")
	}
	if c := post(tok, "", ok); c != 400 {
		t.Fatalf("بلا مفتاح: %d", c)
	}
	if c := post(tok, "k-000001", ok); c != 201 {
		t.Fatalf("الأول: %d", c)
	}
	if c := post(tok, "k-000001", ok); c != 200 { // نفس المفتاح = ما يتكرر
		t.Fatalf("التكرار: %d", c)
	}
	var n, sec int
	var note string
	db.Get(&n, `SELECT count(*) FROM "RemoteWorkEntry" WHERE "employeeId" = $1`, emp)
	db.Get(&sec, `SELECT seconds FROM "RemoteWorkEntry" WHERE "employeeId" = $1`, emp)
	db.Get(&note, `SELECT note FROM "RemoteWorkEntry" WHERE "employeeId" = $1`, emp)
	if n != 1 || sec != 540 || note != "🤖 Claude [k-000001]: فحص ١٠ الصبح" {
		t.Fatalf("السطر: %d %d %q", n, sec, note)
	}
	old := time.Now().In(baghdad).AddDate(0, 0, -3).Format("2006-01-02")
	for name, c := range map[string]struct{ body string }{
		"أكثر من ساعتين":  {`{"date":"` + today + `","minutes":121,"note":"x"}`},
		"بلا ملاحظة":      {`{"date":"` + today + `","minutes":5,"note":""}`},
		"أقدم من يومين":   {`{"date":"` + old + `","minutes":5,"note":"x"}`},
		"تاريخ بالمستقبل": {`{"date":"2999-01-01","minutes":5,"note":"x"}`},
	} {
		if code := post(tok, "k-"+uuid.NewString()[:8], c.body); code != 400 {
			t.Errorf("%s: %d", name, code)
		}
	}
	for i := 0; i < 3; i++ { // ٩ دقايق + ٣×١٢٠ = ٦:٠٩ ← الثالثة تنرفض
		code := post(tok, "k-cap-"+uuid.NewString()[:6], `{"date":"`+today+`","minutes":120,"note":"تصليح"}`)
		if want := map[bool]int{true: 400, false: 201}[i == 2]; code != want {
			t.Fatalf("حد اليوم #%d: %d", i, code)
		}
	}
}
