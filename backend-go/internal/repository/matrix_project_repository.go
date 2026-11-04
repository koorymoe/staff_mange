package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ عين ماتركس على المشاريع — الحقائق الخام ═══
// أي أحد يتوجهله مشروع (مشرف، تقني، مصمم، بغض النظر عن الدور) ينحسب
// من ثلاث مصادر: سجل المراحل (0313)، سجل التوجيه، وسجل النشاط (0305) لكل
// حفظ ناجح على مسار المشروع.

type MatrixProjectRepository struct{ db *sqlx.DB }

func NewMatrixProjectRepository(db *sqlx.DB) *MatrixProjectRepository {
	return &MatrixProjectRepository{db: db}
}

type ProjectFacts struct {
	ID             string     `db:"id" json:"id"`
	Code           string     `db:"code" json:"code"`
	Name           string     `db:"name" json:"name"`
	Stage          string     `db:"stage" json:"stage"`
	WorkType       *string    `db:"workType" json:"workType"`
	DeliveryDate   *string    `db:"deliveryDate" json:"deliveryDate"`
	CreatedAt      time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time  `db:"updatedAt" json:"updatedAt"`
	CreatedByID    *string    `db:"createdById" json:"createdById"`
	CreatedBy      *string    `db:"createdBy" json:"createdBy"`
	ResponsibleID  *string    `db:"responsibleId" json:"responsibleId"`
	Responsible    *string    `db:"responsible" json:"responsible"`
	SurveyorID     *string    `db:"surveyorId" json:"surveyorId"`
	Surveyor       *string    `db:"surveyor" json:"surveyor"`
	DelegatedToID  *string    `db:"delegatedToId" json:"delegatedToId"`
	DelegatedTo    *string    `db:"delegatedTo" json:"delegatedTo"`
	DelegatedAt    *time.Time `db:"delegatedAt" json:"delegatedAt"`
	BookingID      *string    `db:"bookingId" json:"bookingId"`
	BookingCode    *string    `db:"bookingCode" json:"bookingCode"`
	HasContract    bool       `db:"hasContract" json:"hasContract"`
	HasSigned      bool       `db:"hasSigned" json:"hasSigned"`
	StageSince     *time.Time `db:"stageSince" json:"stageSince"`
	LastActivityAt *time.Time `db:"lastActivityAt" json:"lastActivityAt"`
	LastActivityBy *string    `db:"lastActivityBy" json:"lastActivityBy"`
	Touchers       int        `db:"touchers" json:"touchers"`
	ChecklistCount int        `db:"checklistCount" json:"checklistCount"`
	// قرار (ع) 10-10: آخر سبب تأخير انكتب للمرحلة الحالية
	DelayReason   *string    `db:"delayReason" json:"delayReason"`
	DelayReasonBy *string    `db:"delayReasonBy" json:"delayReasonBy"`
	DelayReasonAt *time.Time `db:"delayReasonAt" json:"delayReasonAt"`
}

const projectFactsSQL = `
SELECT p.id, p.code, p.name, p.stage, p."workType", p."deliveryDate", p."createdAt", p."updatedAt",
       p."createdByEmployeeId" AS "createdById", ce.name AS "createdBy",
       p."responsibleEmployeeId" AS "responsibleId", re.name AS responsible,
       p."surveyorEmployeeId" AS "surveyorId", se.name AS surveyor,
       p."delegatedToEmployeeId" AS "delegatedToId", de.name AS "delegatedTo", p."delegatedAt",
       p."bookingId", b.code AS "bookingCode",
       (p."contractPdfBase64" IS NOT NULL) AS "hasContract", (p."signedContractPdfBase64" IS NOT NULL) AS "hasSigned",
       (SELECT max(l."createdAt") FROM "ProjectStageLog" l WHERE l."projectId" = p.id) AS "stageSince",
       act."lastAt" AS "lastActivityAt", ae.name AS "lastActivityBy", COALESCE(act.n, 0) AS touchers,
       (SELECT count(*) FROM "ProjectChecklist" c WHERE c."projectId" = p.id)::int AS "checklistCount",
       dr.reason AS "delayReason", dre.name AS "delayReasonBy", dr."createdAt" AS "delayReasonAt"
FROM "Project" p
LEFT JOIN LATERAL (SELECT reason, "employeeId", "createdAt" FROM "ProjectDelayReason" d
	WHERE d."projectId" = p.id AND d.stage = p.stage ORDER BY d."createdAt" DESC LIMIT 1) dr ON true
LEFT JOIN "Employee" dre ON dre.id = dr."employeeId"
LEFT JOIN "Employee" ce ON ce.id = p."createdByEmployeeId"
LEFT JOIN "Employee" re ON re.id = p."responsibleEmployeeId"
LEFT JOIN "Employee" se ON se.id = p."surveyorEmployeeId"
LEFT JOIN "Employee" de ON de.id = p."delegatedToEmployeeId"
LEFT JOIN "Booking" b ON b.id = p."bookingId"
LEFT JOIN LATERAL (
	SELECT max(a."createdAt") AS "lastAt",
	       (array_agg(a."employeeId" ORDER BY a."createdAt" DESC))[1] AS "lastBy",
	       count(DISTINCT a."employeeId")::int AS n
	FROM "EmployeeActivity" a WHERE a.path LIKE '/api/projects/' || p.id || '%'
) act ON true
LEFT JOIN "Employee" ae ON ae.id = act."lastBy"
`

func (r *MatrixProjectRepository) All() ([]ProjectFacts, error) {
	rows := []ProjectFacts{}
	err := r.db.Select(&rows, projectFactsSQL+` ORDER BY p."createdAt" DESC`)
	return rows, err
}

func (r *MatrixProjectRepository) One(id string) (*ProjectFacts, error) {
	var f ProjectFacts
	if err := r.db.Get(&f, projectFactsSQL+` WHERE p.id = $1`, id); err != nil {
		return nil, err
	}
	return &f, nil
}

type ProjectStageRow struct {
	FromStage *string   `db:"fromStage" json:"fromStage"`
	ToStage   string    `db:"toStage" json:"toStage"`
	At        time.Time `db:"createdAt" json:"at"`
	ByID      *string   `db:"byId" json:"byId"`
	By        *string   `db:"by" json:"by"`
}

// Stages تغييرات المرحلة بالترتيب — ومنو غيّرها: أقرب حفظ على مسار المشروع
// بنفس الدقيقة من سجل النشاط.
func (r *MatrixProjectRepository) Stages(id string) ([]ProjectStageRow, error) {
	rows := []ProjectStageRow{}
	err := r.db.Select(&rows, `
		SELECT l."fromStage", l."toStage", l."createdAt", a."employeeId" AS "byId", e.name AS by
		FROM "ProjectStageLog" l
		LEFT JOIN LATERAL (
			SELECT x."employeeId" FROM "EmployeeActivity" x
			WHERE x.path LIKE '/api/projects%' AND x.method IN ('PUT','POST','PATCH')
			  AND x."createdAt" BETWEEN l."createdAt" - interval '1 minute' AND l."createdAt" + interval '1 minute'
			  AND (x.path LIKE '/api/projects/' || l."projectId" || '%' OR x.path = '/api/projects')
			ORDER BY abs(EXTRACT(EPOCH FROM (x."createdAt" - l."createdAt"))) LIMIT 1
		) a ON true
		LEFT JOIN "Employee" e ON e.id = a."employeeId"
		WHERE l."projectId" = $1 ORDER BY l."createdAt"`, id)
	return rows, err
}

type ProjectTouchRow struct {
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	Name       string    `db:"name" json:"name"`
	Role       string    `db:"role" json:"role"`
	FirstAt    time.Time `db:"firstAt" json:"firstAt"`
	LastAt     time.Time `db:"lastAt" json:"lastAt"`
	Actions    int       `db:"actions" json:"actions"`
}

// Touchers منو اشتغل على المشروع، بترتيب أول مرة لمسه.
func (r *MatrixProjectRepository) Touchers(id string) ([]ProjectTouchRow, error) {
	rows := []ProjectTouchRow{}
	err := r.db.Select(&rows, `
		SELECT a."employeeId", e.name, e.role::text AS role, min(a."createdAt") AS "firstAt", max(a."createdAt") AS "lastAt", count(*)::int AS actions
		FROM "EmployeeActivity" a JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a.path LIKE '/api/projects/' || $1 || '%'
		GROUP BY a."employeeId", e.name, e.role ORDER BY min(a."createdAt")`, id)
	return rows, err
}

type ProjectDelegationRow struct {
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	Name       *string   `db:"name" json:"name"`
	Action     string    `db:"action" json:"action"`
	By         *string   `db:"by" json:"by"`
	At         time.Time `db:"createdAt" json:"at"`
}

func (r *MatrixProjectRepository) Delegations(id string) ([]ProjectDelegationRow, error) {
	rows := []ProjectDelegationRow{}
	err := r.db.Select(&rows, `
		SELECT l."employeeId", e.name, l.action, b.name AS by, l."createdAt"
		FROM "ProjectDelegationLog" l LEFT JOIN "Employee" e ON e.id = l."employeeId"
		LEFT JOIN "Employee" b ON b.id = l."delegatedByEmployeeId"
		WHERE l."projectId" = $1 ORDER BY l."createdAt"`, id)
	return rows, err
}

// StageDurations لكل مرحلة: مدة بقاء كل مشروع بيها (بالأيام) — من آخر ١٨٠ يوم.
type StageDuration struct {
	Stage string  `db:"stage"`
	Days  float64 `db:"days"`
}

func (r *MatrixProjectRepository) StageDurations() ([]StageDuration, error) {
	rows := []StageDuration{}
	err := r.db.Select(&rows, `
		SELECT "toStage" AS stage, EXTRACT(EPOCH FROM (nxt - "createdAt")) / 86400 AS days
		FROM (SELECT "toStage", "createdAt", lead("createdAt") OVER (PARTITION BY "projectId" ORDER BY "createdAt") AS nxt
		      FROM "ProjectStageLog" WHERE "createdAt" > now() - interval '180 days') x
		WHERE nxt IS NOT NULL`)
	return rows, err
}

// ManualStageLimits الحدود الي ثبّتها المالك بإيده (تغلب المتعلَّمة).
func (r *MatrixProjectRepository) ManualStageLimits() (map[string]int, error) {
	rows := []struct {
		Stage string `db:"stage"`
		Days  int    `db:"days"`
	}{}
	out := map[string]int{}
	if err := r.db.Select(&rows, `SELECT stage, days FROM "ProjectStageLimit"`); err != nil {
		return out, err
	}
	for _, x := range rows {
		out[x.Stage] = x.Days
	}
	return out, nil
}
