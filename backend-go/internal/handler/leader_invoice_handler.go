package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// LeaderInvoiceHandler يخدم فوترة الليدر: عرض الكتالوج، البحث بمواد الأرشيف
// بالكود، وإنشاء/عرض فواتير الليدر.
type LeaderInvoiceHandler struct {
	service   *service.LeaderInvoiceService
	catalog   *repository.SystemPriceCatalogRepository
	materials *repository.MaterialRepository
}

func NewLeaderInvoiceHandler(
	s *service.LeaderInvoiceService,
	catalog *repository.SystemPriceCatalogRepository,
	materials *repository.MaterialRepository,
) *LeaderInvoiceHandler {
	return &LeaderInvoiceHandler{service: s, catalog: catalog, materials: materials}
}

// GET /api/leader-invoices?employeeId=
func (h *LeaderInvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employeeId")
	invoices, err := h.service.List(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب فواتير الليدر")
		return
	}
	WriteJSON(w, http.StatusOK, invoices)
}

// GET /api/leader-invoices/{id}
func (h *LeaderInvoiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	inv, err := h.service.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// POST /api/leader-invoices
func (h *LeaderInvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLeaderInvoiceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	employeeID := middleware.EmployeeIDFromContext(r)
	inv, err := h.service.Create(employeeID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, inv)
}

// GET /api/system-price-catalog?systemName=
func (h *LeaderInvoiceHandler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	systemName := r.URL.Query().Get("systemName")
	rows, err := h.catalog.List(systemName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب كتالوج الأسعار")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/materials?code=
func (h *LeaderInvoiceHandler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	rows, err := h.materials.List(code)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المواد")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
