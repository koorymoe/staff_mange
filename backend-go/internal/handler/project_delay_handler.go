package handler

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ تأخير المشاريع (قرار (ع) 10-10) ═══
// حدود المراحل (ماتركس يتعلّم + المالك يثبّت)، وسبب التأخير، وشاشة المراقب.
type ProjectDelayHandler struct {
	db  *sqlx.DB
	svc *service.MatrixProjectService
}

func NewProjectDelayHandler(db *sqlx.DB, svc *service.MatrixProjectService) *ProjectDelayHandler {
	return &ProjectDelayHandler{db: db, svc: svc}
}

// المراحل الي إلها حد (المكتمل والمرفوض ما إلهن)
var projectStages = []string{"1. اتصال بالزبون", "2. مرحلة الكشف", "3. عرض السعر", "4. العقد", "📸 الإعلام", "5. البدء بالتنفيذ"}

// GET /api/projects/stage-limits
func (h *ProjectDelayHandler) StageLimits(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.svc.StageLimits(projectStages))
}

// PUT /api/projects/stage-limits {stage, days} — days فاضي/صفر = رجّع لماتركس
func (h *ProjectDelayHandler) SetStageLimit(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Stage string `json:"stage"`
		Days  *int   `json:"days"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	ok := false
	for _, s := range projectStages {
		if s == in.Stage {
			ok = true
		}
	}
	if !ok {
		WriteError(w, http.StatusBadRequest, "مرحلة غير معروفة")
		return
	}
	var err error
	if in.Days == nil || *in.Days <= 0 {
		_, err = h.db.Exec(`DELETE FROM "ProjectStageLimit" WHERE stage = $1`, in.Stage)
	} else {
		if *in.Days > 365 {
			WriteError(w, http.StatusBadRequest, "الحد لازم يكون أقل من سنة")
			return
		}
		_, err = h.db.Exec(`INSERT INTO "ProjectStageLimit" (stage, days, "updatedById") VALUES ($1, $2, $3)
			ON CONFLICT (stage) DO UPDATE SET days = EXCLUDED.days, "updatedById" = EXCLUDED."updatedById", "updatedAt" = now()`,
			in.Stage, *in.Days, middleware.EmployeeIDFromContext(r))
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	h.StageLimits(w, r)
}

// POST /api/projects/{id}/delay-reason {reason}
// يكتبه المرتبط بالمشروع (المسؤول، المتوجّه إله، الكاشف، المضيف، مشرف الحجز)
// أو المدير أو صاحب إدارة المشاريع.
func (h *ProjectDelayHandler) AddReason(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reason string `json:"reason"`
	}
	if err := DecodeJSON(r, &in); err != nil || len([]rune(strings.TrimSpace(in.Reason))) < 5 {
		WriteError(w, http.StatusBadRequest, "اكتب السبب بوضوح")
		return
	}
	id, me := r.PathValue("id"), middleware.EmployeeIDFromContext(r)
	role := middleware.RoleFromContext(r)
	var allowed bool
	_ = h.db.Get(&allowed, `SELECT EXISTS (SELECT 1 FROM "Project" p WHERE p.id = $1 AND (
		$2 IN (COALESCE(p."responsibleEmployeeId",''), COALESCE(p."delegatedToEmployeeId",''), COALESCE(p."surveyorEmployeeId",''), COALESCE(p."createdByEmployeeId",''))
		OR EXISTS (SELECT 1 FROM "Booking" b WHERE b.id = p."bookingId" AND b."projectSupervisorId" = $2)
		OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" pm ON pm.id = ep."permissionId"
		           WHERE ep."employeeId" = $2 AND pm.name = 'project_management')))`, id, me)
	if !allowed && role != "ADMIN" && role != "OWNER" {
		WriteError(w, http.StatusForbidden, "ما عندك علاقة بهالمشروع")
		return
	}
	if _, err := h.db.Exec(`INSERT INTO "ProjectDelayReason" (id, "projectId", stage, "employeeId", reason)
		SELECT gen_random_uuid()::text, p.id, p.stage, $2, $3 FROM "Project" p WHERE p.id = $1`,
		id, me, strings.TrimSpace(in.Reason)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/projects/delays — المشاريع المفتوحة بحكم ماتركس (للمراقب والمدير)
func (h *ProjectDelayHandler) Delays(w http.ResponseWriter, r *http.Request) {
	rep, err := h.svc.Report()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تحليل المشاريع")
		return
	}
	out := []service.ProjectVerdict{}
	for _, p := range rep.Projects {
		if p.Status != service.ChainNA {
			out = append(out, p)
		}
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/projects/my-delays — مشاريعي المتجاوزة (لشريط «اكتب سبب التأخير»).
// المدير وصاحب إدارة المشاريع يشوفون كل المتجاوز.
func (h *ProjectDelayHandler) MyDelays(w http.ResponseWriter, r *http.Request) {
	rep, err := h.svc.Report()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تحليل المشاريع")
		return
	}
	me, role := middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r)
	all := role == "ADMIN" || role == "OWNER"
	if !all {
		_ = h.db.Get(&all, `SELECT EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" pm ON pm.id = ep."permissionId"
			WHERE ep."employeeId" = $1 AND pm.name = 'project_management')`, me)
	}
	out := []service.ProjectVerdict{}
	for _, p := range rep.Projects {
		if p.OverLimit && p.Status != service.ChainNA && (all || p.OwnerID == me) {
			out = append(out, p)
		}
	}
	WriteJSON(w, http.StatusOK, out)
}
