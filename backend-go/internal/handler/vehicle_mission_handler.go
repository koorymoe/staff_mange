package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type VehicleMissionHandler struct {
	service        *service.VehicleMissionService
	ratingService  *service.VehicleMissionRatingService
	bookingService *service.VehicleBookingService
}

func NewVehicleMissionHandler(s *service.VehicleMissionService, ratingService *service.VehicleMissionRatingService, bookingService *service.VehicleBookingService) *VehicleMissionHandler {
	return &VehicleMissionHandler{service: s, ratingService: ratingService, bookingService: bookingService}
}

// StartMissionResponse نتيجة بدء مهمة، مع تحذير اختياري لو السيارة محجوزة حالياً لموظف آخر.
type StartMissionResponse struct {
	*model.VehicleMission
	BookingWarning *string `json:"bookingWarning,omitempty"`
}

func (h *VehicleMissionHandler) CreateRating(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleMissionRatingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	actorID := middleware.EmployeeIDFromContext(r)
	rating, err := h.ratingService.CreateRating(r.PathValue("id"), actorID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, rating)
}

func (h *VehicleMissionHandler) DriverRatingSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.ratingService.GetDriverRatingSummary(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ملخص تقييم السائق")
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

func isAdminOrMonitor(role string) bool {
	return role == "ADMIN" || role == "MONITOR"
}

func (h *VehicleMissionHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req model.StartVehicleMissionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	actorID := middleware.EmployeeIDFromContext(r)
	// موظف عادي يقدر يبدأ مهمة لنفسه بس — تعيين سائق ثاني مسموح فقط للإداري/المراقب.
	if req.DriverID != nil && *req.DriverID != "" && *req.DriverID != actorID && !isAdminOrMonitor(middleware.RoleFromContext(r)) {
		WriteError(w, http.StatusForbidden, "لا تملك صلاحية بدء مهمة نيابة عن موظف آخر")
		return
	}
	mission, err := h.service.StartMission(actorID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp := StartMissionResponse{VehicleMission: mission}
	if h.bookingService != nil {
		if warning, wErr := h.bookingService.CheckApprovedBookingConflict(mission.VehicleID, mission.DriverID); wErr == nil && warning != "" {
			resp.BookingWarning = &warning
		}
	}
	WriteJSON(w, http.StatusCreated, resp)
}

func (h *VehicleMissionHandler) End(w http.ResponseWriter, r *http.Request) {
	var req model.EndVehicleMissionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	missionID := r.PathValue("id")
	mission, err := h.service.Get(missionID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "المهمة غير موجودة")
		return
	}
	actorID := middleware.EmployeeIDFromContext(r)
	if mission.DriverID != actorID && !isAdminOrMonitor(middleware.RoleFromContext(r)) {
		WriteError(w, http.StatusForbidden, "لا تملك صلاحية إنهاء هذه المهمة")
		return
	}
	updated, err := h.service.EndMission(missionID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, updated)
}

func (h *VehicleMissionHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := model.VehicleMissionFilters{}
	q := r.URL.Query()
	if v := q.Get("vehicleId"); v != "" {
		filters.VehicleID = &v
	}
	if v := q.Get("driverId"); v != "" {
		filters.DriverID = &v
	}
	if v := q.Get("status"); v != "" {
		filters.Status = &v
	}
	if v := q.Get("from"); v != "" {
		filters.From = &v
	}
	if v := q.Get("to"); v != "" {
		filters.To = &v
	}
	missions, err := h.service.List(filters)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة المهام")
		return
	}
	if h.ratingService != nil {
		for i := range missions {
			if missions[i].Status == "COMPLETED" {
				missions[i].Rating, _ = h.ratingService.GetByMission(missions[i].ID)
			}
		}
	}
	WriteJSON(w, http.StatusOK, missions)
}

func (h *VehicleMissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	mission, err := h.service.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "المهمة غير موجودة")
		return
	}
	if h.ratingService != nil && mission.Status == "COMPLETED" {
		mission.Rating, _ = h.ratingService.GetByMission(mission.ID)
	}
	WriteJSON(w, http.StatusOK, mission)
}
