package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/middleware"
)

// (11-28) ملفي الشخصي: معلوماتي وحضوري وخصوماتي — عن نفسي بس.
func TestMeProfileLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	db.Exec(`DELETE FROM "KpiEvaluation" WHERE "employeeId" = 'mp-e'`)
	db.Exec(`DELETE FROM "Attendance" WHERE "employeeId" = 'mp-e'`)
	db.Exec(`DELETE FROM "Employee" WHERE id IN ('mp-e','mp-b')`)
	db.MustExec(`INSERT INTO "Employee"(id, name, username, password, role, status, phone, salary) VALUES
		('mp-e','mp موظف','mp-e','x','TECHNICIAN','ACTIVE','0770',500000), ('mp-b','mp مدير','mp-b','x','ADMIN','ACTIVE',NULL,NULL)`)
	defer db.Exec(`DELETE FROM "Employee" WHERE id IN ('mp-e','mp-b')`)
	defer db.Exec(`DELETE FROM "Attendance" WHERE "employeeId" = 'mp-e'`)
	defer db.Exec(`DELETE FROM "KpiEvaluation" WHERE "employeeId" = 'mp-e'`)
	db.MustExec(`INSERT INTO "Attendance"(id, "employeeId", "checkIn", "checkOut", date) VALUES ('mp-a','mp-e', now() - interval '9 hours', now() - interval '1 hour', now())`)
	db.MustExec(`INSERT INTO "KpiEvaluation"(id, "employeeId", "evaluatorId", points, reason, "deductionAmount", "createdAt") VALUES ('mp-k','mp-e','mp-b',-2,'تأخير',10000, now())`)

	r := httptest.NewRequest("GET", "/api/me/profile", nil)
	r = r.WithContext(context.WithValue(r.Context(), middleware.ContextEmployeeID, "mp-e"))
	w := httptest.NewRecorder()
	NewMeProfileHandler(db).Get(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	var out struct {
		Info struct {
			Name   string   `json:"name"`
			Salary *float64 `json:"salary"`
		} `json:"info"`
		Attendance []struct {
			Minutes *int `json:"minutes"`
		} `json:"attendance"`
		Deductions []struct {
			EmployeeID string  `json:"employeeId"`
			ByName     *string `json:"byName"`
		} `json:"deductions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Info.Name != "mp موظف" || out.Info.Salary == nil || *out.Info.Salary != 500000 {
		t.Fatalf("المعلومات: %+v", out.Info)
	}
	if len(out.Attendance) != 1 || out.Attendance[0].Minutes == nil || *out.Attendance[0].Minutes != 480 {
		t.Fatalf("الحضور: %+v", out.Attendance)
	}
	found := false
	for _, d := range out.Deductions {
		if d.EmployeeID != "mp-e" {
			t.Fatal("طلع خصم لموظف ثاني")
		}
		found = found || (d.ByName != nil && *d.ByName == "mp مدير")
	}
	if !found {
		t.Fatalf("الخصم ومنو خصمه ما طلع: %s", w.Body.String())
	}
}
