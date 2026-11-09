package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// TestSuggestionVerifyFlow (دازاي 11-06، بلاغ #10): «تم» يحوّل الاقتراح لـ VERIFYING،
// والوكيل صاحبه بس يگدر يسجّل التأكد: موجود → DONE، ما موجود → ACCEPTED + إشعار للمالك.
func TestSuggestionVerifyFlow(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	mk := func(agent, title string) string {
		var id string
		if err := db.Get(&id, `INSERT INTO "AiSuggestion" (agent, title, details) VALUES ($1, $2, 'x') RETURNING id`, agent, title); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { db.MustExec(`DELETE FROM "AiSuggestion" WHERE id = $1`, id) })
		return id
	}
	h := NewAiBrainHandler(db)
	decide := func(id, status string) int {
		r := httptest.NewRequest(http.MethodPost, "/api/ai/suggestions/"+id+"/decide", bytes.NewBufferString(`{"status":"`+status+`"}`))
		r.SetPathValue("id", id)
		r = r.WithContext(context.WithValue(r.Context(), middleware.ContextEmployeeID, ""))
		w := httptest.NewRecorder()
		h.Decide(w, r)
		return w.Code
	}
	statusOf := func(id string) (s string) { db.Get(&s, `SELECT status FROM "AiSuggestion" WHERE id = $1`, id); return }

	a, b := mk("MATRIX", "تجربة تأكد أ "+t.Name()), mk("MATRIX", "تجربة تأكد ب "+t.Name())
	other := mk("IRIS", "تجربة تأكد ايريس "+t.Name())
	for _, id := range []string{a, b, other} {
		if c := decide(id, "DONE"); c != 200 {
			t.Fatalf("decide: %d", c)
		}
		if s := statusOf(id); s != "VERIFYING" {
			t.Fatalf("بعد تم لازم VERIFYING، طلع %s", s)
		}
	}

	var notes []string
	run := service.VerifySuggestionForTest(db, "MATRIX", func(m string) { notes = append(notes, m) })
	list, err := run(map[string]any{"action": "list"})
	if err != nil || !strings.Contains(list, a) || strings.Contains(list, other) {
		t.Fatalf("القائمة لازم بيها اقتراحاته بس: %v\n%s", err, list)
	}
	if out, _ := run(map[string]any{"action": "record", "id": other, "found": true, "note": "x"}); !strings.Contains(out, "مو من اقتراحاتك") || statusOf(other) != "VERIFYING" {
		t.Fatalf("وكيل ما يگدر يأكد اقتراح غيره: %s", out)
	}
	run(map[string]any{"action": "record", "id": a, "found": true, "note": "الأداة موجودة"})
	if s := statusOf(a); s != "DONE" {
		t.Fatalf("موجود → DONE، طلع %s", s)
	}
	run(map[string]any{"action": "record", "id": b, "found": false, "note": "ما لگيت الجدول"})
	if s := statusOf(b); s != "ACCEPTED" || len(notes) != 1 || !strings.Contains(notes[0], "ما لگيت الجدول") {
		t.Fatalf("ما موجود → ACCEPTED + إشعار: %s %v", s, notes)
	}
	if list, _ = run(map[string]any{"action": "list"}); !strings.Contains(list, "FOUND") || !strings.Contains(list, "NOT_FOUND") {
		t.Fatalf("القائمة لازم بيها آخر النتائج حتى يتعلم:\n%s", list)
	}
	// «تم» مرة ثانية يمسح النتيجة القديمة ويرجع للتأكد
	decide(a, "DONE")
	var vr *string
	db.Get(&vr, `SELECT "verifyResult" FROM "AiSuggestion" WHERE id = $1`, a)
	if statusOf(a) != "VERIFYING" || vr != nil {
		t.Fatal("تم من جديد لازم يصفّر النتيجة")
	}
}
