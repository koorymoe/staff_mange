package model

import "time"

// EmployeeLetter كتاب رسمي من الموظف للإدارة.
//
// ⚠️ هذي أعمدة الجدول كاملة — الجلب SELECT *، وأي عمود بلا حقل
// يفشّل الاستعلام كله بالسكوت.
type EmployeeLetter struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	AddressedTo  string     `db:"addressedTo" json:"addressedTo"`
	Subject      string     `db:"subject" json:"subject"`
	Body         string     `db:"body" json:"body"`
	Status       string     `db:"status" json:"status"` // PENDING | APPROVED | REJECTED
	DecisionNote *string    `db:"decisionNote" json:"decisionNote"`
	DecidedByID  *string    `db:"decidedById" json:"-"`
	DecidedAt    *time.Time `db:"decidedAt" json:"decidedAt"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`

	Employee  *EmployeeBrief `db:"-" json:"employee"`
	DecidedBy *EmployeeBrief `db:"-" json:"decidedBy"`
	// المسمى الوظيفي للموظف — يطلع بترويسة الكتاب المطبوع
	EmployeeJobTitle *string `db:"-" json:"employeeJobTitle"`
}

// CreateEmployeeLetterRequest طلب جديد.
type CreateEmployeeLetterRequest struct {
	AddressedTo string `json:"addressedTo"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
}

// DecideEmployeeLetterRequest جواب الإدارة.
//
// السبب إلزامي بالرفض: «مرفوض» بلا سبب ما تعلّم الموظف شي وتخلي
// الطلب يتكرر بنفس الشكل.
type DecideEmployeeLetterRequest struct {
	Approve bool   `json:"approve"`
	Note    string `json:"note"`
}

// LetterAddressees الجهات الي يقدر الموظف يوجّهلها كتابه.
var LetterAddressees = []string{
	"السيد المدير المحترم",
	"السيد المالك المحترم",
	"السيد مدير الموارد البشرية المحترم",
	"السيد المسؤول المالي المحترم",
	"السيد مسؤول الوحدة المحترم",
}
