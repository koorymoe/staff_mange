package model

import "time"

// ComplaintTypeLabels هي أنواع الشكاوى الثابتة المسموحة — قائمة منسدلة بدل وصف حر.
var ComplaintTypeLabels = map[string]string{
	"DELAY":           "تأخير بالتنفيذ",
	"DISORGANIZED":    "عمل غير منظم",
	"TECHNICAL":       "مشكلة فنية",
	"EXECUTION_ERROR": "خطأ تنفيذي",
	"INCOMPLETE":      "لم يتم إكمال العمل",
	"OTHER":           "أخرى",
}

type Complaint struct {
	ID                   string  `db:"id" json:"id"`
	CustomerID           string  `db:"customerId" json:"-"`
	BookingID            *string `db:"bookingId" json:"-"`
	Type                 string  `db:"type" json:"type"`
	Description          string  `db:"description" json:"description"`
	RelatedEmployeeID    *string `db:"relatedEmployeeId" json:"-"`
	RelatedEmployeeName  *string `db:"relatedEmployeeName" json:"relatedEmployeeName"`
	Status               string  `db:"status" json:"status"`
	CreatedByEmployeeID  string  `db:"createdByEmployeeId" json:"-"`
	AssignedToEmployeeID *string `db:"assignedToEmployeeId" json:"-"`
	Resolution           *string `db:"resolution" json:"resolution"`
	// حالة الاتصال بالزبون: منو اتصل ومتى، وملاحظات الزبون.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	ContactedAt   *time.Time `db:"contactedAt" json:"contactedAt"`
	ContactedByID *string    `db:"contactedById" json:"contactedById"`
	Notes         *string    `db:"notes" json:"notes"`
	// اسم الي اتصل — يُملأ بالتهيئة حتى تظهر معلوماته كدام الشكوى
	ContactedByName *string    `db:"-" json:"contactedByName"`
	CreatedAt       time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt      *time.Time `db:"resolvedAt" json:"resolvedAt"`

	// تقييم الزبون للخدمة (١..٥) — يسجّله مهندس الجودة وقت التواصل.
	// ⚠️ nullable بقصد: «ما انسأل» مو «تقييمه صفر»، والصفر يهبّط
	// المتوسط ويظلم المهندس على شكوى ماكو بيها تقييم أصلاً.
	CustomerRating *int `db:"customerRating" json:"customerRating"`
	// حكم المدقق على شغل مهندس الجودة بهذي الشكوى.
	AuditVerdict *string        `db:"auditVerdict" json:"auditVerdict"`
	AuditNote    *string        `db:"auditNote" json:"auditNote"`
	AuditedAt    *time.Time     `db:"auditedAt" json:"auditedAt"`
	AuditedByID  *string        `db:"auditedById" json:"-"`
	AuditedBy    *EmployeeBrief `db:"-" json:"auditedBy"`

	Customer           *Customer      `db:"-" json:"customer"`
	Booking            *Booking       `db:"-" json:"booking"`
	CreatedByEmployee  *EmployeeBrief `db:"-" json:"createdByEmployee"`
	AssignedToEmployee *EmployeeBrief `db:"-" json:"assignedToEmployee"`
	RelatedEmployee    *EmployeeBrief `db:"-" json:"relatedEmployee"`
}

type CreateComplaintRequest struct {
	CustomerID          string  `json:"customerId"`
	BookingID           *string `json:"bookingId"`
	Type                string  `json:"type"`
	Description         string  `json:"description"`
	RelatedEmployeeID   *string `json:"relatedEmployeeId"`
	CreatedByEmployeeID string  `json:"createdByEmployeeId"`
}

type UpdateComplaintRequest struct {
	Status               *string `json:"status"`
	AssignedToEmployeeID *string `json:"assignedToEmployeeId"`
	Resolution           *string `json:"resolution"`
}

type ResolveComplaintRequest struct {
	Resolution *string `json:"resolution"`
}

// أحكام المدقق الثلاثة — نفس الي بالتصميم.
const (
	AuditNeedsFollowUp = "NEEDS_FOLLOWUP"
	AuditRecheckRating = "RECHECK_RATING"
	AuditApproved      = "APPROVED"
)

var AuditVerdictLabels = map[string]string{
	AuditNeedsFollowUp: "مطلوب متابعة",
	AuditRecheckRating: "مراجعة التقييم",
	AuditApproved:      "معتمد",
}

// AuditComplaintRequest حكم المدقق على شغل مهندس الجودة.
type AuditComplaintRequest struct {
	Verdict string  `json:"verdict"`
	Note    *string `json:"note"`
}

// SetContactedRequest — التواصل والتقييم بطلب واحد.
//
// ⚠️ التقييم **مع** التواصل مو بمسار لحاله: مهندس الجودة يسأل
// الزبون وهو بالمكالمة، فتقسيمهن لخطوتين يعني نصف التقييمات ما
// تنسجّل. Rating فاضي مسموح — يعني ما سأل.
type SetContactedRequest struct {
	Contacted bool `json:"contacted"`
	Rating    *int `json:"rating"`
}

// ComplaintEvent سطر بسجل إجراءات الشكوى — يُكتب ولا يُعدّل.
type ComplaintEvent struct {
	ID           string    `db:"id" json:"id"`
	ComplaintID  string    `db:"complaintId" json:"complaintId"`
	Kind         string    `db:"kind" json:"kind"`
	Detail       *string   `db:"detail" json:"detail"`
	ByEmployeeID *string   `db:"byEmployeeId" json:"-"`
	ByName       *string   `db:"byName" json:"byName"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

const (
	EventCreated   = "CREATED"
	EventContacted = "CONTACTED"
	EventRated     = "RATED"
	EventNoted     = "NOTED"
	EventAssigned  = "ASSIGNED"
	EventAudited   = "AUDITED"
	EventResolved  = "RESOLVED"
)

// ComplaintCustomerStat يلخص كم مرة اشتكى زبون معيّن — تقرير منفصل عن إحصائيات الحجوزات.
type ComplaintCustomerStat struct {
	CustomerID        string `db:"customerId" json:"customerId"`
	CustomerName      string `db:"customerName" json:"customerName"`
	CustomerPhone     string `db:"customerPhone" json:"customerPhone"`
	ComplaintCount    int    `db:"complaintCount" json:"complaintCount"`
	NotContactedCount int    `db:"notContactedCount" json:"notContactedCount"`
	OpenCount         int    `db:"openCount" json:"openCount"`

	// ⚠️ AvgRating يبقى nil لمن ماكو ولا تقييم — **مو صفر**. الواجهة
	// تعرض «—»، لأن رقم مخترع أسوأ من ماكو رقم.
	AvgRating       *float64   `db:"avgRating" json:"avgRating"`
	LastContactAt   *time.Time `db:"lastContactAt" json:"lastContactAt"`
	ContactedLast30 int        `db:"contactedLast30" json:"contactedLast30"`
	NeedsAuditCount int        `db:"needsAuditCount" json:"needsAuditCount"`
	// حالة ومهندس **آخر** شكوى للزبون — الصف مجمّع، فبلا هذا
	// التوضيح تصير قراءتان مختلفتان لنفس السطر.
	LatestStatus      *string `db:"latestStatus" json:"latestStatus"`
	LatestEngineer    *string `db:"latestEngineer" json:"latestEngineer"`
	LatestComplaintID *string `db:"latestComplaintId" json:"latestComplaintId"`
}
