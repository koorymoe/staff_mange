package repository

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type DepartmentRepository struct {
	db *sqlx.DB
}

func NewDepartmentRepository(db *sqlx.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// List الأقسام مع مسؤوليها.
//
// ⚠️ `includeInactive` للشاشة الإدارية بس: منتقي الحجز يشوف الفعّال
// فقط، وإلا الموظف يحجز لقسم انلغى.
func (r *DepartmentRepository) List(includeInactive bool) ([]model.Department, error) {
	q := `SELECT * FROM "Department"`
	if !includeInactive {
		q += ` WHERE active = true`
	}
	q += ` ORDER BY name`
	depts := []model.Department{}
	if err := r.db.Select(&depts, q); err != nil {
		return nil, err
	}
	if len(depts) == 0 {
		return depts, nil
	}

	hq := `SELECT * FROM "DepartmentHead"`
	if !includeInactive {
		hq += ` WHERE active = true`
	}
	hq += ` ORDER BY name`
	heads := []model.DepartmentHead{}
	if err := r.db.Select(&heads, hq); err != nil {
		return nil, err
	}
	byDept := map[string][]model.DepartmentHead{}
	for _, h := range heads {
		byDept[h.DepartmentID] = append(byDept[h.DepartmentID], h)
	}
	for i := range depts {
		depts[i].Heads = byDept[depts[i].ID]
	}
	return depts, nil
}

func (r *DepartmentRepository) CreateDepartment(req model.SaveDepartmentRequest) (*model.Department, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("اكتب اسم القسم")
	}
	row := model.Department{}
	err := r.db.Get(&row, `
		INSERT INTO "Department" (id, name) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET active = true
		RETURNING *`, uuid.NewString(), name)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpdateDepartment تسمية أو تفعيل/تعطيل.
//
// ⚠️ التعطيل مو حذف: الحجوزات القديمة تبقى تشير للقسم وتبقى مقروءة.
func (r *DepartmentRepository) UpdateDepartment(id string, req model.SaveDepartmentRequest) (*model.Department, error) {
	name := strings.TrimSpace(req.Name)
	row := model.Department{}
	err := r.db.Get(&row, `
		UPDATE "Department" SET
			name   = COALESCE(NULLIF($2,''), name),
			active = COALESCE($3, active)
		WHERE id = $1
		RETURNING *`, id, name, req.Active)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DepartmentRepository) CreateHead(req model.SaveDepartmentHeadRequest) (*model.DepartmentHead, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("اكتب اسم المسؤول")
	}
	if strings.TrimSpace(req.DepartmentID) == "" {
		return nil, fmt.Errorf("اختر القسم")
	}
	row := model.DepartmentHead{}
	err := r.db.Get(&row, `
		INSERT INTO "DepartmentHead" (id, "departmentId", name, phone)
		VALUES ($1, $2, $3, NULLIF($4,''))
		RETURNING *`, uuid.NewString(), req.DepartmentID, name, strings.TrimSpace(derefStr(req.Phone)))
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DepartmentRepository) UpdateHead(id string, req model.SaveDepartmentHeadRequest) (*model.DepartmentHead, error) {
	row := model.DepartmentHead{}
	err := r.db.Get(&row, `
		UPDATE "DepartmentHead" SET
			name   = COALESCE(NULLIF($2,''), name),
			phone  = COALESCE(NULLIF($3,''), phone),
			active = COALESCE($4, active)
		WHERE id = $1
		RETURNING *`, id, strings.TrimSpace(req.Name), strings.TrimSpace(derefStr(req.Phone)), req.Active)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// HeadForEmployee مسؤول القسم المربوط بحساب موظف — يخدم «إذا هوه الي
// مسوي الحجز يطلع اسمه مباشرة». يرجّع nil بلا خطأ لو ما چان مسؤولاً.
//
// ⚠️ يشتغل بس بعد ما تجي مرحلة ربط الحسابات؛ اليوم `employeeId` فارغ
// دائماً فيرجّع nil — وهذا صحيح، مو خلل.
func (r *DepartmentRepository) HeadForEmployee(employeeID string) (*model.DepartmentHead, error) {
	rows := []model.DepartmentHead{}
	err := r.db.Select(&rows, `
		SELECT * FROM "DepartmentHead"
		WHERE "employeeId" = $1 AND active = true LIMIT 1`, employeeID)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

// HeadWithDepartment مسؤول واحد مع قسمه — يخدم تعبئة اسم الطالب
// بالحجز الداخلي بنداء واحد بدل نداءين.
func (r *DepartmentRepository) HeadWithDepartment(headID string) (*model.DepartmentHead, *model.Department, error) {
	heads := []model.DepartmentHead{}
	if err := r.db.Select(&heads, `SELECT * FROM "DepartmentHead" WHERE id = $1`, headID); err != nil {
		return nil, nil, err
	}
	if len(heads) == 0 {
		return nil, nil, nil
	}
	depts := []model.Department{}
	if err := r.db.Select(&depts, `SELECT * FROM "Department" WHERE id = $1`, heads[0].DepartmentID); err != nil {
		return &heads[0], nil, err
	}
	if len(depts) == 0 {
		return &heads[0], nil, nil
	}
	return &heads[0], &depts[0], nil
}
