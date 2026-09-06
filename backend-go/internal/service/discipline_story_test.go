package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// fakeStories يلتقط القصص بدل ما يخزنها — نفحص شنو انبعث بالضبط.
type fakeStories struct{ got []model.EmitStoryRequest }

func (f *fakeStories) Emit(req model.EmitStoryRequest) { f.got = append(f.got, req) }

// TestDisciplinePenalizeStory_Live يحرس شغلتين بمسار الغرامة:
//
// ① الغرامة تولّد قصة **مربوطة بحدثها الرسمي** (`eventId` = id صف
//    `DisciplineEvent`) — بدونها ينكسر مفتاح الـidempotency.
// ② ⚠️⚠️ **والغرامة تشتغل حتى لو انطفى محرّك القصص** — هذا الفحص
//    الحاسم: القصة إضافة على العقوبة مو شرط لها، وأي انقلاب بهالترتيب
//    يخلي إجراءً مالياً يعتمد على ميزة عرض.
func TestDisciplinePenalizeStory_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	const emp = "disc-story-emp"
	cleanup := func() {
		db.Exec(`DELETE FROM "DisciplineEvent" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "DisciplinePoints" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	if _, err := db.Exec(`INSERT INTO "Employee"(id, name) VALUES ($1, 'موظف قصة الانضباط')`, emp); err != nil {
		t.Fatalf("seed employee: %v", err)
	}
	t.Cleanup(cleanup)

	discRepo := repository.NewDisciplineRepository(db)
	svc := NewDisciplineService(discRepo, nil, nil, repository.NewEmployeeRepository(db))

	// ② أولاً: بلا محرّك قصص إطلاقاً — الغرامة لازم تنزل مثل ما هي
	svc.penalize(emp, "موظف قصة الانضباط", "GUARD_NO_STORIES", "غرامة بلا محرّك قصص", nil)
	var n int
	if err := db.Get(&n, `SELECT count(*) FROM "DisciplineEvent" WHERE "employeeId" = $1 AND kind = 'GUARD_NO_STORIES'`, emp); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Fatalf("الغرامة ما انسجّلت بلا محرّك قصص — القصة صارت شرطاً للعقوبة (لگينا %d)", n)
	}

	// ① وبعدها مع المحرّك: قصة وحدة، ومعرّفها من الجدول
	fake := &fakeStories{}
	svc.SetStories(fake)
	svc.penalize(emp, "موظف قصة الانضباط", "GUARD_WITH_STORIES", "ورق متأخر", nil)
	if len(fake.got) != 1 {
		t.Fatalf("توقعنا قصة وحدة، انبعث %d", len(fake.got))
	}
	req := fake.got[0]
	if req.EventKind != model.StoryEventPaperMissing {
		t.Fatalf("نوع الحدث غلط: %s", req.EventKind)
	}
	if req.RecipientID != emp {
		t.Fatalf("المستلم غلط: %s", req.RecipientID)
	}
	if err := db.Get(&n, `SELECT count(*) FROM "DisciplineEvent" WHERE id = $1`, req.EventID); err != nil {
		t.Fatalf("count event: %v", err)
	}
	if n != 1 {
		t.Fatal("eventId مالت القصة مو id صف حقيقي بـDisciplineEvent")
	}

	// ⚠️ وكل غرامة **معرّفها لحاله**: القصتان ما يتصادمان بالفهرس الفريد
	svc.penalize(emp, "موظف قصة الانضباط", "GUARD_SECOND", "تأخير ثانٍ", nil)
	if len(fake.got) != 2 {
		t.Fatalf("توقعنا قصتين، انبعث %d", len(fake.got))
	}
	if fake.got[0].EventID == fake.got[1].EventID {
		t.Fatal("غرامتان مختلفتان بنفس المعرّف — الثانية تنبلع بالفهرس الفريد")
	}
}
