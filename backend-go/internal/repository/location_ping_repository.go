package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type LocationPingRepository struct {
	db *sqlx.DB
}

func NewLocationPingRepository(db *sqlx.DB) *LocationPingRepository {
	return &LocationPingRepository{db: db}
}

func (r *LocationPingRepository) Create(employeeID string, req model.CreateLocationPingRequest) (*model.LocationPing, error) {
	var p model.LocationPing
	err := r.db.Get(&p, `
		INSERT INTO "LocationPing" (id, "employeeId", "bookingId", latitude, longitude)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *
	`, uuid.NewString(), employeeID, req.BookingID, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// LatestPerEmployee يرجع آخر نقطة موقع لكل موظف عنده نقاط بآخر ساعتين — لعرض "مين طالع الحين"
func (r *LocationPingRepository) LatestPerEmployee() ([]model.LocationPing, error) {
	rows := []model.LocationPing{}
	err := r.db.Select(&rows, `
		SELECT DISTINCT ON ("employeeId") *
		FROM "LocationPing"
		WHERE "createdAt" > now() - interval '2 hours'
		ORDER BY "employeeId", "createdAt" DESC
	`)
	return rows, err
}

// Path يرجع مسار موظف معين (اختيارياً لحجز معين) بترتيب زمني — يرسم كخط على الخريطة
func (r *LocationPingRepository) Path(employeeID string, bookingID *string) ([]model.LocationPing, error) {
	rows := []model.LocationPing{}
	var err error
	if bookingID != nil {
		err = r.db.Select(&rows, `
			SELECT * FROM "LocationPing" WHERE "employeeId" = $1 AND "bookingId" = $2 ORDER BY "createdAt"
		`, employeeID, *bookingID)
	} else {
		err = r.db.Select(&rows, `
			SELECT * FROM "LocationPing" WHERE "employeeId" = $1 AND "createdAt" > now() - interval '4 hours' ORDER BY "createdAt"
		`, employeeID)
	}
	return rows, err
}
