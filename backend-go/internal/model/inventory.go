package model

import "time"

type InventoryCheck struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	Complete     bool       `db:"complete" json:"complete"`
	MissingItems *string    `db:"missingItems" json:"missingItems"`
	CheckedAt    time.Time  `db:"checkedAt" json:"checkedAt"`
	Resolved     bool       `db:"resolved" json:"resolved"`
	ResolvedByID *string    `db:"resolvedById" json:"resolvedById"`
	ResolvedAt   *time.Time `db:"resolvedAt" json:"resolvedAt"`

	Employee   *EmployeeBrief `db:"-" json:"employee"`
	ResolvedBy *EmployeeBrief `db:"-" json:"resolvedBy"`
}

type CreateInventoryCheckRequest struct {
	Complete     bool    `json:"complete"`
	MissingItems *string `json:"missingItems"`
}

type PersonalTool struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	Name       string    `db:"name" json:"name"`
	Barcode    string    `db:"barcode" json:"barcode"`
	Status     string    `db:"status" json:"status"`
	CheckedOut bool      `db:"checkedOut" json:"checkedOut"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type CreatePersonalToolRequest struct {
	EmployeeID string `json:"employeeId"`
	Name       string `json:"name"`
	Barcode    string `json:"barcode"`
}

type UpdatePersonalToolRequest struct {
	Status     *string `json:"status"`
	CheckedOut *bool   `json:"checkedOut"`
}

type VehicleTool struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Barcode   string    `db:"barcode" json:"barcode"`
	VehicleID string    `db:"vehicleId" json:"vehicleId"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type CreateVehicleToolRequest struct {
	Name      string `json:"name"`
	Barcode   string `json:"barcode"`
	VehicleID string `json:"vehicleId"`
}

type UpdateVehicleToolRequest struct {
	Name    *string `json:"name"`
	Barcode *string `json:"barcode"`
	Status  *string `json:"status"`
}

type OnDemandTool struct {
	ID                string    `db:"id" json:"id"`
	Name              string    `db:"name" json:"name"`
	Barcode           string    `db:"barcode" json:"barcode"`
	TotalQuantity     int       `db:"totalQuantity" json:"totalQuantity"`
	AvailableQuantity int       `db:"availableQuantity" json:"availableQuantity"`
	Status            string    `db:"status" json:"status"`
	CreatedAt         time.Time `db:"createdAt" json:"createdAt"`
}

type CreateOnDemandToolRequest struct {
	Name              string `json:"name"`
	Barcode           string `json:"barcode"`
	TotalQuantity     int    `json:"totalQuantity"`
	AvailableQuantity int    `json:"availableQuantity"`
}

type UpdateOnDemandToolRequest struct {
	Name              *string `json:"name"`
	Barcode           *string `json:"barcode"`
	TotalQuantity     *int    `json:"totalQuantity"`
	AvailableQuantity *int    `json:"availableQuantity"`
	Status            *string `json:"status"`
}

type ToolRequest struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	ToolID       string     `db:"toolId" json:"toolId"`
	Status       string     `db:"status" json:"status"`
	ApprovedByID *string    `db:"approvedById" json:"approvedById"`
	RequestedAt  time.Time  `db:"requestedAt" json:"requestedAt"`
	ReturnedAt   *time.Time `db:"returnedAt" json:"returnedAt"`

	Employee   *EmployeeBrief `db:"-" json:"employee"`
	Tool       *OnDemandTool  `db:"-" json:"tool"`
	ApprovedBy *EmployeeBrief `db:"-" json:"approvedBy"`
}

type CreateToolRequestRequest struct {
	EmployeeID string `json:"employeeId"`
	ToolID     string `json:"toolId"`
}

type ApproveToolRequestRequest struct {
	ApprovedByID string `json:"approvedById"`
}
