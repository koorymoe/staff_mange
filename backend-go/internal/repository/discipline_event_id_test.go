package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDisciplineEventID_Live يحرس الشرط الي يقوم عليه ربط القصص
// بالانضباط: `Penalize` و`RestoreOne` لازم يرجّعان **id صف حقيقي**
// بجدول `DisciplineEvent`.
//
// ⚠️ ليش هذا حارس مو تفصيل: `eventId` بالقصة هو **مفتاح الـidempotency**
// (الفهرس الفريد `(eventKind, eventId, recipientRef)`). لو رجع فارغاً
// أو رجع رقماً مخترعاً مو من الجدول، نخسر الضمانة كلها — ونفس الغرامة
// تنعرض على الموظف كل مرة تمر المكنسة.
func TestDisciplineEventID_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	const emp = "disc-eventid-emp"
	cleanup := func() {
		db.Exec(`DELETE FROM "DisciplineEvent" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "DisciplinePoints" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	if _, err := db.Exec(`INSERT INTO "Employee"(id, name) VALUES ($1, 'موظف اختبار الانضباط')`, emp); err != nil {
		t.Fatalf("seed employee: %v", err)
	}
	t.Cleanup(cleanup)

	repo := NewDisciplineRepository(db)

	// ① الغرامة ترجّع id، والـid موجود فعلاً بالجدول
	applied, _, eventID, err := repo.Penalize(emp, "TEST_KIND", "اختبار إرجاع المعرّف", nil, 1)
	if err != nil {
		t.Fatalf("penalize: %v", err)
	}
	if !applied {
		t.Fatal("الغرامة ما انطبّقت — الاختبار ما يقدر يكمل")
	}
	if eventID == "" {
		t.Fatal("Penalize رجّع معرّفاً فارغاً — مفتاح الـidempotency ينكسر")
	}
	var n int
	if err := db.Get(&n, `SELECT count(*) FROM "DisciplineEvent" WHERE id = $1 AND "employeeId" = $2`, eventID, emp); err != nil {
		t.Fatalf("count penalize row: %v", err)
	}
	if n != 1 {
		t.Fatalf("المعرّف المرجَّع مو صف حقيقي بالجدول (لگينا %d صف)", n)
	}

	// ② ورجوع النقطة نفس الشي
	restoreID, err := repo.RestoreOne(emp, "اختبار إرجاع معرّف الرجوع")
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restoreID == "" {
		t.Fatal("RestoreOne رجّع معرّفاً فارغاً")
	}
	if err := db.Get(&n, `SELECT count(*) FROM "DisciplineEvent" WHERE id = $1 AND "employeeId" = $2 AND delta = 1`, restoreID, emp); err != nil {
		t.Fatalf("count restore row: %v", err)
	}
	if n != 1 {
		t.Fatalf("معرّف الرجوع مو صف حقيقي بالجدول (لگينا %d صف)", n)
	}

	// ③ والمعرّفان مختلفان — حدثان مختلفان
	if eventID == restoreID {
		t.Fatal("الغرامة والرجوع رجّعا نفس المعرّف — القصتان تتصادمان بالفهرس الفريد")
	}
}
