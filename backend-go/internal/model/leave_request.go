package model

import "time"

// طلب الإجازة: الموظف يقدّم من النظام، والموافقة تروح للشخص المخوّل حسب
// نوع كادره — مو لأي مدير.

// مسارات الموافقة. المسار يتحدد من الموظف الطالب نفسه، مو من اختياره —
// حتى ما يقدر يوجّه طلبه لأسهل واحد يوافق.
const (
	LeaveRouteField   = "FIELD"   // الكوادر الفنية والليدرية
	LeaveRouteEvening = "EVENING" // الكوادر المسائية
	LeaveRouteAdmin   = "ADMIN"   // الكوادر الإدارية
)

var LeaveRouteLabels = map[string]string{
	LeaveRouteField:   "الكوادر الفنية والليدرية",
	LeaveRouteEvening: "الكوادر المسائية",
	LeaveRouteAdmin:   "الكوادر الإدارية",
}

// الصلاحية المطلوبة لكل مسار. المالك يوافق على الكل بلا صلاحية.
var LeaveRoutePermission = map[string]string{
	LeaveRouteField:   "leave_approve_field",
	LeaveRouteEvening: "leave_approve_evening",
	LeaveRouteAdmin:   "leave_approve_admin",
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
// الاتفاق: الإجازة تُطلب قبل يوم أو يومين — يعني ما ينفع الموظف يطلب
// إجازة اليوم نفسه ويختفي.
const LeaveMinNoticeDays = 1

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

// LeaveRouteFor يحدد مسار الموافقة من بيانات الموظف نفسه.
//
// الترتيب مقصود: المسائي أولاً (لأن مسؤول الكوادر المسائية يغطي دوامه
// كله بغض النظر عن نوع الشغل)، بعده الفني/الليدر، وأي واحد غيرهم إداري.
func LeaveRouteFor(role string, shift *string, isLeader bool) string {
	if shift != nil && *shift == "EVENING" {
		return LeaveRouteEvening
	}
	if role == "TECHNICIAN" || isLeader {
		return LeaveRouteField
	}
	return LeaveRouteAdmin
}
