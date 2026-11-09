package service

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// (11-07) دورة طلب الأداة والمخزن:
//   - الإرجاع چان يزيد المخزن ٢ بدل ١ (الريبو والخدمة كلهم يزيدون).
//   - رفض طلب موافَق عليه چان يمر، فالوحدة المنقوصة ما ترجع أبد.
func TestToolRequestStockLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewInventoryRepository(db)
	svc := NewInventoryService(repo)

	emp, boss := uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'فني عدة', $2, 'TECHNICIAN', 'ACTIVE'), ($3, 'إداري كميات', $4, 'ADMIN', 'ACTIVE')`, emp, "tr_"+sfx, boss, "tb_"+sfx)
	tool, err := repo.CreateOnDemandTool("مثقب اختبار "+sfx, "BC-"+sfx, 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "ToolRequest" WHERE "employeeId" = $1`, emp)               //nolint:errcheck
		db.Exec(`DELETE FROM "OnDemandTool" WHERE id = $1`, tool.ID)                    //nolint:errcheck
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" IN ($1, $2)`, emp, boss) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, emp, boss)               //nolint:errcheck
	})
	avail := func() int {
		var n int
		if err := db.Get(&n, `SELECT "availableQuantity" FROM "OnDemandTool" WHERE id = $1`, tool.ID); err != nil {
			t.Fatal(err)
		}
		return n
	}
	newReq := func() string {
		r, err := svc.CreateToolRequest(model.CreateToolRequestRequest{EmployeeID: emp, ToolID: tool.ID, Reason: model.ToolRequestReasonDamaged})
		if err != nil {
			t.Fatal(err)
		}
		return r.ID
	}

	// وحدتين طالعات (٣←١)، إرجاع وحدة ← ٢ بالضبط (مو ٣)، وإرجاع ثاني ينرفض.
	a, a2 := newReq(), newReq()
	for _, id := range []string{a, a2} {
		if _, err := svc.ApproveToolRequest(id, model.ApproveToolRequestRequest{ApprovedByID: boss}); err != nil {
			t.Fatal(err)
		}
	}
	if n := avail(); n != 1 {
		t.Fatalf("بعد موافقتين المتوفر %d، المفروض 1", n)
	}
	if _, err := svc.ReturnToolRequest(a); err != nil {
		t.Fatal(err)
	}
	if n := avail(); n != 2 {
		t.Fatalf("بعد إرجاع وحدة المتوفر %d، المفروض 2", n)
	}
	if _, err := svc.ReturnToolRequest(a2); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReturnToolRequest(a); err == nil {
		t.Fatal("إرجاع ثاني لنفس الطلب انقبل")
	}
	if n := avail(); n != 3 {
		t.Fatalf("الإرجاع الثاني غيّر المخزن: %d", n)
	}

	// رفض طلب موافَق عليه ينرفض، والوحدة تبقى قابلة للإرجاع.
	b := newReq()
	if _, err := svc.ApproveToolRequest(b, model.ApproveToolRequestRequest{ApprovedByID: boss}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RejectToolRequest(b, boss); err == nil {
		t.Fatal("رفض طلب موافَق عليه انقبل")
	}
	if _, err := svc.ReturnToolRequest(b); err != nil {
		t.Fatalf("الطلب الموافَق عليه ما ينرجع بعد محاولة رفض: %v", err)
	}
	if n := avail(); n != 3 {
		t.Fatalf("المخزن النهائي %d، المفروض 3", n)
	}

	// رفض طلب معلّق يشتغل عادي وما يلمس المخزن.
	c := newReq()
	if _, err := svc.RejectToolRequest(c, boss); err != nil {
		t.Fatal(err)
	}
	if n := avail(); n != 3 {
		t.Fatalf("رفض المعلّق غيّر المخزن: %d", n)
	}
}
