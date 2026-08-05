package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

var ErrInvalidCredentials = errors.New("اسم المستخدم أو كلمة المرور غير صحيحة")
var ErrAccountSuspended = errors.New("تم إيقاف هذا الحساب — راجع إدارة النظام")

type AuthService struct {
	employees  *repository.EmployeeRepository
	loginAudit *repository.LoginAuditRepository
	lockout    *repository.SecurityLockoutRepository
	jwtSecret  []byte
}

func NewAuthService(employees *repository.EmployeeRepository, loginAudit *repository.LoginAuditRepository, lockout *repository.SecurityLockoutRepository, jwtSecret string) *AuthService {
	return &AuthService{employees: employees, loginAudit: loginAudit, lockout: lockout, jwtSecret: []byte(jwtSecret)}
}

// ErrAccountLocked الحساب انحظر تلقائياً — ما ينفتح إلا بيد المالك.
var ErrAccountLocked = errors.New("تم حظر هذا الحساب لأسباب أمنية — راجع مالك النظام لإعادة تفعيله")

// ErrAccountTemporarilyLocked تعطيل مؤقت بسبب محاولات كلمة مرور خاطئة —
// ينتهي لحاله، عمداً: الحظر الدائم هنا يخلي المهاجم يقفل حسابات الموظفين
// بدون ما يعرف كلمات سرهم.
var ErrAccountTemporarilyLocked = errors.New("تم تعطيل الدخول مؤقتاً بسبب محاولات خاطئة متكررة — جرّب بعد 15 دقيقة")

type Claims struct {
	EmployeeID string `json:"employeeId"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(username, password, ip, userAgent string) (*model.Employee, string, error) {
	employee, err := s.employees.FindByUsername(username)
	if err != nil || employee == nil || employee.Password == nil {
		_ = s.loginAudit.Record(username, nil, false, ip, userAgent)
		return nil, "", ErrInvalidCredentials
	}

	// الحساب المحظور ما يدخل حتى لو كلمة السر صحيحة.
	//
	// نميّز بين نوعين:
	//  - تعطيل مؤقت (كلمة سر خاطئة متكررة): ينتهي لحاله بعد المدة، وبعدها
	//    نصفّره ونكمل عادي. هذا يمنع التخمين بدون ما يخلي مهاجم يقفل
	//    حسابات الموظفين نهائياً وهو ما يعرف كلمات سرهم.
	//  - حظر دائم (محاولة وصول غير مخوّلة): ما ينفك إلا بيد المالك.
	if employee.LockedAt != nil {
		if repository.IsTemporarilyLocked(employee.LockedAt, employee.LockedReason) {
			_ = s.loginAudit.Record(username, &employee.ID, false, ip, userAgent)
			if s.lockout != nil {
				_ = s.lockout.LogEvent(&employee.ID, employee.Name, "LOGIN_WHILE_LOCKED",
					"محاولة دخول أثناء التعطيل المؤقت", ip, userAgent)
			}
			return nil, "", ErrAccountTemporarilyLocked
		}
		if employee.LockedReason != nil && *employee.LockedReason == repository.LockReasonFailedLogins {
			// انتهت مدة التعطيل المؤقت — نصفّره ونكمل
			if s.lockout != nil {
				_ = s.lockout.ClearTemporaryLock(employee.ID)
			}
		} else {
			_ = s.loginAudit.Record(username, &employee.ID, false, ip, userAgent)
			if s.lockout != nil {
				_ = s.lockout.LogEvent(&employee.ID, employee.Name, "LOGIN_WHILE_LOCKED",
					"محاولة دخول لحساب محظور", ip, userAgent)
			}
			return nil, "", ErrAccountLocked
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*employee.Password), []byte(password)); err != nil {
		_ = s.loginAudit.Record(username, &employee.ID, false, ip, userAgent)
		if s.lockout != nil {
			streak, locked, _ := s.lockout.RegisterFailedLogin(employee.ID)
			detail := fmt.Sprintf("كلمة مرور خاطئة (المحاولة %d من %d)", streak, repository.FailedLoginThreshold)
			_ = s.lockout.LogEvent(&employee.ID, employee.Name, "LOGIN_FAILED", detail, ip, userAgent)
			if locked {
				_ = s.lockout.LogEvent(&employee.ID, employee.Name, "ACCOUNT_TEMP_LOCKED",
					fmt.Sprintf("تعطيل مؤقت (15 دقيقة) بعد %d محاولات كلمة مرور خاطئة", streak), ip, userAgent)
				return nil, "", ErrAccountTemporarilyLocked
			}
			// حساب مستثنى من الحظر (مالك أو مدير نظام) وتجاوز الحد: ما
			// نكذب عليه برسالة «معطّل» — هو مو معطّل فعلاً — بس نرفع
			// حدث مميز حتى المالك يشوف إنه اكو أحد يخمّن على أهم حساب.
			if streak >= repository.FailedLoginThreshold {
				_ = s.lockout.LogEvent(&employee.ID, employee.Name, "PRIVILEGED_LOGIN_ATTACK",
					fmt.Sprintf("%d محاولة خاطئة على حساب %s — الحساب ما ينحظر تلقائياً", streak, employee.Role),
					ip, userAgent)
			}
		}
		return nil, "", ErrInvalidCredentials
	}

	if employee.Status != "ACTIVE" {
		_ = s.loginAudit.Record(username, &employee.ID, false, ip, userAgent)
		return nil, "", ErrAccountSuspended
	}

	// دخول ناجح -> نصفّر عدّاد المحاولات الفاشلة
	if s.lockout != nil {
		_ = s.lockout.ResetFailedLogins(employee.ID)
	}

	token, err := s.GenerateToken(employee)
	if err != nil {
		return nil, "", err
	}

	_ = s.loginAudit.Record(username, &employee.ID, true, ip, userAgent)
	return employee, token, nil
}

// ChangePassword يغيّر كلمة مرور موظف بنفسه — يتحقق من كلمة المرور الحالية
// قبل ما يخليه يحدد وحدة جديدة.
func (s *AuthService) ChangePassword(employeeID, currentPassword, newPassword string) error {
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return err
	}
	employee, err := s.employees.FindByID(employeeID)
	if err != nil || employee == nil || employee.Password == nil {
		return errors.New("تعذر التحقق من الموظف")
	}
	if bcrypt.CompareHashAndPassword([]byte(*employee.Password), []byte(currentPassword)) != nil {
		return errors.New("كلمة المرور الحالية غير صحيحة")
	}
	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.employees.SetPassword(employeeID, hashed); err != nil {
		return err
	}
	// تغيير كلمة السر يبطل كل الجلسات القديمة فوراً — بدونها التوكن
	// المسروق يضل شغّال لين ينتهي (12 ساعة) حتى بعد ما الموظف غيّر سره.
	if s.lockout != nil {
		_ = s.lockout.InvalidateSessions(employeeID)
	}
	return nil
}

func (s *AuthService) Me(employeeID string) (*model.Employee, error) {
	return s.employees.FindByID(employeeID)
}

func (s *AuthService) GenerateToken(employee *model.Employee) (string, error) {
	claims := Claims{
		EmployeeID: employee.ID,
		Role:       employee.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// SessionValid يتأكد إن التوكن صدر *بعد* آخر إبطال جلسات للموظف.
func (s *AuthService) SessionValid(employeeID string, issuedAt *jwt.NumericDate) bool {
	if issuedAt == nil {
		return false
	}
	emp, err := s.employees.FindByID(employeeID)
	if err != nil || emp == nil {
		return false
	}
	if emp.SessionsInvalidatedAt == nil {
		return true
	}
	// هامش ثانية: التوكن يخزن الوقت بدقة الثانية
	return issuedAt.Time.Add(time.Second).After(*emp.SessionsInvalidatedAt)
}

func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	// نثبّت الخوارزمية صراحةً (HS256) — بدونها نظرياً ينفتح باب هجمات
	// الخلط بين خوارزميات التوقيع (alg confusion)
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return nil, errors.New("رمز الدخول غير صالح")
	}
	return claims, nil
}

// ValidatePasswordStrength الحد الأدنى لقوة كلمة المرور.
//
// 6 أحرف بلا أي شرط جان ضعيف جداً: "123456" مقبولة، وتنكسر بثوانٍ لو تسرّبت
// قاعدة البيانات. نطلب 10 أحرف على الأقل مع خليط حروف وأرقام، ونرفض
// الكلمات الشائعة جداً.
func ValidatePasswordStrength(pw string) error {
	if len([]rune(pw)) < 10 {
		return errors.New("كلمة المرور لازم تكون 10 محارف على الأقل")
	}
	var hasLetter, hasDigit bool
	for _, c := range pw {
		switch {
		case c >= '0' && c <= '9':
			hasDigit = true
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c > 127:
			hasLetter = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("كلمة المرور لازم تحتوي حروف وأرقام سوة")
	}
	lower := strings.ToLower(pw)
	for _, bad := range []string{"password", "123456789", "qwerty", "111111", "iraq", "admin"} {
		if strings.Contains(lower, bad) {
			return errors.New("كلمة المرور سهلة التخمين — اختر وحدة أقوى")
		}
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}
