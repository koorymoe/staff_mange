package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

// pointsHash بصمة النقاط الفعّالة (المعرّف + النص بالترتيب) — أي تعديل أو إضافة أو
// تعطيل يغيّرها فينطلب من الكل يوافقون من جديد.
func pointsHash(points []model.PrivacyPolicyPoint) string {
	h := sha256.New()
	for _, p := range points {
		h.Write([]byte(p.ID))
		h.Write([]byte{0})
		h.Write([]byte(p.Content))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Status يحسب إذا الموظف يحتاج يوافق: ما وافق أبداً، أو النص تغيّر بعد آخر موافقة.
// لو ما اكو أي نقطة فعّالة ما نطلب موافقة إطلاقاً.
func (r *PrivacyPolicyRepository) Status(employeeID string) (*model.PrivacyPolicyStatus, error) {
	points, err := r.ListPoints(true)
	if err != nil {
		return nil, err
	}
	st := &model.PrivacyPolicyStatus{Points: points, Hash: pointsHash(points)}
	if len(points) == 0 {
		return st, nil
	}
	var row struct {
		At   time.Time `db:"acceptedAt"`
		Hash *string   `db:"pointsHash"`
	}
	err = r.db.Get(&row, `SELECT "acceptedAt", "pointsHash" FROM "PrivacyPolicyAcceptance" WHERE "employeeId" = $1`, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		st.NeedsAcceptance = true
		return st, nil
	}
	if err != nil {
		return nil, err
	}
	st.AcceptedAt = &row.At
	st.NeedsAcceptance = row.Hash == nil || *row.Hash != st.Hash
	st.Changed = st.NeedsAcceptance
	return st, nil
}

// Accept يوافق على النص الحالي بالضبط: يحدّث آخر موافقة ويضيف سطر بالسجل (ما ينمسح).
// expectHash بصمة النص الي قراه — إذا النص تغيّر (أو ماكو بصمة) نرفض حتى يقرا الجديد.
func (r *PrivacyPolicyRepository) Accept(employeeID string, expectHash string) error {
	points, err := r.ListPoints(true)
	if err != nil {
		return err
	}
	hash := pointsHash(points)
	if expectHash != hash {
		return ErrPolicyChanged
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`
		INSERT INTO "PrivacyPolicyAcceptance" (id, "employeeId", "pointsVersion", "pointsHash")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		ON CONFLICT ("employeeId") DO UPDATE
		SET "acceptedAt" = now(), "pointsVersion" = EXCLUDED."pointsVersion", "pointsHash" = EXCLUDED."pointsHash"
	`, employeeID, len(points), hash); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO "PrivacyPolicyAcceptanceLog" ("employeeId", "pointsHash", "pointsCount") VALUES ($1, $2, $3)`,
		employeeID, hash, len(points)); err != nil {
		return err
	}
	return tx.Commit()
}

var ErrPolicyChanged = errors.New("السياسة تحدّثت هسه — اقرا النسخة الجديدة ووافق عليها")

// CurrentHash بصمة النص الحالي (ترجع ويا الحالة حتى الواجهة ترسلها ويا الموافقة).
func (r *PrivacyPolicyRepository) CurrentHash() (string, error) {
	points, err := r.ListPoints(true)
	if err != nil {
		return "", err
	}
	return pointsHash(points), nil
}

type PolicyAcceptance struct {
	AcceptedAt  time.Time `db:"acceptedAt" json:"acceptedAt"`
	PointsCount int       `db:"pointsCount" json:"pointsCount"`
	Current     bool      `db:"current" json:"current"`
}

// History سجل موافقات موظف (الأحدث أول) — current إذا وافق على النص الحالي.
func (r *PrivacyPolicyRepository) History(employeeID string) ([]PolicyAcceptance, error) {
	hash, err := r.CurrentHash()
	if err != nil {
		return nil, err
	}
	out := []PolicyAcceptance{}
	err = r.db.Select(&out, `SELECT "acceptedAt", "pointsCount", "pointsHash" = $2 AS current FROM "PrivacyPolicyAcceptanceLog"
		WHERE "employeeId" = $1 ORDER BY "acceptedAt" DESC LIMIT 100`, employeeID, hash)
	return out, err
}
