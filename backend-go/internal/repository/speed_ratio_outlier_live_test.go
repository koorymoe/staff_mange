package repository

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (11-10 بلاغ #27) سرعة العمل طلعت 666 مرة أسرع: معدّل النسب ويا عيّنة بدقيقة وحدة.
// هسه: نسبة المجموع، والعيّنة تحت ٥ دقايق ما تنحسب.
func TestEmployeeSpeedRatioIgnoresOutlierLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	repo := NewJobDurationSampleRepository(db)
	emp, cust := uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	sys := "منظومة سرعة " + sfx
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'فني سرعة', $2, 'TECHNICIAN', 'ACTIVE')`, emp, "sp_"+sfx)
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ($1, 'زبون سرعة', $2)`, cust, "0775"+sfx)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "JobDurationSample" WHERE "systemName" = $1`, sys) //nolint:errcheck
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "employeeId" = $1`, emp) //nolint:errcheck
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = $1`, cust)          //nolint:errcheck
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)                   //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)                    //nolint:errcheck
	})
	add := func(minutes int, mine bool) {
		b := uuid.NewString()
		db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", status) VALUES ($1, $2, $3, 'COMPLETED')`, b, "SP"+uuid.NewString()[:8], cust)
		if mine {
			db.MustExec(`INSERT INTO "BookingAssignment" (id, "bookingId", "employeeId", role) VALUES (gen_random_uuid()::text, $1, $2, 'TECH_1')`, b, emp)
		}
		db.MustExec(`INSERT INTO "JobDurationSample" (id, "systemName", "jobType", "itemCount", "crewSize", "durationMinutes", "bookingId") VALUES ($1, $2, 'INSTALL', 1, 1, $3, $4)`,
			uuid.NewString(), sys, minutes, b)
	}
	// الكل ~١٢٠ دقيقة؛ الفني: شغلة ١٢٠ (عادي) + شغلة بدقيقة وحدة (ضغط ورا بعض)
	add(120, false)
	add(120, false)
	add(120, true)
	add(1, true)
	ratio, n, err := repo.EmployeeSpeedRatio(emp, time.Now().In(time.FixedZone("BGW", 3*3600)).Format("2006-01"))
	if err != nil {
		t.Fatal(err)
	}
	if ratio == nil || *ratio > 2 || *ratio < 0.5 {
		t.Fatalf("النسبة %v من %d عيّنة — المفروض قريبة من ١", ratio, n)
	}
}
