package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/middleware"
)

// TestRoutePlanLive: حجوزات الموظف بيومها بس، مرتبة، والي بلا موقع منفصلة، وغيره ياخذ 404.
func TestRoutePlanLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	var hadLoc bool
	_ = db.Get(&hadLoc, `SELECT EXISTS (SELECT 1 FROM "CompanyLocation" WHERE id = 1)`)
	if !hadLoc {
		db.MustExec(`INSERT INTO "CompanyLocation" (id, lat, lng, "radiusM") VALUES (1, 32.6, 44.0, 150)`)
		t.Cleanup(func() { db.Exec(`DELETE FROM "CompanyLocation" WHERE id = 1`) }) //nolint:errcheck
	}
	var co struct{ Lat, Lng float64 }
	_ = db.QueryRow(`SELECT lat, lng FROM "CompanyLocation" WHERE id = 1`).Scan(&co.Lat, &co.Lng)
	emp, other, cust := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'ليدر فحص المسار', $2, 'TECHNICIAN', 'ACTIVE'), ($3, 'فني ثاني', $4, 'TECHNICIAN', 'ACTIVE')`, emp, "rp_"+sfx, other, "rp2_"+sfx)
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون فحص المسار', '07700000004')`, cust)
	loc, _ := time.LoadLocation("Asia/Baghdad")
	day := time.Now().In(loc).AddDate(0, 0, 3)
	at := func(h int) time.Time { return time.Date(day.Year(), day.Month(), day.Day(), h, 0, 0, 0, loc) }
	ids := []string{}
	mk := func(code string, h int, lat *float64, mine bool) {
		id := uuid.NewString()
		ids = append(ids, id)
		var lng *float64
		if lat != nil {
			v := co.Lng
			lng = &v
		}
		db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status, "scheduledAt", "mapLatitude", "mapLongitude") VALUES ($1, $2, $3, 'CONFIRMED', $4, $5, $6)`, id, code, cust, at(h), lat, lng)
		who := other
		if mine {
			who = emp
		}
		db.MustExec(`INSERT INTO "BookingAssignment" (id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, $1, $2, 'TECH_1')`, id, who)
	}
	f := func(d float64) *float64 { v := co.Lat + d; return &v }
	mk("RPfar"+sfx, 9, f(0.3), true)
	mk("RPnear"+sfx, 11, f(0.01), true)
	mk("RPnoloc"+sfx, 12, nil, true)
	mk("RPother"+sfx, 10, f(0.05), false)
	t.Cleanup(func() {
		for _, id := range ids {
			db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" = $1`, id) //nolint:errcheck
			db.Exec(`DELETE FROM "Booking" WHERE id = $1`, id)                    //nolint:errcheck
		}
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)              //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, emp, other) //nolint:errcheck
	})
	call := func(me, q string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/route-plan?date="+day.Format("2006-01-02")+q, nil)
		ctx := context.WithValue(r.Context(), middleware.ContextEmployeeID, me)
		w := httptest.NewRecorder()
		NewRoutePlanHandler(db).Plan(w, r.WithContext(context.WithValue(ctx, middleware.ContextRole, "TECHNICIAN")))
		return w
	}
	w := call(emp, "")
	var out struct {
		Plan struct {
			Stops []struct{ Label string } `json:"stops"`
			KM    float64                  `json:"km"`
			Saved float64                  `json:"savedKm"`
		} `json:"plan"`
		NoLocation []map[string]string `json:"noLocation"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if len(out.Plan.Stops) != 2 || len(out.NoLocation) != 1 {
		t.Fatalf("لازم توقفين ووحد بلا موقع، وحجز الغير ما يطلع: %+v", out)
	}
	if out.Plan.Stops[0].Label[:6] != "RPnear" || out.Plan.Saved <= 0 {
		t.Fatalf("الأقرب أول: %+v", out.Plan)
	}
	if c := call(other, "&employeeId="+emp).Code; c != 404 {
		t.Fatalf("زميل يشوف مسار غيره: %d", c)
	}
}
