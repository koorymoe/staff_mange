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

// ComplaintCustomerStat يلخص كم مرة اشتكى زبون معيّن — تقرير منفصل عن إحصائيات الحجوزات.
type ComplaintCustomerStat struct {
	CustomerID        string `db:"customerId" json:"customerId"`
	CustomerName      string `db:"customerName" json:"customerName"`
	CustomerPhone     string `db:"customerPhone" json:"customerPhone"`
	ComplaintCount    int    `db:"complaintCount" json:"complaintCount"`
	NotContactedCount int    `db:"notContactedCount" json:"notContactedCount"`
	OpenCount         int    `db:"openCount" json:"openCount"`
}
