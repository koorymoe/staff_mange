package model

import "time"

// ServiceStudy خدمة جديدة مقترحة تحتاج دراسة قبل ما تصير خدمة رسمية بالكتالوج
// (وحدة التقنيين → إدارة الخدمات) — يفتحها المدير أو أي تقني، والمدير حصراً
// يحدد أي تقني/تقنيين موكَّلين بدراستها. كل تقني موكَّل يرفع تقارير/دراسات
// تُؤرشف بقاعدة الأرشيف الخاصة بهذا القسم.
type ServiceStudy struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	CreatedByID string    `db:"createdById" json:"-"`
	Archived    bool      `db:"archived" json:"archived"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`

	CreatedBy         *EmployeeBrief       `db:"-" json:"createdBy"`
	AssignedEmployees []EmployeeBrief      `db:"-" json:"assignedEmployees"`
	Reports           []ServiceStudyReport `db:"-" json:"reports"`
}

type ServiceStudyAssignment struct {
	ID             string    `db:"id" json:"id"`
	ServiceStudyID string    `db:"serviceStudyId" json:"serviceStudyId"`
	EmployeeID     string    `db:"employeeId" json:"employeeId"`
	CreatedAt      time.Time `db:"createdAt" json:"createdAt"`
}

type ServiceStudyReport struct {
	ID             string    `db:"id" json:"id"`
	ServiceStudyID string    `db:"serviceStudyId" json:"serviceStudyId"`
	EmployeeID     string    `db:"employeeId" json:"-"`
	Content        string    `db:"content" json:"content"`
	CreatedAt      time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type CreateServiceStudyRequest struct {
	Name string `json:"name"`
}

type AssignServiceStudyRequest struct {
	EmployeeIDs []string `json:"employeeIds"`
}

type CreateServiceStudyReportRequest struct {
	Content string `json:"content"`
}
