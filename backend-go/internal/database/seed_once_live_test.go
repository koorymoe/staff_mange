package database

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
)

// (دازاي 11-15) «كل تشغيل يرجّع الي غيّرته»: سعر غيّره المدير ومعيار تقييم حذفه يبقون بعد التشغيل.
func TestSeedsDoNotUndoAdminEditsLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	var price struct {
		ID    string  `db:"id"`
		Value float64 `db:"value"`
	}
	if err := db.Get(&price, `SELECT id, value FROM "SystemPriceCatalog" ORDER BY id LIMIT 1`); err != nil {
		t.Skip("ماكو كتالوج")
	}
	t.Cleanup(func() { db.Exec(`UPDATE "SystemPriceCatalog" SET value = $2 WHERE id = $1`, price.ID, price.Value) })
	db.MustExec(`UPDATE "SystemPriceCatalog" SET value = 123456 WHERE id = $1`, price.ID)
	var kpi struct {
		ID    string `db:"id"`
		Label string `db:"label"`
	}
	hasKpi := db.Get(&kpi, `SELECT id, label FROM "KpiCriterion" ORDER BY id LIMIT 1`) == nil
	if hasKpi {
		var backup []byte
		db.Get(&backup, `SELECT to_jsonb(k)::text FROM "KpiCriterion" k WHERE id = $1`, kpi.ID)
		t.Cleanup(func() {
			db.Exec(`INSERT INTO "KpiCriterion" SELECT * FROM jsonb_populate_record(NULL::"KpiCriterion", $1::jsonb) ON CONFLICT DO NOTHING`, string(backup))
		})
		if _, err := db.Exec(`DELETE FROM "KpiCriterion" WHERE id = $1`, kpi.ID); err != nil {
			hasKpi = false // مربوط بتقييمات — ما ينحذف بهالقاعدة
		}
	}
	if err := Migrate(db, "", ""); err != nil { // = إعادة تشغيل
		t.Fatal(err)
	}
	var v float64
	db.Get(&v, `SELECT value FROM "SystemPriceCatalog" WHERE id = $1`, price.ID)
	if v != 123456 {
		t.Fatalf("السعر رجع لقيمة البذرة: %v", v)
	}
	if hasKpi {
		var back bool
		db.Get(&back, `SELECT EXISTS (SELECT 1 FROM "KpiCriterion" WHERE label = $1)`, kpi.Label)
		if back {
			t.Fatal("معيار محذوف رجع بإعادة التشغيل")
		}
	}
}
