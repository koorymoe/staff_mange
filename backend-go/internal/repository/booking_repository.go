package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

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
	if err := r.hydrateAll(bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}

// ListForAssignedEmployee يرجّع الحجوزات اللي الموظف معيّن عليها بـ BookingAssignment
// (مثلاً موظف مبيعات أو فني مرتبط بيها) — يستخدمها المساعد الذكي لعرض "حجوزاتي".
func (r *BookingRepository) ListForAssignedEmployee(employeeID string, limit int) ([]model.Booking, error) {
	bookings := []model.Booking{}
	err := r.db.Select(&bookings, `
		SELECT b.* FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1
		ORDER BY b."createdAt" DESC
		LIMIT $2
	`, employeeID, limit)
	if err != nil {
		return nil, err
	}
	if err := r.hydrateAll(bookings); err != nil {
		return nil, err
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
	if err := r.hydrateAll([]model.Booking{b}); err != nil {
		return nil, err
	}
	return &b, nil
}

// hydrate يجلب علاقات حجز واحد — يلف hydrateAll تفادياً لتكرار المنطق.
func (r *BookingRepository) hydrate(b *model.Booking) error {
	return r.hydrateAll([]model.Booking{*b})
}

// hydrateAll يجلب كل العلاقات المرتبطة بمجموعة حجوزات دفعة وحدة (batch) بدل استعلام
// منفصل لكل حجز — قبل هذا التعديل كل حجز كان يسوي 8-15+ استعلام لحاله (N+1)، فلما
// صار عدد الحجوزات بالآلاف بعد استيراد البيانات القديمة، صفحة الحجوزات صارت تسوي
// عشرات الآلاف من الاستعلامات المتسلسلة وتعلق. الحل: نجمع كل الـ IDs المطلوبة أول
// وبعدين نجيبهم بدفعة وحدة لكل نوع (WHERE id = ANY(...))، ونوزعهم بالذاكرة.
func (r *BookingRepository) hydrateAll(bookings []model.Booking) error {
	if len(bookings) == 0 {
		return nil
	}

	customerIDs := make([]string, 0, len(bookings))
	serviceIDs := make([]string, 0)
	bookingIDs := make([]string, 0, len(bookings))
	empIDSet := map[string]bool{}
	addEmp := func(id *string) {
		if id != nil && *id != "" {
			empIDSet[*id] = true
		}
	}
	for _, b := range bookings {
		customerIDs = append(customerIDs, b.CustomerID)
		if b.ServiceID != nil {
			serviceIDs = append(serviceIDs, *b.ServiceID)
		}
		bookingIDs = append(bookingIDs, b.ID)
		addEmp(b.TransferEmployeeID)
		addEmp(b.ProjectSupervisorID)
		addEmp(b.ConfirmedByEmployeeID)
		addEmp(b.ExpenseResponsibleID)
		addEmp(b.MaterialsReadyByID)
	}

	customers := map[string]model.Customer{}
	if len(customerIDs) > 0 {
		rows := []model.Customer{}
		if err := r.db.Select(&rows, `SELECT * FROM "Customer" WHERE id = ANY($1)`, pq.Array(customerIDs)); err == nil {
			for _, c := range rows {
				customers[c.ID] = c
			}
		}
	}

	services := map[string]model.Service{}
	if len(serviceIDs) > 0 {
		rows := []model.Service{}
		if err := r.db.Select(&rows, `SELECT * FROM "Service" WHERE id = ANY($1)`, pq.Array(serviceIDs)); err == nil {
			for _, s := range rows {
				services[s.ID] = s
			}
		}
	}

	assignmentsByBooking := map[string][]model.BookingAssignment{}
	if len(bookingIDs) > 0 {
		rows := []model.BookingAssignment{}
		if err := r.db.Select(&rows, `SELECT * FROM "BookingAssignment" WHERE "bookingId" = ANY($1)`, pq.Array(bookingIDs)); err == nil {
			for _, a := range rows {
				addEmp(&a.EmployeeID)
				assignmentsByBooking[a.BookingID] = append(assignmentsByBooking[a.BookingID], a)
			}
		}
	}

	cartItemsByBooking := map[string][]model.CartItem{}
	if len(bookingIDs) > 0 {
		rows := []model.CartItem{}
		if err := r.db.Select(&rows, `SELECT * FROM "CartItem" WHERE "bookingId" = ANY($1) ORDER BY "createdAt" ASC`, pq.Array(bookingIDs)); err == nil {
			for _, c := range rows {
				cartItemsByBooking[c.BookingID] = append(cartItemsByBooking[c.BookingID], c)
			}
		}
	}

	logsByBooking := map[string][]model.ScheduleChangeLog{}
	if len(bookingIDs) > 0 {
		rows := []model.ScheduleChangeLog{}
		if err := r.db.Select(&rows, `SELECT * FROM "ScheduleChangeLog" WHERE "bookingId" = ANY($1) ORDER BY "createdAt" DESC`, pq.Array(bookingIDs)); err == nil {
			for _, l := range rows {
				addEmp(&l.ChangedByID)
				logsByBooking[l.BookingID] = append(logsByBooking[l.BookingID], l)
			}
		}
	}

	employees := map[string]model.Employee{}
	if len(empIDSet) > 0 {
		empIDs := make([]string, 0, len(empIDSet))
		for id := range empIDSet {
			empIDs = append(empIDs, id)
		}
		rows := []model.Employee{}
		if err := r.db.Select(&rows, `SELECT * FROM "Employee" WHERE id = ANY($1)`, pq.Array(empIDs)); err == nil {
			for _, e := range rows {
				employees[e.ID] = e
			}
		}
	}
	getEmp := func(id *string) *model.Employee {
		if id == nil {
			return nil
		}
		if e, ok := employees[*id]; ok {
			return &e
		}
		return nil
	}
	getEmpBrief := func(id *string) *model.EmployeeBrief {
		e := getEmp(id)
		if e == nil {
			return nil
		}
		return &model.EmployeeBrief{ID: e.ID, Name: e.Name}
	}

	for i := range bookings {
		b := &bookings[i]
		if c, ok := customers[b.CustomerID]; ok {
			cc := c
			b.Customer = &cc
		}
		if b.ServiceID != nil {
			if s, ok := services[*b.ServiceID]; ok {
				ss := s
				b.Service = &ss
			}
		}
		b.TransferEmployee = getEmp(b.TransferEmployeeID)
		b.ProjectSupervisor = getEmp(b.ProjectSupervisorID)
		b.ConfirmedByEmployee = getEmp(b.ConfirmedByEmployeeID)
		b.ExpenseResponsible = getEmp(b.ExpenseResponsibleID)
		b.MaterialsReadyBy = getEmpBrief(b.MaterialsReadyByID)

		assignments := assignmentsByBooking[b.ID]
		for j := range assignments {
			if e := getEmp(&assignments[j].EmployeeID); e != nil {
				assignments[j].Employee = *e
			}
		}
		if assignments == nil {
			assignments = []model.BookingAssignment{}
		}
		b.Assignments = assignments

		cartItems := cartItemsByBooking[b.ID]
		if cartItems == nil {
			cartItems = []model.CartItem{}
		}
		b.CartItems = cartItems

		logs := logsByBooking[b.ID]
		for j := range logs {
			logs[j].ChangedBy = getEmp(&logs[j].ChangedByID)
		}
		if logs == nil {
			logs = []model.ScheduleChangeLog{}
		}
		b.ScheduleLogs = logs
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
		INSERT INTO "Booking" (id, code, "sequenceNumber", "customerId", "serviceId", notes, "vehicleType", priority, "transferEmployeeId", address, "mapLatitude", "mapLongitude", "updatedAt")
		VALUES (:id, :code, :sequenceNumber, :customerId, :serviceId, :notes, :vehicleType, :priority, :transferEmployeeId, :address, :mapLatitude, :mapLongitude, now())
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

// StartWithResponseTime يبدأ العمل ويحسب كم دقيقة أخذ الفنيون بعد ما تيم ليدر جهّز
// المواد ولحد ما فعلاً بدأوا الشغل — حتى نعرف مين ضيّع وقت.
func (r *BookingRepository) StartWithResponseTime(id string) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET
			status = 'IN_PROGRESS',
			"responseMinutes" = CASE
				WHEN "materialsReadyAt" IS NOT NULL AND "responseMinutes" IS NULL
				THEN GREATEST(0, EXTRACT(EPOCH FROM (now() - "materialsReadyAt"))::int / 60)
				ELSE "responseMinutes"
			END
		WHERE id = $1
	`, id)
	return err
}

// SetMaterialsReady يسجّل لحظة تجهيز المواد من تيم ليدر الفريق، ويبدأ عدّاد الاستجابة
func (r *BookingRepository) SetMaterialsReady(id, byEmployeeID string) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET "materialsReadyAt" = now(), "materialsReadyById" = $2
		WHERE id = $1
	`, id, byEmployeeID)
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
