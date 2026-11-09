package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// DecorRateHandler جدول أسعار الديكور — (ع) 10-16.
type DecorRateHandler struct {
	repo *repository.DecorRateRepository
}

func NewDecorRateHandler(r *repository.DecorRateRepository) *DecorRateHandler { return &DecorRateHandler{repo: r} }

// GET /api/decor-rates — البنود الفعّالة (?all=1 للمدير يشوف المعطّلة هم).
func (h *DecorRateHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.List(r.URL.Query().Get("all") != "1")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أسعار الديكور")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/decor-rates — إضافة أو تعديل بند (المدير).
func (h *DecorRateHandler) Save(w http.ResponseWriter, r *http.Request) {
	var x model.DecorRate
	if err := DecodeJSON(r, &x); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	x.Name, x.Unit = strings.TrimSpace(x.Name), strings.TrimSpace(x.Unit)
	if x.ServiceID == "" || x.Name == "" || x.Unit == "" || x.Price < 0 {
		WriteError(w, http.StatusBadRequest, "المهنة والبند والوحدة والسعر كلهن مطلوبات")
		return
	}
	if err := h.repo.Save(x); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	h.List(w, r)
}
