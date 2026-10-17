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
const productSelect = `SELECT p.*, s.name AS "serviceName", e.name AS "createdByName"
	FROM "Product" p
	LEFT JOIN "Service" s ON s.id = p."serviceId"
	LEFT JOIN "Employee" e ON e.id = p."createdById"`

// decorate يعبّي عنوان «الحاجة» المقروء.
//
// اقتراح النظام للخدمة انشال: التقني هو الي يكتب الخدمة بنفسه، لأنه
// يعرف شغله أحسن من أي مطابقة أسماء.
func (r *ProductRepository) decorate(products []model.Product) []model.Product {
	for i := range products {
		products[i].AvailabilityLabel = model.ProductAvailabilityLabels[products[i].Availability]
	}
	return products
}

// List كل المنتجات — يستخدمها عرض السعر، لأنه لازم يشوف الكتالوج كامل.
func (r *ProductRepository) List() ([]model.Product, error) {
	products := []model.Product{}
	err := r.db.Select(&products, productSelect+` ORDER BY p.name ASC`)
	return r.decorate(products), err
}

// ListAdded المنتجات المضافة من شاشة «إضافة منتج» بس.
//
// شاشة التقنيين ما تعرض كتالوج عروض الأسعار القديم — هذاك مئات
// المنتجات وما تخص شغلهم. المضاف من عدهم إله createdById.
func (r *ProductRepository) ListAdded() ([]model.Product, error) {
	products := []model.Product{}
	err := r.db.Select(&products, productSelect+` WHERE p."createdById" IS NOT NULL ORDER BY p."createdAt" DESC`)
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

func (r *ProductRepository) Create(req model.CreateProductRequest, createdByID string) (*model.Product, error) {
	availability := model.ProductInStock
	if req.Availability != nil && model.ValidProductAvailability(*req.Availability) {
		availability = *req.Availability
	}
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "Product" (id, name, unit, "defaultPrice", "wholesalePrice", "imageBase64",
			availability, "serviceId", specs, source, "modelName", "serviceText", "createdById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, NULLIF($7,''), $8, $9, $10, $11, NULLIF($12,''))
		RETURNING id
	`, req.Name, req.Unit, req.DefaultPrice, req.WholesalePrice, req.ImageBase64, availability,
		derefStr(req.ServiceID), req.Specs, req.Source, req.ModelName, req.ServiceText, createdByID)
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
			"modelName" = COALESCE($12, "modelName"),
			"serviceText" = COALESCE($13, "serviceText")
		WHERE id = $1
	`, id, req.Name, req.Unit, req.DefaultPrice, req.WholesalePrice, req.ImageBase64,
		availability, req.ClearService, derefStr(req.ServiceID), req.Specs, req.Source, req.ModelName, req.ServiceText)
	if err != nil {
		return nil, err
	}
	return r.Get(id)
}

func (r *ProductRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Product" WHERE id = $1`, id)
	return err
}
