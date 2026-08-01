package repository

import (
	"github.com/jmoiron/sqlx"
	"strings"

	"staffmange-api/internal/model"
)

const firstStage = "1. اتصال بالزبون"

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// List يجيب المشاريع بدون أعمدة العقد (PDF بصيغة base64) — هذي الأعمدة ممكن
// تكون ميغابايتات لكل مشروع، وجلبها بالقائمة كان يخلي الصفحة بطيئة جداً. بدالها
// نرجّع علمين بس (هل اكو عقد / هل اكو عقد موقّع)، والملف نفسه ينجلب لما
// المستخدم يفتح نافذة العقد (GET /api/projects/{id}).
// prefixed يضيف بادئة الجدول لكل عمود بالقائمة — لازمة لما نعمل JOIN، وإلا
// يطلع خطأ "column reference is ambiguous" أو ما ينلكه العمود أصلاً.
func prefixed(cols, alias string) string {
	parts := strings.Split(cols, ",")
	out := make([]string, 0, len(parts))
	for _, c := range parts {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		// التعابير المحسوبة (فيها أقواس) تنترك مثل ما هي
		if strings.ContainsAny(c, "()") {
			out = append(out, c)
			continue
		}
		out = append(out, alias+"."+c)
	}
	return strings.Join(out, ", ")
}

const projectListColumns = `id, code, name, rep, phone, location, "locationUrl", "mapLatitude", "mapLongitude",
	"workType", "refPerson", stage, price, staff, time, task, priority, "deliveryDate", survey,
	"bookingId", "responsibleEmployeeId", "surveyorEmployeeId", "createdByEmployeeId", "createdAt", "updatedAt",
	"delegatedToEmployeeId", "delegatedByEmployeeId", "delegatedAt",
	("contractPdfBase64" IS NOT NULL) AS "hasContract",
	("signedContractPdfBase64" IS NOT NULL) AS "hasSignedContract"`

const projectListSelect = `SELECT ` + `%s` + `, e.name AS "createdByName", d.name AS "delegatedToName"
	FROM "Project" p
	LEFT JOIN "Employee" e ON e.id = p."createdByEmployeeId"
	LEFT JOIN "Employee" d ON d.id = p."delegatedToEmployeeId"`

func (r *ProjectRepository) List() ([]model.Project, error) {
	projects := []model.Project{}
	// نجيب اسم مضيف المشروع واسم الموظف المُسلَّم إله بـJOIN بدل استعلام لكل صف
	err := r.db.Select(&projects, strings.Replace(projectListSelect, "%s", prefixed(projectListColumns, "p"), 1)+
		` ORDER BY p."createdAt" DESC`)
	return projects, err
}

// ListDelegatedTo يرجّع بس المشاريع المُسلَّمة لموظف معيّن — يستخدمها الموظف
// الي ما عنده صلاحية إدارة المشاريع العامة، فيشوف مشاريعه هو بس.
func (r *ProjectRepository) ListDelegatedTo(employeeID string) ([]model.Project, error) {
	projects := []model.Project{}
	err := r.db.Select(&projects, strings.Replace(projectListSelect, "%s", prefixed(projectListColumns, "p"), 1)+
		` WHERE p."delegatedToEmployeeId" = $1 ORDER BY p."createdAt" DESC`, employeeID)
	return projects, err
}

// IsDelegatedTo يفحص إذا المشروع مُسلَّم لهذا الموظف — أساس التحقق قبل أي
// تعديل: الموظف يتحكم بمشروعه المُسلَّم بس، مو بكل المشاريع.
func (r *ProjectRepository) IsDelegatedTo(projectID, employeeID string) (bool, error) {
	var n int
	err := r.db.Get(&n, `SELECT COUNT(*) FROM "Project" WHERE id = $1 AND "delegatedToEmployeeId" = $2`,
		projectID, employeeID)
	return n > 0, err
}

// Delegate يسلّم المشروع لموظف (أو يسحبه لو employeeID فاضي) ويسجّل الحركة.
func (r *ProjectRepository) Delegate(projectID string, employeeID *string, byEmployeeID *string, note string) (*model.Project, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var p model.Project
	if err := tx.Get(&p, `
		UPDATE "Project" SET
			-- التحويل الصريح لازم: بدونه Postgres ما يقدر يستنتج نوع $2 داخل CASE
			"delegatedToEmployeeId" = $2::text,
			"delegatedByEmployeeId" = CASE WHEN $2::text IS NULL THEN NULL ELSE $3::text END,
			"delegatedAt"           = CASE WHEN $2::text IS NULL THEN NULL ELSE now() END,
			"updatedAt" = now()
		WHERE id = $1
		RETURNING *,
			("contractPdfBase64" IS NOT NULL) AS "hasContract",
			("signedContractPdfBase64" IS NOT NULL) AS "hasSignedContract"
	`, projectID, employeeID, byEmployeeID); err != nil {
		return nil, err
	}

	action, logEmployee := "REVOKE", ""
	if employeeID != nil && *employeeID != "" {
		action = "ASSIGN"
		logEmployee = *employeeID
	} else {
		// عند السحب نسجّل منو جان مستلمه قبل (لو معروف) حتى يبقى السجل مفيد
		if p.DelegatedToEmployeeID != nil {
			logEmployee = *p.DelegatedToEmployeeID
		}
	}
	if logEmployee != "" {
		if _, err := tx.Exec(`
			INSERT INTO "ProjectDelegationLog" ("projectId", "employeeId", "delegatedByEmployeeId", action, note)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''))
		`, projectID, logEmployee, byEmployeeID, action, note); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &p, nil
}

// DelegationLog سجل تسليم مشروع واحد (أو كل السجل لو projectID فاضي).
func (r *ProjectRepository) DelegationLog(projectID string) ([]model.ProjectDelegationLogEntry, error) {
	rows := []model.ProjectDelegationLogEntry{}
	q := `SELECT l.*, e.name AS "employeeName", b.name AS "delegatedByName",
			p.code AS "projectCode", p.name AS "projectName"
		FROM "ProjectDelegationLog" l
		LEFT JOIN "Employee" e ON e.id = l."employeeId"
		LEFT JOIN "Employee" b ON b.id = l."delegatedByEmployeeId"
		LEFT JOIN "Project"  p ON p.id = l."projectId"`
	if projectID != "" {
		return rows, r.db.Select(&rows, q+` WHERE l."projectId" = $1 ORDER BY l."createdAt" DESC`, projectID)
	}
	return rows, r.db.Select(&rows, q+` ORDER BY l."createdAt" DESC`)
}

// GetByID يرجّع المشروع كامل بما بيه ملفات العقد — يُستخدم لما تنفتح نافذة العقد.
func (r *ProjectRepository) GetByID(id string) (*model.Project, error) {
	var p model.Project
	err := r.db.Get(&p, `SELECT *,
		("contractPdfBase64" IS NOT NULL) AS "hasContract",
		("signedContractPdfBase64" IS NOT NULL) AS "hasSignedContract"
		FROM "Project" WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// NextCodeNumber يرجّع أكبر رقم مستخدم بأكواد المشاريع (PRJ-0007 -> 7).
// نعتمد على أكبر رقم موجود مو على عدد الصفوف: لو انحذف مشروع، العدد ينقص بينما
// الكود القديم يبقى مستخدَم — وهذا الي كان يسبب خطأ
// duplicate key value violates unique constraint "Project_code_key".
func (r *ProjectRepository) NextCodeNumber() (int, error) {
	var maxNum int
	err := r.db.Get(&maxNum, `
		SELECT COALESCE(MAX(NULLIF(regexp_replace(code, '\D', '', 'g'), '')::int), 0)
		FROM "Project"`)
	if err != nil {
		return 0, err
	}
	return maxNum + 1, nil
}

func (r *ProjectRepository) CountAll() (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "Project"`)
	return count, err
}

func (r *ProjectRepository) Create(code, name string, rep, phone, location *string, mapLatitude, mapLongitude *float64, workType, refPerson *string, priority string, deliveryDate, bookingID, responsibleEmployeeID, surveyorEmployeeID, locationUrl, createdByEmployeeID *string) (*model.Project, error) {
	var p model.Project
	err := r.db.Get(&p, `
		INSERT INTO "Project" (id, code, name, rep, phone, location, "locationUrl", "mapLatitude", "mapLongitude", "workType", "refPerson", priority, "deliveryDate", stage, "bookingId", "responsibleEmployeeId", "surveyorEmployeeId", "createdByEmployeeId", "updatedAt")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $16, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $17, now())
		RETURNING *
	`, code, name, rep, phone, location, mapLatitude, mapLongitude, workType, refPerson, priority, deliveryDate, firstStage, bookingID, responsibleEmployeeID, surveyorEmployeeID, locationUrl, createdByEmployeeID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) Update(id string, req model.UpdateProjectRequest) (*model.Project, error) {
	var p model.Project
	err := r.db.Get(&p, `
		UPDATE "Project" SET
			name = COALESCE($2, name),
			rep = COALESCE($3, rep),
			phone = COALESCE($4, phone),
			location = COALESCE($5, location),
			"locationUrl" = COALESCE($24, "locationUrl"),
			"mapLatitude" = COALESCE($6, "mapLatitude"),
			"mapLongitude" = COALESCE($7, "mapLongitude"),
			"workType" = COALESCE($8, "workType"),
			"refPerson" = COALESCE($9, "refPerson"),
			stage = COALESCE($10, stage),
			price = COALESCE($11, price),
			staff = COALESCE($12, staff),
			time = COALESCE($13, time),
			task = COALESCE($14, task),
			priority = COALESCE($15, priority),
			"deliveryDate" = COALESCE($16, "deliveryDate"),
			survey = COALESCE($17::jsonb, survey),
			"contractPdfBase64" = CASE WHEN $22 THEN NULL ELSE COALESCE($18, "contractPdfBase64") END,
			"signedContractPdfBase64" = CASE WHEN $23 THEN NULL ELSE COALESCE($19, "signedContractPdfBase64") END,
			"responsibleEmployeeId" = COALESCE($20, "responsibleEmployeeId"),
			"surveyorEmployeeId" = COALESCE($21, "surveyorEmployeeId"),
			"updatedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, req.Name, req.Rep, req.Phone, req.Location, req.MapLatitude, req.MapLongitude, req.WorkType, req.RefPerson, req.Stage,
		req.Price, req.Staff, req.Time, req.Task, req.Priority, req.DeliveryDate, req.Survey,
		req.ContractPdfBase64, req.SignedContractPdfBase64, req.ResponsibleEmployeeID, req.SurveyorEmployeeID,
		req.ClearContract, req.ClearSignedContract, req.LocationUrl)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Project" WHERE id = $1`, id)
	return err
}
