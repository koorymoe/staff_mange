package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type TeamInventoryCheckRepository struct {
	db *sqlx.DB
}

func NewTeamInventoryCheckRepository(db *sqlx.DB) *TeamInventoryCheckRepository {
	return &TeamInventoryCheckRepository{db: db}
}

func (r *TeamInventoryCheckRepository) loadEmployeeBrief(id *string) *model.EmployeeBrief {
	if id == nil {
		return nil
	}
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name, position FROM "Employee" WHERE id = $1`, *id); err != nil {
		return nil
	}
	return &brief
}

func (r *TeamInventoryCheckRepository) ListTools() ([]model.TeamInventoryToolCatalog, error) {
	tools := []model.TeamInventoryToolCatalog{}
	if err := r.db.Select(&tools, `SELECT * FROM "TeamInventoryToolCatalog" ORDER BY name ASC`); err != nil {
		return nil, err
	}
	return tools, nil
}

func (r *TeamInventoryCheckRepository) CreateTool(name string) (*model.TeamInventoryToolCatalog, error) {
	var t model.TeamInventoryToolCatalog
	err := r.db.Get(&t, `
		INSERT INTO "TeamInventoryToolCatalog" (id, name)
		VALUES (gen_random_uuid()::text, $1)
		RETURNING *
	`, name)
	return &t, err
}

func (r *TeamInventoryCheckRepository) hydrate(c *model.TeamInventoryCheck) error {
	leaderID := c.LeaderID
	c.Leader = r.loadEmployeeBrief(&leaderID)
	c.Employee1 = r.loadEmployeeBrief(c.Employee1ID)
	c.Employee2 = r.loadEmployeeBrief(c.Employee2ID)

	items := []model.TeamInventoryCheckItem{}
	if err := r.db.Select(&items, `SELECT * FROM "TeamInventoryCheckItem" WHERE "checkId" = $1 ORDER BY "toolName" ASC, "personRole" ASC`, c.ID); err != nil {
		return err
	}
	c.Items = items
	return nil
}

// Create ينشئ جلسة جرد فريق كاملة (الليدر + الموظفين المختارين + حالة كل أداة
// لكل شخص) بمعاملة واحدة (transaction) — إما تنحفظ كلها أو ولا وحدة.
func (r *TeamInventoryCheckRepository) Create(leaderID string, req model.CreateTeamInventoryCheckRequest) (*model.TeamInventoryCheck, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var c model.TeamInventoryCheck
	if err := tx.Get(&c, `
		INSERT INTO "TeamInventoryCheck" (id, "leaderId", "employee1Id", "employee2Id")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, leaderID, req.Employee1ID, req.Employee2ID); err != nil {
		return nil, err
	}

	for _, item := range req.Items {
		if _, err := tx.Exec(`
			INSERT INTO "TeamInventoryCheckItem" (id, "checkId", "toolName", "personRole", present, reason)
			VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		`, c.ID, item.ToolName, item.PersonRole, item.Present, item.Reason); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if err := r.hydrate(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *TeamInventoryCheckRepository) List() ([]model.TeamInventoryCheck, error) {
	checks := []model.TeamInventoryCheck{}
	if err := r.db.Select(&checks, `SELECT * FROM "TeamInventoryCheck" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range checks {
		if err := r.hydrate(&checks[i]); err != nil {
			return nil, err
		}
	}
	return checks, nil
}
