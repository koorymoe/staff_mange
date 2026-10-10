package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ دفعات المشاريع ═══

type ProjectPaymentRepository struct{ db *sqlx.DB }

func NewProjectPaymentRepository(db *sqlx.DB) *ProjectPaymentRepository {
	return &ProjectPaymentRepository{db: db}
}

type ProjectPayment struct {
	ID           string     `db:"id" json:"id"`
	ProjectID    string     `db:"projectId" json:"projectId"`
	Amount       float64    `db:"amount" json:"amount"`
	PaidAt       time.Time  `db:"paidAt" json:"paidAt"`
	Method       string     `db:"method" json:"method"`
	ReceiptNo    *string    `db:"receiptNo" json:"receiptNo"`
	Note         *string    `db:"note" json:"note"`
	CreatedBy    *string    `db:"createdBy" json:"createdBy"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
	VerifiedBy   *string    `db:"verifiedBy" json:"verifiedBy"`
	VerifiedAt   *time.Time `db:"verifiedAt" json:"verifiedAt"`
	CancelledAt  *time.Time `db:"cancelledAt" json:"cancelledAt"`
	CancelReason *string    `db:"cancelReason" json:"cancelReason"`
}

// ProjectMoney ملخّص فلوس مشروع.
type ProjectMoney struct {
	ProjectID     string   `db:"projectId" json:"projectId"`
	Code          string   `db:"code" json:"code"`
	Name          string   `db:"name" json:"name"`
	Stage         string   `db:"stage" json:"stage"`
	PriceText     *string  `db:"price" json:"priceText"`
	ContractValue *float64 `db:"contractValue" json:"contractValue"` // المسجّل كرقم، وإلا من نص السعر
	ValueFromText bool     `db:"valueFromText" json:"valueFromText"`
	Paid          float64  `db:"paid" json:"paid"`
	Unverified    float64  `db:"unverified" json:"unverified"`
	Payments      int      `db:"payments" json:"payments"`
	LastPaidAt    *string  `db:"lastPaidAt" json:"lastPaidAt"`
	OwnerID       *string  `db:"ownerId" json:"ownerId"`
	OwnerName     *string  `db:"ownerName" json:"ownerName"`
}

const projectPriceNum = `NULLIF(regexp_replace(COALESCE(p.price, ''), '\D', '', 'g'), '')::numeric`

const projectMoneySQL = `
	SELECT p.id AS "projectId", p.code, p.name, p.stage, p.price,
	       COALESCE(cv.amount, ` + projectPriceNum + `)::float8 AS "contractValue",
	       (cv.amount IS NULL AND ` + projectPriceNum + ` IS NOT NULL) AS "valueFromText",
	       COALESCE(pp.paid, 0)::float8 AS paid, COALESCE(pp.unverified, 0)::float8 AS unverified,
	       COALESCE(pp.n, 0) AS payments, to_char(pp.last, 'YYYY-MM-DD') AS "lastPaidAt",
	       COALESCE(p."delegatedToEmployeeId", p."responsibleEmployeeId", p."surveyorEmployeeId") AS "ownerId",
	       oe.name AS "ownerName"
	FROM "Project" p
	LEFT JOIN "ProjectContractValue" cv ON cv."projectId" = p.id
	LEFT JOIN LATERAL (
		SELECT SUM(amount) AS paid, SUM(amount) FILTER (WHERE "verifiedAt" IS NULL) AS unverified,
		       count(*) AS n, max("paidAt") AS last
		FROM "ProjectPayment" WHERE "projectId" = p.id AND "cancelledAt" IS NULL
	) pp ON true
	LEFT JOIN "Employee" oe ON oe.id = COALESCE(p."delegatedToEmployeeId", p."responsibleEmployeeId", p."surveyorEmployeeId")`

func (r *ProjectPaymentRepository) Money(projectID string) (*ProjectMoney, error) {
	var m ProjectMoney
	err := r.db.Get(&m, projectMoneySQL+` WHERE p.id = $1`, projectID)
	return &m, err
}

// AllMoney كل المشاريع (للشاشة ولعين الإيرادات).
func (r *ProjectPaymentRepository) AllMoney() ([]ProjectMoney, error) {
	rows := []ProjectMoney{}
	err := r.db.Select(&rows, projectMoneySQL+` ORDER BY p."updatedAt" DESC`)
	return rows, err
}

func (r *ProjectPaymentRepository) List(projectID string) ([]ProjectPayment, error) {
	rows := []ProjectPayment{}
	err := r.db.Select(&rows, `
		SELECT pp.id, pp."projectId", pp.amount::float8 AS amount, pp."paidAt", pp.method, pp."receiptNo", pp.note,
		       ce.name AS "createdBy", pp."createdAt", ve.name AS "verifiedBy", pp."verifiedAt", pp."cancelledAt", pp."cancelReason"
		FROM "ProjectPayment" pp
		LEFT JOIN "Employee" ce ON ce.id = pp."createdById"
		LEFT JOIN "Employee" ve ON ve.id = pp."verifiedById"
		WHERE pp."projectId" = $1 ORDER BY pp."paidAt" DESC, pp."createdAt" DESC`, projectID)
	return rows, err
}

type ProjectPaymentIn struct {
	Amount    float64 `json:"amount"`
	PaidAt    string  `json:"paidAt"`
	Method    string  `json:"method"`
	ReceiptNo *string `json:"receiptNo"`
	Note      *string `json:"note"`
}

func (r *ProjectPaymentRepository) Add(projectID, byID string, in ProjectPaymentIn, verified bool) (string, error) {
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "ProjectPayment" ("projectId", amount, "paidAt", method, "receiptNo", note, "createdById", "verifiedById", "verifiedAt")
		VALUES ($1, $2, $3::date, $4, NULLIF($5, ''), NULLIF($6, ''), $7,
		        CASE WHEN $8 THEN $7 END, CASE WHEN $8 THEN now() END)
		RETURNING id`, projectID, in.Amount, in.PaidAt, in.Method, in.ReceiptNo, in.Note, byID, verified)
	return id, err
}

func (r *ProjectPaymentRepository) ProjectOf(paymentID string) string {
	var id string
	_ = r.db.Get(&id, `SELECT "projectId" FROM "ProjectPayment" WHERE id = $1`, paymentID)
	return id
}

func (r *ProjectPaymentRepository) Verify(id, byID string) error {
	_, err := r.db.Exec(`UPDATE "ProjectPayment" SET "verifiedById" = $2, "verifiedAt" = now()
		WHERE id = $1 AND "verifiedAt" IS NULL AND "cancelledAt" IS NULL`, id, byID)
	return err
}

func (r *ProjectPaymentRepository) Cancel(id, byID, reason string) error {
	_, err := r.db.Exec(`UPDATE "ProjectPayment" SET "cancelledAt" = now(), "cancelledById" = $2, "cancelReason" = $3
		WHERE id = $1 AND "cancelledAt" IS NULL`, id, byID, reason)
	return err
}

func (r *ProjectPaymentRepository) SetContractValue(projectID, byID string, amount float64) error {
	_, err := r.db.Exec(`INSERT INTO "ProjectContractValue" ("projectId", amount, "setById") VALUES ($1, $2, $3)
		ON CONFLICT ("projectId") DO UPDATE SET amount = EXCLUDED.amount, "setById" = EXCLUDED."setById", "setAt" = now()`,
		projectID, amount, byID)
	return err
}

// IsSupervisor هل الموظف مشرف هالمشروع (المسؤول أو المتوجّه إله أو المسّاح).
func (r *ProjectPaymentRepository) IsSupervisor(projectID, employeeID string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Project" WHERE id = $1
		AND ($2 IN (COALESCE("responsibleEmployeeId", ''), COALESCE("delegatedToEmployeeId", ''), COALESCE("surveyorEmployeeId", ''))
		     -- (ع) 10-09: مشرف الحجز المربوط بالمشروع يرفع فلوسه هم
		     OR EXISTS (SELECT 1 FROM "Booking" b WHERE b.id = "Project"."bookingId" AND b."projectSupervisorId" = $2)))`,
		projectID, employeeID)
	return ok
}

// StaleUnverified دفعات مسجّلة من +٣ أيام وما تأكدت.
func (r *ProjectPaymentRepository) StaleUnverified() ([]ProjectPayment, error) {
	rows := []ProjectPayment{}
	err := r.db.Select(&rows, `
		SELECT pp.id, pp."projectId", pp.amount::float8 AS amount, pp."paidAt", pp.method, pp."receiptNo", pp.note,
		       ce.name AS "createdBy", pp."createdAt", NULL::text AS "verifiedBy", pp."verifiedAt", pp."cancelledAt", pp."cancelReason"
		FROM "ProjectPayment" pp LEFT JOIN "Employee" ce ON ce.id = pp."createdById"
		WHERE pp."verifiedAt" IS NULL AND pp."cancelledAt" IS NULL AND pp."createdAt" < now() - interval '3 days'
		ORDER BY pp."createdAt"`)
	return rows, err
}
