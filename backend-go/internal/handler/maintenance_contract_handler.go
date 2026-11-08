package handler

import (
	"log"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// 🔄 عقود الصيانة الدورية (ع) 11-08 — المدير والإداري المنسّق.
type MaintenanceContractHandler struct {
	svc *service.MaintenanceContractService
}

func NewMaintenanceContractHandler(s *service.MaintenanceContractService) *MaintenanceContractHandler {
	return &MaintenanceContractHandler{svc: s}
}

func (h *MaintenanceContractHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List()
	if err != nil {
		log.Printf("contracts: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب العقود")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *MaintenanceContractHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in service.ContractInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	id, err := h.svc.Create(middleware.EmployeeIDFromContext(r), in)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *MaintenanceContractHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in service.ContractInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.Update(r.PathValue("id"), in); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
