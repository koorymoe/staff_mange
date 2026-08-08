package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ComplaintHandler struct {
	service *service.ComplaintService
}

func NewComplaintHandler(s *service.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{service: s}
}

// GET /api/v1/complaints
func (h *ComplaintHandler) List(w http.ResponseWriter, r *http.Request) {
	complaints, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الشكاوى")
		return
	}
	WriteJSON(w, http.StatusOK, complaints)
}

// GET /api/v1/complaints/stats
func (h *ComplaintHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.StatsByCustomer()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب إحصائيات الشكاوى")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// POST /api/v1/complaints
func (h *ComplaintHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateComplaintRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	complaint, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, complaint)
}

// PUT /api/v1/complaints/{id}
func (h *ComplaintHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateComplaintRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	complaint, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, complaint)
}

// PUT /api/v1/complaints/{id}/resolve
// PUT /api/complaints/{id}/contact — أي موظف يأشر إنه اتصل بالزبون
func (h *ComplaintHandler) SetContacted(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Contacted bool `json:"contacted"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	out, err := h.service.SetContacted(r.PathValue("id"), req.Contacted, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تحديث حالة الاتصال")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// PUT /api/complaints/{id}/notes — ملاحظات الزبون
func (h *ComplaintHandler) SetNotes(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Notes string `json:"notes"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	out, err := h.service.SetNotes(r.PathValue("id"), req.Notes)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حفظ الملاحظات")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

func (h *ComplaintHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	var req model.ResolveComplaintRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	complaint, err := h.service.Resolve(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, complaint)
}
