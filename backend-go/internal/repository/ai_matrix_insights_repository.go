package repository

import (
	"time"

	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// ═══ ماتركس — الكادر الأنسب، منحنى الموظف الجديد، المخزون، الاستبدال، التقرير الأسبوعي ═══
//
// قراءات بس من أعمدة موجودة — بلا جداول جديدة وبلا نموذج خارجي.
// الأعمدة «timestamp without time zone» مخزونة UTC، فنحوّلها
// لـtimestamptz بـ AT TIME ZONE 'UTC' قبل المقارنة بوقت Go.

// ── ١) توصية الكادر ──

// CrewBookingInfo خدمة الحجز ويومه (بغداد) — اليوم فارغ الموعد = اليوم.
type CrewBookingInfo struct {
	ServiceID   string `db:"serviceId"`
	ServiceName string `db:"serviceName"`
	Day         string `db:"day"`
}

func (r *AiRepository) CrewBookingInfo(bookingID string) (*CrewBookingInfo, error) {
	var out CrewBookingInfo
	err := r.db.Get(&out, `
		SELECT COALESCE(b."serviceId", '') AS "serviceId", COALESCE(s.name, '') AS "serviceName",
		       to_char(COALESCE(baghdad_date(b."scheduledAt"), baghdad_date(now())), 'YYYY-MM-DD') AS day
		FROM "Booking" b LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.id = $1`, bookingID)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CrewCandidateRow موظف ميداني بالدوام مع أرقامه لهالخدمة.
type CrewCandidateRow struct {
	EmployeeID       string `db:"employeeId"`
	Name             string `db:"name"`
	IsLeader         bool   `db:"isLeader"`
	HasSkill         bool   `db:"hasSkill"`
	ServiceHasSkills bool   `db:"serviceHasSkills"`
	DoneCount        int    `db:"doneCount"`
	ProblemCount     int    `db:"problemCount"`
	DayLoad          int    `db:"dayLoad"`
}

// CrewCandidates ليدرية وفنيين ACTIVE وبالدوام (onDuty). DoneCount =
// حجوزات منجزة من نفس الخدمة بآخر ٣٦٥ يوم وهو بكادرها؛ ProblemCount =
// منها الي عليها إشارة توقف شغل أو شكوى؛ DayLoad = حجوزاته بنفس يوم
// الحجز (غير هذا الحجز).
func (r *AiRepository) CrewCandidates(bookingID, serviceID, day string) ([]CrewCandidateRow, error) {
	rows := []CrewCandidateRow{}
	err := r.db.Select(&rows, `
		WITH done AS (
			SELECT b.id, ba."employeeId" AS eid
			FROM "Booking" b JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
			WHERE b.status = 'COMPLETED' AND b."serviceId" = $2 AND b."serviceId" <> ''
			  AND b."archivedAt" IS NULL AND b."settledLegacyAt" IS NULL
			  AND b."completedAt" >= (now() AT TIME ZONE 'UTC') - interval '365 days'
		)
		SELECT e.id AS "employeeId", e.name, e."isLeader",
		       EXISTS (SELECT 1 FROM "EmployeeSkill" es JOIN "Skill" s ON s.id = es."skillId"
		               WHERE es."employeeId" = e.id AND es."canPerform" AND s."serviceId" = $2) AS "hasSkill",
		       EXISTS (SELECT 1 FROM "Skill" s WHERE s."serviceId" = $2) AS "serviceHasSkills",
		       (SELECT COUNT(*) FROM done d WHERE d.eid = e.id) AS "doneCount",
		       (SELECT COUNT(*) FROM done d WHERE d.eid = e.id AND (
		           EXISTS (SELECT 1 FROM "AiSignal" sg WHERE sg."entityType" = 'BOOKING' AND sg."entityId" = d.id AND sg.kind = 'WORK_STOPPED')
		           OR EXISTS (SELECT 1 FROM "Complaint" c WHERE c."bookingId" = d.id))) AS "problemCount",
		       (SELECT COUNT(DISTINCT b2.id) FROM "Booking" b2
		        JOIN "BookingAssignment" ba2 ON ba2."bookingId" = b2.id
		        WHERE ba2."employeeId" = e.id AND b2.id <> $1 AND b2."archivedAt" IS NULL
		          AND b2.status IN ('PENDING', 'CONFIRMED', 'IN_PROGRESS', 'COMPLETED')
		          AND baghdad_date(b2."scheduledAt") = $3::date) AS "dayLoad"
		FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND e."onDuty" = true AND (e."isLeader" = true OR e.role = 'TECHNICIAN')
		ORDER BY e.name`, bookingID, serviceID, day)
	return rows, err
}

// ── ٧) + ١٢) أحداث الموظفين ──

// NewFieldEmployee موظف ميداني بأول ٩٠ يوم.
type NewFieldEmployee struct {
	ID       string `db:"id"`
	Name     string `db:"name"`
	IsLeader bool   `db:"isLeader"`
	Start    string `db:"start"`
}

// NewFieldEmployees فنيين/ليدرية ACTIVE بداية عملهم (hireDate وإلا
// يوم إنشاء الحساب) خلال آخر ٩٠ يوم.
func (r *AiRepository) NewFieldEmployees() ([]NewFieldEmployee, error) {
	rows := []NewFieldEmployee{}
	err := r.db.Select(&rows, `
		SELECT e.id, e.name, e."isLeader",
		       to_char(COALESCE(e."hireDate", baghdad_date(e."createdAt")), 'YYYY-MM-DD') AS start
		FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND (e."isLeader" = true OR e.role = 'TECHNICIAN')
		  AND COALESCE(e."hireDate", baghdad_date(e."createdAt")) >= baghdad_date(now()) - 90
		ORDER BY start DESC, e.name`)
	return rows, err
}

// EmployeeEvent حدث واحد: DONE حجز منجز بكادره، STOP إشارة توقف شغل
// باسمه، COMPLAINT شكوى مربوطة بيه. PaperDue/PaperOK لحجوزات DONE
// الي هو ليدرها وخلصت من أكثر من ٤٨ ساعة.
type EmployeeEvent struct {
	EmployeeID string    `db:"eid"`
	Name       string    `db:"name"`
	Kind       string    `db:"kind"`
	At         time.Time `db:"at"`
	PaperDue   bool      `db:"paperDue"`
	PaperOK    bool      `db:"paperOk"`
}

// EmployeeEvents أحداث بين from وto. ids فارغة = كل الموظفين.
func (r *AiRepository) EmployeeEvents(ids []string, from, to time.Time) ([]EmployeeEvent, error) {
	rows := []EmployeeEvent{}
	err := r.db.Select(&rows, `
		SELECT x.eid, e.name, x.kind, x.at, x."paperDue", x."paperOk" FROM (
			SELECT ba."employeeId" AS eid, 'DONE' AS kind, (b."completedAt" AT TIME ZONE 'UTC') AS at,
			       (le."isLeader" AND b."completedAt" < (now() AT TIME ZONE 'UTC') - interval '48 hours') AS "paperDue",
			       `+paperworkDoneSQL+` AS "paperOk"
			FROM "Booking" b
			JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
			JOIN "Employee" le ON le.id = ba."employeeId"
			WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL AND b."settledLegacyAt" IS NULL
			  AND NOT (`+isLegacyImportSQL+`)
			  AND (b."completedAt" AT TIME ZONE 'UTC') >= $2 AND (b."completedAt" AT TIME ZONE 'UTC') < $3
			UNION ALL
			SELECT s."employeeId", 'STOP', s."occurredAt", false, false
			FROM "AiSignal" s
			WHERE s.kind = 'WORK_STOPPED' AND s."employeeId" IS NOT NULL
			  AND s."occurredAt" >= $2 AND s."occurredAt" < $3
			UNION ALL
			SELECT c."relatedEmployeeId", 'COMPLAINT', (c."createdAt" AT TIME ZONE 'UTC'), false, false
			FROM "Complaint" c
			WHERE c."relatedEmployeeId" IS NOT NULL
			  AND (c."createdAt" AT TIME ZONE 'UTC') >= $2 AND (c."createdAt" AT TIME ZONE 'UTC') < $3
		) x JOIN "Employee" e ON e.id = x.eid
		WHERE (cardinality($1::text[]) = 0 OR x.eid = ANY($1))`, pq.Array(ids), from, to)
	return rows, err
}

// ── ١٢) أرقام الأسبوع ──

// WeekTotals أرقام الشركة بنافذة زمنية.
type WeekTotals struct {
	Completed     int     `db:"completed" json:"completed"`
	Revenue       float64 `db:"revenue" json:"revenue"`
	Complaints    int     `db:"complaints" json:"complaints"`
	WorkStops     int     `db:"workStops" json:"workStops"`
	LatePaperwork int     `db:"latePaperwork" json:"latePaperwork"`
}

// WeekTotals: منجز ومحصّل (amountCollected للحجوزات المنجزة بالنافذة)،
// شكاوى انفتحت، إشارات توقف، ومنجز بالنافذة صار له +٤٨ ساعة وورقه ناقص.
func (r *AiRepository) WeekTotals(from, to time.Time) (*WeekTotals, error) {
	var out WeekTotals
	err := r.db.Get(&out, `
		SELECT
		  (SELECT COUNT(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL
		     AND NOT (`+isLegacyImportSQL+`) AND b."settledLegacyAt" IS NULL
		     AND (b."completedAt" AT TIME ZONE 'UTC') >= $1 AND (b."completedAt" AT TIME ZONE 'UTC') < $2) AS completed,
		  (SELECT COALESCE(SUM(b."amountCollected"), 0) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL
		     AND NOT (`+isLegacyImportSQL+`) AND b."settledLegacyAt" IS NULL
		     AND (b."completedAt" AT TIME ZONE 'UTC') >= $1 AND (b."completedAt" AT TIME ZONE 'UTC') < $2) AS revenue,
		  (SELECT COUNT(*) FROM "Complaint" c
		     WHERE (c."createdAt" AT TIME ZONE 'UTC') >= $1 AND (c."createdAt" AT TIME ZONE 'UTC') < $2) AS complaints,
		  (SELECT COUNT(*) FROM "AiSignal" s WHERE s.kind = 'WORK_STOPPED'
		     AND s."occurredAt" >= $1 AND s."occurredAt" < $2) AS "workStops",
		  (SELECT COUNT(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL
		     AND b."settledLegacyAt" IS NULL`+NotDeletePendingSQL("b")+`
		     AND (b."completedAt" AT TIME ZONE 'UTC') >= $1 AND (b."completedAt" AT TIME ZONE 'UTC') < $2
		     AND b."completedAt" < (now() AT TIME ZONE 'UTC') - interval '48 hours'
		     AND NOT `+paperworkDoneSQL+`) AS "latePaperwork"`, from, to)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// OpenDecisions قرارات تنتظر الإدارة هسه.
type OpenDecisions struct {
	PendingLeaves      int `db:"pendingLeaves" json:"pendingLeaves"`
	DeleteRequests     int `db:"deleteRequests" json:"deleteRequests"`
	UnreviewedVerdicts int `db:"unreviewedVerdicts" json:"unreviewedVerdicts"`
}

func (r *AiRepository) OpenDecisions() (*OpenDecisions, error) {
	var out OpenDecisions
	err := r.db.Get(&out, `
		SELECT
		  (SELECT COUNT(*) FROM "LeaveRequest" WHERE status = ANY($1)) AS "pendingLeaves",
		  (SELECT COUNT(*) FROM "BookingDeleteRequest" WHERE status = 'PENDING') AS "deleteRequests",
		  (SELECT COUNT(*) FROM "MonitorReview" WHERE "entityType" = 'AI_VERDICT' AND status = 'PENDING') AS "unreviewedVerdicts"`,
		pq.Array(model.LeaveOpenStatuses()))
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ── ١٠) استهلاك المواد ──

// MaterialConsumption استهلاك مادة من فواتير الليدرية (غير الملغاة).
type MaterialConsumption struct {
	Key      string  `db:"key"`
	Name     string  `db:"name"`
	Usages   int     `db:"usages"`
	Quantity float64 `db:"quantity"`
}

// MaterialConsumptionSince المواد المستعملة ≥minUsages مرة من since.
func (r *AiRepository) MaterialConsumptionSince(since time.Time, minUsages int) ([]MaterialConsumption, error) {
	rows := []MaterialConsumption{}
	err := r.db.Select(&rows, `
		SELECT COALESCE(i."materialId", 'name:' || i.name) AS key,
		       COALESCE(MAX(m.name), MAX(i.name)) AS name,
		       COUNT(*) AS usages, COALESCE(SUM(i.quantity), 0)::float8 AS quantity
		FROM "LeaderInvoiceMaterialItem" i
		JOIN "LeaderInvoice" li ON li.id = i."leaderInvoiceId"
		LEFT JOIN "Material" m ON m.id = i."materialId"
		WHERE li."revokedAt" IS NULL AND i.quantity > 0
		  AND (li."createdAt" AT TIME ZONE 'UTC') >= $1
		GROUP BY 1
		HAVING COUNT(*) >= $2
		ORDER BY quantity DESC
		LIMIT 30`, since, minUsages)
	return rows, err
}

// ── ١١) الاستبدال ──

// ItRepairHeavy جهاز IT عليه تصليحات/صيانات كثيرة.
type ItRepairHeavy struct {
	ID          string  `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	Kind        string  `db:"kind" json:"kind"`
	Status      string  `db:"status" json:"status"`
	RepairCount int     `db:"repairCount" json:"repairCount"`
	RepairCost  float64 `db:"repairCost" json:"repairCost"`
}

// ItRepairCounts أجهزة غير متقاعدة عليها ≥minCount سجل REPAIR/MAINTENANCE بآخر days يوم.
func (r *AiRepository) ItRepairCounts(days, minCount int) ([]ItRepairHeavy, error) {
	rows := []ItRepairHeavy{}
	err := r.db.Select(&rows, `
		SELECT a.id, a.name, a.kind, a.status, COUNT(l.id) AS "repairCount", COALESCE(SUM(l.cost), 0) AS "repairCost"
		FROM "ItAsset" a
		JOIN "ItAssetLog" l ON l."assetId" = a.id AND l.kind IN ('REPAIR', 'MAINTENANCE')
		     AND l."createdAt" >= now() - make_interval(days => $1)
		WHERE a.status <> 'RETIRED'
		GROUP BY a.id
		HAVING COUNT(l.id) >= $2
		ORDER BY COUNT(l.id) DESC, a.name`, days, minCount)
	return rows, err
}

// VehicleCostRow كلفة صيانة وحوادث مركبة فعّالة.
type VehicleCostRow struct {
	ID            string  `db:"id"`
	Name          string  `db:"name"`
	PlateNumber   string  `db:"plateNumber"`
	MaintCost12m  float64 `db:"maintCost12m"`
	IncidentCost  float64 `db:"incidentCost12m"`
	Incidents180d int     `db:"incidents180d"`
}

// VehicleCosts كل المركبات الفعّالة: صيانة (كل أنواع السجل عدا الوقود
// والغسيل) + كلفة الحوادث (كلفة التصليح وإلا الكلفة) بآخر ١٢ شهر، وعدد
// الحوادث بآخر ١٨٠ يوم.
func (r *AiRepository) VehicleCosts() ([]VehicleCostRow, error) {
	rows := []VehicleCostRow{}
	err := r.db.Select(&rows, `
		SELECT v.id, v.name, COALESCE(v."plateNumber", '') AS "plateNumber",
		  COALESCE((SELECT SUM(l.cost) FROM "VehicleLog" l WHERE l."vehicleId" = v.id
		     AND l.type NOT IN ('FUEL', 'CLEANING')
		     AND l."performedAt" >= (now() AT TIME ZONE 'UTC') - interval '365 days'), 0) AS "maintCost12m",
		  COALESCE((SELECT SUM(COALESCE(i."repairCost", i.cost, 0)) FROM "VehicleIncident" i WHERE i."vehicleId" = v.id
		     AND i."createdAt" >= (now() AT TIME ZONE 'UTC') - interval '365 days'), 0) AS "incidentCost12m",
		  (SELECT COUNT(*) FROM "VehicleIncident" i WHERE i."vehicleId" = v.id
		     AND i."createdAt" >= (now() AT TIME ZONE 'UTC') - interval '180 days') AS "incidents180d"
		FROM "Vehicle" v
		WHERE v."isActive" = true
		ORDER BY v.name`)
	return rows, err
}
