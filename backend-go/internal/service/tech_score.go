package service

import (
	"fmt"
	"math"

	"staffmange-api/internal/repository"
)

// ═══ التقييم الجديد دور دور (دازاي 11-18، docs/التقييم-الجديد.md) ═══
// كل دور: بنود بأوزان (٨٠٪) + رأي ماتركس (٢٠٪). البند الي ماله بيانات وزنه يتوزع.
// السلبي ما ينحسب إلا بعد تأكيد المسؤول — الفلترة بالاستعلامات (score_review.go).

type TechPart struct {
	Key    string  `json:"key"`
	Label  string  `json:"label"`
	Weight float64 `json:"weight"` // الوزن بعد التوزيع (من ١٠٠)
	Pct    float64 `json:"pct"`
	Detail string  `json:"detail"`
}

type roleWeight struct {
	Key    string
	Label  string
	Weight float64
}

// الفني: ليدر 30 · حضور 20 · عدّة 15 · مهام 10 · زملاء 5
var techWeights = []roleWeight{
	{"LEADER", "تقييم الليدر", 30},
	{"ATTENDANCE", "الحضور", 20},
	{"INVENTORY", "جرد العدّة", 15},
	{"TASK_DUE", "المهام الإضافية", 10},
	{"PEERS", "الزملاء (المشاكل الوظيفية)", 5},
}

// الليدر: الوقت 20 · مسؤوله 20 · الورق 15 · تقييم فنيينه 5 · حضور 10 · عدّة 5 · مهام 3 · زملاء 2
var leaderWeights = []roleWeight{
	{"TIME", "وقت الشغل (مقابل الوقت القياسي)", 20},
	{"SUPERVISOR", "تقييم مدير العمليات/المنسّق", 20},
	{"PAPER", "التقارير والفواتير (بوقتها ومطابقة للشغل)", 15},
	{"CREW", "تقييم فنيينه بعد كل حجز", 5},
	{"ATTENDANCE", "الحضور", 10},
	{"INVENTORY", "جرد العدّة", 5},
	{"TASK_DUE", "المهام الإضافية", 3},
	{"PEERS", "الزملاء (المشاكل الوظيفية)", 2},
}

var roleWeightsOf = map[string][]roleWeight{GroupTechs: techWeights, GroupLeaders: leaderWeights}

const (
	techRulesShare  = 80.0
	techMatrixShare = 20.0
	// وقت الشغل: لحد +١٥ دقيقة فوگ القياسي كاملة، لحد +٦٠ نص، أكثر صفر
	TimeGraceOK      = 15.0
	TimeGracePartial = 60.0
)

type opinionRule struct {
	Key     string
	Below   float64
	Penalty float64
	Note    string
}

var techOpinion = []opinionRule{
	{"LEADER", 70, 30, "يحتاج تدريب — تقييم ليدره واطي"},
	{"ATTENDANCE", 90, 30, "يحتاج يجي بوقته ويسجّل حضوره كل يوم"},
	{"TASK_DUE", 80, 20, "يحتاج يكمّل مهامه أسرع"},
	{"INVENTORY", 90, 20, "يحتاج ينتبه لعدّته ويجردها"},
}

var leaderOpinion = []opinionRule{
	{"TIME", 70, 30, "يحتاج يضبط وقت الشغل — حجوزاته تطوّل فوگ الوقت القياسي"},
	{"PAPER", 80, 25, "تقاريره وفواتيره تحتاج دقة ووقت"},
	{"SUPERVISOR", 70, 25, "يحتاج تدريب — تقييم مسؤوله واطي"},
	{"CREW", 80, 10, "لازم يقيّم فنيينه بعد كل حجز"},
	{"ATTENDANCE", 90, 10, "يحتاج يجي بوقته"},
}

var opinionOf = map[string][]opinionRule{GroupTechs: techOpinion, GroupLeaders: leaderOpinion}

// RoleMatrixOpinion رأي ماتركس (٢٠٪): يبدي ١٠٠ وينقص على الي يحتاجه، ويا ملاحظة لكل وحدة.
func RoleMatrixOpinion(group string, pct map[string]float64) (float64, []string) {
	v := 100.0
	notes := []string{}
	for _, r := range opinionOf[group] {
		if x, ok := pct[r.Key]; ok && x < r.Below {
			v -= r.Penalty
			notes = append(notes, r.Note)
		}
	}
	return math.Max(0, v), notes
}

// TechMatrixOpinion (للتوافق) رأي ماتركس للفني.
func TechMatrixOpinion(pct map[string]float64) (float64, []string) {
	return RoleMatrixOpinion(GroupTechs, pct)
}

// JobTimePoints حجز واحد مقابل الوقت القياسي: ٢ بالوقت، ١ متأخر شوية، ٠ متأخر هواية.
func JobTimePoints(minutes, expected float64) int {
	switch over := minutes - expected; {
	case over <= TimeGraceOK:
		return 2
	case over <= TimeGracePartial:
		return 1
	}
	return 0
}

// LeaderTimes نقاط الوقت لكل ليدر (مكسوب/أقصى + عدد المتأخرة هواية). الخدمة الي ماله وقت ما تنحسب.
func LeaderTimes(jobs []repository.LeaderJob, expected map[string]float64) map[string][3]int {
	out := map[string][3]int{}
	for _, j := range jobs {
		if j.LeaderID == nil || j.ServiceID == nil {
			continue
		}
		e, ok := expected[*j.ServiceID]
		if !ok || e <= 0 {
			continue
		}
		p := JobTimePoints(j.Minutes, e)
		v := out[*j.LeaderID]
		v[0] += p
		v[1] += 2
		if p == 0 {
			v[2]++
		}
		out[*j.LeaderID] = v
	}
	return out
}

func ratio(v [2]int) (float64, bool) {
	if v[1] == 0 {
		return 0, false
	}
	return float64(v[0]) * 100 / float64(v[1]), true
}

// applyRoleScores يبدّل رقم الأدوار الي إلها حسبة جديدة (قبل الفلوس والانضباط).
func applyRoleScores(rows []StaffScore, hum []repository.HumanRow, times map[string][3]int) {
	type agg [2]int
	byStage := map[string]map[string]agg{}
	for _, h := range hum {
		if byStage[h.RateeID] == nil {
			byStage[h.RateeID] = map[string]agg{}
		}
		v := byStage[h.RateeID][h.Stage]
		byStage[h.RateeID][h.Stage] = agg{v[0] + h.Score, v[1] + 1}
	}
	stars := func(id string, stages ...string) (float64, string, bool) {
		sum, n := 0, 0
		for _, st := range stages {
			v := byStage[id][st]
			sum, n = sum+v[0], n+v[1]
		}
		if n == 0 {
			return 0, "", false
		}
		avg := float64(sum) / float64(n)
		return avg * 20, fmt.Sprintf("%.1f نجمة من %d تقييم", avg, n), true
	}
	for i := range rows {
		s := &rows[i]
		weights, ok := roleWeightsOf[s.Group]
		if !ok || s.AIOnly {
			continue
		}
		if s.Final != nil {
			old := *s.Final
			s.LegacyBase = &old
		}
		pct := map[string]float64{}
		detail := map[string]string{}
		set := func(k string, v float64, d string) { pct[k], detail[k] = v, d }
		for _, p := range s.RelParts {
			switch p.Key {
			case "ATTENDANCE", "INVENTORY", "TASK_DUE", "PEERS":
				set(p.Key, p.Pct, p.Detail)
			}
		}
		switch s.Group {
		case GroupTechs:
			if v, d, ok := stars(s.ID, "LEADER_CREW"); ok {
				set("LEADER", v, d)
			}
		case GroupLeaders:
			if v, d, ok := stars(s.ID, "COORD_LEADER", "OPS_DAILY"); ok {
				set("SUPERVISOR", v, d)
			}
			if t := times[s.ID]; t[1] > 0 {
				d := fmt.Sprintf("%d حجز", t[1]/2)
				if t[2] > 0 {
					d += fmt.Sprintf("، %d طوّل أكثر من ساعة فوگ القياسي", t[2])
				}
				set("TIME", float64(t[0])*100/float64(t[1]), d)
			}
			r := s.ruleTotals
			if v, ok := ratio([2]int{r["PAPER"][0] + r["REPORT_MATCH"][0], r["PAPER"][1] + r["REPORT_MATCH"][1]}); ok {
				set("PAPER", v, "الفاتورة والتقرير بوقتها + تقريره يطابق شغله الفعلي")
			}
			if v, ok := ratio([2]int{r["RATING"][0] + r["RATING_DUTY"][0], r["RATING"][1] + r["RATING_DUTY"][1]}); ok {
				set("CREW", v, "قيّم كادره بعد الحجز")
			}
		}
		total := 0.0
		for _, w := range weights {
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
		for _, w := range weights {
			v, ok := pct[w.Key]
			if !ok {
				continue
			}
			weight := w.Weight / total * techRulesShare
			sum += v * weight / 100
			s.TechParts = append(s.TechParts, TechPart{Key: w.Key, Label: w.Label, Weight: weight, Pct: v, Detail: detail[w.Key]})
		}
		op, notes := RoleMatrixOpinion(s.Group, pct)
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

// applyTechScores (للتوافق ويا الاختبارات) — الفنيين بلا أوقات.
func applyTechScores(rows []StaffScore, hum []repository.HumanRow) { applyRoleScores(rows, hum, nil) }
