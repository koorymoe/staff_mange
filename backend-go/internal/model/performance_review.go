package model

import "time"

// PerformanceReview تقييم أداء منفصل تماماً عن KPI (الغرامات المالية) — يحدد فقط
// هل الموظف يستحق تدريب أو لا. سلسلة هرمية: التيم ليدر يقيّم فنييه، والإداري
// (HR_COORDINATOR) يقيّم التيم ليدر نفسه، والأدمن يقدر يقيّم أي أحد.
type PerformanceReview struct {
	ID          string    `db:"id" json:"id"`
	EmployeeID  string    `db:"employeeId" json:"employeeId"`
	EvaluatorID string    `db:"evaluatorId" json:"evaluatorId"`
	Rating      string    `db:"rating" json:"rating"` // POSITIVE | NEGATIVE
	Reason      string    `db:"reason" json:"reason"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`

	Employee  *EmployeeBrief `db:"-" json:"employee"`
	Evaluator *EmployeeBrief `db:"-" json:"evaluator"`
}

type CreatePerformanceReviewRequest struct {
	EmployeeID string `json:"employeeId"`
	Rating     string `json:"rating"`
	Reason     string `json:"reason"`
}
