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

// List يجيب كل الموردين بـ3 استعلامات ثابتة بدل (1 + 2×عدد الموردين) —
// كان يسوي استعلامين لكل مورد (تخصصاته + كل تقييماته) وهذا سبب بطء الصفحة.
// هسه: استعلام للموردين، استعلام واحد لكل التخصصات، واستعلام واحد يحسب
// متوسط/عدد التقييمات بالسيرفر (بدون جلب كل صفوف التقييمات أصلاً).
func (r *SupplierRepository) List() ([]model.Supplier, error) {
	suppliers := []model.Supplier{}
	if err := r.db.Select(&suppliers, `SELECT * FROM "Supplier" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	if len(suppliers) == 0 {
		return suppliers, nil
	}

	ids := make([]string, len(suppliers))
	byID := make(map[string]*model.Supplier, len(suppliers))
	for i := range suppliers {
		ids[i] = suppliers[i].ID
		suppliers[i].Specialties = []model.SupplierSpecialty{}
		byID[suppliers[i].ID] = &suppliers[i]
	}

	var links []struct {
		SupplierID string `db:"supplierId"`
		model.SupplierSpecialty
	}
	if err := r.db.Select(&links, `
		SELECT l."supplierId", sp.*
		FROM "SupplierSpecialtyLink" l
		JOIN "SupplierSpecialty" sp ON sp.id = l."specialtyId"
		WHERE l."supplierId" = ANY($1)
	`, pq.Array(ids)); err != nil {
		return nil, err
	}
	for _, l := range links {
		if s := byID[l.SupplierID]; s != nil {
			s.Specialties = append(s.Specialties, l.SupplierSpecialty)
		}
	}

	var aggs []struct {
		SupplierID string  `db:"supplierId"`
		Count      int     `db:"cnt"`
		Avg        float64 `db:"avg"`
	}
	if err := r.db.Select(&aggs, `
		SELECT "supplierId", COUNT(*) AS cnt, COALESCE(AVG(value), 0) AS avg
		FROM "SupplierRating"
		WHERE "supplierId" = ANY($1)
		GROUP BY "supplierId"
	`, pq.Array(ids)); err != nil {
		return nil, err
	}
	for _, a := range aggs {
		if s := byID[a.SupplierID]; s != nil {
			s.RatingCount = a.Count
			s.AvgRating = a.Avg
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
