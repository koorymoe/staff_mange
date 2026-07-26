package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type TeamInventoryCheckHandler struct {
	service *service.TeamInventoryCheckService
}

func NewTeamInventoryCheckHandler(s *service.TeamInventoryCheckService) *TeamInventoryCheckHandler {
	return &TeamInventoryCheckHandler{service: s}
}

func (h *TeamInventoryCheckHandler) ListTools(w http.ResponseWriter, r *http.Request) {
	tools, err := h.service.ListTools()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة الأدوات المطلوبة")
		return
	}
	WriteJSON(w, http.StatusOK, tools)
}

func (h *TeamInventoryCheckHandler) CreateTool(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTeamInventoryToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.CreateTool(req.Name)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, tool)
}

func (h *TeamInventoryCheckHandler) List(w http.ResponseWriter, r *http.Request) {
	checks, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب جلسات الجرد")
		return
	}
	WriteJSON(w, http.StatusOK, checks)
}

func (h *TeamInventoryCheckHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTeamInventoryCheckRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	check, err := h.service.Create(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, check)
}
