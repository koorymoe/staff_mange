package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// NetworkCostHandler استمارة تكلفة الشبكات + إدارة أسعارها.
type NetworkCostHandler struct {
	prices *repository.NetworkPriceRepository
}

func NewNetworkCostHandler(prices *repository.NetworkPriceRepository) *NetworkCostHandler {
	return &NetworkCostHandler{prices: prices}
}

// GET /api/network-cost/items — الفقرات الفعّالة للاستمارة.
func (h *NetworkCostHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	items, err := h.prices.List(true)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// GET /api/network-cost/prices — الكل (شامل المعطّلة) لشاشة الإعدادات.
func (h *NetworkCostHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.prices.List(false)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// validate يمنع الفقرات الي تطلّع فاتورة بصفر أو برقم ما ينفهم.
func validateNetworkPrice(req *model.SaveNetworkPriceItemRequest) string {
	req.Label = strings.TrimSpace(req.Label)
	req.Unit = strings.TrimSpace(req.Unit)
	if TextLen(req.Label) < 2 {
		return "اسم الفقرة قصير"
	}
	if req.Unit == "" {
		req.Unit = "قطعة"
	}
	switch req.PricingMode {
	case model.NetworkPricingFlat:
		if req.BasePrice <= 0 {
			return "سعر الوحدة لازم يكون أكبر من صفر"
		}
	case model.NetworkPricingTiered:
		if req.BasePrice <= 0 {
			return "المبلغ المقطوع لازم يكون أكبر من صفر"
		}
		if req.IncludedQty <= 0 {
			return "الكمية المشمولة لازم تكون أكبر من صفر"
		}
		if req.ExtraPerUnit < 0 {
			return "سعر الوحدة الزايدة ما يصير بالسالب"
		}
	case model.NetworkPricingBracket:
		if len(req.Brackets) == 0 {
			return "لازم شريحة وحدة على الأقل"
		}
		for _, b := range req.Brackets {
			if b.UnitPrice <= 0 {
				return "سعر كل شريحة لازم يكون أكبر من صفر"
			}
		}
	default:
		return "نمط تسعير غير معروف"
	}
	return ""
}

// POST /api/network-cost/prices — إضافة فقرة (المالك ومدير النظام).
func (h *NetworkCostHandler) CreatePrice(w http.ResponseWriter, r *http.Request) {
	var req model.SaveNetworkPriceItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	if msg := validateNetworkPrice(&req); msg != "" {
		WriteError(w, http.StatusBadRequest, msg)
		return
	}
	it, err := h.prices.Create(req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, it)
}

// PUT /api/network-cost/prices/{id}
func (h *NetworkCostHandler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
	var req model.SaveNetworkPriceItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	if msg := validateNetworkPrice(&req); msg != "" {
		WriteError(w, http.StatusBadRequest, msg)
		return
	}
	it, err := h.prices.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, it)
}

// DELETE /api/network-cost/prices/{id} — تعطيل مو محو.
func (h *NetworkCostHandler) DeactivatePrice(w http.ResponseWriter, r *http.Request) {
	if err := h.prices.Deactivate(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/network-cost/calculate — حساب الاستمارة.
//
// الحساب بالسيرفر مو بالواجهة: نفس الاستمارة تنفتح بأكثر من جهاز،
// ولو الحساب بالجافاسكربت كل نسخة قديمة من الواجهة تطلّع رقم غير.
func (h *NetworkCostHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req model.NetworkCostRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	items, err := h.prices.List(false)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	byID := map[string]model.NetworkPriceItem{}
	for _, it := range items {
		byID[it.ID] = it
	}

	res := model.NetworkCostResponse{Lines: []model.NetworkCostLineResult{}}
	for _, ln := range req.Lines {
		it, ok := byID[ln.ItemID]
		if !ok {
			WriteError(w, http.StatusBadRequest, "فقرة تسعيرة غير موجودة")
			return
		}
		if ln.Quantity < 0 {
			WriteError(w, http.StatusBadRequest, "الكمية ما تصير بالسالب")
			return
		}
		line := it.CalculateLine(ln.Quantity)
		res.Lines = append(res.Lines, line)
		res.Subtotal += line.Total
	}

	res.Discount = req.Discount
	if res.Discount < 0 {
		res.Discount = 0
	}
	// الخصم ما يتعدى المجموع — وإلا تطلع فاتورة بالسالب
	if res.Discount > res.Subtotal {
		res.Discount = res.Subtotal
	}
	res.FinalAmount = res.Subtotal - res.Discount
	WriteJSON(w, http.StatusOK, res)
}
