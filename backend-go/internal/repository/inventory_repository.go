package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type InventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name, position FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

// ── Inventory Checks (جرد يومي) ──────────────────────────────────────────────

func (r *InventoryRepository) CreateInventoryCheck(employeeID string, req model.CreateInventoryCheckRequest) (*model.InventoryCheck, error) {
	var c model.InventoryCheck
	// ⚠️ ON CONFLICT يحدّث بدل ما يفشل: الفني يجرد، ويلگه أداة ناقصة
	// بعدين، ويعيد الجرد. رفضه برسالة «جردت قبل» يخلي السجل يكذب.
	// (الفهرس جزئي، فالجرد العام بلا حجز ما يمرّ من هنا ويظل يتراكم
	// بصفوف مستقلة مثل ما كان.)
	err := r.db.Get(&c, `
		INSERT INTO "InventoryCheck" (id, "employeeId", complete, "missingItems", "bookingId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, NULLIF($4,''))
		ON CONFLICT ("bookingId", "employeeId") WHERE "bookingId" IS NOT NULL
		DO UPDATE SET complete = EXCLUDED.complete,
		              "missingItems" = EXCLUDED."missingItems",
		              "checkedAt" = CURRENT_TIMESTAMP,
		              -- الجرد الجديد يلغي «انحلّت»: نقص جديد يحتاج حل جديد
		              resolved = false, "resolvedById" = NULL, "resolvedAt" = NULL
		RETURNING *
	`, employeeID, req.Complete, req.MissingItems, derefStr(req.BookingID))
	if err != nil {
		return nil, err
	}
	c.Employee = r.loadEmployeeBrief(c.EmployeeID)
	return &c, nil
}

// LastInventoryCheck يرجّع آخر جرد سوّاه الموظف نفسه (أو nil لو ما جرد
// أبداً) — تنستعمل حتى نعرف هل حان وقت جرده الأسبوعي. الفني يشوف
// جرده هو بس، مو جرد بقية الفنيين.
func (r *InventoryRepository) LastInventoryCheck(employeeID string) (*model.InventoryCheck, error) {
	var c model.InventoryCheck
	err := r.db.Get(&c, `
		SELECT * FROM "InventoryCheck"
		WHERE "employeeId" = $1
		ORDER BY "checkedAt" DESC
		LIMIT 1`, employeeID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// TodaysInventoryChecks يرجع آخر سجل جرد لكل موظف سجّل اليوم (للإداري يشوف مين ناقصه شي)
func (r *InventoryRepository) TodaysInventoryChecks() ([]model.InventoryCheck, error) {
	checks := []model.InventoryCheck{}
	if err := r.db.Select(&checks, `
		SELECT DISTINCT ON ("employeeId") *
		FROM "InventoryCheck"
		WHERE "checkedAt" >= baghdad_today()
		ORDER BY "employeeId", "checkedAt" DESC
	`); err != nil {
		return nil, err
	}
	for i := range checks {
		checks[i].Employee = r.loadEmployeeBrief(checks[i].EmployeeID)
		if checks[i].ResolvedByID != nil {
			checks[i].ResolvedBy = r.loadEmployeeBrief(*checks[i].ResolvedByID)
		}
	}
	return checks, nil
}

// ═══ جرد كادر حجز واحد ═══
//
// «الليدر يجرد أدواته ويشوف منو من الموظفين الي راح يطلعون وياه
// **بهذا الحجز** جرد».
//
// ⚠️ الكادر يجي من مصدرين لازم يتوحّدون: جدول التعيينات (الفنيين)
// **وعمود المشرف بالحجز** (الليدر). الليدر ما ينحفظ بجدول التعيينات
// — فلو أخذنا التعيينات بس، الليدر ما يطلع بقائمة فريقه أبداً وما
// يشوف جرده هو.
func (r *InventoryRepository) BookingCrewInventory(bookingID string) ([]model.BookingCrewInventoryState, error) {
	rows := []model.BookingCrewInventoryState{}
	err := r.db.Select(&rows, `
		WITH crew AS (
			SELECT ba."employeeId" AS id, false AS "isLeader"
			FROM "BookingAssignment" ba WHERE ba."bookingId" = $1
			UNION
			SELECT b."projectSupervisorId", true
			FROM "Booking" b
			WHERE b.id = $1 AND b."projectSupervisorId" IS NOT NULL
		)
		SELECT e.id AS "employeeId", e.name, e.position,
		       bool_or(crew."isLeader") AS "isLeader",
		       ic."checkedAt", ic.complete, ic."missingItems"
		FROM crew
		JOIN "Employee" e ON e.id = crew.id
		LEFT JOIN "InventoryCheck" ic
		       ON ic."bookingId" = $1 AND ic."employeeId" = e.id
		GROUP BY e.id, e.name, e.position, ic."checkedAt", ic.complete, ic."missingItems"
		ORDER BY bool_or(crew."isLeader") DESC, e.name
	`, bookingID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ResolveInventoryCheck يؤشر إنو الإداري/الأدمن وفّر النقص المسجّل بهذا الجرد.
func (r *InventoryRepository) ResolveInventoryCheck(id string, resolvedByID string) (*model.InventoryCheck, error) {
	var c model.InventoryCheck
	err := r.db.Get(&c, `
		UPDATE "InventoryCheck"
		SET resolved = true, "resolvedById" = $2, "resolvedAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING *
	`, id, resolvedByID)
	if err != nil {
		return nil, err
	}
	c.Employee = r.loadEmployeeBrief(c.EmployeeID)
	c.ResolvedBy = r.loadEmployeeBrief(resolvedByID)
	return &c, nil
}

// ── Personal Tools ──────────────────────────────────────────────────────────

func (r *InventoryRepository) ListPersonalTools(employeeID string) ([]model.PersonalTool, error) {
	tools := []model.PersonalTool{}
	var err error
	if employeeID != "" {
		err = r.db.Select(&tools, `SELECT * FROM "PersonalTool" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID)
	} else {
		err = r.db.Select(&tools, `SELECT * FROM "PersonalTool" ORDER BY "createdAt" DESC`)
	}
	if err != nil {
		return nil, err
	}
	for i := range tools {
		tools[i].Employee = r.loadEmployeeBrief(tools[i].EmployeeID)
	}
	return tools, nil
}

func (r *InventoryRepository) GetPersonalTool(id string) (*model.PersonalTool, error) {
	var t model.PersonalTool
	if err := r.db.Get(&t, `SELECT * FROM "PersonalTool" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &t, nil
}

// logToolEvent يكتب حدث بسجل حركة الأداة. فشل الكتابة ما يفشّل العملية
// الأساسية — السجل توثيق مساعد، مو شرط لنجاح التعديل.
func (r *InventoryRepository) logToolEvent(t *model.PersonalTool, eventType string, from, to, note, actorID *string) {
	_, _ = r.db.Exec(`
		INSERT INTO "PersonalToolEvent" (id, "toolId", "toolName", "employeeId", "eventType", "fromStatus", "toStatus", note, "actorId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8)
	`, t.ID, t.Name, t.EmployeeID, eventType, from, to, note, actorID)
}

// CreatePersonalTool ينشئ أداة خاصة لموظف محدد.
//
// الباركود ما عاد يُطلب من المستخدم (شيلناه من الفورم) — يبقى عمود NOT NULL
// بقاعدة البيانات، فنولّده هنا لما يجي فاضي. التتبع الفريد صار بمعرّف
// الأداة نفسه، مو بالباركود.
func (r *InventoryRepository) CreatePersonalTool(employeeID, name, barcode string, actorID *string) (*model.PersonalTool, error) {
	var t model.PersonalTool
	err := r.db.Get(&t, `
		INSERT INTO "PersonalTool" (id, "employeeId", name, barcode)
		VALUES (gen_random_uuid()::text, $1, $2,
			COALESCE(NULLIF($3, ''), 'TOOL-' || substr(gen_random_uuid()::text, 1, 12)))
		RETURNING *
	`, employeeID, name, barcode)
	if err != nil {
		return nil, err
	}
	r.logToolEvent(&t, model.ToolEventCreated, nil, &t.Status, nil, actorID)
	return &t, nil
}

func (r *InventoryRepository) UpdatePersonalTool(id string, req model.UpdatePersonalToolRequest, actorID *string) (*model.PersonalTool, error) {
	before, err := r.GetPersonalTool(id)
	if err != nil {
		return nil, err
	}
	var t model.PersonalTool
	err = r.db.Get(&t, `
		UPDATE "PersonalTool" SET
			name = COALESCE($2, name),
			barcode = COALESCE($3, barcode),
			status = COALESCE($4, status),
			"checkedOut" = COALESCE($5, "checkedOut")
		WHERE id = $1
		RETURNING *
	`, id, req.Name, req.Barcode, req.Status, req.CheckedOut)
	if err != nil {
		return nil, err
	}

	// حدث لكل تغيير فعلي — تغيير الحالة هو الي يوثّق وقت الفقدان
	if before.Status != t.Status {
		r.logToolEvent(&t, model.ToolEventStatusChanged, &before.Status, &t.Status, req.Note, actorID)
	}
	if before.Name != t.Name || before.Barcode != t.Barcode {
		r.logToolEvent(&t, model.ToolEventRenamed, &before.Name, &t.Name, req.Note, actorID)
	}
	if before.CheckedOut != t.CheckedOut {
		ev := model.ToolEventReturned
		if t.CheckedOut {
			ev = model.ToolEventCheckedOut
		}
		r.logToolEvent(&t, ev, nil, nil, req.Note, actorID)
	}
	return &t, nil
}

func (r *InventoryRepository) DeletePersonalTool(id string, actorID *string) error {
	// نسجّل الحذف قبل ما ننفذه — بعده ما نعرف اسم الأداة ولا صاحبها
	t, getErr := r.GetPersonalTool(id)
	if getErr == nil {
		r.logToolEvent(t, model.ToolEventDeleted, &t.Status, nil, nil, actorID)
	}
	if _, err := r.db.Exec(`DELETE FROM "PersonalTool" WHERE id = $1`, id); err != nil {
		return err
	}
	// ═══ ⚠️ الحذف يستثنيها من نواقص هذا الموظف بنفس الوقت ═══
	//
	// بدون هذا السطر، حذف الأداة **هو الي يخلقها نقصاً**: اسمها لسه
	// بالقالب القياسي، فترجع فوراً بتقرير النواقص كـ«خالصة من
	// المخزن». وهاي بالضبط شكوى أبو الكميات — يحذفها وترجعله.
	//
	// ⚠️ والاستثناء لهذا الموظف وحده — القالب مشترك، وحذفها منه
	// يعني ماكو ولا فني يتحاسب عليها.
	//
	// ⚠️ وما نفشّل الحذف لو فشل الاستثناء: الأداة انحذفت فعلاً،
	// وإرجاع خطأ هنا يخلّي أبو الكميات يظن إنها ما انحذفت فيعيد.
	// بس ما ننبلع الخطأ بصمت — ينسجّل بالسجل.
	if getErr == nil {
		if err := r.ExemptPersonalTool(t.EmployeeID, t.Name, nil, actorID); err != nil {
			log.Printf("استثناء الأداة بعد الحذف فشل (employee=%s tool=%s): %v", t.EmployeeID, t.Name, err)
		}
	}
	return nil
}

// ExemptPersonalTool يشيل أداة من نواقص موظف بعينه (بلا مساس بالقالب).
//
// ⚠️ الاسم ينخزن **مقصوصاً** — نفس ما يتحسب النقص بالضبط (بالمقارنة
// النصية بعد `trim`)، وإلا «مفتاح » ما تطابق «مفتاح» والاستثناء ما
// يشتغل والمستخدم ما يفهم ليش.
func (r *InventoryRepository) ExemptPersonalTool(employeeID, toolName string, note, actorID *string) error {
	name := strings.TrimSpace(toolName)
	if employeeID == "" || name == "" {
		return fmt.Errorf("لازم تحدد الموظف واسم الأداة")
	}
	byName := ""
	if actorID != nil {
		_ = r.db.Get(&byName, `SELECT name FROM "Employee" WHERE id = $1`, *actorID)
	}
	_, err := r.db.Exec(`
		INSERT INTO "PersonalToolExemption" (id, "employeeId", "toolName", note, "byEmployeeId", "byName")
		VALUES (gen_random_uuid()::text, $1, $2, NULLIF($3,''), $4, $5)
		ON CONFLICT ("employeeId", "toolName") DO NOTHING`,
		employeeID, name, deref(note), actorID, byName)
	return err
}

// UnexemptPersonalTool يرجّع الأداة لنواقص الموظف — تراجع عن الاستثناء.
func (r *InventoryRepository) UnexemptPersonalTool(employeeID, toolName string) error {
	_, err := r.db.Exec(`DELETE FROM "PersonalToolExemption"
		WHERE "employeeId" = $1 AND "toolName" = $2`, employeeID, strings.TrimSpace(toolName))
	return err
}

// ListPersonalToolExemptions كل الاستثناءات — الواجهة تطرحها من النواقص.
func (r *InventoryRepository) ListPersonalToolExemptions() ([]model.PersonalToolExemption, error) {
	rows := []model.PersonalToolExemption{}
	err := r.db.Select(&rows, `SELECT * FROM "PersonalToolExemption" ORDER BY "createdAt" DESC`)
	return rows, err
}

const toolEventSelect = `SELECT ev.*, a.name AS "actorName", e.name AS "employeeName"
	FROM "PersonalToolEvent" ev
	LEFT JOIN "Employee" a ON a.id = ev."actorId"
	LEFT JOIN "Employee" e ON e.id = ev."employeeId"`

func (r *InventoryRepository) decorateEvents(events []model.PersonalToolEvent) []model.PersonalToolEvent {
	for i := range events {
		events[i].EventLabel = model.ToolEventLabels[events[i].EventType]
		if events[i].FromStatus != nil {
			events[i].FromStatusText = model.PersonalToolStatusLabels[*events[i].FromStatus]
		}
		if events[i].ToStatus != nil {
			events[i].ToStatusText = model.PersonalToolStatusLabels[*events[i].ToStatus]
		}
	}
	return events
}

// ListToolEvents سجل حركة: لأداة وحدة (toolID)، أو لكل أدوات موظف
// (employeeID)، أو الأحداث كلها لما الاثنين فاضيين.
func (r *InventoryRepository) ListToolEvents(toolID, employeeID string) ([]model.PersonalToolEvent, error) {
	events := []model.PersonalToolEvent{}
	var err error
	switch {
	case toolID != "":
		err = r.db.Select(&events, toolEventSelect+` WHERE ev."toolId" = $1 ORDER BY ev."createdAt" DESC`, toolID)
	case employeeID != "":
		err = r.db.Select(&events, toolEventSelect+` WHERE ev."employeeId" = $1 ORDER BY ev."createdAt" DESC LIMIT 200`, employeeID)
	default:
		err = r.db.Select(&events, toolEventSelect+` ORDER BY ev."createdAt" DESC LIMIT 200`)
	}
	if err != nil {
		return nil, err
	}
	return r.decorateEvents(events), nil
}

// ── Booking Tool Check (لقطة الأدوات الناقصة عند استلام حجز) ────────────────

// ListPersonalToolsByIDs يرجّع أدوات شخصية محددة بالمعرّفات — تستخدم لتحويل
// missingToolIds المرسلة من الواجهة لأسماء أدوات قابلة للقراءة عند تسجيل اللقطة.
func (r *InventoryRepository) ListPersonalToolsByIDs(ids []string) ([]model.PersonalTool, error) {
	tools := []model.PersonalTool{}
	if len(ids) == 0 {
		return tools, nil
	}
	err := r.db.Select(&tools, `SELECT * FROM "PersonalTool" WHERE id = ANY($1)`, pq.Array(ids))
	return tools, err
}

func (r *InventoryRepository) CreateBookingToolCheck(bookingID, employeeID string, missingItems *string) (*model.BookingToolCheck, error) {
	var c model.BookingToolCheck
	err := r.db.Get(&c, `
		INSERT INTO "BookingToolCheck" (id, "bookingId", "employeeId", "missingItems")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, bookingID, employeeID, missingItems)
	if err != nil {
		return nil, err
	}
	c.Employee = r.loadEmployeeBrief(c.EmployeeID)
	return &c, nil
}

// ListAllBookingToolChecks يرجّع كل لقطات الأدوات الناقصة عند استلام الحجوزات
// (بدون تصفية حسب حجز معين) — تستخدم بتبويب "تقارير النواقص" بصفحة المخزون.
func (r *InventoryRepository) ListAllBookingToolChecks() ([]model.BookingToolCheck, error) {
	checks := []model.BookingToolCheck{}
	err := r.db.Select(&checks, `SELECT * FROM "BookingToolCheck" ORDER BY "checkedAt" DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	for i := range checks {
		checks[i].Employee = r.loadEmployeeBrief(checks[i].EmployeeID)
	}
	return checks, nil
}

func (r *InventoryRepository) ListBookingToolChecks(bookingID string) ([]model.BookingToolCheck, error) {
	checks := []model.BookingToolCheck{}
	err := r.db.Select(&checks, `SELECT * FROM "BookingToolCheck" WHERE "bookingId" = $1 ORDER BY "checkedAt" DESC`, bookingID)
	if err != nil {
		return nil, err
	}
	for i := range checks {
		checks[i].Employee = r.loadEmployeeBrief(checks[i].EmployeeID)
	}
	return checks, nil
}

// ── Vehicle Tools ───────────────────────────────────────────────────────────

// vehicleToolSelect يجيب اسم ورقم لوحة السيارة مع الأداة باستعلام واحد، حتى
// الجدول يعرض اسم السيارة بدل معرّفها بدون استعلام إضافي لكل صف.
const vehicleToolSelect = `SELECT t.*,
	COALESCE(v.name, '') AS "vehicleName",
	COALESCE(v."plateNumber", '') AS "vehiclePlate"
	FROM "VehicleTool" t LEFT JOIN "Vehicle" v ON v.id = t."vehicleId"`

func (r *InventoryRepository) ListVehicleTools(vehicleID string) ([]model.VehicleTool, error) {
	tools := []model.VehicleTool{}
	var err error
	if vehicleID != "" {
		err = r.db.Select(&tools, vehicleToolSelect+` WHERE t."vehicleId" = $1 ORDER BY t."createdAt" DESC`, vehicleID)
	} else {
		err = r.db.Select(&tools, vehicleToolSelect+` ORDER BY t."createdAt" DESC`)
	}
	return tools, err
}

func (r *InventoryRepository) getVehicleTool(id string) (*model.VehicleTool, error) {
	var t model.VehicleTool
	if err := r.db.Get(&t, vehicleToolSelect+` WHERE t.id = $1`, id); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) CreateVehicleTool(name string, barcode *string, quantity int, vehicleID string) (*model.VehicleTool, error) {
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "VehicleTool" (id, name, barcode, quantity, "vehicleId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING id
	`, name, barcode, quantity, vehicleID)
	if err != nil {
		return nil, err
	}
	return r.getVehicleTool(id)
}

func (r *InventoryRepository) UpdateVehicleTool(id string, name, barcode, status, vehicleID *string, quantity *int) (*model.VehicleTool, error) {
	_, err := r.db.Exec(`
		UPDATE "VehicleTool" SET
			name = COALESCE($2, name),
			barcode = COALESCE($3, barcode),
			status = COALESCE($4, status),
			"vehicleId" = COALESCE($5, "vehicleId"),
			quantity = COALESCE($6, quantity)
		WHERE id = $1
	`, id, name, barcode, status, vehicleID, quantity)
	if err != nil {
		return nil, err
	}
	return r.getVehicleTool(id)
}

func (r *InventoryRepository) DeleteVehicleTool(id string) error {
	_, err := r.db.Exec(`DELETE FROM "VehicleTool" WHERE id = $1`, id)
	return err
}

// ── On-demand Tools ─────────────────────────────────────────────────────────

func (r *InventoryRepository) ListOnDemandTools() ([]model.OnDemandTool, error) {
	tools := []model.OnDemandTool{}
	err := r.db.Select(&tools, `SELECT * FROM "OnDemandTool" ORDER BY name ASC`)
	return tools, err
}

func (r *InventoryRepository) GetOnDemandTool(id string) (*model.OnDemandTool, error) {
	var t model.OnDemandTool
	if err := r.db.Get(&t, `SELECT * FROM "OnDemandTool" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) CreateOnDemandTool(name, barcode string, totalQty, availableQty int) (*model.OnDemandTool, error) {
	var t model.OnDemandTool
	err := r.db.Get(&t, `
		INSERT INTO "OnDemandTool" (id, name, barcode, "totalQuantity", "availableQuantity")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, name, barcode, totalQty, availableQty)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) UpdateOnDemandTool(id string, name, barcode, status *string, totalQty, availableQty *int) (*model.OnDemandTool, error) {
	var t model.OnDemandTool
	err := r.db.Get(&t, `
		UPDATE "OnDemandTool" SET
			name = COALESCE($2, name),
			barcode = COALESCE($3, barcode),
			"totalQuantity" = COALESCE($4, "totalQuantity"),
			"availableQuantity" = COALESCE($5, "availableQuantity"),
			status = COALESCE($6, status)
		WHERE id = $1
		RETURNING *
	`, id, name, barcode, totalQty, availableQty, status)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *InventoryRepository) IncrementAvailableQuantity(id string) error {
	_, err := r.db.Exec(`UPDATE "OnDemandTool" SET "availableQuantity" = "availableQuantity" + 1 WHERE id = $1`, id)
	return err
}

// ── Tool Requests ───────────────────────────────────────────────────────────

func (r *InventoryRepository) hydrateRequest(req *model.ToolRequest) {
	req.Employee = r.loadEmployeeBrief(req.EmployeeID)
	if tool, err := r.GetOnDemandTool(req.ToolID); err == nil {
		req.Tool = tool
	}
	if req.ApprovedByID != nil {
		req.ApprovedBy = r.loadEmployeeBrief(*req.ApprovedByID)
	}
	if req.Reason != nil {
		req.ReasonLabel = model.ToolRequestReasonLabels[*req.Reason]
	}
	if req.RequestKind != nil {
		req.KindLabel = model.ToolRequestKindLabels[*req.RequestKind]
	}
}

func (r *InventoryRepository) ListToolRequests(employeeID string) ([]model.ToolRequest, error) {
	requests := []model.ToolRequest{}
	var err error
	if employeeID != "" {
		err = r.db.Select(&requests, `SELECT * FROM "ToolRequest" WHERE "employeeId" = $1 ORDER BY "requestedAt" DESC`, employeeID)
	} else {
		err = r.db.Select(&requests, `SELECT * FROM "ToolRequest" ORDER BY "requestedAt" DESC`)
	}
	if err != nil {
		return nil, err
	}
	for i := range requests {
		r.hydrateRequest(&requests[i])
	}
	return requests, nil
}

func (r *InventoryRepository) CreateToolRequest(employeeID, toolID, reason string, kind *string, description *string) (*model.ToolRequest, error) {
	// لو الواجهة ما بعثت تصنيف، نستنتجه من السبب — حتى ما يضل طلب بلا سلة
	k := model.KindForReason(reason)
	if kind != nil && *kind != "" {
		k = *kind
	}
	var req model.ToolRequest
	err := r.db.Get(&req, `
		INSERT INTO "ToolRequest" (id, "employeeId", "toolId", reason, "requestKind", description)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, employeeID, toolID, reason, k, description)
	if err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

func (r *InventoryRepository) DeleteToolRequest(id string) error {
	_, err := r.db.Exec(`DELETE FROM "ToolRequest" WHERE id = $1`, id)
	return err
}

func (r *InventoryRepository) GetToolRequest(id string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	if err := r.db.Get(&req, `SELECT * FROM "ToolRequest" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

// ApproveToolRequest يوافق على الطلب، ولو الأداة انشترت (مو متوفرة بالمخزن)
// يخزن سعرها ومعرّف طلب المشتريات المتولّد حتى تنربط السلسلة كاملة.
// ApproveToolRequest يوافق على الطلب *وينقّص الكمية من مخزن إداري الكميات*.
//
// deductStock=true لما الأداة موجودة بالمخزن فعلاً (يعني انعطت من
// الرف). لما تكون مو موجودة وينشترى إلها، ماكو شي ينتنقص.
//
// قبل هيچي الموافقة كانت تقلب الحالة بس والكمية ما تتحرك أبداً — يعني
// إداري الكميات يشوف رقم ما يمثّل شي، وينطي أدوات وهو يحسب إنها لسّه
// عنده. الخصم داخل نفس المعاملة، وشرط availableQuantity > 0 داخل
// التحديث نفسه حتى ما ينخصم رقمين بنفس اللحظة ويطلع بالسالب.
func (r *InventoryRepository) ApproveToolRequest(id, approvedByID string, purchasePrice *float64, procurementRequestID *string, deductStock bool) (*model.ToolRequest, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var req model.ToolRequest
	if err := tx.Get(&req, `
		UPDATE "ToolRequest" SET status = 'APPROVED', "approvedById" = $2, "approvedAt" = now(),
			"purchasePrice" = COALESCE($3, "purchasePrice"),
			"procurementRequestId" = COALESCE($4, "procurementRequestId")
		WHERE id = $1 AND status = 'PENDING'
		RETURNING *
	`, id, approvedByID, purchasePrice, procurementRequestID); err != nil {
		return nil, fmt.Errorf("الطلب مو معلّق — يمكن انبتّ بيه من قبل")
	}

	if deductStock {
		var left int
		if err := tx.Get(&left, `
			UPDATE "OnDemandTool"
			SET "availableQuantity" = "availableQuantity" - 1
			WHERE id = $1 AND "availableQuantity" > 0
			RETURNING "availableQuantity"`, req.ToolID); err != nil {
			return nil, fmt.Errorf("الأداة خلصت من المخزن — حدّث الكمية أو اشترِ وحدة")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

// ReturnToolStock يرجّع وحدة للمخزن لما الأداة تنرجع.
func (r *InventoryRepository) ReturnToolStock(toolID string) error {
	_, err := r.db.Exec(`
		UPDATE "OnDemandTool"
		SET "availableQuantity" = LEAST("availableQuantity" + 1, "totalQuantity")
		WHERE id = $1`, toolID)
	return err
}

func (r *InventoryRepository) RejectToolRequest(id string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	err := r.db.Get(&req, `
		UPDATE "ToolRequest" SET status = 'REJECTED'
		WHERE id = $1
		RETURNING *
	`, id)
	if err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

// ── Personal Tool Template (العدة القياسية) ─────────────────────────────────

// toolKitEligibleSQL شرط "منو يستحق عدة قياسية".
//
// العدة تخص الي يشتغل بيدينه بالميدان: الفني والليدر. موظف المبيعات
// والمصمم ومهندس الجودة وإداري الكوادر والمدقق ومدير النظام وإدارة
// المشاريع — ما عندهم عدة أصلاً وما يتحاسبون عليها.
//
// قبل هذا الشرط، أي أداة تنضاف للعدة القياسية كانت تروح لـ"كل" موظف بلا
// استثناء، فطلع موظف مبيعات معلّق برقبته ٣٩ أداة وإداري كوادر ٤١.
// التقني يشتغل بالميدان مثل الفني — فتنطبق عليه العدة القياسية بعد
const toolKitEligibleSQL = `(e.role IN ('TECHNICIAN', 'TECHNICAL') OR e."isLeader" = true)`

func (r *InventoryRepository) ListPersonalToolTemplateItems() ([]model.PersonalToolTemplateItem, error) {
	items := []model.PersonalToolTemplateItem{}
	err := r.db.Select(&items, `SELECT * FROM "PersonalToolTemplateItem" ORDER BY "createdAt" ASC`)
	return items, err
}

// CreatePersonalToolTemplateItem يضيف عنصر جديد للعدة القياسية، وفوراً ينشئ
// PersonalTool مطابق لكل موظف حالي (barcode مولّد تلقائياً لأنه عمود NOT NULL
// وما عنده معنى حقيقي هون — التتبع الفريد صار بالعدة القياسية نفسها، مو بالباركود).
func (r *InventoryRepository) CreatePersonalToolTemplateItem(name string) (*model.PersonalToolTemplateItem, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// اسم مكرر يعني نسخة ثانية من نفس الأداة تنضاف لكل موظف — لاحظناها
	// بالاختبار: إضافة "مفك" مرتين خلّت كل فني عنده مفكين.
	var exists bool
	if err := tx.Get(&exists, `
		SELECT EXISTS(SELECT 1 FROM "PersonalToolTemplateItem" WHERE lower(btrim(name)) = lower(btrim($1)))
	`, name); err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("الأداة %q موجودة أصلاً بالعدة القياسية", name)
	}

	var item model.PersonalToolTemplateItem
	if err := tx.Get(&item, `
		INSERT INTO "PersonalToolTemplateItem" (id, name)
		VALUES (gen_random_uuid()::text, $1)
		RETURNING *
	`, name); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(`
		INSERT INTO "PersonalTool" (id, "employeeId", name, barcode)
		SELECT gen_random_uuid()::text, e.id, $1, 'TPL-' || substr(gen_random_uuid()::text, 1, 12)
		FROM "Employee" e
		WHERE `+toolKitEligibleSQL+`
	`, name); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *InventoryRepository) DeletePersonalToolTemplateItem(id string) error {
	_, err := r.db.Exec(`DELETE FROM "PersonalToolTemplateItem" WHERE id = $1`, id)
	return err
}

// ApplyPersonalToolTemplateToEmployee يضيف PersonalTool لكل عنصر بالعدة القياسية
// لموظف واحد — يُستدعى فور إنشاء موظف جديد (employee_service.go) حتى ياخذ
// القائمة كاملة تلقائياً بدون أي خطوة يدوية من الإداري.
//
// ما ينطبق إلا على المستحقين (شوف toolKitEligibleSQL) — موظف مبيعات أو
// مصمم جديد ما ينضاف له ولا أداة.
func (r *InventoryRepository) ApplyPersonalToolTemplateToEmployee(employeeID string) error {
	_, err := r.db.Exec(`
		INSERT INTO "PersonalTool" (id, "employeeId", name, barcode)
		SELECT gen_random_uuid()::text, e.id, t.name, 'TPL-' || substr(gen_random_uuid()::text, 1, 12)
		FROM "PersonalToolTemplateItem" t
		CROSS JOIN "Employee" e
		WHERE e.id = $1 AND `+toolKitEligibleSQL+`
	`, employeeID)
	return err
}

// SyncPersonalToolKitForEmployee يضبط عدة موظف بعد ما يتغيّر دوره أو صفة
// الليدر: إذا صار مستحق ياخذ العدة القياسية الناقصة، وإذا ما عاد مستحق
// تنشال عدته. بدونها، فني يتحوّل لمبيعات يضل معلّق برقبته ٤٠ أداة.
func (r *InventoryRepository) SyncPersonalToolKitForEmployee(employeeID string) error {
	var eligible bool
	if err := r.db.Get(&eligible, `SELECT `+toolKitEligibleSQL+` FROM "Employee" e WHERE e.id = $1`, employeeID); err != nil {
		return err
	}
	if !eligible {
		return r.RemovePersonalToolKit(employeeID)
	}
	// المستحق ياخذ الناقص بس — ما نكرّر أداة موجودة عنده أصلاً
	_, err := r.db.Exec(`
		INSERT INTO "PersonalTool" (id, "employeeId", name, barcode)
		SELECT gen_random_uuid()::text, $1, t.name, 'TPL-' || substr(gen_random_uuid()::text, 1, 12)
		FROM "PersonalToolTemplateItem" t
		WHERE NOT EXISTS (
			SELECT 1 FROM "PersonalTool" p WHERE p."employeeId" = $1 AND p.name = t.name
		)
	`, employeeID)
	return err
}

// RemovePersonalToolKit يشيل عدة موظف ما عاد مستحق. نسجّل حدث حذف لكل أداة
// قبل ما نشيلها — سجل الحركة (PersonalToolEvent) ما عنده مفتاح خارجي وموجود
// أصلاً حتى يوثّق الي راح، فيبقى أثر يقدر المالك يراجعه.
func (r *InventoryRepository) RemovePersonalToolKit(employeeID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO "PersonalToolEvent" (id, "toolId", "toolName", "employeeId", "eventType", note)
		SELECT gen_random_uuid()::text, p.id, p.name, p."employeeId", $2,
			'انشالت لأن دور الموظف ما يستحق عدة قياسية'
		FROM "PersonalTool" p WHERE p."employeeId" = $1
	`, employeeID, model.ToolEventDeleted); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM "PersonalTool" WHERE "employeeId" = $1`, employeeID); err != nil {
		return err
	}
	return tx.Commit()
}

// ── Vehicle Tool Check (لقطة أدوات المركبة الناقصة عند بدء مهمة من ليدر) ────

func (r *InventoryRepository) CreateVehicleToolCheck(vehicleID, missionID, employeeID string, missingToolNames *string) (*model.VehicleToolCheck, error) {
	var c model.VehicleToolCheck
	err := r.db.Get(&c, `
		INSERT INTO "VehicleToolCheck" (id, "vehicleId", "missionId", "employeeId", "missingToolNames")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, vehicleID, missionID, employeeID, missingToolNames)
	if err != nil {
		return nil, err
	}
	c.Employee = r.loadEmployeeBrief(c.EmployeeID)
	return &c, nil
}

func (r *InventoryRepository) ListVehicleToolChecks() ([]model.VehicleToolCheck, error) {
	checks := []model.VehicleToolCheck{}
	err := r.db.Select(&checks, `SELECT * FROM "VehicleToolCheck" ORDER BY "createdAt" DESC`)
	if err != nil {
		return nil, err
	}
	for i := range checks {
		checks[i].Employee = r.loadEmployeeBrief(checks[i].EmployeeID)
	}
	return checks, nil
}

func (r *InventoryRepository) ReturnToolRequest(id string) (*model.ToolRequest, error) {
	var req model.ToolRequest
	err := r.db.Get(&req, `
		UPDATE "ToolRequest" SET status = 'RETURNED', "returnedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id)
	if err != nil {
		return nil, err
	}
	if err := r.IncrementAvailableQuantity(req.ToolID); err != nil {
		return nil, err
	}
	r.hydrateRequest(&req)
	return &req, nil
}

// ── إضافة الكميات للمخزون ────────────────────────────────────────────────────

// AddStock يزيد كمية أداة "حسب الحاجة" ويسجّل الإضافة بأثر كامل.
//
// نزيد المجموع والمتوفر سوا — الكمية الجديدة داخلة للمخزن، مو مصروفة.
func (r *InventoryRepository) AddStock(req model.CreateStockIntakeRequest, byID *string) (*model.StockIntake, error) {
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("الكمية لازم تكون أكبر من صفر")
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE "OnDemandTool"
		SET "totalQuantity" = "totalQuantity" + $2,
			"availableQuantity" = "availableQuantity" + $2
		WHERE id = $1`, req.ToolID, req.Quantity); err != nil {
		return nil, err
	}

	var in model.StockIntake
	if err := tx.Get(&in, `
		INSERT INTO "StockIntake" (id, "toolId", quantity, "unitPrice", supplier, notes, "createdById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, NULLIF($4,''), NULLIF($5,''), $6)
		RETURNING *`, req.ToolID, req.Quantity, req.UnitPrice,
		strDeref(req.Supplier), strDeref(req.Notes), byID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &in, nil
}

// ListStockIntakes سجل إضافات الكميات — منو أضاف وشكد ومتى.
func (r *InventoryRepository) ListStockIntakes(toolID string) ([]model.StockIntake, error) {
	rows := []model.StockIntake{}
	q := `SELECT s.*, t.name AS "toolName", e.name AS "createdName"
		FROM "StockIntake" s
		JOIN "OnDemandTool" t ON t.id = s."toolId"
		LEFT JOIN "Employee" e ON e.id = s."createdById"`
	var err error
	if toolID != "" {
		err = r.db.Select(&rows, q+` WHERE s."toolId" = $1 ORDER BY s."createdAt" DESC`, toolID)
	} else {
		err = r.db.Select(&rows, q+` ORDER BY s."createdAt" DESC LIMIT 300`)
	}
	return rows, err
}

func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
