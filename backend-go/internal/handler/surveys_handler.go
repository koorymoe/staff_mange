package handler

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ 📋 كل الكشوفات بمكان واحد (دازاي 11-15) ═══
// «ليش ماكو كشوفات مجاي تنرفع الكشوفات هنانه؟» — كشف المشروع ينحفظ بالمشروع، وكشف الحجز
// بـBookingSurveyReport، والفورمات الورقية بـProjectChecklist. هالمسار يجمعهن.
// مدير المشاريع/الإدارة يشوف الكل؛ غيره يشوف بس الي هو مسؤول/منفّذ/كاتب بيها (والي ما يخصه = ما ينرجع).

type SurveyRow struct {
	Kind       string     `db:"kind" json:"kind"` // PROJECT | BOOKING
	ID         string     `db:"id" json:"id"`
	RefID      string     `db:"refId" json:"refId"`
	Code       *string    `db:"code" json:"code"`
	Title      *string    `db:"title" json:"title"`
	Customer   *string    `db:"customer" json:"customer"`
	Stage      *string    `db:"stage" json:"stage"`
	Location   *string    `db:"location" json:"location"`
	When       *string    `db:"when" json:"when"`
	Team       *string    `db:"team" json:"team"`
	Details    *string    `db:"details" json:"details"`
	ByName     *string    `db:"byName" json:"byName"`
	OwnerName  *string    `db:"ownerName" json:"ownerName"`
	HasAnswers bool       `db:"hasAnswers" json:"hasAnswers"`
	CreatedAt  *time.Time `db:"createdAt" json:"createdAt"`
}

type SurveysHandler struct {
	db      *sqlx.DB
	seesAll func(*http.Request) bool
}

func NewSurveysHandler(db *sqlx.DB, seesAll func(*http.Request) bool) *SurveysHandler {
	return &SurveysHandler{db: db, seesAll: seesAll}
}

const surveysSQL = `
  SELECT 'PROJECT' AS kind, p.id, p.id AS "refId", p.code, p.name AS title, p.rep AS customer, p.stage,
         p.location, p.time AS "when", p.staff AS team, p.task AS details,
         sv.name AS "byName", rs.name AS "ownerName",
         (p.survey IS NOT NULL AND p.survey::text NOT IN ('null','[]','{}')) AS "hasAnswers",
         p."updatedAt" AS "createdAt"
  FROM "Project" p
  LEFT JOIN "Employee" sv ON sv.id = p."surveyorEmployeeId"
  LEFT JOIN "Employee" rs ON rs.id = p."responsibleEmployeeId"
  WHERE (p."surveyorEmployeeId" IS NOT NULL OR (p.survey IS NOT NULL AND p.survey::text NOT IN ('null','[]','{}')) OR p.stage LIKE '%كشف%')
    AND ($1 = '' OR p."surveyorEmployeeId" = $1 OR p."responsibleEmployeeId" = $1 OR p."delegatedToEmployeeId" = $1)
  UNION ALL
  SELECT 'BOOKING', r.id, b.id, b.code, s.name, c.name, b.status::text,
         b.address, NULL, NULL,
         concat_ws(' · ', 'يريد: ' || r."customerWants", CASE WHEN r."siteAreaSqm" IS NOT NULL THEN 'المساحة: ' || r."siteAreaSqm"::text || ' م²' END, r."siteDetails", r."otherNotes"),
         e.name, NULL, true, r."createdAt"
  FROM "BookingSurveyReport" r
  JOIN "Booking" b ON b.id = r."bookingId"
  LEFT JOIN "Service" s ON s.id = b."serviceId"
  LEFT JOIN "Customer" c ON c.id = b."customerId"
  LEFT JOIN "Employee" e ON e.id = r."employeeId"
  WHERE ($1 = '' OR r."employeeId" = $1)
ORDER BY "createdAt" DESC NULLS LAST LIMIT 500`

// GET /api/surveys
func (h *SurveysHandler) List(w http.ResponseWriter, r *http.Request) {
	me := ""
	if !h.seesAll(r) {
		me = middleware.EmployeeIDFromContext(r)
	}
	rows := []SurveyRow{}
	if err := h.db.Select(&rows, surveysSQL, me); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الكشوفات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
