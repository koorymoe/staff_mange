package model

import "time"

// MatrixGuideRule تعليمة توجيه لماتركس — يعدّلها المدير.
type MatrixGuideRule struct {
	ID            string     `db:"id" json:"id"`
	Route         string     `db:"route" json:"route"`
	Match         string     `db:"match" json:"match"`
	Groups        string     `db:"groups" json:"groups"`
	Text          string     `db:"text" json:"text"`
	OnlyIfPending bool       `db:"onlyIfPending" json:"onlyIfPending"`
	Priority      int        `db:"priority" json:"priority"`
	Enabled       bool       `db:"enabled" json:"enabled"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	Hits          int        `db:"hits" json:"hits"`
	LastHitAt     *time.Time `db:"lastHitAt" json:"lastHitAt"`
	Source        string     `db:"source" json:"source"`
}

// اقتراحات ماتركس.
const (
	ProposalGuideRule  = "GUIDE_RULE"
	ProposalPrediction = "PREDICTION"
	ProposalRuleTune   = "RULE_TUNE"
)

// MatrixProposal اقتراح من ماتركس ينتظر قرار المدير — ما ينفّذ بلا موافقة.
type MatrixProposal struct {
	ID          string     `db:"id" json:"id"`
	Kind        string     `db:"kind" json:"kind"`
	Title       string     `db:"title" json:"title"`
	Rationale   string     `db:"rationale" json:"rationale"`
	Evidence    NullJSON   `db:"evidence" json:"evidence"`
	Payload     NullJSON   `db:"payload" json:"payload"`
	Signature   string     `db:"signature" json:"-"`
	Source      string     `db:"source" json:"source"`
	Status      string     `db:"status" json:"status"`
	DecidedByID *string    `db:"decidedById" json:"decidedById"`
	DecidedAt   *time.Time `db:"decidedAt" json:"decidedAt"`
	Note        *string    `db:"note" json:"note"`
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
	// الموظف المعني (من payload.employeeId) — للعرض بصندوق القرارات.
	EmployeeID   *string `db:"employeeId" json:"employeeId,omitempty"`
	EmployeeName *string `db:"employeeName" json:"employeeName,omitempty"`
	EmployeeRole *string `db:"employeeRole" json:"employeeRole,omitempty"`
}
