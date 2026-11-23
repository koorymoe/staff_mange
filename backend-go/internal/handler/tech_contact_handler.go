package handler

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/errlog"
	"staffmange-api/internal/middleware"
)

// ═══ تقارير اتصال التقنيين (ك 11-23، طلب زينا) ═══
// «لازم يطلعن التقارير عند: المراقب، ومدير العمليات ونائبه، ومدير الإدارة ونائبه».
// شاشة وحدة بيها كل شي انحوّل لتقني (حجز أو استفسار طاقة شمسية) آخر ٣٠ يوم:
// منو التقني، متى انحوّل، متى رفع تقرير الاتصال وشنو كتب، وإذا انخصمت عليه نقطة.
// الفحص هنا: غير هذولا (ومدير النظام والمالك) يطلعله 404.

type TechContactHandler struct{ db *sqlx.DB }

func NewTechContactHandler(db *sqlx.DB) *TechContactHandler { return &TechContactHandler{db: db} }

// techContactWatcherSQL $1 = الموظف: مراقب، أو مدير/نائب العمليات أو الإدارة، أو مدير النظام/المالك.
const techContactWatcherSQL = `SELECT EXISTS (SELECT 1 FROM "Employee" e WHERE e.id = $1 AND e.status = 'ACTIVE' AND (
	e.role::text IN ('ADMIN', 'OWNER', 'MONITOR')
	OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = e.id AND p.name = 'monitoring')
	OR EXISTS (SELECT 1 FROM "OrgAssignment" oa WHERE oa."employeeId" = e.id AND oa."endedAt" IS NULL AND oa."unitId" IS NULL
		AND oa."divisionId" IN ('ops', 'admin') AND oa.position IN ('DIRECTOR', 'DEPUTY'))))`

type techContactRow struct {
	Kind        string     `db:"kind" json:"kind"`
	Ref         string     `db:"ref" json:"ref"`
	Customer    *string    `db:"customer" json:"customer"`
	Tech        string     `db:"tech" json:"tech"`
	By          *string    `db:"by" json:"by"`
	Since       time.Time  `db:"since" json:"since"`
	ContactedAt *time.Time `db:"contactedAt" json:"contactedAt"`
	Note        *string    `db:"note" json:"note"`
	KpiAt       *time.Time `db:"kpiAt" json:"kpiAt"`
	Closed      bool       `db:"closed" json:"closed"`
}

// GET /api/tech-contact/reports
func (h *TechContactHandler) Reports(w http.ResponseWriter, r *http.Request) {
	var ok bool
	if err := h.db.Get(&ok, techContactWatcherSQL, middleware.EmployeeIDFromContext(r)); err != nil || !ok {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	rows := []techContactRow{}
	errlog.Err(h.db.Select(&rows, `
		SELECT 'booking' AS kind, b.code AS ref, c.name AS customer, t.name AS tech, hb.name AS "by",
			b."handoverAt" AT TIME ZONE current_setting('TimeZone') AS since,
			b."techContactedAt" AT TIME ZONE current_setting('TimeZone') AS "contactedAt",
			b."techContactNote" AS note, b."techContactKpiAt" AS "kpiAt",
			(b.status::text IN ('COMPLETED', 'CANCELLED')) AS closed
		FROM "Booking" b JOIN "Employee" t ON t.id = b."handoverToId"
		LEFT JOIN "Employee" hb ON hb.id = b."handoverById"
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		WHERE b."handoverAt" > now() - interval '30 days' AND b."archivedAt" IS NULL
		UNION ALL
		SELECT 'solar', '#' || q.number::text, q."customerName", t.name, cb.name,
			q."techClaimedAt", q."techContactAt",
			concat_ws(' — ', CASE q."techOutcome" WHEN 'JUST_ASKING' THEN 'بس يستفسر' WHEN 'WANTS_INSTALL' THEN 'يريد ينصب منظومة'
				WHEN 'NO_ANSWER' THEN 'ما رد' WHEN 'NOT_INTERESTED' THEN 'ما مهتم' END, q."techNote"),
			q."techContactKpiAt", (q."qualityAt" IS NOT NULL)
		FROM "SolarInquiry" q JOIN "Employee" t ON t.id = q."techId"
		LEFT JOIN "Employee" cb ON cb.id = q."createdById"
		WHERE q."techClaimedAt" > now() - interval '30 days'
		ORDER BY since DESC LIMIT 300`))
	for i := range rows {
		if rows[i].Note != nil && *rows[i].Note == "" {
			rows[i].Note = nil
		}
	}
	WriteJSON(w, http.StatusOK, rows)
}
