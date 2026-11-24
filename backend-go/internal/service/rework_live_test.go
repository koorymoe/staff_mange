package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/timeutil"
)

// (ل) 11-24 دازاي: كادر الصبح غلط، كادر المسا صلّح ٣ ساعات — المراقب يوافق ← ٣ × سعر ساعة الي صلّح
// ينضاف إله وينخصم من الصبح، ويطلع بـ«الفلوس الي دخّلها».
func TestReworkLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "ReworkClaim" WHERE "createdById" LIKE 'rw-%'`)
		db.Exec(`DELETE FROM "CostEmployee" WHERE "employeeId" LIKE 'rw-%'`)
		db.Exec(`DELETE FROM "BookingAssignment" WHERE "bookingId" LIKE 'rw-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'rw-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'rw-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'rw-%'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role, status, "isLeader") VALUES ('rw-am', 'rw صبح', 'TECHNICIAN', 'ACTIVE', true),
		('rw-pm', 'rw مسا', 'TECHNICIAN', 'ACTIVE', true), ('rw-mon', 'rw مراقب', 'MONITOR', 'ACTIVE', false), ('rw-x', 'rw غريب', 'TECHNICIAN', 'ACTIVE', false)`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('rw-cu', 'x', '07700007777')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "completedAt", "arrivedAt") VALUES
		('rw-orig', 'RW-ORIG', 'rw-cu', 'COMPLETED', now() - interval '1 day', now() - interval '26 hours'),
		('rw-fix', 'RW-FIX', 'rw-cu', 'COMPLETED', now() - interval '1 hour', now() - interval '4 hours')`)
	db.MustExec(`INSERT INTO "BookingAssignment"(id, "bookingId", "employeeId", role) VALUES ('rw-a1', 'rw-orig', 'rw-am', 'TECH_1'), ('rw-a2', 'rw-fix', 'rw-pm', 'TECH_1')`)
	c := NewCompanyCosts(db)
	// سعر ساعة الي صلّح: 624,000 ÷ 26 يوم ÷ 8 = 3,000 (بلا مصاريف ولا ربح لأن الوحيدين)
	if err := c.SaveEmployee(CostEmployee{EmployeeID: "rw-pm", Kind: "EARNER", Salary: 624000, HoursPerDay: 8, DaysPerMonth: 26}); err != nil {
		t.Fatal(err)
	}
	rw := NewRework(db)
	if err := rw.Submit("rw-x", "RW-FIX", "RW-ORIG", "الكاميرا ما چانت مربوطة"); err == nil {
		t.Fatal("غير الكادر ما يسجّل")
	}
	if err := rw.Submit("rw-pm", "RW-FIX", "RW-ORIG", "الكاميرا ما چانت مربوطة"); err != nil {
		t.Fatal(err)
	}
	items, _ := rw.List("rw-pm")
	if len(items) != 1 || items[0].Status != "PENDING" {
		t.Fatalf("لازم يطلع معلّق: %+v", items)
	}
	if err := rw.Decide("rw-pm", items[0].ID, true, ""); err != ErrDesignHidden {
		t.Fatalf("الفني ما يوافق: %v", err)
	}
	if err := rw.Decide("rw-mon", items[0].ID, true, "تمام"); err != nil {
		t.Fatal(err)
	}
	if err := rw.Decide("rw-mon", items[0].ID, true, ""); err == nil {
		t.Fatal("ما تنعاد الموافقة")
	}
	items, _ = rw.List("rw-mon")
	// الساعات من الحجز: الوصول قبل ٤ ساعات والإنهاء قبل ساعة = ٣
	var got *ReworkClaim
	for i := range items {
		if items[i].FixCode == "RW-FIX" {
			got = &items[i]
		}
	}
	if got == nil || got.Hours == nil || *got.Hours < 2.99 || *got.Hours > 3.01 {
		t.Fatalf("الساعات غلط: %+v", got)
	}
	ex := *got.Hours * 3000
	if got.Amount == nil || *got.Amount < ex-2 || *got.Amount > ex+2 {
		t.Fatalf("المبلغ لازم ≈ %v: %+v", ex, got)
	}
	from, to, _ := timeutil.MonthStart(timeutil.ThisMonth())
	rev, _ := repository.NewMatrixScoreRepository(db).Revenue(from.UTC(), to.Add(time.Hour).UTC(), "")
	sum := map[string]float64{}
	for _, x := range rev {
		if x.Kind == "WORK" {
			sum[x.EmployeeID] += x.Amount
		}
	}
	if sum["rw-pm"] < ex-2 || sum["rw-am"] > -ex+2 {
		t.Fatalf("الانتقال ما طلع بالفلوس: %v", sum)
	}
}
