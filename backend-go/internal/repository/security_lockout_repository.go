package repository

import (
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// SecurityLockoutRepository يدير الحظر التلقائي وسجل الأحداث الأمنية.
//
// القاعدة: أي حساب ينحظر ما يفتح إلا بيد المالك — لا بمرور الوقت ولا بإعادة
// المحاولة. هذا مقصود: نريد المالك يشوف شنو صار قبل ما يرجّع الحساب.
type SecurityLockoutRepository struct {
	db *sqlx.DB
}

func NewSecurityLockoutRepository(db *sqlx.DB) *SecurityLockoutRepository {
	return &SecurityLockoutRepository{db: db}
}

// أسباب الحظر
const (
	LockReasonFailedLogins = "FAILED_LOGINS"
	LockReasonAuthzAbuse   = "AUTHZ_ABUSE"
	LockReasonManual       = "MANUAL"
)

// FailedLoginThreshold عدد محاولات كلمة السر الغلط المتتالية قبل التعطيل.
const FailedLoginThreshold = 3

// FailedLoginCooldown مدة التعطيل المؤقت بعد تجاوز الحد.
//
// ليش مؤقت مو دائم؟ الحظر الدائم عند كلمة سر غلط يعطي المهاجم سلاح: يجرّب
// أسماء المستخدمين ويقفل حسابات الموظفين واحد واحد *بدون ما يعرف ولا كلمة
// سر*. التعطيل المؤقت يوقف التخمين (اللي هدفه أصلاً محاولات كثيرة بسرعة)
// بدون ما يخلي المهاجم يشلّ الشركة. الحظر *الدائم* يبقى بس لمحاولة الوصول
// غير المخوّلة — وهذي تحتاج حساب مسجّل دخول أصلاً، فما تنستغل من بره.
const FailedLoginCooldown = 15 * time.Minute

// LogEvent يسجّل حدث أمني (يظهر بلوحة المراقبة مال المالك).
func (r *SecurityLockoutRepository) LogEvent(employeeID *string, employeeName, kind, detail, ip, userAgent string) error {
	_, err := r.db.Exec(`
		INSERT INTO "SecurityEvent" ("employeeId", "employeeName", kind, detail, ip, "userAgent")
		VALUES ($1, NULLIF($2,''), $3, NULLIF($4,''), NULLIF($5,''), NULLIF($6,''))
	`, employeeID, employeeName, kind, detail, ip, userAgent)
	return err
}

// RegisterFailedLogin يزيد عدّاد المحاولات الفاشلة المتتالية ويرجّع العدد
// الجديد. لما يوصل الحد، ينحظر الحساب.
func (r *SecurityLockoutRepository) RegisterFailedLogin(employeeID string) (streak int, locked bool, err error) {
	err = r.db.Get(&streak, `
		UPDATE "Employee" SET "failedLoginStreak" = "failedLoginStreak" + 1
		WHERE id = $1 RETURNING "failedLoginStreak"`, employeeID)
	if err != nil {
		return 0, false, err
	}
	if streak < FailedLoginThreshold {
		return streak, false, nil
	}
	// تعطيل مؤقت: نسجّل لحظة التعطيل بس ما نغيّر حالة الحساب — ينفك لحاله
	// بعد انتهاء المدة، ويظهر للمالك بلوحة المراقبة طول ما هو فعّال.
	_, err = r.db.Exec(`
		UPDATE "Employee"
		SET "lockedAt" = now(), "lockedReason" = $2,
			"lockedDetail" = 'تعطيل مؤقت — ينتهي تلقائياً بعد 15 دقيقة'
		WHERE id = $1 AND role NOT IN ('OWNER', 'ADMIN')`,
		employeeID, LockReasonFailedLogins)
	if err != nil {
		return streak, false, err
	}
	return streak, true, nil
}

// IsTemporarilyLocked هل التعطيل المؤقت لسه فعّال (ما انتهت مدته).
func IsTemporarilyLocked(lockedAt *time.Time, reason *string) bool {
	if lockedAt == nil || reason == nil || *reason != LockReasonFailedLogins {
		return false
	}
	return time.Since(*lockedAt) < FailedLoginCooldown
}

// ResetFailedLogins يصفّر العدّاد بعد دخول ناجح.
func (r *SecurityLockoutRepository) ResetFailedLogins(employeeID string) error {
	_, err := r.db.Exec(`UPDATE "Employee" SET "failedLoginStreak" = 0 WHERE id = $1`, employeeID)
	return err
}

// Lock يحظر الحساب. المالك ومدير النظام ما ينحظرون تلقائياً أبداً — وإلا
// ممكن نقفل النظام على نفسنا وما يبقى أحد يقدر يفك الحظر.
//
// نستخدم حالة SUSPENDED الموجودة أصلاً بالـenum (ما نضيف قيمة جديدة للنوع)،
// والتمييز بين الإيقاف اليدوي والحظر التلقائي يجي من lockedAt/lockedReason.
func (r *SecurityLockoutRepository) Lock(employeeID, reason, detail string) (bool, error) {
	var affected int
	err := r.db.Get(&affected, `
		WITH upd AS (
			UPDATE "Employee"
			SET status = 'SUSPENDED', "lockedAt" = now(), "lockedReason" = $2, "lockedDetail" = $3,
				-- الحظر يبطل جلساته الحالية فوراً، مو بس يمنع دخول جديد
				"sessionsInvalidatedAt" = now()
			WHERE id = $1 AND role NOT IN ('OWNER', 'ADMIN') AND "lockedAt" IS NULL
			RETURNING 1
		)
		SELECT COUNT(*) FROM upd`, employeeID, reason, detail)
	return affected > 0, err
}

// Unlock يفك الحظر — للمالك حصراً (الراوت محمي بـrequireOwner).
func (r *SecurityLockoutRepository) Unlock(employeeID string) error {
	_, err := r.db.Exec(`
		UPDATE "Employee"
		SET status = 'ACTIVE', "lockedAt" = NULL, "lockedReason" = NULL, "lockedDetail" = NULL,
			"failedLoginStreak" = 0, "authzViolations" = 0
		WHERE id = $1`, employeeID)
	return err
}

// ClearTemporaryLock يشيل التعطيل المؤقت بعد انتهاء مدته.
func (r *SecurityLockoutRepository) ClearTemporaryLock(employeeID string) error {
	_, err := r.db.Exec(`
		UPDATE "Employee"
		SET "lockedAt" = NULL, "lockedReason" = NULL, "lockedDetail" = NULL, "failedLoginStreak" = 0
		WHERE id = $1 AND "lockedReason" = $2`, employeeID, LockReasonFailedLogins)
	return err
}

// InvalidateSessions يبطل كل التوكنات الصادرة قبل هالحظة لهذا الموظف.
func (r *SecurityLockoutRepository) InvalidateSessions(employeeID string) error {
	_, err := r.db.Exec(`UPDATE "Employee" SET "sessionsInvalidatedAt" = now() WHERE id = $1`, employeeID)
	return err
}

// LockedEmployees قائمة الحسابات المحظورة حالياً مع سبب كل واحد.
func (r *SecurityLockoutRepository) LockedEmployees() ([]model.LockedEmployee, error) {
	rows := []model.LockedEmployee{}
	err := r.db.Select(&rows, `
		SELECT id, name, username, role, "lockedAt", "lockedReason", "lockedDetail",
			"failedLoginStreak", "authzViolations"
		FROM "Employee"
		WHERE "lockedAt" IS NOT NULL
		ORDER BY "lockedAt" DESC NULLS LAST`)
	return rows, err
}

// RecentEvents آخر الأحداث الأمنية.
func (r *SecurityLockoutRepository) RecentEvents(limit int) ([]model.SecurityEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows := []model.SecurityEvent{}
	err := r.db.Select(&rows, `
		SELECT * FROM "SecurityEvent" ORDER BY "createdAt" DESC LIMIT $1`, limit)
	return rows, err
}
