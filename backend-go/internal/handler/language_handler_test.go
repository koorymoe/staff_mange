package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// بلاغ #11: الموظف يغيّر لغته هو بس، وبس ar أو ur.
func TestSetMyLanguageLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	me, other := uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'فني لغة', $2, 'TECHNICIAN', 'ACTIVE'), ($3, 'زميل', $4, 'TECHNICIAN', 'ACTIVE')`, me, "lg_"+sfx, other, "lg2_"+sfx)
	t.Cleanup(func() { db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, me, other) }) //nolint:errcheck
	call := func(body string) int {
		r := httptest.NewRequest(http.MethodPut, "/api/me/language", strings.NewReader(body))
		r = r.WithContext(context.WithValue(r.Context(), middleware.ContextEmployeeID, me))
		w := httptest.NewRecorder()
		SetMyLanguage(db)(w, r)
		return w.Code
	}
	if c := call(`{"language":"fr"}`); c != 400 {
		t.Fatalf("لغة غلط: %d", c)
	}
	if c := call(`{"language":"ur"}`); c != 200 {
		t.Fatalf("%d", c)
	}
	var mine, theirs string
	_ = db.Get(&mine, `SELECT language FROM "Employee" WHERE id = $1`, me)
	_ = db.Get(&theirs, `SELECT language FROM "Employee" WHERE id = $1`, other)
	if mine != "ur" || theirs != "ar" {
		t.Fatalf("mine=%s theirs=%s", mine, theirs)
	}

	// الإدارة: تعديل حقل ثاني ما يمسح اللغة، وتگدر تحدد اللغة، وقيمة غلط تنرفض.
	svc := service.NewEmployeeService(repository.NewEmployeeRepository(db))
	ph := "0770"
	if _, err := svc.Update(me, model.UpdateEmployeeRequest{Phone: &ph}); err != nil {
		t.Fatal(err)
	}
	_ = db.Get(&mine, `SELECT language FROM "Employee" WHERE id = $1`, me)
	if mine != "ur" {
		t.Fatalf("تعديل الهاتف مسح اللغة: %s", mine)
	}
	bad, ar := "xx", "ar"
	if _, err := svc.Update(me, model.UpdateEmployeeRequest{Language: &bad}); err == nil {
		t.Fatal("لغة غلط انقبلت")
	}
	if _, err := svc.Update(me, model.UpdateEmployeeRequest{Language: &ar}); err != nil {
		t.Fatal(err)
	}
	_ = db.Get(&mine, `SELECT language FROM "Employee" WHERE id = $1`, me)
	if mine != "ar" {
		t.Fatalf("الإدارة ما غيّرت اللغة: %s", mine)
	}
}
