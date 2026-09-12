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
	ID          string  `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	Certificate *string `db:"certificate" json:"certificate"`
	// مسار صورة الموظف (`/api/files/...`) — فاضي = نعرض أول حرف اسمه
	PhotoURL        *string `db:"photoUrl" json:"photoUrl"`
	Position        *string `db:"position" json:"position"`
	Phone           *string `db:"phone" json:"phone"`
	Status          string  `db:"status" json:"status"`
	AuthzViolations int     `db:"authzViolations" json:"-"`
	// الحظر التلقائي — لازم تكون موجودة بالموديل وإلا sqlx يفشل بـSELECT *
	// ("missing destination name") ويطلع الخطأ كأنه كلمة مرور غلط.
	FailedLoginStreak int        `db:"failedLoginStreak" json:"-"`
	LockedAt          *time.Time `db:"lockedAt" json:"lockedAt"`
	LockedReason      *string    `db:"lockedReason" json:"lockedReason"`
	LockedDetail      *string    `db:"lockedDetail" json:"lockedDetail"`
	// أي توكن صدر قبل هذي اللحظة يُرفض (تغيير كلمة سر / حظر / إنهاء جلسات)
	SessionsInvalidatedAt *time.Time `db:"sessionsInvalidatedAt" json:"-"`
	Role                  string     `db:"role" json:"role"`
	OnDuty                bool       `db:"onDuty" json:"onDuty"`
	Username              *string    `db:"username" json:"username"`
	Password              *string    `db:"password" json:"-"`
	// باسورد مركز القيادة — منفصل تماماً عن العادي حتى تنعزل الطبقتين:
	// لو انسرق العادي، مركز القيادة يضل مقفول.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (الجلب SELECT *). وjson:"-"
	// إجباري — ما يطلع بأي رد أبداً.
	CommandPassword      *string    `db:"commandPassword" json:"-"`
	CommandPasswordSetAt *time.Time `db:"commandPasswordSetAt" json:"-"`
	HasDrivingLicense    bool       `db:"hasDrivingLicense" json:"hasDrivingLicense"`
	HasSafetyCertificate bool       `db:"hasSafetyCertificate" json:"hasSafetyCertificate"`
	Salary               *float64   `db:"salary" json:"salary"`
	Shift                *string    `db:"shift" json:"shift"`
	ShiftStart           *string    `db:"shiftStart" json:"shiftStart"`
	ShiftEnd             *string    `db:"shiftEnd" json:"shiftEnd"`
	MonthlyLeaves        int        `db:"monthlyLeaves" json:"monthlyLeaves"`
	JobTitle             *string    `db:"jobTitle" json:"jobTitle"`
	LeaderSkillLevel     int        `db:"leaderSkillLevel" json:"leaderSkillLevel"`
	IsLeader             bool       `db:"isLeader" json:"isLeader"`
	IsTrainee            bool       `db:"isTrainee" json:"isTrainee"`
	// Division تفصل موظفي الشعبة الهندسية (ENGINEERING، الافتراضي) عن موظفي
	// شعبة الديكور (DECOR) — تحدد أي كتالوج مهارات ينطبق عليهم، انظر
	// model.DivisionEngineering / model.DivisionDecor.
	Division       string    `db:"division" json:"division"`
	AttendanceIcon *string   `db:"attendanceIcon" json:"attendanceIcon"`
	CreatedAt      time.Time `db:"createdAt" json:"createdAt"`

	// ═══ ملف الموارد البشرية (منقول من نظام الطاقة الشمسية) ═══
	// ⚠️ هذولا أعمدة بالجدول → لازم حقول هنا. الجلب يستعمل SELECT * وأي
	// عمود بلا حقل يفشّل الاستعلام كله بالسكوت — حتى تسجيل الدخول يوكف
	// وما يطلع ولا خطأ بالكونسول، بس «اسم المستخدم أو كلمة المرور غير
	// صحيحة». صارت مرتين، فلا تضيف عمود على "Employee" بدون حقل هنا.
	Department      *string    `db:"department" json:"department"`
	HireDate        *time.Time `db:"hireDate" json:"hireDate"`
	ExperienceYears *float64   `db:"experienceYears" json:"experienceYears"`
	LastReview      *string    `db:"lastReview" json:"lastReview"`
	CareerStatus    string     `db:"careerStatus" json:"careerStatus"`
	// JobLevel المستوى الوظيفي ١-١٠ — غير leaderSkillLevel (درجة مهارة
	// الليدر). افتراضيته ٥ حتى ما ينوسم موظف بتقييم ما انعطى له.
	JobLevel      int     `db:"jobLevel" json:"jobLevel"`
	NextRole      *string `db:"nextRole" json:"nextRole"`
	TrainingNeeds *string `db:"trainingNeeds" json:"trainingNeeds"`

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
	Division *string `json:"division"`
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
	PhotoURL             *string  `json:"photoUrl"`
	Salary               *float64 `json:"salary"`
	Shift                *string  `json:"shift"`
	ShiftStart           *string  `json:"shiftStart"`
	ShiftEnd             *string  `json:"shiftEnd"`
	MonthlyLeaves        *int     `json:"monthlyLeaves"`
	JobTitle             *string  `json:"jobTitle"`

	// ملف الموارد البشرية — منقول من نظام الطاقة الشمسية
	Department      *string  `json:"department"`
	HireDate        *string  `json:"hireDate"`
	ExperienceYears *float64 `json:"experienceYears"`
	LastReview      *string  `json:"lastReview"`
	CareerStatus    *string  `json:"careerStatus"`
	JobLevel        *int     `json:"jobLevel"`
	NextRole        *string  `json:"nextRole"`
	TrainingNeeds   *string  `json:"trainingNeeds"`
}

// ═══ أهلية الترقية ═══
//
// نفس قواعد النظام القديم، بس محسوبة بالسيرفر مو بالمتصفح — القاعدة
// وحدة للكل ولا تنغيّر لو أحد فتح أدوات المطوّر.
//
//	خبرة ≥٣ سنوات و مستوى ≥٧ و تقييم ممتاز/جيد جداً  →  يحتاج ترقية
//	تقييم «يحتاج تحسين» أو مستوى <٤                    →  يحتاج تدريب
//	غيرها                                              →  مستقر
const (
	CareerStable     = "مستقر"
	CareerNeedsPromo = "يحتاج ترقية"
	CareerNeedsTrain = "يحتاج تدريب"
	CareerWatched    = "تحت المراقبة"
)

// EvaluateCareerStatus يحسب الحالة الوظيفية من الخبرة والمستوى والتقييم.
func EvaluateCareerStatus(experienceYears float64, level int, lastReview string) string {
	if experienceYears >= 3 && level >= 7 && (lastReview == "ممتاز" || lastReview == "جيد جداً") {
		return CareerNeedsPromo
	}
	if lastReview == "يحتاج تحسين" || level < 4 {
		return CareerNeedsTrain
	}
	return CareerStable
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
	// Realm: staff = نظام الشركة، command = مركز القيادة.
	Realm string `json:"realm,omitempty"`
}

// LockedEmployee حساب محظور تلقائياً — يظهر بلوحة المراقبة مال المالك مع
// سبب الحظر، والمالك وحده يقدر يفكّه.
type LockedEmployee struct {
	ID                string     `db:"id" json:"id"`
	Name              string     `db:"name" json:"name"`
	Username          *string    `db:"username" json:"username"`
	Role              string     `db:"role" json:"role"`
	LockedAt          *time.Time `db:"lockedAt" json:"lockedAt"`
	LockedReason      *string    `db:"lockedReason" json:"lockedReason"`
	LockedDetail      *string    `db:"lockedDetail" json:"lockedDetail"`
	FailedLoginStreak int        `db:"failedLoginStreak" json:"failedLoginStreak"`
	AuthzViolations   int        `db:"authzViolations" json:"authzViolations"`
}

// SecurityEvent حدث أمني مسجّل (محاولة دخول فاشلة، محاولة وصول غير مخوّلة،
// حظر، فك حظر، منح صلاحيات...).
type SecurityEvent struct {
	ID           string    `db:"id" json:"id"`
	EmployeeID   *string   `db:"employeeId" json:"employeeId"`
	EmployeeName *string   `db:"employeeName" json:"employeeName"`
	Kind         string    `db:"kind" json:"kind"`
	Detail       *string   `db:"detail" json:"detail"`
	IP           *string   `db:"ip" json:"ip"`
	UserAgent    *string   `db:"userAgent" json:"userAgent"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

// ═══ مهارات القيادة ═══
//
// ⚠️ القائمة **مصدر واحد** بالخادم: الواجهة چانت تعرّفها بنفسها،
// فأي اسم ينكتب غلط يخزن مهارة ما تنقرا بأي مكان ثاني.
var LeaderSkills = []string{
	"القيادة", "إدارة الفريق", "حل المشكلات", "التواصل", "اتخاذ القرار",
}

func IsLeaderSkill(s string) bool {
	for _, k := range LeaderSkills {
		if k == s {
			return true
		}
	}
	return false
}

type LeaderSkillRating struct {
	ID          string    `db:"id" json:"id"`
	EmployeeID  string    `db:"employeeId" json:"employeeId"`
	Skill       string    `db:"skill" json:"skill"`
	Score       int       `db:"score" json:"score"`
	RatedByID   *string   `db:"ratedById" json:"-"`
	RatedByName *string   `db:"ratedByName" json:"ratedByName"`
	RatedAt     time.Time `db:"ratedAt" json:"ratedAt"`
}

// SetLeaderSkillsRequest الدرجات بطلب واحد: {"القيادة": 7, ...}
type SetLeaderSkillsRequest struct {
	Scores map[string]int `json:"scores"`
}
