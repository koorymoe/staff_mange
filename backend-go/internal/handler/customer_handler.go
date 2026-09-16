package handler

import (
	"net/http"
	"strconv"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// GET /api/v1/customers/gps
func (h *CustomerHandler) ListGps(w http.ResponseWriter, r *http.Request) {
	customers, err := h.service.ListGpsCustomers()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة زبائن الجي بي اس")
		return
	}
	WriteJSON(w, http.StatusOK, customers)
}

// PUT /api/v1/customers/{id}
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.UpdateCustomerRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	customer, err := h.service.Update(id, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if customer == nil {
		WriteError(w, http.StatusNotFound, "الزبون غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, customer)
}

type CustomerHandler struct {
	service *service.CustomerService
}

func NewCustomerHandler(s *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: s}
}

// GET /api/v1/customers
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	// بلا وسائط يرجّع كل الزبائن متل ما كان — أي مستدعي قديم ما ينكسر.
	search := r.URL.Query().Get("search")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit > 2000 {
		limit = 2000
	}
	customers, err := h.service.Search(search, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة العملاء")
		return
	}
	WriteJSON(w, http.StatusOK, customers)
}

// GET /api/v1/customers/lookup?phone=xxx
func (h *CustomerHandler) Lookup(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")
	if phone == "" {
		WriteError(w, http.StatusBadRequest, "رقم الهاتف مطلوب")
		return
	}

	customer, err := h.service.Lookup(phone)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر البحث عن العميل")
		return
	}
	if customer == nil {
		WriteError(w, http.StatusNotFound, "العميل غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, customer)
}

// POST /api/v1/customers
func (h *CustomerHandler) FindOrCreate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCustomerRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	customer, err := h.service.FindOrCreate(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, customer)
}
