package model

// ProjectCandidateGroup مجموعات الموظفين المرشحين لمشروع، مرتّبة بنفس التسلسل
// الي يريده المدير بالقوائم المنسدلة: مهندسين ← تقنيين ← ليدريه ← فنيين ←
// إداريين ← مصممين.
const (
	CandidateGroupEngineers = "ENGINEERS"
	CandidateGroupTechnical = "TECHNICAL"
	CandidateGroupLeaders   = "LEADERS"
	CandidateGroupTechs     = "TECHS"
	CandidateGroupAdmins    = "ADMINS"
	CandidateGroupDesigners = "DESIGNERS"
	// خانة «الإداريون» جانت تكوّم كل واحد مو فني ولا مهندس: مبيعات
	// وحسابات وكوادر ومخازن ومسؤولي خدمات ورقابة وإدارة — اثنعش موظف
	// بخانة وحدة، فالمدير ما يلكه منو يريد. صارت خانات مفهومة.
	CandidateGroupServices = "SERVICES"
	CandidateGroupSales    = "SALES"
	CandidateGroupFinance  = "FINANCE"
	CandidateGroupHR       = "HR"
	CandidateGroupStore    = "STORE"
	CandidateGroupMonitor  = "MONITOR"
)

// ProjectCandidateGroupOrder ترتيب ظهور المجموعات بالقائمة المنسدلة.
var ProjectCandidateGroupOrder = []string{
	CandidateGroupEngineers,
	CandidateGroupTechnical,
	CandidateGroupLeaders,
	CandidateGroupTechs,
	CandidateGroupDesigners,
	CandidateGroupServices,
	CandidateGroupSales,
	CandidateGroupStore,
	CandidateGroupFinance,
	CandidateGroupHR,
	CandidateGroupMonitor,
	CandidateGroupAdmins,
}

// ProjectCandidateGroupLabels التسميات العربية الي تظهر كعناوين مجموعات.
var ProjectCandidateGroupLabels = map[string]string{
	CandidateGroupEngineers: "المهندسون",
	CandidateGroupTechnical: "التقنيون",
	CandidateGroupLeaders:   "الليدرية",
	CandidateGroupTechs:     "الفنيون",
	CandidateGroupAdmins:    "الإداريون",
	CandidateGroupDesigners: "المصممون",
	CandidateGroupServices:  "مسؤولو الخدمات",
	CandidateGroupSales:     "المبيعات",
	CandidateGroupFinance:   "الحسابات",
	CandidateGroupHR:        "الكوادر",
	CandidateGroupStore:     "المشتريات والمخازن",
	CandidateGroupMonitor:   "الرقابة",
}

// ProjectCandidate موظف مرشح لمشروع، مع مجموعته وهل هو مهندس.
// IsEngineer يعني: دوره مهندس (ENGINEER/QUALITY_ENGINEER) أو عنده مهارات
// الهندسة (تصميم/تخطيط/تنفيذ) — وهذا وحده الي يقدر يكون "المسؤول عن المشروع".
// أما "منفّذ الكشف" فأي موظف يقدر يكونه، بس القائمة تنعرض بالتسلسل أعلاه.
type ProjectCandidate struct {
	ID         string `db:"id" json:"id"`
	Name       string `db:"name" json:"name"`
	Role       string `db:"role" json:"role"`
	IsLeader   bool   `db:"isLeader" json:"isLeader"`
	HasEngSkl  bool   `db:"hasEngSkill" json:"-"`
	IsTechPerm bool   `db:"isTechPerm" json:"-"`
	Group      string `db:"-" json:"group"`
	GroupLabel string `db:"-" json:"groupLabel"`
	IsEngineer bool   `db:"-" json:"isEngineer"`
}

// ClassifyProjectCandidate يحدد مجموعة الموظف وهل يُعتبر مهندساً.
func ClassifyProjectCandidate(c *ProjectCandidate) {
	c.IsEngineer = c.Role == "ENGINEER" || c.Role == "QUALITY_ENGINEER" || c.HasEngSkl
	switch {
	case c.IsEngineer:
		c.Group = CandidateGroupEngineers
	case c.IsTechPerm:
		c.Group = CandidateGroupTechnical
	// الليدر أول — تيم ليدر يبقى بخانة «الليدرية» سواء دوره فني أو تقني
	case (c.Role == "TECHNICIAN" || c.Role == "TECHNICAL") && c.IsLeader:
		c.Group = CandidateGroupLeaders
	// TECHNICAL دور ميداني مثل الفني بس يتولى أكثر من خدمة — محله
	// «التقنيون»، مو مكوّم مع الإداريين مثل ما جان.
	case c.Role == "TECHNICAL":
		c.Group = CandidateGroupTechnical
	case c.Role == "TECHNICIAN":
		c.Group = CandidateGroupTechs
	case c.Role == "DESIGNER":
		c.Group = CandidateGroupDesigners
	case c.Role == "GPS_ADMIN" || c.Role == "SERVICE_MANAGER":
		c.Group = CandidateGroupServices
	case c.Role == "SALES":
		c.Group = CandidateGroupSales
	case c.Role == "FINANCE":
		c.Group = CandidateGroupFinance
	case c.Role == "HR_COORDINATOR":
		c.Group = CandidateGroupHR
	case c.Role == "PROCUREMENT_ADMIN":
		c.Group = CandidateGroupStore
	case c.Role == "MONITOR" || c.Role == "QUALITY_ENGINEER":
		c.Group = CandidateGroupMonitor
	default:
		c.Group = CandidateGroupAdmins
	}
	c.GroupLabel = ProjectCandidateGroupLabels[c.Group]
}
