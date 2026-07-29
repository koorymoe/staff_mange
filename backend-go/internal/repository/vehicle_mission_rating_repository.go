package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type VehicleMissionRatingRepository struct {
	db *sqlx.DB
}

func NewVehicleMissionRatingRepository(db *sqlx.DB) *VehicleMissionRatingRepository {
	return &VehicleMissionRatingRepository{db: db}
}

func (r *VehicleMissionRatingRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *VehicleMissionRatingRepository) GetByMission(missionID string) (*model.VehicleMissionRating, error) {
	var rating model.VehicleMissionRating
	if err := r.db.Get(&rating, `SELECT * FROM "VehicleMissionRating" WHERE "missionId" = $1`, missionID); err != nil {
		return nil, err
	}
	rating.RatedBy = r.loadEmployeeBrief(rating.RatedByID)
	return &rating, nil
}

func (r *VehicleMissionRatingRepository) Create(missionID, ratedByID string, req model.CreateVehicleMissionRatingRequest) (*model.VehicleMissionRating, error) {
	var rating model.VehicleMissionRating
	err := r.db.Get(&rating, `
		INSERT INTO "VehicleMissionRating" (id, "missionId", "ratedById", commitment, "vehicleCare", driving, cleanliness, notes)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`, missionID, ratedByID, req.Commitment, req.VehicleCare, req.Driving, req.Cleanliness, req.Notes)
	if err != nil {
		return nil, err
	}
	rating.RatedBy = r.loadEmployeeBrief(rating.RatedByID)
	return &rating, nil
}

// GetCleanlinessAvgForDriverMonth يرجّع متوسط بند النظافة (cleanliness) من
// "تقييم السائقين بعد المهمة" لسائق معيّن خلال شهر معيّن (monthPrefix بصيغة
// "YYYY-MM")، مع عدد التقييمات — nil لو ما عنده أي مهام مقيَّمة بهذا الشهر.
func (r *VehicleMissionRatingRepository) GetCleanlinessAvgForDriverMonth(employeeID, monthPrefix string) (avg *float64, count int, err error) {
	var row struct {
		Count int      `db:"count"`
		Avg   *float64 `db:"avg"`
	}
	err = r.db.Get(&row, `
		SELECT COUNT(mr.id) AS count, AVG(mr.cleanliness) AS avg
		FROM "VehicleMission" m
		JOIN "VehicleMissionRating" mr ON mr."missionId" = m.id
		WHERE m."driverId" = $1 AND to_char(m."startedAt", 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	if err != nil {
		return nil, 0, err
	}
	return row.Avg, row.Count, nil
}

// GetCleanlinessAvgForDriverRange نفس GetCleanlinessAvgForDriverMonth لكن لمدى
// تاريخ حر (from/to بصيغة "YYYY-MM-DD").
func (r *VehicleMissionRatingRepository) GetCleanlinessAvgForDriverRange(employeeID, from, to string) (avg *float64, count int, err error) {
	var row struct {
		Count int      `db:"count"`
		Avg   *float64 `db:"avg"`
	}
	err = r.db.Get(&row, `
		SELECT COUNT(mr.id) AS count, AVG(mr.cleanliness) AS avg
		FROM "VehicleMission" m
		JOIN "VehicleMissionRating" mr ON mr."missionId" = m.id
		WHERE m."driverId" = $1 AND m."startedAt"::date BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	if err != nil {
		return nil, 0, err
	}
	return row.Avg, row.Count, nil
}

// GetDriverRatingSummary يحسب متوسط تقييمات سائق عبر كل مهامه المقيَّمة.
func (r *VehicleMissionRatingRepository) GetDriverRatingSummary(employeeID string) (*model.DriverRatingSummary, error) {
	var summary model.DriverRatingSummary
	err := r.db.Get(&summary, `
		SELECT
			COUNT(mr.id) AS "ratingsCount",
			COALESCE(AVG(mr.commitment), 0) AS "avgCommitment",
			COALESCE(AVG(mr."vehicleCare"), 0) AS "avgVehicleCare",
			COALESCE(AVG(mr.driving), 0) AS "avgDriving",
			COALESCE(AVG(mr.cleanliness), 0) AS "avgCleanliness",
			COALESCE(AVG((mr.commitment + mr."vehicleCare" + mr.driving + mr.cleanliness) / 4.0), 0) AS "avgOverall"
		FROM "VehicleMission" m
		JOIN "VehicleMissionRating" mr ON mr."missionId" = m.id
		WHERE m."driverId" = $1
	`, employeeID)
	if err != nil {
		return nil, err
	}
	summary.EmployeeID = employeeID
	return &summary, nil
}
