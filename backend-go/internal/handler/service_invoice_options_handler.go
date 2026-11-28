package handler

import (
	"errors"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ خيارات الفاتورة لكل خدمة — تجريبي، المالك ومدير النظام (ل) 11-24 ═══

type ServiceInvoiceOptionsHandler struct {
	svc *service.ServiceInvoiceOptions
}

func NewServiceInvoiceOptionsHandler(s *service.ServiceInvoiceOptions) *ServiceInvoiceOptionsHandler {
	return &ServiceInvoiceOptionsHandler{svc: s}
}

// GET /api/service-invoice-options
func (h *ServiceInvoiceOptionsHandler) List(w http.ResponseWriter, r *http.Request) {
	ts, err := h.svc.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, ts)
}

// PUT /api/service-invoice-options/{serviceId}  {minWork, options:[{label,unit,unitPrice}]}
func (h *ServiceInvoiceOptionsHandler) Save(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MinWork  float64                        `json:"minWork"`
		MinTiers []float64                      `json:"minTiers"`
		Options  []service.ServiceInvoiceOption `json:"options"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.Save(r.PathValue("serviceId"), middleware.EmployeeIDFromContext(r), in.MinWork, in.MinTiers, in.Options); err != nil {
		if errors.Is(err, service.ErrDesignHidden) {
			WriteError(w, http.StatusNotFound, "الخدمة مو موجودة")
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.List(w, r)
}

// POST /api/service-invoice-options/{serviceId}/quote  {qty:{optionId:n}, materials}
func (h *ServiceInvoiceOptionsHandler) Quote(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Qty       map[string]float64 `json:"qty"`
		Materials float64            `json:"materials"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	q, err := h.svc.Quote(r.PathValue("serviceId"), in.Qty, in.Materials)
	if err != nil {
		WriteError(w, http.StatusNotFound, "الخدمة مو موجودة")
		return
	}
	WriteJSON(w, http.StatusOK, q)
}
