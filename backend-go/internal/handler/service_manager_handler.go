package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ServiceManagerHandler struct {
	repo *repository.ServiceManagerRepository
}

func NewServiceManagerHandler(repo *repository.ServiceManagerRepository) *ServiceManagerHandler {
	return &ServiceManagerHandler{repo: repo}
}

func (h *ServiceManagerHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مسؤولي الخدمات")
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

// Set — الأدمن يحدد مجموعة الخدمات اللي موظف معين مسؤول عنها (يستبدل القديم بالكامل)
func (h *ServiceManagerHandler) Set(w http.ResponseWriter, r *http.Request) {
	var req model.SetServiceManagerRequest
	if err := DecodeJSON(r, &req); err != nil || req.EmployeeID == "" {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if err := h.repo.SetForEmployee(req.EmployeeID, req.ServiceIDs); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ مسؤوليات الخدمة")
		return
	}
	list, err := h.repo.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مسؤولي الخدمات")
		return
	}
	WriteJSON(w, http.StatusOK, list)
}
