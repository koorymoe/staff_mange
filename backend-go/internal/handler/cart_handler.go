package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type CartHandler struct {
	service *service.CartService
}

func NewCartHandler(s *service.CartService) *CartHandler {
	return &CartHandler{service: s}
}

// GET /api/v1/cart/booking/{bookingId}
func (h *CartHandler) ListForBooking(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListForBooking(r.PathValue("bookingId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مواد السلة")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// POST /api/v1/cart/booking/{bookingId}
func (h *CartHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCartItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.Create(r.PathValue("bookingId"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

// PUT /api/v1/cart/{id}
func (h *CartHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateCartItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

// DELETE /api/v1/cart/{id}
func (h *CartHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}
