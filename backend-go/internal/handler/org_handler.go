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
