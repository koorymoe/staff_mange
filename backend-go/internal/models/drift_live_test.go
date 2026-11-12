package models

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// 🏴‍☠️ ون بيس: القاعدة (بعد كل التحديثات) والموديلات لازم يتطابقون بكل شي —
// جدول، عمود، نوع، NOT NULL، افتراضي، قيد، فهرس، ENUM. أي فرق = حد كتب تحديث بلا
// موديل، أو موديل بلا تحديث (cmd/makemigration يولّده).
func TestModelsMatchDatabase(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	have, haveE, err := Introspect(db)
	haveO, _ := IntrospectObjects(db)
	if err != nil {
		t.Fatal(err)
	}
	p := DiffAll(Tables(), Enums(), Objects(), have, haveE, haveO)
	for _, s := range p.Safe {
		t.Errorf("الموديل بيه شي القاعدة ما بيها (شغّل cmd/makemigration): %s", s)
	}
	for _, s := range p.NeedsApproval {
		t.Errorf("القاعدة بيها شي الموديل ما بيه (ضيفه للموديل بـinternal/models): %s", s)
	}
}
