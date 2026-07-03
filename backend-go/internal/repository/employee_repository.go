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

func (r *EmployeeRepository) List() ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.Select(&employees, `SELECT * FROM "Employee" ORDER BY name ASC`)
	return employees, err
}

func (r *EmployeeRepository) FindByID(id string) (*model.Employee, error) {
	var e model.Employee
	err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) FindByUsername(username string) (*model.Employee, error) {
	var e model.Employee
	err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE username = $1`, username)
	if err != nil {
		return nil, err
	}
	return &e, nil
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
			"jobTitle" = :jobTitle
		WHERE id = :id
	`, e)
	return err
}

// Supervisors يرجّع تيم ليدرز ومدراء المشاريع النشطين المؤهلين للإشراف على تكليف الفنيين
func (r *EmployeeRepository) Supervisors() ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.Select(&employees, `
		SELECT * FROM "Employee"
		WHERE status = 'ACTIVE' AND (role = 'PROJECT_MANAGER' OR "isLeader" = true)
		ORDER BY name ASC
	`)
	return employees, err
}

// MatchForService يرجّع الفنيين النشطين والمتاحين حالياً مع علامة إذا يمتلكون مهارة الخدمة المطلوبة
func (r *EmployeeRepository) MatchForService(serviceID string) ([]model.Employee, error) {
	var employees []model.Employee
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
	var skills []model.EmployeeSkillDetail
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
