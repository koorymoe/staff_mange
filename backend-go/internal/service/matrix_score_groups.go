package service

import (
	"fmt"
	"staffmange-api/internal/model"

	"staffmange-api/internal/repository"
)

// ═══ فصل التقييم حسب المجموعة — قرار (ع) 10-06 ═══
// «التقييم مال الفنيين لازم ينعزل عن تقييم الليدرية، والمبيعات ينعزلون عن
// الكل». كل مجموعة ترتيبها وحدها، وتقييم البشر ينحسب بس من المقيّمين
// المسموحين إلها:
//
//	الفنيين         ← الليدر مالتهم (بعد كل حجز)
//	الليدرية        ← المراقب (تدقيق) + الإداري + الجودة
//	الإداريين       ← المراقب + الجودة
//	المراقبين       ← المدير/المالك (دوري)
//	المبيعات، المحاسبة، التقنيين ومسؤولي الخدمات، والبقية ← المراقب (دوري)

const (
	GroupTechs       = "TECHS"
	GroupLeaders     = "LEADERS"
	GroupSales       = "SALES"
	GroupMonitors    = "MONITORS"
	GroupFinance     = "FINANCE"
	GroupTechnical   = "TECHNICAL"
	GroupCoordinator = "COORDINATORS"
	GroupOthers      = "OTHERS"
	GroupProcurement = "PROCUREMENT"
	// (ع) 10-11: الجودة والتصميم والآيتي كل وحدة مجموعة، والديكور منعزل عن الهندسية
	GroupQuality = "QUALITY"
	GroupDesign  = "DESIGN"
	GroupIT      = "IT"
	GroupDecor   = "DECOR"
)

var ScoreGroupLabels = map[string]string{
	GroupTechs: "الفنيين", GroupLeaders: "الليدرية", GroupSales: "المبيعات", GroupMonitors: "المراقبين",
	GroupFinance: "المحاسبة", GroupTechnical: "التقنيين ومسؤولي الخدمات", GroupCoordinator: "الإداريين", GroupProcurement: "المخازن والأسطول", GroupOthers: "بقية الأقسام",
	GroupQuality: "الجودة", GroupDesign: "التصميم", GroupIT: "تقنية المعلومات", GroupDecor: "شعبة الديكور",
}

var ScoreGroupOrder = []string{GroupLeaders, GroupTechs, GroupDecor, GroupCoordinator, GroupSales, GroupTechnical, GroupFinance, GroupProcurement, GroupMonitors, GroupQuality, GroupDesign, GroupIT, GroupOthers}

// ScoreGroup مجموعة الموظف بالتقييم. الليدر ينفصل عن الفنيين بعلامة الليدر،
// والمبيعات بالدور.
func ScoreGroup(p repository.Scorable) string {
	switch {
	// (دازاي 11-13) منسّق الحجوزات (دوره SALES بس عنده صلاحية التنسيق) شغله تثبيت
	// وتعيين كادر — ينحسب ويا الإداريين، مو ويا البياعين.
	case p.Role == "SALES" && p.IsCoordinator:
		return GroupCoordinator
	case p.Role == "SALES":
		return GroupSales
	case p.Role == "MONITOR":
		return GroupMonitors
	case p.Role == "FINANCE":
		return GroupFinance
	case p.Role == "HR_COORDINATOR":
		return GroupCoordinator
	case p.Role == "PROCUREMENT_ADMIN":
		return GroupProcurement
	// الدور يغلب علامة الليدر: ليدر صار مهندس جودة ينحسب جودة (10-11)
	case p.Role == "QUALITY_ENGINEER":
		return GroupQuality
	case p.Role == "DESIGNER":
		return GroupDesign
	case p.Role == "IT_SUPPORT":
		return GroupIT
	// شعبة الديكور (فنيها وليدرها) منعزلة عن الهندسية
	case p.Division == "DECOR" && (p.IsLeader || p.Role == "TECHNICIAN"):
		return GroupDecor
	case p.IsLeader:
		return GroupLeaders
	case p.Role == "ENGINEER" || p.Role == "TECHNICAL" || p.Role == "SERVICE_MANAGER" || p.IsServiceManager:
		return GroupTechnical
	case p.Role == "TECHNICIAN":
		return GroupTechs
	}
	return GroupOthers
}

// groupRaterStages محطات التقييم البشري الي تنحسب لكل مجموعة.
var groupRaterStages = map[string]map[string]bool{
	GroupTechs:       {"LEADER_CREW": true, "OPS_DAILY": true},
	GroupLeaders:     {"COORD_LEADER": true, "AUDIT": true, "QUALITY_CALL": true, "OPS_DAILY": true},
	GroupCoordinator: {"AUDIT": true, "QUALITY_CALL": true, "MONITOR_DAILY": true},
	GroupMonitors:    {"ADMIN_PERIODIC": true},
	GroupSales:       {"MONITOR_PERIODIC": true, "MONITOR_DAILY": true},
	GroupFinance:     {"MONITOR_PERIODIC": true, "MONITOR_DAILY": true},
	GroupTechnical:   {"MONITOR_PERIODIC": true, "MONITOR_DAILY": true},
	GroupProcurement: {"MONITOR_PERIODIC": true, "MONITOR_DAILY": true},
	GroupOthers:      {"MONITOR_PERIODIC": true, "MONITOR_DAILY": true},
	GroupQuality:     {"MONITOR_PERIODIC": true, "MONITOR_DAILY": true},
	GroupDesign:      {"MONITOR_PERIODIC": true, "MONITOR_DAILY": true},
	GroupIT:          {"MONITOR_PERIODIC": true, "MONITOR_DAILY": true},
	GroupDecor:       {"LEADER_CREW": true, "COORD_LEADER": true, "AUDIT": true, "QUALITY_CALL": true, "OPS_DAILY": true},
}

// periodicTargets منو يقيّم منو دورياً: المدير ← المراقبين؛ المراقب ← المجموعات
// الي ماكو إلها تقييم بمحطة (مو الفنيين ولا الليدرية ولا المراقبين).
func periodicStage(raterRole string) string {
	if model.IsAdminRole(raterRole) {
		return "ADMIN_PERIODIC"
	}
	return "MONITOR_PERIODIC"
}

func periodicAllowed(stage string, p repository.Scorable) bool {
	return groupRaterStages[ScoreGroup(p)][stage]
}

// ═══ 🛡️ الاعتمادية — رقم منفصل جنب التقييم ═══
// «موظف سجل حضور؟ سجل انصراف؟ ادّى العمل الي عليه، جرد أدواته؟ ماعليه
// شكوى…». كل جزء نسبة من ١٠٠، والاعتمادية معدّل الأجزاء الي تنطبق عليه.

type ReliabilityPart struct {
	Key    string  `json:"key"`
	Label  string  `json:"label"`
	Pct    float64 `json:"pct"`
	Detail string  `json:"detail"`
}

func ratioPart(key, label string, earned, max int, unit string) *ReliabilityPart {
	if max == 0 {
		return nil
	}
	return &ReliabilityPart{Key: key, Label: label, Pct: float64(earned) * 100 / float64(max),
		Detail: fmt.Sprintf("%d من %d %s", earned/2, max/2, unit)}
}

// BuildReliability من نقاط ماتركس (الحضور، الجرد، الورق، المهام) ووقائع
// الانصراف والشكاوى.
func BuildReliability(rules map[string][2]int, f repository.ReliabilityFact, group ...string) ([]ReliabilityPart, *float64) {
	// (دازاي 11-13) العدّة والجرد شغل الميدان، والشكوى من زبون لمن يتعامل ويا الزبون —
	// المصمم والمحاسب چانوا ياخذون ١٠٠٪ «محافظ على عدّته» على شي مو شغلهم.
	g := ""
	if len(group) > 0 {
		g = group[0]
	}
	field := g == "" || fieldGroups[g] || g == GroupTechnical
	customerFacing := field || officeBookingGroups[g] || g == GroupQuality
	parts := []ReliabilityPart{}
	add := func(p *ReliabilityPart) {
		if p != nil {
			parts = append(parts, *p)
		}
	}
	a := rules["ATTENDANCE"]
	add(ratioPart("ATTENDANCE", "يسجّل حضوره", a[0], a[1], "يوم سجّل"))
	// (دازاي 11-11): «ما يتحاسب ع الانصراف، يتحاسب بس ع الحضور» — عدنا انصراف تلقائي
	// بعد نهاية الشفت، فنسيان الانصراف ما ينزّل الاعتمادية. والحساب حي من البيانات،
	// فكل الأشهر السابقة رجعت لحالها بلا هذا الجزء.
	if field {
		i := rules["INVENTORY"]
		add(ratioPart("INVENTORY", "يجرد عدّته", i[0], i[1], "مرة"))
	}
	pw := rules["PAPER"]
	add(ratioPart("PAPER", "يخلّص ورقه بوقته", pw[0], pw[1], "حجز"))
	t := rules["TASK_DUE"]
	add(ratioPart("TASK_DUE", "يلتزم بتوجيهات مسؤوليه ومهامه بوقتها", t[0], t[1], "مهمة"))
	if customerFacing {
		c := 100 - 25*f.Complaints
		if c < 0 {
			c = 0
		}
		detail := "ماكو عليه ولا شكوى"
		if f.Complaints > 0 {
			detail = fmt.Sprintf("%d شكوى من زبون", f.Complaints)
		}
		parts = append(parts, ReliabilityPart{Key: "COMPLAINTS", Label: "بلا شكاوى", Pct: float64(c), Detail: detail})
	}
	// قرار (ع) 10-08: الغيابات والإجازات، المشاكل ويا الزملاء، العدّة، الـKPI.
	penalty := func(key, label string, n, per int, zero, unit string) {
		v := 100 - per*n
		if v < 0 {
			v = 0
		}
		d := zero
		if n > 0 {
			d = fmt.Sprintf("%d %s", n, unit)
		}
		parts = append(parts, ReliabilityPart{Key: key, Label: label, Pct: float64(v), Detail: d})
	}
	extraLeave := f.LeaveDays - 2 // يومين بالشهر طبيعي
	if extraLeave < 0 {
		extraLeave = 0
	}
	leaveDetail := "ماكو إجازات"
	if f.LeaveDays > 0 {
		leaveDetail = fmt.Sprintf("%d يوم إجازة", f.LeaveDays)
	}
	parts = append(parts, ReliabilityPart{Key: "LEAVES", Label: "ماعنده غيابات وإجازات كثيرة",
		Pct: float64(max(0, 100-10*extraLeave)), Detail: leaveDetail})
	penalty("PEERS", "ماعنده مشاكل ويا زملائه", f.Issues, 25, "ماكو مشاكل", "مشكلة وظيفية")
	if field {
		penalty("TOOLS", "محافظ على عدّته", f.ToolLosses, 25, "ماكو ضياع ولا تلف", "عدّة ضاعت أو تلفت")
	}
	penalty("KPI", "تقييم الـKPI", f.KpiHits, 20, "ماكو عليه نقاط", "نقطة KPI عليه")
	if len(parts) <= 1 && f.Complaints == 0 && a[1] == 0 {
		return parts, nil // ماكو بيانات كافية
	}
	sum := 0.0
	for _, p := range parts {
		sum += p.Pct
	}
	v := sum / float64(len(parts))
	return parts, &v
}

// ═══ «محد يتقيّم ع شغل مو شغله» — دازاي 11-13 ═══
// كل محطة بالحجز إلها أهلها. إذا واحد من غير أهلها سوّاها (فني سجّل حجز،
// بياع ثبّت…) ما تنحسب له ولا عليه — لا تنزّله ولا ترفعه. القواعد العامة
// (الحضور، التقرير اليومي، المهام…) والقواعد الخاصة بدور واحد تنحسب عادي.
var fieldGroups = map[string]bool{GroupLeaders: true, GroupTechs: true, GroupDecor: true}
var officeBookingGroups = map[string]bool{GroupSales: true, GroupCoordinator: true, GroupTechnical: true}

var ruleGroups = map[string]map[string]bool{
	"CREATE":       officeBookingGroups,
	"CONTACT":      {GroupCoordinator: true, GroupTechnical: true},
	"HANDOFF":      {GroupCoordinator: true, GroupTechnical: true},
	"CONFIRM":      {GroupCoordinator: true, GroupTechnical: true},
	"CREW":         {GroupCoordinator: true, GroupTechnical: true},
	"TECH_CONTACT": {GroupTechnical: true},
	"TECH_DECIDE":  {GroupTechnical: true},
	"TECH_VISIT":   {GroupTechnical: true},
	"TECH_DIAG":    {GroupTechnical: true},
	"TECH_FIX":     {GroupTechnical: true},
	"RECEIVE":      fieldGroups,
	"ROUTE":        fieldGroups,
	"WORK":         fieldGroups,
	"COMPLETE":     fieldGroups,
	"PAPER":        fieldGroups,
	"STAGE_LOG":    fieldGroups,
	"RATING":       {GroupLeaders: true, GroupDecor: true},
	"ACCOUNT":      {GroupFinance: true},
	"QUALITY":      {GroupQuality: true},
	"MONITOR":      {GroupMonitors: true},
}

// RuleFitsGroup هل هالقاعدة من شغل هالمجموعة؟ قاعدة مو بالجدول = تنطبق.
func RuleFitsGroup(rule, group string) bool {
	g, ok := ruleGroups[rule]
	return !ok || group == "" || g[group]
}
