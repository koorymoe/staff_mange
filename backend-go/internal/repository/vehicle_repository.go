package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type VehicleRepository struct {
	db *sqlx.DB
}

func NewVehicleRepository(db *sqlx.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) loadEmployeeBrief(id *string) *model.EmployeeBrief {
	if id == nil {
		return nil
	}
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, *id); err != nil {
		return nil
	}
	return &brief
}

// ── Vehicle ──

func (r *VehicleRepository) List() ([]model.Vehicle, error) {
	vehicles := []model.Vehicle{}
	err := r.db.Select(&vehicles, `SELECT * FROM "Vehicle" ORDER BY "createdAt" DESC`)
	return vehicles, err
}

func (r *VehicleRepository) Create(req model.CreateVehicleRequest) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.Get(&v, `
		INSERT INTO "Vehicle" (id, name, "plateNumber", color, type)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, req.Name, req.PlateNumber, req.Color, req.Type)
	return &v, err
}

// ── VehicleLog (fuel / cleaning / oil change) ──

func (r *VehicleRepository) ListLogs(vehicleID string) ([]model.VehicleLog, error) {
	logs := []model.VehicleLog{}
	if err := r.db.Select(&logs, `SELECT * FROM "VehicleLog" WHERE "vehicleId" = $1 ORDER BY "performedAt" DESC`, vehicleID); err != nil {
		return nil, err
	}
	for i := range logs {
		logs[i].RecordedBy = r.loadEmployeeBrief(logs[i].RecordedByID)
	}
	return logs, nil
}

func (r *VehicleRepository) CreateLog(vehicleID string, req model.CreateVehicleLogRequest, recordedByID string) (*model.VehicleLog, error) {
	var l model.VehicleLog
	err := r.db.Get(&l, `
		INSERT INTO "VehicleLog" (id, "vehicleId", type, "performedAt", "nextDueAt", odometer, cost, notes, "recordedById")
		VALUES (gen_random_uuid()::text, $1, $2, COALESCE($3::timestamp, now()), $4::timestamp, $5, $6, $7, $8)
		RETURNING *
	`, vehicleID, req.Type, req.PerformedAt, req.NextDueAt, req.Odometer, req.Cost, req.Notes, recordedByID)
	if err != nil {
		return nil, err
	}
	l.RecordedBy = r.loadEmployeeBrief(l.RecordedByID)
	return &l, nil
}

// ── VehicleIncident (fault / damage) ──

func (r *VehicleRepository) ListIncidents(vehicleID string) ([]model.VehicleIncident, error) {
	incidents := []model.VehicleIncident{}
	if err := r.db.Select(&incidents, `SELECT * FROM "VehicleIncident" WHERE "vehicleId" = $1 ORDER BY "createdAt" DESC`, vehicleID); err != nil {
		return nil, err
	}
	for i := range incidents {
		incidents[i].ResponsibleEmployee = r.loadEmployeeBrief(incidents[i].ResponsibleEmployeeID)
		incidents[i].ReportedBy = r.loadEmployeeBrief(incidents[i].ReportedByID)
	}
	return incidents, nil
}

func (r *VehicleRepository) CreateIncident(vehicleID string, req model.CreateVehicleIncidentRequest, reportedByID string) (*model.VehicleIncident, error) {
	var inc model.VehicleIncident
	err := r.db.Get(&inc, `
		INSERT INTO "VehicleIncident" (id, "vehicleId", type, description, "responsibleEmployeeId", cost, "reportedById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, vehicleID, req.Type, req.Description, req.ResponsibleEmployeeID, req.Cost, reportedByID)
	if err != nil {
		return nil, err
	}
	inc.ResponsibleEmployee = r.loadEmployeeBrief(inc.ResponsibleEmployeeID)
	inc.ReportedBy = r.loadEmployeeBrief(inc.ReportedByID)
	return &inc, nil
}

func (r *VehicleRepository) UpdateIncident(id string, req model.UpdateVehicleIncidentRequest) (*model.VehicleIncident, error) {
	var resolvedAtExpr string
	if req.Status != nil && *req.Status == "RESOLVED" {
		resolvedAtExpr = "now()"
	} else {
		resolvedAtExpr = `"resolvedAt"`
	}
	var inc model.VehicleIncident
	err := r.db.Get(&inc, `
		UPDATE "VehicleIncident" SET
			status = COALESCE($2, status),
			cost = COALESCE($3, cost),
			"resolvedAt" = `+resolvedAtExpr+`
		WHERE id = $1
		RETURNING *
	`, id, req.Status, req.Cost)
	if err != nil {
		return nil, err
	}
	inc.ResponsibleEmployee = r.loadEmployeeBrief(inc.ResponsibleEmployeeID)
	inc.ReportedBy = r.loadEmployeeBrief(inc.ReportedByID)
	return &inc, nil
}

// ── VehicleMonthlyStatus ──

func (r *VehicleRepository) ListMonthlyStatus(vehicleID string) ([]model.VehicleMonthlyStatus, error) {
	statuses := []model.VehicleMonthlyStatus{}
	err := r.db.Select(&statuses, `SELECT * FROM "VehicleMonthlyStatus" WHERE "vehicleId" = $1 ORDER BY month DESC`, vehicleID)
	return statuses, err
}

func (r *VehicleRepository) SetMonthlyStatus(vehicleID string, req model.SetVehicleMonthlyStatusRequest, recordedByID string) (*model.VehicleMonthlyStatus, error) {
	var s model.VehicleMonthlyStatus
	err := r.db.Get(&s, `
		INSERT INTO "VehicleMonthlyStatus" (id, "vehicleId", month, "hasIssue", "issueDescription", resolved, notes, "recordedById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ("vehicleId", month) DO UPDATE SET
			"hasIssue" = EXCLUDED."hasIssue",
			"issueDescription" = EXCLUDED."issueDescription",
			resolved = EXCLUDED.resolved,
			notes = EXCLUDED.notes,
			"recordedById" = EXCLUDED."recordedById"
		RETURNING *
	`, vehicleID, req.Month, req.HasIssue, req.IssueDescription, req.Resolved, req.Notes, recordedByID)
	return &s, err
}
