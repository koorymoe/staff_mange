package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (دازاي 11-16) من يتغيّر دور الموظف، إشعارات دوره القديم ما تطلعله — بس الجديدة.
func TestNotificationsHiddenAfterRoleChange_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	for _, v := range []string{"0502_employee_role_changed_at", "0503_employee_prev_role"} {
		if err := database.ApplyMigrationForTest(db, v); err != nil {
			t.Fatal(err)
		}
	}
	const emp = "role-change-notif-emp"
	cleanup := func() { db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp) }
	cleanup()
	t.Cleanup(cleanup)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ($1, 'موظف تغيير دور', 'HR_COORDINATOR')`, emp)
	repo := NewNotificationRepository(db)
	db.MustExec(`INSERT INTO "Notification"(id, "employeeId", type, message, "createdAt") VALUES ('rc-old', $1, 'X', 'قديم', CURRENT_TIMESTAMP - interval '1 day')`, emp)

	// تعديل بلا تغيير دور: القديم يبقى
	db.MustExec(`UPDATE "Employee" SET name = 'موظف تغيير دور ٢' WHERE id = $1`, emp)
	if l, _ := repo.ListForEmployee(emp, 10); len(l) != 1 {
		t.Fatalf("بلا تغيير دور لازم يبقى القديم: %d", len(l))
	}

	db.MustExec(`UPDATE "Employee" SET role = 'TECHNICIAN', "isLeader" = true WHERE id = $1`, emp)
	db.MustExec(`INSERT INTO "Notification"(id, "employeeId", type, message) VALUES ('rc-new', $1, 'X', 'جديد')`, emp)
	l, err := repo.ListForEmployee(emp, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(l) != 1 || l[0].ID != "rc-new" {
		t.Fatalf("لازم يطلع الجديد بس: %+v", l)
	}
	if n, _ := repo.UnreadCount(emp); n != 1 {
		t.Fatalf("العدّاد لازم ١، طلع %d", n)
	}
	// والدور القبلي ينحفظ — للتقييم حسب كل دور
	var prev string
	var prevLeader bool
	if err := db.QueryRow(`SELECT "prevRole", "prevIsLeader" FROM "Employee" WHERE id = $1`, emp).Scan(&prev, &prevLeader); err != nil || prev != "HR_COORDINATOR" || prevLeader {
		t.Fatalf("الدور القبلي غلط: %q %v %v", prev, prevLeader, err)
	}
}
