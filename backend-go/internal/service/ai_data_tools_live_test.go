package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// وصول الذكاء الشامل: قراءة بس، والأسرار والهواتف ممنوعة على الاثنين،
// وأسماء الزبائن ممنوعة على ماتركس ومسموحة لايريس.
func TestAIDataAccess_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	database.EnsureAIReaders(db)
	if !database.AIReadersReady {
		t.Skip("مستخدم القاعدة ما يكدر ينشئ أدوار")
	}
	cleanup := func() {
		db.Exec(`DELETE FROM "Customer" WHERE id = 'aid-c'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'aid-e'`)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status) VALUES ('aid-e', 'موظف وصول', 'ACTIVE')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('aid-c', 'زبون وصول', '07700001239')`)

	run := func(plain bool, sql string) (string, error) {
		a := NewMatrixAgent(db, nil, "")
		if plain {
			a.ShowNames()
		}
		a.registerDataTools()
		for _, tl := range a.tools {
			if tl.Name == "query_data" {
				return tl.Run(map[string]any{"sql": sql}, nil)
			}
		}
		t.Fatal("no query_data")
		return "", nil
	}
	// قراءة عادية
	if out, err := run(false, `SELECT name, status FROM "Employee" WHERE id = 'aid-e'`); err != nil || !strings.Contains(out, "موظف وصول") {
		t.Fatalf("matrix read employee: %q %v", out, err)
	}
	// أسرار ممنوعة على الاثنين
	for _, plain := range []bool{false, true} {
		if _, err := run(plain, `SELECT password FROM "Employee" LIMIT 1`); err == nil {
			t.Fatalf("password readable (plain=%v)", plain)
		}
		if _, err := run(plain, `SELECT phone FROM "Customer" WHERE id = 'aid-c'`); err == nil {
			t.Fatalf("phone readable (plain=%v)", plain)
		}
		if _, err := run(plain, `SELECT * FROM "SecretaryNote" LIMIT 1`); err == nil {
			t.Fatalf("owner diary readable (plain=%v)", plain)
		}
	}
	// اسم الزبون: ماتركس لا، ايريس اي
	if _, err := run(false, `SELECT name FROM "Customer" WHERE id = 'aid-c'`); err == nil {
		t.Fatal("matrix can read customer name")
	}
	if out, err := run(true, `SELECT name FROM "Customer" WHERE id = 'aid-c'`); err != nil || !strings.Contains(out, "زبون وصول") {
		t.Fatalf("iris customer name: %q %v", out, err)
	}
	// الكتابة مرفوضة
	for _, q := range []string{`DELETE FROM "Employee"`, `SELECT 1; DELETE FROM "Employee"`, `WITH x AS (DELETE FROM "Employee" RETURNING id) SELECT * FROM x`, `SELECT set_config('role','staffmange',true)`} {
		if _, err := run(true, q); err == nil {
			t.Fatalf("write/escape allowed: %s", q)
		}
	}
	var n int
	db.Get(&n, `SELECT COUNT(*) FROM "Employee" WHERE id = 'aid-e'`)
	if n != 1 {
		t.Fatal("employee row vanished")
	}
}
