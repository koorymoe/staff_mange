package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// p.* يجيب كل أعمدة المنتج، ونضم اسم الخدمة حتى الواجهة ما تحتاج
// استعلام ثاني لكل صف.
const productSelect = `SELECT p.*, s.name AS "serviceName"
	FROM "Product" p
	LEFT JOIN "Service" s ON s.id = p."serviceId"`

// serviceHints خدمات الشركة مع مهاراتها — أساس اقتراح النظام للتصنيف.
func (r *ProductRepository) serviceHints() []model.ServiceHint {
	var rows []struct {
		ID    string  `db:"id"`
		Name  string  `db:"name"`
		Skill *string `db:"skill"`
	}
	if err := r.db.Select(&rows, `
		SELECT s.id, s.name, k.name AS skill
		FROM "Service" s
		LEFT JOIN "Skill" k ON k."serviceId" = s.id
	`); err != nil {
		return nil // الاقتراح كماليات — فشله ما يوقف عرض المنتجات
	}
	byID := map[string]*model.ServiceHint{}
	order := []string{}
	for _, row := range rows {
		h, ok := byID[row.ID]
		if !ok {
			h = &model.ServiceHint{ID: row.ID, Name: row.Name, Terms: []string{row.Name}}
			byID[row.ID] = h
			order = append(order, row.ID)
		}
		if row.Skill != nil && *row.Skill != "" {
			h.Terms = append(h.Terms, *row.Skill)
		}
	}
	out := make([]model.ServiceHint, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out
}

// decorate يعبّي الحقول المحسوبة: عنوان التوفر، واقتراح النظام للتصنيف.
//
// الاقتراح يُحسب بالسيرفر مو بالواجهة، حتى ما يختلف من شاشة لشاشة.
func (r *ProductRepository) decorate(products []model.Product) []model.Product {
	hints := r.serviceHints()
	for i := range products {
		products[i].AvailabilityLabel = model.ProductAvailabilityLabels[products[i].Availability]
		if id, name, ok := model.SuggestServiceFor(products[i].Name, hints); ok {
			products[i].SuggestedServiceID = &id
			products[i].SuggestedServiceName = &name
		}
	}
	return products
}

func (r *ProductRepository) List() ([]model.Product, error) {
	products := []model.Product{}
	err := r.db.Select(&products, productSelect+` ORDER BY p.name ASC`)
	return r.decorate(products), err
}

func (r *ProductRepository) Get(id string) (*model.Product, error) {
	var p model.Product
	if err := r.db.Get(&p, productSelect+` WHERE p.id = $1`, id); err != nil {
		return nil, err
	}
	out := r.decorate([]model.Product{p})
	return &out[0], nil
}

func (r *ProductRepository) Create(req model.CreateProductRequest) (*model.Product, error) {
	availability := model.ProductInStock
	if req.Availability != nil && model.ValidProductAvailability(*req.Availability) {
		availability = *req.Availability
	}
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "Product" (id, name, unit, "defaultPrice", "wholesalePrice", "imageBase64",
			availability, "serviceId", specs, source, "modelName")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, NULLIF($7,''), $8, $9, $10)
		RETURNING id
	`, req.Name, req.Unit, req.DefaultPrice, req.WholesalePrice, req.ImageBase64, availability,
		derefStr(req.ServiceID), req.Specs, req.Source, req.ModelName)
	if err != nil {
		return nil, err
	}
	return r.Get(id)
}

func (r *ProductRepository) Update(id string, req model.UpdateProductRequest) (*model.Product, error) {
	// التوفر ما ينقبل إلا بقيمة صحيحة — أي شي غيرها نتجاهله ونبقي القديم
	var availability *string
	if req.Availability != nil && model.ValidProductAvailability(*req.Availability) {
		availability = req.Availability
	}
	_, err := r.db.Exec(`
		UPDATE "Product" SET
			name = COALESCE($2, name),
			unit = COALESCE($3, unit),
			"defaultPrice" = COALESCE($4, "defaultPrice"),
			"wholesalePrice" = COALESCE($5, "wholesalePrice"),
			"imageBase64" = COALESCE($6, "imageBase64"),
			availability = COALESCE($7, availability),
			"serviceId" = CASE WHEN $8 THEN NULL ELSE COALESCE(NULLIF($9,''), "serviceId") END,
			specs = COALESCE($10, specs),
			source = COALESCE($11, source),
			"modelName" = COALESCE($12, "modelName")
		WHERE id = $1
	`, id, req.Name, req.Unit, req.DefaultPrice, req.WholesalePrice, req.ImageBase64,
		availability, req.ClearService, derefStr(req.ServiceID), req.Specs, req.Source, req.ModelName)
	if err != nil {
		return nil, err
	}
	return r.Get(id)
}

func (r *ProductRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Product" WHERE id = $1`, id)
	return err
}
