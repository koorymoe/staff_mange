package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (ل) 11-25 دازاي: الموظف ياخذ أداة بنفسه ووقته، تبقى بعهدته، يسلّم بصورة، المالك يدقق
// (غلط ← ترجع لعهدته، صح ← ترجع للمخزن)؛ والمدير ياخذ عن موظف.
func TestToolCustodyLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "ToolRequest" WHERE "toolId" = 'tc-tool'`)
		db.Exec(`DELETE FROM "OnDemandTool" WHERE id = 'tc-tool'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'tc-%'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role, status) VALUES ('tc-t', 'tc فني', 'TECHNICIAN', 'ACTIVE'), ('tc-o', 'tc مالك', 'OWNER', 'ACTIVE'), ('tc-u', 'tc ثاني', 'TECHNICIAN', 'ACTIVE')`)
	db.MustExec(`INSERT INTO "OnDemandTool"(id, name, barcode, "totalQuantity", "availableQuantity") VALUES ('tc-tool', 'دريل tc', 'TC-BAR-1', 1, 1)`)
	c := NewToolCustody(db, NewInventoryService(repository.NewInventoryRepository(db)))
	now := time.Now()
	if _, err := c.Take("tc-t", "tc-t", "tc-tool", now.Add(time.Hour).Format(time.RFC3339), "", now); err == nil {
		t.Fatal("وقت بالمستقبل لازم ينرفض")
	}
	it, err := c.Take("tc-t", "tc-t", "tc-tool", now.Add(-2*time.Hour).Format(time.RFC3339), "للحجز B1", now)
	if err != nil {
		t.Fatal(err)
	}
	if !it.SelfCheckout || it.TakenAt == nil || now.Sub(*it.TakenAt) < 119*time.Minute {
		t.Fatalf("الأخذ غلط: %+v", it)
	}
	if _, err := c.Take("tc-o", "tc-u", "tc-tool", "", "", now); err == nil {
		t.Fatal("ماكو كمية — لازم ينرفض")
	}
	if err := c.Verify("tc-o", it.ID, true, ""); err == nil {
		t.Fatal("ما يتدقق قبل التسليم")
	}
	if err := c.RequestReturn("tc-u", it.ID, "p.jpg", ""); err != ErrDesignHidden {
		t.Fatalf("غير صاحب العهدة ما يسلّم: %v", err)
	}
	if err := c.RequestReturn("tc-t", it.ID, "", ""); err == nil {
		t.Fatal("التسليم بلا صورة لازم ينرفض")
	}
	if err := c.RequestReturn("tc-t", it.ID, "photo-1.jpg", "سلّمتها للمخزن"); err != nil {
		t.Fatal(err)
	}
	if err := c.Verify("tc-o", it.ID, false, ""); err == nil {
		t.Fatal("الرفض يحتاج سبب")
	}
	if err := c.Verify("tc-o", it.ID, false, "الصورة مو واضحة"); err != nil {
		t.Fatal(err)
	}
	mine, _ := c.Mine("tc-t")
	if len(mine) != 1 || mine[0].ReturnRequestedAt != nil || mine[0].VerifyNote == nil {
		t.Fatalf("بعد الرفض لازم ترجع لعهدته ويا ملاحظة: %+v", mine)
	}
	_ = c.RequestReturn("tc-t", it.ID, "photo-2.jpg", "")
	if err := c.Verify("tc-o", it.ID, true, ""); err != nil {
		t.Fatal(err)
	}
	mine, _ = c.Mine("tc-t")
	var avail int
	db.Get(&avail, `SELECT "availableQuantity" FROM "OnDemandTool" WHERE id = 'tc-tool'`)
	if len(mine) != 0 || avail != 1 {
		t.Fatalf("بعد التدقيق لازم ترجع للمخزن: عهدة %d، كمية %d", len(mine), avail)
	}
	// المدير ياخذ عن موظف
	it2, err := c.Take("tc-o", "tc-u", "tc-tool", "", "", now)
	if err != nil || it2.SelfCheckout || it2.EmployeeID != "tc-u" || it2.ByName == nil {
		t.Fatalf("المدير عن الموظف: %+v %v", it2, err)
	}
}
