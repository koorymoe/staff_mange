package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

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

	employee, token, err := h.auth.Login(req.Username, req.Password)
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
