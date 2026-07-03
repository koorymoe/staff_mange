package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) List(status, customerID string) ([]model.Booking, error) {
	query := `SELECT * FROM "Booking" WHERE 1=1`
	args := []any{}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(` AND status = $%d`, len(args))
	}
	if customerID != "" {
		args = append(args, customerID)
		query += fmt.Sprintf(` AND "customerId" = $%d`, len(args))
	}
	query += ` ORDER BY "createdAt" DESC`

	bookings := []model.Booking{}
	if err := r.db.Select(&bookings, query, args...); err != nil {
		return nil, err
	}
	for i := range bookings {
		if err := r.hydrate(&bookings[i]); err != nil {
			return nil, err
		}
	}
	return bookings, nil
}

func (r *BookingRepository) FindByID(id string) (*model.Booking, error) {
	var b model.Booking
	err := r.db.Get(&b, `SELECT * FROM "Booking" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

// hydrate يجلب كل العلاقات المرتبطة بالحجز (زبون، خدمة، موظفين، تعيينات، مواد السلة، سجل التعديلات)
func (r *BookingRepository) hydrate(b *model.Booking) error {
	var customer model.Customer
	if err := r.db.Get(&customer, `SELECT * FROM "Customer" WHERE id = $1`, b.CustomerID); err == nil {
		b.Customer = &customer
	}

	if b.ServiceID != nil {
		var svc model.Service
		if err := r.db.Get(&svc, `SELECT * FROM "Service" WHERE id = $1`, *b.ServiceID); err == nil {
			b.Service = &svc
		}
	}

	loadEmployee := func(id *string) *model.Employee {
		if id == nil {
			return nil
		}
		var e model.Employee
		if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, *id); err != nil {
			return nil
		}
		return &e
	}
	b.TransferEmployee = loadEmployee(b.TransferEmployeeID)
	b.ProjectSupervisor = loadEmployee(b.ProjectSupervisorID)
	b.ConfirmedByEmployee = loadEmployee(b.ConfirmedByEmployeeID)
	b.ExpenseResponsible = loadEmployee(b.ExpenseResponsibleID)

	assignments := []model.BookingAssignment{}
	if err := r.db.Select(&assignments, `SELECT * FROM "BookingAssignment" WHERE "bookingId" = $1`, b.ID); err == nil {
		for i := range assignments {
			if e := loadEmployee(&assignments[i].EmployeeID); e != nil {
				assignments[i].Employee = *e
			}
		}
		b.Assignments = assignments
	}
	if b.Assignments == nil {
		b.Assignments = []model.BookingAssignment{}
	}

	cartItems := []model.CartItem{}
	if err := r.db.Select(&cartItems, `SELECT * FROM "CartItem" WHERE "bookingId" = $1 ORDER BY "createdAt" ASC`, b.ID); err == nil {
		b.CartItems = cartItems
	}
	if b.CartItems == nil {
		b.CartItems = []model.CartItem{}
	}

	logs := []model.ScheduleChangeLog{}
	if err := r.db.Select(&logs, `SELECT * FROM "ScheduleChangeLog" WHERE "bookingId" = $1 ORDER BY "createdAt" DESC`, b.ID); err == nil {
		for i := range logs {
			logs[i].ChangedBy = loadEmployee(&logs[i].ChangedByID)
		}
		b.ScheduleLogs = logs
	}
	if b.ScheduleLogs == nil {
		b.ScheduleLogs = []model.ScheduleChangeLog{}
	}

	return nil
}

func (r *BookingRepository) NextSequenceNumber() (int, error) {
	var seq sql.NullInt64
	err := r.db.Get(&seq, `SELECT MAX("sequenceNumber") FROM "Booking"`)
	if err != nil {
		return 1, err
	}
	if !seq.Valid {
		return 1, nil
	}
	return int(seq.Int64) + 1, nil
}

func (r *BookingRepository) Create(b *model.Booking) error {
	_, err := r.db.NamedExec(`
		INSERT INTO "Booking" (id, code, "sequenceNumber", "customerId", "serviceId", notes, "vehicleType", priority, "transferEmployeeId", "updatedAt")
		VALUES (:id, :code, :sequenceNumber, :customerId, :serviceId, :notes, :vehicleType, :priority, :transferEmployeeId, now())
	`, b)
	return err
}

func (r *BookingRepository) Confirm(id string, req model.ConfirmBookingRequest, scheduledAt *string) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET
			status = 'CONFIRMED',
			"confirmedByName" = COALESCE($2, "confirmedByName"),
			"confirmedByEmployeeId" = COALESCE($3, "confirmedByEmployeeId"),
			"adminNotes" = COALESCE($4, "adminNotes"),
			"transferToProjects" = $5,
			"quotedPrice" = COALESCE($6, "quotedPrice"),
			address = COALESCE($7, address),
			"scheduledAt" = COALESCE($8::timestamp, "scheduledAt")
		WHERE id = $1
	`, id, req.ConfirmedByName, req.ConfirmedByEmployeeID, req.AdminNotes, req.TransferToProjects, req.QuotedPrice, req.Address, scheduledAt)
	return err
}

func (r *BookingRepository) UpdateDetails(id string, req model.UpdateBookingDetailsRequest) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET
			"quotedPrice" = $2,
			address = COALESCE($3, address),
			"assignedVehicle" = COALESCE($4, "assignedVehicle"),
			"mapLocation" = COALESCE($5, "mapLocation"),
			"mapLatitude" = COALESCE($6, "mapLatitude"),
			"mapLongitude" = COALESCE($7, "mapLongitude"),
			"expenseResponsibleId" = COALESCE($8, "expenseResponsibleId")
		WHERE id = $1
	`, id, req.QuotedPrice, req.Address, req.AssignedVehicle, req.MapLocation, req.MapLatitude, req.MapLongitude, req.ExpenseResponsibleID)
	return err
}

func (r *BookingRepository) ScheduleLog(bookingID string) ([]model.ScheduleChangeLog, error) {
	logs := []model.ScheduleChangeLog{}
	err := r.db.Select(&logs, `SELECT * FROM "ScheduleChangeLog" WHERE "bookingId" = $1 ORDER BY "createdAt" DESC`, bookingID)
	if err != nil {
		return nil, err
	}
	for i := range logs {
		var e model.Employee
		if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, logs[i].ChangedByID); err == nil {
			logs[i].ChangedBy = &e
		}
	}
	return logs, nil
}

func (r *BookingRepository) SetStatus(id, status string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *BookingRepository) Complete(id string, req model.CompleteBookingRequest) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET
			status = 'COMPLETED',
			"completedAt" = now(),
			"completionNotes" = COALESCE($2, "completionNotes"),
			"amountCollected" = COALESCE($3, "amountCollected"),
			"advancePaid" = COALESCE($4, "advancePaid")
		WHERE id = $1
	`, id, req.CompletionNotes, req.AmountCollected, req.AdvancePaid)
	return err
}

func (r *BookingRepository) Verify(id string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET "amountVerified" = true WHERE id = $1`, id)
	return err
}

func (r *BookingRepository) SetSchedule(id, scheduledAt string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET "scheduledAt" = $2::timestamp WHERE id = $1`, id, scheduledAt)
	return err
}

func (r *BookingRepository) SetSupervisor(id string, employeeID *string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET "projectSupervisorId" = $2, "expenseResponsibleId" = $2 WHERE id = $1`, id, employeeID)
	return err
}

func (r *BookingRepository) SetAdminNotes(id, notes string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET "adminNotes" = $2 WHERE id = $1`, id, notes)
	return err
}

func (r *BookingRepository) SetExpenseResponsible(id, employeeID string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET "expenseResponsibleId" = $2 WHERE id = $1`, id, employeeID)
	return err
}

func (r *BookingRepository) SetAssignedVehicle(id, vehicle string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET "assignedVehicle" = $2 WHERE id = $1`, id, vehicle)
	return err
}

func (r *BookingRepository) CreateScheduleLog(bookingID, changedByID string, oldTime *string, newTime string) error {
	_, err := r.db.Exec(`
		INSERT INTO "ScheduleChangeLog" (id, "bookingId", "changedById", "oldTime", "newTime")
		VALUES (gen_random_uuid()::text, $1, $2, $3::timestamp, $4::timestamp)
	`, bookingID, changedByID, oldTime, newTime)
	return err
}

func (r *BookingRepository) UpsertAssignment(bookingID, employeeID, role string) error {
	_, err := r.db.Exec(`
		INSERT INTO "BookingAssignment" (id, "bookingId", "employeeId", role)
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		ON CONFLICT ("bookingId", role) DO UPDATE SET "employeeId" = EXCLUDED."employeeId"
	`, bookingID, employeeID, role)
	return err
}

func (r *BookingRepository) ListAssignments(bookingID string) ([]model.BookingAssignment, error) {
	assignments := []model.BookingAssignment{}
	err := r.db.Select(&assignments, `SELECT * FROM "BookingAssignment" WHERE "bookingId" = $1`, bookingID)
	return assignments, err
}

func (r *BookingRepository) EmployeeHasSkillForService(employeeID, serviceID string) (bool, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "EmployeeSkill" es
		JOIN "Skill" sk ON sk.id = es."skillId"
		WHERE es."employeeId" = $1 AND es."canPerform" = true AND sk."serviceId" = $2
	`, employeeID, serviceID)
	return count > 0, err
}
