package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(db *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

// List يرجع الموظفين النشطين فقط (يستثني المؤرشفين والمحذوفين والموقوفين) —
// هذا افتراضي لكل واجهات النظام العادية.
func (r *EmployeeRepository) List() ([]model.Employee, error) {
	employees := []model.Employee{}
	// حساب المالك (OWNER) ما يطلع بأي قائمة موظفين عادية أبداً — حساب مخفي تماماً
	if err := r.db.Select(&employees, `SELECT * FROM "Employee" WHERE status NOT IN ('ARCHIVED', 'DELETED', 'SUSPENDED') AND role != 'OWNER' ORDER BY name ASC`); err != nil {
		return nil, err
	}
	for i := range employees {
		skills, err := r.SkillsForEmployee(employees[i].ID)
		if err != nil {
			return nil, err
		}
		employees[i].Skills = skills
	}
	return employees, nil
}

// ListArchived يرجع المؤرشفين والمحذوفين والموقوفين تلقائياً (بسبب محاولات
// اختراق) — للأدمن/المالك حصراً، لمراجعة تاريخهم أو استرجاعهم عند الحاجة.
func (r *EmployeeRepository) ListArchived() ([]model.Employee, error) {
	employees := []model.Employee{}
	if err := r.db.Select(&employees, `SELECT * FROM "Employee" WHERE status IN ('ARCHIVED', 'DELETED', 'SUSPENDED') ORDER BY name ASC`); err != nil {
		return nil, err
	}
	for i := range employees {
		skills, err := r.SkillsForEmployee(employees[i].ID)
		if err != nil {
			return nil, err
		}
		employees[i].Skills = skills
	}
	return employees, nil
}

func (r *EmployeeRepository) FindByID(id string) (*model.Employee, error) {
	var e model.Employee
	if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	skills, err := r.SkillsForEmployee(e.ID)
	if err != nil {
		return nil, err
	}
	e.Skills = skills
	return &e, nil
}

// StatusByID استعلام خفيف يرجع حالة الموظف بس (بدون بقية بياناته) — يستخدمه
// RequireAuth بكل طلب يتحقق الحساب لسه فعّال (مو موقوف SUSPENDED) حتى لو
// التوكن نفسه لسه صالح.
func (r *EmployeeRepository) StatusByID(id string) (string, error) {
	var status string
	err := r.db.Get(&status, `SELECT status FROM "Employee" WHERE id = $1`, id)
	return status, err
}

// RecordAuthzViolation تسجل محاولة وصول مرفوضة (طلب عملية الموظف مو مخوّل
// لها فعلياً حسب دوره الحقيقي بالتوكن) — إذا تكررت 3 مرات نوقف الحساب
// تلقائياً (SUSPENDED) حماية من محاولات التلاعب بجلسة المتصفح.
func (r *EmployeeRepository) RecordAuthzViolation(id string) (violations int, suspended bool, err error) {
	err = r.db.Get(&violations, `
		UPDATE "Employee" SET "authzViolations" = "authzViolations" + 1
		WHERE id = $1
		RETURNING "authzViolations"
	`, id)
	if err != nil {
		return 0, false, err
	}
	if violations >= 3 {
		if _, execErr := r.db.Exec(`UPDATE "Employee" SET status = 'SUSPENDED' WHERE id = $1 AND status = 'ACTIVE'`, id); execErr != nil {
			return violations, false, execErr
		}
		suspended = true
	}
	return violations, suspended, nil
}

func (r *EmployeeRepository) FindByUsername(username string) (*model.Employee, error) {
	var e model.Employee
	if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE username = $1`, username); err != nil {
		return nil, err
	}
	skills, err := r.SkillsForEmployee(e.ID)
	if err != nil {
		return nil, err
	}
	e.Skills = skills
	return &e, nil
}

func (r *EmployeeRepository) SetTrainee(id string, isTrainee bool) error {
	_, err := r.db.Exec(`UPDATE "Employee" SET "isTrainee" = $2 WHERE id = $1`, id, isTrainee)
	return err
}

func (r *EmployeeRepository) Create(e *model.Employee) error {
	_, err := r.db.NamedExec(`
		INSERT INTO "Employee" (id, name, certificate, position, phone, username, password, "jobTitle", salary, shift, "shiftStart", "shiftEnd", role)
		VALUES (:id, :name, :certificate, :position, :phone, :username, :password, :jobTitle, :salary, :shift, :shiftStart, :shiftEnd, :role)
	`, e)
	return err
}

func (r *EmployeeRepository) Update(e *model.Employee) error {
	_, err := r.db.NamedExec(`
		UPDATE "Employee" SET
			name = :name,
			certificate = :certificate,
			position = :position,
			phone = :phone,
			status = :status,
			role = :role,
			"onDuty" = :onDuty,
			username = :username,
			password = COALESCE(NULLIF(:password, ''), password),
			"hasDrivingLicense" = :hasDrivingLicense,
			"hasSafetyCertificate" = :hasSafetyCertificate,
			"isLeader" = :isLeader,
			"isTrainee" = :isTrainee,
			salary = :salary,
			shift = :shift,
			"shiftStart" = :shiftStart,
			"shiftEnd" = :shiftEnd,
			"monthlyLeaves" = :monthlyLeaves,
			"jobTitle" = :jobTitle,
			"authzViolations" = :authzViolations
		WHERE id = :id
	`, e)
	return err
}

// Supervisors يرجّع تيم ليدرز ومدراء المشاريع النشطين المؤهلين للإشراف على تكليف الفنيين
func (r *EmployeeRepository) Supervisors() ([]model.Employee, error) {
	employees := []model.Employee{}
	err := r.db.Select(&employees, `
		SELECT * FROM "Employee"
		WHERE status = 'ACTIVE' AND (role = 'PROJECT_MANAGER' OR "isLeader" = true)
		ORDER BY name ASC
	`)
	return employees, err
}

// MatchForService يرجّع الفنيين النشطين والمتاحين حالياً مع علامة إذا يمتلكون مهارة الخدمة المطلوبة
func (r *EmployeeRepository) MatchForService(serviceID string) ([]model.Employee, error) {
	employees := []model.Employee{}
	if err := r.db.Select(&employees, `
		SELECT * FROM "Employee"
		WHERE status = 'ACTIVE' AND "onDuty" = true AND role = 'TECHNICIAN'
		ORDER BY name ASC
	`); err != nil {
		return nil, err
	}
	for i := range employees {
		skills, err := r.SkillsForEmployee(employees[i].ID)
		if err != nil {
			return nil, err
		}
		employees[i].Skills = skills
		hasSkill := false
		for _, s := range skills {
			if s.CanPerform && s.Skill != nil && s.Skill.ServiceID == serviceID {
				hasSkill = true
				break
			}
		}
		employees[i].HasRequiredSkill = &hasSkill
	}
	return employees, nil
}

// SkillsForEmployee يجلب كل مهارات موظف مع تفاصيل المهارة والخدمة المرتبطة بيها
func (r *EmployeeRepository) SkillsForEmployee(employeeID string) ([]model.EmployeeSkillDetail, error) {
	skills := []model.EmployeeSkillDetail{}
	if err := r.db.Select(&skills, `SELECT * FROM "EmployeeSkill" WHERE "employeeId" = $1`, employeeID); err != nil {
		return nil, err
	}
	for i := range skills {
		var skill model.Skill
		if err := r.db.Get(&skill, `SELECT * FROM "Skill" WHERE id = $1`, skills[i].SkillID); err == nil {
			skills[i].Skill = &skill
		}
	}
	return skills, nil
}

// SetSkills يستبدل مهارات الموظف بالكامل (حذف القديم وإدخال الجديد بنفس المعاملة)
func (r *EmployeeRepository) SetSkills(employeeID string, skills []model.EmployeeSkillInput) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM "EmployeeSkill" WHERE "employeeId" = $1`, employeeID); err != nil {
		return err
	}
	for _, s := range skills {
		if _, err := tx.Exec(`
			INSERT INTO "EmployeeSkill" (id, "employeeId", "skillId", "canPerform")
			VALUES (gen_random_uuid()::text, $1, $2, $3)
		`, employeeID, s.SkillID, s.CanPerform); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LinkHistoricalRecords يربط سجلات تاريخية (حجوزات/شكاوى مستوردة من نظام قديم بس
// بالاسم النصي، بدون حساب فعلي وقتها) بحساب موظف حالي بنفس الاسم — يستخدمها الأدمن
// يدوياً لما موظف قديم يرجع ويصير له حساب جديد بنفس اسمه بالضبط.
func (r *EmployeeRepository) LinkHistoricalRecords(employeeID, employeeName string) (bookingsLinked int, complaintsLinked int, err error) {
	res, err := r.db.Exec(`
		UPDATE "Booking" SET "confirmedByEmployeeId" = $1
		WHERE "confirmedByName" = $2 AND "confirmedByEmployeeId" IS NULL
	`, employeeID, employeeName)
	if err != nil {
		return 0, 0, err
	}
	if n, e := res.RowsAffected(); e == nil {
		bookingsLinked = int(n)
	}

	res, err = r.db.Exec(`
		UPDATE "Complaint" SET "relatedEmployeeId" = $1
		WHERE "relatedEmployeeName" = $2 AND "relatedEmployeeId" IS NULL
	`, employeeID, employeeName)
	if err != nil {
		return bookingsLinked, 0, err
	}
	if n, e := res.RowsAffected(); e == nil {
		complaintsLinked = int(n)
	}

	return bookingsLinked, complaintsLinked, nil
}
