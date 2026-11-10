package main

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// قفل «ون بيس» (11-10): أعمدة القاعدة الي ماكو إلها مكان بموديلات Go — الرقم ينزل بس.
// عمود جديد؟ ضيفه لـstruct بوسم db:"…". صلّحت عمود قديم؟ نزّل الحد هنا.
const maxMissing = 0 // (11-10) كل عمود إله موديل بـinternal/models

func TestSchemaColumnsHaveGoModels(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	tags, err := Tags("../../internal")
	if err != nil {
		t.Fatal(err)
	}
	miss, err := Check(db, tags)
	if err != nil {
		t.Fatal(err)
	}
	if len(miss) > maxMissing {
		for _, m := range miss {
			t.Logf("%s.%s", m.Table, m.Column)
		}
		t.Fatalf("أعمدة بلا موديل Go: %d > %d — العمود الجديد لازم ينضاف لـstruct بوسم db", len(miss), maxMissing)
	}
	if len(miss) < maxMissing {
		t.Logf("👍 نزل (%d) — نزّل maxMissing", len(miss))
	}
}
