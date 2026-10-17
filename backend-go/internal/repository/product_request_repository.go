package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ProductRequestRepository struct {
	db *sqlx.DB
}

func NewProductRequestRepository(db *sqlx.DB) *ProductRequestRepository {
	return &ProductRequestRepository{db: db}
}

func (r *ProductRequestRepository) hydrate(items []model.ProductRequest) {
	for i := range items {
		p := &items[i]
		var requester model.Employee
		if err := r.db.Get(&requester, `SELECT * FROM "Employee" WHERE id = $1`, p.RequestedByID); err == nil {
			p.RequestedBy = &model.EmployeeBrief{ID: requester.ID, Name: requester.Name}
		}
		if p.ResolvedByID != nil {
			var resolver model.Employee
			if err := r.db.Get(&resolver, `SELECT * FROM "Employee" WHERE id = $1`, *p.ResolvedByID); err == nil {
				p.ResolvedBy = &model.EmployeeBrief{ID: resolver.ID, Name: resolver.Name}
			}
		}
	}
}

func (r *ProductRequestRepository) List() ([]model.ProductRequest, error) {
	items := []model.ProductRequest{}
	if err := r.db.Select(&items, `SELECT * FROM "ProductRequest" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	r.hydrate(items)
	return items, nil
}

func (r *ProductRequestRepository) Create(id string, req model.CreateProductProposalRequest, requestedByID string) (*model.ProductRequest, error) {
	var p model.ProductRequest
	err := r.db.Get(&p, `
		INSERT INTO "ProductRequest" (id, "requestedById", "productName", specs, source, model, category, price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`, id, requestedByID, req.ProductName, req.Specs, req.Source, req.Model, req.Category, req.Price)
	if err != nil {
		return nil, err
	}
	items := []model.ProductRequest{p}
	r.hydrate(items)
	return &items[0], nil
}

func (r *ProductRequestRepository) Resolve(id, status, resolvedByID string) (*model.ProductRequest, error) {
	var p model.ProductRequest
	err := r.db.Get(&p, `
		UPDATE "ProductRequest" SET status = $2, "resolvedAt" = now(), "resolvedById" = $3
		WHERE id = $1
		RETURNING *
	`, id, status, resolvedByID)
	if err != nil {
		return nil, err
	}
	items := []model.ProductRequest{p}
	r.hydrate(items)
	return &items[0], nil
}
