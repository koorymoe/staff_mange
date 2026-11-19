package service

import (
	"fmt"
	"sort"

	"staffmange-api/internal/repository"
)

// ═══ «على شنو ينقيّم» — كل موظف وليش (دازاي 11-18) ═══
// «اريد موضوع التقييم يكمل بافضل صوره… كل موظف ع شنو يتقيم وليش». ينبني من نفس جداول
// الحساب (ScoreGroup، ruleGroups، groupRaterStages، revenueKindOf) — مو نص منفصل يتقادم.

type ScoreExplainRow struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Role       string   `json:"role"`
	IsLeader   bool     `json:"isLeader"`
	Group      string   `json:"group"`
	GroupLabel string   `json:"groupLabel"`
	Why        string   `json:"why"`
	Rules      []string `json:"rules"`
	Raters     []string `json:"raters"`
	Revenue    string   `json:"revenue,omitempty"` // مصدر الفلوس، فاضي = ما تنحسب
	AIOnly     bool     `json:"aiOnly,omitempty"`
	OffSystem  bool     `json:"offSystem,omitempty"`
	RoleChange string   `json:"roleChange,omitempty"`
}

var raterStageLabels = map[string]string{
	"LEADER_CREW": "ليدره بعد كل حجز", "OPS_DAILY": "إدارة العمليات (يومي)", "COORD_LEADER": "المنسّق بعد الحجز",
	"AUDIT": "المراقب (تدقيق الحجز)", "QUALITY_CALL": "الجودة (رأي الزبون)", "MONITOR_DAILY": "المراقب (يومي على تقريره)",
	"MONITOR_PERIODIC": "المراقب (دوري)", "ADMIN_PERIODIC": "المدير/المالك (دوري)",
}

var revenueKindLabels = map[string]string{
	"BOOKED": "الحجوزات الي سجّلها", "WORK": "الشغل الي اشتغله (حصته من الكادر)", "PROJECT": "دفعات المشاريع الي هو مسؤولها",
}

var groupWhy = map[string]string{
	GroupTechs:       "(11-18) ليدره هو الي يشوف شغله: تقييم الليدر + حضوره + عدّته + مهامه + مشاكله ويا زملائه، والسلبي بس إذا أكده المسؤول.",
	GroupLeaders:     "(11-18) مسؤول عن الحجز كله: وقته مقابل الوقت القياسي للخدمة، وتقاريره وفواتيره، وتقييم فنيينه، ورأي مسؤوله — ويدخّل ١٣٠٪ من راتبه.",
	GroupDecor:       "شعبة الديكور منعزلة عن الهندسية — نفس قواعد الميدان بترتيب خاص بيها.",
	GroupCoordinator: "شغله تنظيم: سرعة الاتصال والتثبيت وتحديد الكادر. ما تنحسب عليه فلوس لأنه ينسّق مو يجيب.",
	GroupSales:       "شغله يجيب زبائن: ينقيّم على تسجيل الحجوزات والفلوس الي جابها.",
	GroupTechnical:   "شغله فني وتنسيقي سوة: الاتصال والقرار والزيارة والتشخيص والتصليح.",
	GroupFinance:     "تأخير الفواتير يأخّر فلوس الشركة، فينقيّم على سرعة التدقيق والاعتماد.",
	GroupProcurement: "يحمي المواد والعدّة: الرد على الطلبات وحل النواقص ومتابعة السيارات.",
	GroupMonitors:    "هو الي يقيّم الكل، فالمدير يقيّمه، وماتركس على سرعة حسمه.",
	GroupQuality:     "رأي الزبون لازم يوصل بسرعة: الاتصال خلال يومين من الإنجاز.",
	GroupDesign:      "التصميم المتأخر يأخّر المشروع: تسليم الأقسام بموعدها.",
	GroupProjects:    "المشروع الواگف خسارة: لازم يتحرك كل أسبوع.",
	GroupIT:          "دعم فني داخلي: القواعد العامة وتقييم المراقب.",
	GroupOffSystem:   "شغله ما ينسجّل بالنظام، فينقيّم على تقريره وصدقه وحضوره ومهامه بس.",
	GroupOthers:      "القواعد العامة وتقييم المراقب.",
}

// قواعد خاصة ببعض المجاميع ما مكتوبة بـruleGroups (تنسجّل بس لأهلها).
var groupExtraRules = map[string][]string{
	GroupProcurement: {"PROC_RESPONSE", "TOOL_RESPONSE", "SHORTAGE_RESOLVE", "VEHICLE_DAILY"},
	GroupDesign:      {"DESIGN_SECTION"},
	GroupMonitors:    {"MONITOR_AUDIT"},
}

// ExplainScores الكل (عدا المالك ومدير النظام).
func (s *MatrixScoreService) ExplainScores() ([]ScoreExplainRow, error) {
	people, err := s.repo.Scorables()
	if err != nil {
		return nil, err
	}
	out := make([]ScoreExplainRow, 0, len(people))
	for _, p := range people {
		out = append(out, ExplainOne(p))
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].GroupLabel != out[j].GroupLabel {
			return out[i].GroupLabel < out[j].GroupLabel
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// ExplainOne على شنو ينقيّم موظف واحد (منفصل حتى ينفحص بلا قاعدة).
func ExplainOne(p repository.Scorable) ScoreExplainRow {
	g := ScoreGroup(p)
	r := ScoreExplainRow{ID: p.ID, Name: p.Name, Role: p.Role, IsLeader: p.IsLeader, Group: g, GroupLabel: ScoreGroupLabels[g],
		Why: groupWhy[g], AIOnly: p.AIOnly, OffSystem: p.OffSystem, Rules: []string{}, Raters: []string{}}
	seen := map[string]bool{}
	add := func(rule string) {
		if seen[rule] || !RuleFitsGroup(rule, g) {
			return
		}
		seen[rule] = true
		r.Rules = append(r.Rules, ruleTitle(rule))
	}
	keys := make([]string, 0, len(ruleGroups))
	for k := range ruleGroups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if g != GroupOffSystem {
		for _, k := range keys {
			if ruleGroups[k][g] {
				add(k)
			}
		}
		for _, k := range groupExtraRules[g] {
			add(k)
		}
	}
	for _, k := range []string{"ATTENDANCE", "TASK_DUE", "DAILY_REPORT", "REPORT_MATCH", "RATING_DUTY"} {
		add(k)
	}
	if fieldGroups[g] {
		add("INVENTORY")
	}
	if weights, ok := roleWeightsOf[g]; ok && !p.AIOnly {
		// (دازاي 11-18) التقييم الجديد دور دور — tech_score.go
		r.Rules = []string{}
		for _, w := range weights {
			r.Rules = append(r.Rules, fmt.Sprintf("%s (%.0f)", w.Label, w.Weight*techRulesShare/100))
		}
		r.Rules = append(r.Rules, fmt.Sprintf("رأي ماتركس (%.0f): يحتاج تدريب؟ يجي بوقته؟ يكمّل مهامه أسرع؟", techMatrixShare))
		r.Raters = []string{"ليدره بعد كل حجز — السيء (١–٢) يأكده مدير الإدارة أو نائبه"}
		if g == GroupLeaders {
			r.Raters = []string{"المنسّق بعد الحجز", "إدارة العمليات (يومي)"}
		}
		if defs, ok := genericParts[g]; ok {
			r.Raters = []string{}
			for _, d := range defs {
				for _, st := range d.Stages {
					if l := raterStageLabels[st]; l != "" {
						r.Raters = append(r.Raters, l)
					}
				}
			}
			r.Revenue = ""
			if k, ok := revenueKindOf[g]; ok {
				r.Revenue = fmt.Sprintf("%s — الهدف %.0f٪ من راتبه", revenueKindLabels[k], targetOf(g))
			}
		}
		if g == GroupCoordinator {
			r.Raters = []string{"المراقب (تدقيق + يومي)", "الجودة (رأي الزبون)", "المدير/المالك (دوري)"}
			r.Rules = append(r.Rules, "مدير الإدارة: شغل الحجز ٤٠ + خطة إدارته ٤٠ · النائب: ٥٠ + بنوده بالخطة ٣٠ (الهيكل الجديد)")
		}
		if k, ok := revenueKindOf[g]; ok {
			r.Revenue = fmt.Sprintf("%s — الهدف %.0f٪ من راتبه", revenueKindLabels[k], targetOf(g))
		}
		return r
	}
	if p.AIOnly {
		r.Raters = append(r.Raters, "🤖 ماتركس بس — البشر ما يقيّمونه")
	} else {
		stages := make([]string, 0)
		for st := range groupRaterStages[g] {
			stages = append(stages, st)
		}
		sort.Strings(stages)
		for _, st := range stages {
			if l := raterStageLabels[st]; l != "" {
				r.Raters = append(r.Raters, l)
			}
		}
	}
	if k, ok := revenueKindOf[g]; ok {
		r.Revenue = revenueKindLabels[k]
	}
	if p.RoleChangedAt != nil && p.PrevRole != nil {
		r.RoleChange = *p.PrevRole + "|" + p.RoleChangedAt.Format("2006-01-02")
	}
	return r
}
