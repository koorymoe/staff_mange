package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
)

// (دازاي 11-15) «لازم كلشي يتسجل»: الحذف ينحفظ بنسخة كاملة؛ والحذف الي ما صار ما يبقى بالسجل.
// ويا: الدور الجاي من المنصب ينشال بـPermissionRevoke "role:X".
func TestDeleteLogAndOrgRoleRevokeLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "DeleteLog" WHERE "rowId" LIKE 'dl-%'`)
		db.Exec(`DELETE FROM "Announcement" WHERE id LIKE 'dl-%'`)
		db.Exec(`DELETE FROM "PermissionRevoke" WHERE "employeeId" = 'dl-emp'`)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = 'dl-emp'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'dl-emp'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('dl-emp', 'dl', 'DESIGNER')`)
	db.MustExec(`INSERT INTO "Announcement"(id, body, active, "createdById") VALUES ('dl-a1', 'x', true, 'dl-emp'), ('dl-a2', 'y', true, 'dl-emp')`)

	mux := http.NewServeMux()
	mw := DeleteLogged(db, "Announcement", "id")
	mux.Handle("DELETE /del/{id}", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db.MustExec(`DELETE FROM "Announcement" WHERE id = $1`, r.PathValue("id"))
	})))
	mux.Handle("DELETE /fail/{id}", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})))
	for _, p := range []string{"/del/dl-a1", "/fail/dl-a2"} {
		mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodDelete, p, nil))
	}
	var body string
	if err := db.Get(&body, `SELECT data->>'body' FROM "DeleteLog" WHERE "tableName" = 'Announcement' AND "rowId" = 'dl-a1'`); err != nil || body != "x" {
		t.Fatalf("الحذف لازم ينسجل بنسخة كاملة: %v %q", err, body)
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "DeleteLog" WHERE "rowId" = 'dl-a2'`)
	if n != 0 {
		t.Fatal("حذف ما صار ما يبقى بالسجل")
	}

	// الدور من المنصب
	var div string
	if err := db.Get(&div, `SELECT "divisionId" FROM "OrgUnit" WHERE id = 'admin.monitoring'`); err != nil {
		t.Skip("ماكو وحدة المراقبة بالهيكل")
	}
	db.MustExec(`INSERT INTO "OrgAssignment"(id, "employeeId", "divisionId", "unitId", position) VALUES ('dl-oa', 'dl-emp', $1, 'admin.monitoring', 'MEMBER')`, div)
	roles := func() []string {
		r := []string{}
		db.Select(&r, model.OrgRolesOfEmployeeSQL(), "dl-emp")
		return r
	}
	if r := roles(); len(r) != 1 || r[0] != "MONITOR" {
		t.Fatalf("المنصب لازم ينطي مراقب: %v", r)
	}
	db.MustExec(`INSERT INTO "PermissionRevoke"("employeeId", permission) VALUES ('dl-emp', $1)`, model.RoleRevokeKey("MONITOR"))
	if r := roles(); len(r) != 0 {
		t.Fatalf("الدور المشال رجع: %v", r)
	}
	var eff string
	db.Get(&eff, `SELECT `+model.EffectiveRoleSQL(`'dl-emp'`, `'DESIGNER'`))
	if eff != "DESIGNER" {
		t.Fatalf("الدور الفعلي = %s", eff)
	}
}
