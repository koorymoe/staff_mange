package service

import (
	"time"

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

// (دازاي 11-19) المنسّق — شغل الحجز: التواصل والتثبيت 25 · الكادر قبل الموعد 20 · التأجيلات 10 ·
// تقييم مسؤوله 15 · حضور 5 · مهام 3 · زملاء 2. ومدير/نائب الإدارة ياخذون ويا هذا خطتهم (coordShares).
var coordWeights = []roleWeight{
	{"CONFIRM", "التواصل ويا الزبون والتثبيت بوقته", 25},
	{"CREW", "الكادر يتحدد قبل الموعد", 20},
	{"POSTPONE", "التأجيلات (كل تأجيل ينحسب)", 10},
	{"SUPERVISOR", "تقييم مسؤوله", 15},
	{"ATTENDANCE", "الحضور", 5},
	{"TASK_DUE", "المهام الإضافية", 3},
	{"PEERS", "الزملاء (المشاكل الوظيفية)", 2},
}

// moneyPart (دازاي 11-28) «هذا الموظف ما دخل فلوسه لازم يطلعلي بتقييمه»: الفلوس مقابل هدفه
// (إداري الحجوزات: هدف شفته — الحجوزات الي ثبّتها منسّقين شفته). الوزن يتغيّر من الإعدادات.
var moneyPart = roleWeight{"MONEY", "الفلوس مقابل الهدف", 25}

var roleWeightsOf = map[string][]roleWeight{
	GroupTechs:       append(techWeights, roleWeight{projectsPart.Key, projectsPart.Label, projectsPart.Weight}, moneyPart),
	GroupLeaders:     append(leaderWeights, roleWeight{projectsPart.Key, projectsPart.Label, projectsPart.Weight}, moneyPart),
	GroupCoordinator: append(coordWeights, roleWeight{projectsPart.Key, projectsPart.Label, projectsPart.Weight}, moneyPart),
}

// coordShares (دازاي 11-19) نسب المنسّق حسب موقعه بالهيكل الجديد: [شغل الحجز، خطته]، وماتركس ٢٠ دائماً.
//
//	مدير الإدارة: ٤٠ + ٤٠ · النائب: ٥٠ + ٣٠ · الإداري: ٨٠ + ٠
var coordShares = map[string][2]float64{"DIRECTOR": {40, 40}, "DEPUTY": {50, 30}}

// PostponePct كل تأجيل ينحسب: ١٠٠ − (التأجيلات ÷ الحجوزات الي ثبّتها) × ١٠٠.
func PostponePct(postpones, confirmed int) (float64, bool) {
	if confirmed == 0 {
		return 0, false
	}
	return math.Max(0, 100-float64(postpones)*100/float64(confirmed)), true
}

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
	{"MONEY", 80, 25, "ما دخّل الفلوس المطلوبة منه — أقل من هدفه"},
	{"LEADER", 70, 30, "يحتاج تدريب — تقييم ليدره واطي"},
	{"ATTENDANCE", 90, 30, "يحتاج يجي بوقته ويسجّل حضوره كل يوم"},
	{"TASK_DUE", 80, 20, "يحتاج يكمّل مهامه أسرع"},
	{"INVENTORY", 90, 20, "يحتاج ينتبه لعدّته ويجردها"},
}

var leaderOpinion = []opinionRule{
	{"MONEY", 80, 25, "ما دخّل الفلوس المطلوبة منه — أقل من هدفه"},
	{"TIME", 70, 30, "يحتاج يضبط وقت الشغل — حجوزاته تطوّل فوگ الوقت القياسي"},
	{"PAPER", 80, 25, "تقاريره وفواتيره تحتاج دقة ووقت"},
	{"SUPERVISOR", 70, 25, "يحتاج تدريب — تقييم مسؤوله واطي"},
	{"CREW", 80, 10, "لازم يقيّم فنيينه بعد كل حجز"},
	{"ATTENDANCE", 90, 10, "يحتاج يجي بوقته"},
}

var coordOpinion = []opinionRule{
	{"MONEY", 80, 25, "شفته ما طلّع الفلوس المطلوبة — أقل من هدف الشفت"},
	{"CONFIRM", 70, 30, "يحتاج يتواصل ويثبّت أسرع"},
	{"CREW", 80, 25, "الكادر يتحدد متأخر — لازم قبل الموعد"},
	{"POSTPONE", 80, 20, "تأجيلات هواية — يحتاج يضبط المواعيد"},
	{"PLAN", 70, 25, "خطة الإدارة متأخرة"},
	{"ATTENDANCE", 90, 10, "يحتاج يجي بوقته"},
}

var opinionOf = map[string][]opinionRule{GroupTechs: techOpinion, GroupLeaders: leaderOpinion, GroupCoordinator: coordOpinion}

// RoleMatrixOpinion رأي ماتركس (٢٠٪): يبدي ١٠٠ وينقص على الي يحتاجه، ويا ملاحظة لكل وحدة.
func RoleMatrixOpinion(group string, pct map[string]float64) (float64, []string) {
	return roleMatrixOpinionW(group, pct, nil)
}

// roleMatrixOpinionW (دازاي 11-28) نفسها ويا أوزان الإعدادات: البند المطفي ما ينقّص،
// والبند الي وزنه تغيّر خصمه يتغيّر بنفس النسبة (سقفه ٤٠).
func roleMatrixOpinionW(group string, pct map[string]float64, ov map[string]float64) (float64, []string) {
	def := map[string]float64{}
	for _, w := range roleWeightsOf[group] {
		def[w.Key] = w.Weight
	}
	v := 100.0
	notes := []string{}
	for _, r := range opinionOf[group] {
		pen := r.Penalty
		if nw, set := ov[r.Key]; set {
			if nw <= 0 {
				continue
			}
			if d := def[r.Key]; d > 0 {
				pen = math.Min(40, pen*nw/d)
			}
		}
		if x, ok := pct[r.Key]; ok && x < r.Below {
			v -= pen
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
// scoreWeightsLoader (11-27) يقرا أوزان الإعدادات — ينضبط ويا الخدمة؛ بالاختبارات nil = الأوزان الأصلية.
var scoreWeightsLoader func() ScoreWeights

// moneyFactsLoader (دازاي 11-28) بند الفلوس مقابل الهدف — ينضبط ويا الخدمة؛ بالاختبارات nil = ماكو بند.
var moneyFactsLoader func(from, to time.Time) map[string]MoneyFact

func applyRoleScores(rows []StaffScore, hum []repository.HumanRow, times map[string][3]int, coord ...map[string]repository.CoordFact) {
	applyRoleScoresM(rows, hum, times, nil, coord...)
}

func applyRoleScoresM(rows []StaffScore, hum []repository.HumanRow, times map[string][3]int, money map[string]MoneyFact, coord ...map[string]repository.CoordFact) {
	overrides := ScoreWeights{}
	if scoreWeightsLoader != nil {
		overrides = scoreWeightsLoader()
	}
	var cf map[string]repository.CoordFact
	if len(coord) > 0 {
		cf = coord[0]
	}
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
		// (دازاي 11-21) الديكور: نفس قواعد الفني والليدر (تقييم الليدر، الجرد، الوقت، الورق…)
		// بس يبقون بفئتهم وترتيبهم الخاص.
		rg := s.Group
		if rg == GroupDecor {
			rg = GroupTechs
			if s.IsLeader {
				rg = GroupLeaders
			}
		}
		weights, ok := roleWeightsOf[rg]
		if !ok || s.AIOnly {
			continue
		}
		// (دازاي 11-27) الأوزان من الإعدادات — البند المطفي يطلع ووزنه يتوزع
		weights = applyWeights(rg, weights, overrides)
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
		switch rg {
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
		case GroupCoordinator:
			r := s.ruleTotals
			if v, ok := ratio([2]int{r["CONTACT"][0] + r["CONFIRM"][0], r["CONTACT"][1] + r["CONFIRM"][1]}); ok {
				set("CONFIRM", v, "التواصل والتثبيت بالوقت المعتاد")
			}
			if v, ok := ratio(r["CREW"]); ok {
				set("CREW", v, "الكادر انحدد بوقته")
			}
			if f, ok := cf[s.ID]; ok {
				if v, ok := PostponePct(f.Postpones, f.Confirmed); ok {
					set("POSTPONE", v, fmt.Sprintf("%d تأجيل على %d حجز ثبّته", f.Postpones, f.Confirmed))
				}
			}
			if v, d, ok := stars(s.ID, "AUDIT", "QUALITY_CALL", "MONITOR_DAILY", "MONITOR_PERIODIC", "ADMIN_PERIODIC"); ok {
				set("SUPERVISOR", v, d)
			}
		default:
			if defs, ok := genericParts[s.Group]; ok {
				genericPct(defs, s, stars, cf, set)
			}
		}
		if rg == GroupTechs || rg == GroupLeaders || rg == GroupCoordinator {
			if m, ok := money[s.ID]; ok {
				set("MONEY", m.Pct, m.Detail)
			}
		}
		// (دازاي 11-21) المشاريع للفني والليدر والمنسّق هم — بقية الأدوار من genericParts
		if rg == GroupTechs || rg == GroupLeaders || rg == GroupCoordinator {
			if v, ok := ratio(s.ruleTotals["PROJECT_WEEK"]); ok {
				set("PROJECTS", v, fmt.Sprintf("%d من %d نقطة", s.ruleTotals["PROJECT_WEEK"][0], s.ruleTotals["PROJECT_WEEK"][1]))
			}
		}
		rulesShare, planShare := techRulesShare, 0.0
		if f, ok := cf[s.ID]; ok && s.Group == GroupCoordinator && f.Position != nil {
			if sh, ok := coordShares[*f.Position]; ok && f.PlanItems > 0 && f.Plan != nil {
				rulesShare, planShare = sh[0], sh[1]
				pct["PLAN"] = *f.Plan
				who := "خطة إدارته"
				if *f.Position == "DEPUTY" {
					who = "بنوده بالخطة"
				}
				detail["PLAN"] = fmt.Sprintf("%s: %d بند", who, f.PlanItems)
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
			weight := w.Weight / total * rulesShare
			sum += v * weight / 100
			s.TechParts = append(s.TechParts, TechPart{Key: w.Key, Label: w.Label, Weight: weight, Pct: v, Detail: detail[w.Key]})
		}
		if planShare > 0 {
			sum += pct["PLAN"] * planShare / 100
			s.TechParts = append(s.TechParts, TechPart{Key: "PLAN", Label: "الخطة وتنفيذها (الهيكل الجديد)", Weight: planShare, Pct: pct["PLAN"], Detail: detail["PLAN"]})
		}
		op, notes := roleMatrixOpinionW(rg, pct, overrides[rg])
		if defs, ok := genericParts[s.Group]; ok {
			op, notes = genericOpinion(withWeights(defs, overrides[s.Group]), pct)
		}
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
