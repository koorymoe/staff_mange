package handler

import (
	"net/http"
	"strconv"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ مختبر المحاكاة ═══
//
// ⚠️ كل مسارات هذا المعالج مقفولة بـ`middleware.RequireOwner()` بهالمرحلة
// — ترجّع **404** لأي حساب ثاني حتى مدير النظام، بلا ما تسجّل مخالفة.
// السبب: صاحب النظام طلبها صراحةً «ما أريدها تظهر عند مدير النظام إلى أن
// تكتمل». والصلاحيات ما تگدر تسوي هذا: كل وسائط الصلاحيات بالنظام
// تمرّر ADMIN بلا شرط.
//
// ⚠️ و`isOwner` ينمرّر للمستودع حتى يقرر شنو يشوف السائل: المالك يشوف
// المحتوى غير المحقّق (هو الي يراجعه)، وغيره ما يشوف إلا المنشور
// المحقّق. لمن ينفتح المختبر للموظفين بعدين، البوابة تتغيّر بمكان واحد
// وهذا الشرط يبقى شغّالاً لحاله.
type SimHandler struct {
	repo *repository.SimRepository
}

func NewSimHandler(repo *repository.SimRepository) *SimHandler {
	return &SimHandler{repo: repo}
}

// isOwner يقرا الدور من السياق — `RequireAuth` يجيبه طازجاً من قاعدة
// البيانات بكل طلب، مو من التوكن.
func isOwner(r *http.Request) bool {
	role, _ := r.Context().Value(middleware.ContextRole).(string)
	return role == "OWNER"
}

// GET /api/sim/categories
func (h *SimHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListCategories(isOwner(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب فئات المختبر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/sim/categories/{id}/exercises
func (h *SimHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListExercises(r.PathValue("id"), isOwner(r), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التمارين")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/sim/categories/{id}/lessons
func (h *SimHandler) ListLessons(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListLessons(r.PathValue("id"), isOwner(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الدروس")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/sim/exercises/{id}
func (h *SimHandler) GetExercise(w http.ResponseWriter, r *http.Request) {
	ex, err := h.repo.GetExercise(r.PathValue("id"), isOwner(r))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, ex)
}

// POST /api/sim/exercises/{id}/attempts
func (h *SimHandler) StartAttempt(w http.ResponseWriter, r *http.Request) {
	a, err := h.repo.StartAttempt(r.PathValue("id"), middleware.EmployeeIDFromContext(r), isOwner(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, a)
}

// PUT /api/sim/attempts/{id}/progress
func (h *SimHandler) SaveProgress(w http.ResponseWriter, r *http.Request) {
	var req model.SaveAttemptProgressRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	if err := h.repo.SaveProgress(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PUT /api/sim/attempts/{id}/finish
func (h *SimHandler) FinishAttempt(w http.ResponseWriter, r *http.Request) {
	var req model.FinishAttemptRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	a, err := h.repo.FinishAttempt(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, a)
}

// GET /api/sim/attempts/mine
func (h *SimHandler) MyAttempts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := h.repo.MyAttempts(middleware.EmployeeIDFromContext(r), limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المحاولات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
