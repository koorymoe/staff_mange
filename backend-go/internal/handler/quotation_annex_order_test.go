package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// TestAnnexNewTemplateGoesLast: القالب الجديد (مثلاً «نسخ كقالب جديد») ما يسبق القالب الأساسي
// بالترتيب، لأن العرض الجديد ياخذ أول قالب افتراضي.
func TestAnnexNewTemplateGoesLast(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	name := "أ قالب فحص " + uuid.NewString()[:6]                                                // يبدي بـ«أ» حتى لو الترتيب بالاسم يطلع أول
	t.Cleanup(func() { db.Exec(`DELETE FROM "QuotationAnnexTemplate" WHERE name = $1`, name) }) //nolint:errcheck
	body := `{"name":"` + name + `","parts":[{"title":"جزء","items":[{"title":"بند","text":"نص"}]}]}`
	w := httptest.NewRecorder()
	NewQuotationAnnexHandler(db).Save(w, httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	var first string
	if err := db.Get(&first, `SELECT name FROM "QuotationAnnexTemplate" ORDER BY "sortOrder", name LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	if first == name {
		t.Fatal("القالب الجديد صار الأول (الافتراضي)")
	}
}
