package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

func (r *BookingRepository) List(status, customerID, date string) ([]model.Booking, error) {
	query := `SELECT * FROM "Booking" WHERE 1=1`
	args := []any{}
	if date != "" {
		// نفس منطق "موعد الحجز الفعلي" بالواجهة: scheduledAt لو موجود، وإلا createdAt —
		// حتى فلترة التاريخ تصير بالسيرفر (نجيب يوم وحد فقط) بدل ما نجيب كل أرشيف
		// الحجوزات التاريخي ونفلتره بالواجهة، وهذا كان يبطّئ الصفحة مع تراكم البيانات.
		args = append(args, date)
		query += fmt.Sprintf(` AND (
			(("scheduledAt" IS NOT NULL) AND "scheduledAt"::date = $%d::date)
			OR (("scheduledAt" IS NULL) AND "createdAt"::date = $%d::date)
		)`, len(args), len(args))
	}
	if status != "" {
		// يدعم قائمة حالات مفصولة بفاصلة (مثلاً "PENDING,CONFIRMED,IN_PROGRESS") حتى
		// يقدر المنسق يجيب الحجوزات الفعالة بس بدل كل الأرشيف التاريخي (آلاف الحجوزات
		// القديمة المكتملة/الملغاة) بطلب واحد بدل ما يجيب كل شي ويفلتر بالواجهة.
		statuses := strings.Split(status, ",")
		if len(statuses) > 1 {
			args = append(args, pq.Array(statuses))
			query += fmt.Sprintf(` AND status = ANY($%d)`, len(args))
		} else {
			args = append(args, status)
			query += fmt.Sprintf(` AND status = $%d`, len(args))
		}
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
	if err := r.hydrateAll(toPointers(bookings)); err != nil {
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
	if err := r.hydrateAll(toPointers(bookings)); err != nil {
		return nil, err
	}
	return bookings, nil
}

// toPointers تحول []model.Booking إلى []*model.Booking تشاور نفس عناصر المصفوفة
// الأصلية — لازم نمرر مؤشرات لـ hydrateAll حتى التعديلات (Customer, Service...)
// توصل فعلاً للسلايس الي يرجعه الكولر، مو لنسخة مؤقتة تنرمى بعد ما تخلص الدالة.
func toPointers(bookings []model.Booking) []*model.Booking {
	ptrs := make([]*model.Booking, len(bookings))
	for i := range bookings {
		ptrs[i] = &bookings[i]
	}
	return ptrs
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
	if err := r.hydrateAll([]*model.Booking{&b}); err != nil {
		return nil, err
	}
	return &b, nil
}

// hydrate يجلب علاقات حجز واحد — يلف hydrateAll تفادياً لتكرار المنطق.
func (r *BookingRepository) hydrate(b *model.Booking) error {
	return r.hydrateAll([]*model.Booking{b})
}

// hydrateAll يجلب كل العلاقات المرتبطة بمجموعة حجوزات دفعة وحدة (batch) بدل استعلام
// منفصل لكل حجز — قبل هذا التعديل كل حجز كان يسوي 8-15+ استعلام لحاله (N+1)، فلما
// صار عدد الحجوزات بالآلاف بعد استيراد البيانات القديمة، صفحة الحجوزات صارت تسوي
// عشرات الآلاف من الاستعلامات المتسلسلة وتعلق. الحل: نجمع كل الـ IDs المطلوبة أول
// وبعدين نجيبهم بدفعة وحدة لكل نوع (WHERE id = ANY(...))، ونوزعهم بالذاكرة.
func (r *BookingRepository) hydrateAll(bookings []*model.Booking) error {
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
		addEmp(b.ConfirmationContactedByID)
		addEmp(b.LastEditedByID)
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

	// كل الخدمات المرتبطة بكل حجز (خدمات متعددة بالحجز الواحد) — استعلام
	// واحد لكل الحجوزات بدل استعلام لكل حجز.
	extraServicesByBooking := map[string][]model.Service{}
	if len(bookingIDs) > 0 {
		type bsRow struct {
			BookingID     string `db:"bookingId"`
			model.Service `db:",inline"`
		}
		rows := []bsRow{}
		if err := r.db.Select(&rows, `
			SELECT bs."bookingId", s.*
			FROM "BookingService" bs
			JOIN "Service" s ON s.id = bs."serviceId"
			WHERE bs."bookingId" = ANY($1)
			ORDER BY bs."createdAt"`, pq.Array(bookingIDs)); err == nil {
			for _, row := range rows {
				extraServicesByBooking[row.BookingID] = append(extraServicesByBooking[row.BookingID], row.Service)
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
		b := bookings[i]
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
		// قائمة الخدمات: لو ما اكو صفوف بالجدول الجديد (حجز قديم) ننزل على
		// الخدمة المفردة حتى الواجهة تلاقي دائماً قائمة مو فاضية
		if list := extraServicesByBooking[b.ID]; len(list) > 0 {
			// الخدمة الرئيسية دائماً أول القائمة حتى العرض يكون ثابت ومفهوم
			if b.ServiceID != nil {
				for i, sv := range list {
					if sv.ID == *b.ServiceID && i != 0 {
						list[0], list[i] = list[i], list[0]
						break
					}
				}
			}
			b.Services = list
		} else if b.Service != nil {
			b.Services = []model.Service{*b.Service}
		}
		b.TransferEmployee = getEmp(b.TransferEmployeeID)
		b.ProjectSupervisor = getEmp(b.ProjectSupervisorID)
		b.ConfirmedByEmployee = getEmp(b.ConfirmedByEmployeeID)
		b.ExpenseResponsible = getEmp(b.ExpenseResponsibleID)
		b.MaterialsReadyBy = getEmpBrief(b.MaterialsReadyByID)
		b.ConfirmationContactedBy = getEmpBrief(b.ConfirmationContactedByID)
		b.LastEditedBy = getEmpBrief(b.LastEditedByID)

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
		INSERT INTO "Booking" (id, code, "sequenceNumber", "customerId", "serviceId", notes, "vehicleType", priority, "transferEmployeeId", address, "mapLatitude", "mapLongitude", "locationUrl", "updatedAt")
		VALUES (:id, :code, :sequenceNumber, :customerId, :serviceId, :notes, :vehicleType, :priority, :transferEmployeeId, :address, :mapLatitude, :mapLongitude, :locationUrl, now())
	`, b)
	return err
}

// SetServices يحدّث قائمة خدمات الحجز (يستبدلها بالكامل). أول خدمة بالقائمة
// تصير الخدمة الرئيسية بعمود serviceId حتى الشاشات القديمة تضل تشتغل.
func (r *BookingRepository) SetServices(bookingID string, serviceIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM "BookingService" WHERE "bookingId" = $1`, bookingID); err != nil {
		return err
	}
	seen := map[string]bool{}
	var primary *string
	for _, sid := range serviceIDs {
		if sid == "" || seen[sid] {
			continue
		}
		seen[sid] = true
		if primary == nil {
			v := sid
			primary = &v
		}
		if _, err := tx.Exec(`
			INSERT INTO "BookingService" ("bookingId", "serviceId") VALUES ($1, $2)
			ON CONFLICT DO NOTHING`, bookingID, sid); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`UPDATE "Booking" SET "serviceId" = $2, "updatedAt" = now() WHERE id = $1`,
		bookingID, primary); err != nil {
		return err
	}
	return tx.Commit()
}

// ServiceIDsFor يرجّع معرّفات خدمات الحجز بالترتيب.
func (r *BookingRepository) ServiceIDsFor(bookingID string) ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `
		SELECT "serviceId" FROM "BookingService" WHERE "bookingId" = $1 ORDER BY "createdAt"`, bookingID)
	return ids, err
}

func (r *BookingRepository) Confirm(id string, req model.ConfirmBookingRequest, scheduledAt *string) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET
			status = 'CONFIRMED',
			"confirmedByName" = COALESCE($2, "confirmedByName"),
			"confirmedByEmployeeId" = COALESCE($3, "confirmedByEmployeeId"),
			"adminNotes" = COALESCE($4, "adminNotes"),
			"transferToProjects" = $5,
			-- وقت التحويل لتنسيق الحجوزات — أول مرة بس، ما ينداس بإعادة التثبيت
			"confirmedAt" = COALESCE("confirmedAt", now()),
			"quotedPrice" = COALESCE($6, "quotedPrice"),
			address = COALESCE($7, address),
			"scheduledAt" = COALESCE($8::timestamp, "scheduledAt")
		WHERE id = $1
	`, id, req.ConfirmedByName, req.ConfirmedByEmployeeID, req.AdminNotes, req.TransferToProjects, req.QuotedPrice, req.Address, scheduledAt)
	return err
}

// MarkConfirmationContacted يسجّل لحظة "تم" الإداري بعد ما اتصل بالزبون وأكّد
// معه الموعد/الاتفاق — قبل الضغط الفعلي على "تثبيت" (Confirm). هذا يفصل خطوة
// "تواصلت مع الزبون وأقفلت الاتفاق" عن تغيير حالة الحجز نفسها، حتى يقدر المراقب
// (بصلاحية crew_management) يدقق هل الإداري فعلاً تواصل قبل ما يثبّت الحجز.
func (r *BookingRepository) MarkConfirmationContacted(id, byEmployeeID string) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET "confirmationContactedAt" = now(), "confirmationContactedById" = $2
		WHERE id = $1
	`, id, byEmployeeID)
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
			"expenseResponsibleId" = COALESCE($8, "expenseResponsibleId"),
			"locationUrl" = COALESCE($9, "locationUrl")
		WHERE id = $1
	`, id, req.QuotedPrice, req.Address, req.AssignedVehicle, req.MapLocation, req.MapLatitude, req.MapLongitude, req.ExpenseResponsibleID, req.LocationUrl)
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

// TouchLastEdited يسجّل مين آخر موظف عدّل تفاصيل/تكليف الحجز — منفصل عن "من أكّده".
func (r *BookingRepository) TouchLastEdited(id, editorID string) error {
	if editorID == "" {
		return nil
	}
	_, err := r.db.Exec(`UPDATE "Booking" SET "lastEditedById" = $2, "lastEditedAt" = now() WHERE id = $1`, id, editorID)
	return err
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
			"startedAt" = COALESCE("startedAt", now()),
			"responseMinutes" = CASE
				WHEN "materialsReadyAt" IS NOT NULL AND "responseMinutes" IS NULL
				THEN GREATEST(0, EXTRACT(EPOCH FROM (now() - "materialsReadyAt"))::int / 60)
				ELSE "responseMinutes"
			END
		WHERE id = $1
	`, id)
	return err
}

// MarkArrived يسجّل لحظة وصول الفنيين لموقع الزبون (قبل بدء العمل فعلياً) — يُستخدم
// لاحقاً كبديل عن startedAt عند حساب عيّنات مدة العمل لو startedAt غير متوفر.
func (r *BookingRepository) MarkArrived(id string) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET "arrivedAt" = COALESCE("arrivedAt", now()) WHERE id = $1
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

// CountCompletedForEmployeeMonth يرجّع عدد الحجوزات المكتملة (COMPLETED) خلال
// شهر معيّن (monthPrefix بصيغة "YYYY-MM") التي كان موظف معيّن مربوطاً بها عبر
// "BookingAssignment" (أي دور: ليدر أو فني).
func (r *BookingRepository) CountCompletedForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND b.status = 'COMPLETED' AND b."completedAt" IS NOT NULL
			AND to_char(b."completedAt", 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return count, err
}

// CountAssignedForEmployeeMonth يرجّع عدد كل الحجوزات المسندة لموظف معيّن خلال
// شهر معيّن، بكل الحالات (مثبت/ملغى/منجز...) — يُحسب حسب تاريخ الإنجاز أو
// الإلغاء الفعلي لو موجود، وإلا تاريخ الإسناد (createdAt) — مو الموعد
// المجدول، حتى تأجيل الموعد ما يغيّر الشهر الي تنحسب فيه.
func (r *BookingRepository) CountAssignedForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1
			AND to_char(COALESCE(b."completedAt", b."createdAt"), 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return count, err
}

// CountMaintenanceForEmployeeMonth يرجّع عدد حجوزات الصيانة (bookingType =
// 'MAINTENANCE') المسندة لموظف معيّن خلال شهر معيّن.
func (r *BookingRepository) CountMaintenanceForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND b."bookingType" = 'MAINTENANCE'
			AND to_char(COALESCE(b."completedAt", b."createdAt"), 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return count, err
}

// CountFreeMaintenanceForEmployeeMonth يرجّع عدد حجوزات الصيانة المجانية (بدون
// تكلفة مقدّرة) المسندة لموظف معيّن خلال شهر معيّن.
func (r *BookingRepository) CountFreeMaintenanceForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND b."bookingType" = 'MAINTENANCE'
			AND (b."quotedPrice" IS NULL OR b."quotedPrice" = 0)
			AND to_char(COALESCE(b."completedAt", b."createdAt"), 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return count, err
}

// CountCompletedForEmployeeRange نفس CountCompletedForEmployeeMonth لكن لمدى
// تاريخ حر (from/to بصيغة "YYYY-MM-DD").
func (r *BookingRepository) CountCompletedForEmployeeRange(employeeID, from, to string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND b.status = 'COMPLETED' AND b."completedAt" IS NOT NULL
			AND b."completedAt"::date BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return count, err
}

// CountAssignedForEmployeeRange نفس CountAssignedForEmployeeMonth لكن لمدى
// تاريخ حر.
func (r *BookingRepository) CountAssignedForEmployeeRange(employeeID, from, to string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1
			AND COALESCE(b."completedAt", b."createdAt")::date BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return count, err
}

// CountMaintenanceForEmployeeRange نفس CountMaintenanceForEmployeeMonth لكن
// لمدى تاريخ حر.
func (r *BookingRepository) CountMaintenanceForEmployeeRange(employeeID, from, to string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND b."bookingType" = 'MAINTENANCE'
			AND COALESCE(b."completedAt", b."createdAt")::date BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return count, err
}

// CountFreeMaintenanceForEmployeeRange نفس CountFreeMaintenanceForEmployeeMonth
// لكن لمدى تاريخ حر.
func (r *BookingRepository) CountFreeMaintenanceForEmployeeRange(employeeID, from, to string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND b."bookingType" = 'MAINTENANCE'
			AND (b."quotedPrice" IS NULL OR b."quotedPrice" = 0)
			AND COALESCE(b."completedAt", b."createdAt")::date BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return count, err
}

// dailyDateExpr هو نفس منطق "relevantDate" المعتمد بصفحة الحجوزات: الموعد
// المحدد (scheduledAt) وإلا تاريخ التسجيل (createdAt) — حتى الإحصائية اليومية
// تلتقط الحجز باليوم الصحيح حتى لو تأجّل موعده لاحقاً.
const dailyDateExpr = `COALESCE(b."scheduledAt", b."createdAt")`

// CountForDate يرجّع عدد كل الحجوزات (بكل الحالات) بتاريخ معيّن.
func (r *BookingRepository) CountForDate(date string) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "Booking" b WHERE `+dailyDateExpr+`::date = $1::date`, date)
	return count, err
}

// CountMorningEveningForDate يرجّع عدد الحجوزات الصباحية (قبل الساعة 12
// ظهراً) والمسائية بتاريخ معيّن.
func (r *BookingRepository) CountMorningEveningForDate(date string) (morning int, evening int, err error) {
	err = r.db.Get(&morning, `SELECT COUNT(*) FROM "Booking" b WHERE `+dailyDateExpr+`::date = $1::date AND EXTRACT(HOUR FROM `+dailyDateExpr+`) < 12`, date)
	if err != nil {
		return 0, 0, err
	}
	err = r.db.Get(&evening, `SELECT COUNT(*) FROM "Booking" b WHERE `+dailyDateExpr+`::date = $1::date AND EXTRACT(HOUR FROM `+dailyDateExpr+`) >= 12`, date)
	return morning, evening, err
}

// CountDistinctCrewForDate يرجّع عدد الموظفين المميزين المكلّفين بأي حجز
// بتاريخ معيّن (الكادر الي طلع للحجوزات).
func (r *BookingRepository) CountDistinctCrewForDate(date string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT ba."employeeId") FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE `+dailyDateExpr+`::date = $1::date
	`, date)
	return count, err
}

// CountDistinctVehiclesForDate يرجّع عدد السيارات المميزة المستخدمة بحجوزات
// تاريخ معيّن.
func (r *BookingRepository) CountDistinctVehiclesForDate(date string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b."assignedVehicle") FROM "Booking" b
		WHERE `+dailyDateExpr+`::date = $1::date AND b."assignedVehicle" IS NOT NULL AND b."assignedVehicle" != ''
	`, date)
	return count, err
}

// AssignedAndCompletedForEmployeeOnDate يرجّع عدد الحجوزات المكلّف بيها موظف
// معيّن بتاريخ معيّن، وشكد منهن أنجزهن فعلاً.
func (r *BookingRepository) AssignedAndCompletedForEmployeeOnDate(employeeID, date string) (assigned int, completed int, err error) {
	err = r.db.Get(&assigned, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND `+dailyDateExpr+`::date = $2::date
	`, employeeID, date)
	if err != nil {
		return 0, 0, err
	}
	err = r.db.Get(&completed, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND `+dailyDateExpr+`::date = $2::date AND b.status = 'COMPLETED'
	`, employeeID, date)
	return assigned, completed, err
}

// CountEnteredThisWeekForEmployee يرجّع عدد الحجوزات الي أدخلها موظف مبيعات
// بالنظام (createdAt) خلال آخر 7 أيام — يُستخدم لقياس إنتاجية موظف المبيعات.
func (r *BookingRepository) CountEnteredThisWeekForEmployee(transferEmployeeID string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "Booking"
		WHERE "transferEmployeeId" = $1 AND "createdAt" >= now() - interval '7 days'
	`, transferEmployeeID)
	return count, err
}

// CountCompletedForEmployeeLast7Days يرجّع عدد الحجوزات المكتملة لموظف معيّن
// خلال آخر 7 أيام — يُستخدم لقياس إنتاجية الكوادر الأسبوعية.
func (r *BookingRepository) CountCompletedForEmployeeLast7Days(employeeID string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT b.id) FROM "Booking" b
		JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND b.status = 'COMPLETED' AND b."completedAt" >= now() - interval '7 days'
	`, employeeID)
	return count, err
}

// CrewMatesForEmployee يرجّع كل الموظفين المميزين الي طلعوا مع موظف معيّن
// بنفس الحجز (عبر BookingAssignment) — يُستخدم لتقييد "تقييم فريقي" بالليدر
// على زملاء حجوزاته الفعليين بس، مو كل الموظفين بالنظام.
func (r *BookingRepository) CrewMatesForEmployee(employeeID string) ([]model.EmployeeBrief, error) {
	mates := []model.EmployeeBrief{}
	err := r.db.Select(&mates, `
		SELECT DISTINCT e.id, e.name FROM "BookingAssignment" ba
		JOIN "BookingAssignment" mine ON mine."bookingId" = ba."bookingId" AND mine."employeeId" = $1
		JOIN "Employee" e ON e.id = ba."employeeId"
		WHERE ba."employeeId" != $1
	`, employeeID)
	return mates, err
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

// ReturnToCrew يرجّع حجز محوّل لإدارة المشاريع رجعة لكادر الشد — يستخدمه
// مدير المشاريع لما يفتح تفاصيل الحجز ويلكاه مو مال مشروع أصلاً. الحجز يبقى
// مثبّتاً (CONFIRMED)، بس ينشال من قائمة "بانتظار الاستلام كمشروع".
func (r *BookingRepository) ReturnToCrew(id string, note *string) error {
	_, err := r.db.Exec(`
		UPDATE "Booking" SET
			"transferToProjects" = false,
			"adminNotes" = CASE
				WHEN $2::text IS NULL OR $2::text = '' THEN "adminNotes"
				ELSE COALESCE("adminNotes" || E'\n', '') || 'أُعيد لكادر الشد: ' || $2::text
			END
		WHERE id = $1 AND "transferToProjects" = true
	`, id, note)
	return err
}

// IsAssignedTo يفحص إذا الموظف مكلّف فعلاً بهذا الحجز — أساس التحقق قبل أي
// إجراء على مسار العمل (وصلت/بدأت/أنهيت). بدونه أي موظف مسجّل دخول يقدر
// "ينهي" حجز موظف ثاني أو يغيّر موعده (ثغرة IDOR).
//
// نعتبره مكلّفاً كذلك لو هو مسؤول المصاريف أو مشرف المشروع أو الي رحّل الحجز
// — هذول أطراف شرعية بنفس الحجز.
func (r *BookingRepository) IsAssignedTo(bookingID, employeeID string) (bool, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "Booking" b
		WHERE b.id = $1 AND (
			b."expenseResponsibleId" = $2
			OR b."projectSupervisorId" = $2
			OR b."transferEmployeeId" = $2
			OR b."inspectionSupervisorId" = $2
			OR EXISTS (
				SELECT 1 FROM "BookingAssignment" ba
				WHERE ba."bookingId" = b.id AND ba."employeeId" = $2
			)
		)`, bookingID, employeeID)
	return n > 0, err
}

// IsCartItemOfAssignedBooking يفحص إذا عنصر السلة يتبع حجز الموظف طرف بيه —
// بدونه أي موظف يعدّل أو يحذف عناصر سلة أي حجز بمجرد معرفة رقم العنصر.
func (r *BookingRepository) IsCartItemOfAssignedBooking(cartItemID, employeeID string) (bool, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "CartItem" ci
		JOIN "Booking" b ON b.id = ci."bookingId"
		WHERE ci.id = $1 AND (
			b."expenseResponsibleId" = $2
			OR b."projectSupervisorId" = $2
			OR b."transferEmployeeId" = $2
			OR b."inspectionSupervisorId" = $2
			OR EXISTS (
				SELECT 1 FROM "BookingAssignment" ba
				WHERE ba."bookingId" = b.id AND ba."employeeId" = $2
			)
		)`, cartItemID, employeeID)
	return n > 0, err
}
