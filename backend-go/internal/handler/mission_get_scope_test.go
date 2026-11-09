package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// TestMissionGetScope (فحص الصلاحيات 11-06): المهمة يشوفها ليدرها وأعضاؤها والمراقبة بس؛
// غيرهم ياخذ 404 (مو 403) حتى ما يعرف المهمة موجودة.
func TestMissionGetScope(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	cust, book, mis := uuid.NewString(), uuid.NewString(), uuid.NewString()
	leader, member, other := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	for i, e := range []string{leader, member, other} {
		db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'فني فحص المهام', $2, 'TECHNICIAN', 'ACTIVE')`, e, "msc_"+sfx+string(rune('a'+i)))
	}
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون فحص المهام', '07700000003')`, cust)
	db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "scheduledAt") VALUES ($1, $2, $3, 'CONFIRMED', now())`, book, "MS"+sfx, cust)
	db.MustExec(`INSERT INTO "Mission" (id, code, "bookingId", "leaderId", "memberIds") VALUES ($1, $2, $3, $4, $5)`, mis, "M"+sfx, book, leader, "{"+member+"}")
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "Mission" WHERE id = $1`, mis)                                       //nolint:errcheck
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, book)                                      //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)                                     //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, "{"+leader+","+member+","+other+"}") //nolint:errcheck
	})
	h := NewMissionHandler(service.NewMissionService(repository.NewMissionRepository(db, repository.NewBookingRepository(db))))
	h.Supervisor = func(r *http.Request) bool { return middleware.RoleFromContext(r) == "MONITOR" }
	get := func(emp, role string) int {
		r := httptest.NewRequest(http.MethodGet, "/api/missions/"+mis, nil)
		r.SetPathValue("id", mis)
		ctx := context.WithValue(r.Context(), middleware.ContextEmployeeID, emp)
		ctx = context.WithValue(ctx, middleware.ContextRole, role)
		w := httptest.NewRecorder()
		h.Get(w, r.WithContext(ctx))
		return w.Code
	}
	for name, c := range map[string]struct {
		emp, role string
		want      int
	}{
		"ليدر": {leader, "TECHNICIAN", 200}, "عضو": {member, "TECHNICIAN", 200},
		"غريب": {other, "TECHNICIAN", 404}, "مراقب": {other, "MONITOR", 200},
	} {
		if got := get(c.emp, c.role); got != c.want {
			t.Errorf("%s: %d، المتوقع %d", name, got, c.want)
		}
	}
}
