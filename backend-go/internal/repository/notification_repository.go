package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"log"

	"staffmange-api/internal/model"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// logFail — أغلب المستدعين (٥٠+ مكان) يكتبون `_ = notif.Create(...)`، فإذا
// فشل الإرسال (قرار إجازة، طلب حذف، غرامة، تنبيه أمني) چان يضيع بلا أثر.
// هسه كل فشل ينكتب بالسجل من مكان واحد.
func logFail(kind, target, notifType string, err error) error {
	if err != nil {
		log.Printf("notification %s failed (target=%s type=%s): %v", kind, target, notifType, err)
	}
	return err
}

// Create ينشئ إشعاراً لموظف واحد.
func (r *NotificationRepository) Create(employeeID, notifType, message string) error {
	_, err := r.db.Exec(`
		INSERT INTO "Notification" (id, "employeeId", type, message)
		VALUES (gen_random_uuid()::text, $1, $2, $3)
	`, employeeID, notifType, message)
	return logFail("Create", employeeID, notifType, err)
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
	return logFail("CreateForRole", role, notifType, err)
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

// CreateForPermission ينبّه كل من عنده صلاحية معيّنة.
//
// نستعمله بطلبات الإجازة: التنبيه يروح للمخوّل بهذا المسار بالذات، فلو
// انتقلت المسؤولية من شخص لشخص تنتقل معها التنبيهات — بدون تعديل كود.
func (r *NotificationRepository) CreateForPermission(permissionName, notifType, message string) error {
	_, err := r.db.Exec(`
		INSERT INTO "Notification" (id, "employeeId", type, message)
		SELECT gen_random_uuid()::text, e.id, $2, $3
		FROM "Employee" e
		JOIN "EmployeePermission" ep ON ep."employeeId" = e.id
		JOIN "Permission" p ON p.id = ep."permissionId"
		WHERE p.name = $1 AND e.status = 'ACTIVE'
	`, permissionName, notifType, message)
	return logFail("CreateForPermission", permissionName, notifType, err)
}

// CreateForRolesOrPermission ينبّه كل موظف نشط دوره من الأدوار المعطاة
// أو عنده الصلاحية — مرة وحدة للشخص حتى لو انطبق عليه الشرطين.
func (r *NotificationRepository) CreateForRolesOrPermission(roles []string, permissionName, notifType, message string) error {
	_, err := r.db.Exec(`
		INSERT INTO "Notification" (id, "employeeId", type, message)
		SELECT gen_random_uuid()::text, e.id, $3, $4
		FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND (
			e.role::text = ANY($1) OR EXISTS (
				SELECT 1 FROM "EmployeePermission" ep
				JOIN "Permission" p ON p.id = ep."permissionId"
				WHERE ep."employeeId" = e.id AND p.name = $2))
	`, pq.Array(roles), permissionName, notifType, message)
	return logFail("CreateForRolesOrPermission", permissionName, notifType, err)
}
