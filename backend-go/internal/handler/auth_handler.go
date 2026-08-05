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
	auth *service.AuthService
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

	WriteJSON(w, http.StatusOK, model.LoginResponse{Employee: *employee, Token: token})
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
