package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ProductProcurementRepository تجهيز طلبات المنتجات من الدوار.
//
// كل عملية تلمس رصيد الدوار تنمشي داخل معاملة واحدة مع سطر التجهيز،
// حتى ما يصير سطر بلا خصم ولا خصم بلا سطر.
type ProductProcurementRepository struct {
	db *sqlx.DB
}

func NewProductProcurementRepository(db *sqlx.DB) *ProductProcurementRepository {
	return &ProductProcurementRepository{db: db}
}

const procurementSelect = `SELECT p.*, f.name AS "fundName",
		s."companyName" AS "supplierName",
		b.name AS "purchasedByName", st.name AS "settledByName"
	FROM "ProductProcurement" p
	JOIN "RevolvingFund" f ON f.id = p."fundId"
	JOIN "Supplier" s ON s.id = p."supplierId"
	JOIN "Employee" b ON b.id = p."purchasedById"
	LEFT JOIN "Employee" st ON st.id = p."settledById"`

func decorateProcurements(rows []model.ProductProcurement) []model.ProductProcurement {
	for i := range rows {
		rows[i].StatusLabel = model.ProcurementStatusLabels[rows[i].Status]
		rows[i].PayerLabel = model.PayerKindLabels[rows[i].PayerKind]
	}
	return rows
}

// List التجهيزات. status فاضي = الكل.
func (r *ProductProcurementRepository) List(status string) ([]model.ProductProcurement, error) {
	rows := []model.ProductProcurement{}
	q := procurementSelect
	args := []any{}
	if status != "" {
		args = append(args, status)
		q += ` WHERE p.status = $1`
	}
	q += ` ORDER BY p."createdAt" DESC LIMIT 500`
	if err := r.db.Select(&rows, q, args...); err != nil {
		return nil, err
	}
	return decorateProcurements(rows), nil
}

func (r *ProductProcurementRepository) Get(id string) (*model.ProductProcurement, error) {
	var p model.ProductProcurement
	if err := r.db.Get(&p, procurementSelect+` WHERE p.id = $1`, id); err != nil {
		return nil, err
	}
	out := decorateProcurements([]model.ProductProcurement{p})
	return &out[0], nil
}

// Fulfill أبو الحسابات يجهّز الطلب: يشتري من مورد مضاف مسبقاً بفلوس
// الدوار، ويسجّل المبلغ والسبب والوصل ومنو يعوّض.
//
// المبلغ ينخصم من الدوار فوراً — لأنه فعلاً انصرف. الطلب يبقى PENDING
// لحد ما المحاسب يرجّع الفلوس بـSettle.
func (r *ProductProcurementRepository) Fulfill(requestID string, req model.FulfillProductRequest, purchasedByID string) (*model.ProductProcurement, error) {
	if req.SpentAmount <= 0 {
		return nil, fmt.Errorf("لازم تكتب المبلغ المصروف")
	}
	if strings.TrimSpace(req.Reason) == "" {
		return nil, fmt.Errorf("لازم تكتب سبب الصرف")
	}
	if req.ReceiptImage == nil || strings.TrimSpace(*req.ReceiptImage) == "" {
		return nil, fmt.Errorf("لازم ترفع صورة الوصل")
	}
	if req.SupplierID == "" {
		return nil, fmt.Errorf("لازم تحدد المورد الي اشتريت منه")
	}
	if req.FundID == "" {
		return nil, fmt.Errorf("لازم تحدد الدوار الي انصرف منه")
	}
	if !model.ValidPayerKind(req.PayerKind) {
		return nil, fmt.Errorf("لازم تحدد المنتج للشركة لو للزبون")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// الطلب لازم يكون معلّق — ما نجهّز طلب مرفوض ولا مجهّز من قبل
	var productName string
	if err := tx.Get(&productName, `
		SELECT "productName" FROM "ProductRequest"
		WHERE id = $1 AND status = 'PENDING'`, requestID); err != nil {
		return nil, fmt.Errorf("الطلب مو موجود أو انبتّ بيه من قبل")
	}

	// شرط balance >= المبلغ داخل التحديث نفسه — يمنع الصرف الزائد حتى
	// لو صرفين انطلبو بنفس اللحظة
	// RETURNING مع Get: لو الشرط ما انطبق ما يرجع ولا صف، فـGet يطيح —
	// وهذا بالضبط اللي نريده بدل ما نمشي بلا خصم.
	var newBalance float64
	if err := tx.Get(&newBalance, `
		UPDATE "RevolvingFund" SET balance = balance - $2, "updatedAt" = now()
		WHERE id = $1 AND "isActive" = true AND balance >= $2
		RETURNING balance`, req.FundID, req.SpentAmount); err != nil {
		return nil, fmt.Errorf("رصيد الدوار ما يكفي، أو الدوار موقوف")
	}

	var p model.ProductProcurement
	if err := tx.Get(&p, `
		INSERT INTO "ProductProcurement"
			(id, "requestId", "fundId", "supplierId", "productName", "spentAmount", reason,
			 "receiptImage", "payerKind", "customerNote", "bookingId", status, "purchasedById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9,''), $10, 'PENDING', $11)
		RETURNING *`,
		requestID, req.FundID, req.SupplierID, productName, req.SpentAmount, strings.TrimSpace(req.Reason),
		req.ReceiptImage, req.PayerKind, derefStr(req.CustomerNote), req.BookingID, purchasedByID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(p.ID)
}

// Settle المحاسب يرجّع المبلغ للدوار — وهنا بس يخلص الطلب.
//
// المبلغ يرجع للدوار وتنكتب حركة TOPUP حتى يبقى أثر بكشف الدوار.
// شرط status='PENDING' بالتحديث يمنع الإرجاع مرتين.
func (r *ProductProcurementRepository) Settle(id string, note *string, byID string) (*model.ProductProcurement, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var p model.ProductProcurement
	if err := tx.Get(&p, `
		UPDATE "ProductProcurement"
		SET status = 'SETTLED', "settledById" = $2, "settledAt" = now(), "settleNote" = NULLIF($3,'')
		WHERE id = $1 AND status = 'PENDING'
		RETURNING *`, id, byID, derefStr(note)); err != nil {
		return nil, fmt.Errorf("الطلب مو معلّق — يمكن انرجع المبلغ من قبل")
	}

	if _, err := tx.Exec(`
		UPDATE "RevolvingFund" SET balance = balance + $2, "updatedAt" = now()
		WHERE id = $1`, p.FundID, p.SpentAmount); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`
		INSERT INTO "RevolvingFundTxn" (id, "fundId", kind, amount, notes, status, "createdById")
		VALUES (gen_random_uuid()::text, $1, 'TOPUP', $2, $3, 'APPROVED', $4)`,
		p.FundID, p.SpentAmount,
		fmt.Sprintf("إرجاع مبلغ تجهيز منتج: %s", p.ProductName), byID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(p.ID)
}
