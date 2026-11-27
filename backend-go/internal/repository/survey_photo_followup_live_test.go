package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// بلاغ #62: صور المتابعة تنحفظ ويا المرحلة وما تختلط ويا صور الكشف.
func TestSurveyPhotoFollowupLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	var emp string
	proj := "fu-test-proj"
	db.Exec(`DELETE FROM "Project" WHERE id = $1`, proj)
	db.MustExec(`INSERT INTO "Project" (id, code, name) VALUES ($1, 'FU-TEST', 'مشروع صور')`, proj)
	defer db.Exec(`DELETE FROM "Project" WHERE id = $1`, proj)
	if err := db.Get(&emp, `SELECT id FROM "Employee" LIMIT 1`); err != nil {
		t.Skip("ماكو موظف")
	}
	defer db.Exec(`DELETE FROM "SurveyPhoto" WHERE "fileKey" LIKE 'fu-test-%'`)
	r := NewSurveyPhotoRepository(db)
	if _, err := r.CreateKind(nil, &proj, "fu-test-1", "image/jpeg", "a.jpg", emp, "FOLLOWUP", "تنفيذ"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Create(nil, &proj, "fu-test-2", "image/jpeg", "b.jpg", emp); err != nil {
		t.Fatal(err)
	}
	fu, _ := r.ListForProjectKind(proj, "FOLLOWUP")
	sv, _ := r.ListForProject(proj)
	has := func(l []SurveyPhoto, name string) (bool, *string) {
		for _, p := range l {
			if p.FileName == name {
				return true, p.Stage
			}
		}
		return false, nil
	}
	if ok, st := has(fu, "a.jpg"); !ok || st == nil || *st != "تنفيذ" {
		t.Fatalf("المتابعة: %v %v", ok, st)
	}
	if ok, _ := has(fu, "b.jpg"); ok {
		t.Fatal("صورة الكشف طلعت بالمتابعة")
	}
	if ok, _ := has(sv, "a.jpg"); ok {
		t.Fatal("صورة المتابعة طلعت بالكشف")
	}
}
