package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type PrivacyPolicyRepository struct {
	db *sqlx.DB
}

func NewPrivacyPolicyRepository(db *sqlx.DB) *PrivacyPolicyRepository {
	return &PrivacyPolicyRepository{db: db}
}

// اسم مضيف النقطة يجي بـJOIN — المالك ومدير النظام يشوفونه، وغيرهم ينحذف
// بطبقة الخدمة قبل الإرجاع.
const privacyPointSelect = `SELECT p.*, e.name AS "createdByName"
	FROM "PrivacyPolicyPoint" p
	LEFT JOIN "Employee" e ON e.id = p."createdByEmployeeId"`

func (r *PrivacyPolicyRepository) ListPoints(activeOnly bool) ([]model.PrivacyPolicyPoint, error) {
	points := []model.PrivacyPolicyPoint{}
	q := privacyPointSelect
	if activeOnly {
		q += ` WHERE p."isActive" = true`
	}
	q += ` ORDER BY p."order" ASC, p."createdAt" ASC`
	err := r.db.Select(&points, q)
	return points, err
}

func (r *PrivacyPolicyRepository) getPoint(id string) (*model.PrivacyPolicyPoint, error) {
	var p model.PrivacyPolicyPoint
	if err := r.db.Get(&p, privacyPointSelect+` WHERE p.id = $1`, id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PrivacyPolicyRepository) CreatePoint(content string, order int, createdBy *string) (*model.PrivacyPolicyPoint, error) {
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "PrivacyPolicyPoint" (id, content, "order", "createdByEmployeeId")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING id
	`, content, order, createdBy)
	if err != nil {
		return nil, err
	}
	return r.getPoint(id)
}

func (r *PrivacyPolicyRepository) UpdatePoint(id string, req model.UpsertPrivacyPolicyPointRequest) (*model.PrivacyPolicyPoint, error) {
	var content *string
	if req.Content != "" {
		content = &req.Content
	}
	_, err := r.db.Exec(`
		UPDATE "PrivacyPolicyPoint" SET
			content = COALESCE($2, content),
			"order" = COALESCE($3, "order"),
			"isActive" = COALESCE($4, "isActive"),
			"updatedAt" = now()
		WHERE id = $1
	`, id, content, req.Order, req.IsActive)
	if err != nil {
		return nil, err
	}
	return r.getPoint(id)
}

func (r *PrivacyPolicyRepository) DeletePoint(id string) error {
	_, err := r.db.Exec(`DELETE FROM "PrivacyPolicyPoint" WHERE id = $1`, id)
	return err
}

func (r *PrivacyPolicyRepository) countActivePoints() (int, error) {
	var n int
	err := r.db.Get(&n, `SELECT COUNT(*) FROM "PrivacyPolicyPoint" WHERE "isActive" = true`)
	return n, err
}

// Status يحسب إذا الموظف يحتاج يوافق: ما وافق أبداً، أو انضافت نقاط بعد آخر
// موافقة. لو ما اكو أي نقطة فعّالة ما نطلب موافقة إطلاقاً.
func (r *PrivacyPolicyRepository) Status(employeeID string) (*model.PrivacyPolicyStatus, error) {
	points, err := r.ListPoints(true)
	if err != nil {
		return nil, err
	}
	st := &model.PrivacyPolicyStatus{Points: points}
	if len(points) == 0 {
		return st, nil
	}

	var acceptedAt time.Time
	var version int
	err = r.db.QueryRow(`
		SELECT "acceptedAt", "pointsVersion" FROM "PrivacyPolicyAcceptance" WHERE "employeeId" = $1
	`, employeeID).Scan(&acceptedAt, &version)
	if errors.Is(err, sql.ErrNoRows) {
		st.NeedsAcceptance = true
		return st, nil
	}
	if err != nil {
		return nil, err
	}
	st.AcceptedAt = &acceptedAt
	st.NeedsAcceptance = version < len(points)
	return st, nil
}

func (r *PrivacyPolicyRepository) Accept(employeeID string) error {
	n, err := r.countActivePoints()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		INSERT INTO "PrivacyPolicyAcceptance" (id, "employeeId", "pointsVersion")
		VALUES (gen_random_uuid()::text, $1, $2)
		ON CONFLICT ("employeeId") DO UPDATE
		SET "acceptedAt" = now(), "pointsVersion" = EXCLUDED."pointsVersion"
	`, employeeID, n)
	return err
}
