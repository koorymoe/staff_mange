package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type QualityHandler struct {
	service *service.QualityService
}

func NewQualityHandler(s *service.QualityService) *QualityHandler {
	return &QualityHandler{service: s}
}

func (h *QualityHandler) List(w http.ResponseWriter, r *http.Request) {
	issues, err := h.service.List(r.URL.Query().Get("category"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مشاكل الجودة")
		return
	}
	WriteJSON(w, http.StatusOK, issues)
}

func (h *QualityHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateQualityIssueRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	issue, err := h.service.Create(req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, issue)
}

func (h *QualityHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateQualityIssueRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	issue, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, issue)
}
