package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// يحتاج قاعدة حية: الملف ينبني لكل موظف فعّال، وما بيه راتب.
func TestDossier_Live(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	ids := []string{}
	_ = db.Select(&ids, `SELECT id FROM "Employee" WHERE status = 'ACTIVE'`)
	for _, id := range ids {
		d, err := BuildDossier(db, id)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if len(d.Months) != 6 {
			t.Fatalf("%s: months %d", id, len(d.Months))
		}
		txt := d.Text()
		if strings.Contains(txt, "راتب") {
			t.Fatalf("الملف بيه راتب: %s", id)
		}
		if id == os.Getenv("DOSSIER_SHOW") {
			t.Log("\n" + txt)
		}
	}
}
