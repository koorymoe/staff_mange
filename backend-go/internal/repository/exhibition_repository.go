package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type ExhibitionRepository struct {
	db *sqlx.DB
}

func NewExhibitionRepository(db *sqlx.DB) *ExhibitionRepository {
	return &ExhibitionRepository{db: db}
}

func (r *ExhibitionRepository) hydrate(items []model.Exhibition) {
	for i := range items {
		e := &items[i]
		var creator model.Employee
		if err := r.db.Get(&creator, `SELECT * FROM "Employee" WHERE id = $1`, e.CreatedByID); err == nil {
			e.CreatedBy = &model.EmployeeBrief{ID: creator.ID, Name: creator.Name}
		}
		e.NominatedEmployee = []model.EmployeeBrief{}
		if len(e.NominatedEmployeeIDs) > 0 {
			nominees := []model.EmployeeBrief{}
			query, args, err := sqlx.In(`SELECT id, name FROM "Employee" WHERE id IN (?)`, []string(e.NominatedEmployeeIDs))
			if err == nil {
				query = r.db.Rebind(query)
				if err := r.db.Select(&nominees, query, args...); err == nil {
					e.NominatedEmployee = nominees
				}
			}
		}
	}
}

func (r *ExhibitionRepository) List() ([]model.Exhibition, error) {
	items := []model.Exhibition{}
	if err := r.db.Select(&items, `SELECT * FROM "Exhibition" ORDER BY archived ASC, "startDate" DESC`); err != nil {
		return nil, err
	}
	r.hydrate(items)
	return items, nil
}

func (r *ExhibitionRepository) FindByID(id string) (*model.Exhibition, error) {
	var e model.Exhibition
	if err := r.db.Get(&e, `SELECT * FROM "Exhibition" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	items := []model.Exhibition{e}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ExhibitionRepository) Create(id, title, location, startDate, endDate string, companies, productsToShow []string, createdByID string) (*model.Exhibition, error) {
	var e model.Exhibition
	err := r.db.Get(&e, `
		INSERT INTO "Exhibition" (id, title, location, "startDate", "endDate", companies, "productsToShow", "createdById")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`, id, title, location, startDate, endDate, pq.Array(companies), pq.Array(productsToShow), createdByID)
	if err != nil {
		return nil, err
	}
	items := []model.Exhibition{e}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ExhibitionRepository) Nominate(id string, employeeIDs []string) (*model.Exhibition, error) {
	var e model.Exhibition
	err := r.db.Get(&e, `
		UPDATE "Exhibition" SET "nominatedEmployeeIds" = $2 WHERE id = $1 RETURNING *
	`, id, pq.Array(employeeIDs))
	if err != nil {
		return nil, err
	}
	items := []model.Exhibition{e}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ExhibitionRepository) AddPhotos(id string, photoUrls []string) (*model.Exhibition, error) {
	var e model.Exhibition
	err := r.db.Get(&e, `
		UPDATE "Exhibition" SET "businessCardPhotos" = "businessCardPhotos" || $2 WHERE id = $1 RETURNING *
	`, id, pq.Array(photoUrls))
	if err != nil {
		return nil, err
	}
	items := []model.Exhibition{e}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ExhibitionRepository) SetFindings(id, keyFindings string) (*model.Exhibition, error) {
	var e model.Exhibition
	err := r.db.Get(&e, `
		UPDATE "Exhibition" SET "keyFindings" = $2 WHERE id = $1 RETURNING *
	`, id, keyFindings)
	if err != nil {
		return nil, err
	}
	items := []model.Exhibition{e}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ExhibitionRepository) SetVisitReport(id, report string) (*model.Exhibition, error) {
	var e model.Exhibition
	err := r.db.Get(&e, `
		UPDATE "Exhibition" SET "visitReport" = $2 WHERE id = $1 RETURNING *
	`, id, report)
	if err != nil {
		return nil, err
	}
	items := []model.Exhibition{e}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ExhibitionRepository) Archive(id string) (*model.Exhibition, error) {
	var e model.Exhibition
	err := r.db.Get(&e, `UPDATE "Exhibition" SET archived = true WHERE id = $1 RETURNING *`, id)
	if err != nil {
		return nil, err
	}
	items := []model.Exhibition{e}
	r.hydrate(items)
	return &items[0], nil
}
