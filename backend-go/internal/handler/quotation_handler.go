package handler

import (
	"net/http"
	"strconv"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type QuotationHandler struct {
	service *service.QuotationService
}

func NewQuotationHandler(s *service.QuotationService) *QuotationHandler {
	return &QuotationHandler{service: s}
}

// GET /api/quotations?search=...
func (h *QuotationHandler) List(w http.ResponseWriter, r *http.Request) {
	// ?page= — القائمة المقسّمة الخفيفة (بلا بنود ولا صور).
	if p := r.URL.Query().Get("page"); p != "" {
		page, _ := strconv.Atoi(p)
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 50
		}
		res, err := h.service.ListPage(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.URL.Query().Get("search"), page, limit)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر جلب عروض الأسعار")
			return
		}
		WriteJSON(w, http.StatusOK, res)
		return
	}
	quotations, err := h.service.List(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.URL.Query().Get("search"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب عروض الأسعار")
		return
	}
	WriteJSON(w, http.StatusOK, quotations)
}

// GET /api/quotations/{id}
func (h *QuotationHandler) Get(w http.ResponseWriter, r *http.Request) {
	quotation, err := h.service.Get(r.PathValue("id"), middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, quotation)
}

// GET /api/quotation-by-project/{projectId} — عرض المشروع أو null
func (h *QuotationHandler) ByProject(w http.ResponseWriter, r *http.Request) {
	q, err := h.service.ByProject(r.PathValue("projectId"), middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, q)
}

// POST /api/quotations
func (h *QuotationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateQuotationRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	quotation, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, quotation)
}

// PUT /api/quotations/{id}
func (h *QuotationHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateQuotationRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	quotation, err := h.service.Update(r.PathValue("id"), middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, quotation)
}

// DELETE /api/quotations/{id}
func (h *QuotationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GET /api/quotations/{id}/versions — النسخ القديمة المؤرشفة
func (h *QuotationHandler) Versions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.Versions(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب النسخ القديمة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
