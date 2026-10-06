package service

import (
	"fmt"

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
)

var ScoreGroupLabels = map[string]string{
	GroupTechs: "الفنيين", GroupLeaders: "الليدرية", GroupSales: "المبيعات", GroupMonitors: "المراقبين",
	GroupFinance: "المحاسبة", GroupTechnical: "التقنيين ومسؤولي الخدمات", GroupCoordinator: "الإداريين", GroupOthers: "بقية الأقسام",
}

var ScoreGroupOrder = []string{GroupLeaders, GroupTechs, GroupCoordinator, GroupSales, GroupTechnical, GroupFinance, GroupMonitors, GroupOthers}

// ScoreGroup مجموعة الموظف بالتقييم. الليدر ينفصل عن الفنيين بعلامة الليدر،
// والمبيعات بالدور.
func ScoreGroup(p repository.Scorable) string {
	switch {
	case p.Role == "SALES":
		return GroupSales
	case p.Role == "MONITOR":
		return GroupMonitors
	case p.Role == "FINANCE":
		return GroupFinance
	case p.Role == "HR_COORDINATOR":
		return GroupCoordinator
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
	GroupTechs:       {"LEADER_CREW": true},
	GroupLeaders:     {"COORD_LEADER": true, "AUDIT": true, "QUALITY_CALL": true},
	GroupCoordinator: {"AUDIT": true, "QUALITY_CALL": true},
	GroupMonitors:    {"ADMIN_PERIODIC": true},
	GroupSales:       {"MONITOR_PERIODIC": true},
	GroupFinance:     {"MONITOR_PERIODIC": true},
	GroupTechnical:   {"MONITOR_PERIODIC": true},
	GroupOthers:      {"MONITOR_PERIODIC": true},
}

// periodicTargets منو يقيّم منو دورياً: المدير ← المراقبين؛ المراقب ← المجموعات
// الي ماكو إلها تقييم بمحطة (مو الفنيين ولا الليدرية ولا المراقبين).
func periodicStage(raterRole string) string {
	if raterRole == "ADMIN" || raterRole == "OWNER" {
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
func BuildReliability(rules map[string][2]int, f repository.ReliabilityFact) ([]ReliabilityPart, *float64) {
	parts := []ReliabilityPart{}
	add := func(p *ReliabilityPart) {
		if p != nil {
			parts = append(parts, *p)
		}
	}
	a := rules["ATTENDANCE"]
	add(ratioPart("ATTENDANCE", "يسجّل حضوره", a[0], a[1], "يوم سجّل"))
	if f.Sessions > 0 {
		self := f.Sessions - f.Auto
		if self < 0 {
			self = 0
		}
		parts = append(parts, ReliabilityPart{Key: "CHECKOUT", Label: "يسجّل انصرافه بنفسه",
			Pct: float64(self) * 100 / float64(f.Sessions), Detail: fmt.Sprintf("%d من %d مرة بنفسه", self, f.Sessions)})
	}
	i := rules["INVENTORY"]
	add(ratioPart("INVENTORY", "يجرد عدّته", i[0], i[1], "مرة"))
	pw := rules["PAPER"]
	add(ratioPart("PAPER", "يخلّص ورقه بوقته", pw[0], pw[1], "حجز"))
	t := rules["TASK_DUE"]
	add(ratioPart("TASK_DUE", "يخلّص مهامه بوقتها", t[0], t[1], "مهمة"))
	c := 100 - 25*f.Complaints
	if c < 0 {
		c = 0
	}
	detail := "ماكو عليه ولا شكوى"
	if f.Complaints > 0 {
		detail = fmt.Sprintf("%d شكوى من زبون", f.Complaints)
	}
	parts = append(parts, ReliabilityPart{Key: "COMPLAINTS", Label: "بلا شكاوى", Pct: float64(c), Detail: detail})
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
