package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// SolarInquiryHandler استفسارات الطاقة الشمسية (دازاي 11-12). الصلاحيات بالخدمة؛ الرفض 404.
type SolarInquiryHandler struct{ svc *service.SolarInquiryService }

func NewSolarInquiryHandler(s *service.SolarInquiryService) *SolarInquiryHandler {
	return &SolarInquiryHandler{svc: s}
}

func (h *SolarInquiryHandler) reply(w http.ResponseWriter, x any, err error, created bool) {
	if err == service.ErrInquiryHidden {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	code := http.StatusOK
	if created {
		code = http.StatusCreated
	}
	WriteJSON(w, code, x)
}

// GET /api/solar-inquiries
func (h *SolarInquiryHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, access, err := h.svc.List(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r))
	h.reply(w, map[string]any{"items": rows, "access": access}, err, false)
}

// POST /api/solar-inquiries {customerName, phone, area, note}
func (h *SolarInquiryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct{ CustomerName, Phone, Area, Note string }
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	x, err := h.svc.Create(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), in.CustomerName, in.Phone, in.Area, in.Note)
	h.reply(w, x, err, true)
}

// POST /api/solar-inquiries/{id}/claim
func (h *SolarInquiryHandler) Claim(w http.ResponseWriter, r *http.Request) {
	x, err := h.svc.Claim(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"))
	h.reply(w, x, err, false)
}

// POST /api/solar-inquiries/{id}/contact {outcome, note, quoteGiven, quoteAmount}
func (h *SolarInquiryHandler) Contact(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Outcome     string   `json:"outcome"`
		Note        string   `json:"note"`
		QuoteGiven  *bool    `json:"quoteGiven"`
		QuoteAmount *float64 `json:"quoteAmount"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	x, err := h.svc.TechContact(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"), in.Outcome, in.Note, in.QuoteGiven, in.QuoteAmount)
	h.reply(w, x, err, false)
}

// POST /api/solar-inquiries/{id}/quality
func (h *SolarInquiryHandler) Quality(w http.ResponseWriter, r *http.Request) {
	var in service.QualityInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	x, err := h.svc.QualityCheck(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"), in)
	h.reply(w, x, err, false)
}
