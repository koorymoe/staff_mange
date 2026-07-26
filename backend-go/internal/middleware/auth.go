package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// authzViolationNotifyThreshold أول عدد محاولات وصول مرفوضة نبعث بعده تنبيه
// للإدارة — وبعدها كل مضاعف له (3، 6, 9...) حتى ننبه لو المحاولات استمرت
// بدون ما نغرق الإدارة برسالة كل محاولة وحدة.
const authzViolationNotifyThreshold = 3

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

// RequireAuth يتحقق من وجود JWT صالح بالطلب ويرفض أي طلب بدونه. كمان يتحقق
// من حالة الحساب والدور الحقيقيين بقاعدة البيانات بكل طلب (مو بس وقت تسجيل
// الدخول) — حتى لو التوكن نفسه لسه صالح تقنياً:
//   - حساب موقوف (SUSPENDED) أو محذوف ما يقدر يستخدم النظام أبداً بعدها.
//   - الدور المعتمد بكل فحص صلاحيات هو دور الموظف الحالي بقاعدة البيانات،
//     مو الدور القديم المخزّن جوا التوكن وقت تسجيل الدخول — قبل هذا التعديل،
//     تنزيل موظف من ADMIN لدور عادي ما كان يبطل صلاحياته العملية إلا بعد
//     انتهاء التوكن (لغاية ١٢ ساعة) أو تسجيل خروج/دخول يدوي.
func RequireAuth(auth *service.AuthService, employees *repository.EmployeeRepository) func(http.Handler) http.Handler {
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

			status, role, err := employees.StatusAndRoleByID(claims.EmployeeID)
			if err != nil || status != "ACTIVE" {
				writeError(w, http.StatusUnauthorized, "تم إيقاف هذا الحساب — راجع إدارة النظام")
				return
			}

			ctx := context.WithValue(r.Context(), ContextEmployeeID, claims.EmployeeID)
			ctx = context.WithValue(ctx, ContextRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// recordViolationAndBlock تسجل محاولة وصول مرفوضة وترد "ممنوع" — بدون أي
// إيقاف تلقائي للحساب (شوف تعليق RecordAuthzViolation). إذا العدّاد وصل
// عتبة التنبيه (وكل مضاعف لها بعدين)، نبعث تنبيه للإدارة (ADMIN) حتى تراجع
// الموظف يدوياً وتقرر هي — قرار بشري، مو إيقاف آلي أعمى.
func recordViolationAndBlock(w http.ResponseWriter, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, employeeID string) {
	if employeeID != "" && employees != nil {
		if violations, err := employees.RecordAuthzViolation(employeeID); err == nil {
			if violations > 0 && violations%authzViolationNotifyThreshold == 0 && notifications != nil {
				name, nameErr := employees.NameByID(employeeID)
				if nameErr != nil || name == "" {
					name = employeeID
				}
				_ = notifications.CreateForRole("ADMIN", "authz_violation",
					fmt.Sprintf("⚠️ الموظف %s حاول %d مرة يوصل لعملية مو مخوّل لها — راجع صلاحياته/دوره", name, violations))
			}
		}
	}
	writeError(w, http.StatusForbidden, "لا تملك صلاحية الوصول لهذه العملية")
}

// RequireRole يمنع الوصول إلا لأصحاب الأدوار المذكورة (يُستخدم بعد RequireAuth).
// أي محاولة وصول مرفوضة تُسجَّل، وتكرارها ينبّه الإدارة (بدون إيقاف تلقائي).
func RequireRole(employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextRole).(string)
			// OWNER يتخطى أي قيد أدوار — حساب المالك الأساسي، أقوى من أي دور ثاني بما فيه ADMIN
			if role != "OWNER" && !allowed[role] {
				recordViolationAndBlock(w, employees, notifications, EmployeeIDFromContext(r))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission يسمح بالوصول لـ ADMIN دائماً، أو لأي موظف عنده الصلاحية المذكورة
// من جدول الصلاحيات المخصصة (يُستخدم بعد RequireAuth). أي محاولة وصول مرفوضة
// تُسجَّل، وتكرارها ينبّه الإدارة (بدون إيقاف تلقائي).
func RequirePermission(permissions *repository.PermissionRepository, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, permissionName string) func(http.Handler) http.Handler {
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
				recordViolationAndBlock(w, employees, notifications, employeeID)
				return
			}
			for _, p := range perms {
				if p.Name == permissionName {
					next.ServeHTTP(w, r)
					return
				}
			}
			recordViolationAndBlock(w, employees, notifications, employeeID)
		})
	}
}

// RequireLeader يسمح بالوصول فقط للموظفين "ليدر" (isLeader=true) — يُقرأ العلم
// طازج من قاعدة البيانات بكل طلب (مو من التوكن) لنفس سبب StatusAndRoleByID:
// تنزيل موظف من ليدر ما لازم يبقى فعّال إلا بعد تحديث قاعدة البيانات مباشرة.
// ADMIN وOWNER يتخطون هذا الشرط دايماً.
func RequireLeader(employees *repository.EmployeeRepository, notifications *repository.NotificationRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextRole).(string)
			if role == "ADMIN" || role == "OWNER" {
				next.ServeHTTP(w, r)
				return
			}
			employeeID := EmployeeIDFromContext(r)
			isLeader, err := employees.IsLeaderFreshByID(employeeID)
			if err != nil || !isLeader {
				recordViolationAndBlock(w, employees, notifications, employeeID)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireLeaderOrPermission يسمح بالوصول لليدر (isLeader فريش من قاعدة البيانات،
// نفس RequireLeader) أو لأي موظف عنده صلاحية مخصصة معينة (نفس RequirePermission)
// أو ADMIN/OWNER. تُستخدم لسلة الليدر (leader_basket): افتراضياً حصراً لليدر،
// لكن الإدارة تقدر تمنحها لموظف MONITOR أيضاً عبر صفحة الصلاحيات بدون ما يصير
// ليدر فعلياً (isLeader=false يبقى كما هو).
func RequireLeaderOrPermission(permissions *repository.PermissionRepository, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, permissionName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextRole).(string)
			if role == "ADMIN" || role == "OWNER" {
				next.ServeHTTP(w, r)
				return
			}
			employeeID := EmployeeIDFromContext(r)
			if isLeader, err := employees.IsLeaderFreshByID(employeeID); err == nil && isLeader {
				next.ServeHTTP(w, r)
				return
			}
			if perms, err := permissions.ListForEmployee(employeeID); err == nil {
				for _, p := range perms {
					if p.Name == permissionName {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			recordViolationAndBlock(w, employees, notifications, employeeID)
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
