package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
)

// TestStoryQueue_Live يحرس الضمانات الأربع الي يقوم عليها محرّك
// القصص. كل وحدة منهن **علّة حقيقية لو انكسرت**:
//
// ① التكرار: نفس الحدث ينشئ قصتين ← الموظف يستلم نفس الخصم مرتين.
// ② التجميع: ثلاث أوراق ناقصة بنفس الحجز ← ثلاث ركضات ورا بعض،
//    والكيان يصير مزعجاً وينتجاهل — ووقتها العقوبة تفقد أثرها.
// ③ الأولوية: التهنئة تسبق التحذير ← موظف متأخر يستلم احتفالاً.
// ④ المرحلة ما ترجع للورا: استطلاع متأخر يرجّع «أقرّ» لـ«وصل» ←
//    نخسر حقيقة مسجَّلة، ونعيد عرض مشهد انتهى.
func TestStoryQueue_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	const emp = "story-test-emp"
	db.Exec(`DELETE FROM "StoryInstance" WHERE "recipientRef" = $1`, emp)
	db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	if _, err := db.Exec(`INSERT INTO "Employee"(id, name) VALUES ($1, 'موظف اختبار القصص')`, emp); err != nil {
		t.Fatalf("seed employee: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "StoryInstance" WHERE "recipientRef" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	})

	r := NewStoryRepository(db)

	// ① نفس الحدث مرتين ← قصة وحدة
	base := model.StoryInstance{
		EventID: "story-test-ev-1", EventKind: model.StoryEventPointDeducted,
		StoryType: model.StoryEventPointDeducted, RecipientRef: emp, RecipientName: "موظف اختبار القصص",
		Priority: model.StoryPriority[model.StoryEventPointDeducted], Physical: true,
		Payload: json.RawMessage(`{"title":"خصم"}`),
	}
	if created, err := r.Enqueue(base); err != nil || !created {
		t.Fatalf("الإدراج الأول لازم ينجح: created=%v err=%v", created, err)
	}
	base.ID = "" // معرّف جديد، نفس الحدث
	if created, err := r.Enqueue(base); err != nil || created {
		t.Fatalf("نفس الحدث لازم ما ينشئ قصة ثانية: created=%v err=%v", created, err)
	}

	// ② التجميع: ورقة ناقصة ثانية بنفس الحجز تندمج
	group := "paper:BK-1"
	grouped := model.StoryInstance{
		EventID: "story-test-ev-2", EventKind: model.StoryEventPaperMissing,
		StoryType: model.StoryEventPaperMissing, RecipientRef: emp, RecipientName: "موظف اختبار القصص",
		Priority: model.StoryPriority[model.StoryEventPaperMissing], Physical: true,
		GroupKey: &group, Payload: json.RawMessage(`{"mergedCount":1,"lines":["تقرير"]}`),
	}
	if created, err := r.Enqueue(grouped); err != nil || !created {
		t.Fatalf("قصة المجموعة الأولى لازم تنجح: %v %v", created, err)
	}
	merged, err := r.MergeIntoGroup(emp, group, "فاتورة")
	if err != nil || !merged {
		t.Fatalf("الدمج لازم ينجح: merged=%v err=%v", merged, err)
	}
	var count int
	var lines string
	if err := db.QueryRow(`SELECT (payload->>'mergedCount')::int, payload->>'lines'
	                       FROM "StoryInstance" WHERE "eventId" = 'story-test-ev-2'`).Scan(&count, &lines); err != nil {
		t.Fatalf("قراءة المدموج: %v", err)
	}
	if count != 2 {
		t.Fatalf("العدّاد لازم يصير ٢ بعد الدمج، صار %d", count)
	}

	// ③ الأولوية: تهنئة أقدم ما تسبق تحذيراً أحدث
	praise := model.StoryInstance{
		EventID: "story-test-ev-3", EventKind: model.StoryEventPraise,
		StoryType: model.StoryEventPraise, RecipientRef: emp, RecipientName: "موظف اختبار القصص",
		Priority: model.StoryPriority[model.StoryEventPraise], Physical: true,
	}
	if _, err := r.Enqueue(praise); err != nil {
		t.Fatalf("إدراج التهنئة: %v", err)
	}
	db.Exec(`UPDATE "StoryInstance" SET "createdAt" = now() - interval '1 hour' WHERE "eventId" = 'story-test-ev-3'`)
	next, err := r.NextForEmployee(emp)
	if err != nil || next == nil {
		t.Fatalf("لازم ترجّع قصة: %v", err)
	}
	if next.EventKind != model.StoryEventPointDeducted {
		t.Fatalf("التحذير لازم يسبق الفرح — رجّعت %s", next.EventKind)
	}

	// ④ المرحلة ما ترجع للورا
	if err := r.Advance(next.ID, emp, model.StoryStatusAcknowledged, 7); err != nil {
		t.Fatalf("التقدّم لـACKNOWLEDGED: %v", err)
	}
	if err := r.Advance(next.ID, emp, model.StoryStatusDelivered, 1); err != nil {
		t.Fatalf("محاولة الرجوع ما لازم ترجّع خطأ: %v", err)
	}
	var status string
	db.QueryRow(`SELECT status FROM "StoryInstance" WHERE id = $1`, next.ID).Scan(&status)
	if status != model.StoryStatusAcknowledged {
		t.Fatalf("المرحلة رجعت للورا: صارت %s", status)
	}

	// ⑤ العزل: موظف ثاني ما يقدر يعلّم قصة مو مالته
	if err := r.Advance(next.ID, "someone-else", model.StoryStatusSeen, 1); err != nil {
		t.Fatalf("محاولة غريب ما لازم ترجّع خطأ: %v", err)
	}

	// ⑥ موظف نظيف: «ماكو قصة» مو خطأ
	if got, err := r.NextForEmployee("no-such-employee"); err != nil || got != nil {
		t.Fatalf("موظف بلا قصص لازم يرجّع (nil,nil) — got=%v err=%v", got, err)
	}
}

// TestStoryClaim_Live يحرس فجوة **نافذتان لنفس الموظف**: بلا حجز
// ذرّي، التبويبان يشغّلان نفس المشهد — نفس العقوبة تنعرض مرتين،
// والإقرار ينسجّل من وحدة والثانية تبقى تعرض شي انبتّ فيه.
func TestStoryClaim_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	const emp = "story-claim-emp"
	db.Exec(`DELETE FROM "StoryInstance" WHERE "recipientRef" = $1`, emp)
	db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	db.Exec(`INSERT INTO "Employee"(id, name) VALUES ($1, 'موظف حجز')`, emp)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "StoryInstance" WHERE "recipientRef" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	})

	r := NewStoryRepository(db)
	empID := emp
	if _, err := r.Enqueue(model.StoryInstance{
		EventID: "claim-ev-1", EventKind: model.StoryEventPointDeducted,
		StoryType: model.StoryEventPointDeducted,
		RecipientEmployeeID: &empID, RecipientRef: emp, RecipientName: "موظف حجز",
		Priority: 100, Physical: true,
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	story, err := r.NextForEmployee(emp)
	if err != nil || story == nil {
		t.Fatalf("لازم ترجّع قصة: %v", err)
	}

	first, err := r.Claim(story.ID, emp)
	if err != nil || !first {
		t.Fatalf("النافذة الأولى لازم تربح الحجز: %v %v", first, err)
	}
	second, err := r.Claim(story.ID, emp)
	if err != nil {
		t.Fatalf("الحجز الثاني ما لازم يرجّع خطأ: %v", err)
	}
	if second {
		t.Fatal("نافذتان حجزتا نفس القصة — نفس المشهد راح ينعرض مرتين")
	}
	// غريب ما يحجز قصة مو مالته
	if got, err := r.Claim(story.ID, "someone-else"); err != nil || got {
		t.Fatalf("غريب حجز قصة مو مالته: got=%v err=%v", got, err)
	}
}

// TestStoryDeleteEmployee_KeepsEvidence يحرس **العيب الي لگاه (م)**:
// الترحيل 0272 چان `ON DELETE CASCADE`، فحذف حساب الموظف **يمحي كل
// دليل إنه انخصم وأقرّ بالاطلاع** — وهذا يكسر نمط مطبَّق بأربع
// جداول ثانية بالمشروع (الاسم منسوخ نصاً والمفتاح SET NULL).
//
// ⚠️ ويفحص شغلة ثانية: الـidempotency **لازم تبقى شغّالة بعد الحذف**،
// ولذلك الفهرس الفريد على `recipientRef` الثابت مو على المفتاح الي
// يصير NULL.
func TestStoryDeleteEmployee_KeepsEvidence(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	const emp = "story-del-emp"
	db.Exec(`DELETE FROM "StoryInstance" WHERE "recipientRef" = $1`, emp)
	db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	db.Exec(`INSERT INTO "Employee"(id, name) VALUES ($1, 'موظف راح ينحذف')`, emp)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "StoryInstance" WHERE "recipientRef" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	})

	r := NewStoryRepository(db)
	empID := emp
	if _, err := r.Enqueue(model.StoryInstance{
		EventID: "del-ev-1", EventKind: model.StoryEventPointDeducted,
		StoryType: model.StoryEventPointDeducted,
		RecipientEmployeeID: &empID, RecipientRef: emp,
		RecipientName: "موظف راح ينحذف", Priority: 100, Physical: true,
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp); err != nil {
		t.Fatalf("حذف الموظف: %v", err)
	}

	var count int
	var name string
	var fk *string
	if err := db.QueryRow(`SELECT count(*) FROM "StoryInstance" WHERE "recipientRef" = $1`, emp).Scan(&count); err != nil {
		t.Fatalf("عدّ بعد الحذف: %v", err)
	}
	if count != 1 {
		t.Fatalf("انمحى دليل وصول العقوبة بعد حذف الموظف — صفوف=%d", count)
	}
	if err := db.QueryRow(`SELECT "recipientName", "recipientEmployeeId"
	                       FROM "StoryInstance" WHERE "recipientRef" = $1`, emp).Scan(&name, &fk); err != nil {
		t.Fatalf("قراءة بعد الحذف: %v", err)
	}
	if name == "" {
		t.Fatal("السطر صار مجهولاً — الاسم المنسوخ ضاع")
	}
	if fk != nil {
		t.Fatalf("المفتاح لازم يصير NULL بعد الحذف، صار %q", *fk)
	}

	// الـidempotency لازم تبقى شغّالة حتى بعد ما صار المفتاح NULL
	created, err := r.Enqueue(model.StoryInstance{
		EventID: "del-ev-1", EventKind: model.StoryEventPointDeducted,
		StoryType: model.StoryEventPointDeducted,
		RecipientRef: emp, RecipientName: "موظف راح ينحذف", Priority: 100,
	})
	if err != nil {
		t.Fatalf("إعادة الإدراج: %v", err)
	}
	if created {
		t.Fatal("نفس الحدث انشأ قصة ثانية بعد حذف الحساب — الـidempotency انكسرت")
	}
}
