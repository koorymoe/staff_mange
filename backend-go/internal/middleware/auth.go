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
// lockoutRepo يُحقن مرة وحدة عند الإقلاع — الميدل وير ما ياخذه بكل نداء حتى
// ما نغيّر توقيع كل الدوال الموجودة.
var lockoutRepo *repository.SecurityLockoutRepository

// SetLockoutRepository يربط مستودع الحظر بالميدل وير (يُنادى من main).
func SetLockoutRepository(r *repository.SecurityLockoutRepository) { lockoutRepo = r }

func recordViolationAndBlock(w http.ResponseWriter, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, employeeID string) {
	if employeeID != "" && employees != nil {
		if violations, err := employees.RecordAuthzViolation(employeeID); err == nil {
			name, nameErr := employees.NameByID(employeeID)
			if nameErr != nil || name == "" {
				name = employeeID
			}
			// أي محاولة وصول لعملية مو مخوّل لها = حظر فوري للحساب. المالك
			// ومدير النظام مستثنون (Lock نفسها تستثنيهم) حتى ما ننقفل بره
			// النظام. الحساب ما ينفتح إلا بيد المالك.
			if lockoutRepo != nil {
				locked, _ := lockoutRepo.Lock(employeeID, repository.LockReasonAuthzAbuse,
					"حاول يوصل لعملية مو مخوّل لها")
				_ = lockoutRepo.LogEvent(&employeeID, name, "AUTHZ_VIOLATION",
					fmt.Sprintf("محاولة وصول غير مخوّلة رقم %d", violations), "", "")
				if locked {
					_ = lockoutRepo.LogEvent(&employeeID, name, "ACCOUNT_LOCKED",
						"انحظر تلقائياً بعد محاولة وصول غير مخوّلة", "", "")
					if notifications != nil {
						_ = notifications.CreateForRole("ADMIN", "authz_violation",
							fmt.Sprintf("🔒 انحظر حساب %s تلقائياً — حاول يوصل لعملية مو مخوّل لها", name))
					}
				}
			}
			if violations > 0 && violations%authzViolationNotifyThreshold == 0 && notifications != nil {
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

// RequireAnyPermission نفس RequirePermission لكن يسمح لو الموظف عنده أي وحدة
// من عدة صلاحيات معطاة (OR مو AND) — يُستخدم لما ميزة وحدة يقدر يوصلها أكثر
// من مستوى صلاحية (مثلاً عروض الأسعار: إضافة فقط / إضافة وتعديل / الكل).
func RequireAnyPermission(permissions *repository.PermissionRepository, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, permissionNames ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(permissionNames))
	for _, name := range permissionNames {
		allowed[name] = true
	}
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
				if allowed[p.Name] {
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
// RequireRoleOrPermission يسمح بالوصول لأي موظف دوره ضمن roles المذكورة (زي
// RequireRole) أو لأي موظف عنده الصلاحية المخصصة permissionName (زي RequirePermission)
// أو ADMIN/OWNER دايماً. يُستخدم لتوسيع وصول مبني على الدور (مثلاً requireHR)
// ليشمل أيضاً أي موظف مُنح صلاحية مخصصة مكافئة من صفحة الصلاحيات، بدون ما نلغي
// وصول الأدوار الأصلية.
func RequireRoleOrPermission(permissions *repository.PermissionRepository, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, roles []string, permissionName string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextRole).(string)
			if role == "ADMIN" || role == "OWNER" || allowed[role] {
				next.ServeHTTP(w, r)
				return
			}
			employeeID := EmployeeIDFromContext(r)
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

// RequireRoleOrAnyPermission نفس RequireRoleOrPermission لكن يقبل أكثر من صلاحية
// (OR) — يُستخدم لما ميزة وحدة يقدر يوصلها دور معيّن أو أي وحدة من عدة صلاحيات
// (مثلاً إنشاء مشروع: مدير المشاريع، أو صلاحية إدارة المشاريع الكاملة، أو صلاحية
// "إضافة مشروع فقط" المبسّطة).
func RequireRoleOrAnyPermission(permissions *repository.PermissionRepository, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, roles []string, permissionNames ...string) func(http.Handler) http.Handler {
	allowedRoles := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowedRoles[role] = true
	}
	allowedPerms := make(map[string]bool, len(permissionNames))
	for _, name := range permissionNames {
		allowedPerms[name] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextRole).(string)
			if role == "ADMIN" || role == "OWNER" || allowedRoles[role] {
				next.ServeHTTP(w, r)
				return
			}
			employeeID := EmployeeIDFromContext(r)
			if perms, err := permissions.ListForEmployee(employeeID); err == nil {
				for _, p := range perms {
					if allowedPerms[p.Name] {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			recordViolationAndBlock(w, employees, notifications, employeeID)
		})
	}
}

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

// RequireLeaderOrAnyPermission نفس السابقة بس تقبل أكثر من صلاحية — تُستخدم
// لفواتير الليدر: يشوفها الليدر، أو صاحب سلة الليدر، أو المحاسب.
func RequireLeaderOrAnyPermission(permissions *repository.PermissionRepository, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, permissionNames ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(permissionNames))
	for _, n := range permissionNames {
		allowed[n] = true
	}
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
					if allowed[p.Name] {
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
