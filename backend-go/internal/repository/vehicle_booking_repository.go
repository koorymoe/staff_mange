package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type VehicleBookingRepository struct {
	db *sqlx.DB
}

func NewVehicleBookingRepository(db *sqlx.DB) *VehicleBookingRepository {
	return &VehicleBookingRepository{db: db}
}

func (r *VehicleBookingRepository) loadEmployeeBrief(id *string) *model.EmployeeBrief {
	if id == nil || *id == "" {
		return nil
	}
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, *id); err != nil {
		return nil
	}
	return &brief
}

func (r *VehicleBookingRepository) loadVehicle(id string) *model.Vehicle {
	var v model.Vehicle
	if err := r.db.Get(&v, `SELECT * FROM "Vehicle" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &v
}

func (r *VehicleBookingRepository) hydrate(b *model.VehicleBooking) {
	b.Vehicle = r.loadVehicle(b.VehicleID)
	b.RequestedBy = r.loadEmployeeBrief(&b.RequestedByID)
	b.ApprovedBy = r.loadEmployeeBrief(b.ApprovedByID)
}

func (r *VehicleBookingRepository) Get(id string) (*model.VehicleBooking, error) {
	var b model.VehicleBooking
	if err := r.db.Get(&b, `SELECT * FROM "VehicleBooking" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate(&b)
	return &b, nil
}

// FindOverlapping يرجع الحجوزات المتعارضة زمنياً على نفس السيارة بحالة PENDING أو APPROVED.
// overlap = startAt < existing.endAt AND endAt > existing.startAt
func (r *VehicleBookingRepository) FindOverlapping(vehicleID string, startAt, endAt string, excludeBookingID *string) ([]model.VehicleBooking, error) {
	bookings := []model.VehicleBooking{}
	query := `
		SELECT * FROM "VehicleBooking"
		WHERE "vehicleId" = $1
		  AND status IN ('PENDING','APPROVED')
		  AND "startAt" < $3::timestamp
		  AND "endAt" > $2::timestamp
	`
	args := []any{vehicleID, startAt, endAt}
	if excludeBookingID != nil && *excludeBookingID != "" {
		query += ` AND id != $4`
		args = append(args, *excludeBookingID)
	}
	if err := r.db.Select(&bookings, query, args...); err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *VehicleBookingRepository) Create(requestedByID, vehicleID, purpose, startAt, endAt string) (*model.VehicleBooking, error) {
	var b model.VehicleBooking
	err := r.db.Get(&b, `
		INSERT INTO "VehicleBooking" (id, "vehicleId", "requestedById", purpose, "startAt", "endAt")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4::timestamp, $5::timestamp)
		RETURNING *
	`, vehicleID, requestedByID, purpose, startAt, endAt)
	if err != nil {
		return nil, err
	}
	r.hydrate(&b)
	return &b, nil
}

func (r *VehicleBookingRepository) Decide(bookingID, approverID string, approve bool, rejectionReason *string) (*model.VehicleBooking, error) {
	status := "REJECTED"
	if approve {
		status = "APPROVED"
	}
	var b model.VehicleBooking
	err := r.db.Get(&b, `
		UPDATE "VehicleBooking" SET
			status = $2,
			"approvedById" = $3,
			"rejectionReason" = $4,
			"decidedAt" = now()
		WHERE id = $1
		RETURNING *
	`, bookingID, status, approverID, rejectionReason)
	if err != nil {
		return nil, err
	}
	r.hydrate(&b)
	return &b, nil
}

func (r *VehicleBookingRepository) Cancel(bookingID string) (*model.VehicleBooking, error) {
	var b model.VehicleBooking
	err := r.db.Get(&b, `
		UPDATE "VehicleBooking" SET status = 'CANCELLED'
		WHERE id = $1
		RETURNING *
	`, bookingID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&b)
	return &b, nil
}

// FindApprovedCoveringNow يرجع حجزاً معتمَداً يغطي الوقت الحالي لسيارة معيّنة (لتحذير بدء المهمة).
func (r *VehicleBookingRepository) FindApprovedCoveringNow(vehicleID string) (*model.VehicleBooking, error) {
	var b model.VehicleBooking
	err := r.db.Get(&b, `
		SELECT * FROM "VehicleBooking"
		WHERE "vehicleId" = $1 AND status = 'APPROVED'
		  AND "startAt" <= now() AND "endAt" >= now()
		LIMIT 1
	`, vehicleID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&b)
	return &b, nil
}

func (r *VehicleBookingRepository) List(filters model.VehicleBookingFilters) ([]model.VehicleBooking, error) {
	bookings := []model.VehicleBooking{}
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
	if filters.RequestedByID != nil && *filters.RequestedByID != "" {
		add(`"requestedById" = $%d`, *filters.RequestedByID)
	}
	if filters.Status != nil && *filters.Status != "" {
		add(`status = $%d`, *filters.Status)
	}
	if filters.From != nil && *filters.From != "" {
		add(`"startAt" >= $%d::timestamp`, *filters.From)
	}
	if filters.To != nil && *filters.To != "" {
		add(`"startAt" <= $%d::timestamp`, *filters.To)
	}
	query := `SELECT * FROM "VehicleBooking" WHERE ` + strings.Join(clauses, " AND ") + ` ORDER BY "startAt" DESC`
	if err := r.db.Select(&bookings, query, args...); err != nil {
		return nil, err
	}
	for i := range bookings {
		r.hydrate(&bookings[i])
	}
	return bookings, nil
}
