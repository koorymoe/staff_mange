package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ عيون الرقابة — الكميات والسيارات وتقييمات الزبائن ═══
// طلب (ع) 10-05: «لازم اكو رقابه بكل مكان». كلها قواعد على الجداول الموجودة
// (بلا ذكاء اصطناعي = بلا كلفة). كل استعلام يرجع البنود الي تحتاج انتباه بس.

type MatrixWatchRepository struct{ db *sqlx.DB }

func NewMatrixWatchRepository(db *sqlx.DB) *MatrixWatchRepository {
	return &MatrixWatchRepository{db: db}
}

// ── الكميات ──

type StockRow struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	StockQty  float64    `db:"stockQty"`
	Used      float64    `db:"used"`
	CountedAt *time.Time `db:"countedAt"`
}

// NegativeStock مواد رصيدها المحسوب (الجرد − المصروف بالفواتير بعده) نزل تحت الصفر.
func (r *MatrixWatchRepository) NegativeStock() ([]StockRow, error) {
	rows := []StockRow{}
	err := r.db.Select(&rows, `
		SELECT m.id, m.name, m."stockQty", m."stockCountedAt" AS "countedAt",
		       COALESCE(sum(i.quantity) FILTER (WHERE li.id IS NOT NULL), 0)::float8 AS used
		FROM "Material" m
		LEFT JOIN "LeaderInvoiceMaterialItem" i ON i."materialId" = m.id
		LEFT JOIN "LeaderInvoice" li ON li.id = i."leaderInvoiceId" AND li."createdAt" > m."stockCountedAt"
		WHERE m."stockQty" IS NOT NULL AND m."stockCountedAt" IS NOT NULL
		GROUP BY m.id
		HAVING m."stockQty" - COALESCE(sum(i.quantity) FILTER (WHERE li.id IS NOT NULL), 0) < 0
		ORDER BY m."stockQty" - COALESCE(sum(i.quantity) FILTER (WHERE li.id IS NOT NULL), 0)`)
	return rows, err
}

// StaleCount مواد انصرفت بآخر ٣٠ يوم وجردها أقدم من ٣٠ يوم (أو ما انجردت أبداً).
func (r *MatrixWatchRepository) StaleCount() ([]StockRow, error) {
	rows := []StockRow{}
	err := r.db.Select(&rows, `
		SELECT m.id, m.name, COALESCE(m."stockQty", 0) AS "stockQty", m."stockCountedAt" AS "countedAt",
		       sum(i.quantity)::float8 AS used
		FROM "Material" m
		JOIN "LeaderInvoiceMaterialItem" i ON i."materialId" = m.id
		JOIN "LeaderInvoice" li ON li.id = i."leaderInvoiceId" AND li."createdAt" > now() - interval '30 days'
		WHERE m."stockCountedAt" IS NULL OR m."stockCountedAt" < now() - interval '30 days'
		GROUP BY m.id ORDER BY sum(i.quantity) DESC LIMIT 30`)
	return rows, err
}

type ToolOutRow struct {
	ID         string    `db:"id"`
	Tool       string    `db:"tool"`
	EmployeeID string    `db:"employeeId"`
	Employee   string    `db:"employee"`
	ApprovedAt time.Time `db:"approvedAt"`
}

// ToolsNotReturned أدوات انطلعت (معتمدة) وما رجعت من +٧ أيام. البدل (مفقود/تالف)
// ما ينحسب — هذا يبقى ويا الموظف.
func (r *MatrixWatchRepository) ToolsNotReturned() ([]ToolOutRow, error) {
	rows := []ToolOutRow{}
	err := r.db.Select(&rows, `
		SELECT q.id, t.name AS tool, q."employeeId", e.name AS employee, q."approvedAt"
		FROM "ToolRequest" q JOIN "OnDemandTool" t ON t.id = q."toolId" JOIN "Employee" e ON e.id = q."employeeId"
		WHERE q.status::text = 'APPROVED' AND q."returnedAt" IS NULL AND q."approvedAt" < now() - interval '7 days'
		  AND COALESCE(q."requestKind", '') NOT IN ('REPLACE_LOST', 'REPLACE_DAMAGED')
		ORDER BY q."approvedAt" LIMIT 50`)
	return rows, err
}

// ToolReturnedOrGone للتذكير «انحل؟».
func (r *MatrixWatchRepository) ToolReturned(requestID string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT NOT EXISTS (SELECT 1 FROM "ToolRequest" WHERE id = $1 AND status::text = 'APPROVED' AND "returnedAt" IS NULL)`, requestID)
	return ok
}

type ToolCountRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Total     int    `db:"totalQuantity"`
	Available int    `db:"availableQuantity"`
}

// ToolCountsWrong أداة المتوفر منها أكثر من الكلي أو أقل من صفر.
func (r *MatrixWatchRepository) ToolCountsWrong() ([]ToolCountRow, error) {
	rows := []ToolCountRow{}
	err := r.db.Select(&rows, `SELECT id, name, "totalQuantity", "availableQuantity" FROM "OnDemandTool"
		WHERE "availableQuantity" > "totalQuantity" OR "availableQuantity" < 0`)
	return rows, err
}

type ShortageRow struct {
	Source     string    `db:"source"`
	EmployeeID string    `db:"employeeId"`
	Employee   string    `db:"employee"`
	Missing    *string   `db:"missing"`
	BookingID  *string   `db:"bookingId"`
	Code       *string   `db:"code"`
	At         time.Time `db:"at"`
}

// ShortagesOpen نقص بالجرد ما انحل من +٣ أيام (قبل الحجز)، ونقص بعد الحجز (آخر ١٤ يوم).
func (r *MatrixWatchRepository) ShortagesOpen() ([]ShortageRow, error) {
	rows := []ShortageRow{}
	err := r.db.Select(&rows, `
		SELECT 'BEFORE' AS source, c."employeeId", e.name AS employee, c."missingItems" AS missing, c."bookingId", b.code, c."checkedAt" AS at
		FROM "InventoryCheck" c JOIN "Employee" e ON e.id = c."employeeId" LEFT JOIN "Booking" b ON b.id = c."bookingId"
		WHERE NOT c.complete AND NOT c.resolved AND c."checkedAt" < now() - interval '3 days' AND c."checkedAt" > now() - interval '60 days'
		UNION ALL
		SELECT 'AFTER', a."employeeId", e.name, a."missingItems", a."bookingId", b.code, a."checkedAt"
		FROM "BookingAfterInventory" a JOIN "Employee" e ON e.id = a."employeeId" JOIN "Booking" b ON b.id = a."bookingId"
		WHERE NOT a.complete AND a."checkedAt" > now() - interval '14 days'
		ORDER BY at DESC LIMIT 50`)
	return rows, err
}

type OddQtyRow struct {
	Material string    `db:"material"`
	Qty      float64   `db:"qty"`
	Median   float64   `db:"median"`
	Code     *string   `db:"code"`
	Booking  *string   `db:"bookingId"`
	Leader   string    `db:"leader"`
	LeaderID string    `db:"leaderId"`
	At       time.Time `db:"at"`
}

// OddInvoiceQty بنود فاتورة (آخر ٣٠ يوم) كميتها أكثر من ٣ أضعاف المعتاد لنفس المادة.
func (r *MatrixWatchRepository) OddInvoiceQty() ([]OddQtyRow, error) {
	rows := []OddQtyRow{}
	err := r.db.Select(&rows, `
		WITH med AS (
			SELECT "materialId", percentile_cont(0.5) WITHIN GROUP (ORDER BY quantity)::float8 AS median, count(*) AS n
			FROM "LeaderInvoiceMaterialItem" WHERE "materialId" IS NOT NULL AND "createdAt" > now() - interval '180 days'
			GROUP BY "materialId"
		)
		SELECT i.name AS material, i.quantity::float8 AS qty, med.median, b.code, li."bookingId", e.name AS leader, e.id AS "leaderId", li."createdAt" AS at
		FROM "LeaderInvoiceMaterialItem" i
		JOIN med ON med."materialId" = i."materialId" AND med.n >= 5 AND med.median > 0
		JOIN "LeaderInvoice" li ON li.id = i."leaderInvoiceId"
		JOIN "Employee" e ON e.id = li."employeeId"
		LEFT JOIN "Booking" b ON b.id = li."bookingId"
		WHERE li."createdAt" > now() - interval '30 days' AND i.quantity > med.median * 3
		ORDER BY i.quantity / med.median DESC LIMIT 30`)
	return rows, err
}

// ── السيارات ──

type FuelRow struct {
	Vehicle string    `db:"vehicle"`
	Plate   *string   `db:"plate"`
	Liters  float64   `db:"liters"`
	Km      float64   `db:"km"`
	Per100  float64   `db:"per100"`
	Avg100  float64   `db:"avg100"`
	By      *string   `db:"by"`
	ByID    *string   `db:"byId"`
	At      time.Time `db:"at"`
}

// FuelAbnormal تعبئات استهلاكها (لتر/١٠٠كم من التعبئة الي قبلها) أعلى من معدل
// نفس السيارة ×١٫٤ — آخر ٦٠ يوم، والسيارة عندها ٤ قراءات على الأقل.
func (r *MatrixWatchRepository) FuelAbnormal() ([]FuelRow, error) {
	rows := []FuelRow{}
	err := r.db.Select(&rows, `
		WITH f AS (
			SELECT l.*, l.odometer - lag(l.odometer) OVER (PARTITION BY l."vehicleId" ORDER BY l."performedAt") AS km
			FROM "VehicleLog" l WHERE l.type = 'FUEL' AND l.liters > 0 AND l.odometer IS NOT NULL
		), c AS (
			SELECT f.*, f.liters / f.km * 100 AS per100 FROM f WHERE f.km > 20
		), a AS (
			SELECT "vehicleId", avg(per100) AS avg100, count(*) AS n FROM c GROUP BY "vehicleId"
		)
		SELECT v.name AS vehicle, v."plateNumber" AS plate, c.liters, c.km::float8, c.per100::float8, a.avg100::float8,
		       e.name AS by, c."filledByEmployeeId" AS "byId", c."performedAt" AS at
		FROM c JOIN a ON a."vehicleId" = c."vehicleId" AND a.n >= 4 JOIN "Vehicle" v ON v.id = c."vehicleId"
		LEFT JOIN "Employee" e ON e.id = c."filledByEmployeeId"
		WHERE c."performedAt" > now() - interval '60 days' AND c.per100 > a.avg100 * 1.4
		ORDER BY c."performedAt" DESC LIMIT 30`)
	return rows, err
}

type MissionGapRow struct {
	Vehicle  string    `db:"vehicle"`
	Plate    *string   `db:"plate"`
	PrevEnd  int       `db:"prevEnd"`
	Start    int       `db:"start"`
	Driver   *string   `db:"driver"`
	DriverID *string   `db:"driverId"`
	At       time.Time `db:"at"`
}

// OdometerGaps العدّاد رجع لورا، أو +٥٠ كم مشتها السيارة بين مهمتين بلا مهمة مسجّلة.
func (r *MatrixWatchRepository) OdometerGaps() ([]MissionGapRow, error) {
	rows := []MissionGapRow{}
	err := r.db.Select(&rows, `
		WITH m AS (
			SELECT vm.*, lag(vm."endOdometer") OVER (PARTITION BY vm."vehicleId" ORDER BY vm."startedAt") AS "prevEnd"
			FROM "VehicleMission" vm WHERE vm."startOdometer" IS NOT NULL
		)
		SELECT v.name AS vehicle, v."plateNumber" AS plate, m."prevEnd", m."startOdometer" AS start, e.name AS driver, m."driverId", m."startedAt" AS at
		FROM m JOIN "Vehicle" v ON v.id = m."vehicleId" LEFT JOIN "Employee" e ON e.id = m."driverId"
		WHERE m."prevEnd" IS NOT NULL AND m."startedAt" > now() - interval '60 days'
		  AND (m."startOdometer" < m."prevEnd" OR m."startOdometer" - m."prevEnd" > 50)
		ORDER BY m."startedAt" DESC LIMIT 30`)
	return rows, err
}

type MissionRow struct {
	ID       string    `db:"id"`
	Vehicle  string    `db:"vehicle"`
	Plate    *string   `db:"plate"`
	Driver   *string   `db:"driver"`
	DriverID *string   `db:"driverId"`
	Purpose  *string   `db:"purpose"`
	At       time.Time `db:"at"`
}

// MissionsUnbooked مهمات (آخر ٣٠ يوم) ماكو حجز سيارة معتمد يغطّي وقتها (±ساعة).
func (r *MatrixWatchRepository) MissionsUnbooked() ([]MissionRow, error) {
	rows := []MissionRow{}
	err := r.db.Select(&rows, `
		SELECT vm.id, v.name AS vehicle, v."plateNumber" AS plate, e.name AS driver, vm."driverId", vm.purpose, vm."startedAt" AS at
		FROM "VehicleMission" vm JOIN "Vehicle" v ON v.id = vm."vehicleId" LEFT JOIN "Employee" e ON e.id = vm."driverId"
		WHERE vm."startedAt" > now() - interval '30 days'
		  AND NOT EXISTS (SELECT 1 FROM "VehicleBooking" bk WHERE bk."vehicleId" = vm."vehicleId" AND bk.status = 'APPROVED'
		                  AND vm."startedAt" BETWEEN bk."startAt" - interval '1 hour' AND bk."endAt" + interval '1 hour')
		ORDER BY vm."startedAt" DESC LIMIT 30`)
	return rows, err
}

// MissionsOpen مهمات مفتوحة من +٢٤ ساعة.
func (r *MatrixWatchRepository) MissionsOpen() ([]MissionRow, error) {
	rows := []MissionRow{}
	err := r.db.Select(&rows, `
		SELECT vm.id, v.name AS vehicle, v."plateNumber" AS plate, e.name AS driver, vm."driverId", vm.purpose, vm."startedAt" AS at
		FROM "VehicleMission" vm JOIN "Vehicle" v ON v.id = vm."vehicleId" LEFT JOIN "Employee" e ON e.id = vm."driverId"
		WHERE vm.status = 'IN_PROGRESS' AND vm."startedAt" < now() - interval '24 hours'
		ORDER BY vm."startedAt" LIMIT 30`)
	return rows, err
}

type MaintRow struct {
	Vehicle  string     `db:"vehicle"`
	Plate    *string    `db:"plate"`
	Type     string     `db:"type"`
	DueAt    *time.Time `db:"dueAt"`
	DueOdo   *int       `db:"dueOdo"`
	Odometer *int       `db:"odometer"`
}

// MaintenanceOverdue آخر سجل من كل نوع صيانة لكل سيارة وموعده الجاي فات (تاريخ أو عدّاد).
func (r *MatrixWatchRepository) MaintenanceOverdue() ([]MaintRow, error) {
	rows := []MaintRow{}
	err := r.db.Select(&rows, `
		SELECT v.name AS vehicle, v."plateNumber" AS plate, x.type, x."nextDueAt" AS "dueAt", x."nextDueOdometer" AS "dueOdo", v."currentOdometer" AS odometer
		FROM (SELECT DISTINCT ON ("vehicleId", type) * FROM "VehicleLog" WHERE type <> 'FUEL' ORDER BY "vehicleId", type, "performedAt" DESC) x
		JOIN "Vehicle" v ON v.id = x."vehicleId" AND v."isActive"
		WHERE (x."nextDueAt" IS NOT NULL AND x."nextDueAt" < now())
		   OR (x."nextDueOdometer" IS NOT NULL AND v."currentOdometer" IS NOT NULL AND v."currentOdometer" > x."nextDueOdometer")
		LIMIT 30`)
	return rows, err
}

type FuelNoProofRow struct {
	Vehicle   string    `db:"vehicle"`
	Plate     *string   `db:"plate"`
	By        *string   `db:"by"`
	ByID      *string   `db:"byId"`
	Cost      *float64  `db:"cost"`
	NoReceipt bool      `db:"noReceipt"`
	NoMission bool      `db:"noMission"`
	At        time.Time `db:"at"`
}

// FuelNoProof تعبئات (آخر ٣٠ يوم) بلا صورة وصل، أو الي عبّى ما عنده مهمة بالسيارة بنفس اليوم.
func (r *MatrixWatchRepository) FuelNoProof() ([]FuelNoProofRow, error) {
	rows := []FuelNoProofRow{}
	err := r.db.Select(&rows, `
		SELECT v.name AS vehicle, v."plateNumber" AS plate, e.name AS by, l."filledByEmployeeId" AS "byId", l.cost,
		       (COALESCE(l."receiptPhotoBase64", '') = '') AS "noReceipt",
		       (l."filledByEmployeeId" IS NOT NULL AND NOT EXISTS (SELECT 1 FROM "VehicleMission" vm WHERE vm."vehicleId" = l."vehicleId"
		          AND vm."driverId" = l."filledByEmployeeId" AND baghdad_date(vm."startedAt") = baghdad_date(l."performedAt"))) AS "noMission",
		       l."performedAt" AS at
		FROM "VehicleLog" l JOIN "Vehicle" v ON v.id = l."vehicleId" LEFT JOIN "Employee" e ON e.id = l."filledByEmployeeId"
		WHERE l.type = 'FUEL' AND l."performedAt" > now() - interval '30 days'
		  AND (COALESCE(l."receiptPhotoBase64", '') = ''
		       OR (l."filledByEmployeeId" IS NOT NULL AND NOT EXISTS (SELECT 1 FROM "VehicleMission" vm WHERE vm."vehicleId" = l."vehicleId"
		          AND vm."driverId" = l."filledByEmployeeId" AND baghdad_date(vm."startedAt") = baghdad_date(l."performedAt"))))
		ORDER BY l."performedAt" DESC LIMIT 30`)
	return rows, err
}

// ── تقييمات الزبائن ──

type CustomerScoreRow struct {
	EmployeeID   string   `db:"employeeId" json:"employeeId"`
	Name         string   `db:"name" json:"name"`
	FollowUps    int      `db:"followUps" json:"followUps"`
	Happy        int      `db:"happy" json:"happy"`
	Unhappy      int      `db:"unhappy" json:"unhappy"`
	Complaints   int      `db:"complaints" json:"complaints"`
	Complaints30 int      `db:"complaints30" json:"complaints30"`
	AvgRating    *float64 `db:"avgRating" json:"avgRating"`
	PrevHappyPct *float64 `db:"prevHappyPct" json:"prevHappyPct"`
}

// CustomerScores لكل موظف بالكادر (ليدر أو فني): متابعات الجودة لحجوزاته آخر ٦٠ يوم،
// كم زبون راضي وكم عنده ملاحظة، والشكاوى عليه، ومعدل تقييم الزبون بالشكاوى.
func (r *MatrixWatchRepository) CustomerScores() ([]CustomerScoreRow, error) {
	rows := []CustomerScoreRow{}
	err := r.db.Select(&rows, `
		WITH crew AS (
			SELECT b.id AS "bookingId", b."projectSupervisorId" AS "employeeId" FROM "Booking" b WHERE b."projectSupervisorId" IS NOT NULL
			UNION SELECT a."bookingId", a."employeeId" FROM "BookingAssignment" a
		), fu AS (
			SELECT crew."employeeId", q."contactedAt",
			       (q.status IN ('CONTACTED_OK') OR q."reportType" = 'POSITIVE') AS happy,
			       (q.status IN ('CONTACTED_ISSUE', 'CONVERTED') OR q."reportType" = 'NEGATIVE' OR q."inspectionResult" = 'CUSTOMER_RIGHT') AS unhappy
			FROM "QualityFollowUp" q JOIN crew ON crew."bookingId" = q."bookingId"
			WHERE q."contactedAt" > now() - interval '120 days'
		), cur AS (
			SELECT "employeeId", count(*)::int AS "followUps", count(*) FILTER (WHERE happy)::int AS happy, count(*) FILTER (WHERE unhappy)::int AS unhappy
			FROM fu WHERE "contactedAt" > now() - interval '60 days' GROUP BY 1
		), prev AS (
			SELECT "employeeId", 100.0 * count(*) FILTER (WHERE happy) / NULLIF(count(*), 0) AS pct
			FROM fu WHERE "contactedAt" <= now() - interval '60 days' GROUP BY 1
		), comp AS (
			SELECT "relatedEmployeeId" AS "employeeId", count(*)::int AS complaints,
			       count(*) FILTER (WHERE "createdAt" > now() - interval '30 days')::int AS complaints30,
			       avg("customerRating")::float8 AS "avgRating"
			FROM "Complaint" WHERE "relatedEmployeeId" IS NOT NULL AND "createdAt" > now() - interval '60 days' GROUP BY 1
		)
		SELECT e.id AS "employeeId", e.name, COALESCE(cur."followUps", 0) AS "followUps", COALESCE(cur.happy, 0) AS happy,
		       COALESCE(cur.unhappy, 0) AS unhappy, COALESCE(comp.complaints, 0) AS complaints, COALESCE(comp.complaints30, 0) AS complaints30,
		       comp."avgRating", prev.pct::float8 AS "prevHappyPct"
		FROM "Employee" e
		LEFT JOIN cur ON cur."employeeId" = e.id LEFT JOIN comp ON comp."employeeId" = e.id LEFT JOIN prev ON prev."employeeId" = e.id
		WHERE e.status = 'ACTIVE' AND (cur."employeeId" IS NOT NULL OR comp."employeeId" IS NOT NULL)
		ORDER BY COALESCE(comp.complaints, 0) DESC, COALESCE(cur.unhappy, 0) DESC`)
	return rows, err
}

// ── الآيتي ──

type ItRow struct {
	ID       string     `db:"id"`
	Name     string     `db:"name"`
	Count    int        `db:"n"`
	Cost     *float64   `db:"cost"`
	Holder   *string    `db:"holder"`
	HolderID *string    `db:"holderId"`
	At       *time.Time `db:"at"`
}

// ItRepeatRepairs جهاز انصلح/انصان ٣ مرات أو أكثر بآخر ٩٠ يوم.
func (r *MatrixWatchRepository) ItRepeatRepairs() ([]ItRow, error) {
	rows := []ItRow{}
	err := r.db.Select(&rows, `
		SELECT a.id, a.name, count(l.id)::int AS n, sum(l.cost)::float8 AS cost, e.name AS holder, a."assignedEmployeeId" AS "holderId", max(l."createdAt") AS at
		FROM "ItAsset" a JOIN "ItAssetLog" l ON l."assetId" = a.id AND l.kind IN ('REPAIR', 'MAINTENANCE') AND l."createdAt" > now() - interval '90 days'
		LEFT JOIN "Employee" e ON e.id = a."assignedEmployeeId"
		GROUP BY a.id, e.name HAVING count(l.id) >= 3 ORDER BY count(l.id) DESC LIMIT 30`)
	return rows, err
}

// ItInUseNoHolder جهاز «مستعمل» بلا موظف مسؤول عنه.
func (r *MatrixWatchRepository) ItInUseNoHolder() ([]ItRow, error) {
	rows := []ItRow{}
	err := r.db.Select(&rows, `SELECT id, name, 0 AS n, NULL::float8 AS cost, NULL AS holder, NULL AS "holderId", "updatedAt" AS at
		FROM "ItAsset" WHERE status IN ('IN_USE', 'ACTIVE') AND "assignedEmployeeId" IS NULL LIMIT 30`)
	return rows, err
}

type TicketRow struct {
	ID       string    `db:"id"`
	Device   *string   `db:"device"`
	Serial   *string   `db:"serial"`
	Employee *string   `db:"employee"`
	EmpID    *string   `db:"employeeId"`
	Received time.Time `db:"received"`
	Count    int       `db:"n"`
}

// RepairTicketsLate أجهزة زبائن انستلمت للتصليح وما انسلّمت من +٧ أيام.
func (r *MatrixWatchRepository) RepairTicketsLate() ([]TicketRow, error) {
	rows := []TicketRow{}
	err := r.db.Select(&rows, `SELECT t.id, t."deviceTypeName" AS device, t."deviceSerial" AS serial, e.name AS employee, t."employeeId", t."receivedAt" AS received, 0 AS n
		FROM "DeviceMaintenanceTicket" t LEFT JOIN "Employee" e ON e.id = t."employeeId"
		WHERE t."receivedAt" IS NOT NULL AND t."deliveredAt" IS NULL AND t."receivedAt" < now() - interval '7 days'
		ORDER BY t."receivedAt" LIMIT 30`)
	return rows, err
}

// RepairRepeats نفس الجهاز (رقم تسلسلي) رجع للتصليح مرتين أو أكثر بآخر ٩٠ يوم.
func (r *MatrixWatchRepository) RepairRepeats() ([]TicketRow, error) {
	rows := []TicketRow{}
	err := r.db.Select(&rows, `SELECT max(t.id) AS id, max(t."deviceTypeName") AS device, t."deviceSerial" AS serial, NULL AS employee, NULL AS "employeeId",
		       max(t."receivedAt") AS received, count(*)::int AS n
		FROM "DeviceMaintenanceTicket" t
		WHERE COALESCE(t."deviceSerial", '') <> '' AND t."receivedAt" > now() - interval '90 days'
		GROUP BY t."deviceSerial" HAVING count(*) >= 2 ORDER BY count(*) DESC LIMIT 30`)
	return rows, err
}

// ── الطلبات المعلّقة (كل سلسلة بيها «ينتظر قرار») ──

type PendingRow struct {
	Kind    string    `db:"kind"`
	ID      string    `db:"id"`
	Title   string    `db:"title"`
	Who     *string   `db:"who"`
	WhoID   *string   `db:"whoId"`
	Since   time.Time `db:"since"`
	Decider string    `db:"decider"`
}

// PendingRequests طلبات تنتظر قرار أكثر من المعتاد: إجازات، كتب، مشتريات،
// طلبات منتجات، أجهزة جي بي اس وتجديدها، شكاوى الزبائن.
func (r *MatrixWatchRepository) PendingRequests() ([]PendingRow, error) {
	rows := []PendingRow{}
	err := r.db.Select(&rows, `
		SELECT 'LEAVE' AS kind, l.id, 'طلب إجازة' AS title, e.name AS who, l."employeeId" AS "whoId", l."createdAt" AS since, 'الموارد البشرية / المدير' AS decider
		  FROM "LeaveRequest" l JOIN "Employee" e ON e.id = l."employeeId"
		 WHERE l.status IN ('PENDING', 'PRELIMINARY') AND l."createdAt" < now() - interval '2 days'
		UNION ALL
		SELECT 'LETTER', k.id, 'كتاب: ' || k.subject, e.name, k."employeeId", k."createdAt", 'المدير'
		  FROM "EmployeeLetter" k JOIN "Employee" e ON e.id = k."employeeId"
		 WHERE k.status = 'PENDING' AND k."createdAt" < now() - interval '3 days'
		UNION ALL
		SELECT 'PROCUREMENT', p.id, 'طلب مشتريات ' || p.code, e.name, p."requestedById", p."createdAt", 'المشتريات / المحاسب'
		  FROM "ProcurementRequest" p LEFT JOIN "Employee" e ON e.id = p."requestedById"
		 WHERE p.status::text IN ('PENDING', 'IN_PROGRESS') AND p."createdAt" < now() - interval '3 days'
		UNION ALL
		SELECT 'PRODUCT', p.id, 'طلب منتج: ' || p."productName", e.name, p."requestedById", p."createdAt", 'المدير'
		  FROM "ProductRequest" p LEFT JOIN "Employee" e ON e.id = p."requestedById"
		 WHERE p.status = 'PENDING' AND p."createdAt" < now() - interval '5 days'
		UNION ALL
		SELECT 'GPS_DEVICE', g.id, 'طلب جهاز جي بي اس', e.name, g."employeeId", g."createdAt", 'مسؤول الجي بي اس'
		  FROM "GpsDeviceRequest" g LEFT JOIN "Employee" e ON e.id = g."employeeId"
		 WHERE g.status::text IN ('PENDING', 'APPROVED') AND g."createdAt" < now() - interval '3 days'
		UNION ALL
		SELECT 'GPS_RENEWAL', g.id, 'طلب تجديد اشتراك جي بي اس', e.name, g."employeeId", g."createdAt", 'مسؤول الجي بي اس'
		  FROM "GpsRenewalRequest" g LEFT JOIN "Employee" e ON e.id = g."employeeId"
		 WHERE g.status::text = 'PENDING' AND g."createdAt" < now() - interval '2 days'
		UNION ALL
		SELECT 'COMPLAINT', c.id, 'شكوى زبون', e.name, c."assignedToEmployeeId", c."createdAt", 'الجودة / المكلّف بالشكوى'
		  FROM "Complaint" c LEFT JOIN "Employee" e ON e.id = c."assignedToEmployeeId"
		 WHERE c.status::text IN ('NEW', 'IN_PROGRESS') AND c."createdAt" < now() - interval '5 days'
		UNION ALL
		SELECT 'MEDIA', m.id, 'تصوير مشروع ' || p.name || ' (الإعلام)', NULL, NULL, m."createdAt", 'الإعلام والعلاقات العامة'
		  FROM "MediaBrief" m JOIN "Project" p ON p.id = m."projectId"
		 WHERE m.status = 'NEW' AND m."createdAt" < now() - interval '3 days'
		ORDER BY since LIMIT 80`)
	return rows, err
}

// ═══ عين الإيرادات (قرار (ع) 10-06): «ماتركس راح يكون العروق مال الشجره» ═══

// MoneyGap حجز منجز بيه مشكلة فلوس.
type MoneyGap struct {
	BookingID   string    `db:"bookingId"`
	Code        string    `db:"code"`
	CompletedAt time.Time `db:"completedAt"`
	Collected   *float64  `db:"collected"`
	InvoiceNet  *float64  `db:"invoiceNet"`
	LeaderID    *string   `db:"leaderId"`
	LeaderName  *string   `db:"leaderName"`
}

// MoneyGaps حجوزات منجزة بآخر ٦٠ يوم (مو قديمة، مو داخلية): بلا مبلغ ولا فاتورة
// من +٢٤ ساعة، أو المحصّل يختلف عن صافي الفاتورة. حجوزات المشاريع المدفوعة
// كمشروع مستثناة (فلوسها بالدفعات).
func (r *MatrixWatchRepository) MoneyGaps() ([]MoneyGap, error) {
	rows := []MoneyGap{}
	err := r.db.Select(&rows, `
		SELECT b.id AS "bookingId", b.code, b."completedAt", b."amountCollected"::float8 AS collected,
		       li."netTotal"::float8 AS "invoiceNet",
		       COALESCE(b."projectSupervisorId", ld."employeeId") AS "leaderId", le.name AS "leaderName"
		FROM "Booking" b
		LEFT JOIN LATERAL (SELECT "netTotal" FROM "LeaderInvoice" WHERE "bookingId" = b.id ORDER BY "createdAt" DESC LIMIT 1) li ON true
		LEFT JOIN LATERAL (SELECT ba."employeeId" FROM "BookingAssignment" ba JOIN "Employee" x ON x.id = ba."employeeId"
		                   WHERE ba."bookingId" = b.id AND x."isLeader" LIMIT 1) ld ON true
		LEFT JOIN "Employee" le ON le.id = COALESCE(b."projectSupervisorId", ld."employeeId")
		WHERE b.status = 'COMPLETED' AND b."completedAt" IS NOT NULL AND b."archivedAt" IS NULL
		  AND b."bookingType" IS DISTINCT FROM 'INTERNAL' AND upper(b.code) NOT LIKE 'OLD%'
		  AND b."completedAt" >= now() - interval '60 days'
		  AND NOT EXISTS (SELECT 1 FROM "Project" pj JOIN "ProjectPayment" pp ON pp."projectId" = pj.id AND pp."cancelledAt" IS NULL WHERE pj."bookingId" = b.id)
		  AND (
		    (b."completedAt" < now() - interval '24 hours' AND COALESCE(b."amountCollected", 0) = 0 AND li."netTotal" IS NULL)
		    OR (b."amountCollected" IS NOT NULL AND li."netTotal" IS NOT NULL AND abs(b."amountCollected" - li."netTotal") >= 1000)
		  )
		ORDER BY b."completedAt" DESC LIMIT 200`)
	return rows, err
}

// AutoCheckouts كم انصراف تلقائي لكل موظف بآخر ٧ أيام (٢ وأكثر).
type AutoCheckoutRow struct {
	EmployeeID string `db:"employeeId"`
	Name       string `db:"name"`
	N          int    `db:"n"`
}

func (r *MatrixWatchRepository) AutoCheckouts() ([]AutoCheckoutRow, error) {
	rows := []AutoCheckoutRow{}
	err := r.db.Select(&rows, `SELECT a."employeeId", e.name, count(*)::int AS n FROM "AttendanceAuto" a
		JOIN "Employee" e ON e.id = a."employeeId" WHERE a."createdAt" >= now() - interval '7 days'
		GROUP BY 1, 2 HAVING count(*) >= 2 ORDER BY 3 DESC`)
	return rows, err
}
