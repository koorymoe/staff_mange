package service

import (
	"errors"
	"fmt"
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

// RealmStaff نظام إدارة الشركة (الي موجود)، RealmCommand مركز القيادة.
//
// ⚠️ الطبقة تنكتب بالتوكن نفسه مو بالواجهة: بدونها أي واحد يبدّل
// المسار بالمتصفح ويدخل الطبقة الثانية بتوكن الطبقة الأولى — يعني
// الفصل يصير بالاسم بس.
const (
	RealmStaff   = "staff"
	RealmCommand = "command"
)

type Claims struct {
	EmployeeID string `json:"employeeId"`
	Role       string `json:"role"`
	// Realm فارغ = staff (التوكنات القديمة الي صدرت قبل هاي الميزة).
	Realm string `json:"realm,omitempty"`
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

	// ═══ باسورد مركز القيادة ═══
	// نفس اليوزر وباسورد ثاني يفتح الطبقة العليا (فكرة PPSK). ينتفحص
	// **قبل** العادي لأنه لو طابق ما نريد نعدّ محاولة فاشلة على العادي.
	if employee.CommandPassword != nil && *employee.CommandPassword != "" &&
		bcrypt.CompareHashAndPassword([]byte(*employee.CommandPassword), []byte(password)) == nil {
		_ = s.loginAudit.Record(username+" [مركز القيادة]", &employee.ID, true, ip, userAgent)
		if s.lockout != nil {
			_ = s.lockout.LogEvent(&employee.ID, employee.Name, "COMMAND_LOGIN",
				"دخول مركز القيادة", ip, userAgent)
			_ = s.lockout.ResetFailedLogins(employee.ID)
		}
		token, err := s.generateTokenForRealm(employee, RealmCommand)
		return employee, token, err
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
//
// يرجّع توكن جديد: إبطال الجلسات يقتل توكن الموظف الحالي هو بعد، فبدون
// توكن بديل يبقى بالشاشة وكل طلب بعدها يطلع 401 — يعني «التغيير يضل
// معلّق» بنظر المستخدم. الجديد يخلي جهازه شغّال، وباقي الأجهزة تنطرد
// وهذا كل المطلوب أمنياً.
func (s *AuthService) ChangePassword(employeeID, currentPassword, newPassword string) (string, error) {
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return "", err
	}
	employee, err := s.employees.FindByID(employeeID)
	if err != nil || employee == nil || employee.Password == nil {
		return "", errors.New("تعذر التحقق من الموظف")
	}
	if bcrypt.CompareHashAndPassword([]byte(*employee.Password), []byte(currentPassword)) != nil {
		return "", errors.New("كلمة المرور الحالية غير صحيحة")
	}
	hashed, err := HashPassword(newPassword)
	if err != nil {
		return "", err
	}
	if err := s.employees.SetPassword(employeeID, hashed); err != nil {
		return "", err
	}
	// تغيير كلمة السر يبطل كل الجلسات القديمة فوراً — بدونها التوكن
	// المسروق يضل شغّال لين ينتهي (12 ساعة) حتى بعد ما الموظف غيّر سره.
	if s.lockout != nil {
		_ = s.lockout.InvalidateSessions(employeeID)
	}
	// التوكن الجديد ينصدر بعد الإبطال حتى يعدّي فحص SessionValid
	return s.GenerateToken(employee)
}

func (s *AuthService) Me(employeeID string) (*model.Employee, error) {
	return s.employees.FindByID(employeeID)
}

func (s *AuthService) GenerateToken(employee *model.Employee) (string, error) {
	return s.generateTokenForRealm(employee, RealmStaff)
}

func (s *AuthService) generateTokenForRealm(employee *model.Employee, realm string) (string, error) {
	claims := Claims{
		EmployeeID: employee.ID,
		Role:       employee.Role,
		Realm:      realm,
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

// PasswordMinLength الحد الأدنى لطول كلمة المرور.
const PasswordMinLength = 4

// ValidatePasswordStrength الحد الأدنى لكلمة المرور: 4 محارف وبس.
//
// كانت تطلب 10 محارف + خليط حروف وأرقام + ترفض كلمات شائعة. عملياً
// هذا خلّى الموظفين ما يقدرون يغيّرون سرهم أصلاً — والنتيجة إن الكل
// يبقى على كلمة السر الافتراضية، وهذا أسوأ من كلمة سر قصيرة. الحماية
// الحقيقية جاية من تعطيل الدخول بعد 3 محاولات خاطئة ومن تحديد معدّل
// الطلبات، مو من شروط كتابة تعجّز المستخدم.
func ValidatePasswordStrength(pw string) error {
	if len([]rune(pw)) < PasswordMinLength {
		return fmt.Errorf("كلمة المرور لازم تكون %d محارف على الأقل", PasswordMinLength)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}


// SetCommandPassword يحط أو يغيّر باسورد مركز القيادة.
//
// ⚠️ لازم يختلف عن الباسورد العادي — لو تطابقن انلغى الفصل كله:
// الي يعرف باسورد الموظف يفتح الطبقة العليا.
func (s *AuthService) SetCommandPassword(employeeID, newPassword string) error {
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return err
	}
	employee, err := s.employees.FindByID(employeeID)
	if err != nil || employee == nil {
		return errors.New("تعذر التحقق من الموظف")
	}
	if employee.Password != nil &&
		bcrypt.CompareHashAndPassword([]byte(*employee.Password), []byte(newPassword)) == nil {
		return errors.New("باسورد مركز القيادة لازم يكون مختلف عن باسوردك العادي")
	}
	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.employees.SetCommandPassword(employeeID, hashed)
}
