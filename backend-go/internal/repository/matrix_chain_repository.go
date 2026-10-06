package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ═══ سلسلة الحجز — الحقائق الخام ═══
//
// صف واحد لكل حجز يجمع أوقات كل محطة ومنو سوّاها، من الجداول الموجودة
// (الحجز، التكليف، المهمة، الفاتورة، التقرير، الجرد، الجودة، المراقب،
// والتقييم الجديد). الحكم (بوقتها/متأخرة/ما صارت) ينحسب بالخدمة.
//
// ⚠️ الحجوزات القديمة (OLD…) والملغية برّا: ما مرّت بالنظام فقياسها كذب.
type ChainFacts struct {
	ID            string     `db:"id"`
	Code          string     `db:"code"`
	Status        string     `db:"status"`
	BookingType   string     `db:"bookingType"`
	ServiceID     *string    `db:"serviceId"`
	ServiceName   *string    `db:"serviceName"`
	Solo          bool       `db:"solo"`
	CreatedAt     time.Time  `db:"createdAt"`
	CreatedByID   *string    `db:"createdById"`
	CreatedByName *string    `db:"createdByName"`
	HasAddress    bool       `db:"hasAddress"`
	HasLocation   bool       `db:"hasLocation"`
	NameWords     int        `db:"nameWords"`
	Duplicate     bool       `db:"duplicate"`
	ContactedAt   *time.Time `db:"contactedAt"`
	ContactedByID *string    `db:"contactedById"`
	ContactedBy   *string    `db:"contactedBy"`
	ConfirmedAt   *time.Time `db:"confirmedAt"`
	ConfirmedByID *string    `db:"confirmedById"`
	ConfirmedBy   *string    `db:"confirmedBy"`
	ScheduledAt   *time.Time `db:"scheduledAt"`
	// حجز محوّل لإدارة المشاريع: ينقفل عندهم لحد «البدء بالتنفيذ».
	ToProjects    bool       `db:"toProjects"`
	ProjectExecAt *time.Time `db:"projectExecAt"`
	ProjectStage  *string    `db:"projectStage"`
	PostponeCount int        `db:"postponeCount"`
	ScheduleMoves int        `db:"scheduleMoves"`
	ContactTries  int        `db:"contactTries"`

	FirstAssignAt *time.Time     `db:"firstAssignAt"`
	CrewIDs       pq.StringArray `db:"crewIds"`
	TechIDs       pq.StringArray `db:"techIds"`
	LeaderID      *string        `db:"leaderId"`
	LeaderName    *string        `db:"leaderName"`

	MissionAt    *time.Time `db:"missionAt"`
	MaterialsAt  *time.Time `db:"materialsAt"`
	DepartedAt   *time.Time `db:"departedAt"`
	ArrivedAt    *time.Time `db:"arrivedAt"`
	WorkStartAt  *time.Time `db:"workStartAt"`
	StartedAt    *time.Time `db:"startedAt"`
	CompletedAt  *time.Time `db:"completedAt"`
	StoppedAt    *time.Time `db:"stoppedAt"`
	PartialCount int        `db:"partialCount"`

	InvoiceAt     *time.Time `db:"invoiceAt"`
	InvoiceByID   *string    `db:"invoiceById"`
	InvoiceNet    *float64   `db:"invoiceNet"`
	ReportAt      *time.Time `db:"reportAt"`
	AuditedAt     *time.Time `db:"auditedAt"`
	AuditedByID   *string    `db:"auditedById"`
	AuditedBy     *string    `db:"auditedBy"`
	ApprovedAt    *time.Time `db:"approvedAt"`
	ApprovedByID  *string    `db:"approvedById"`
	ApprovedBy    *string    `db:"approvedBy"`
	Collected     *float64   `db:"collected"`
	AmountChecked bool       `db:"amountVerified"`

	RatedCount     int            `db:"ratedCount"`
	RatedAt        *time.Time     `db:"ratedAt"`
	InventoryIDs   pq.StringArray `db:"inventoryIds"`        // جردوا بعد الحجز
	InventoryMiss  pq.StringArray `db:"inventoryMissingIds"` // جردوا ولگوا نقص
	QualityAt      *time.Time     `db:"qualityAt"`
	QualityByID    *string        `db:"qualityById"`
	QualityBy      *string        `db:"qualityBy"`
	QualityStatus  *string        `db:"qualityStatus"`
	QualityCreated *time.Time     `db:"qualityCreated"`

	MonitorRows    int        `db:"monitorRows"`
	MonitorPending int        `db:"monitorPending"`
	MonitorFirstAt *time.Time `db:"monitorFirstAt"`
	MonitorLastAt  *time.Time `db:"monitorLastAt"`
	MonitorByID    *string    `db:"monitorById"`
	MonitorBy      *string    `db:"monitorBy"`
	MonitorMinutes *float64   `db:"monitorMinutes"`
}

type MatrixChainRepository struct{ db *sqlx.DB }

func NewMatrixChainRepository(db *sqlx.DB) *MatrixChainRepository {
	return &MatrixChainRepository{db: db}
}

const chainFactsSQL = `
SELECT b.id, b.code, b.status::text AS status, b."bookingType"::text AS "bookingType",
       b."serviceId", s.name AS "serviceName", COALESCE(s."managerHandlesPaperwork", false) AS solo,
       b."createdAt", b."createdById", ce.name AS "createdByName",
       (COALESCE(btrim(b.address), '') <> '') AS "hasAddress",
       (b."mapLatitude" IS NOT NULL OR COALESCE(b."locationUrl", '') <> '' OR COALESCE(b."mapLocation", '') <> '') AS "hasLocation",
       COALESCE(array_length(regexp_split_to_array(btrim(c.name), '\s+'), 1), 0) AS "nameWords",
       EXISTS (SELECT 1 FROM "DuplicateCandidate" d WHERE d.kind = 'BOOKING'
               AND (d."entityAId" = b.id OR d."entityBId" = b.id)) AS duplicate,
       b."confirmationContactedAt" AS "contactedAt", b."confirmationContactedById" AS "contactedById", ke.name AS "contactedBy",
       b."confirmedAt", b."confirmedByEmployeeId" AS "confirmedById", fe.name AS "confirmedBy",
       b."scheduledAt", b."postponeCount",
       COALESCE(b."transferToProjects", false) AS "toProjects", b."projectExecutionAt" AS "projectExecAt",
       (SELECT p.stage FROM "Project" p WHERE p."bookingId" = b.id ORDER BY p."createdAt" DESC LIMIT 1) AS "projectStage",
       (SELECT count(*) FROM "ScheduleChangeLog" l WHERE l."bookingId" = b.id)::int AS "scheduleMoves",
       b."contactAttempts" AS "contactTries",
       asg."firstAssignAt", COALESCE(asg.crew, '{}') AS "crewIds", COALESCE(asg.techs, '{}') AS "techIds",
       COALESCE(b."projectSupervisorId", asg.leader, m."leaderId") AS "leaderId", le.name AS "leaderName",
       m."assignedAt" AS "missionAt", COALESCE(m."materialsReadyAt", b."materialsReadyAt") AS "materialsAt",
       m."departedAt", COALESCE(m."arrivedAt", b."arrivedAt") AS "arrivedAt", m."workStartedAt" AS "workStartAt",
       b."startedAt", b."completedAt", b."workStoppedAt" AS "stoppedAt", b."partialCount",
       inv."createdAt" AS "invoiceAt", inv."employeeId" AS "invoiceById", inv."netTotal" AS "invoiceNet",
       (SELECT min(w."createdAt") FROM "WorkReport" w WHERE w."bookingId" = b.id) AS "reportAt",
       inv."auditedAt", inv."auditedById", ae.name AS "auditedBy",
       inv."approvedAt", inv."approvedByEmployeeId" AS "approvedById", pe.name AS "approvedBy",
       b."amountCollected" AS collected, b."amountVerified",
       (SELECT count(*) FROM "CrewRating" r WHERE r."bookingId" = b.id)::int AS "ratedCount",
       (SELECT min(r."createdAt") FROM "CrewRating" r WHERE r."bookingId" = b.id) AS "ratedAt",
       COALESCE((SELECT array_agg(i."employeeId") FROM "BookingAfterInventory" i WHERE i."bookingId" = b.id), '{}') AS "inventoryIds",
       COALESCE((SELECT array_agg(i."employeeId") FROM "BookingAfterInventory" i WHERE i."bookingId" = b.id AND NOT i.complete), '{}') AS "inventoryMissingIds",
       q."contactedAt" AS "qualityAt", q."contactedByEmployeeId" AS "qualityById", qe.name AS "qualityBy",
       q.status AS "qualityStatus", q."createdAt" AS "qualityCreated",
       COALESCE(mr.n, 0) AS "monitorRows", COALESCE(mr.pending, 0) AS "monitorPending",
       mr."firstAt" AS "monitorFirstAt", mr."lastAt" AS "monitorLastAt", mr."byId" AS "monitorById", me.name AS "monitorBy",
       mr.minutes AS "monitorMinutes"
FROM "Booking" b
JOIN "Customer" c ON c.id = b."customerId"
LEFT JOIN "Service" s ON s.id = b."serviceId"
LEFT JOIN "Employee" ce ON ce.id = b."createdById"
LEFT JOIN "Employee" ke ON ke.id = b."confirmationContactedById"
LEFT JOIN "Employee" fe ON fe.id = b."confirmedByEmployeeId"
LEFT JOIN LATERAL (
	SELECT min(a."createdAt") AS "firstAssignAt",
	       array_agg(a."employeeId") AS crew,
	       array_agg(a."employeeId") FILTER (WHERE NOT e."isLeader") AS techs,
	       (array_agg(a."employeeId" ORDER BY a."createdAt") FILTER (WHERE e."isLeader"))[1] AS leader
	FROM "BookingAssignment" a JOIN "Employee" e ON e.id = a."employeeId"
	WHERE a."bookingId" = b.id
) asg ON true
LEFT JOIN LATERAL (
	SELECT * FROM "Mission" mm WHERE mm."bookingId" = b.id ORDER BY mm."assignedAt" DESC LIMIT 1
) m ON true
LEFT JOIN "Employee" le ON le.id = COALESCE(b."projectSupervisorId", asg.leader, m."leaderId")
LEFT JOIN LATERAL (
	SELECT li."createdAt", li."employeeId", li."netTotal", li."auditedAt", li."auditedById",
	       li."approvedAt", li."approvedByEmployeeId"
	FROM "LeaderInvoice" li WHERE li."bookingId" = b.id ORDER BY li."createdAt" LIMIT 1
) inv ON true
LEFT JOIN "Employee" ae ON ae.id = inv."auditedById"
LEFT JOIN "Employee" pe ON pe.id = inv."approvedByEmployeeId"
LEFT JOIN LATERAL (
	SELECT qq.* FROM "QualityFollowUp" qq WHERE qq."bookingId" = b.id ORDER BY qq."createdAt" LIMIT 1
) q ON true
LEFT JOIN "Employee" qe ON qe.id = q."contactedByEmployeeId"
LEFT JOIN LATERAL (
	SELECT count(*)::int AS n,
	       count(*) FILTER (WHERE r.status = 'PENDING')::int AS pending,
	       min(r."createdAt") AS "firstAt", max(r."reviewedAt") AS "lastAt",
	       (array_agg(r."reviewedById" ORDER BY r."reviewedAt" DESC) FILTER (WHERE r."reviewedById" IS NOT NULL))[1] AS "byId",
	       avg(EXTRACT(EPOCH FROM (r."reviewedAt" - r."createdAt")) / 60) FILTER (WHERE r."reviewedAt" IS NOT NULL) AS minutes
	FROM "MonitorReview" r
	WHERE (r."entityType" = 'BOOKING' AND r."entityId" = b.id)
	   OR (r."entityType" = 'LEADER_INVOICE' AND r."entityId" IN (SELECT id FROM "LeaderInvoice" WHERE "bookingId" = b.id))
) mr ON true
LEFT JOIN "Employee" me ON me.id = mr."byId"
`

// ForBooking حقائق حجز واحد (حتى لو قديم — العرض يگول هذا).
func (r *MatrixChainRepository) ForBooking(id string) (*ChainFacts, error) {
	var f ChainFacts
	if err := r.db.Get(&f, chainFactsSQL+` WHERE b.id = $1`, id); err != nil {
		return nil, err
	}
	return &f, nil
}

// Since حقائق كل الحجوزات المسجّلة من تاريخ — للوسيط وتقارير الأدوار.
func (r *MatrixChainRepository) Since(from time.Time) ([]ChainFacts, error) {
	rows := []ChainFacts{}
	err := r.db.Select(&rows, chainFactsSQL+`
		WHERE b."createdAt" >= $1 AND upper(b.code) NOT LIKE 'OLD%' AND b.status::text <> 'CANCELLED'
		  AND b."bookingType"::text <> 'INTERNAL'
		ORDER BY b."createdAt"`, from)
	return rows, err
}

// ═══ تقييم الليدر لفنيّيه ═══

type CrewRatingRow struct {
	BookingID    string    `db:"bookingId" json:"bookingId"`
	BookingCode  string    `db:"bookingCode" json:"bookingCode"`
	LeaderID     string    `db:"leaderId" json:"leaderId"`
	LeaderName   string    `db:"leaderName" json:"leaderName"`
	TechnicianID string    `db:"technicianId" json:"technicianId"`
	Score        int       `db:"score" json:"score"`
	Note         *string   `db:"note" json:"note"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

// PendingCrewRating حجز خلّصه الليدر وبعده ما قيّم كل فنيّيه.
type PendingCrewRating struct {
	BookingID   string         `db:"bookingId" json:"bookingId"`
	BookingCode string         `db:"bookingCode" json:"bookingCode"`
	CompletedAt time.Time      `db:"completedAt" json:"completedAt"`
	TechIDs     pq.StringArray `db:"techIds" json:"-"`
	Techs       []EmployeeName `db:"-" json:"techs"`
}

type EmployeeName struct {
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// leaderBookingsSQL الحجوزات المنجزة الي الموظف ليدرها (بالتكليف أو بالمهمة)
// وبيها فنيين مو ليدرية ما انقيّموا — آخر ٣٠ يوم.
const pendingRatingSQL = `
	SELECT b.id AS "bookingId", b.code AS "bookingCode", b."completedAt",
	       array_agg(a."employeeId") AS "techIds"
	FROM "Booking" b
	JOIN "BookingAssignment" a ON a."bookingId" = b.id
	JOIN "Employee" t ON t.id = a."employeeId" AND NOT t."isLeader" AND t.id <> $1
	WHERE b.status::text IN ('COMPLETED', 'PARTIAL') AND b."completedAt" IS NOT NULL
	  AND b."completedAt" > now() - interval '30 days'
	  AND upper(b.code) NOT LIKE 'OLD%'
	  AND (b."projectSupervisorId" = $1
	       OR EXISTS (SELECT 1 FROM "BookingAssignment" la JOIN "Employee" le ON le.id = la."employeeId"
	               WHERE la."bookingId" = b.id AND la."employeeId" = $1 AND le."isLeader")
	       OR EXISTS (SELECT 1 FROM "Mission" m WHERE m."bookingId" = b.id AND m."leaderId" = $1))
	  AND NOT EXISTS (SELECT 1 FROM "CrewRating" r WHERE r."bookingId" = b.id AND r."technicianId" = a."employeeId")
	GROUP BY b.id
	ORDER BY b."completedAt"`

func (r *MatrixChainRepository) PendingRatings(leaderID string) ([]PendingCrewRating, error) {
	rows := []PendingCrewRating{}
	if err := r.db.Select(&rows, pendingRatingSQL, leaderID); err != nil {
		return nil, err
	}
	for i := range rows {
		names := []EmployeeName{}
		_ = r.db.Select(&names, `SELECT id, name FROM "Employee" WHERE id = ANY($1) ORDER BY name`, rows[i].TechIDs)
		rows[i].Techs = names
	}
	return rows, nil
}

// IsBookingLeader الموظف ليدر هذا الحجز؟
func (r *MatrixChainRepository) IsBookingLeader(bookingID, employeeID string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Booking" WHERE id = $1 AND "projectSupervisorId" = $2)
	                OR EXISTS (SELECT 1 FROM "BookingAssignment" a JOIN "Employee" e ON e.id = a."employeeId"
	                     WHERE a."bookingId" = $1 AND a."employeeId" = $2 AND e."isLeader")
	                OR EXISTS (SELECT 1 FROM "Mission" m WHERE m."bookingId" = $1 AND m."leaderId" = $2)`, bookingID, employeeID)
	return ok
}

// IsBookingTech الفني مكلّف بهذا الحجز؟
func (r *MatrixChainRepository) IsBookingTech(bookingID, employeeID string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "BookingAssignment" WHERE "bookingId" = $1 AND "employeeId" = $2)`, bookingID, employeeID)
	return ok
}

func (r *MatrixChainRepository) SaveRating(bookingID, leaderID, techID string, score int, note *string) error {
	_, err := r.db.Exec(`INSERT INTO "CrewRating" (id, "bookingId", "leaderId", "technicianId", score, note)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		ON CONFLICT ("bookingId", "technicianId") DO UPDATE SET score = EXCLUDED.score, note = EXCLUDED.note,
		  "leaderId" = EXCLUDED."leaderId", "createdAt" = now()`, bookingID, leaderID, techID, score, note)
	return err
}

// RatingsFor تقييمات الفنيين — لكل حجز أو لفني.
func (r *MatrixChainRepository) RatingsByBooking(bookingID string) ([]CrewRatingRow, error) {
	return r.ratings(`r."bookingId" = $1`, bookingID)
}

func (r *MatrixChainRepository) RatingsSince(from time.Time) ([]CrewRatingRow, error) {
	return r.ratings(`r."createdAt" >= $1`, from)
}

func (r *MatrixChainRepository) ratings(where string, arg any) ([]CrewRatingRow, error) {
	rows := []CrewRatingRow{}
	err := r.db.Select(&rows, `SELECT r."bookingId", b.code AS "bookingCode", r."leaderId", l.name AS "leaderName",
	        r."technicianId", r.score, r.note, r."createdAt"
	   FROM "CrewRating" r JOIN "Booking" b ON b.id = r."bookingId" JOIN "Employee" l ON l.id = r."leaderId"
	  WHERE `+where+` ORDER BY r."createdAt" DESC`, arg)
	return rows, err
}

// LeadersPendingRating الليدرية الي عندهم حجوزات منجزة ما قيّموا فنيّيها
// من يوم أو أكثر — لتذكير ماتركس.
type LeaderPending struct {
	LeaderID string         `db:"leaderId"`
	Codes    pq.StringArray `db:"codes"`
}

func (r *MatrixChainRepository) LeadersPendingRating() ([]LeaderPending, error) {
	rows := []LeaderPending{}
	err := r.db.Select(&rows, `
		WITH lb AS (
			SELECT DISTINCT b.id, b.code, COALESCE(b."projectSupervisorId", la."employeeId", m."leaderId") AS "leaderId"
			FROM "Booking" b
			LEFT JOIN LATERAL (SELECT a."employeeId" FROM "BookingAssignment" a JOIN "Employee" e ON e.id = a."employeeId"
			                   WHERE a."bookingId" = b.id AND e."isLeader" ORDER BY a."createdAt" LIMIT 1) la ON true
			LEFT JOIN LATERAL (SELECT "leaderId" FROM "Mission" WHERE "bookingId" = b.id ORDER BY "assignedAt" DESC LIMIT 1) m ON true
			WHERE b.status::text IN ('COMPLETED', 'PARTIAL') AND b."completedAt" IS NOT NULL
			  AND b."completedAt" BETWEEN now() - interval '30 days' AND now() - interval '12 hours'
			  AND upper(b.code) NOT LIKE 'OLD%'
		)
		SELECT lb."leaderId", array_agg(lb.code ORDER BY lb.code) AS codes
		FROM lb JOIN "Employee" le ON le.id = lb."leaderId" AND le.status = 'ACTIVE'
		WHERE EXISTS (SELECT 1 FROM "BookingAssignment" a JOIN "Employee" t ON t.id = a."employeeId"
		              WHERE a."bookingId" = lb.id AND NOT t."isLeader" AND t.id <> lb."leaderId"
		                AND NOT EXISTS (SELECT 1 FROM "CrewRating" r WHERE r."bookingId" = lb.id AND r."technicianId" = t.id))
		GROUP BY lb."leaderId"`)
	return rows, err
}

// LeaderHasPending بعده عنده حجوزات بلا تقييم؟ (للتذكير «انحل؟»)
func (r *MatrixChainRepository) LeaderHasPending(leaderID string) bool {
	rows, err := r.PendingRatings(leaderID)
	return err == nil && len(rows) > 0
}

// ═══ جرد العدّة بعد الحجز ═══

// afterPendingSQL حجوزات منجزة (آخر ١٤ يوم) الموظف مكلّف بيها كفني ويا ليدر
// (مو ليدر، والخدمة مو فردية) وبعده ما جرد عدّته بعدها.
const afterPendingSQL = `
	SELECT b.id AS "bookingId", b.code AS "bookingCode", b."completedAt"
	FROM "Booking" b
	JOIN "BookingAssignment" a ON a."bookingId" = b.id AND a."employeeId" = $1
	JOIN "Employee" me ON me.id = $1 AND NOT me."isLeader" AND b."projectSupervisorId" IS DISTINCT FROM $1
	LEFT JOIN "Service" s ON s.id = b."serviceId"
	WHERE b.status::text IN ('COMPLETED', 'PARTIAL') AND b."completedAt" IS NOT NULL
	  AND b."completedAt" > now() - interval '14 days'
	  AND upper(b.code) NOT LIKE 'OLD%' AND b."bookingType"::text <> 'SURVEY'
	  AND NOT COALESCE(s."managerHandlesPaperwork", false)
	  AND (b."projectSupervisorId" IS NOT NULL OR EXISTS (SELECT 1 FROM "BookingAssignment" la JOIN "Employee" le ON le.id = la."employeeId"
	              WHERE la."bookingId" = b.id AND le."isLeader"))
	  AND NOT EXISTS (SELECT 1 FROM "BookingAfterInventory" x WHERE x."bookingId" = b.id AND x."employeeId" = $1)
	ORDER BY b."completedAt"`

type AfterInventoryPending struct {
	BookingID   string    `db:"bookingId" json:"bookingId"`
	BookingCode string    `db:"bookingCode" json:"bookingCode"`
	CompletedAt time.Time `db:"completedAt" json:"completedAt"`
}

func (r *MatrixChainRepository) AfterInventoryPending(employeeID string) ([]AfterInventoryPending, error) {
	rows := []AfterInventoryPending{}
	err := r.db.Select(&rows, afterPendingSQL, employeeID)
	return rows, err
}

func (r *MatrixChainRepository) SaveAfterInventory(bookingID, employeeID string, complete bool, missing *string) error {
	_, err := r.db.Exec(`INSERT INTO "BookingAfterInventory" ("bookingId", "employeeId", complete, "missingItems")
		VALUES ($1, $2, $3, $4) ON CONFLICT ("bookingId", "employeeId") DO UPDATE
		SET complete = EXCLUDED.complete, "missingItems" = EXCLUDED."missingItems", "checkedAt" = now()`,
		bookingID, employeeID, complete, missing)
	return err
}

// TechsPendingAfterInventory الفنيين الي عندهم حجوزات منجزة من ٦ ساعات أو أكثر
// بلا جرد بعدها — لتذكير ماتركس.
func (r *MatrixChainRepository) TechsPendingAfterInventory() ([]LeaderPending, error) {
	rows := []LeaderPending{}
	err := r.db.Select(&rows, `
		SELECT a."employeeId" AS "leaderId", array_agg(DISTINCT b.code) AS codes
		FROM "Booking" b
		JOIN "BookingAssignment" a ON a."bookingId" = b.id
		JOIN "Employee" t ON t.id = a."employeeId" AND NOT t."isLeader" AND t.status = 'ACTIVE'
		     AND b."projectSupervisorId" IS DISTINCT FROM t.id
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.status::text IN ('COMPLETED', 'PARTIAL') AND b."completedAt" IS NOT NULL
		  AND b."completedAt" BETWEEN now() - interval '14 days' AND now() - interval '6 hours'
		  AND upper(b.code) NOT LIKE 'OLD%' AND b."bookingType"::text <> 'SURVEY'
		  AND NOT COALESCE(s."managerHandlesPaperwork", false)
		  AND (b."projectSupervisorId" IS NOT NULL OR EXISTS (SELECT 1 FROM "BookingAssignment" la JOIN "Employee" le ON le.id = la."employeeId"
		              WHERE la."bookingId" = b.id AND le."isLeader"))
		  AND NOT EXISTS (SELECT 1 FROM "BookingAfterInventory" x WHERE x."bookingId" = b.id AND x."employeeId" = a."employeeId")
		GROUP BY a."employeeId"`)
	return rows, err
}

// BookingCodeAndEmployee كود الحجز واسم الموظف — لرسالة النقص الداخلية.
func (r *MatrixChainRepository) BookingCodeAndEmployee(bookingID, employeeID string) (string, string) {
	var row struct {
		Code string `db:"code"`
		Name string `db:"name"`
	}
	_ = r.db.Get(&row, `SELECT b.code, e.name FROM "Booking" b, "Employee" e WHERE b.id = $1 AND e.id = $2`, bookingID, employeeID)
	return row.Code, row.Name
}

// ═══ ماتركس على المراقب: بنود الصندوق المتأخرة ═══

// monitorBacklogSQL بنود صندوق المراقب الي تنتظر حكم من أكثر من ٢٤ ساعة.
const monitorBacklogSQL = `SELECT count(*) FROM "MonitorReview" WHERE status = 'PENDING' AND "createdAt" < now() - interval '24 hours'`

type MonitorBacklog struct {
	Count  int        `db:"n"`
	Oldest *time.Time `db:"oldest"`
}

func (r *MatrixChainRepository) MonitorBacklog() (MonitorBacklog, error) {
	var b MonitorBacklog
	err := r.db.Get(&b, `SELECT count(*)::int AS n, min("createdAt") AS oldest FROM "MonitorReview"
		WHERE status = 'PENDING' AND "createdAt" < now() - interval '24 hours'`)
	return b, err
}

// Monitors المراقبين الفعّالين: دور MONITOR أو صلاحية monitoring/auditing (بلا المدير والمالك).
func (r *MatrixChainRepository) Monitors() ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `SELECT e.id FROM "Employee" e WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER')
		AND (e.role::text = 'MONITOR' OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
		     WHERE ep."employeeId" = e.id AND p.name IN ('monitoring', 'auditing')))`)
	return ids, err
}

// ContactFeatureSince أول مرة انضغط «تواصلت ويا الزبون» بالنظام — الحجوزات
// الأقدم ما ينحاسب عليها أحد (الزر ما چان موجود).
func (r *MatrixChainRepository) ContactFeatureSince() *time.Time {
	var t *time.Time
	_ = r.db.Get(&t, `SELECT min("confirmationContactedAt") FROM "Booking"`)
	return t
}
