package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type VipCustomerRepository struct {
	db *sqlx.DB
}

func NewVipCustomerRepository(db *sqlx.DB) *VipCustomerRepository {
	return &VipCustomerRepository{db: db}
}

// vipSelect استعلام واحد بـJOIN يجيب كل تفاصيل العرض (اسم الزبون ورقمه، رمز
// الحجز، واسم الموظف الي علّمه) — بدون استعلام إضافي لكل صف.
const vipSelect = `
	SELECT v.*, c.name AS "customerName", c.phone AS "customerPhone",
		b.code AS "bookingCode", p.code AS "projectCode", p.name AS "projectName",
		e.name AS "markedByName"
	FROM "VipCustomer" v
	JOIN "Customer" c ON c.id = v."customerId"
	LEFT JOIN "Booking" b ON b.id = v."bookingId"
	LEFT JOIN "Project" p ON p.id = v."projectId"
	JOIN "Employee" e ON e.id = v."markedByEmployeeId"
`

func decorateVips(items []model.VipCustomer) []model.VipCustomer {
	for i := range items {
		items[i].SourceLabel = model.VipSourceLabels[items[i].Source]
	}
	return items
}

func (r *VipCustomerRepository) List() ([]model.VipCustomer, error) {
	items := []model.VipCustomer{}
	err := r.db.Select(&items, vipSelect+` ORDER BY v."createdAt" DESC`)
	return decorateVips(items), err
}

// ListCustomerIDs يرجّع معرّفات الزبائن المعلّمين — تستخدمها الواجهة لتبيّن
// أي زبون معلّم أصلاً (بدون كشف تفاصيل الـVIP لغير المدير).
func (r *VipCustomerRepository) ListCustomerIDs() ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `SELECT "customerId" FROM "VipCustomer"`)
	return ids, err
}

// Mark يعلّم الزبون كشخصية مهمة — لو معلّم أصلاً يحدّث السجل الموجود بدل ما يفشل.
func (r *VipCustomerRepository) Mark(id string, req model.MarkVipCustomerRequest, markedBy string) (*model.VipCustomer, error) {
	return r.markWithSource(id, req, markedBy, model.VipSourceManual)
}

func (r *VipCustomerRepository) markWithSource(id string, req model.MarkVipCustomerRequest, markedBy, source string) (*model.VipCustomer, error) {
	// المنصب يُخزن بسجل الزبون حتى يبقى معه حتى لو انشال من قائمة الشخصيات
	// المهمة ورجع انضاف. ولو ما انبعث منصب، ناخذ الي مخزون بسجل الزبون.
	if req.CustomerPosition != nil && *req.CustomerPosition != "" {
		_, _ = r.db.Exec(`UPDATE "Customer" SET position = $2 WHERE id = $1`, req.CustomerID, *req.CustomerPosition)
	} else {
		var pos *string
		if err := r.db.Get(&pos, `SELECT position FROM "Customer" WHERE id = $1`, req.CustomerID); err == nil && pos != nil {
			req.CustomerPosition = pos
		}
	}

	var saved model.VipCustomer
	if err := r.db.Get(&saved, `
		INSERT INTO "VipCustomer" (id, "customerId", "bookingId", "projectId", "requestSummary",
			"customerPosition", note, "markedByEmployeeId", source, "boughtFromUs")
		VALUES (COALESCE(NULLIF($1,''), gen_random_uuid()::text), $2, $3, $4, $5, $6, $7, $8, $9, COALESCE($10, true))
		ON CONFLICT ("customerId") DO UPDATE SET
			"bookingId" = COALESCE(EXCLUDED."bookingId", "VipCustomer"."bookingId"),
			"projectId" = COALESCE(EXCLUDED."projectId", "VipCustomer"."projectId"),
			"requestSummary" = COALESCE(EXCLUDED."requestSummary", "VipCustomer"."requestSummary"),
			"customerPosition" = COALESCE(EXCLUDED."customerPosition", "VipCustomer"."customerPosition"),
			note = COALESCE(EXCLUDED.note, "VipCustomer".note),
			"markedByEmployeeId" = EXCLUDED."markedByEmployeeId",
			"boughtFromUs" = EXCLUDED."boughtFromUs"
		RETURNING *
	`, id, req.CustomerID, req.BookingID, req.ProjectID, req.RequestSummary,
		req.CustomerPosition, req.Note, markedBy, source, req.BoughtFromUs); err != nil {
		return nil, err
	}
	var full model.VipCustomer
	if err := r.db.Get(&full, vipSelect+` WHERE v.id = $1`, saved.ID); err != nil {
		return nil, err
	}
	full.SourceLabel = model.VipSourceLabels[full.Source]
	return &full, nil
}

// MarkFromProject يرحّل صاحب المشروع تلقائياً للشخصيات المهمة.
//
// السبب: الناس الي يطلبون مشاريع دائماً ناس مهمين، فما ننتظر أحد يعلّمهم
// يدوياً. لو الزبون معلّم أصلاً، نكتفي بربط المشروع بسجله الموجود.
func (r *VipCustomerRepository) MarkFromProject(projectID, customerID, projectName string, position *string, markedBy string) error {
	if customerID == "" || markedBy == "" {
		return nil
	}
	summary := "مشروع: " + projectName
	_, err := r.markWithSource("", model.MarkVipCustomerRequest{
		CustomerID:       customerID,
		ProjectID:        &projectID,
		RequestSummary:   &summary,
		CustomerPosition: position,
	}, markedBy, model.VipSourceProject)
	return err
}

func (r *VipCustomerRepository) Unmark(customerID string) error {
	_, err := r.db.Exec(`DELETE FROM "VipCustomer" WHERE "customerId" = $1`, customerID)
	return err
}
