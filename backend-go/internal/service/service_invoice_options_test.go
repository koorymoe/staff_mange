package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// (ل) 11-24 دازاي: الشغل = مجموع الخيارات، وإذا أقل من الحد الأدنى يطلع الحد؛ المواد فوگه.
func TestQuoteFor(t *testing.T) {
	tpl := ServiceInvoiceTemplate{MinWork: 14000, Options: []ServiceInvoiceOption{
		{ID: "a", Label: "تسليك كيبل", Unit: "متر", UnitPrice: 500},
		{ID: "b", Label: "غسل لوح", Unit: "لوح", UnitPrice: 2000},
	}}
	q := QuoteFor(tpl, map[string]float64{"a": 10, "b": 2}, 0) // 5000 + 4000 = 9000 ← الحد 14000
	if q.Sum != 9000 || q.Work != 14000 || !q.MinUsed || q.Total != 14000 || len(q.Lines) != 2 {
		t.Fatalf("الحد الأدنى ما انطبق: %+v", q)
	}
	q = QuoteFor(tpl, map[string]float64{"b": 10}, 6000) // 20000 فوگ الحد + مواد
	if q.Work != 20000 || q.MinUsed || q.Total != 26000 {
		t.Fatalf("الحساب غلط: %+v", q)
	}
	q = QuoteFor(tpl, map[string]float64{"x": 5, "a": -3}, 0) // خيار مو موجود أو سالب ← يتجاهله
	if len(q.Lines) != 0 || q.Work != 14000 {
		t.Fatalf("لازم يتجاهل الغلط: %+v", q)
	}
}

func TestServiceInvoiceOptionsLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	db.Exec(`DELETE FROM "Service" WHERE id = 'sio-svc'`)
	t.Cleanup(func() { db.Exec(`DELETE FROM "Service" WHERE id = 'sio-svc'`) })
	db.MustExec(`INSERT INTO "Service"(id, name) VALUES ('sio-svc', 'تنظيف ألواح sio')`)
	s := NewServiceInvoiceOptions(db)
	if err := s.Save("sio-svc", "", 14000, nil, []ServiceInvoiceOption{{Label: " غسل لوح ", Unit: "لوح", UnitPrice: 2000}, {Label: "صعود سطح", UnitPrice: 3000}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Save("sio-svc", "", 0, nil, []ServiceInvoiceOption{{Label: "", UnitPrice: 1}}); err == nil {
		t.Fatal("خيار بلا اسم لازم ينرفض")
	}
	if err := s.Save("no-such", "", 0, nil, nil); err != ErrDesignHidden {
		t.Fatalf("خدمة مو موجودة: %v", err)
	}
	ts, _ := s.List()
	var tpl *ServiceInvoiceTemplate
	for i := range ts {
		if ts[i].ServiceID == "sio-svc" {
			tpl = &ts[i]
		}
	}
	if tpl == nil || tpl.MinWork != 14000 || len(tpl.Options) != 2 || tpl.Options[0].Label != "غسل لوح" || tpl.Options[1].Unit != "قطعة" {
		t.Fatalf("الحفظ غلط: %+v", tpl)
	}
	q, err := s.Quote("sio-svc", map[string]float64{tpl.Options[0].ID: 3}, 0)
	if err != nil || q.Work != 14000 || q.Sum != 6000 {
		t.Fatalf("التجربة غلط: %+v %v", q, err)
	}
}

// (دازاي 11-28) «إذا طلعت أقل من 14 ينحسب 14، وإذا أكثر من 14 ينحسب أقل شي 24».
func TestQuoteForTiers(t *testing.T) {
	tpl := ServiceInvoiceTemplate{MinTiers: []float64{24000, 14000}, Options: []ServiceInvoiceOption{{ID: "a", UnitPrice: 1000}}}
	for _, c := range []struct{ n, want float64 }{{5, 14000}, {14, 14000}, {16, 24000}, {30, 30000}} {
		if q := QuoteFor(tpl, map[string]float64{"a": c.n}, 0); q.Work != c.want {
			t.Fatalf("n=%v: work=%v want %v", c.n, q.Work, c.want)
		}
	}
}
