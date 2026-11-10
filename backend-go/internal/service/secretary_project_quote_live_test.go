package service

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// TestKamalProjectQuote_Live (دازاي 11-06): كمال يسوي «عرض مشاريع» — الغلاف والملحق ينحطون،
// والبند يسحب مواصفاته وموديله من الكتالوج، والسعر يبقى مثل ما انكتب.
func TestKamalProjectQuote_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	emp := "kq-" + uuid.NewString()[:8]
	prod := "لابتوب فحص كمال " + uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, role, status) VALUES ($1, 'مالك فحص', 'OWNER', 'ACTIVE')`, emp)
	db.MustExec(`INSERT INTO "Product" (id, name, unit, "defaultPrice", "modelName", specs) VALUES ($1, $2, 'قطعة', 900000, 'LENOVO V15', 'المعالج: i5')`, uuid.NewString(), prod)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Quotation" WHERE "createdByEmployeeId" = $1`, emp) //nolint:errcheck
		db.Exec(`DELETE FROM "Product" WHERE name = $1`, prod)                   //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)                     //nolint:errcheck
	})
	a := NewMatrixAgent(db, nil, "")
	tc := &SecretaryTurnCtx{UserID: emp}
	registerQuoteTools(a, NewQuotationService(repository.NewQuotationRepository(db), repository.NewPermissionRepository(db)), tc)
	var draft agentTool
	for _, x := range a.tools {
		if x.Name == "draft_quotation" {
			draft = x
		}
	}
	out, err := draft.Run(map[string]any{
		"customer_name": "جامعة فحص", "layout": "PROJECT", "cover_title": "تجهيز المختبرات",
		"items": []any{map[string]any{"name": prod, "quantity": float64(50), "unit_price": float64(925000), "category": "أنظمة الحوسبة"}},
	}, nil)
	if err != nil || !strings.Contains(out, "انحفظت") {
		t.Fatalf("%v %s", err, out)
	}
	var q struct {
		Layout string  `db:"layout"`
		Cover  string  `db:"cover"`
		Annex  *string `db:"annex"`
		Model  *string `db:"modelName"`
		Specs  *string `db:"specs"`
		Cat    *string `db:"category"`
		Price  float64 `db:"unitPrice"`
	}
	if err := db.Get(&q, `SELECT q.layout, q.cover::text AS cover, q.annex::text AS annex, i."modelName", i.specs, i.category, i."unitPrice"
		FROM "Quotation" q JOIN "QuotationItem" i ON i."quotationId" = q.id WHERE q."createdByEmployeeId" = $1`, emp); err != nil {
		t.Fatal(err)
	}
	if q.Layout != "PROJECT" || !strings.Contains(q.Cover, "تجهيز المختبرات") || q.Annex == nil || !strings.Contains(*q.Annex, "الإشراف الفني") {
		t.Fatalf("الغلاف/الملحق: %+v", q)
	}
	if q.Model == nil || *q.Model != "LENOVO V15" || q.Specs == nil || q.Cat == nil || *q.Cat != "أنظمة الحوسبة" || q.Price != 925000 {
		t.Fatalf("البند: %+v", q)
	}
	if len(tc.Cards) == 0 || tc.Cards[0].Link == nil || !strings.Contains(*tc.Cards[0].Link, "/quotations/project/") {
		t.Fatal("البطاقة لازم تفتح عرض المشاريع")
	}
}
