package handler

import (
	"errors"
	"fmt"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// ═══ 🏛️ الهيكل الجديد (دازاي 10-29) ═══
// الشجرة والتعيين والإنهاء للمالك ومدير النظام (فوك الكل). «مكاني» لأي موظف.

type OrgHandler struct {
	svc           *service.OrgService
	notifications *repository.NotificationRepository
}

func NewOrgHandler(svc *service.OrgService, n *repository.NotificationRepository) *OrgHandler {
	return &OrgHandler{svc: svc, notifications: n}
}

// guard: غير المالك ما يعيّن ولا ينهي منصب المالك أو مدير النظام، ولا منصبه هو.
func (h *OrgHandler) guard(r *http.Request, target string) string {
	if middleware.RoleFromContext(r) == "OWNER" {
		return ""
	}
	if target == "" {
		return ""
	}
	if target == middleware.EmployeeIDFromContext(r) {
		return "منصبك إنت يعيّنه المالك بس"
	}
	if rl := h.svc.TargetRole(target); rl == "OWNER" || rl == "ADMIN" {
		return "مناصب المالك ومدير النظام يعيّنها المالك بس"
	}
	return ""
}

var orgPositionLabel = map[string]string{"DIRECTOR": "مدير", "DEPUTY": "نائب", "MEMBER": "عضو"}

// GET /api/org/chart
func (h *OrgHandler) Chart(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Chart()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الهيكل")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/org/me
func (h *OrgHandler) Me(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Mine(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب منصبك")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// POST /api/org/assign — {employeeId, divisionId, unitId?, position, allowanceIQD, note}
func (h *OrgHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var in service.OrgAssignInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	if why := h.guard(r, in.EmployeeID); why != "" {
		WriteError(w, http.StatusForbidden, why)
		return
	}
	a, err := h.svc.Assign(middleware.EmployeeIDFromContext(r), in)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrOrgSeatTaken) {
			status = http.StatusConflict
		}
		WriteError(w, status, err.Error())
		return
	}
	if h.notifications != nil {
		_ = h.notifications.Create(a.EmployeeID, "ORG_ASSIGN",
			fmt.Sprintf("🏛️ صار عندك منصب جديد: %s — صلاحياتك تحدّثت حسب مكانك.", orgPositionLabel[a.Position]))
	}
	WriteJSON(w, http.StatusOK, a)
}

// POST /api/org/assignments/{id}/end — {reason}
func (h *OrgHandler) End(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reason string `json:"reason"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	if why := h.guard(r, h.svc.AssignmentEmployee(r.PathValue("id"))); why != "" {
		WriteError(w, http.StatusForbidden, why)
		return
	}
	a, err := h.svc.End(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.Reason)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.notifications != nil {
		_ = h.notifications.Create(a.EmployeeID, "ORG_END", "🏛️ انتهى منصبك ("+orgPositionLabel[a.Position]+") — السبب: "+in.Reason)
	}
	WriteJSON(w, http.StatusOK, a)
}

// ── 🏛️ مكتب المدير (المرحلة ٢) ──
// الي ما يحق له يشوف إدارة ياخذ 404 «مو موجود» (مو 403) — حتى ما تنسجل مخالفة على فضول.

func (h *OrgHandler) orgErr(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrOrgHidden) {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	WriteError(w, http.StatusBadRequest, err.Error())
}

func who(r *http.Request) (string, string) {
	return middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r)
}

// GET /api/org/desk/{division}?period=YYYY-MM
func (h *OrgHandler) Desk(w http.ResponseWriter, r *http.Request) {
	emp, role := who(r)
	d, err := h.svc.Desk(emp, role, r.PathValue("division"), r.URL.Query().Get("period"))
	if err != nil {
		h.orgErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, d)
}

// POST /api/org/issues
func (h *OrgHandler) AddIssue(w http.ResponseWriter, r *http.Request) {
	var in service.OrgIssueInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	emp, role := who(r)
	id, err := h.svc.AddIssue(emp, role, in)
	if err != nil {
		h.orgErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"id": id})
}

// PATCH /api/org/issues/{id} — {status, resolution}
func (h *OrgHandler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status     string `json:"status"`
		Resolution string `json:"resolution"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	emp, role := who(r)
	if err := h.svc.UpdateIssue(emp, role, r.PathValue("id"), in.Status, in.Resolution); err != nil {
		h.orgErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/org/plans — {divisionId, period?, title}
func (h *OrgHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var in struct {
		DivisionID string `json:"divisionId"`
		UnitID     string `json:"unitId"`
		Period     string `json:"period"`
		Title      string `json:"title"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	emp, role := who(r)
	id, err := h.svc.CreatePlan(emp, role, in.DivisionID, in.UnitID, in.Period, in.Title)
	if err != nil {
		h.orgErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"id": id})
}

// POST /api/org/plans/{id}/items — {kind, title, issueId?, metric?, target?}
func (h *OrgHandler) AddPlanItem(w http.ResponseWriter, r *http.Request) {
	var in service.OrgPlanItemInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	emp, role := who(r)
	id, err := h.svc.AddPlanItem(emp, role, r.PathValue("id"), in)
	if err != nil {
		h.orgErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"id": id})
}

// PATCH /api/org/plan-items/{id} — {manualProgress?, done?}
func (h *OrgHandler) UpdatePlanItem(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ManualProgress *int  `json:"manualProgress"`
		Done           *bool `json:"done"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	emp, role := who(r)
	if err := h.svc.UpdatePlanItem(emp, role, r.PathValue("id"), in.ManualProgress, in.Done); err != nil {
		h.orgErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/org/comments — {targetType: ISSUE|PLAN_ITEM, targetId, text}
func (h *OrgHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	var in struct {
		TargetType string `json:"targetType"`
		TargetID   string `json:"targetId"`
		Text       string `json:"text"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	emp, role := who(r)
	if err := h.svc.AddComment(emp, role, in.TargetType, in.TargetID, in.Text); err != nil {
		h.orgErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PUT /api/org/authority/{division} — [{decision, mode, limitIQD}] (المالك ومدير النظام)
func (h *OrgHandler) SetAuthority(w http.ResponseWriter, r *http.Request) {
	var rows []service.OrgAuthorityRow
	if err := DecodeJSON(r, &rows); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	if err := h.svc.SetAuthority(middleware.EmployeeIDFromContext(r), r.PathValue("division"), rows); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── 🏛️ تطوير الوحدة (10-30) ──

// POST /api/org/plans/{id}/submit — رئيس الوحدة/الإدارة يرفع الخطة
func (h *OrgHandler) SubmitPlan(w http.ResponseWriter, r *http.Request) {
	emp, role := who(r)
	to, err := h.svc.SubmitPlan(emp, role, r.PathValue("id"))
	if err != nil {
		h.orgErr(w, err)
		return
	}
	if h.notifications != nil {
		for _, id := range to {
			_ = h.notifications.Create(id, "ORG_PLAN_SUBMITTED", "🏛️ انرفعتلك خطة تطوير للاعتماد — بمكتب الإدارة.")
		}
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/org/plans/{id}/review — {approve, note}
func (h *OrgHandler) ReviewPlan(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	emp, role := who(r)
	author, err := h.svc.ReviewPlan(emp, role, r.PathValue("id"), in.Approve, in.Note)
	if err != nil {
		h.orgErr(w, err)
		return
	}
	if h.notifications != nil && author != "" {
		msg := "✅ خطة التطوير مالتك انعتمدت."
		if !in.Approve {
			msg = "↩️ خطة التطوير مالتك رجعت للتعديل: " + in.Note
		}
		_ = h.notifications.Create(author, "ORG_PLAN_REVIEWED", msg)
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PUT /api/org/issues/{id}/solution — {plan}
func (h *OrgHandler) SetSolution(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Plan string `json:"plan"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	emp, role := who(r)
	if err := h.svc.SetSolutionPlan(emp, role, r.PathValue("id"), in.Plan); err != nil {
		h.orgErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/org/regulations — لائحة العمل (المواد والفقرات الفعّالة) لربط المشاكل
func (h *OrgHandler) Regulations(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Regulations()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب اللائحة")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/org/my-items — بنود الخطط الي الموظف مسؤول عنها (آمن لأي موظف)
func (h *OrgHandler) MyItems(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.MyItems(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب بنودك")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}
