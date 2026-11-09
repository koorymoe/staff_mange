package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
)

// ٦ح — قائمة المراحل بالقاعدة (0438) لازم تطابق projectStages بالكود: أي مرحلة
// تنضاف هنا وما تنضاف هناك = المشروع ما ينتقل لها.
func TestProjectStagesMatchDB(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	for _, s := range projectStages {
		var ok bool
		if err := db.Get(&ok, `SELECT project_stage_valid($1)`, s); err != nil || !ok {
			t.Errorf("المرحلة %q مو بقائمة القاعدة (0438) — ضيفها بترحيل جديد", s)
		}
	}
	var bogus bool
	_ = db.Get(&bogus, `SELECT project_stage_valid('مرحلة وهمية')`)
	if bogus {
		t.Error("القاعدة تقبل مرحلة وهمية")
	}
}
