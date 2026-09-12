package model

import "time"

// ═══ سجل الأقسام ومسؤوليها ═══
//
// يخدم «الحجز داخل الشركة»: القسم ينتخب من السجل، والطالب واحد من
// مسؤولي ذاك القسم — بدل اسم قسم ينكتب بالإيد كل مرة بإملاء مختلف.
type Department struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Active    bool      `db:"active" json:"active"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`

	// المسؤولون ينتعبون بنداء منفصل — قسم بلا مسؤول يبقى صالحاً
	// للاختيار، وطالب الحجز ينكتب بالإيد.
	Heads []DepartmentHead `db:"-" json:"heads,omitempty"`
}

// DepartmentHead مسؤول قسم — «اكثر من شخص يكدر يطلب حجز لنفس القسم».
//
// ⚠️ `EmployeeID` يبقى فارغاً بهذي المرحلة: حسابات الدخول للمسؤولين
// مرحلة ثانية بقراره، والعمود موجود من هسه حتى ما نحتاج ترحيلاً ثانياً.
type DepartmentHead struct {
	ID           string    `db:"id" json:"id"`
	DepartmentID string    `db:"departmentId" json:"departmentId"`
	Name         string    `db:"name" json:"name"`
	Phone        *string   `db:"phone" json:"phone"`
	EmployeeID   *string   `db:"employeeId" json:"employeeId"`
	Active       bool      `db:"active" json:"active"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

type SaveDepartmentRequest struct {
	Name   string `json:"name"`
	Active *bool  `json:"active,omitempty"`
}

type SaveDepartmentHeadRequest struct {
	DepartmentID string  `json:"departmentId"`
	Name         string  `json:"name"`
	Phone        *string `json:"phone,omitempty"`
	Active       *bool   `json:"active,omitempty"`
}
