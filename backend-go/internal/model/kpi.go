package model

import "time"

type KpiEvaluation struct {
	ID              string    `db:"id" json:"id"`
	EmployeeID      string    `db:"employeeId" json:"employeeId"`
	EvaluatorID     string    `db:"evaluatorId" json:"evaluatorId"`
	Points          int       `db:"points" json:"points"`
	Reason          string    `db:"reason" json:"reason"`
	DeductionAmount float64   `db:"deductionAmount" json:"deductionAmount"`
	CreatedAt       time.Time `db:"createdAt" json:"createdAt"`

	Employee  *EmployeeBrief `db:"-" json:"employee"`
	Evaluator *EmployeeBrief `db:"-" json:"evaluator"`
}

type CreateKpiEvaluationRequest struct {
	EmployeeID  string `json:"employeeId"`
	EvaluatorID string `json:"evaluatorId"`
	Points      *int   `json:"points"`
	Reason      string `json:"reason"`
}

type KpiLeaderboardEntry struct {
	EmployeeID        string `db:"employeeId" json:"employeeId"`
	EmployeeName      string `db:"employeeName" json:"employeeName"`
	Points            int    `db:"points" json:"points"`
	EvaluationCount   int    `db:"evaluationCount" json:"evaluationCount"`
	CompletedBookings int    `db:"completedBookings" json:"completedBookings"`
}

type RoleKpiLeaderboard struct {
	Role    string                 `json:"role"`
	Weekly  []KpiLeaderboardEntry  `json:"weekly"`
	Monthly []KpiLeaderboardEntry  `json:"monthly"`
}
