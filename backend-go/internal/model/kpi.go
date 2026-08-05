package model

import "time"

type KpiEvaluation struct {
	ID                    string     `db:"id" json:"id"`
	EmployeeID            string     `db:"employeeId" json:"employeeId"`
	EvaluatorID           string     `db:"evaluatorId" json:"evaluatorId"`
	Points                int        `db:"points" json:"points"`
	Reason                string     `db:"reason" json:"reason"`
	DeductionAmount       float64    `db:"deductionAmount" json:"deductionAmount"`
	Cancelled             bool       `db:"cancelled" json:"cancelled"`
	CancelledAt           *time.Time `db:"cancelledAt" json:"cancelledAt"`
	CancelledByEmployeeID *string    `db:"cancelledByEmployeeId" json:"-"`
	CreatedAt             time.Time  `db:"createdAt" json:"createdAt"`

	Employee            *EmployeeBrief `db:"-" json:"employee"`
	Evaluator           *EmployeeBrief `db:"-" json:"evaluator"`
	CancelledByEmployee *EmployeeBrief `db:"-" json:"cancelledByEmployee"`
}

// KpiCriterion هي نقاط الكي بي اي القابلة للإضافة والحذف من الواجهة (صلاحية
// kpi_criteria_management منفصلة عن صلاحية تسجيل التقييمات نفسها).
type KpiCriterion struct {
	ID        string    `db:"id" json:"id"`
	Label     string    `db:"label" json:"label"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type CreateKpiCriterionRequest struct {
	Label string `json:"label"`
}

type CreateKpiEvaluationRequest struct {
	EmployeeID  string `json:"employeeId"`
	EvaluatorID string `json:"evaluatorId"`
	Points      *int   `json:"points"`
	Reason      string `json:"reason"`
	// Announce ينشر المخالفة بلوحة الإعلانات لمدة 3 أيام. يشتغل مع
	// المخالفات بس (نقاط بالسالب) — المدير يختاره وقت التسجيل.
	Announce bool `json:"announce"`
}

type KpiLeaderboardEntry struct {
	EmployeeID        string `db:"employeeId" json:"employeeId"`
	EmployeeName      string `db:"employeeName" json:"employeeName"`
	Points            int    `db:"points" json:"points"`
	EvaluationCount   int    `db:"evaluationCount" json:"evaluationCount"`
	CompletedBookings int    `db:"completedBookings" json:"completedBookings"`
}

type RoleKpiLeaderboard struct {
	Role    string                `json:"role"`
	Weekly  []KpiLeaderboardEntry `json:"weekly"`
	Monthly []KpiLeaderboardEntry `json:"monthly"`
}
