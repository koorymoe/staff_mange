package middleware

import (
	"context"
	"net/http"
	"strings"

	"staffmange-api/internal/handler"
	"staffmange-api/internal/service"
)

type contextKey string

const (
	ContextEmployeeID contextKey = "employeeId"
	ContextRole       contextKey = "role"
)

// RequireAuth يتحقق من وجود JWT صالح بالطلب ويرفض أي طلب بدونه
func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				handler.WriteError(w, http.StatusUnauthorized, "يجب تسجيل الدخول")
				return
			}

			tokenString := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.ParseToken(tokenString)
			if err != nil {
				handler.WriteError(w, http.StatusUnauthorized, "جلسة الدخول منتهية، الرجاء تسجيل الدخول مجدداً")
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
			if !allowed[role] {
				handler.WriteError(w, http.StatusForbidden, "لا تملك صلاحية الوصول لهذه العملية")
				return
			}
			next.ServeHTTP(w, r)
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
