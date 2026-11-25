package handler

import (
	"errors"
	"net/http"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ 🧾 سندات الصرف (ل) 11-25 ═══

type VoucherHandler struct{ svc *service.Vouchers }

func NewVoucherHandler(s *service.Vouchers) *VoucherHandler { return &VoucherHandler{svc: s} }

// GET /api/vouchers?month=YYYY-MM
func (h *VoucherHandler) List(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.List(r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// POST /api/vouchers
func (h *VoucherHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in service.VoucherInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	v, err := h.svc.Create(middleware.EmployeeIDFromContext(r), in, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, v)
}

// POST /api/vouchers/{id}/cancel {reason}
func (h *VoucherHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reason string `json:"reason"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.Cancel(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.Reason); err != nil {
		if errors.Is(err, service.ErrDesignHidden) {
			WriteError(w, http.StatusNotFound, "مو موجود أو ملغي")
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
