package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

func TestApplyWeights(t *testing.T) {
	defs := []roleWeight{{"A", "a", 30}, {"B", "b", 20}, {"C", "c", 10}}
	out := applyWeights("g", defs, ScoreWeights{"g": {"A": 50, "C": 0}})
	if len(out) != 2 || out[0].Weight != 50 || out[1].Key != "B" {
		t.Fatalf("%+v", out)
	}
	if got := applyWeights("x", defs, ScoreWeights{"g": {"A": 1}}); len(got) != 3 || got[0].Weight != 30 {
		t.Fatalf("مجموعة ثانية تأثرت: %+v", got)
	}
}

func TestScoreWeightsLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DELETE FROM "ScoreWeightConfig"`)
	if err := SaveScoreWeights(db, "", ScoreWeights{GroupTechs: {"LEADER": 40, "PEERS": 0}}); err != nil {
		t.Fatal(err)
	}
	w := LoadScoreWeights(db)
	if w[GroupTechs]["LEADER"] != 40 || w[GroupTechs]["PEERS"] != 0 {
		t.Fatalf("%+v", w)
	}
	// كل البنود مطفية = مرفوض
	all := map[string]float64{}
	for _, d := range roleWeightsOf[GroupTechs] {
		all[d.Key] = 0
	}
	if err := SaveScoreWeights(db, "", ScoreWeights{GroupTechs: all}); err == nil {
		t.Fatal("لازم يبقى بند شغّال")
	}
	// نفس الأصل ما ينخزن
	if err := SaveScoreWeights(db, "", ScoreWeights{GroupTechs: {"LEADER": 30}}); err != nil {
		t.Fatal(err)
	}
	if len(LoadScoreWeights(db)) != 0 {
		t.Fatal("الوزن الأصلي ما لازم ينخزن")
	}
	v := ScoreWeightsView(db)
	if len(v) == 0 || len(v[0].Items) == 0 {
		t.Fatal("العرض فاضي")
	}
}
