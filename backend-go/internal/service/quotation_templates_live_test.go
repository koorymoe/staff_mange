package service

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

func TestQuotationTemplateSaveUseLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	emp, q, sfx := uuid.NewString(), uuid.NewString(), uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'منشئ', $2, 'ADMIN', 'ACTIVE')`, emp, "qt_"+sfx)
	db.MustExec(`INSERT INTO "Quotation" (id, "quotationNumber", "customerName", "customerPhone", "createdByEmployeeId", layout, terms, duration)
		VALUES ($1, $2, 'زبون سري', '07700000000', $3, 'PROJECT', '["دفعة ٥٠٪"]', '٣٠ يوم')`, q, "QT-T-"+sfx, emp)
	db.MustExec(`INSERT INTO "QuotationItem" (id, "quotationId", "productName", quantity, "unitPrice", "totalPrice", "sortIndex", category)
		VALUES ($1, $2, 'لوح شمسي 550', 10, 100, 1000, 0, 'طاقة شمسية'), ($3, $2, 'انفرتر 10k', 1, 900, 900, 1, 'طاقة شمسية')`, uuid.NewString(), q, uuid.NewString())
	created := []string{}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "QuotationTemplate" WHERE "createdById" = $1`, emp) //nolint:errcheck
		for _, id := range append(created, q) {
			db.Exec(`DELETE FROM "Quotation" WHERE id = $1`, id) //nolint:errcheck
		}
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp) //nolint:errcheck
	})
	s := NewQuotationTemplateService(db)
	id, sum, err := s.Save(emp, "ADMIN", "QT-T-"+sfx, "نموذج شمسي "+sfx, "")
	if err != nil || !strings.Contains(sum, "طاقة شمسية (2 بند)") {
		t.Fatalf("الحفظ: %v | %s", err, sum)
	}
	var data string
	db.Get(&data, `SELECT data::text FROM "QuotationTemplate" WHERE id = $1`, id) //nolint:errcheck
	if strings.Contains(data, "زبون سري") || strings.Contains(data, "0770") {
		t.Fatal("بيانات الزبون انحفظت بالنموذج")
	}
	qty, price := 20, 120.0
	nq, err := s.Use(emp, "ADMIN", "نموذج شمسي "+sfx, TemplateUse{CustomerName: "زبون ثاني", Items: []TemplateItemChange{
		{ProductName: "لوح شمسي 550", Quantity: &qty, UnitPrice: &price},
		{ProductName: "انفرتر 10k", Remove: true},
		{ProductName: "بطارية ليثيوم", Category: "طاقة شمسية", UnitPrice: &price},
	}, Terms: []string{"دفعة ٧٠٪"}})
	if err != nil {
		t.Fatal(err)
	}
	created = append(created, nq.ID)
	items := []struct {
		N string  `db:"productName"`
		Q int     `db:"quantity"`
		P float64 `db:"unitPrice"`
	}{}
	db.Select(&items, `SELECT "productName", quantity, "unitPrice" FROM "QuotationItem" WHERE "quotationId" = $1 ORDER BY "sortIndex"`, nq.ID) //nolint:errcheck
	if len(items) != 2 || items[0].N != "لوح شمسي 550" || items[0].Q != 20 || items[0].P != 120 || items[1].N != "بطارية ليثيوم" {
		t.Fatalf("البنود: %+v", items)
	}
	var terms, layout string
	db.Get(&terms, `SELECT terms::text FROM "Quotation" WHERE id = $1`, nq.ID)  //nolint:errcheck
	db.Get(&layout, `SELECT layout FROM "Quotation" WHERE id = $1`, nq.ID)      //nolint:errcheck
	if !strings.Contains(terms, "٧٠٪") || layout != "PROJECT" {
		t.Fatalf("الشروط/الشكل: %s %s", terms, layout)
	}
	if s.CanUse(uuid.NewString(), "TECHNICIAN") {
		t.Fatal("فني بلا صلاحية يگدر")
	}
}
