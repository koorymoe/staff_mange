package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// ═══ تعلّم أسعار الخدمات ═══
//
// المسار الي طلبه (ع): النظام **يقترح**، والمحاسب **يراجع**، والمالك
// **يعتمد**. وكل خطوة حارسها بـmain.go.

type ServicePriceLearningHandler struct {
	repo *repository.ServicePriceLearningRepository
}

func NewServicePriceLearningHandler(r *repository.ServicePriceLearningRepository) *ServicePriceLearningHandler {
	return &ServicePriceLearningHandler{repo: r}
}

// GET /api/service-prices/samples — كم عيّنة لكل خدمة ومعدّلها
func (h *ServicePriceLearningHandler) Samples(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Stats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب عيّنات الأسعار")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/service-prices/suggestions?status=
func (h *ServicePriceLearningHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Suggestions(r.URL.Query().Get("status"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الاقتراحات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/service-prices/suggestions/{id}/review — المحاسب
func (h *ServicePriceLearningHandler) Review(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Note string `json:"note"`
	}
	_ = DecodeJSON(r, &body)
	if err := h.repo.MarkReviewed(r.PathValue("id"), middleware.EmployeeIDFromContext(r), body.Note); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// PUT /api/service-prices/suggestions/{id}/decide — المالك
func (h *ServicePriceLearningHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Approve bool `json:"approve"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if err := h.repo.Decide(r.PathValue("id"), middleware.EmployeeIDFromContext(r), body.Approve); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}
