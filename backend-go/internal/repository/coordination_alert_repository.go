package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type CoordinationAlertRepository struct {
	db *sqlx.DB
}

func NewCoordinationAlertRepository(db *sqlx.DB) *CoordinationAlertRepository {
	return &CoordinationAlertRepository{db: db}
}

// Add يسجّل تنبيهاً ويرجّع عدد المفتوحة بعده.
//
// ⚠️ العدّ يرجع من نفس المعاملة: المستدعي يقرر النشر على أساسه،
// ولو عدّ بطلب منفصل ممكن ضغطتان متزامنتان تشوفان نفس الرقم
// وتنشران الإعلان مرتين.
func (r *CoordinationAlertRepository) Add(
	bookingID string, coordinatorID, coordinatorName *string,
	reason, byID, byName string,
) (openCount int, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err = tx.Exec(`
		INSERT INTO "CoordinationAlert"
			(id, "bookingId", "coordinatorId", "coordinatorName", reason, "byEmployeeId", "byName")
		VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), $6, NULLIF($7,''))`,
		uuid.NewString(), bookingID, coordinatorID, deref(coordinatorName),
		reason, byID, byName); err != nil {
		return 0, err
	}
	if err = tx.Get(&openCount, `
		SELECT COUNT(*) FROM "CoordinationAlert"
		WHERE "bookingId" = $1 AND "resolvedAt" IS NULL`, bookingID); err != nil {
		return 0, err
	}
	return openCount, tx.Commit()
}

// Resolve يأشّر تنبيهات الحجز المفتوحة «انعالجت».
//
// ⚠️ ما يمحيهن: المعالجة واقعة تنضاف للسجل مو ممحاة له — وإلا
// يقدر أحد ينظّف تاريخه بضغطة.
func (r *CoordinationAlertRepository) Resolve(bookingID, byID, byName, note string) error {
	_, err := r.db.Exec(`
		UPDATE "CoordinationAlert"
		SET "resolvedAt" = now(), "resolvedById" = $2,
		    "resolvedByName" = NULLIF($3,''), "resolveNote" = NULLIF($4,'')
		WHERE "bookingId" = $1 AND "resolvedAt" IS NULL`, bookingID, byID, byName, note)
	return err
}

// ListForBooking سجل الحجز — الأحدث أولاً.
func (r *CoordinationAlertRepository) ListForBooking(bookingID string) ([]model.CoordinationAlert, error) {
	rows := []model.CoordinationAlert{}
	err := r.db.Select(&rows, `
		SELECT * FROM "CoordinationAlert"
		WHERE "bookingId" = $1 ORDER BY "createdAt" DESC`, bookingID)
	return rows, err
}

// Summaries ملخص كل الحجوزات الي عندها تنبيهات — دفعة وحدة.
//
// ⚠️ استعلام واحد مو واحد لكل حجز: الشاشة تعرض عشرات الحجوزات،
// وطلب لكل صف يخليها تزحف.
func (r *CoordinationAlertRepository) Summaries() ([]model.CoordinationAlertSummary, error) {
	rows := []model.CoordinationAlertSummary{}
	err := r.db.Select(&rows, `
		SELECT "bookingId",
		       COUNT(*) FILTER (WHERE "resolvedAt" IS NULL) AS "openCount",
		       COUNT(*)                                     AS "totalCount",
		       MAX("createdAt")                             AS "lastAlertAt"
		FROM "CoordinationAlert"
		GROUP BY "bookingId"`)
	return rows, err
}

// ═══ (ع) 10-13: التقصير على موظف الفترة ═══
// الحجز يجي صباحي أو مسائي (حقل الفترة، وإذا فاضي من ساعة إنشائه ببغداد:
// قبل ٣ العصر صباحي). المسؤول = موظفين التنسيق (صلاحية coordinator) بنفس
// الفترة. وكل ١٠ تقصيرات على الموظف = نقطة كي بي اي «تنفيذ المهام».

type ShiftOwner struct {
	ID      string `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	Strikes int    `db:"strikes" json:"strikes"`
}

type BookingShiftInfo struct {
	BookingID string       `json:"bookingId"`
	Shift     string       `json:"shift"`
	Owners    []ShiftOwner `json:"owners"`
}

const bookingShiftSQL = `CASE WHEN b.shift IS NOT NULL THEN b.shift::text
	WHEN EXTRACT(HOUR FROM (b."createdAt" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad')) < 15 THEN 'MORNING' ELSE 'EVENING' END`

func (r *CoordinationAlertRepository) BookingShift(bookingID string) (string, error) {
	var s string
	err := r.db.Get(&s, `SELECT `+bookingShiftSQL+` FROM "Booking" b WHERE b.id = $1`, bookingID)
	return s, err
}

func (r *CoordinationAlertRepository) ShiftOwners(shift string) ([]ShiftOwner, error) {
	out := []ShiftOwner{}
	err := r.db.Select(&out, `
		SELECT e.id, e.name,
		       (SELECT COUNT(*) FROM "CoordinationAlert" a WHERE a."coordinatorId" = e.id) AS strikes
		FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER', 'MONITOR')
		  AND COALESCE(e.shift::text, 'MORNING') = $1
		  AND EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
		              WHERE ep."employeeId" = e.id AND p.name = 'coordinator')
		ORDER BY e.name`, shift)
	return out, err
}

// PendingShifts فترة كل حجز ينتظر التثبيت وما تواصل وياه الإداري بعد.
func (r *CoordinationAlertRepository) PendingShifts() ([]BookingShiftInfo, error) {
	rows := []struct {
		ID    string `db:"id"`
		Shift string `db:"shift"`
	}{}
	if err := r.db.Select(&rows, `SELECT b.id, `+bookingShiftSQL+` AS shift FROM "Booking" b
		WHERE b.status = 'PENDING' AND b."archivedAt" IS NULL AND b."confirmationContactedAt" IS NULL`); err != nil {
		return nil, err
	}
	owners := map[string][]ShiftOwner{}
	out := make([]BookingShiftInfo, 0, len(rows))
	for _, x := range rows {
		if _, ok := owners[x.Shift]; !ok {
			o, _ := r.ShiftOwners(x.Shift)
			owners[x.Shift] = o
		}
		out = append(out, BookingShiftInfo{BookingID: x.ID, Shift: x.Shift, Owners: owners[x.Shift]})
	}
	return out, nil
}

// StrikesFor مجموع التقصيرات المسجّلة على موظف (كل الوقت).
func (r *CoordinationAlertRepository) StrikesFor(employeeID string) int {
	var n int
	_ = r.db.Get(&n, `SELECT COUNT(*) FROM "CoordinationAlert" WHERE "coordinatorId" = $1`, employeeID)
	return n
}

// KpiTaskCriterion نص بند «تنفيذ المهام» كما هو بالقاعدة (حتى رصيد النقاط يطابقه).
func (r *CoordinationAlertRepository) KpiTaskCriterion() string {
	var l string
	if err := r.db.Get(&l, `SELECT label FROM "KpiCriterion" WHERE label LIKE '%تنفيذ المهام%' OR label LIKE '%إتمام المهام%' OR label LIKE '%اتمام المهام%' ORDER BY "createdAt" LIMIT 1`); err != nil || l == "" {
		return "تنفيذ المهام الموكلة إليه"
	}
	return l
}
