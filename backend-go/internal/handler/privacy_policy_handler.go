package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type PrivacyPolicyHandler struct {
	repo *repository.PrivacyPolicyRepository
}

func NewPrivacyPolicyHandler(repo *repository.PrivacyPolicyRepository) *PrivacyPolicyHandler {
	return &PrivacyPolicyHandler{repo: repo}
}

// منو أضاف كل نقطة معلومة إدارية — تظهر للمالك ومدير النظام بس.
func canSeeAuthors(r *http.Request) bool {
	role := middleware.RoleFromContext(r)
	return role == "ADMIN" || role == "OWNER"
}

func stripAuthors(points []model.PrivacyPolicyPoint) []model.PrivacyPolicyPoint {
	for i := range points {
		points[i].CreatedByName = nil
		points[i].CreatedByEmployeeID = nil
	}
	return points
}

func (h *PrivacyPolicyHandler) List(w http.ResponseWriter, r *http.Request) {
	// الإدارة تشوف كل النقاط (حتى المعطّلة)، والموظف العادي الفعّالة بس
	activeOnly := !canSeeAuthors(r) && r.URL.Query().Get("all") != "true"
	points, err := h.repo.ListPoints(activeOnly)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سياسة الخصوصية")
		return
	}
	if !canSeeAuthors(r) {
		points = stripAuthors(points)
	}
	WriteJSON(w, http.StatusOK, points)
}

func (h *PrivacyPolicyHandler) Status(w http.ResponseWriter, r *http.Request) {
	st, err := h.repo.Status(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حالة الموافقة")
		return
	}
	if !canSeeAuthors(r) {
		st.Points = stripAuthors(st.Points)
	}
	WriteJSON(w, http.StatusOK, st)
}

func (h *PrivacyPolicyHandler) Accept(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Accept(middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تسجيل الموافقة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"accepted": true})
}

func (h *PrivacyPolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertPrivacyPolicyPointRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		WriteError(w, http.StatusBadRequest, "نص النقطة ما يصير فاضي")
		return
	}
	order := 0
	if req.Order != nil {
		order = *req.Order
	} else if existing, err := h.repo.ListPoints(false); err == nil {
		order = len(existing) // تنضاف بآخر القائمة افتراضياً
	}
	actor := middleware.EmployeeIDFromContext(r)
	var by *string
	if actor != "" {
		by = &actor
	}
	p, err := h.repo.CreatePoint(req.Content, order, by)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إضافة النقطة")
		return
	}
	WriteJSON(w, http.StatusCreated, p)
}

func (h *PrivacyPolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertPrivacyPolicyPointRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	p, err := h.repo.UpdatePoint(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تعديل النقطة")
		return
	}
	WriteJSON(w, http.StatusOK, p)
}

func (h *PrivacyPolicyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeletePoint(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حذف النقطة")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
