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
)

// ProjectCandidateGroupOrder ترتيب ظهور المجموعات بالقائمة المنسدلة.
var ProjectCandidateGroupOrder = []string{
	CandidateGroupEngineers,
	CandidateGroupTechnical,
	CandidateGroupLeaders,
	CandidateGroupTechs,
	CandidateGroupAdmins,
	CandidateGroupDesigners,
}

// ProjectCandidateGroupLabels التسميات العربية الي تظهر كعناوين مجموعات.
var ProjectCandidateGroupLabels = map[string]string{
	CandidateGroupEngineers: "المهندسون",
	CandidateGroupTechnical: "التقنيون",
	CandidateGroupLeaders:   "الليدرية",
	CandidateGroupTechs:     "الفنيون",
	CandidateGroupAdmins:    "الإداريون",
	CandidateGroupDesigners: "المصممون",
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
	case c.Role == "TECHNICIAN" && c.IsLeader:
		c.Group = CandidateGroupLeaders
	case c.Role == "TECHNICIAN":
		c.Group = CandidateGroupTechs
	case c.Role == "DESIGNER":
		c.Group = CandidateGroupDesigners
	default:
		c.Group = CandidateGroupAdmins
	}
	c.GroupLabel = ProjectCandidateGroupLabels[c.Group]
}
