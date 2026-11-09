package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (ع) 11-08: صيت التصاميم — المسار الكامل والصلاحيات.
func TestDesignSet_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	const adm, des, other = "__ds_admin", "__ds_designer", "__ds_other"
	emps := []string{adm, des, other}
	clean := func() {
		db.Exec(`DELETE FROM "DesignSet" WHERE "createdById" = ANY($1)`, emps)   //nolint:errcheck
		db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" = ANY($1)`, emps)  //nolint:errcheck
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" = ANY($1)`, emps) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, emps)               //nolint:errcheck
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'مدير نظام','ACTIVE','ADMIN'),($2,'مصمم','ACTIVE','TECHNICIAN'),($3,'غريب','ACTIVE','TECHNICIAN')`, adm, des, other)
	s := NewDesignSetService(db, repository.NewExtraTaskRepository(db), repository.NewNotificationRepository(db))

	var tools int
	db.Get(&tools, `SELECT count(*) FROM "DesignTool"`) //nolint:errcheck
	if tools < 10 {
		t.Fatalf("أدوات البداية: %d", tools)
	}
	if _, err := s.Create(other, "TECHNICIAN", nil, DesignSetInput{Title: "ممنوع", Disciplines: []string{"ARCH"}}); err == nil {
		t.Fatal("الموظف العادي ما ينشئ صيت")
	}
	set, err := s.Create(adm, "ADMIN", nil, DesignSetInput{Title: "فيلا تجربة", Disciplines: []string{"ARCH", "CAMERAS", "SOLAR", "BOGUS"}})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Detail(adm, "ADMIN", nil, set)
	if err != nil || len(d.Sections) != 3 || d.Sections[0].Discipline != "ARCH" {
		t.Fatalf("التخصصات: %v %+v", err, d)
	}
	cam := d.Sections[1].ID
	if _, err := s.Detail(other, "TECHNICIAN", nil, set); err == nil {
		t.Fatal("الغريب ما يشوف الصيت")
	}
	due := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
	if err := s.Assign(adm, "ADMIN", nil, cam, des, due); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Detail(des, "TECHNICIAN", nil, set); err != nil {
		t.Fatal("المصمم المكلّف لازم يشوف صيته")
	}
	if err := s.Submit(des, "TECHNICIAN", nil, cam); err == nil {
		t.Fatal("ما يرفع للمراجعة بلا ملف")
	}
	if _, err := s.AddFile(des, "TECHNICIAN", nil, cam, DesignFileInput{URL: "http://x"}); err == nil {
		t.Fatal("الرابط لازم https")
	}
	if _, err := s.AddFile(des, "TECHNICIAN", nil, cam, DesignFileInput{FileKey: "misc/x.png"}); err == nil {
		t.Fatal("الملف لازم من مجلد التصاميم")
	}
	if _, err := s.AddFile(des, "TECHNICIAN", nil, cam, DesignFileInput{URL: "https://design.example/p/1", ToolName: "Dahua"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(des, "TECHNICIAN", nil, cam); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(des, "TECHNICIAN", cam, true, ""); err == nil {
		t.Fatal("المصمم ما يعتمد شغله")
	}
	if err := s.Review(adm, "ADMIN", cam, false, "قص"); err == nil {
		t.Fatal("الإرجاع لازم بملاحظة")
	}
	if err := s.Review(adm, "ADMIN", cam, true, ""); err != nil {
		t.Fatal(err)
	}
	var taskStatus string
	db.Get(&taskStatus, `SELECT t.status FROM "ExtraTask" t JOIN "DesignSection" x ON x."taskId" = t.id WHERE x.id = $1`, cam) //nolint:errcheck
	if taskStatus != "DONE" {
		t.Fatalf("الاعتماد لازم يخلّص مهمة المصمم: %q", taskStatus)
	}
	if _, err := s.AddFile(des, "TECHNICIAN", nil, cam, DesignFileInput{URL: "https://x.example"}); err == nil {
		t.Fatal("ما ينضاف على تخصص معتمد")
	}
	// يرجّعه بعد الاعتماد
	if err := s.Review(adm, "ADMIN", cam, false, "عدّل زاوية الكاميرا الخلفية"); err != nil {
		t.Fatal(err)
	}
	list, _ := s.List(des, "TECHNICIAN", nil)
	if len(list) != 1 || list[0].Sections != 3 {
		t.Fatalf("قائمة المصمم: %+v", list)
	}
}
