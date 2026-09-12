package database

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// seedOwnerAccount يزرع/يحدّث حساب مالك النظام، حساب واحد فوق كل شي حتى فوق ADMIN
// (يشوف كل شي بما فيه لوحة المراقبة الخلفية الحصرية)، ما يطلع بأي قائمة موظفين
// عادية، ومحد يقدر يمنح هذا الدور لغيره من الواجهة (مقفول بـ employee_service.go).
//
// بيانات الدخول تنجي حصراً من متغيرات البيئة OWNER_USERNAME/OWNER_PASSWORD —
// ممنوع أي قيمة افتراضية ثابتة بالكود (كانت هذي ثغرة أمنية حرجة سابقاً: حساب
// المالك الحقيقي بكلمة مرور مكتوبة نص صريح بالكود ومرفوعة لتاريخ Git للأبد).
// لو المتغيرين غير معرّفين، ما نلمس الحساب الموجود إطلاقاً (نتجنب قفل وصول
// المالك الحالي بالغلط لو نسى أحد يضيف المتغيرات وقت الترقية لهذا الإصدار).
func seedOwnerAccount(db *sqlx.DB, username, password string) error {
	if username == "" || password == "" {
		return nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO "Employee" (id, name, username, password, role, status, "jobTitle")
		VALUES ('owner_root', 'المالك', $1, $2, 'OWNER', 'ACTIVE', 'مالك النظام')
		ON CONFLICT (id) DO UPDATE SET username = $1, password = $2, role = 'OWNER', status = 'ACTIVE'
	`, username, string(hashed))
	return err
}

// grantGpsSystemToMonitors تضمن كل موظف بدور "مراقب" عنده صلاحية "gps_system"
// (مراقبة قسم الجي بي اس) — الجي بي اس صارت خدمة بصلاحية مو دور وظيفي منفصل،
// والمراقب المفروض يشوف ويتدخل بيها متل باقي الخدمات. idempotent بالكامل.
func grantGpsSystemToMonitors(db *sqlx.DB) error {
	return grantRolePermission(db, "MONITOR", "gps_system", "نظام GPS")
}

// grantLeaderBasketToLeaders تضمن كل موظف عنده isLeader=true (تيم ليدر فريق)
// عنده صلاحية "leader_basket" (سلة الليدر/فاتورة الليدر) تلقائياً — هذي الصلاحية
// الافتراضية لليدر بحكم دوره الفعلي (isLeader، مو دور وظيفي role عادي). الأدمن
// يقدر لاحقاً يمنحها يدوياً لموظف MONITOR أيضاً من صفحة الصلاحيات إذا احتاج،
// بدون ما يمس هذا الافتراضي. idempotent بالكامل.
func grantLeaderBasketToLeaders(db *sqlx.DB) error {
	if _, err := db.Exec(`
		INSERT INTO "Permission" (id, name, label)
		VALUES (gen_random_uuid()::text, 'leader_basket', 'سلة الليدر (فاتورة الليدر / المواد والمنظومات المختارة)')
		ON CONFLICT (name) DO NOTHING
	`); err != nil {
		return err
	}
	_, err := db.Exec(`
		INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
		SELECT gen_random_uuid()::text, e.id, p.id
		FROM "Employee" e, "Permission" p
		WHERE e."isLeader" = true AND p.name = 'leader_basket'
		ON CONFLICT ("employeeId", "permissionId") DO NOTHING
	`)
	return err
}

// grantRolePermission يضمن كل موظف بدور معيّن عنده صلاحية معيّنة — يستخدم
// لتحديث RoleDefaultPermissions بأثر رجعي على الموظفين الموجودين فعلاً
// (تغيير خارطة الصلاحيات بالكود لحاله ما يوصل تلقائياً لحسابات منشأة سابقاً).
func grantRolePermission(db *sqlx.DB, role, permissionName, permissionLabel string) error {
	if _, err := db.Exec(`
		INSERT INTO "Permission" (id, name, label)
		VALUES (gen_random_uuid()::text, $1, $2)
		ON CONFLICT (name) DO NOTHING
	`, permissionName, permissionLabel); err != nil {
		return err
	}
	_, err := db.Exec(`
		INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
		SELECT gen_random_uuid()::text, e.id, p.id
		FROM "Employee" e, "Permission" p
		WHERE e.role = $1 AND p.name = $2
		ON CONFLICT ("employeeId", "permissionId") DO NOTHING
	`, role, permissionName)
	return err
}
