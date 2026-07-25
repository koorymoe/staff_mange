package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type VehicleMissionHandler struct {
	service *service.VehicleMissionService
}

func NewVehicleMissionHandler(s *service.VehicleMissionService) *VehicleMissionHandler {
	return &VehicleMissionHandler{service: s}
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
	WriteJSON(w, http.StatusCreated, mission)
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
	WriteJSON(w, http.StatusOK, missions)
}

func (h *VehicleMissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	mission, err := h.service.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "المهمة غير موجودة")
		return
	}
	WriteJSON(w, http.StatusOK, mission)
}
