package repository

import (
	"errors"
	"staffmange-api/internal/errlog"
	"time"

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
	err := r.db.Select(&rows, `SELECT p.*, e.id AS "employeeId", e.name AS "employeeName", e.role::text AS "employeeRole"
		FROM "MatrixProposal" p
		LEFT JOIN "Employee" e ON e.id = p.payload->>'employeeId'
		WHERE ($1 = '' OR p.status = $1) ORDER BY p."createdAt" DESC LIMIT $2`, status, limit)
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
	errlog.Err(r.db.Get(&n, `SELECT COUNT(*) FROM "MatrixProposal" WHERE status = 'PENDING'`))
	return n
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

// ExpireStalePredictions توقع يوم فات ما إله معنى بعد — ينتهي بدل ما يضل معلّق
// بصندوق القرارات («منذ ٢٤ ساعة… باقي 410»). وتراجع سرعة بأرقام مستحيلة
// (أكثر من ×٥ المتوقع) ينتهي هم — بيانات أوقات غلط مو أداء.
func (r *MatrixProposalRepository) ExpireStalePredictions(today string) {
	errlog.Exec(r.db.Exec(`UPDATE "MatrixProposal" SET status = 'EXPIRED', "decidedAt" = now(),
		note = 'انتهى تلقائياً — توقع يوم فات أو أرقام غير منطقية'
		WHERE status = 'PENDING' AND (
		  (signature LIKE 'PRED|%' AND signature NOT LIKE '%|' || $1)
		  OR (signature LIKE 'DECLINE|%' AND (COALESCE((evidence->>'ratio')::float, 0) > 5 OR COALESCE((evidence->>'prevRatio')::float, 0) > 5))
		)`, today))
}

// ═══ (ل) 11-12: دقة توقعات «ما راح يخلّص اليوم» ═══

// PredictionToCheck توقع يوم معيّن بعده ما انحكم.
type PredictionToCheck struct {
	ID       string         `db:"id"`
	Evidence model.NullJSON `db:"evidence"`
	Payload  model.NullJSON `db:"payload"`
}

// PredictionsToVerify توقعات اليوم (أي حالة: انوافق عليها، انرفضت، أو انتهت) الي ما انحكمت.
func (r *MatrixProposalRepository) PredictionsToVerify(day string) ([]PredictionToCheck, error) {
	rows := []PredictionToCheck{}
	err := r.db.Select(&rows, `SELECT id, evidence, payload FROM "MatrixProposal"
		WHERE kind = 'PREDICTION' AND signature LIKE 'PRED|%|' || $1 AND verdict IS NULL`, day)
	return rows, err
}

func (r *MatrixProposalRepository) SetPredictionVerdict(id string, correct bool, leftAtEnd int) {
	v := "WRONG"
	if correct {
		v = "CORRECT"
	}
	errlog.Exec(r.db.Exec(`UPDATE "MatrixProposal" SET verdict = $2, "verifiedAt" = now(), "leftAtEnd" = $3 WHERE id = $1`, id, v, leftAtEnd))
}

// PredictionVerdictRow توقع انحكم — للإثبات بالشاشة.
type PredictionVerdictRow struct {
	Title     string    `db:"title" json:"title"`
	Verdict   string    `db:"verdict" json:"verdict"`
	LeftAtEnd *int      `db:"leftAtEnd" json:"leftAtEnd"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

// PredictionAccuracy آخر `days` يوم: كم توقع صح من كم انحكم + آخر ٢٠ للإثبات.
func (r *MatrixProposalRepository) PredictionAccuracy(days int) (correct, judged int, recent []PredictionVerdictRow, err error) {
	err = r.db.QueryRow(`SELECT count(*) FILTER (WHERE verdict = 'CORRECT'), count(*) FROM "MatrixProposal"
		WHERE kind = 'PREDICTION' AND verdict IS NOT NULL AND "createdAt" > now() - make_interval(days => $1)`, days).Scan(&correct, &judged)
	if err != nil {
		return
	}
	recent = []PredictionVerdictRow{}
	err = r.db.Select(&recent, `SELECT title, verdict, "leftAtEnd", "createdAt" FROM "MatrixProposal"
		WHERE kind = 'PREDICTION' AND verdict IS NOT NULL AND "createdAt" > now() - make_interval(days => $1)
		ORDER BY "createdAt" DESC LIMIT 20`, days)
	return
}
