package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type InventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name, position FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

// ── Inventory Checks (جرد يومي) ──────────────────────────────────────────────

func (r *InventoryRepository) CreateInventoryCheck(employeeID string, req model.CreateInventoryCheckRequest) (*model.InventoryCheck, error) {
	var c model.InventoryCheck
	err := r.db.Get(&c, `
		INSERT INTO "InventoryCheck" (id, "employeeId", complete, "missingItems")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, employeeID, req.Complete, req.MissingItems)
	if err != nil {
		return nil, err
	}
	c.Employee = r.loadEmployeeBrief(c.EmployeeID)
	return &c, nil
}

// TodaysInventoryChecks يرجع آخر سجل جرد لكل موظف سجّل اليوم (للإداري يشوف مين ناقصه شي)
func (r *InventoryRepository) TodaysInventoryChecks() ([]model.InventoryCheck, error) {
	checks := []model.InventoryCheck{}
	if err := r.db.Select(&checks, `
		SELECT DISTINCT ON ("employeeId") *
		FROM "InventoryCheck"
		WHERE "checkedAt" >= CURRENT_DATE
		ORDER BY "employeeId", "checkedAt" DESC
	`); err != nil {
		return nil, err
	}
	for i := range checks {
		checks[i].Employee = r.loadEmployeeBrief(checks[i].EmployeeID)
		if checks[i].ResolvedByID != nil {
			checks[i].ResolvedBy = r.loadEmployeeBrief(*checks[i].ResolvedByID)
		}
	}
	return checks, nil
}

// ResolveInventoryCheck يؤشر إنو الإداري/الأدمن وفّر النقص المسجّل بهذا الجرد.
func (r *InventoryRepository) ResolveInventoryCheck(id string, resolvedByID string) (*model.InventoryCheck, error) {
	var c model.InventoryCheck
	err := r.db.Get(&c, `
		UPDATE "InventoryCheck"
		SET resolved = true, "resolvedById" = $2, "resolvedAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING *
	`, id, resolvedByID)
	if err != nil {
		return nil, err
	}
	c.Employee = r.loadEmployeeBrief(c.EmployeeID)
	c.ResolvedBy = r.loadEmployeeBrief(resolvedByID)
	return &c, nil
}

// ── Personal Tools ──────────────────────────────────────────────────────────

func (r *InventoryRepository) ListPersonalTools(employeeID string) ([]model.PersonalTool, error) {
	tools := []model.PersonalTool{}
	var err error
	if employeeID != "" {
		err = r.db.Select(&tools, `SELECT * FROM "PersonalTool" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID)
	} else {
		err = r.db.Select(&tools, `SELECT * FROM "PersonalTool" ORDER BY "createdAt" DESC`)
	}
	if err != nil {
		return nil, err
	}
	for i := range tools {
		tools[i].Employee = r.loadEmployeeBrief(tools[i].EmployeeID)
	}
	return tools, nil
}

func (r *InventoryRepository) CreatePersonalTool(employeeID, name, barcode string) (*model.PersonalTool, error) {
	var t model.PersonalTool
	err := r.db.Get(&t, `
		INSERT INTO "PersonalTool" (id, "employeeId", name, barcode)
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, employeeID, name, barcode)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) UpdatePersonalTool(id string, status *string, checkedOut *bool) (*model.PersonalTool, error) {
	var t model.PersonalTool
	err := r.db.Get(&t, `
		UPDATE "PersonalTool" SET
			status = COALESCE($2, status),
			"checkedOut" = COALESCE($3, "checkedOut")
		WHERE id = $1
		RETURNING *
	`, id, status, checkedOut)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) DeletePersonalTool(id string) error {
	_, err := r.db.Exec(`DELETE FROM "PersonalTool" WHERE id = $1`, id)
	return err
}

// ── Booking Tool Check (لقطة الأدوات الناقصة عند استلام حجز) ────────────────

// ListPersonalToolsByIDs يرجّع أدوات شخصية محددة بالمعرّفات — تستخدم لتحويل
// missingToolIds المرسلة من الواجهة لأسماء أدوات قابلة للقراءة عند تسجيل اللقطة.
func (r *InventoryRepository) ListPersonalToolsByIDs(ids []string) ([]model.PersonalTool, error) {
	tools := []model.PersonalTool{}
	if len(ids) == 0 {
		return tools, nil
	}
	err := r.db.Select(&tools, `SELECT * FROM "PersonalTool" WHERE id = ANY($1)`, pq.Array(ids))
	return tools, err
}

func (r *InventoryRepository) CreateBookingToolCheck(bookingID, employeeID string, missingItems *string) (*model.BookingToolCheck, error) {
	var c model.BookingToolCheck
	err := r.db.Get(&c, `
		INSERT INTO "BookingToolCheck" (id, "bookingId", "employeeId", "missingItems")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, bookingID, employeeID, missingItems)
	if err != nil {
		return nil, err
	}
	c.Employee = r.loadEmployeeBrief(c.EmployeeID)
	return &c, nil
}

// ListAllBookingToolChecks يرجّع كل لقطات الأدوات الناقصة عند استلام الحجوزات
// (بدون تصفية حسب حجز معين) — تستخدم بتبويب "تقارير النواقص" بصفحة المخزون.
func (r *InventoryRepository) ListAllBookingToolChecks() ([]model.BookingToolCheck, error) {
	checks := []model.BookingToolCheck{}
	err := r.db.Select(&checks, `SELECT * FROM "BookingToolCheck" ORDER BY "checkedAt" DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	for i := range checks {
		checks[i].Employee = r.loadEmployeeBrief(checks[i].EmployeeID)
	}
	return checks, nil
}

func (r *InventoryRepository) ListBookingToolChecks(bookingID string) ([]model.BookingToolCheck, error) {
	checks := []model.BookingToolCheck{}
	err := r.db.Select(&checks, `SELECT * FROM "BookingToolCheck" WHERE "bookingId" = $1 ORDER BY "checkedAt" DESC`, bookingID)
	if err != nil {
		return nil, err
	}
	for i := range checks {
		checks[i].Employee = r.loadEmployeeBrief(checks[i].EmployeeID)
	}
	return checks, nil
}

// ── Vehicle Tools ───────────────────────────────────────────────────────────

// vehicleToolSelect يجيب اسم ورقم لوحة السيارة مع الأداة باستعلام واحد، حتى
// الجدول يعرض اسم السيارة بدل معرّفها بدون استعلام إضافي لكل صف.
const vehicleToolSelect = `SELECT t.*,
	COALESCE(v.name, '') AS "vehicleName",
	COALESCE(v."plateNumber", '') AS "vehiclePlate"
	FROM "VehicleTool" t LEFT JOIN "Vehicle" v ON v.id = t."vehicleId"`

func (r *InventoryRepository) ListVehicleTools(vehicleID string) ([]model.VehicleTool, error) {
	tools := []model.VehicleTool{}
	var err error
	if vehicleID != "" {
		err = r.db.Select(&tools, vehicleToolSelect+` WHERE t."vehicleId" = $1 ORDER BY t."createdAt" DESC`, vehicleID)
	} else {
		err = r.db.Select(&tools, vehicleToolSelect+` ORDER BY t."createdAt" DESC`)
	}
	return tools, err
}

func (r *InventoryRepository) getVehicleTool(id string) (*model.VehicleTool, error) {
	var t model.VehicleTool
	if err := r.db.Get(&t, vehicleToolSelect+` WHERE t.id = $1`, id); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) CreateVehicleTool(name string, barcode *string, quantity int, vehicleID string) (*model.VehicleTool, error) {
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "VehicleTool" (id, name, barcode, quantity, "vehicleId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING id
	`, name, barcode, quantity, vehicleID)
	if err != nil {
		return nil, err
	}
	return r.getVehicleTool(id)
}

func (r *InventoryRepository) UpdateVehicleTool(id string, name, barcode, status, vehicleID *string, quantity *int) (*model.VehicleTool, error) {
	_, err := r.db.Exec(`
		UPDATE "VehicleTool" SET
			name = COALESCE($2, name),
			barcode = COALESCE($3, barcode),
			status = COALESCE($4, status),
			"vehicleId" = COALESCE($5, "vehicleId"),
			quantity = COALESCE($6, quantity)
		WHERE id = $1
	`, id, name, barcode, status, vehicleID, quantity)
	if err != nil {
		return nil, err
	}
	return r.getVehicleTool(id)
}

func (r *InventoryRepository) DeleteVehicleTool(id string) error {
	_, err := r.db.Exec(`DELETE FROM "VehicleTool" WHERE id = $1`, id)
	return err
}

// ── On-demand Tools ─────────────────────────────────────────────────────────

func (r *InventoryRepository) ListOnDemandTools() ([]model.OnDemandTool, error) {
	tools := []model.OnDemandTool{}
	err := r.db.Select(&tools, `SELECT * FROM "OnDemandTool" ORDER BY name ASC`)
	return tools, err
}

func (r *InventoryRepository) GetOnDemandTool(id string) (*model.OnDemandTool, error) {
	var t model.OnDemandTool
	if err := r.db.Get(&t, `SELECT * FROM "OnDemandTool" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) CreateOnDemandTool(name, barcode string, totalQty, availableQty int) (*model.OnDemandTool, error) {
	var t model.OnDemandTool
	err := r.db.Get(&t, `
		INSERT INTO "OnDemandTool" (id, name, barcode, "totalQuantity", "availableQuantity")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, name, barcode, totalQty, availableQty)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) UpdateOnDemandTool(id string, name, barcode, status *string, totalQty, availableQty *int) (*model.OnDemandTool, error) {
	var t model.OnDemandTool
	err := r.db.Get(&t, `
		UPDATE "OnDemandTool" SET
			name = COALESCE($2, name),
			barcode = COALESCE($3, barcode),
			"totalQuantity" = COALESCE($4, "totalQuantity"),
			"availableQuantity" = COALESCE($5, "availableQuantity"),
			status = COALESCE($6, status)
		WHERE id = $1
		RETURNING *
	`, id, name, barcode, totalQty, availableQty, status)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) IncrementAvailableQuantity(id string) error {
	_, err := r.db.Exec(`UPDATE "OnDemandTool" SET "availableQuantity" = "availableQuantity" + 1 WHERE id = $1`, id)
	return err
}

// ── Tool Requests ───────────────────────────────────────────────────────────

func (r *InventoryRepository) hydrateRequest(req *model.ToolRequest) {
	req.Employee = r.loadEmployeeBrief(req.EmployeeID)
	if tool, err := r.GetOnDemandTool(req.ToolID); err == nil {
		req.Tool = tool
	}
	if req.ApprovedByID != nil {
		req.ApprovedBy = r.loadEmployeeBrief(*req.ApprovedByID)
	}
	if req.Reason != nil {
		req.ReasonLabel = model.ToolRequestReasonLabels[*req.Reason]
	}
}

func (r *InventoryRepository) ListToolRequests(employeeID string) ([]model.ToolRequest, error) {
	requests := []model.ToolRequest{}
	var err error
	if employeeID != "" {
		err = r.db.Select(&requests, `SELECT * FROM "ToolRequest" WHERE "employeeId" = $1 ORDER BY "requestedAt" DESC`, employeeID)
	} else {
		err = r.db.Select(&requests, `SELECT * FROM "ToolRequest" ORDER BY "requestedAt" DESC`)
	}
	if err != nil {
		return nil, err
	}
	for i := range requests {
		r.hydrateRequest(&requests[i])
	}
	return requests, nil
}

func (r *InventoryRepository) CreateToolRequest(employeeID, toolID, reason string, description *string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	err := r.db.Get(&req, `
		INSERT INTO "ToolRequest" (id, "employeeId", "toolId", reason, description)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, employeeID, toolID, reason, description)
	if err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

func (r *InventoryRepository) DeleteToolRequest(id string) error {
	_, err := r.db.Exec(`DELETE FROM "ToolRequest" WHERE id = $1`, id)
	return err
}

func (r *InventoryRepository) GetToolRequest(id string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	if err := r.db.Get(&req, `SELECT * FROM "ToolRequest" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

// ApproveToolRequest يوافق على الطلب، ولو الأداة انشترت (مو متوفرة بالمخزن)
// يخزن سعرها ومعرّف طلب المشتريات المتولّد حتى تنربط السلسلة كاملة.
func (r *InventoryRepository) ApproveToolRequest(id, approvedByID string, purchasePrice *float64, procurementRequestID *string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	err := r.db.Get(&req, `
		UPDATE "ToolRequest" SET status = 'APPROVED', "approvedById" = $2, "approvedAt" = now(),
			"purchasePrice" = COALESCE($3, "purchasePrice"),
			"procurementRequestId" = COALESCE($4, "procurementRequestId")
		WHERE id = $1
		RETURNING *
	`, id, approvedByID, purchasePrice, procurementRequestID)
	if err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

func (r *InventoryRepository) RejectToolRequest(id string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	err := r.db.Get(&req, `
		UPDATE "ToolRequest" SET status = 'REJECTED'
		WHERE id = $1
		RETURNING *
	`, id)
	if err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

// ── Personal Tool Template (العدة القياسية) ─────────────────────────────────

func (r *InventoryRepository) ListPersonalToolTemplateItems() ([]model.PersonalToolTemplateItem, error) {
	items := []model.PersonalToolTemplateItem{}
	err := r.db.Select(&items, `SELECT * FROM "PersonalToolTemplateItem" ORDER BY "createdAt" ASC`)
	return items, err
}

// CreatePersonalToolTemplateItem يضيف عنصر جديد للعدة القياسية، وفوراً ينشئ
// PersonalTool مطابق لكل موظف حالي (barcode مولّد تلقائياً لأنه عمود NOT NULL
// وما عنده معنى حقيقي هون — التتبع الفريد صار بالعدة القياسية نفسها، مو بالباركود).
func (r *InventoryRepository) CreatePersonalToolTemplateItem(name string) (*model.PersonalToolTemplateItem, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var item model.PersonalToolTemplateItem
	if err := tx.Get(&item, `
		INSERT INTO "PersonalToolTemplateItem" (id, name)
		VALUES (gen_random_uuid()::text, $1)
		RETURNING *
	`, name); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(`
		INSERT INTO "PersonalTool" (id, "employeeId", name, barcode)
		SELECT gen_random_uuid()::text, e.id, $1, 'TPL-' || substr(gen_random_uuid()::text, 1, 12)
		FROM "Employee" e
	`, name); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *InventoryRepository) DeletePersonalToolTemplateItem(id string) error {
	_, err := r.db.Exec(`DELETE FROM "PersonalToolTemplateItem" WHERE id = $1`, id)
	return err
}

// ApplyPersonalToolTemplateToEmployee يضيف PersonalTool لكل عنصر بالعدة القياسية
// لموظف واحد — يُستدعى فور إنشاء موظف جديد (employee_service.go) حتى ياخذ
// القائمة كاملة تلقائياً بدون أي خطوة يدوية من الإداري.
func (r *InventoryRepository) ApplyPersonalToolTemplateToEmployee(employeeID string) error {
	_, err := r.db.Exec(`
		INSERT INTO "PersonalTool" (id, "employeeId", name, barcode)
		SELECT gen_random_uuid()::text, $1, t.name, 'TPL-' || substr(gen_random_uuid()::text, 1, 12)
		FROM "PersonalToolTemplateItem" t
	`, employeeID)
	return err
}

// ── Vehicle Tool Check (لقطة أدوات المركبة الناقصة عند بدء مهمة من ليدر) ────

func (r *InventoryRepository) CreateVehicleToolCheck(vehicleID, missionID, employeeID string, missingToolNames *string) (*model.VehicleToolCheck, error) {
	var c model.VehicleToolCheck
	err := r.db.Get(&c, `
		INSERT INTO "VehicleToolCheck" (id, "vehicleId", "missionId", "employeeId", "missingToolNames")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, vehicleID, missionID, employeeID, missingToolNames)
	if err != nil {
		return nil, err
	}
	c.Employee = r.loadEmployeeBrief(c.EmployeeID)
	return &c, nil
}

func (r *InventoryRepository) ListVehicleToolChecks() ([]model.VehicleToolCheck, error) {
	checks := []model.VehicleToolCheck{}
	err := r.db.Select(&checks, `SELECT * FROM "VehicleToolCheck" ORDER BY "createdAt" DESC`)
	if err != nil {
		return nil, err
	}
	for i := range checks {
		checks[i].Employee = r.loadEmployeeBrief(checks[i].EmployeeID)
	}
	return checks, nil
}

func (r *InventoryRepository) ReturnToolRequest(id string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	err := r.db.Get(&req, `
		UPDATE "ToolRequest" SET status = 'RETURNED', "returnedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id)
	if err != nil {
		return nil, err
	}
	if err := r.IncrementAvailableQuantity(req.ToolID); err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}
