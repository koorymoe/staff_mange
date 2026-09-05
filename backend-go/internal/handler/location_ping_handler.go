package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type LocationPingHandler struct {
	repo *repository.LocationPingRepository
}

func NewLocationPingHandler(repo *repository.LocationPingRepository) *LocationPingHandler {
	return &LocationPingHandler{repo: repo}
}

// Create — الفني (نفسه بس، من رمز الدخول) يرسل نقطة موقعه الحالية
func (h *LocationPingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLocationPingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if req.Latitude == 0 && req.Longitude == 0 {
		WriteError(w, http.StatusBadRequest, "إحداثيات غير صحيحة")
		return
	}
	ping, err := h.repo.Create(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تسجيل الموقع")
		return
	}
	WriteJSON(w, http.StatusCreated, ping)
}

// Latest — لوحة متابعة الفرق الميدانية: مين طالع الحين وآخر موقع له
func (h *LocationPingHandler) Latest(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.LatestPerEmployee()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المواقع")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// Path — مسار موظف معين (?employeeId=&bookingId=) لرسمه كخط على الخريطة
func (h *LocationPingHandler) Path(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employeeId")
	if employeeID == "" {
		WriteError(w, http.StatusBadRequest, "employeeId مطلوب")
		return
	}
	var bookingID *string
	if v := r.URL.Query().Get("bookingId"); v != "" {
		bookingID = &v
	}
	rows, err := h.repo.Path(employeeID, bookingID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المسار")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
