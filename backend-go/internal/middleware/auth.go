package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type contextKey string

const (
	ContextEmployeeID contextKey = "employeeId"
	ContextRole       contextKey = "role"
)

// writeError يكتب استجابة خطأ بنفس شكل handler.WriteError دون استيراد حزمة
// handler هنا (تفادياً لحلقة استيراد بما إن handler يستورد middleware أيضاً).
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// RequireAuth يتحقق من وجود JWT صالح بالطلب ويرفض أي طلب بدونه
func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "يجب تسجيل الدخول")
				return
			}

			tokenString := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.ParseToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "جلسة الدخول منتهية، الرجاء تسجيل الدخول مجدداً")
				return
			}

			ctx := context.WithValue(r.Context(), ContextEmployeeID, claims.EmployeeID)
			ctx = context.WithValue(ctx, ContextRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole يمنع الوصول إلا لأصحاب الأدوار المذكورة (يُستخدم بعد RequireAuth)
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextRole).(string)
			// OWNER يتخطى أي قيد أدوار — حساب المالك الأساسي، أقوى من أي دور ثاني بما فيه ADMIN
			if role != "OWNER" && !allowed[role] {
				writeError(w, http.StatusForbidden, "لا تملك صلاحية الوصول لهذه العملية")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission يسمح بالوصول لـ ADMIN دائماً، أو لأي موظف عنده الصلاحية المذكورة
// من جدول الصلاحيات المخصصة (يُستخدم بعد RequireAuth)
func RequirePermission(permissions *repository.PermissionRepository, permissionName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextRole).(string)
			if role == "ADMIN" || role == "OWNER" {
				next.ServeHTTP(w, r)
				return
			}
			employeeID, _ := r.Context().Value(ContextEmployeeID).(string)
			perms, err := permissions.ListForEmployee(employeeID)
			if err != nil {
				writeError(w, http.StatusForbidden, "لا تملك صلاحية الوصول لهذه العملية")
				return
			}
			for _, p := range perms {
				if p.Name == permissionName {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeError(w, http.StatusForbidden, "لا تملك صلاحية الوصول لهذه العملية")
		})
	}
}

func EmployeeIDFromContext(r *http.Request) string {
	id, _ := r.Context().Value(ContextEmployeeID).(string)
	return id
}

func RoleFromContext(r *http.Request) string {
	role, _ := r.Context().Value(ContextRole).(string)
	return role
}
