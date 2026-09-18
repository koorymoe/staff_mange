package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// SystemSwitchHandler مفاتيح إطفاء الميزات — القراءة للكل والكتابة
// للمالك حصراً.
type SystemSwitchHandler struct {
	repo *repository.SystemSwitchRepository
}

func NewSystemSwitchHandler(r *repository.SystemSwitchRepository) *SystemSwitchHandler {
	return &SystemSwitchHandler{repo: r}
}

// GET /api/system/switches — حالة كل المفاتيح.
//
// ⚠️ **القراءة لكل موظف مو للمالك**: واجهة **كل** موظف تحتاج تعرف
// هل الميزة مطفية، وإلا تبقى تعرضها. وماكو شي سرّي بالجواب — هي
// حالة ميزة يشوفها بعينه على شاشته.
func (h *SystemSwitchHandler) List(w http.ResponseWriter, _ *http.Request) {
	switches, err := h.repo.All()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مفاتيح النظام")
		return
	}
	WriteJSON(w, http.StatusOK, switches)
}

// PUT /api/system/switches/{key} — المالك دايماً، ومدير النظام
// (ADMIN) بس بالمفاتيح المسموحة له (model.SystemSwitchAdminAllowed).
func (h *SystemSwitchHandler) Set(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/system/switches/")
	if !model.KnownSystemSwitch(key) {
		// ⚠️ اسم غلط يرجع خطأ صريح مو نجاحاً صامتاً: النجاح الصامت
		// يخلي (ع) يحسب إنه أطفى شي وهو ما انطفى.
		WriteError(w, http.StatusBadRequest, "مفتاح غير معروف")
		return
	}
	role, _ := r.Context().Value(middleware.ContextRole).(string)
	if role != "OWNER" && !(role == "ADMIN" && model.SystemSwitchAdminAllowed(key)) {
		WriteError(w, http.StatusForbidden, "هذا المفتاح للمالك حصراً")
		return
	}
	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Enabled == nil {
		WriteError(w, http.StatusBadRequest, "لازم تحدد enabled")
		return
	}
	actor, _ := r.Context().Value(middleware.ContextEmployeeID).(string)
	if err := h.repo.Set(key, *body.Enabled, actor); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ المفتاح")
		return
	}
	switches, err := h.repo.All()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "انحفظ بس تعذر جلب الحالة")
		return
	}
	WriteJSON(w, http.StatusOK, switches)
}
