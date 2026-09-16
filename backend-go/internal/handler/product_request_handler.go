package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ProductRequestHandler struct {
	service *service.ProductRequestService
}

func NewProductRequestHandler(s *service.ProductRequestService) *ProductRequestHandler {
	return &ProductRequestHandler{service: s}
}

func (h *ProductRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات المنتجات")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *ProductRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductProposalRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.Create(req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *ProductRequestHandler) Approve(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Approve(r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		// لو الموافقة نجحت وفشل إنشاء المنتج بس، نرجّع نص الخطأ الحقيقي
		// حتى المستخدم يعرف إنه لازم يضيفه يدوي — مو «تعذر الموافقة»
		// وهو موافق عليه فعلاً.
		if item != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		WriteError(w, http.StatusBadRequest, "تعذر الموافقة")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *ProductRequestHandler) Reject(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Reject(r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الرفض")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
