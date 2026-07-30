package repository

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type QuotationRepository struct {
	db *sqlx.DB
}

func NewQuotationRepository(db *sqlx.DB) *QuotationRepository {
	return &QuotationRepository{db: db}
}

func (r *QuotationRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *QuotationRepository) hydrate(q *model.Quotation) error {
	items := []model.QuotationItem{}
	if err := r.db.Select(&items, `SELECT * FROM "QuotationItem" WHERE "quotationId" = $1`, q.ID); err != nil {
		return err
	}
	q.Items = items
	q.CreatedByEmployee = r.loadEmployeeBrief(q.CreatedByEmployeeID)
	return nil
}

// List يرجّع كل عروض الأسعار — search (اختياري) يفلتر باسم الزبون أو المشروع
// أو رقم العرض (بحث سلس، جزئي، case-insensitive).
func (r *QuotationRepository) List(search string) ([]model.Quotation, error) {
	quotations := []model.Quotation{}
	var err error
	if search == "" {
		err = r.db.Select(&quotations, `SELECT * FROM "Quotation" ORDER BY "createdAt" DESC`)
	} else {
		err = r.db.Select(&quotations, `
			SELECT * FROM "Quotation"
			WHERE "customerName" ILIKE '%' || $1 || '%'
				OR "projectName" ILIKE '%' || $1 || '%'
				OR "quotationNumber" ILIKE '%' || $1 || '%'
			ORDER BY "createdAt" DESC
		`, search)
	}
	if err != nil {
		return nil, err
	}
	for i := range quotations {
		if err := r.hydrate(&quotations[i]); err != nil {
			return nil, err
		}
	}
	return quotations, nil
}

func (r *QuotationRepository) FindByID(id string) (*model.Quotation, error) {
	var q model.Quotation
	err := r.db.Get(&q, `SELECT * FROM "Quotation" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&q); err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuotationRepository) nextQuotationNumber(year int) (string, error) {
	prefix := "AM-" + strconv.Itoa(year) + "-"
	var lastNumber string
	err := r.db.Get(&lastNumber, `
		SELECT "quotationNumber" FROM "Quotation"
		WHERE "quotationNumber" LIKE $1
		ORDER BY "quotationNumber" DESC LIMIT 1
	`, prefix+"%")
	seq := 1
	if err == nil {
		parts := strings.Split(lastNumber, "-")
		if len(parts) == 3 {
			if n, convErr := strconv.Atoi(parts[2]); convErr == nil {
				seq = n + 1
			}
		}
	}
	return prefix + padLeft(strconv.Itoa(seq), 4, '0'), nil
}

func padLeft(s string, length int, pad byte) string {
	for len(s) < length {
		s = string(pad) + s
	}
	return s
}

func (r *QuotationRepository) insertItems(tx *sqlx.Tx, quotationID string, items []model.QuotationItemInput) error {
	for _, item := range items {
		totalPrice := float64(item.Quantity) * item.UnitPrice
		if _, err := tx.Exec(`
			INSERT INTO "QuotationItem" (id, "quotationId", "productName", unit, quantity, "unitPrice", "totalPrice")
			VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		`, quotationID, item.ProductName, item.Unit, item.Quantity, item.UnitPrice, totalPrice); err != nil {
			return err
		}
	}
	return nil
}

func computeTotals(items []model.QuotationItemInput, discountPercent float64) (grandTotal, discountValue, netTotal float64) {
	for _, item := range items {
		grandTotal += float64(item.Quantity) * item.UnitPrice
	}
	discountValue = grandTotal * (discountPercent / 100)
	netTotal = grandTotal - discountValue
	return
}

func (r *QuotationRepository) Create(req model.CreateQuotationRequest) (*model.Quotation, error) {
	year := time.Now().Year()
	quotationNumber, err := r.nextQuotationNumber(year)
	if err != nil {
		return nil, err
	}

	discountPercent := 0.0
	if req.DiscountPercent != nil {
		discountPercent = *req.DiscountPercent
	}
	grandTotal, discountValue, netTotal := computeTotals(req.Items, discountPercent)

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var q model.Quotation
	err = tx.Get(&q, `
		INSERT INTO "Quotation" (id, "quotationNumber", "customerName", "customerPhone", "customerAddress", "projectName", "discountPercent", "discountValue", "grandTotal", "netTotal", notes, duration, "createdByEmployeeId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING *
	`, quotationNumber, req.CustomerName, req.CustomerPhone, req.CustomerAddress, req.ProjectName, discountPercent, discountValue, grandTotal, netTotal, req.Notes, req.Duration, req.CreatedByEmployeeID)
	if err != nil {
		return nil, err
	}

	if err := r.insertItems(tx, q.ID, req.Items); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if err := r.hydrate(&q); err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuotationRepository) Update(id string, req model.UpdateQuotationRequest) (*model.Quotation, error) {
	existing, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("Quotation not found")
	}

	discountPercent := existing.DiscountPercent
	if req.DiscountPercent != nil {
		discountPercent = *req.DiscountPercent
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var grandTotal, discountValue, netTotal *float64
	if req.Items != nil {
		if _, err := tx.Exec(`DELETE FROM "QuotationItem" WHERE "quotationId" = $1`, id); err != nil {
			return nil, err
		}
		if err := r.insertItems(tx, id, req.Items); err != nil {
			return nil, err
		}
		gt, dv, nt := computeTotals(req.Items, discountPercent)
		grandTotal, discountValue, netTotal = &gt, &dv, &nt
	}

	var q model.Quotation
	err = tx.Get(&q, `
		UPDATE "Quotation" SET
			"customerName" = COALESCE($2, "customerName"),
			"customerPhone" = COALESCE($3, "customerPhone"),
			"customerAddress" = COALESCE($4, "customerAddress"),
			"projectName" = COALESCE($5, "projectName"),
			"discountPercent" = COALESCE($6, "discountPercent"),
			notes = COALESCE($7, notes),
			duration = COALESCE($8, duration),
			status = COALESCE($9, status),
			"grandTotal" = COALESCE($10, "grandTotal"),
			"discountValue" = COALESCE($11, "discountValue"),
			"netTotal" = COALESCE($12, "netTotal")
		WHERE id = $1
		RETURNING *
	`, id, req.CustomerName, req.CustomerPhone, req.CustomerAddress, req.ProjectName,
		req.DiscountPercent, req.Notes, req.Duration, req.Status, grandTotal, discountValue, netTotal)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if err := r.hydrate(&q); err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuotationRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Quotation" WHERE id = $1`, id)
	return err
}
