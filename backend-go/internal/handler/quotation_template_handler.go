package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// 📋 نماذج عروض الأسعار (ل) 11-10 — نفس أدوات ماتركس بس يدوي.
type QuotationTemplateHandler struct {
	svc *service.QuotationTemplateService
}

func NewQuotationTemplateHandler(s *service.QuotationTemplateService) *QuotationTemplateHandler {
	return &QuotationTemplateHandler{svc: s}
}

func (h *QuotationTemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.svc.CanUse(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r)) {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	rows, err := h.svc.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/quotation-templates {quotationId, name, styleNotes}
func (h *QuotationTemplateHandler) Save(w http.ResponseWriter, r *http.Request) {
	var in struct {
		QuotationID string `json:"quotationId"`
		Name        string `json:"name"`
		StyleNotes  string `json:"styleNotes"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	id, sum, err := h.svc.Save(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), in.QuotationID, in.Name, in.StyleNotes)
	if err == service.ErrDesignHidden {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id, "summary": sum})
}

// POST /api/quotation-templates/{id}/use — عرض جديد من النموذج
func (h *QuotationTemplateHandler) Use(w http.ResponseWriter, r *http.Request) {
	var in service.TemplateUse
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	q, err := h.svc.Use(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"), in)
	if err == service.ErrDesignHidden {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, q)
}
