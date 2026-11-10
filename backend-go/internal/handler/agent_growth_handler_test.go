package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// TestAgentFeedbackOwnership (دازاي 11-06): 👍/👎 لصاحب المحادثة بس، وغيره ٤٠٤.
// ومسار طلبات التحديث للقارئ: بلا رمز ٤٠٤، وبالرمز يرجّع.
func TestAgentFeedbackOwnership(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	me, other := "fb-"+uuid.NewString()[:8], "fb-"+uuid.NewString()[:8]
	db.MustExec(`INSERT INTO "Employee" (id, name, role, status) VALUES ($1, 'أ', 'ADMIN', 'ACTIVE'), ($2, 'ب', 'ADMIN', 'ACTIVE')`, me, other)
	chat, msg := uuid.NewString(), uuid.NewString()
	db.MustExec(`INSERT INTO "MatrixTalk" (id, "employeeId", title) VALUES ($1, $2, 'x')`, chat, me)
	db.MustExec(`INSERT INTO "MatrixTalkMessage" (id, "chatId", role, text) VALUES ($1, $2, 'ASSISTANT', 'جواب')`, msg, chat)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "AgentFeedback" WHERE "messageId" = $1`, msg) //nolint:errcheck
		db.Exec(`DELETE FROM "MatrixTalk" WHERE id = $1`, chat)            //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, me, other)  //nolint:errcheck
	})
	h := NewAgentGrowthHandler(db)
	send := func(emp string) int {
		r := httptest.NewRequest(http.MethodPost, "/x", bytes.NewBufferString(`{"good":false,"note":"الرقم غلط 07701234567"}`))
		r.SetPathValue("id", chat)
		r.SetPathValue("msgId", msg)
		r = r.WithContext(context.WithValue(r.Context(), middleware.ContextEmployeeID, emp))
		w := httptest.NewRecorder()
		h.MatrixFeedback(w, r)
		return w.Code
	}
	if c := send(other); c != 404 {
		t.Fatalf("غير صاحب المحادثة لازم ٤٠٤: %d", c)
	}
	if c := send(me); c != 200 {
		t.Fatalf("صاحبها: %d", c)
	}
	if c := send(me); c != 200 { // تعديل نفس التقييم
		t.Fatalf("إعادة: %d", c)
	}
	var note string
	var n int
	db.Get(&n, `SELECT count(*) FROM "AgentFeedback" WHERE "messageId" = $1`, msg)
	db.Get(&note, `SELECT note FROM "AgentFeedback" WHERE "messageId" = $1`, msg)
	if n != 1 || note == "" || bytes.Contains([]byte(note), []byte("07701234567")) {
		t.Fatalf("تقييم واحد بلا أرقام: %d %q", n, note)
	}

	tg := NewTelegramBotHandler(repository.NewBotReportRepository(db), nil, nil, nil, nil, "jwt-secret-for-test-32-characters!!")
	tg.SetReaderToken("reader-token-for-tests-0123456789abcdef")
	get := func(tok string) int {
		r := httptest.NewRequest(http.MethodGet, "/api/bot-reader/suggestions", nil)
		if tok != "" {
			r.Header.Set("X-Bot-Reader-Token", tok)
		}
		w := httptest.NewRecorder()
		tg.ReaderSuggestions(w, r)
		return w.Code
	}
	if get("") != 404 || get("wrong-token-wrong-token-wrong-token") != 404 || get("reader-token-for-tests-0123456789abcdef") != 200 {
		t.Fatal("مسار طلبات التحديث: بلا رمز ٤٠٤ وبالرمز ٢٠٠")
	}
}
