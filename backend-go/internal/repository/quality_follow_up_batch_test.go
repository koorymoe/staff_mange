package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
)

// TestQualityFollowUpBatchMatchesRow — (ع) 10-16: List صار بدفعات حتى
// مكتب الجودة ما ياخذ نص دقيقة. لازم يرجّع نفس الشي بالضبط الي يرجّعه
// hydrate للصف الواحد.
func TestQualityFollowUpBatchMatchesRow(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Skip("القاعدة ما تشتغل: ", err)
	}
	defer db.Close()
	r := NewQualityFollowUpRepository(db)
	batch, err := r.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) == 0 {
		t.Skip("ماكو متابعات جودة بالقاعدة")
	}
	for _, b := range batch {
		var one model.QualityFollowUp
		if err := db.Get(&one, `SELECT * FROM "QualityFollowUp" WHERE id = $1`, b.ID); err != nil {
			t.Fatal(err)
		}
		r.hydrate(&one)
		x, _ := json.Marshal(b)
		y, _ := json.Marshal(one)
		if string(x) != string(y) {
			t.Fatalf("%s: الدفعة تفرق عن الصف\nbatch=%s\nrow=%s", b.ID, x, y)
		}
	}
}
