package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type GpsHandler struct {
	service *service.GpsService
}

func NewGpsHandler(s *service.GpsService) *GpsHandler {
	return &GpsHandler{service: s}
}

// ── GPS Customers ────────────────────────────────────────────────────────────

func (h *GpsHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := h.service.ListCustomers()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب عملاء GPS")
		return
	}
	WriteJSON(w, http.StatusOK, customers)
}

func (h *GpsHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsCustomerRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	customer, err := h.service.CreateCustomer(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, customer)
}

func (h *GpsHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsCustomerRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	customer, err := h.service.UpdateCustomer(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, customer)
}

// ── SIM Cards ────────────────────────────────────────────────────────────────

func (h *GpsHandler) ListSims(w http.ResponseWriter, r *http.Request) {
	sims, err := h.service.ListSims()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب شرائح SIM")
		return
	}
	WriteJSON(w, http.StatusOK, sims)
}

func (h *GpsHandler) CreateSim(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertSimCardRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	sim, err := h.service.CreateSim(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, sim)
}

func (h *GpsHandler) UpdateSim(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertSimCardRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	sim, err := h.service.UpdateSim(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, sim)
}

// ── Device Requests ──────────────────────────────────────────────────────────

func (h *GpsHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.service.ListDevices()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الأجهزة")
		return
	}
	WriteJSON(w, http.StatusOK, devices)
}

func (h *GpsHandler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsDeviceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	device, err := h.service.CreateDevice(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, device)
}

func (h *GpsHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsDeviceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	device, err := h.service.UpdateDevice(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, device)
}

// ── Renewals ─────────────────────────────────────────────────────────────────

func (h *GpsHandler) ListRenewals(w http.ResponseWriter, r *http.Request) {
	renewals, err := h.service.ListRenewals()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات التجديد")
		return
	}
	WriteJSON(w, http.StatusOK, renewals)
}

func (h *GpsHandler) CreateRenewal(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsRenewalRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	renewal, err := h.service.CreateRenewal(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, renewal)
}

func (h *GpsHandler) UpdateRenewal(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsRenewalRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	renewal, err := h.service.UpdateRenewal(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, renewal)
}

// ── Maintenance ──────────────────────────────────────────────────────────────

func (h *GpsHandler) ListMaintenance(w http.ResponseWriter, r *http.Request) {
	records, err := h.service.ListMaintenance()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الصيانة")
		return
	}
	WriteJSON(w, http.StatusOK, records)
}

func (h *GpsHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsMaintenanceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	record, err := h.service.CreateMaintenance(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, record)
}

func (h *GpsHandler) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsMaintenanceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	record, err := h.service.UpdateMaintenance(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, record)
}

// ── Settings (prices) ────────────────────────────────────────────────────────

func (h *GpsHandler) ListSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.ListSettings()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب إعدادات الأسعار")
		return
	}
	WriteJSON(w, http.StatusOK, settings)
}

func (h *GpsHandler) UpsertSettings(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertGpsSubscriptionPriceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	settings, err := h.service.UpsertSetting(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, settings)
}

// ── Stats ────────────────────────────────────────────────────────────────────

func (h *GpsHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Stats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب إحصائيات GPS")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// ── دورة حياة الشريحة ومتابعة التجديد ────────────────────────────────────────

// GET /api/gps/sims/available — الشرائح المتوفرة للربط بزبون جديد
func (h *GpsHandler) ListAvailableSims(w http.ResponseWriter, r *http.Request) {
	sims, err := h.service.ListAvailableSims()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الشرائح المتوفرة")
		return
	}
	WriteJSON(w, http.StatusOK, sims)
}

// POST /api/gps/sims/{id}/assign — ربط شريحة متوفرة بزبون
func (h *GpsHandler) AssignSim(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customerId"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	sim, err := h.service.AssignSim(r.PathValue("id"), req.CustomerID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, sim)
}

// POST /api/gps/sims/{id}/release — تحرير الشريحة وإرجاعها للمتوفر
func (h *GpsHandler) ReleaseSim(w http.ResponseWriter, r *http.Request) {
	sim, err := h.service.ReleaseSim(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تحرير الشريحة")
		return
	}
	WriteJSON(w, http.StatusOK, sim)
}

// POST /api/gps/sims/{id}/burn — تأشير إن الشريحة انحرقت
func (h *GpsHandler) BurnSim(w http.ResponseWriter, r *http.Request) {
	sim, err := h.service.BurnSim(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تأشير حرق الشريحة")
		return
	}
	WriteJSON(w, http.StatusOK, sim)
}

// GET /api/gps/subscriptions/follow-up — قائمة متابعة الاشتراكات المنتهية
func (h *GpsHandler) SubscriptionFollowUps(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.SubscriptionFollowUps()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة المتابعة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/gps/devices/{id}/follow-up — مهندس الجودة يسجّل نتيجة الاتصال
func (h *GpsHandler) CreateFollowUp(w http.ResponseWriter, r *http.Request) {
	var req model.CreateFollowUpRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	var by *string
	if id := middleware.EmployeeIDFromContext(r); id != "" {
		by = &id
	}
	f, err := h.service.CreateFollowUp(r.PathValue("id"), by, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, f)
}

// GET /api/gps/devices/{id}/follow-up — سجل الاتصالات على جهاز
func (h *GpsHandler) ListFollowUps(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.ListFollowUps(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل الاتصالات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
