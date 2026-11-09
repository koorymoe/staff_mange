package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (ع) 11-08: ماتركس يقرا غسيل اليوم من التقييم اليومي.
func TestWashToday_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	clean := func() {
		db.Exec(`DELETE FROM "VehicleWashRating" WHERE "employeeId" = 'wt-emp'`)
		db.Exec(`DELETE FROM "VehicleDailyRating" WHERE "vehicleId" IN ('wt-v1','wt-v2')`)
		db.Exec(`DELETE FROM "Vehicle" WHERE id IN ('wt-v1','wt-v2')`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'wt-emp'`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name) VALUES ('wt-emp','غاسل تجربة')`)
	db.MustExec(`INSERT INTO "Vehicle"(id, name, "plateNumber", "isActive") VALUES ('wt-v1','سيارة نظيفة','WT-1',true), ('wt-v2','سيارة وسخة','WT-2',true)`)
	db.MustExec(`INSERT INTO "VehicleDailyRating"(id, "vehicleId", "ratedDate", wash) VALUES ('wt-r1','wt-v1',baghdad_today(),4), ('wt-r2','wt-v2',baghdad_today(),1)`)
	db.MustExec(`INSERT INTO "VehicleWashRating"(id, "dailyRatingId", "employeeId", score) VALUES ('wt-w1','wt-r1','wt-emp',2)`)
	w, err := NewMatrixScoreRepository(db).WashToday()
	if err != nil {
		t.Fatal(err)
	}
	if w.Rated < 2 || w.Washed < 1 || len(w.LowCars) < 1 || len(w.Techs) < 1 {
		t.Fatalf("الغسيل غلط: %+v", w)
	}
}
