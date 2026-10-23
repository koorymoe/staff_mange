package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type ProjectWorkTypeRepository struct {
	db *sqlx.DB
}

func NewProjectWorkTypeRepository(db *sqlx.DB) *ProjectWorkTypeRepository {
	return &ProjectWorkTypeRepository{db: db}
}

func (r *ProjectWorkTypeRepository) List() ([]model.ProjectWorkType, error) {
	items := []model.ProjectWorkType{}
	err := r.db.Select(&items, `SELECT * FROM "ProjectWorkType" ORDER BY "createdAt" ASC`)
	return items, err
}

func (r *ProjectWorkTypeRepository) Create(id, name string) (*model.ProjectWorkType, error) {
	var item model.ProjectWorkType
	err := r.db.Get(&item, `
		INSERT INTO "ProjectWorkType" (id, name) VALUES ($1, $2)
		RETURNING *
	`, id, name)
	return &item, err
}

func (r *ProjectWorkTypeRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "ProjectWorkType" WHERE id = $1`, id)
	return err
}

// ListProjectCandidates يرجّع كل الموظفين النشطين مع علمين محسوبين: هل عنده
// مهارات الهندسة (تصميم/تخطيط/تنفيذ بـcanPerform)، وهل عنده صلاحية التقني —
// يُستخدمان لتصنيفه بالمجموعة الصحيحة بقوائم "المسؤول عن المشروع"/"منفّذ الكشف".
func (r *ProjectWorkTypeRepository) ListProjectCandidates(engineeringSkills []string) ([]model.ProjectCandidate, error) {
	candidates := []model.ProjectCandidate{}
	err := r.db.Select(&candidates, `
		SELECT e.id, e.name, e.role, e."isLeader", e."isTrainee",
			EXISTS (
				SELECT 1 FROM "EmployeeSkill" es
				JOIN "Skill" s ON s.id = es."skillId"
				WHERE es."employeeId" = e.id AND es."canPerform" = true AND s.name = ANY($1)
			) AS "hasEngSkill",
			EXISTS (
				SELECT 1 FROM "EmployeePermission" ep
				JOIN "Permission" p ON p.id = ep."permissionId"
				WHERE ep."employeeId" = e.id AND p.name = 'content_technician'
			) AS "isTechPerm"
		FROM "Employee" e
		-- المتدرب ما ينشال من القائمة: قبل جان ينختفي بالسكوت، فالمدير
		-- ينطي موظف جديد دور تقني وبعدين ما يلكاه بقائمة التوجيه ولا
		-- يعرف ليش. هسه يطلع مؤشّر «متدرب» والمدير هو الي يقرر.
		WHERE e.status = 'ACTIVE'
		ORDER BY e.name
	`, pq.Array(engineeringSkills))
	if err != nil {
		return nil, err
	}
	for i := range candidates {
		model.ClassifyProjectCandidate(&candidates[i])
	}
	return candidates, nil
}
