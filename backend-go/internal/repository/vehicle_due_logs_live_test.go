package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// بلاغ #46 (11-16): تنبيه الصيانة ياخذ آخر سجل *فعلي* من كل نوع، والوقود ما يطلّع تنبيه صيانة.
func TestDueSoonLogs_LatestAndNoFuel_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const v = "due-logs-veh"
	cleanup := func() {
		db.Exec(`DELETE FROM "VehicleLog" WHERE "vehicleId" = $1`, v)
		db.Exec(`DELETE FROM "Vehicle" WHERE id = $1`, v)
	}
	cleanup()
	t.Cleanup(cleanup)
	db.MustExec(`INSERT INTO "Vehicle"(id, name, "plateNumber") VALUES ($1, 'سيارة اختبار', 'T-1')`, v)
	ins := `INSERT INTO "VehicleLog"(id, "vehicleId", type, "performedAt", "nextDueOdometer") VALUES ($1, $2, $3, now() - $4::interval, $5)`
	db.MustExec(ins, "dl-1", v, "FUEL", "10 days", 1000)       // وقود قديم بي عداد استحقاق
	db.MustExec(ins, "dl-2", v, "OIL_CHANGE", "10 days", 5000) // دهن قديم بي استحقاق
	db.MustExec(ins, "dl-3", v, "OIL_CHANGE", "1 day", nil)    // دهن أحدث بلا استحقاق

	rows, err := NewVehicleRepository(db).DueSoonLogs()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.VehicleID == v {
			t.Fatalf("ما لازم يطلع تنبيه: %+v", r)
		}
	}
}
