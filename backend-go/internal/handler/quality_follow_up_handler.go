package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type QualityFollowUpHandler struct {
	service *service.QualityFollowUpService
}

func NewQualityFollowUpHandler(s *service.QualityFollowUpService) *QualityFollowUpHandler {
	return &QualityFollowUpHandler{service: s}
}

// GET /api/quality-follow-ups
func (h *QualityFollowUpHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب متابعات الجودة")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// PUT /api/quality-follow-ups/{id}
func (h *QualityFollowUpHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateQualityFollowUpRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.Update(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
