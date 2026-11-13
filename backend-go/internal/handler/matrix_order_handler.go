package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// ═══ 🎯 أوامر ماتركس — المالك والمدير بس (دازاي 11-13) ═══
// غيرهم ياخذ 404 (الرفض = 404).
type MatrixOrderHandler struct{ svc *service.MatrixOrderService }

func NewMatrixOrderHandler(svc *service.MatrixOrderService) *MatrixOrderHandler {
	return &MatrixOrderHandler{svc: svc}
}

func (h *MatrixOrderHandler) boss(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	role := middleware.RoleFromContext(r)
	if !model.IsAdminRole(role) {
		WriteError(w, http.StatusNotFound, "غير موجود")
		return "", "", false
	}
	return middleware.EmployeeIDFromContext(r), role, true
}

// GET /api/matrix/orders
func (h *MatrixOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	me, _, ok := h.boss(w, r)
	if !ok {
		return
	}
	rows, err := h.svc.List(me)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأوامر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/matrix/orders {text}
func (h *MatrixOrderHandler) Propose(w http.ResponseWriter, r *http.Request) {
	me, _, ok := h.boss(w, r)
	if !ok {
		return
	}
	var in struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	o, err := h.svc.Propose(me, in.Text)
	if errors.Is(err, service.ErrAgentOff) {
		WriteError(w, http.StatusServiceUnavailable, "ماتركس مطفي — شغّله من إعدادات الذكاء الاصطناعي")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, o)
}

// POST /api/matrix/orders/{id}/run {skip:[i…]}
func (h *MatrixOrderHandler) Run(w http.ResponseWriter, r *http.Request) {
	me, role, ok := h.boss(w, r)
	if !ok {
		return
	}
	var in struct {
		Skip []int `json:"skip"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	o, err := h.svc.Run(r.PathValue("id"), me, role, in.Skip)
	if err != nil {
		WriteError(w, http.StatusNotFound, "الأمر غير موجود أو انفّذ")
		return
	}
	WriteJSON(w, http.StatusOK, o)
}

// POST /api/matrix/orders/{id}/cancel
func (h *MatrixOrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	me, _, ok := h.boss(w, r)
	if !ok {
		return
	}
	if err := h.svc.Cancel(r.PathValue("id"), me); err != nil {
		WriteError(w, http.StatusNotFound, "الأمر غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
