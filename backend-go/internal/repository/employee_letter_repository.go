package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// EmployeeLetterRepository الطلبات — كتاب رسمي من الموظف للإدارة.
type EmployeeLetterRepository struct {
	db *sqlx.DB
}

func NewEmployeeLetterRepository(db *sqlx.DB) *EmployeeLetterRepository {
	return &EmployeeLetterRepository{db: db}
}

func (r *EmployeeLetterRepository) Create(employeeID string, req model.CreateEmployeeLetterRequest) (*model.EmployeeLetter, error) {
	var l model.EmployeeLetter
	err := r.db.Get(&l, `
		INSERT INTO "EmployeeLetter" (id, "employeeId", "addressedTo", subject, body)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *`,
		uuid.NewString(), employeeID, req.AddressedTo, req.Subject, req.Body)
	if err != nil {
		return nil, err
	}
	r.hydrateOne(&l)
	return &l, nil
}

// Mine طلبات الموظف نفسه — كل واحد يشوف طلباته وأجوبتها.
func (r *EmployeeLetterRepository) Mine(employeeID string) ([]model.EmployeeLetter, error) {
	rows := []model.EmployeeLetter{}
	if err := r.db.Select(&rows, `
		SELECT * FROM "EmployeeLetter" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID); err != nil {
		return nil, err
	}
	r.hydrateAll(rows)
	return rows, nil
}

// Inbox صندوق الإدارة — المعلّقة أول، لأنها الي تحتاج قرار.
func (r *EmployeeLetterRepository) Inbox(status string) ([]model.EmployeeLetter, error) {
	rows := []model.EmployeeLetter{}
	query := `SELECT * FROM "EmployeeLetter"`
	args := []any{}
	if status != "" {
		args = append(args, status)
		query += ` WHERE status = $1`
	}
	query += ` ORDER BY (status = 'PENDING') DESC, "createdAt" DESC`
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	r.hydrateAll(rows)
	return rows, nil
}

// Decide جواب الإدارة — مرة وحدة بس.
//
// الشرط status = 'PENDING' هو الي يمنع تغيير القرار بعد ما ينصدر:
// الموظف انبلغ بالجواب، وتغييره بالخفية بعدها يكسر ثقته بالنظام.
func (r *EmployeeLetterRepository) Decide(id, byEmployeeID string, req model.DecideEmployeeLetterRequest) (*model.EmployeeLetter, error) {
	status := "REJECTED"
	if req.Approve {
		status = "APPROVED"
	}
	var l model.EmployeeLetter
	err := r.db.Get(&l, `
		UPDATE "EmployeeLetter"
		SET status = $2, "decisionNote" = NULLIF($3,''), "decidedById" = $4, "decidedAt" = now()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING *`, id, status, req.Note, byEmployeeID)
	if err != nil {
		return nil, errors.New("الطلب مو موجود أو انبتّ بيه من قبل")
	}
	r.hydrateOne(&l)
	return &l, nil
}

func (r *EmployeeLetterRepository) PendingCount() (int, error) {
	var n int
	err := r.db.Get(&n, `SELECT COUNT(*) FROM "EmployeeLetter" WHERE status = 'PENDING'`)
	return n, err
}

func (r *EmployeeLetterRepository) hydrateOne(l *model.EmployeeLetter) {
	rows := []model.EmployeeLetter{*l}
	r.hydrateAll(rows)
	*l = rows[0]
}

// hydrateAll يجيب أسماء الموظفين بدفعة وحدة مو استعلام لكل صف.
func (r *EmployeeLetterRepository) hydrateAll(rows []model.EmployeeLetter) {
	if len(rows) == 0 {
		return
	}
	ids := map[string]bool{}
	for i := range rows {
		ids[rows[i].EmployeeID] = true
		if rows[i].DecidedByID != nil {
			ids[*rows[i].DecidedByID] = true
		}
	}
	list := make([]string, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}
	type brief struct {
		ID       string  `db:"id"`
		Name     string  `db:"name"`
		JobTitle *string `db:"jobTitle"`
	}
	briefs := []brief{}
	q, args, err := sqlx.In(`SELECT id, name, "jobTitle" FROM "Employee" WHERE id IN (?)`, list)
	if err != nil {
		return
	}
	if err := r.db.Select(&briefs, r.db.Rebind(q), args...); err != nil {
		return
	}
	byID := map[string]brief{}
	for _, b := range briefs {
		byID[b.ID] = b
	}
	for i := range rows {
		if b, ok := byID[rows[i].EmployeeID]; ok {
			rows[i].Employee = &model.EmployeeBrief{ID: b.ID, Name: b.Name}
			rows[i].EmployeeJobTitle = b.JobTitle
		}
		if rows[i].DecidedByID != nil {
			if b, ok := byID[*rows[i].DecidedByID]; ok {
				rows[i].DecidedBy = &model.EmployeeBrief{ID: b.ID, Name: b.Name}
			}
		}
	}
}
