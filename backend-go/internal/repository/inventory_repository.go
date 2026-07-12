package repository

import (
	"github.com/jmoiron/sqlx"

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
	}
	return checks, nil
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

// ── Vehicle Tools ───────────────────────────────────────────────────────────

func (r *InventoryRepository) ListVehicleTools(vehicleID string) ([]model.VehicleTool, error) {
	tools := []model.VehicleTool{}
	var err error
	if vehicleID != "" {
		err = r.db.Select(&tools, `SELECT * FROM "VehicleTool" WHERE "vehicleId" = $1 ORDER BY "createdAt" DESC`, vehicleID)
	} else {
		err = r.db.Select(&tools, `SELECT * FROM "VehicleTool" ORDER BY "createdAt" DESC`)
	}
	return tools, err
}

func (r *InventoryRepository) CreateVehicleTool(name, barcode, vehicleID string) (*model.VehicleTool, error) {
	var t model.VehicleTool
	err := r.db.Get(&t, `
		INSERT INTO "VehicleTool" (id, name, barcode, "vehicleId")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, name, barcode, vehicleID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) UpdateVehicleTool(id string, name, barcode, status *string) (*model.VehicleTool, error) {
	var t model.VehicleTool
	err := r.db.Get(&t, `
		UPDATE "VehicleTool" SET
			name = COALESCE($2, name),
			barcode = COALESCE($3, barcode),
			status = COALESCE($4, status)
		WHERE id = $1
		RETURNING *
	`, id, name, barcode, status)
	if err != nil {
		return nil, err
	}
	return &t, nil
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

func (r *InventoryRepository) CreateToolRequest(employeeID, toolID string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	err := r.db.Get(&req, `
		INSERT INTO "ToolRequest" (id, "employeeId", "toolId")
		VALUES (gen_random_uuid()::text, $1, $2)
		RETURNING *
	`, employeeID, toolID)
	if err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

func (r *InventoryRepository) ApproveToolRequest(id, approvedByID string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	err := r.db.Get(&req, `
		UPDATE "ToolRequest" SET status = 'APPROVED', "approvedById" = $2
		WHERE id = $1
		RETURNING *
	`, id, approvedByID)
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
