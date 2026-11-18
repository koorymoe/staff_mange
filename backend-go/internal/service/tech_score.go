package service

import (
	"fmt"
	"math"

	"staffmange-api/internal/repository"
)

// ═══ التقييم الجديد للفني — الشعبة الهندسية (دازاي 11-18، docs/التقييم-الجديد.md) ═══
// الرقم = ٨٠٪ بنود + ٢٠٪ رأي ماتركس:
//   تقييم الليدر 30 · الحضور 20 · جرد العدّة 15 · المهام الإضافية 10 · الزملاء (المشاكل) 5
// البند الي ماله بيانات (مثلاً ماكو مهام) وزنه يتوزع على البقية.
// السلبي ما ينحسب إلا بعد تأكيد المسؤول — الفلترة بالاستعلامات (score_review.go).
// محطات الحجز وتقرير الإنجاز وواجب التقييم وإدارة العمليات ما تدخل برقم الفني.

type TechPart struct {
	Key    string  `json:"key"`
	Label  string  `json:"label"`
	Weight float64 `json:"weight"` // الوزن بعد التوزيع (من ١٠٠)
	Pct    float64 `json:"pct"`
	Detail string  `json:"detail"`
}

var techWeights = []struct {
	Key    string
	Label  string
	Weight float64
}{
	{"LEADER", "تقييم الليدر", 30},
	{"ATTENDANCE", "الحضور", 20},
	{"INVENTORY", "جرد العدّة", 15},
	{"TASK_DUE", "المهام الإضافية", 10},
	{"PEERS", "الزملاء (المشاكل الوظيفية)", 5},
}

const (
	techRulesShare  = 80.0
	techMatrixShare = 20.0
)

// TechMatrixOpinion رأي ماتركس (٢٠٪): يبدي ١٠٠ وينقص على الي يحتاجه الفني، ويا ملاحظة لكل وحدة.
func TechMatrixOpinion(pct map[string]float64) (float64, []string) {
	v := 100.0
	notes := []string{}
	if x, ok := pct["LEADER"]; ok && x < 70 {
		v -= 30
		notes = append(notes, "يحتاج تدريب — تقييم ليدره واطي")
	}
	if x, ok := pct["ATTENDANCE"]; ok && x < 90 {
		v -= 30
		notes = append(notes, "يحتاج يجي بوقته ويسجّل حضوره كل يوم")
	}
	if x, ok := pct["TASK_DUE"]; ok && x < 80 {
		v -= 20
		notes = append(notes, "يحتاج يكمّل مهامه أسرع")
	}
	if x, ok := pct["INVENTORY"]; ok && x < 90 {
		v -= 20
		notes = append(notes, "يحتاج ينتبه لعدّته ويجردها")
	}
	return math.Max(0, v), notes
}

// applyTechScores يبدّل رقم الفنيين بالحسبة الجديدة (قبل الفلوس والانضباط).
func applyTechScores(rows []StaffScore, hum []repository.HumanRow) {
	leader := map[string][2]int{}
	for _, h := range hum {
		if h.Stage == "LEADER_CREW" {
			v := leader[h.RateeID]
			leader[h.RateeID] = [2]int{v[0] + h.Score, v[1] + 1}
		}
	}
	for i := range rows {
		s := &rows[i]
		if s.Group != GroupTechs || s.AIOnly {
			continue
		}
		pct := map[string]float64{}
		detail := map[string]string{}
		if l := leader[s.ID]; l[1] > 0 {
			avg := float64(l[0]) / float64(l[1])
			pct["LEADER"] = avg * 20
			detail["LEADER"] = fmtStars(avg, l[1])
		}
		for _, p := range s.RelParts {
			switch p.Key {
			case "ATTENDANCE", "INVENTORY", "TASK_DUE", "PEERS":
				pct[p.Key] = p.Pct
				detail[p.Key] = p.Detail
			}
		}
		total := 0.0
		for _, w := range techWeights {
			if _, ok := pct[w.Key]; ok {
				total += w.Weight
			}
		}
		if total == 0 {
			s.TechParts = nil
			s.Final = nil
			continue
		}
		s.TechParts = []TechPart{}
		sum := 0.0
		for _, w := range techWeights {
			v, ok := pct[w.Key]
			if !ok {
				continue
			}
			weight := w.Weight / total * techRulesShare
			sum += v * weight / 100
			s.TechParts = append(s.TechParts, TechPart{Key: w.Key, Label: w.Label, Weight: weight, Pct: v, Detail: detail[w.Key]})
		}
		op, notes := TechMatrixOpinion(pct)
		s.TechParts = append(s.TechParts, TechPart{Key: "MATRIX", Label: "رأي ماتركس", Weight: techMatrixShare, Pct: op})
		s.MatrixNotes = notes
		sum += op * techMatrixShare / 100
		f := sum
		if s.ExtraBonus > 0 {
			f = math.Min(100, f+s.ExtraBonus)
		}
		s.Final = &f
		s.TechScore = true
	}
}

func fmtStars(avg float64, n int) string {
	return fmt.Sprintf("%.1f نجمة من %d تقييم", avg, n)
}
