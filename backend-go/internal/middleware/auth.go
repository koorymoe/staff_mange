package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

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
	ContextRealm      contextKey = "realm"
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

			// إبطال الجلسات: أي توكن صدر قبل آخر إبطال (تغيير كلمة سر، حظر،
			// إنهاء جلسات) يُرفض فوراً — بدونه التوكن المسروق يضل شغّال 12 ساعة.
			if !auth.SessionValid(claims.EmployeeID, claims.IssuedAt) {
				writeError(w, http.StatusUnauthorized, "انتهت صلاحية الجلسة — سجّل دخول مجدداً")
				return
			}

			// الطبقة: التوكنات القديمة بلا realm تنعتبر staff.
			realm := claims.Realm
			if realm == "" {
				realm = service.RealmStaff
			}
			// ⚠️ توكن مركز القيادة **ما يشتغل** على مسارات الموظفين.
			// بدون هذا الحاجز، الفصل بين الطبقتين يصير بالاسم بس: يكفي
			// تبدّل المسار بالمتصفح وتشتغل بكل النظام بتوكن القيادة.
			// مسارات القيادة تسمح لنفسها بحارس مستقل (RequireCommandRealm).
			if realm != service.RealmStaff && !strings.HasPrefix(r.URL.Path, "/api/command/") {
				writeError(w, http.StatusForbidden, "توكن مركز القيادة ما يشتغل على هذا المسار")
				return
			}

			ctx := context.WithValue(r.Context(), ContextEmployeeID, claims.EmployeeID)
			ctx = context.WithValue(ctx, ContextRole, role)
			ctx = context.WithValue(ctx, ContextRealm, realm)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// lockoutRepo يُحقن مرة وحدة عند الإقلاع — الميدل وير ما ياخذه بكل نداء حتى
// ما نغيّر توقيع كل الدوال الموجودة.
var lockoutRepo *repository.SecurityLockoutRepository

// SetLockoutRepository يربط مستودع الحظر بالميدل وير (يُنادى من main).
func SetLockoutRepository(r *repository.SecurityLockoutRepository) { lockoutRepo = r }

// عتبة الحظر التلقائي على محاولات الوصول غير المخوّلة.
//
// ليش عتبة مو حظر من أول محاولة؟ الحظر الفوري كان يوكع بيه الموظف العادي:
// القائمة الجانبية تعرض رابط، الصفحة تنادي خادم ما يخصه، يجي ٤٠٣ واحد،
// وينحظر الحساب حظر دائم ما ينفك إلا بيد المالك. المهاجم الحقيقي يجرّب
// عشرات المسارات، فالعتبة تمسكه، والموظف الي دس دوسة غلط ما تأذيه.
const (
	authzLockThreshold = 5
	authzLockWindow    = 10 * time.Minute
)

// authzViolationLog يخزّن أوقات المحاولات المرفوضة الأخيرة لكل موظف حتى
// نحسب "٥ محاولات خلال ١٠ دقائق" بدل العدّاد التراكمي مدى الحياة.
var authzViolationLog = struct {
	sync.Mutex
	hits map[string][]time.Time
}{hits: make(map[string][]time.Time)}

// registerAuthzViolation يسجّل محاولة جديدة ويرجّع عددها داخل النافذة.
func registerAuthzViolation(employeeID string, now time.Time) int {
	authzViolationLog.Lock()
	defer authzViolationLog.Unlock()

	cutoff := now.Add(-authzLockWindow)
	kept := authzViolationLog.hits[employeeID][:0]
	for _, t := range authzViolationLog.hits[employeeID] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	kept = append(kept, now)
	authzViolationLog.hits[employeeID] = kept

	// تنظيف دوري: نشيل الموظفين الي ما عدهم محاولات فعّالة حتى الخريطة ما
	// تكبر بلا حدود على سيرفر شغّال أشهر.
	if len(authzViolationLog.hits) > 1000 {
		for id, times := range authzViolationLog.hits {
			if len(times) == 0 || !times[len(times)-1].After(cutoff) {
				delete(authzViolationLog.hits, id)
			}
		}
	}
	return len(kept)
}

// clearAuthzViolations يصفّي سجل الموظف (يُستعمل بالاختبارات وبعد فك الحظر).
func clearAuthzViolations(employeeID string) {
	authzViolationLog.Lock()
	defer authzViolationLog.Unlock()
	delete(authzViolationLog.hits, employeeID)
}

// recordViolationAndBlock تسجل محاولة وصول مرفوضة وترد "ممنوع".
//
// الحظر التلقائي ما ينحرك إلا بشرطين مع بعض:
//  1. الطلب مو مجرد قراءة (GET/HEAD/OPTIONS) — فتح صفحة غلط مو اعتداء.
//  2. الموظف تجاوز authzLockThreshold محاولة خلال authzLockWindow.
//
// وبكل الحالات ننبّه الإدارة حتى القرار البشري يضل موجود.
func recordViolationAndBlock(w http.ResponseWriter, r *http.Request, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, employeeID string) {
	if employeeID != "" && employees != nil {
		if violations, err := employees.RecordAuthzViolation(employeeID); err == nil {
			name, nameErr := employees.NameByID(employeeID)
			if nameErr != nil || name == "" {
				name = employeeID
			}
			readOnly := r != nil && (r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions)
			// القراءات ما تُحسب بالنافذة أصلاً — لا تحظر ولا تقرّب الموظف من
			// الحظر. لو حسبناها، موظف جمّع ٤٠٣ من روابط قائمة مكسورة ينحظر
			// بأول عملية كتابة عادية.
			recent := 0
			if !readOnly {
				recent = registerAuthzViolation(employeeID, time.Now())
			}

			if lockoutRepo != nil {
				_ = lockoutRepo.LogEvent(&employeeID, name, "AUTHZ_VIOLATION",
					fmt.Sprintf("محاولة وصول غير مخوّلة رقم %d (%d كتابة خلال %s)", violations, recent, authzLockWindow), "", "")

				if recent >= authzLockThreshold {
					locked, _ := lockoutRepo.Lock(employeeID, repository.LockReasonAuthzAbuse,
						fmt.Sprintf("حاول %d مرات يوصل لعمليات مو مخوّل لها خلال %s", recent, authzLockWindow))
					if locked {
						_ = lockoutRepo.LogEvent(&employeeID, name, "ACCOUNT_LOCKED",
							"انحظر تلقائياً بعد تكرار محاولات وصول غير مخوّلة", "", "")
						if notifications != nil {
							_ = notifications.CreateForRole("ADMIN", "authz_violation",
								fmt.Sprintf("🔒 انحظر حساب %s تلقائياً — كرّر محاولات وصول غير مخوّلة", name))
						}
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
				recordViolationAndBlock(w, r, employees, notifications, EmployeeIDFromContext(r))
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
				recordViolationAndBlock(w, r, employees, notifications, employeeID)
				return
			}
			for _, p := range perms {
				if p.Name == permissionName {
					next.ServeHTTP(w, r)
					return
				}
			}
			recordViolationAndBlock(w, r, employees, notifications, employeeID)
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
				recordViolationAndBlock(w, r, employees, notifications, employeeID)
				return
			}
			for _, p := range perms {
				if allowed[p.Name] {
					next.ServeHTTP(w, r)
					return
				}
			}
			recordViolationAndBlock(w, r, employees, notifications, employeeID)
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
				recordViolationAndBlock(w, r, employees, notifications, employeeID)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireLeaderOrBookingServiceManager نفس RequireLeader بالضبط، بس
// يسمح زيادةً لـ**مسؤول خدمة الحجز** الي بالطلب.
//
// ليش موجود: بخدمات مثل الجي بي اس والداش كام، الفاتورة على مسؤول
// الخدمة مو على الفني — ومسؤول الخدمة مو بالضرورة ليدر.
//
// ⚠️⚠️ **الحارس يتوسّع ما ينفكّ.** الي مو ليدر ولا مسؤول الخدمة يبقى
// يمر بنفس `recordViolationAndBlock` — يعني تنسجّل مخالفته ويتقرّب من
// الحظر التلقائي، بالضبط مثل اليوم. لو نقلنا الفحص جوّا الهاندلر
// (أسهل بكثير) چان ضاع هالجزء بصمت: الرفض يصير رسالة خطأ عادية،
// والي يجرّب يفوّت الحارس مية مرة ما ينحظر ولا أحد ينتبه.
//
// ⚠️ ويقرا الجسم ويرجّعه: `bookingId` ما يجي بالمسار، يجي بجسم الطلب.
// بلا الإرجاع، الهاندلر يلگه جسماً فاضياً — وتصير الفاتورة تفشل
// بـ«بيانات الطلب غير صحيحة» بلا أي علاقة بالسبب الحقيقي.
func RequireLeaderOrBookingServiceManager(
	employees *repository.EmployeeRepository,
	notifications *repository.NotificationRepository,
	isManagerOfBooking func(employeeID, bookingID string) (bool, error),
) func(http.Handler) http.Handler {
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

			// مو ليدر — نشوف إذا مسؤول خدمة هالحجز
			bookingID := ""
			if r.Body != nil {
				raw, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
				_ = r.Body.Close()
				if err == nil {
					r.Body = io.NopCloser(bytes.NewReader(raw))
					var probe struct {
						BookingID string `json:"bookingId"`
					}
					_ = json.Unmarshal(raw, &probe)
					bookingID = probe.BookingID
				}
			}
			if bookingID != "" && isManagerOfBooking != nil {
				if ok, err := isManagerOfBooking(employeeID, bookingID); err == nil && ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			recordViolationAndBlock(w, r, employees, notifications, employeeID)
		})
	}
}

// RequireOwner يقصر المسار على حساب المالك وحده — ولا حتى ADMIN.
//
// ثلاث فروقات مقصودة عن RequireRole("OWNER"):
//
//  1. يرجّع 404 مو 403. الـ403 يعترف إن المسار موجود ويكشف وجود الميزة
//     لأي واحد يجرّب. الـ404 ما يفرّق عن أي رابط مو موجود أصلاً.
//  2. ما ينبّه الإدارة ولا يسجّل مخالفة. تنبيه «فلان حاول يوصل لعملية
//     مو مخوّل لها» يوصل لـADMIN — يعني نفس الشي الي نخبّيه ينكشف
//     بالإشعار.
//  3. ما يقرّب أحد من الحظر التلقائي.
//
// تُستخدم لمراقبة النسخ الاحتياطية: المالك سلّم إدارة النظام لكن
// الإشراف على النسخ يبقى عنده هو بس.
func RequireOwner() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if role, _ := r.Context().Value(ContextRole).(string); role != "OWNER" {
				writeError(w, http.StatusNotFound, "المسار غير موجود")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwnerOnly يقصر المسار على المالك — ولا حتى ADMIN — بس **بلا
// عقوبة**: يرجّع 403 برسالة واضحة، ما يسجّل مخالفة ولا يقرّب أحد من
// الحظر التلقائي.
//
// ⚠️ ليش مو RequireRole("OWNER")؟
// لأنها تنادي recordViolationAndBlock، ويعني مدير النظام الي يضغط زر
// باقي بواجهة قديمة بالكاش **ينحظر حسابه بعد ٣ ضغطات**. الميزة هنا
// محصورة مو مخفية، والمنع المشروع مو محاولة اختراق.
//
// ⚠️ وليش مو RequireOwner (الـ404)؟
// الـ404 للأشياء الي وجودها نفسه سر (النسخ الاحتياطية). فتح الحسابات
// مو سر — مدير النظام يعرف إن الحسابات موجودة، بس ما إله يفتحها.
// رسالة واضحة أحسن من إخفاء يخلّيه يظن النظام مكسور.
//
// تُستخدم لفتح الحسابات: المالك سلّم إدارة النظام، بس **منو يدخل
// النظام** يبقى قراره هو.
func RequireOwnerOnly(message string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if role, _ := r.Context().Value(ContextRole).(string); role != "OWNER" {
				writeError(w, http.StatusForbidden, message)
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
			recordViolationAndBlock(w, r, employees, notifications, employeeID)
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
			recordViolationAndBlock(w, r, employees, notifications, employeeID)
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
			recordViolationAndBlock(w, r, employees, notifications, employeeID)
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
			recordViolationAndBlock(w, r, employees, notifications, employeeID)
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


// RequireCommandRealm يقبل توكنات مركز القيادة بس.
//
// العكس مقصود: توكن الموظف العادي ما يفتح مركز القيادة حتى لو صاحبه
// مالك — الطبقة العليا تتطلب الباسورد الثاني، وهذا كل معناها.
func RequireCommandRealm() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if realm, _ := r.Context().Value(ContextRealm).(string); realm != service.RealmCommand {
				writeError(w, http.StatusNotFound, "المسار غير موجود")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
