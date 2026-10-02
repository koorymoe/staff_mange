package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type MatrixProposalRepository struct{ db *sqlx.DB }

func NewMatrixProposalRepository(db *sqlx.DB) *MatrixProposalRepository {
	return &MatrixProposalRepository{db: db}
}

// proposalRejectMemory المرفوض ما ينعاد اقتراحه هالمدة.
const proposalRejectMemory = 60

// Create يسجّل اقتراح — إلا لو نفس التوقيع معلّق، أو انرفض قريباً.
func (r *MatrixProposalRepository) Create(p model.MatrixProposal) (bool, error) {
	var blocked bool
	if err := r.db.Get(&blocked, `SELECT EXISTS (SELECT 1 FROM "MatrixProposal" WHERE signature = $1 AND
		(status = 'PENDING' OR (status IN ('REJECTED','APPROVED') AND "decidedAt" > now() - make_interval(days => $2))))`,
		p.Signature, proposalRejectMemory); err != nil || blocked {
		return false, err
	}
	var ev, pl any
	if len(p.Evidence) > 0 {
		ev = []byte(p.Evidence)
	}
	if len(p.Payload) > 0 {
		pl = []byte(p.Payload)
	}
	res, err := r.db.Exec(`INSERT INTO "MatrixProposal" (id, kind, title, rationale, evidence, payload, signature, source)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT DO NOTHING`,
		uuid.NewString(), p.Kind, p.Title, p.Rationale, ev, pl, p.Signature, p.Source)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

func (r *MatrixProposalRepository) List(status string, limit int) ([]model.MatrixProposal, error) {
	rows := []model.MatrixProposal{}
	err := r.db.Select(&rows, `SELECT * FROM "MatrixProposal" WHERE ($1 = '' OR status = $1) ORDER BY "createdAt" DESC LIMIT $2`, status, limit)
	return rows, err
}

func (r *MatrixProposalRepository) Find(id string) (*model.MatrixProposal, error) {
	var p model.MatrixProposal
	if err := r.db.Get(&p, `SELECT * FROM "MatrixProposal" WHERE id = $1`, id); err != nil {
		return nil, errors.New("الاقتراح مو موجود")
	}
	return &p, nil
}

func (r *MatrixProposalRepository) Decide(id, status, by, note string) error {
	var n *string
	if note != "" {
		n = &note
	}
	res, err := r.db.Exec(`UPDATE "MatrixProposal" SET status = $2, "decidedById" = $3, "decidedAt" = now(), note = $4
		WHERE id = $1 AND status = 'PENDING'`, id, status, by, n)
	if err != nil {
		return err
	}
	if c, _ := res.RowsAffected(); c == 0 {
		return errors.New("الاقتراح انقرر عليه من قبل")
	}
	return nil
}

func (r *MatrixProposalRepository) PendingCount() int {
	var n int
	_ = r.db.Get(&n, `SELECT COUNT(*) FROM "MatrixProposal" WHERE status = 'PENDING'`)
	return n
}

// Stats نسبة الموافقة — جزء من «دقة ماتركس».
func (r *MatrixProposalRepository) Stats() (approved, rejected int) {
	_ = r.db.Get(&approved, `SELECT COUNT(*) FROM "MatrixProposal" WHERE status = 'APPROVED' AND "decidedAt" > now() - interval '30 days'`)
	_ = r.db.Get(&rejected, `SELECT COUNT(*) FROM "MatrixProposal" WHERE status = 'REJECTED' AND "decidedAt" > now() - interval '30 days'`)
	return
}

// ═══ مادة التعلّم ═══

type KindOutcome struct {
	Kind      string `db:"kind" json:"kind"`
	Total     int    `db:"total" json:"total"`
	Escalated int    `db:"escalated" json:"escalated"`
	Resolved  int    `db:"resolved" json:"resolved"`
	Rejected  int    `db:"rejected" json:"rejected"`
}

// KindOutcomes نتائج كل نوع تذكير بآخر ٣٠ يوم.
func (r *MatrixProposalRepository) KindOutcomes() ([]KindOutcome, error) {
	rows := []KindOutcome{}
	err := r.db.Select(&rows, `SELECT kind, COUNT(*) AS total,
		COUNT(*) FILTER (WHERE "escalatedAt" IS NOT NULL) AS escalated,
		COUNT(*) FILTER (WHERE "resolvedAt" IS NOT NULL) AS resolved,
		COUNT(*) FILTER (WHERE status = 'UNDONE') AS rejected
		FROM "AiAction" WHERE "createdAt" > now() - interval '30 days' GROUP BY kind`)
	return rows, err
}

type RepeatOffender struct {
	EmployeeID string `db:"employeeId"`
	Escalated  int    `db:"escalated"`
}

// RepeatEscalations موظفين صعد عليهم ٣ مرات فأكثر بشهر.
func (r *MatrixProposalRepository) RepeatEscalations() ([]RepeatOffender, error) {
	rows := []RepeatOffender{}
	err := r.db.Select(&rows, `SELECT "targetEmployeeId" AS "employeeId", COUNT(*) AS escalated FROM "AiAction"
		WHERE "targetEmployeeId" IS NOT NULL AND "escalatedAt" IS NOT NULL AND "createdAt" > now() - interval '30 days'
		GROUP BY "targetEmployeeId" HAVING COUNT(*) >= 3`)
	return rows, err
}

// UnusedRules تعليمات فعّالة عمرها ٣٠ يوم وما تطابقت من ٣٠ يوم.
func (r *MatrixProposalRepository) UnusedRules() ([]model.MatrixGuideRule, error) {
	rows := []model.MatrixGuideRule{}
	err := r.db.Select(&rows, `SELECT * FROM "MatrixGuideRule" WHERE enabled AND route <> '/'
		AND "createdAt" < now() - interval '30 days'
		AND ("lastHitAt" IS NULL OR "lastHitAt" < now() - interval '30 days')`)
	return rows, err
}

// RecentRejections المرفوض وسببه — يندز لهايكو حتى ما يكرر.
func (r *MatrixProposalRepository) RecentRejections(limit int) ([]model.MatrixProposal, error) {
	rows := []model.MatrixProposal{}
	err := r.db.Select(&rows, `SELECT * FROM "MatrixProposal" WHERE status = 'REJECTED' ORDER BY "decidedAt" DESC LIMIT $1`, limit)
	return rows, err
}

