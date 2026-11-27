package service

import (
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"
)

// ═══ ⚖️ بنود التقييم وأوزانها من الإعدادات (دازاي 11-27، المرحلة الثانية) ═══
// «كلشي يدوي يكون اعداد»: لكل مجموعة (فنيين، ليدرية، إداريين…) بنودها بأوزانها، والمالك
// يغيّر الوزن أو يطفّي البند (وزن صفر). البند المطفي وزنه يتوزع على الباقي مثل البند البلا بيانات.
// الي ما تغيّر يبقى بوزنه الأصلي بالكود — فالجدول الفاضي = نفس التقييم الحالي بالضبط.

// ScoreWeights {مجموعة: {بند: وزن}} — الي يخالف الأصل بس.
type ScoreWeights map[string]map[string]float64

func LoadScoreWeights(db *sqlx.DB) ScoreWeights {
	w := ScoreWeights{}
	if db == nil {
		return w
	}
	var raw []byte
	if err := db.Get(&raw, `SELECT weights FROM "ScoreWeightConfig" WHERE id = 1`); err == nil {
		_ = json.Unmarshal(raw, &w)
	}
	return w
}

// applyWeights بنود المجموعة بالأوزان من الإعدادات (المطفي يطلع).
func applyWeights(group string, defs []roleWeight, w ScoreWeights) []roleWeight {
	o, ok := w[group]
	if !ok || len(o) == 0 {
		return defs
	}
	out := make([]roleWeight, 0, len(defs))
	for _, d := range defs {
		if v, ok := o[d.Key]; ok {
			if v <= 0 {
				continue
			}
			d.Weight = v
		}
		out = append(out, d)
	}
	return out
}

type ScoreWeightItem struct {
	Key     string  `json:"key"`
	Label   string  `json:"label"`
	Default float64 `json:"default"`
	Weight  float64 `json:"weight"`
}

type ScoreWeightGroup struct {
	Group string            `json:"group"`
	Label string            `json:"label"`
	Items []ScoreWeightItem `json:"items"`
}

// ScoreWeightsView كل المجموعات ببنودها: الأصل والحالي.
func ScoreWeightsView(db *sqlx.DB) []ScoreWeightGroup {
	w := LoadScoreWeights(db)
	out := []ScoreWeightGroup{}
	for _, g := range ScoreGroupOrder {
		defs, ok := roleWeightsOf[g]
		if !ok {
			continue
		}
		grp := ScoreWeightGroup{Group: g, Label: ScoreGroupLabels[g], Items: []ScoreWeightItem{}}
		for _, d := range defs {
			cur := d.Weight
			if v, ok := w[g][d.Key]; ok {
				cur = v
			}
			grp.Items = append(grp.Items, ScoreWeightItem{Key: d.Key, Label: d.Label, Default: d.Weight, Weight: cur})
		}
		out = append(out, grp)
	}
	return out
}

var ErrScoreWeights = errors.New("الأوزان من 0 لـ100، ولازم يبقى بند واحد شغّال على الأقل بكل مجموعة")

// SaveScoreWeights يحفظ الي يخالف الأصل بس، ويمسح لوحات الترتيب المحفوظة.
func SaveScoreWeights(db *sqlx.DB, by string, in ScoreWeights) error {
	clean := ScoreWeights{}
	for g, items := range in {
		defs, ok := roleWeightsOf[g]
		if !ok {
			continue
		}
		on := 0
		for _, d := range defs {
			v, set := items[d.Key]
			if !set {
				v = d.Weight
			}
			if v < 0 || v > 100 {
				return ErrScoreWeights
			}
			if v > 0 {
				on++
			}
			if set && v != d.Weight {
				if clean[g] == nil {
					clean[g] = map[string]float64{}
				}
				clean[g][d.Key] = v
			}
		}
		if on == 0 {
			return ErrScoreWeights
		}
	}
	raw, _ := json.Marshal(clean)
	if _, err := db.Exec(`INSERT INTO "ScoreWeightConfig" (id, weights, "updatedById", "updatedAt") VALUES (1, $1, NULLIF($2,''), now())
		ON CONFLICT (id) DO UPDATE SET weights = $1, "updatedById" = NULLIF($2,''), "updatedAt" = now()`, string(raw), by); err != nil {
		return err
	}
	InvalidateBoard()
	return nil
}
