package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestAssistantKnowledgeSearchRelevant_Live يتأكد فعلياً (على قاعدة بيانات حية،
// إذا متوفر DATABASE_URL بالبيئة) إن SearchRelevant يرجع سطور المعرفة المطابقة
// لكلمات مفتاحية بالرسالة، ويشيل بياناته بعد الاختبار (best-effort cleanup).
func TestAssistantKnowledgeSearchRelevant_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer db.Close()

	repo := NewAssistantKnowledgeRepository(db)

	var empID string
	if err := db.Get(&empID, `SELECT id FROM "Employee" LIMIT 1`); err != nil {
		t.Fatalf("could not find sample employee: %v", err)
	}

	testTopic1 := "__test_إنفرترات_الطاقة_الشمسية"
	testTopic2 := "__test_موضوع_غير_ذي_علاقة"
	if err := repo.Create(testTopic1, "الإنفرتر الهجين أفضل للاستخدام المنزلي حسب بحث سابق", empID); err != nil {
		t.Fatalf("create 1 error: %v", err)
	}
	if err := repo.Create(testTopic2, "معلومة عن شي ثاني كلياً ما إلها علاقة بالبحث", empID); err != nil {
		t.Fatalf("create 2 error: %v", err)
	}

	defer func() {
		db.MustExec(`DELETE FROM "AssistantKnowledge" WHERE topic IN ($1, $2)`, testTopic1, testTopic2)
	}()

	keywords := ExtractKeywords("شنو أحسن إنفرترات للألواح الشمسية؟")
	if len(keywords) == 0 {
		t.Fatalf("expected non-empty keywords")
	}
	t.Logf("extracted keywords: %v", keywords)

	rows, err := repo.SearchRelevant(keywords, 8)
	if err != nil {
		t.Fatalf("search error: %v", err)
	}

	found := false
	for _, r := range rows {
		if r.Topic == testTopic1 {
			found = true
		}
		if r.Topic == testTopic2 {
			t.Fatalf("unrelated row unexpectedly matched: %s", r.Topic)
		}
	}
	if !found {
		t.Fatalf("expected to find test row %s among results, got: %+v", testTopic1, rows)
	}
}
