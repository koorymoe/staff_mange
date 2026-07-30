package repository

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type SupplierRepository struct {
	db *sqlx.DB
}

func NewSupplierRepository(db *sqlx.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// ── Specialties ─────────────────────────────────────────────────────────────

func (r *SupplierRepository) ListSpecialties() ([]model.SupplierSpecialty, error) {
	specs := []model.SupplierSpecialty{}
	err := r.db.Select(&specs, `SELECT * FROM "SupplierSpecialty" ORDER BY "order" ASC`)
	return specs, err
}

func (r *SupplierRepository) SpecialtyExists(name string) (bool, error) {
	var id string
	err := r.db.Get(&id, `SELECT id FROM "SupplierSpecialty" WHERE name = $1`, name)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *SupplierRepository) CountSpecialties() (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "SupplierSpecialty"`)
	return count, err
}

func (r *SupplierRepository) CreateSpecialty(name string, order int) (*model.SupplierSpecialty, error) {
	var s model.SupplierSpecialty
	err := r.db.Get(&s, `
		INSERT INTO "SupplierSpecialty" (id, name, "order")
		VALUES (gen_random_uuid()::text, $1, $2)
		RETURNING *
	`, name, order)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SupplierRepository) DeleteSpecialty(id string) error {
	_, err := r.db.Exec(`DELETE FROM "SupplierSpecialty" WHERE id = $1`, id)
	return err
}

// ── Suppliers ───────────────────────────────────────────────────────────────

func (r *SupplierRepository) hydrate(s *model.Supplier) error {
	specialties := []model.SupplierSpecialty{}
	err := r.db.Select(&specialties, `
		SELECT sp.* FROM "SupplierSpecialtyLink" l
		JOIN "SupplierSpecialty" sp ON sp.id = l."specialtyId"
		WHERE l."supplierId" = $1
	`, s.ID)
	if err != nil {
		return err
	}
	s.Specialties = specialties

	ratings := []model.SupplierRating{}
	if err := r.db.Select(&ratings, `SELECT * FROM "SupplierRating" WHERE "supplierId" = $1 ORDER BY "createdAt" DESC`, s.ID); err != nil {
		return err
	}
	s.RatingCount = len(ratings)
	if len(ratings) > 0 {
		sum := 0
		for _, rt := range ratings {
			sum += rt.Value
		}
		s.AvgRating = float64(sum) / float64(len(ratings))
	}
	return nil
}

func (r *SupplierRepository) List() ([]model.Supplier, error) {
	suppliers := []model.Supplier{}
	if err := r.db.Select(&suppliers, `SELECT * FROM "Supplier" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range suppliers {
		if err := r.hydrate(&suppliers[i]); err != nil {
			return nil, err
		}
	}
	return suppliers, nil
}

func (r *SupplierRepository) linkSpecialties(supplierID string, specialtyIDs []string) error {
	for _, sid := range specialtyIDs {
		if _, err := r.db.Exec(`
			INSERT INTO "SupplierSpecialtyLink" ("supplierId", "specialtyId") VALUES ($1, $2)
		`, supplierID, sid); err != nil {
			return err
		}
	}
	return nil
}

func (r *SupplierRepository) Create(req model.UpsertSupplierRequest) (*model.Supplier, error) {
	traderTypes := req.TraderTypes
	if traderTypes == nil {
		traderTypes = []string{}
	}
	var s model.Supplier
	err := r.db.Get(&s, `
		INSERT INTO "Supplier" (id, "companyName", "ownerName", phone, address, lat, lng, "isMaterialSupplier", "isContractor", "traderTypes", notes, "updatedAt")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
		RETURNING *
	`, req.CompanyName, req.OwnerName, req.Phone, req.Address, req.Lat, req.Lng, req.IsMaterialSupplier, req.IsContractor, pq.Array(traderTypes), req.Notes)
	if err != nil {
		return nil, err
	}
	if err := r.linkSpecialties(s.ID, req.SpecialtyIDs); err != nil {
		return nil, err
	}
	if err := r.hydrate(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SupplierRepository) Update(id string, req model.UpsertSupplierRequest) (*model.Supplier, error) {
	if _, err := r.db.Exec(`DELETE FROM "SupplierSpecialtyLink" WHERE "supplierId" = $1`, id); err != nil {
		return nil, err
	}
	traderTypes := req.TraderTypes
	if traderTypes == nil {
		traderTypes = []string{}
	}
	var s model.Supplier
	err := r.db.Get(&s, `
		UPDATE "Supplier" SET
			"companyName" = $2,
			"ownerName" = $3,
			phone = $4,
			address = $5,
			lat = $6,
			lng = $7,
			"isMaterialSupplier" = $8,
			"isContractor" = $9,
			"traderTypes" = $10,
			notes = $11,
			"updatedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, req.CompanyName, req.OwnerName, req.Phone, req.Address, req.Lat, req.Lng, req.IsMaterialSupplier, req.IsContractor, pq.Array(traderTypes), req.Notes)
	if err != nil {
		return nil, err
	}
	if err := r.linkSpecialties(s.ID, req.SpecialtyIDs); err != nil {
		return nil, err
	}
	if err := r.hydrate(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SupplierRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Supplier" WHERE id = $1`, id)
	return err
}

func (r *SupplierRepository) Rate(supplierID string, value int, note *string, ratedByID, ratedByName string) (*model.SupplierRating, error) {
	var rating model.SupplierRating
	err := r.db.Get(&rating, `
		INSERT INTO "SupplierRating" (id, "supplierId", value, note, "ratedById", "ratedByName")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, supplierID, value, note, ratedByID, ratedByName)
	if err != nil {
		return nil, err
	}
	return &rating, nil
}
