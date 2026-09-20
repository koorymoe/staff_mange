package handler

import (
	"net/http"
	"strconv"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// AchievementHandler الإنجازات — تقرير يومي حر من أي موظف بأي دور.
//
// ⚠️ `Create` بلا حراسة دور (أي موظف مسجّل دخول)؛ `List`/`Review`
// محميين بـ`requireAdmin` بمسارات `main.go` — الوجهة حصراً مدير النظام
// والمالك.
type AchievementHandler struct {
	service *service.AchievementService
}

func NewAchievementHandler(s *service.AchievementService) *AchievementHandler {
	return &AchievementHandler{service: s}
}

func (h *AchievementHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAchievementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	a, err := h.service.Create(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, a)
}

// GET /api/achievements?employeeId=&day=&limit=
func (h *AchievementHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	rows, err := h.service.List(q.Get("employeeId"), q.Get("day"), limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإنجازات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/achievements/{id}/review
func (h *AchievementHandler) Review(w http.ResponseWriter, r *http.Request) {
	var req model.ReviewAchievementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	a, err := h.service.Review(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, a)
}
