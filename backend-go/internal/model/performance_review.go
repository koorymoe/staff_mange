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
	// الحجز الي انقيّم عليه. فاضي بالتقييمات القديمة الي انسجّلت قبل
	// ما يصير التقييم مربوط بشغل.
	BookingID *string `db:"bookingId" json:"bookingId"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`

	Employee  *EmployeeBrief `db:"-" json:"employee"`
	Evaluator *EmployeeBrief `db:"-" json:"evaluator"`
}

type CreatePerformanceReviewRequest struct {
	EmployeeID string  `json:"employeeId"`
	Rating     string  `json:"rating"`
	Reason     string  `json:"reason"`
	BookingID  *string `json:"bookingId"`
}

// ═══ حجز ينتظر تقييم كادره ═══
//
// الليدر ما يحتاج يدور على موظفيه بقائمة — النظام يگله «هذني
// الحجوزات الي خلّصتها، ومنو طلع وياك بكل وحدة».
type BookingAwaitingReview struct {
	BookingID    string     `db:"bookingId" json:"bookingId"`
	Code         string     `db:"code" json:"code"`
	CustomerName string     `db:"customerName" json:"customerName"`
	ServiceName  *string    `db:"serviceName" json:"serviceName"`
	CompletedAt  *time.Time `db:"completedAt" json:"completedAt"`

	// الكادر الي طلع بهذا الحجز، وحالة تقييم كل واحد
	Crew []CrewReviewState `db:"-" json:"crew"`
}

type CrewReviewState struct {
	EmployeeID string  `db:"employeeId" json:"employeeId"`
	Name       string  `db:"name" json:"name"`
	Position   *string `db:"position" json:"position"`
	// التقييم الي انسجّل إذا انقيّم — فاضي إذا لسه
	Rating *string `db:"rating" json:"rating"`
	Reason *string `db:"reason" json:"reason"`
}
