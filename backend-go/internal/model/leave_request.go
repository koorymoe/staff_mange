package model

import "time"

// طلب الإجازة: الموظف يقدّم من النظام، والموافقة تروح للشخص المخوّل حسب
// نوع كادره — مو لأي مدير.

// مسارات الموافقة = الشفتات. المسار يتحدد من شفت الموظف الطالب، مو من
// اختياره — حتى ما يقدر يوجّه طلبه لأسهل واحد يوافق.
//
// إداري الكوادر يوافق على *الشفت الي يداوم بيه هو* — أي موظف بهذا الشفت
// مهما كان دوره (فني، مصمم، إداري). يعني إداري الصباحي ما يقدر ينطي
// إجازة لموظف مسائي، والعكس. مدير النظام والمالك يوافقون على الكل.
const (
	LeaveRouteMorning = "MORNING"
	LeaveRouteEvening = "EVENING"
)

var LeaveRouteLabels = map[string]string{
	LeaveRouteMorning: "الشفت الصباحي",
	LeaveRouteEvening: "الشفت المسائي",
}

// الصلاحية المطلوبة لكل شفت. المالك ومدير النظام يوافقون على الكل بلا صلاحية.
var LeaveRoutePermission = map[string]string{
	LeaveRouteMorning: "leave_approve_morning",
	LeaveRouteEvening: "leave_approve_evening",
}

// AllLeaveRoutes كل الشفتات — للمالك ومدير النظام.
func AllLeaveRoutes() []string {
	return []string{LeaveRouteMorning, LeaveRouteEvening}
}

const (
	LeaveStatusPending   = "PENDING"
	LeaveStatusApproved  = "APPROVED"
	LeaveStatusRejected  = "REJECTED"
	LeaveStatusCancelled = "CANCELLED" // الموظف سحب طلبه قبل البت بيه
)

var LeaveStatusLabels = map[string]string{
	LeaveStatusPending:   "بانتظار الموافقة",
	LeaveStatusApproved:  "مقبولة",
	LeaveStatusRejected:  "مرفوضة",
	LeaveStatusCancelled: "ملغاة",
}

// LeaveMinNoticeDays أقل مهلة بين تقديم الطلب وأول يوم إجازة.
//
// الاتفاق: الإجازة تُطلب قبل يومين على الأقل — حتى يلحكون يرتبون الشفت
// قبل ما يغيب الموظف.
const LeaveMinNoticeDays = 2

type LeaveRequest struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	StartDate    time.Time  `db:"startDate" json:"startDate"`
	EndDate      time.Time  `db:"endDate" json:"endDate"`
	Reason       *string    `db:"reason" json:"reason"`
	Route        string     `db:"route" json:"route"`
	Status       string     `db:"status" json:"status"`
	DecidedByID  *string    `db:"decidedById" json:"decidedById"`
	DecidedAt    *time.Time `db:"decidedAt" json:"decidedAt"`
	DecisionNote *string    `db:"decisionNote" json:"decisionNote"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`

	EmployeeName  string  `db:"employeeName" json:"employeeName"`
	EmployeeRole  string  `db:"employeeRole" json:"employeeRole"`
	EmployeeShift *string `db:"employeeShift" json:"employeeShift"`
	JobTitle      *string `db:"jobTitle" json:"jobTitle"`
	DecidedByName *string `db:"decidedByName" json:"decidedByName"`

	RouteLabel  string `db:"-" json:"routeLabel"`
	StatusLabel string `db:"-" json:"statusLabel"`
	Days        int    `db:"-" json:"days"`
}

type CreateLeaveRequest struct {
	StartDate string  `json:"startDate"` // YYYY-MM-DD
	EndDate   string  `json:"endDate"`   // اختياري — لو فاضي يصير نفس البداية
	Reason    *string `json:"reason"`
}

type DecideLeaveRequest struct {
	Approve bool    `json:"approve"`
	Note    *string `json:"note"`
}

// LeaveRouteFor يحدد مسار الموافقة من شفت الموظف — الدور ما دخل بيه.
//
// الشفت الفاضي يُحسب صباحي، لأن هذا هو الافتراضي بعمود shift
// بقاعدة البيانات (DEFAULT 'MORNING').
func LeaveRouteFor(shift *string) string {
	if shift != nil && *shift == LeaveRouteEvening {
		return LeaveRouteEvening
	}
	return LeaveRouteMorning
}
