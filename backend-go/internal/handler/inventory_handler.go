package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type InventoryHandler struct {
	service *service.InventoryService
}

func NewInventoryHandler(s *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: s}
}

// ── Inventory Checks ──────────────────────────────────────────────────────────

func (h *InventoryHandler) CreateInventoryCheck(w http.ResponseWriter, r *http.Request) {
	var req model.CreateInventoryCheckRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	check, err := h.service.CreateInventoryCheck(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, check)
}

func (h *InventoryHandler) TodaysInventoryChecks(w http.ResponseWriter, r *http.Request) {
	checks, err := h.service.TodaysInventoryChecks()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب نتائج الجرد")
		return
	}
	WriteJSON(w, http.StatusOK, checks)
}

// ── Personal Tools ──────────────────────────────────────────────────────────

func (h *InventoryHandler) ListPersonalTools(w http.ResponseWriter, r *http.Request) {
	tools, err := h.service.ListPersonalTools(r.URL.Query().Get("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأدوات الشخصية")
		return
	}
	WriteJSON(w, http.StatusOK, tools)
}

func (h *InventoryHandler) CreatePersonalTool(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePersonalToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.CreatePersonalTool(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, tool)
}

func (h *InventoryHandler) UpdatePersonalTool(w http.ResponseWriter, r *http.Request) {
	var req model.UpdatePersonalToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.UpdatePersonalTool(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tool)
}

func (h *InventoryHandler) DeletePersonalTool(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeletePersonalTool(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ── Vehicle Tools ───────────────────────────────────────────────────────────

func (h *InventoryHandler) ListVehicleTools(w http.ResponseWriter, r *http.Request) {
	tools, err := h.service.ListVehicleTools(r.URL.Query().Get("vehicleId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أدوات المركبة")
		return
	}
	WriteJSON(w, http.StatusOK, tools)
}

func (h *InventoryHandler) CreateVehicleTool(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.CreateVehicleTool(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, tool)
}

func (h *InventoryHandler) UpdateVehicleTool(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateVehicleToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.UpdateVehicleTool(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tool)
}

func (h *InventoryHandler) DeleteVehicleTool(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteVehicleTool(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ── On-demand Tools ─────────────────────────────────────────────────────────

func (h *InventoryHandler) ListOnDemandTools(w http.ResponseWriter, r *http.Request) {
	tools, err := h.service.ListOnDemandTools()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأدوات المشتركة")
		return
	}
	WriteJSON(w, http.StatusOK, tools)
}

func (h *InventoryHandler) CreateOnDemandTool(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOnDemandToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.CreateOnDemandTool(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, tool)
}

func (h *InventoryHandler) UpdateOnDemandTool(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateOnDemandToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.UpdateOnDemandTool(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tool)
}

// ── Tool Requests ───────────────────────────────────────────────────────────

func (h *InventoryHandler) ListToolRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.service.ListToolRequests(r.URL.Query().Get("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الأدوات")
		return
	}
	WriteJSON(w, http.StatusOK, requests)
}

func (h *InventoryHandler) CreateToolRequest(w http.ResponseWriter, r *http.Request) {
	var req model.CreateToolRequestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	request, err := h.service.CreateToolRequest(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, request)
}

func (h *InventoryHandler) ApproveToolRequest(w http.ResponseWriter, r *http.Request) {
	var req model.ApproveToolRequestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	request, err := h.service.ApproveToolRequest(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}

func (h *InventoryHandler) RejectToolRequest(w http.ResponseWriter, r *http.Request) {
	request, err := h.service.RejectToolRequest(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}

func (h *InventoryHandler) ReturnToolRequest(w http.ResponseWriter, r *http.Request) {
	request, err := h.service.ReturnToolRequest(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}
