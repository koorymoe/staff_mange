package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// TrainingProgramHandler برامج التدريب المنقولة من نظام الطاقة الشمسية.
type TrainingProgramHandler struct {
	repo *repository.TrainingProgramRepository
}

func NewTrainingProgramHandler(repo *repository.TrainingProgramRepository) *TrainingProgramHandler {
	return &TrainingProgramHandler{repo: repo}
}

// GET /api/training-programs?status=
func (h *TrainingProgramHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.List(r.URL.Query().Get("status"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب البرامج التدريبية")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/training-programs
func (h *TrainingProgramHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeProgram(w, r)
	if !ok {
		return
	}
	p, err := h.repo.Save("", req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, p)
}

// PUT /api/training-programs/{id}
func (h *TrainingProgramHandler) Update(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeProgram(w, r)
	if !ok {
		return
	}
	p, err := h.repo.Save(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, p)
}

// PUT /api/training-programs/{id}/complete
//
// إكمال البرنامج يمنح مهاراته لكل المشاركين فعلاً — مو بس يغيّر الحالة.
func (h *TrainingProgramHandler) Complete(w http.ResponseWriter, r *http.Request) {
	p, err := h.repo.Complete(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, p)
}

// DELETE /api/training-programs/{id}
func (h *TrainingProgramHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حذف البرنامج")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func decodeProgram(w http.ResponseWriter, r *http.Request) (model.SaveTrainingProgramRequest, bool) {
	var req model.SaveTrainingProgramRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات البرنامج غير صحيحة")
		return req, false
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "اسم البرنامج مطلوب")
		return req, false
	}
	switch req.Level {
	case "مبتدئ", "متوسط", "متقدم":
	default:
		req.Level = "مبتدئ"
	}
	switch req.Status {
	case "قيد التخطيط", "جاري التنفيذ", "مكتمل":
	default:
		req.Status = "قيد التخطيط"
	}
	if req.DurationDays < 1 {
		req.DurationDays = 1
	}
	if req.PassRate < 0 || req.PassRate > 100 {
		req.PassRate = 80
	}
	if req.Progress < 0 || req.Progress > 100 {
		req.Progress = 0
	}
	return req, true
}
