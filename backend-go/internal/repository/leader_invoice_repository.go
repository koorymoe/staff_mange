package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// LeaderInvoiceRepository يخزّن فواتير الليدر وبنود المواد التابعة لها.
type LeaderInvoiceRepository struct {
	db *sqlx.DB
}

func NewLeaderInvoiceRepository(db *sqlx.DB) *LeaderInvoiceRepository {
	return &LeaderInvoiceRepository{db: db}
}

// Create يحفظ فاتورة ليدر جديدة مع كل بنود موادها بمعاملة واحدة (كلها تنجح أو
// كلها تفشل معاً).
func (r *LeaderInvoiceRepository) Create(inv *model.LeaderInvoice, materials []model.LeaderInvoiceMaterialItem) (*model.LeaderInvoice, error) {
	systemsJSON, err := json.Marshal(inv.Systems)
	if err != nil {
		return nil, err
	}
	itemsJSON, err := json.Marshal(inv.Items)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// الكود المحاسبي النهائي يعتمد على معرف الفاتورة (gen_random_uuid) وتاريخ
	// الإنشاء، وكلاهما لا يُعرف إلا بعد الإدراج — لذا نستخدم كوداً مؤقتاً فريداً
	// بالإدراج الأول ثم نحدّثه بالكود الحقيقي بنفس المعاملة.
	var saved model.LeaderInvoice
	err = tx.Get(&saved, `
		INSERT INTO "LeaderInvoice" (
			id, "bookingId", "employeeId", "customerName", "customerPhone", "customerAddress",
			systems, items, "totalDeviceCount", "executionCost", "materialsTotal",
			"discountValue", "netTotal", "accountingCode", status
		) VALUES (
			gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, gen_random_uuid()::text, $13
		) RETURNING *
	`,
		inv.BookingID, inv.EmployeeID, inv.CustomerName, inv.CustomerPhone, inv.CustomerAddress,
		string(systemsJSON), string(itemsJSON), inv.TotalDeviceCount, inv.ExecutionCost, inv.MaterialsTotal,
		inv.DiscountValue, inv.NetTotal, inv.Status,
	)
	if err != nil {
		return nil, err
	}

	for i := range materials {
		materials[i].LeaderInvoiceID = saved.ID
		var savedItem model.LeaderInvoiceMaterialItem
		err = tx.Get(&savedItem, `
			INSERT INTO "LeaderInvoiceMaterialItem" (
				id, "leaderInvoiceId", "materialId", name, quantity, "unitPrice", "profitPerUnit", "lineTotal"
			) VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7)
			RETURNING *
		`, materials[i].LeaderInvoiceID, materials[i].MaterialID, materials[i].Name,
			materials[i].Quantity, materials[i].UnitPrice, materials[i].ProfitPerUnit, materials[i].LineTotal)
		if err != nil {
			return nil, err
		}
		saved.Materials = append(saved.Materials, savedItem)
	}

	finalCode := model.GenerateAccountingCode(saved.ID, saved.CreatedAt)
	if _, err := tx.Exec(`UPDATE "LeaderInvoice" SET "accountingCode" = $2 WHERE id = $1`, saved.ID, finalCode); err != nil {
		return nil, err
	}
	saved.AccountingCode = finalCode

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	saved.Systems = inv.Systems
	saved.Items = inv.Items
	return &saved, nil
}

// CountForEmployeeMonth يرجّع عدد فواتير الليدر (المبيعات) لموظف معيّن خلال
// شهر معيّن (monthPrefix بصيغة "YYYY-MM").
func (r *LeaderInvoiceRepository) CountForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "LeaderInvoice"
		WHERE "employeeId" = $1 AND to_char("createdAt", 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return count, err
}

func (r *LeaderInvoiceRepository) hydrate(inv *model.LeaderInvoice) error {
	if inv.SystemsJSON != "" {
		_ = json.Unmarshal([]byte(inv.SystemsJSON), &inv.Systems)
	}
	if inv.ItemsJSON != "" {
		_ = json.Unmarshal([]byte(inv.ItemsJSON), &inv.Items)
	}
	materials := []model.LeaderInvoiceMaterialItem{}
	if err := r.db.Select(&materials, `SELECT * FROM "LeaderInvoiceMaterialItem" WHERE "leaderInvoiceId" = $1 ORDER BY "createdAt"`, inv.ID); err != nil {
		return err
	}
	inv.Materials = materials
	return nil
}

// GetByID يرجّع فاتورة واحدة مع بنودها.
func (r *LeaderInvoiceRepository) GetByID(id string) (*model.LeaderInvoice, error) {
	var inv model.LeaderInvoice
	err := r.db.Get(&inv, `SELECT * FROM "LeaderInvoice" WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

// List يرجّع الفواتير، مصفّاة حسب الموظف لو تم تمريره.
func (r *LeaderInvoiceRepository) List(employeeID string) ([]model.LeaderInvoice, error) {
	invoices := []model.LeaderInvoice{}
	var err error
	if employeeID == "" {
		err = r.db.Select(&invoices, `SELECT * FROM "LeaderInvoice" ORDER BY "createdAt" DESC`)
	} else {
		err = r.db.Select(&invoices, `SELECT * FROM "LeaderInvoice" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID)
	}
	if err != nil {
		return nil, err
	}
	for i := range invoices {
		if err := r.hydrate(&invoices[i]); err != nil {
			return nil, err
		}
	}
	return invoices, nil
}
