package model

import "time"

type Vehicle struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	PlateNumber string    `db:"plateNumber" json:"plateNumber"`
	Color       *string   `db:"color" json:"color"`
	Type        *string   `db:"type" json:"type"`
	IsActive    bool      `db:"isActive" json:"isActive"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
}

type CreateVehicleRequest struct {
	Name        string  `json:"name"`
	PlateNumber string  `json:"plateNumber"`
	Color       *string `json:"color"`
	Type        *string `json:"type"`
}

// VehicleLog يغطي وقود/تنظيف/تبديل زيت — سجل واحد بنوع محدد لكل حدث
type VehicleLog struct {
	ID            string     `db:"id" json:"id"`
	VehicleID     string     `db:"vehicleId" json:"vehicleId"`
	Type          string     `db:"type" json:"type"` // FUEL | CLEANING | OIL_CHANGE
	PerformedAt   time.Time  `db:"performedAt" json:"performedAt"`
	NextDueAt     *time.Time `db:"nextDueAt" json:"nextDueAt"`
	Odometer      *int       `db:"odometer" json:"odometer"`
	Cost          *float64   `db:"cost" json:"cost"`
	Notes         *string    `db:"notes" json:"notes"`
	RecordedByID  *string    `db:"recordedById" json:"-"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	RecordedBy    *EmployeeBrief `db:"-" json:"recordedBy"`
}

type CreateVehicleLogRequest struct {
	Type        string   `json:"type"`
	PerformedAt *string  `json:"performedAt"`
	NextDueAt   *string  `json:"nextDueAt"`
	Odometer    *int     `json:"odometer"`
	Cost        *float64 `json:"cost"`
	Notes       *string  `json:"notes"`
}

// VehicleIncident يغطي الأعطال والأضرار (صدمات) مع تحديد المسبب والتكلفة
type VehicleIncident struct {
	ID                    string     `db:"id" json:"id"`
	VehicleID             string     `db:"vehicleId" json:"vehicleId"`
	Type                  string     `db:"type" json:"type"` // FAULT | DAMAGE
	Description           string     `db:"description" json:"description"`
	ResponsibleEmployeeID *string    `db:"responsibleEmployeeId" json:"-"`
	Cost                  *float64   `db:"cost" json:"cost"`
	Status                string     `db:"status" json:"status"` // OPEN | RESOLVED
	ReportedByID          *string    `db:"reportedById" json:"-"`
	CreatedAt             time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt            *time.Time `db:"resolvedAt" json:"resolvedAt"`

	ResponsibleEmployee *EmployeeBrief `db:"-" json:"responsibleEmployee"`
	ReportedBy          *EmployeeBrief `db:"-" json:"reportedBy"`
}

type CreateVehicleIncidentRequest struct {
	Type                  string   `json:"type"`
	Description           string   `json:"description"`
	ResponsibleEmployeeID *string  `json:"responsibleEmployeeId"`
	Cost                  *float64 `json:"cost"`
}

type UpdateVehicleIncidentRequest struct {
	Status *string  `json:"status"`
	Cost   *float64 `json:"cost"`
}

// VehicleMonthlyStatus يوثّق حالة كل سيارة شهرياً
type VehicleMonthlyStatus struct {
	ID               string    `db:"id" json:"id"`
	VehicleID        string    `db:"vehicleId" json:"vehicleId"`
	Month            string    `db:"month" json:"month"`
	HasIssue         bool      `db:"hasIssue" json:"hasIssue"`
	IssueDescription *string   `db:"issueDescription" json:"issueDescription"`
	Resolved         bool      `db:"resolved" json:"resolved"`
	Notes            *string   `db:"notes" json:"notes"`
	RecordedByID     *string   `db:"recordedById" json:"-"`
	CreatedAt        time.Time `db:"createdAt" json:"createdAt"`
}

type SetVehicleMonthlyStatusRequest struct {
	Month            string  `json:"month"`
	HasIssue         bool    `json:"hasIssue"`
	IssueDescription *string `json:"issueDescription"`
	Resolved         bool    `json:"resolved"`
	Notes            *string `json:"notes"`
}
