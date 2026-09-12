package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ServiceRepository struct {
	db *sqlx.DB
}

func NewServiceRepository(db *sqlx.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) List() ([]model.Service, error) {
	services := []model.Service{}
	if err := r.db.Select(&services, `SELECT * FROM "Service" ORDER BY name ASC`); err != nil {
		return nil, err
	}

	for i := range services {
		skills := []model.Skill{}
		if err := r.db.Select(&skills, `SELECT * FROM "Skill" WHERE "serviceId" = $1 ORDER BY name ASC`, services[i].ID); err != nil {
			return nil, err
		}
		services[i].Skills = skills
	}

	return services, nil
}

// SetManagerHandlesPaperwork يأشّر/يشيل قاعدة «الورق على مسؤول الخدمة».
//
// ⚠️ قرار صاحب العمل بالبيانات مو بالكود: يأشّر الجي بي اس والداش كام
// من الشاشة، وأي خدمة جديدة تنضاف بضغطة — بلا نشر ولا تعديل كود.
func (r *ServiceRepository) SetManagerHandlesPaperwork(serviceID string, on bool) error {
	_, err := r.db.Exec(`UPDATE "Service" SET "managerHandlesPaperwork" = $1 WHERE id = $2`, on, serviceID)
	return err
}

func (r *ServiceRepository) Create(s *model.Service) error {
	_, err := r.db.NamedExec(`
		INSERT INTO "Service" (id, name, category, division)
		VALUES (:id, :name, :category, :division)
	`, s)
	return err
}

// CreateSkill يضيف مهارة جديدة لخدمة معيّنة — حتى الإداري يقدر يعرّف مهارات كل خدمة
// (بدل ما تكون بس خدمة "الهندسة" هي الوحيدة الي عندها مهارات جاهزة بالنظام).
func (r *ServiceRepository) CreateSkill(sk *model.Skill) error {
	_, err := r.db.NamedExec(`
		INSERT INTO "Skill" (id, name, "serviceId")
		VALUES (:id, :name, :serviceId)
		ON CONFLICT ("serviceId", name) DO NOTHING
	`, sk)
	return err
}

func (r *ServiceRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Service" WHERE id = $1`, id)
	return err
}

// AllSkills يرجّع كل المهارات بتصنيفها واسم خدمتها.
//
// ماكو مسار يجيب المهارات لحالها بالنظام — تنجاب دائماً مدفونة جوّا
// الخدمات. برامج التدريب تحتاجها قائمة مسطّحة عشان تختار «شنو راح
// يتعلّم المتدرّب».
func (r *ServiceRepository) AllSkills() ([]SkillWithService, error) {
	rows := []SkillWithService{}
	err := r.db.Select(&rows, `
		SELECT s.id, s.name, s.category, s.description,
		       s."serviceId", srv.name AS "serviceName"
		FROM "Skill" s
		LEFT JOIN "Service" srv ON srv.id = s."serviceId"
		ORDER BY s.category, srv.name NULLS LAST, s.name`)
	return rows, err
}

type SkillWithService struct {
	ID          string  `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	Category    string  `db:"category" json:"category"`
	Description *string `db:"description" json:"description"`
	ServiceID   *string `db:"serviceId" json:"serviceId"`
	ServiceName *string `db:"serviceName" json:"serviceName"`
}
