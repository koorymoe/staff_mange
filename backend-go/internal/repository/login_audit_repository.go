package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type LoginAuditRepository struct {
	db *sqlx.DB
}

func NewLoginAuditRepository(db *sqlx.DB) *LoginAuditRepository {
	return &LoginAuditRepository{db: db}
}

func (r *LoginAuditRepository) Record(username string, employeeID *string, success bool, ip, userAgent string) error {
	_, err := r.db.Exec(`
		INSERT INTO "LoginAudit" (id, username, "employeeId", success, "ipAddress", "userAgent")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
	`, username, employeeID, success, ip, userAgent)
	return err
}

// Recent يرجع آخر N محاولة دخول (ناجحة وفاشلة) — أساس لوحة المراقبة الأمنية.
func (r *LoginAuditRepository) Recent(limit int) ([]model.LoginAudit, error) {
	logs := []model.LoginAudit{}
	if err := r.db.Select(&logs, `SELECT * FROM "LoginAudit" ORDER BY "createdAt" DESC LIMIT $1`, limit); err != nil {
		return nil, err
	}
	for i := range logs {
		if logs[i].EmployeeID != nil {
			var brief model.EmployeeBrief
			if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, *logs[i].EmployeeID); err == nil {
				logs[i].Employee = &brief
			}
		}
	}
	return logs, nil
}

// FailedAttemptsCount يرجع عدد محاولات الدخول الفاشلة خلال آخر ساعة — مؤشر أمني بسيط.
func (r *LoginAuditRepository) FailedAttemptsCount() (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "LoginAudit" WHERE success = false AND "createdAt" > now() - interval '1 hour'`)
	return count, err
}

// OnlineEmployeesCount يرجع عدد الموظفين المميزين اللي دخلوا خلال آخر 15 دقيقة —
// تقدير بسيط لعدد "المتواجدين حالياً" بدون الحاجة لنظام جلسات حي منفصل.
func (r *LoginAuditRepository) OnlineEmployeesCount() (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT "employeeId") FROM "LoginAudit"
		WHERE success = true AND "employeeId" IS NOT NULL AND "createdAt" > now() - interval '15 minutes'
	`)
	return count, err
}
