package service

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (ل) 11-12 دازاي: فورما المخطط — المصممة تختار المكان وتعدّل الاسم والملاحظات والسكيل،
// الفورما تنقل ويا النسخ، والغريب ما يعدّل (404).
func TestPlanSheetLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	set, dst, own, tech := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'مالك', $2, 'OWNER', 'ACTIVE'), ($3, 'فني', $4, 'TECHNICIAN', 'ACTIVE')`, own, "sh_"+own[:6], tech, "sh_"+tech[:6])
	t.Cleanup(func() { db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, own, tech) }) //nolint:errcheck
	db.MustExec(`INSERT INTO "PlanSet" (id, title) VALUES ($1, 'سيت فورما'), ($2, 'سيت ثاني')`, set, dst)
	t.Cleanup(func() { db.Exec(`DELETE FROM "PlanSet" WHERE id IN ($1, $2)`, set, dst) }) //nolint:errcheck
	var f string
	db.Get(&f, `INSERT INTO "PlanFile" ("setId","typeCode",title,"fileKey","uploadedById") VALUES ($1,'7.1','أبواب','/api/files/designs/x1',$2) RETURNING id`, set, own) //nolint:errcheck
	s := NewPlanSetService(db, NewDesignSetService(db, nil, nil))
	d, _ := s.Detail(own, "OWNER", nil, set)
	if len(d.Items) != 1 || d.Items[0].SheetPos != "right" || d.Items[0].SheetName != nil {
		t.Fatalf("الافتراضي: يمين وفاضي: %+v", d.Items)
	}
	if err := s.UpdateSheet(tech, "TECHNICIAN", nil, f, PlanSheetInput{Pos: "bottom"}); err != ErrDesignHidden {
		t.Fatalf("الفني ما يعدّل: %v", err)
	}
	if err := s.UpdateSheet(own, "OWNER", nil, f, PlanSheetInput{Pos: "bottom", Name: " مخطط الطابق الأرضي: أبواب وشبابيك ", Notes: "القياسات بالمتر", Scale: "1:100"}); err != nil {
		t.Fatal(err)
	}
	d, _ = s.Detail(own, "OWNER", nil, set)
	it := d.Items[0]
	if it.SheetPos != "bottom" || it.SheetName == nil || *it.SheetName != "مخطط الطابق الأرضي: أبواب وشبابيك" || it.SheetScale == nil || *it.SheetScale != "1:100" {
		t.Fatalf("الحفظ: %+v", it)
	}
	if err := s.UpdateSheet(own, "OWNER", nil, f, PlanSheetInput{Pos: "anything"}); err != nil {
		t.Fatal(err)
	}
	d, _ = s.Detail(own, "OWNER", nil, set)
	if d.Items[0].SheetPos != "right" || d.Items[0].SheetName != nil {
		t.Fatalf("مكان غلط = يمين، والفاضي ينمسح: %+v", d.Items[0])
	}
	s.UpdateSheet(own, "OWNER", nil, f, PlanSheetInput{Pos: "bottom", Scale: "1:50"}) //nolint:errcheck
	if n, err := s.Copy(own, "OWNER", nil, dst, PlanCopyInput{FromSetID: set, FileIDs: []string{f}}); err != nil || n != 1 {
		t.Fatalf("النسخ: %d %v", n, err)
	}
	d, _ = s.Detail(own, "OWNER", nil, dst)
	if d.Items[0].SheetPos != "bottom" || d.Items[0].SheetScale == nil || *d.Items[0].SheetScale != "1:50" {
		t.Fatalf("الفورما تنقل ويا النسخ: %+v", d.Items[0])
	}
}
