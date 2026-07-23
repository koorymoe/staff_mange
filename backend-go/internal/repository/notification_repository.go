package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create ينشئ إشعاراً لموظف واحد.
func (r *NotificationRepository) Create(employeeID, notifType, message string) error {
	_, err := r.db.Exec(`
		INSERT INTO "Notification" (id, "employeeId", type, message)
		VALUES (gen_random_uuid()::text, $1, $2, $3)
	`, employeeID, notifType, message)
	return err
}

// CreateForRole يبث نفس الإشعار لكل الموظفين النشطين بدور معيّن (مثال: كل الفنيين
// يشوفون مين تصدر ترتيبهم الشهري).
func (r *NotificationRepository) CreateForRole(role, notifType, message string) error {
	_, err := r.db.Exec(`
		INSERT INTO "Notification" (id, "employeeId", type, message)
		SELECT gen_random_uuid()::text, id, $2, $3
		FROM "Employee"
		WHERE role = $1 AND status = 'ACTIVE'
	`, role, notifType, message)
	return err
}

func (r *NotificationRepository) ListForEmployee(employeeID string, limit int) ([]model.Notification, error) {
	notifications := []model.Notification{}
	err := r.db.Select(&notifications, `
		SELECT * FROM "Notification"
		WHERE "employeeId" = $1
		ORDER BY "createdAt" DESC
		LIMIT $2
	`, employeeID, limit)
	return notifications, err
}

func (r *NotificationRepository) UnreadCount(employeeID string) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "Notification" WHERE "employeeId" = $1 AND read = false`, employeeID)
	return count, err
}

func (r *NotificationRepository) MarkRead(id, employeeID string) error {
	_, err := r.db.Exec(`UPDATE "Notification" SET read = true WHERE id = $1 AND "employeeId" = $2`, id, employeeID)
	return err
}

func (r *NotificationRepository) MarkAllRead(employeeID string) error {
	_, err := r.db.Exec(`UPDATE "Notification" SET read = true WHERE "employeeId" = $1 AND read = false`, employeeID)
	return err
}
