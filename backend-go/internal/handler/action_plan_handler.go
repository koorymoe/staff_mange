package handler

// ═══ معالجة الحالات الحرجة (ك) 11-28 زينا ═══
// موظف عنده مشكلة بالشغل (بطء، دقة، تضييع وقت…) — جاي من ماتركس أو من المدير/المالك.
// الخطة: الهدف، والخطوات، والمدة. وبنهاية المدة: تعالجت؟ وشنو الإجراء التالي.
// داخلية: للمالك ومدير النظام والي عنده صلاحية action_plan (حزمة «الإدارة») — الموظف ما يشوفها.
// ⚠️ ماكو غرامات فلوس هنا: الإجراء التالي نص يكتبه المدير.

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

type ActionPlanHandler struct{ db *sqlx.DB }

func NewActionPlanHandler(db *sqlx.DB) *ActionPlanHandler { return &ActionPlanHandler{db: db} }

type ActionPlanRow struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	EmployeeName string     `db:"employeeName" json:"employeeName"`
	Position     *string    `db:"position" json:"position"`
	Problem      string     `db:"problem" json:"problem"`
	Details      string     `db:"details" json:"details"`
	Source       string     `db:"source" json:"source"`
	Target       string     `db:"target" json:"target"`
	Steps        string     `db:"steps" json:"steps"`
	StartDate    string     `db:"startDate" json:"startDate"`
	DueDate      string     `db:"dueDate" json:"dueDate"`
	DaysLeft     int        `db:"daysLeft" json:"daysLeft"`
	Status       string     `db:"status" json:"status"`
	Outcome      *string    `db:"outcome" json:"outcome"`
	NextAction   *string    `db:"nextAction" json:"nextAction"`
	CreatedBy    string     `db:"createdBy" json:"createdBy"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
	ClosedBy     *string    `db:"closedBy" json:"closedBy"`
	ClosedAt     *time.Time `db:"closedAt" json:"closedAt"`
	Notes        int        `db:"notes" json:"notes"`
}

type ActionPlanNote struct {
	ID        string    `db:"id" json:"id"`
	Note      string    `db:"note" json:"note"`
	By        string    `db:"by" json:"by"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

const actionPlanSelect = `
	SELECT p.id, p."employeeId", e.name AS "employeeName", e."jobTitle" AS position, p.problem, p.details, p.source,
	       p.target, p.steps, p."startDate"::text AS "startDate", p."dueDate"::text AS "dueDate",
	       (p."dueDate" - baghdad_today())::int AS "daysLeft", p.status, p.outcome, p."nextAction",
	       c.name AS "createdBy", p."createdAt", cl.name AS "closedBy", p."closedAt",
	       (SELECT count(*) FROM "ActionPlanNote" n WHERE n."planId" = p.id)::int AS notes
	FROM "ActionPlan" p
	JOIN "Employee" e ON e.id = p."employeeId"
	JOIN "Employee" c ON c.id = p."createdById"
	LEFT JOIN "Employee" cl ON cl.id = p."closedById"`

var actionPlanProblems = map[string]bool{"SLOW": true, "ACCURACY": true, "TIME_WASTE": true, "BEHAVIOR": true, "OTHER": true}
var actionPlanSources = map[string]bool{"MATRIX": true, "MANAGER": true, "OWNER": true}

// GET /api/action-plans — كل الخطط (المفتوحة أول، والي خلصت مدتها فوگ).
func (h *ActionPlanHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []ActionPlanRow{}
	if err := h.db.Select(&rows, actionPlanSelect+`
		ORDER BY (p.status = 'OPEN') DESC, p."dueDate" ASC, p."createdAt" DESC`); err != nil {
		log.Printf("action plans: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الخطط")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/action-plans/{id}/notes — متابعات الخطة.
func (h *ActionPlanHandler) Notes(w http.ResponseWriter, r *http.Request) {
	rows := []ActionPlanNote{}
	if err := h.db.Select(&rows, `SELECT n.id, n.note, e.name AS by, n."createdAt" FROM "ActionPlanNote" n
		JOIN "Employee" e ON e.id = n."byId" WHERE n."planId" = $1 ORDER BY n."createdAt"`, r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المتابعات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/action-plans — خطة جديدة.
func (h *ActionPlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		EmployeeID string `json:"employeeId"`
		Problem    string `json:"problem"`
		Details    string `json:"details"`
		Source     string `json:"source"`
		Target     string `json:"target"`
		Steps      string `json:"steps"`
		DueDate    string `json:"dueDate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	in.Target = strings.TrimSpace(in.Target)
	switch {
	case in.EmployeeID == "":
		WriteError(w, http.StatusBadRequest, "اختار الموظف")
		return
	case !actionPlanProblems[in.Problem]:
		WriteError(w, http.StatusBadRequest, "اختار نوع المشكلة")
		return
	case !actionPlanSources[in.Source]:
		WriteError(w, http.StatusBadRequest, "اختار منين جاي التقرير")
		return
	case in.Target == "":
		WriteError(w, http.StatusBadRequest, "اكتب الهدف المطلوب")
		return
	}
	due, err := time.Parse("2006-01-02", in.DueDate)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "اختار نهاية المدة")
		return
	}
	var id string
	err = h.db.Get(&id, `INSERT INTO "ActionPlan" (id, "employeeId", problem, details, source, target, steps, "startDate", "dueDate", "createdById")
		SELECT gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, baghdad_today(), $7::date, $8
		WHERE $7::date >= baghdad_today()
		  AND EXISTS (SELECT 1 FROM "Employee" e WHERE e.id = $1 AND e.status = 'ACTIVE' AND e.role NOT IN ('ADMIN', 'OWNER'))
		RETURNING id`, in.EmployeeID, in.Problem, strings.TrimSpace(in.Details), in.Source, in.Target, strings.TrimSpace(in.Steps),
		due.Format("2006-01-02"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ما انحفظت: المدة لازم تنتهي اليوم أو بعده، والموظف فعّال (المدير والمالك ما تنفتحلهم خطط)")
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// POST /api/action-plans/{id}/notes — متابعة.
func (h *ActionPlanHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Note string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if strings.TrimSpace(in.Note) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب المتابعة")
		return
	}
	res, err := h.db.Exec(`INSERT INTO "ActionPlanNote" (id, "planId", note, "byId")
		SELECT gen_random_uuid()::text, id, $2, $3 FROM "ActionPlan" WHERE id = $1`,
		r.PathValue("id"), strings.TrimSpace(in.Note), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		WriteError(w, http.StatusNotFound, "الخطة مو موجودة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/action-plans/{id}/extend — تمديد المدة (يتسجّل متابعة بالسبب).
func (h *ActionPlanHandler) Extend(w http.ResponseWriter, r *http.Request) {
	var in struct {
		DueDate string `json:"dueDate"`
		Reason  string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if _, err := time.Parse("2006-01-02", in.DueDate); err != nil || strings.TrimSpace(in.Reason) == "" {
		WriteError(w, http.StatusBadRequest, "اختار التاريخ الجديد واكتب السبب")
		return
	}
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	defer tx.Rollback() //nolint:errcheck
	var old string
	if err := tx.Get(&old, `UPDATE "ActionPlan" p SET "dueDate" = $2::date FROM (SELECT "dueDate" FROM "ActionPlan" WHERE id = $1) o
		WHERE p.id = $1 AND p.status = 'OPEN' AND $2::date > p."dueDate" RETURNING o."dueDate"::text`, r.PathValue("id"), in.DueDate); err != nil {
		WriteError(w, http.StatusBadRequest, "التمديد للخطة المفتوحة بس، والتاريخ الجديد بعد القديم")
		return
	}
	if _, err := tx.Exec(`INSERT INTO "ActionPlanNote" (id, "planId", note, "byId") VALUES (gen_random_uuid()::text, $1, $2, $3)`,
		r.PathValue("id"), "⏳ تمديد من "+old+" إلى "+in.DueDate+" — "+strings.TrimSpace(in.Reason), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/action-plans/{id}/close — النتيجة: تعالجت أو لا، والإجراء التالي.
func (h *ActionPlanHandler) Close(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status     string `json:"status"`
		Outcome    string `json:"outcome"`
		NextAction string `json:"nextAction"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.Status != "RESOLVED" && in.Status != "NOT_RESOLVED" {
		WriteError(w, http.StatusBadRequest, "اختار: تعالجت أو ما تعالجت")
		return
	}
	if strings.TrimSpace(in.Outcome) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب شنو صار")
		return
	}
	if in.Status == "NOT_RESOLVED" && strings.TrimSpace(in.NextAction) == "" {
		WriteError(w, http.StatusBadRequest, "ما تعالجت — اكتب الإجراء التالي")
		return
	}
	res, err := h.db.Exec(`UPDATE "ActionPlan" SET status = $2, outcome = $3, "nextAction" = NULLIF($4, ''),
		"closedById" = $5, "closedAt" = now() WHERE id = $1 AND status = 'OPEN'`,
		r.PathValue("id"), in.Status, strings.TrimSpace(in.Outcome), strings.TrimSpace(in.NextAction), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		WriteError(w, http.StatusBadRequest, "الخطة مسكّرة من قبل")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/action-plans/employees — قائمة الموظفين لاختيار صاحب الخطة (بلا ما يحتاج صلاحية الموظفين).
func (h *ActionPlanHandler) Employees(w http.ResponseWriter, r *http.Request) {
	rows := []struct {
		ID       string  `db:"id" json:"id"`
		Name     string  `db:"name" json:"name"`
		JobTitle *string `db:"jobTitle" json:"jobTitle"`
	}{}
	if err := h.db.Select(&rows, `SELECT id, name, "jobTitle" FROM "Employee"
		WHERE status = 'ACTIVE' AND role NOT IN ('ADMIN', 'OWNER') ORDER BY name`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الموظفين")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
