package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type VehicleHandler struct {
	service *service.VehicleService
}

func NewVehicleHandler(s *service.VehicleService) *VehicleHandler {
	return &VehicleHandler{service: s}
}

func (h *VehicleHandler) List(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة المركبات")
		return
	}
	WriteJSON(w, http.StatusOK, vehicles)
}

func (h *VehicleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	vehicle, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, vehicle)
}

func (h *VehicleHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.ListLogs(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجلات السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, logs)
}

func (h *VehicleHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleLogRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	log, err := h.service.CreateLog(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, log)
}

func (h *VehicleHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := h.service.ListIncidents(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أعطال/أضرار السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, incidents)
}

func (h *VehicleHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleIncidentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	incident, err := h.service.CreateIncident(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, incident)
}

func (h *VehicleHandler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateVehicleIncidentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	incident, err := h.service.UpdateIncident(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, incident)
}

func (h *VehicleHandler) ListMonthlyStatus(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.service.ListMonthlyStatus(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حالة السيارة الشهرية")
		return
	}
	WriteJSON(w, http.StatusOK, statuses)
}

func (h *VehicleHandler) SetMonthlyStatus(w http.ResponseWriter, r *http.Request) {
	var req model.SetVehicleMonthlyStatusRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	status, err := h.service.SetMonthlyStatus(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, status)
}

// ── VehicleDailyRating (تقييم يومي + جودة غسيل الفنيين) ──

func (h *VehicleHandler) CreateDailyRating(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleDailyRatingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	rating, err := h.service.CreateDailyRating(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, rating)
}

func (h *VehicleHandler) ListDailyRatings(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "1900-01-01"
	}
	ratings, err := h.service.ListDailyRatings(r.PathValue("id"), since)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تقييمات السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, ratings)
}

func (h *VehicleHandler) VehicleScoreSummaries(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "1900-01-01"
	}
	summaries, err := h.service.VehicleScoreSummaries(since)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ملخص تقييم السيارات")
		return
	}
	WriteJSON(w, http.StatusOK, summaries)
}

func (h *VehicleHandler) TechnicianWashSummaries(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "1900-01-01"
	}
	summaries, err := h.service.TechnicianWashSummaries(since)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ملخص تقييم الفنيين")
		return
	}
	WriteJSON(w, http.StatusOK, summaries)
}
