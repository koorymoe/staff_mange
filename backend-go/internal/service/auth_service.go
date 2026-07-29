package service

import (
	"errors"
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
	jwtSecret  []byte
}

func NewAuthService(employees *repository.EmployeeRepository, loginAudit *repository.LoginAuditRepository, jwtSecret string) *AuthService {
	return &AuthService{employees: employees, loginAudit: loginAudit, jwtSecret: []byte(jwtSecret)}
}

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

	if err := bcrypt.CompareHashAndPassword([]byte(*employee.Password), []byte(password)); err != nil {
		_ = s.loginAudit.Record(username, &employee.ID, false, ip, userAgent)
		return nil, "", ErrInvalidCredentials
	}

	if employee.Status != "ACTIVE" {
		_ = s.loginAudit.Record(username, &employee.ID, false, ip, userAgent)
		return nil, "", ErrAccountSuspended
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
	if len(newPassword) < 6 {
		return errors.New("كلمة المرور الجديدة لازم تكون 6 أحرف على الأقل")
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
	return s.employees.SetPassword(employeeID, hashed)
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

func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("رمز الدخول غير صالح")
	}
	return claims, nil
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}
