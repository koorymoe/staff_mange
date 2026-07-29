package model

import "time"

// EngineeringSkillNames هي المهارات الأربع المشروطة قبل ما ينعطى موظف دور "مهندس"
// (ENGINEER) — يشتغل عليها كل من زرع بيانات الهندسة الأولية والتحقق وقت تغيير الدور.
var EngineeringSkillNames = []string{"تصميم", "تخطيط", "تنفيذ", "إشراف"}

// DivisionEngineering/DivisionDecor هما القيمتان الوحيدتان المسموحتان لحقل
// "division" بالموظف/الخدمة — يفصلان كادر الهندسة (كاميرات/شبكات/GPS... القديم)
// عن كادر الديكور الجديد (حدادة/نجارة/صباغة/سيراميك/لبخ/تأسيس ماء ومجاري/جبس بورد).
// ENGINEERING هي الافتراضي دايماً حتى ينسجم مع كل الموظفين والخدمات الموجودة أصلاً
// قبل هذا التقسيم (migration بالـ backfill، انظر schema_division.go).
const (
	DivisionEngineering = "ENGINEERING"
	DivisionDecor       = "DECOR"
)

type Employee struct {
	ID                   string    `db:"id" json:"id"`
	Name                 string    `db:"name" json:"name"`
	Certificate          *string   `db:"certificate" json:"certificate"`
	Position             *string   `db:"position" json:"position"`
	Phone                *string   `db:"phone" json:"phone"`
	Status               string    `db:"status" json:"status"`
	AuthzViolations      int       `db:"authzViolations" json:"-"`
	Role                 string    `db:"role" json:"role"`
	OnDuty               bool      `db:"onDuty" json:"onDuty"`
	Username             *string   `db:"username" json:"username"`
	Password             *string   `db:"password" json:"-"`
	HasDrivingLicense    bool      `db:"hasDrivingLicense" json:"hasDrivingLicense"`
	HasSafetyCertificate bool      `db:"hasSafetyCertificate" json:"hasSafetyCertificate"`
	Salary               *float64  `db:"salary" json:"salary"`
	Shift                *string   `db:"shift" json:"shift"`
	ShiftStart           *string   `db:"shiftStart" json:"shiftStart"`
	ShiftEnd             *string   `db:"shiftEnd" json:"shiftEnd"`
	MonthlyLeaves        int       `db:"monthlyLeaves" json:"monthlyLeaves"`
	JobTitle             *string   `db:"jobTitle" json:"jobTitle"`
	LeaderSkillLevel     int       `db:"leaderSkillLevel" json:"leaderSkillLevel"`
	IsLeader             bool      `db:"isLeader" json:"isLeader"`
	IsTrainee            bool      `db:"isTrainee" json:"isTrainee"`
	// Division تفصل موظفي الشعبة الهندسية (ENGINEERING، الافتراضي) عن موظفي
	// شعبة الديكور (DECOR) — تحدد أي كتالوج مهارات ينطبق عليهم، انظر
	// model.DivisionEngineering / model.DivisionDecor.
	Division             string    `db:"division" json:"division"`
	AttendanceIcon       *string   `db:"attendanceIcon" json:"attendanceIcon"`
	CreatedAt            time.Time `db:"createdAt" json:"createdAt"`

	Skills           []EmployeeSkillDetail `db:"-" json:"skills"`
	HasRequiredSkill *bool                 `db:"-" json:"hasRequiredSkill,omitempty"`
}

type EmployeeSkillDetail struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	SkillID    string    `db:"skillId" json:"skillId"`
	CanPerform bool      `db:"canPerform" json:"canPerform"`
	CreatedAt  time.Time `db:"createdAt" json:"-"`
	Skill      *Skill    `db:"-" json:"skill"`
}

type SetEmployeeSkillsRequest struct {
	Skills []EmployeeSkillInput `json:"skills"`
}

type EmployeeSkillInput struct {
	SkillID    string `json:"skillId"`
	CanPerform bool   `json:"canPerform"`
}

type CreateEmployeeRequest struct {
	Name        string   `json:"name"`
	Certificate *string  `json:"certificate"`
	Position    *string  `json:"position"`
	Phone       *string  `json:"phone"`
	Username    *string  `json:"username"`
	Password    *string  `json:"password"`
	JobTitle    *string  `json:"jobTitle"`
	Salary      *float64 `json:"salary"`
	Shift       *string  `json:"shift"`
	ShiftStart  *string  `json:"shiftStart"`
	ShiftEnd    *string  `json:"shiftEnd"`
	Role        *string  `json:"role"`
	// Division: "ENGINEERING" (افتراضي) أو "DECOR" — أول سؤال يظهر بفورم إضافة
	// موظف جديد بالواجهة، قبل ما تظهر بقية الحقول.
	Division    *string  `json:"division"`
}

type UpdateEmployeeRequest struct {
	Name                 *string  `json:"name"`
	Certificate          *string  `json:"certificate"`
	Position             *string  `json:"position"`
	Phone                *string  `json:"phone"`
	Status               *string  `json:"status"`
	Role                 *string  `json:"role"`
	OnDuty               *bool    `json:"onDuty"`
	Username             *string  `json:"username"`
	Password             *string  `json:"password"`
	HasDrivingLicense    *bool    `json:"hasDrivingLicense"`
	HasSafetyCertificate *bool    `json:"hasSafetyCertificate"`
	IsLeader             *bool    `json:"isLeader"`
	IsTrainee            *bool    `json:"isTrainee"`
	Salary               *float64 `json:"salary"`
	Shift                *string  `json:"shift"`
	ShiftStart           *string  `json:"shiftStart"`
	ShiftEnd             *string  `json:"shiftEnd"`
	MonthlyLeaves        *int     `json:"monthlyLeaves"`
	JobTitle             *string  `json:"jobTitle"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse يدمج حقول الموظف مع التوكن بنفس مستوى واحد (flatten) مطابقاً
// لـ `{ ...rest, token }` بالباك إند القديم — لا يحتاج الفرونت إند أي تعديل
// عند القراءة من أي من الباك إندين.
type LoginResponse struct {
	Employee
	Token string `json:"token"`
}
