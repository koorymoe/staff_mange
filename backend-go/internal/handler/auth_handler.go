package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// clientIP يفضّل X-Forwarded-For (اللي يحطه nginx وراه) وإلا يرجع لعنوان الاتصال
// المباشر. المتصفح لا يقدر تقنياً يكشف عنوان MAC الفعلي لأي جهاز — قيد أمان
// بكل المتصفحات الحديثة — فـIP + بصمة المتصفح (User-Agent) هي أقصى ما يمكن تتبعه.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}
	return r.RemoteAddr
}

type AuthHandler struct {
	auth       *service.AuthService
	stealthLog func(ownerID, employeeID, method, path string, status int)
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if req.Username == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "اسم المستخدم وكلمة المرور مطلوبان")
		return
	}

	employee, token, err := h.auth.Login(req.Username, req.Password, clientIP(r), r.UserAgent())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// نستخرج الطبقة من التوكن نفسه مو من متغيّر جانبي — مصدر واحد.
	realm := service.RealmStaff
	if claims, err := h.auth.ParseToken(token); err == nil && claims.Realm != "" {
		realm = claims.Realm
	}
	WriteJSON(w, http.StatusOK, model.LoginResponse{Employee: *employee, Token: token, Realm: realm})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	if employeeID == "" {
		WriteError(w, http.StatusUnauthorized, "غير مسجل الدخول")
		return
	}
	employee, err := h.auth.Me(employeeID)
	if err != nil || employee == nil {
		WriteError(w, http.StatusUnauthorized, "الحساب غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, employee)
}

// PUT /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	employeeID := middleware.EmployeeIDFromContext(r)
	token, err := h.auth.ChangePassword(employeeID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// الواجهة لازم تخزن التوكن الجديد — القديم انبطل بنفس اللحظة
	WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

// PUT /api/auth/command-password — يحط أو يغيّر باسورد مركز القيادة.
//
// قرار (ع) 10-09: المالك يختار أي حساب موجود وينطيه الرمز الثاني. المسار
// محروس بـRequireOwner، فغير المالك ما يوصل له أصلاً؛ وبلا employeeId يصير
// لحساب المالك نفسه مثل قبل.
func (h *AuthHandler) SetCommandPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NewPassword string `json:"newPassword"`
		EmployeeID  string `json:"employeeId"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	employeeID, _ := r.Context().Value(middleware.ContextEmployeeID).(string)
	if req.EmployeeID != "" {
		employeeID = req.EmployeeID
	}
	if err := h.auth.SetCommandPassword(employeeID, req.NewPassword); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/auth/command-accounts — الحسابات الي تفتح مركز القيادة (للمالك)
func (h *AuthHandler) CommandAccounts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.auth.CommandAccounts()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// DELETE /api/auth/command-password/{employeeId} — يسكّر مركز القيادة عن حساب
func (h *AuthHandler) ClearCommandPassword(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.ClearCommandPassword(r.PathValue("employeeId")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الشيل")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/auth/stealth — 🕶️ المالك يدخل بحساب موظف (ع 10-23). الحارس
// RequireOwner (الدور من القاعدة)، فجلسة متخفية ما تكدر تفتح ثانية.
func (h *AuthHandler) Stealth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EmployeeID string `json:"employeeId"`
	}
	if err := DecodeJSON(r, &req); err != nil || req.EmployeeID == "" {
		WriteError(w, http.StatusBadRequest, "اختار الموظف")
		return
	}
	owner := middleware.EmployeeIDFromContext(r)
	employee, token, err := h.auth.StealthToken(owner, req.EmployeeID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.stealthLog != nil {
		h.stealthLog(owner, employee.ID, "STEALTH", "بداية دخول متخفي", http.StatusOK)
	}
	WriteJSON(w, http.StatusOK, model.LoginResponse{Employee: *employee, Token: token, Realm: service.RealmStaff})
}

// SetStealthLog يربط سجل الدخول المتخفي.
func (h *AuthHandler) SetStealthLog(f func(ownerID, employeeID, method, path string, status int)) {
	h.stealthLog = f
}
