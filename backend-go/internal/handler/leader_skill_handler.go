package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type LeaderSkillHandler struct {
	repo *repository.LeaderSkillRepository
}

func NewLeaderSkillHandler(r *repository.LeaderSkillRepository) *LeaderSkillHandler {
	return &LeaderSkillHandler{repo: r}
}

// GET /api/employees/{id}/leader-skills
//
// القراءة لأي موظف مسجَّل: الموظف نفسه يشوف تقييمه، وتقييم يشوفه
// صاحبه أنفع من تقييم مخبّى عنه.
func (h *LeaderSkillHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListForEmployee(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تقييم القيادة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/employees/{id}/leader-skills
//
// ⚠️ الحارس بالمسار يحصرها بالمالك والمدير وإداري الكوادر.
func (h *LeaderSkillHandler) Set(w http.ResponseWriter, r *http.Request) {
	var req model.SetLeaderSkillsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	target := r.PathValue("id")
	// ⚠️ ما يقيّم نفسه: التقييم الذاتي يفرّغ التقييم من معناه.
	if target == middleware.EmployeeIDFromContext(r) {
		WriteError(w, http.StatusForbidden, "ما تكدر تقيّم نفسك")
		return
	}
	if err := h.repo.SetRatings(target, req.Scores, middleware.EmployeeIDFromContext(r)); err != nil {
		// نمرّر النص: «الدرجة لازم تكون من ١ إلى ١٠» تفيد المستخدم.
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.repo.ListForEmployee(target)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "انحفظ بس تعذر جلب النتيجة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
