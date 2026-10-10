package repository

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// AiActionRepository سجل أفعال ماتركس + استعلامات صندوق القرارات.
type AiActionRepository struct {
	db *sqlx.DB
}

func NewAiActionRepository(db *sqlx.DB) *AiActionRepository {
	return &AiActionRepository{db: db}
}

// aiActionSuppressDays المدير رفض فعل؟ نفس النوع على نفس الشي ما
// يرجع يتكرر هالمدة — «لا تذكّر بهذا» قرار، مو إلغاء لمرة وحدة.
const aiActionSuppressDays = 30

// Claim يحجز فعلاً قبل تنفيذه. يرجّع false إذا نفس الفعل انسوّى بنفس
// الفترة، أو المدير رفضه خلال آخر ٣٠ يوم — وبالحالتين ما ننفّذ.
func (r *AiActionRepository) Claim(a model.AiAction) (bool, error) {
	var suppressed bool
	if err := r.db.Get(&suppressed, `
		SELECT EXISTS (SELECT 1 FROM "AiAction"
		  WHERE kind = $1 AND "entityId" = $2 AND status = 'UNDONE'
		    AND "undoneAt" > now() - make_interval(days => $3))`,
		a.Kind, a.EntityID, aiActionSuppressDays); err != nil {
		return false, err
	}
	if suppressed {
		return false, nil
	}
	if paused, err := r.IsPaused(a.Kind); err != nil || paused {
		return false, err
	}
	var details any
	if len(a.Details) > 0 {
		details = []byte(a.Details)
	}
	res, err := r.db.Exec(`
		INSERT INTO "AiAction" (id, kind, "entityType", "entityId", period, "targetEmployeeId", "targetLabel", summary, details)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (kind, "entityId", period) DO NOTHING`,
		uuid.NewString(), a.Kind, a.EntityType, a.EntityID, a.Period, a.TargetEmployeeID, a.TargetLabel, a.Summary, details)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

// CountSince عدد أفعال نوع معيّن من وقت — للحد اليومي.
func (r *AiActionRepository) CountSince(kind string, since time.Time) (int, error) {
	var n int
	err := r.db.Get(&n, `SELECT COUNT(*) FROM "AiAction" WHERE kind = $1 AND "createdAt" >= $2`, kind, since)
	return n, err
}

// ListBetween أفعال فترة، الأحدث أول.
func (r *AiActionRepository) ListBetween(from, to time.Time) ([]model.AiAction, error) {
	rows := []model.AiAction{}
	err := r.db.Select(&rows, `
		SELECT * FROM "AiAction" WHERE "createdAt" >= $1 AND "createdAt" < $2
		ORDER BY "createdAt" DESC LIMIT 300`, from, to)
	return rows, err
}

// Undo المدير يرفض فعلاً.
func (r *AiActionRepository) Undo(id, byEmployeeID string) error {
	res, err := r.db.Exec(`
		UPDATE "AiAction" SET status = 'UNDONE', "undoneById" = $2, "undoneAt" = now()
		WHERE id = $1 AND status = 'DONE'`, id, byEmployeeID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("الفعل مرفوض من قبل أو مو موجود")
	}
	return nil
}

// PendingAiVerdicts أحكام ماتركس بصندوق المراقب الي بعدها ما انراجعت.
func (r *AiActionRepository) PendingAiVerdicts(limit int) ([]model.PendingAiDecision, error) {
	rows := []model.PendingAiDecision{}
	err := r.db.Select(&rows, `
		SELECT mr.id, mr."entityType", mr."entityId", mr.title, mr.summary, mr."createdAt",
		       COALESCE(sg."employeeId", v."blameEmployeeId") AS "employeeId", e.name AS "employeeName",
		       b.id AS "bookingId", b.code AS "bookingCode", sg."occurredAt"
		FROM "MonitorReview" mr
		LEFT JOIN "AiVerdict" v ON mr."entityType" = 'AI_VERDICT' AND v.id = mr."entityId"
		LEFT JOIN "AiSignal" sg ON sg.id = v."signalId"
		LEFT JOIN "Employee" e ON e.id = COALESCE(sg."employeeId", v."blameEmployeeId")
		LEFT JOIN "Booking" b ON sg."entityType" = 'BOOKING' AND b.id = sg."entityId"
		WHERE mr.stage = 'AI_VERDICT' AND mr.status = 'PENDING'
		ORDER BY mr."createdAt" DESC LIMIT $1`, limit)
	return rows, err
}

// UnstaffedOn حجوزات مثبّتة بيوم معيّن (بغداد) وما عليها أي كادر.
func (r *AiActionRepository) UnstaffedOn(day string) ([]model.UnstaffedBooking, error) {
	rows := []model.UnstaffedBooking{}
	err := r.db.Select(&rows, `
		SELECT b.id, b.code, b."scheduledAt"
		FROM "Booking" b
		WHERE b.status = 'CONFIRMED' AND b."scheduledAt" IS NOT NULL
		  AND baghdad_date(b."scheduledAt") = $1::date
		  AND NOT EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id)`+
		BookingCountableAndSQL(`b`)+`
		ORDER BY b."scheduledAt"`, day)
	return rows, err
}

// EmployeeName اسم موظف للعرض — فاضي إذا ماكو.
func (r *AiActionRepository) EmployeeName(id string) string {
	var name string
	_ = r.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, id)
	return name
}

// ═══ المتابعة والتصعيد ═══

// OpenForFollowUp أفعال منفّذة ما انحلت ولا صعدت بعد (آخر ٣٠ يوم).
func (r *AiActionRepository) OpenForFollowUp(kinds []string) ([]model.AiAction, error) {
	rows := []model.AiAction{}
	err := r.db.Select(&rows, `
		SELECT * FROM "AiAction"
		WHERE status = 'DONE' AND "resolvedAt" IS NULL AND kind = ANY($1)
		  AND "createdAt" > now() - interval '30 days'`, pq.Array(kinds))
	return rows, err
}

// OpenFollowableFor تذكيرات موظف واحد المفتوحة (للفحص اللحظي بعد كل حفظ).
func (r *AiActionRepository) OpenFollowableFor(employeeID string, kinds []string) ([]model.AiAction, error) {
	rows := []model.AiAction{}
	err := r.db.Select(&rows, `
		SELECT * FROM "AiAction"
		WHERE "targetEmployeeId" = $1 AND status = 'DONE' AND "resolvedAt" IS NULL AND kind = ANY($2)
		  AND "createdAt" > now() - interval '30 days'`, employeeID, pq.Array(kinds))
	return rows, err
}

func (r *AiActionRepository) MarkResolved(id string) error {
	_, err := r.db.Exec(`UPDATE "AiAction" SET "resolvedAt" = now() WHERE id = $1 AND "resolvedAt" IS NULL`, id)
	return err
}

func (r *AiActionRepository) MarkEscalated(id string) error {
	_, err := r.db.Exec(`UPDATE "AiAction" SET "escalatedAt" = now() WHERE id = $1 AND "escalatedAt" IS NULL`, id)
	return err
}

// IsResolved هل الشي الي ذكّر بي ماتركس انحل؟ السؤال يختلف حسب النوع.
func (r *AiActionRepository) IsResolved(a model.AiAction) (bool, error) {
	var q string
	var args []any
	switch a.Kind {
	case model.AiActionPaperworkReminder:
		// كل حجوزات الليدر الي ذكّرناه بيها صار ورقها كامل.
		var d struct {
			BookingCodes []string `json:"bookingCodes"`
		}
		if err := json.Unmarshal(a.Details, &d); err != nil || len(d.BookingCodes) == 0 {
			return false, nil
		}
		q = `SELECT NOT EXISTS (SELECT 1 FROM "Booking" b WHERE b.code = ANY($1) AND NOT ` + paperworkDoneSQL + `)`
		args = []any{pq.Array(d.BookingCodes)}
	case model.AiActionUnstaffedAlert:
		q = `SELECT EXISTS (SELECT 1 FROM "BookingAssignment" WHERE "bookingId" = $1)
		     OR NOT EXISTS (SELECT 1 FROM "Booking" WHERE id = $1 AND status = 'CONFIRMED')`
		args = []any{a.EntityID}
	case model.AiActionExtraTaskOverdue:
		q = `SELECT NOT EXISTS (SELECT 1 FROM "ExtraTask" WHERE id = $1 AND status IN ('NEW','IN_PROGRESS'))`
		args = []any{a.EntityID}
	case model.AiActionGpsExpiry:
		q = `SELECT EXISTS (SELECT 1 FROM "GpsRenewalRequest" WHERE "deviceRequestId" = $1 AND "createdAt" > $2)
		     OR EXISTS (SELECT 1 FROM "GpsRenewalFollowUp" WHERE "deviceRequestId" = $1 AND "calledAt" > $2)
		     OR NOT EXISTS (SELECT 1 FROM "GpsDeviceRequest" WHERE id = $1
		                    AND "subscriptionEnd" <= now() + interval '14 days')`
		args = []any{a.EntityID, a.CreatedAt}
	case model.AiActionVehicleDocExpiry:
		q = `SELECT NOT EXISTS (SELECT 1 FROM "VehicleDocument" WHERE id = $1
		                        AND "expiryDate" <= now() + interval '30 days')`
		args = []any{a.EntityID}
	case model.AiActionInvoiceApproval:
		q = `SELECT NOT EXISTS (SELECT 1 FROM "LeaderInvoice" WHERE id = $1 AND status = 'SUBMITTED')`
		args = []any{a.EntityID}
	case model.AiActionAttendanceNudge:
		// سجّل حضور بيوم التذكير.
		q = `SELECT EXISTS (SELECT 1 FROM "Attendance" WHERE "employeeId" = $1 AND baghdad_date("checkIn") = $2::date)`
		args = []any{a.EntityID, a.Period}
	case model.AiActionToolReturn:
		q = `SELECT NOT EXISTS (SELECT 1 FROM "ToolRequest" WHERE id = $1 AND status::text = 'APPROVED' AND "returnedAt" IS NULL)`
		args = []any{a.EntityID}
	case model.AiActionMonitorBacklog:
		// الصندوق صار بلا بنود متأخرة (+٢٤ ساعة).
		q = `SELECT (` + monitorBacklogSQL + `) = 0`
	case model.AiActionAfterInventory:
		// جرد عدّته بعد كل حجوزاته المنجزة.
		q = `SELECT NOT EXISTS (SELECT 1 FROM (` + afterPendingSQL + `) x)`
		args = []any{a.EntityID}
	case model.AiActionCrewRating:
		// قيّم كل فنيّي حجوزاته المنجزة.
		q = `SELECT NOT EXISTS (SELECT 1 FROM (` + pendingRatingSQL + `) x)`
		args = []any{a.EntityID}
	case model.AiActionLowStock:
		q = `SELECT NOT EXISTS (SELECT 1 FROM "OnDemandTool" WHERE id = $1
		                        AND "availableQuantity" * 5 <= "totalQuantity")`
		args = []any{a.EntityID}
	default:
		return false, nil
	}
	var ok bool
	err := r.db.Get(&ok, q, args...)
	return ok, err
}

// ═══ التعلّم: الأنواع الموقوفة ═══

func (r *AiActionRepository) IsPaused(kind string) (bool, error) {
	var ok bool
	err := r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "AiActionKindPause" WHERE kind = $1 AND "resumedAt" IS NULL)`, kind)
	return ok, err
}

// RejectsSince كم مرة المدير رفض هذا النوع من وقت — وبعد آخر ترجيع بس،
// وإلا الرفضات القديمة توقفه مرة ثانية بأول رفض جديد.
func (r *AiActionRepository) RejectsSince(kind string, since time.Time) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "AiAction" a
		WHERE a.kind = $1 AND a.status = 'UNDONE' AND a."undoneAt" >= $2
		  AND a."undoneAt" > COALESCE((SELECT "resumedAt" FROM "AiActionKindPause" p WHERE p.kind = $1), '-infinity')`,
		kind, since)
	return n, err
}

func (r *AiActionRepository) Pause(kind, reason string) error {
	_, err := r.db.Exec(`
		INSERT INTO "AiActionKindPause" (kind, reason) VALUES ($1,$2)
		ON CONFLICT (kind) DO UPDATE SET reason = EXCLUDED.reason, "pausedAt" = now(), "resumedAt" = NULL`, kind, reason)
	return err
}

// Resume المدير يرجّع النوع.
func (r *AiActionRepository) Resume(kind string) error {
	res, err := r.db.Exec(`UPDATE "AiActionKindPause" SET "resumedAt" = now() WHERE kind = $1 AND "resumedAt" IS NULL`, kind)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("هذا النوع مو موقوف")
	}
	return nil
}

// ListPaused الأنواع الموقوفة هسه.
func (r *AiActionRepository) ListPaused() ([]model.AiActionKindPause, error) {
	rows := []model.AiActionKindPause{}
	err := r.db.Select(&rows, `SELECT * FROM "AiActionKindPause" WHERE "resumedAt" IS NULL ORDER BY "pausedAt" DESC`)
	return rows, err
}

// KindOf نوع فعل — نحتاجه بعد الرفض حتى نشوف هل لازم نوقف النوع.
func (r *AiActionRepository) KindOf(id string) (string, error) {
	var k string
	err := r.db.Get(&k, `SELECT kind FROM "AiAction" WHERE id = $1`, id)
	return k, err
}

// Escalated أفعال صعدت وبعدها ما انحلت — تطلع بـ«ينتظر قرارك».
func (r *AiActionRepository) Escalated() ([]model.AiAction, error) {
	rows := []model.AiAction{}
	err := r.db.Select(&rows, `
		SELECT * FROM "AiAction" WHERE "escalatedAt" IS NOT NULL AND "resolvedAt" IS NULL AND status = 'DONE'
		  AND "createdAt" > now() - interval '30 days'
		ORDER BY "escalatedAt" DESC LIMIT 100`)
	return rows, err
}

// ═══ الدقة (آخر ٣٠ يوم) ═══

func (r *AiActionRepository) Accuracy() (*model.MatrixAccuracy, error) {
	var a model.MatrixAccuracy
	err := r.db.Get(&a, `
		SELECT
		  COUNT(*)                                                        AS actions,
		  COUNT(*) FILTER (WHERE a."resolvedAt" IS NOT NULL)             AS resolved,
		  COUNT(*) FILTER (WHERE a."escalatedAt" IS NOT NULL)            AS escalated,
		  COUNT(*) FILTER (WHERE a.status = 'UNDONE')                    AS rejected,
		  COUNT(*) FILTER (WHERE a.kind = 'DELAY_WARNING')               AS "delayPredicted",
		  COUNT(b.id)                                                    AS "delayChecked",
		  COUNT(b.id) FILTER (WHERE EXTRACT(EPOCH FROM (b."completedAt" - b."startedAt"))/60
		                          > (a.details->>'availableMinutes')::numeric) AS "delayCorrect"
		FROM "AiAction" a
		LEFT JOIN "Booking" b ON a.kind = 'DELAY_WARNING' AND b.id = a."entityId"
		     AND b."startedAt" IS NOT NULL AND b."completedAt" IS NOT NULL
		     AND a.details ? 'availableMinutes'
		WHERE a."createdAt" > now() - interval '30 days'`)
	return &a, err
}

// ═══ مصادر التذكيرات الجديدة ═══

type GpsExpiringRow struct {
	ID              string    `db:"id"`
	CustomerName    string    `db:"customerName"`
	EmployeeID      string    `db:"employeeId"`
	SubscriptionEnd time.Time `db:"subscriptionEnd"`
	DaysLeft        int       `db:"daysLeft"`
}

// GpsExpiringSoon اشتراكات فعّالة تخلص خلال أيام، ولها بائع معروف.
func (r *AiActionRepository) GpsExpiringSoon(days int) ([]GpsExpiringRow, error) {
	rows := []GpsExpiringRow{}
	err := r.db.Select(&rows, `
		SELECT d.id, COALESCE(c."fullName", '') AS "customerName", d."employeeId",
		       d."subscriptionEnd", (baghdad_date(d."subscriptionEnd") - baghdad_today()) AS "daysLeft"
		FROM "GpsDeviceRequest" d
		LEFT JOIN "GpsCustomer" c ON c.id = d."customerId"
		WHERE d."subscriptionEnd" IS NOT NULL AND d."employeeId" IS NOT NULL
		  AND baghdad_date(d."subscriptionEnd") >= baghdad_today()
		  AND baghdad_date(d."subscriptionEnd") <= baghdad_today() + $1::int
		ORDER BY d."subscriptionEnd"`, days)
	return rows, err
}

type VehicleDocRow struct {
	ID           string    `db:"id"`
	DocumentType string    `db:"documentType"`
	VehicleName  string    `db:"vehicleName"`
	PlateNumber  string    `db:"plateNumber"`
	ExpiryDate   time.Time `db:"expiryDate"`
}

func (r *AiActionRepository) VehicleDocsExpiring(days int) ([]VehicleDocRow, error) {
	rows := []VehicleDocRow{}
	err := r.db.Select(&rows, `
		SELECT d.id, d."documentType", v.name AS "vehicleName", v."plateNumber", d."expiryDate"
		FROM "VehicleDocument" d JOIN "Vehicle" v ON v.id = d."vehicleId"
		WHERE v."isActive" AND d."expiryDate" IS NOT NULL
		  AND d."expiryDate" <= now() + make_interval(days => $1)
		ORDER BY d."expiryDate"`, days)
	return rows, err
}

type OverdueTaskRow struct {
	ID           string    `db:"id"`
	Title        string    `db:"title"`
	AssignedToID string    `db:"assignedToId"`
	AssignedByID *string   `db:"assignedById"`
	DueAt        time.Time `db:"dueAt"`
}

func (r *AiActionRepository) OverdueExtraTasks() ([]OverdueTaskRow, error) {
	rows := []OverdueTaskRow{}
	err := r.db.Select(&rows, `
		SELECT id, title, "assignedToId", "assignedById", "dueAt" FROM "ExtraTask"
		WHERE status IN ('NEW','IN_PROGRESS') AND "dueAt" IS NOT NULL AND "dueAt" < now()
		ORDER BY "dueAt"`)
	return rows, err
}

type LowStockRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Available int    `db:"availableQuantity"`
	Total     int    `db:"totalQuantity"`
}

// LowStockTools أدوات رصيدها ٢٠٪ أو أقل.
func (r *AiActionRepository) LowStockTools() ([]LowStockRow, error) {
	rows := []LowStockRow{}
	err := r.db.Select(&rows, `
		SELECT id, name, "availableQuantity", "totalQuantity" FROM "OnDemandTool"
		WHERE "totalQuantity" > 0 AND "availableQuantity" * 5 <= "totalQuantity"
		ORDER BY "availableQuantity"`)
	return rows, err
}

type StaleInvoiceRow struct {
	ID          string    `db:"id"`
	BookingCode string    `db:"bookingCode"`
	CreatedAt   time.Time `db:"createdAt"`
	Days        int       `db:"days"`
}

// StaleSubmittedInvoices فواتير ليدر مرفوعة وما انعتمدت من أكثر من أيام.
func (r *AiActionRepository) StaleSubmittedInvoices(days int) ([]StaleInvoiceRow, error) {
	rows := []StaleInvoiceRow{}
	err := r.db.Select(&rows, `
		SELECT li.id, COALESCE(b.code, '') AS "bookingCode", li."createdAt",
		       (baghdad_today() - baghdad_date(li."createdAt")) AS days
		FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId"
		WHERE li.status = 'SUBMITTED' AND li."createdAt" < now() - make_interval(days => $1)
		ORDER BY li."createdAt" LIMIT 100`, days)
	return rows, err
}

type NoCheckInRow struct {
	EmployeeID  string `db:"employeeId"`
	BookingCode string `db:"bookingCode"`
}

// WorkTodayNoCheckIn موظفين عندهم حجز اليوم وما سجّلوا حضور.
func (r *AiActionRepository) WorkTodayNoCheckIn() ([]NoCheckInRow, error) {
	rows := []NoCheckInRow{}
	err := r.db.Select(&rows, `
		SELECT DISTINCT ON (a."employeeId") a."employeeId", b.code AS "bookingCode"
		FROM "BookingAssignment" a
		JOIN "Booking" b ON b.id = a."bookingId"
		JOIN "Employee" e ON e.id = a."employeeId" AND e.status = 'ACTIVE'
		WHERE b.status IN ('CONFIRMED','IN_PROGRESS') AND b."scheduledAt" IS NOT NULL
		  AND baghdad_date(b."scheduledAt") = baghdad_today()
		  AND NOT EXISTS (SELECT 1 FROM "Attendance" t WHERE t."employeeId" = a."employeeId"
		                  AND baghdad_date(t."checkIn") = baghdad_today())
		ORDER BY a."employeeId", b."scheduledAt"`)
	return rows, err
}

// TaskAssigner منو كلّف بالمهمة الإضافية — فاضي إذا ماكو.
func (r *AiActionRepository) TaskAssigner(taskID string) string {
	var id string
	_ = r.db.Get(&id, `SELECT COALESCE("assignedById", '') FROM "ExtraTask" WHERE id = $1`, taskID)
	return id
}

// OpenForEmployee تذكيرات ماتركس المفتوحة على موظف — حالة «عينه».
func (r *AiActionRepository) OpenForEmployee(employeeID string) ([]model.AiAction, error) {
	rows := []model.AiAction{}
	err := r.db.Select(&rows, `
		SELECT * FROM "AiAction"
		WHERE "targetEmployeeId" = $1 AND status = 'DONE' AND "resolvedAt" IS NULL
		  AND kind <> 'ATTENDANCE_NUDGE' AND kind <> 'DELAY_WARNING'
		  AND "createdAt" > now() - interval '14 days'
		ORDER BY "createdAt" DESC LIMIT 20`, employeeID)
	return rows, err
}

// ═══ عين ماتركس: «شغلك اليوم» ═══

// WatchSubject الموظف كما تحتاجه العين: دوره، وهل ليدر، وصلاحياته.
type WatchSubject struct {
	ID       string `db:"id"`
	Name     string `db:"name"`
	Role     string `db:"role"`
	IsLeader bool   `db:"isLeader"`
	// IsServiceManager مسجّل مسؤول خدمة بجدول "ServiceManager" — مهما چان دوره.
	IsServiceManager bool `db:"isServiceManager"`
	Perms            []string
}

func (r *AiActionRepository) Subject(id string) (*WatchSubject, error) {
	var s WatchSubject
	if err := r.db.Get(&s, `SELECT id, name, role::text AS role, "isLeader", EXISTS (SELECT 1 FROM "ServiceManager" sm WHERE sm."employeeId" = "Employee".id) AS "isServiceManager" FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	_ = r.db.Select(&s.Perms, `
		SELECT p.name FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
		WHERE ep."employeeId" = $1`, id)
	return &s, nil
}

// ActiveSubjects كل الموظفين الفعّالين — لتقرير المدير.
func (r *AiActionRepository) ActiveSubjects() ([]WatchSubject, error) {
	rows := []WatchSubject{}
	if err := r.db.Select(&rows, `SELECT id, name, role::text AS role, "isLeader", EXISTS (SELECT 1 FROM "ServiceManager" sm WHERE sm."employeeId" = "Employee".id) AS "isServiceManager" FROM "Employee" WHERE status = 'ACTIVE' ORDER BY name`); err != nil {
		return nil, err
	}
	type pr struct {
		EmployeeID string `db:"employeeId"`
		Name       string `db:"name"`
	}
	perms := []pr{}
	_ = r.db.Select(&perms, `SELECT ep."employeeId", p.name FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"`)
	idx := map[string]int{}
	for i := range rows {
		idx[rows[i].ID] = i
	}
	for _, p := range perms {
		if i, ok := idx[p.EmployeeID]; ok {
			rows[i].Perms = append(rows[i].Perms, p.Name)
		}
	}
	return rows, nil
}

func (r *AiActionRepository) count(q string, args ...any) int {
	var n int
	_ = r.db.Get(&n, q, args...)
	return n
}

// AuditToday حجوزات اليوم المنجزة وورقها كامل: كم باقي ما تدقق.
func (r *AiActionRepository) AuditLeftToday() int {
	return r.count(`SELECT COUNT(*) FROM "Booking" b
		WHERE b.status = 'COMPLETED' AND b."completedAt" IS NOT NULL
		  AND baghdad_date(b."completedAt") = baghdad_today()
		  AND NOT b."amountVerified" AND ` + paperworkDoneSQL + BookingCountableAndSQL(`b`))
}

// AuditedByToday فواتير دققها هالموظف اليوم.
func (r *AiActionRepository) AuditedByToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "LeaderInvoice" WHERE "auditedById" = $1 AND baghdad_date("auditedAt") = baghdad_today()`, id)
}

func (r *AiActionRepository) MonitorPending() int {
	return r.count(`SELECT COUNT(*) FROM "MonitorReview" WHERE status = 'PENDING'`)
}

func (r *AiActionRepository) BookingsPendingConfirm() int {
	return r.count(`SELECT COUNT(*) FROM "Booking" b WHERE b.status = 'PENDING'` + BookingCountableAndSQL(`b`))
}

// LeaderToday حجوزات الليدر اليوم: كلها، والمنجز منها.
func (r *AiActionRepository) LeaderToday(id string) (total, done int) {
	q := `SELECT COUNT(*) FROM "Booking" b JOIN "BookingAssignment" a ON a."bookingId" = b.id
		WHERE a."employeeId" = $1 AND b."scheduledAt" IS NOT NULL
		  AND baghdad_date(b."scheduledAt") = baghdad_today() AND b.status <> 'CANCELLED'`
	total = r.count(q, id)
	done = r.count(q+` AND b.status = 'COMPLETED'`, id)
	return
}

func (r *AiActionRepository) InvoicesAwaitingApproval() int {
	return r.count(`SELECT COUNT(*) FROM "LeaderInvoice" WHERE status = 'SUBMITTED'`)
}

func (r *AiActionRepository) ApprovedByToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "LeaderInvoice" WHERE "approvedByEmployeeId" = $1 AND baghdad_date("approvedAt") = baghdad_today()`, id)
}

func (r *AiActionRepository) QualityPending() int {
	return r.count(`SELECT COUNT(*) FROM "QualityFollowUp" WHERE "inspectionStatus" = 'PENDING'`)
}

// ═══ شغل الفني العادي والمصمم والليدر (عين ماتركس لكل دور) ═══

// CheckedInToday سجّل حضور اليوم؟ وانصرف؟
func (r *AiActionRepository) CheckedInToday(id string) (in bool, out bool) {
	_ = r.db.Get(&in, `SELECT EXISTS (SELECT 1 FROM "Attendance" WHERE "employeeId" = $1 AND baghdad_date("checkIn") = baghdad_today())`, id)
	_ = r.db.Get(&out, `SELECT EXISTS (SELECT 1 FROM "Attendance" WHERE "employeeId" = $1 AND baghdad_date("checkIn") = baghdad_today() AND "checkOut" IS NOT NULL)`, id)
	return
}

// ToolsHeld أدوات بعهدته (موافق عليها وما رجعت).
func (r *AiActionRepository) ToolsHeld(id string) int {
	return r.count(`SELECT COUNT(*) FROM "ToolRequest" WHERE "employeeId" = $1 AND status = 'APPROVED' AND "returnedAt" IS NULL`, id)
}

// DesignUploadsToday أعمال تصميم رفعها اليوم.
func (r *AiActionRepository) DesignUploadsToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "DesignAsset" WHERE "uploadedById" = $1 AND baghdad_date("createdAt") = baghdad_today()`, id)
}

// LeaderMaterialsPending حجوزات الليدر اليوم وموادها ما تجهزت بعد — بأكوادها.
func (r *AiActionRepository) LeaderMaterialsPending(id string) []string {
	codes := []string{}
	_ = r.db.Select(&codes, `SELECT b.code FROM "Booking" b JOIN "BookingAssignment" a ON a."bookingId" = b.id
		WHERE a."employeeId" = $1 AND b."scheduledAt" IS NOT NULL AND baghdad_date(b."scheduledAt") = baghdad_today()
		  AND b.status IN ('CONFIRMED','PENDING') AND b."materialsReadyAt" IS NULL ORDER BY b."scheduledAt"`, id)
	return codes
}

// QualityEngineerForGps مهندس الجودة الأقل تذكيرات مفتوحة — له تنروح تذكيرات
// تجديد الجي بي اس (قرار (ع): مو للبائع، والبائع بالاشتراكات القديمة هو المالك).
func (r *AiActionRepository) QualityEngineerForGps() string {
	var id string
	_ = r.db.Get(&id, `SELECT e.id FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND (e.role = 'QUALITY_ENGINEER' OR EXISTS (
			SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
			WHERE ep."employeeId" = e.id AND p.name = 'quality_control' AND e.role NOT IN ('ADMIN','OWNER','MONITOR')))
		ORDER BY (SELECT COUNT(*) FROM "AiAction" a WHERE a."targetEmployeeId" = e.id AND a."resolvedAt" IS NULL), e.name
		LIMIT 1`)
	return id
}

// ReassignOpenGps ينقل تذكيرات التجديد المفتوحة الي على المالك/المدير:
// لمهندس الجودة إذا موجود (toID)، وإلا تتشال من على الشخص (تبقى للجودة كقسم)
// — بالحالتين ما تبقى بعين المالك. والنص القديم «ذكّر البائع» يتحدّث.
func (r *AiActionRepository) ReassignOpenGps(kind, toID, label string) {
	if label == "" {
		label = "الجودة"
	}
	_, _ = r.db.Exec(`UPDATE "AiAction" a SET "targetEmployeeId" = NULLIF($2, ''), "targetLabel" = $3,
		summary = replace(a.summary, 'ذكّر البائع', 'ذكّر الجودة')
		FROM "Employee" e WHERE e.id = a."targetEmployeeId" AND a.kind = $1 AND a."resolvedAt" IS NULL
		  AND e.role IN ('ADMIN', 'OWNER')`, kind, toID, label)
}

// ═══ شغل كل دور بآخر N يوم — ماتركس يقيس كل موظف بشغل دوره هو ═══
// (ع): «مو كل الموظفين ينحسبلهم حجوزات — عندي محاسب ومراقب ومدير».

type RoleCounts struct {
	InvoicesApproved int // فواتير اعتمدها (المحاسبة)
	InvoicesAudited  int // فواتير دقّقها (المراقبة / تدقيق مالي)
	Reviews          int // بنود حكم عليها بصندوق المراقب
	ReviewsFlagged   int // منها «عليه ملاحظة»
	BookingsCreated  int // حجوزات سجّلها (التنسيق / المبيعات)
	QualityDone      int // متابعات جودة: اتصل أو فحص
	DesignUploads    int // تصاميم رفعها
	Projects         int // مشاريع سجّلها أو مسؤول عنها (مدير المشاريع)
	GpsRequests      int // طلبات أجهزة/تجديد/صيانة جي بي اس سجّلها
	GpsCalls         int // اتصالات متابعة تجديد الجي بي اس
	Actions          int // كل عملية حفظ بالنظام (سجل النشاط)
}

func (r *AiActionRepository) RoleCounts(id string, days int) RoleCounts {
	since := `now() - make_interval(days => $2)`
	return RoleCounts{
		InvoicesApproved: r.count(`SELECT COUNT(*) FROM "LeaderInvoice" WHERE "approvedByEmployeeId" = $1 AND "approvedAt" >= `+since, id, days),
		InvoicesAudited:  r.count(`SELECT COUNT(*) FROM "LeaderInvoice" WHERE "auditedById" = $1 AND "auditedAt" >= `+since, id, days),
		Reviews:          r.count(`SELECT COUNT(*) FROM "MonitorReview" WHERE "reviewedById" = $1 AND "reviewedAt" >= `+since, id, days),
		ReviewsFlagged:   r.count(`SELECT COUNT(*) FROM "MonitorReview" WHERE "reviewedById" = $1 AND status = 'FLAGGED' AND "reviewedAt" >= `+since, id, days),
		BookingsCreated:  r.count(`SELECT COUNT(*) FROM "Booking" WHERE "createdById" = $1 AND "createdAt" >= `+since, id, days),
		QualityDone: r.count(`SELECT COUNT(*) FROM "QualityFollowUp" WHERE ("inspectedById" = $1 AND "inspectedAt" >= `+since+`)
			OR ("contactedByEmployeeId" = $1 AND "contactedAt" >= `+since+`)`, id, days),
		DesignUploads: r.count(`SELECT COUNT(*) FROM "DesignAsset" WHERE "uploadedById" = $1 AND "createdAt" >= `+since, id, days),
		Projects: r.count(`SELECT COUNT(*) FROM "Project" WHERE ("createdByEmployeeId" = $1 OR "responsibleEmployeeId" = $1)
			AND "createdAt" >= `+since, id, days),
		GpsRequests: r.count(`SELECT (SELECT COUNT(*) FROM "GpsDeviceRequest" WHERE "employeeId" = $1 AND "createdAt" >= `+since+`)
			+ (SELECT COUNT(*) FROM "GpsRenewalRequest" WHERE "employeeId" = $1 AND "createdAt" >= `+since+`)
			+ (SELECT COUNT(*) FROM "GpsMaintenanceRequest" WHERE "employeeId" = $1 AND "createdAt" >= `+since+`)`, id, days),
		GpsCalls: r.count(`SELECT COUNT(*) FROM "GpsRenewalFollowUp" WHERE "calledById" = $1 AND "calledAt" >= `+since, id, days),
		Actions:  r.count(`SELECT COUNT(*) FROM "EmployeeActivity" WHERE "employeeId" = $1 AND "createdAt" >= `+since, id, days),
	}
}

// ═══ «سوّيت اليوم» لكل دور — حتى عين ماتركس تشوف المنجز مو بس الباقي ═══

func (r *AiActionRepository) ReviewedByToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "MonitorReview" WHERE "reviewedById" = $1 AND baghdad_date("reviewedAt") = baghdad_today()`, id)
}

func (r *AiActionRepository) ConfirmedByToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "Booking" WHERE "confirmedByEmployeeId" = $1 AND baghdad_date("confirmedAt") = baghdad_today()`, id)
}

func (r *AiActionRepository) QualityDoneToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "QualityFollowUp" WHERE ("inspectedById" = $1 AND baghdad_date("inspectedAt") = baghdad_today())
		OR ("contactedByEmployeeId" = $1 AND baghdad_date("contactedAt") = baghdad_today())`, id)
}

func (r *AiActionRepository) BookingsCreatedToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "Booking" WHERE "createdById" = $1 AND baghdad_date("createdAt") = baghdad_today()`, id)
}

func (r *AiActionRepository) GpsCallsToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "GpsRenewalFollowUp" WHERE "calledById" = $1 AND baghdad_date("calledAt") = baghdad_today()`, id)
}

// GpsExpiringCount اشتراكات جي بي اس تخلص خلال ١٤ يوم — طابور اتصالات التجديد.
func (r *AiActionRepository) GpsExpiringCount() int {
	return r.count(`SELECT COUNT(*) FROM "GpsDeviceRequest" WHERE "subscriptionEnd" BETWEEN now() AND now() + interval '14 days'`)
}

// ItLogsToday سجلات صيانة/تصليح أجهزة سجّلها اليوم (عن طريق سجل النشاط على مسارات it).
func (r *AiActionRepository) ItLogsToday(id string) int {
	return r.count(`SELECT COUNT(*) FROM "EmployeeActivity" WHERE "employeeId" = $1 AND path LIKE '/api/it%'
		AND baghdad_date("createdAt") = baghdad_today()`, id)
}

// LeadersPendingRating الليدرية الي ما قيّموا فنيّيهم — لتذكير ماتركس.
func (r *AiActionRepository) LeadersPendingRating() ([]LeaderPending, error) {
	return NewMatrixChainRepository(r.db).LeadersPendingRating()
}

// TechsPendingAfterInventory الفنيين الي ما جردوا بعد حجوزاتهم — لتذكير ماتركس.
func (r *AiActionRepository) TechsPendingAfterInventory() ([]LeaderPending, error) {
	return NewMatrixChainRepository(r.db).TechsPendingAfterInventory()
}

func (r *AiActionRepository) MonitorBacklog() (MonitorBacklog, error) {
	return NewMatrixChainRepository(r.db).MonitorBacklog()
}

func (r *AiActionRepository) InternalNoInvoice() ([]string, error) {
	return NewMatrixChainRepository(r.db).InternalNoInvoice()
}

// ═══ شغل أبو الكميات (قرار (ع) 10-10) ═══
func (r *AiActionRepository) procScore() *MatrixScoreRepository {
	return NewMatrixScoreRepository(r.db)
}
func (r *AiActionRepository) PendingMaterials() ([]ProcDecision, error) {
	return r.procScore().MaterialDecisions(true)
}
func (r *AiActionRepository) PendingTools() ([]ProcDecision, error) {
	return r.procScore().ToolDecisions(true)
}
func (r *AiActionRepository) OpenShortages() ([]ProcDecision, error) {
	return r.procScore().Shortages(true)
}
func (r *AiActionRepository) VehiclesUnratedToday() ([]string, error) {
	return r.procScore().VehiclesUnratedToday()
}
func (r *AiActionRepository) FleetOverdue() ([]string, error) { return r.procScore().FleetOverdue() }
// UnverifiedProjectPayments دفعات مشاريع ما أكدها المحاسب من +٤٨ ساعة (قرار (ع) 10-10).
func (r *AiActionRepository) UnverifiedProjectPayments() (int, float64, string, error) {
	var n int
	var sum float64
	var top string
	err := r.db.QueryRow(`SELECT count(*), COALESCE(sum(pp.amount), 0),
		COALESCE((SELECT p.name || ' (' || p.code || ')' FROM "ProjectPayment" x JOIN "Project" p ON p.id = x."projectId"
		  WHERE x."verifiedAt" IS NULL AND x."cancelledAt" IS NULL AND x."createdAt" < now() - interval '48 hours'
		  GROUP BY p.id, p.name, p.code ORDER BY sum(x.amount) DESC LIMIT 1), '')
		FROM "ProjectPayment" pp WHERE pp."verifiedAt" IS NULL AND pp."cancelledAt" IS NULL AND pp."createdAt" < now() - interval '48 hours'`).
		Scan(&n, &sum, &top)
	return n, sum, top, err
}

func (r *AiActionRepository) ProcurementAdmins() ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `SELECT id FROM "Employee" WHERE status = 'ACTIVE' AND role::text = 'PROCUREMENT_ADMIN'`)
	return ids, err
}

func (r *AiActionRepository) Monitors() ([]string, error) {
	return NewMatrixChainRepository(r.db).Monitors()
}

// Find فعل واحد — للتراجع عن الأفعال التنفيذية.
func (r *AiActionRepository) Find(id string) (*model.AiAction, error) {
	var a model.AiAction
	if err := r.db.Get(&a, `SELECT * FROM "AiAction" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &a, nil
}

// Forget يمسح فعل انحجز وفشل تنفيذه — حتى ما يبين بالسجل كأنه صار.
func (r *AiActionRepository) Forget(kind, entityID, period string) {
	_, _ = r.db.Exec(`DELETE FROM "AiAction" WHERE kind = $1 AND "entityId" = $2 AND period = $3 AND status = 'DONE'`, kind, entityID, period)
}

// RecentOfKinds آخر أفعال أنواع معيّنة.
func (r *AiActionRepository) RecentOfKinds(kinds []string, limit int) ([]model.AiAction, error) {
	rows := []model.AiAction{}
	err := r.db.Select(&rows, `SELECT * FROM "AiAction" WHERE kind = ANY($1) ORDER BY "createdAt" DESC LIMIT $2`, pq.Array(kinds), limit)
	return rows, err
}

// MonitorOverdue بنود صندوق المراقب +٤٨ ساعة بلا حكم.
func (r *AiActionRepository) MonitorOverdue() (int, error) {
	var n int
	err := r.db.Get(&n, `SELECT count(*) FROM "MonitorReview" WHERE status = 'PENDING' AND "createdAt" < now() - interval '48 hours'`)
	return n, err
}
