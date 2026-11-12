package models

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// 🏴‍☠️ ون بيس: القاعدة كلها تنبني من الموديلات بس، على قاعدة فاضية (OPBUILD_URL)،
// وتطلع مطابقة — يعني الموديلات توصف كل شي (جداول، أنواع، دوال، triggers، فهارس).
func TestRebuildFromModels(t *testing.T) {
	if os.Getenv("OPBUILD_URL") == "" {
		t.Skip("بلا OPBUILD_URL (قاعدة فاضية للبناء)")
	}
	db := sqlx.MustConnect("postgres", os.Getenv("OPBUILD_URL"))
	defer db.Close()
	have, haveE, _ := Introspect(db)
	haveO, _ := IntrospectObjects(db)
	p := DiffAll(Tables(), Enums(), Objects(), have, haveE, haveO)
	t.Logf("عبارات: %d، تحتاج موافقة: %d", len(p.Safe), len(p.NeedsApproval))
	tx := db.MustBegin()
	for _, s := range p.Safe {
		if _, err := tx.Exec(s); err != nil {
			t.Fatalf("فشل: %v\n%s", err, s[:min(len(s), 300)])
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	have, haveE, _ = Introspect(db)
	haveO, _ = IntrospectObjects(db)
	p = DiffAll(Tables(), Enums(), Objects(), have, haveE, haveO)
	if !p.Empty() {
		t.Fatalf("بعد البناء بعد أكو فرق: safe=%d\n%s\n%s", len(p.Safe), strings.Join(p.Safe[:min(5, len(p.Safe))], "\n"), strings.Join(p.NeedsApproval[:min(10, len(p.NeedsApproval))], "\n"))
	}
}
