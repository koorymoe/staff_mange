package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type VehicleMissionRepository struct {
	db *sqlx.DB
}

func NewVehicleMissionRepository(db *sqlx.DB) *VehicleMissionRepository {
	return &VehicleMissionRepository{db: db}
}

func (r *VehicleMissionRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *VehicleMissionRepository) loadVehicle(id string) *model.Vehicle {
	var v model.Vehicle
	if err := r.db.Get(&v, `SELECT * FROM "Vehicle" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &v
}

func (r *VehicleMissionRepository) loadPassengers(missionID string) []model.VehicleMissionPassenger {
	passengers := []model.VehicleMissionPassenger{}
	if err := r.db.Select(&passengers, `SELECT * FROM "VehicleMissionPassenger" WHERE "missionId" = $1`, missionID); err != nil {
		return passengers
	}
	for i := range passengers {
		passengers[i].Employee = r.loadEmployeeBrief(passengers[i].EmployeeID)
	}
	return passengers
}

func (r *VehicleMissionRepository) hydrate(m *model.VehicleMission) {
	m.Vehicle = r.loadVehicle(m.VehicleID)
	m.Driver = r.loadEmployeeBrief(m.DriverID)
	m.Passengers = r.loadPassengers(m.ID)
}

// GetActiveMissionForVehicle يرجع المهمة الفعالة (IN_PROGRESS) الحالية لسيارة معيّنة، لو موجودة.
func (r *VehicleMissionRepository) GetActiveMissionForVehicle(vehicleID string) (*model.VehicleMission, error) {
	var m model.VehicleMission
	err := r.db.Get(&m, `SELECT * FROM "VehicleMission" WHERE "vehicleId" = $1 AND status = 'IN_PROGRESS' LIMIT 1`, vehicleID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&m)
	return &m, nil
}

func (r *VehicleMissionRepository) Get(id string) (*model.VehicleMission, error) {
	var m model.VehicleMission
	if err := r.db.Get(&m, `SELECT * FROM "VehicleMission" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate(&m)
	return &m, nil
}

func (r *VehicleMissionRepository) StartMission(driverID, vehicleID, purpose, destination string, startOdometer int, passengerIDs []string) (*model.VehicleMission, error) {
	var m model.VehicleMission
	err := r.db.Get(&m, `
		INSERT INTO "VehicleMission" (id, "vehicleId", "driverId", purpose, destination, "startOdometer")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, vehicleID, driverID, purpose, destination, startOdometer)
	if err != nil {
		return nil, err
	}
	for _, empID := range passengerIDs {
		if _, err := r.db.Exec(`
			INSERT INTO "VehicleMissionPassenger" (id, "missionId", "employeeId")
			VALUES (gen_random_uuid()::text, $1, $2)
			ON CONFLICT ("missionId", "employeeId") DO NOTHING
		`, m.ID, empID); err != nil {
			return nil, err
		}
	}
	r.hydrate(&m)
	return &m, nil
}

func (r *VehicleMissionRepository) EndMission(missionID string, endOdometer int, notes *string) (*model.VehicleMission, error) {
	var m model.VehicleMission
	err := r.db.Get(&m, `
		UPDATE "VehicleMission" SET
			"endOdometer" = $2,
			"distanceKm" = $2 - "startOdometer",
			"endedAt" = now(),
			status = 'COMPLETED',
			notes = COALESCE($3, notes)
		WHERE id = $1
		RETURNING *
	`, missionID, endOdometer, notes)
	if err != nil {
		return nil, err
	}
	if _, err := r.db.Exec(`UPDATE "Vehicle" SET "currentOdometer" = $2 WHERE id = $1`, m.VehicleID, endOdometer); err != nil {
		return nil, err
	}
	r.hydrate(&m)
	return &m, nil
}

func (r *VehicleMissionRepository) List(filters model.VehicleMissionFilters) ([]model.VehicleMission, error) {
	missions := []model.VehicleMission{}
	clauses := []string{"1=1"}
	args := []any{}
	idx := 1
	add := func(clause string, val any) {
		clauses = append(clauses, fmt.Sprintf(clause, idx))
		args = append(args, val)
		idx++
	}
	if filters.VehicleID != nil && *filters.VehicleID != "" {
		add(`"vehicleId" = $%d`, *filters.VehicleID)
	}
	if filters.DriverID != nil && *filters.DriverID != "" {
		add(`"driverId" = $%d`, *filters.DriverID)
	}
	if filters.Status != nil && *filters.Status != "" {
		add(`status = $%d`, *filters.Status)
	}
	if filters.From != nil && *filters.From != "" {
		add(`"startedAt" >= $%d::timestamp`, *filters.From)
	}
	if filters.To != nil && *filters.To != "" {
		add(`"startedAt" <= $%d::timestamp`, *filters.To)
	}
	query := `SELECT * FROM "VehicleMission" WHERE ` + strings.Join(clauses, " AND ") + ` ORDER BY "startedAt" DESC`
	if err := r.db.Select(&missions, query, args...); err != nil {
		return nil, err
	}
	for i := range missions {
		r.hydrate(&missions[i])
	}
	return missions, nil
}
