package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// (ك 11-23) تقرير اتصال التقني: بعد ٢٤ ساعة بلا تقرير نقطة وحدة بلا فلوس، مرة وحدة،
// وبلا أثر رجعي على الي انحوّل قبل القاعدة؛ و«الزبون ما رد» ينحسب تقرير.
func TestTechContactWatchLive(t *testing.T) {
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
		db.Exec(`DELETE FROM "KpiEvaluation" WHERE "employeeId" LIKE 'tcw-%'`)
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" LIKE 'tcw-%'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'tcw-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'tcw-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id LIKE 'tcw-%'`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('tcw-t', 'tcw تقني', 'TECHNICAL'), ('tcw-o', 'tcw مدير', 'ADMIN'), ('tcw-m', 'tcw مراقب', 'MONITOR')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('tcw-cu', 'x', '07700003434')`)
	// القاعدة «انطبقت» قبل ساعتين يوم حتى نكدر نجرّب حجز عابر ٢٤ ساعة بعدها
	var applied string
	db.Get(&applied, `SELECT "appliedAt"::text FROM "SchemaMigration" WHERE version = '0522_tech_contact_report'`)
	db.MustExec(`UPDATE "SchemaMigration" SET "appliedAt" = now() - interval '48 hours' WHERE version = '0522_tech_contact_report'`)
	t.Cleanup(func() {
		db.Exec(`UPDATE "SchemaMigration" SET "appliedAt" = $1::timestamp WHERE version = '0522_tech_contact_report'`, applied)
	})
	ins := func(id, ago string) {
		db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "handoverToId", "handoverById", "handoverAt", "handoverReason", "confirmationContactedAt")
			VALUES ($1, upper($1), 'tcw-cu', 'CONFIRMED', 'tcw-t', 'tcw-o', now() - $2::interval, 'تجربة', now())`, id, ago)
	}
	ins("tcw-late", "25 hours") // عابر ٢٤ ساعة بعد القاعدة ← نقطة
	ins("tcw-old", "60 hours")  // قبل القاعدة ← ولا شي
	ins("tcw-new", "5 hours")   // جديد ← تذكير بس
	ins("tcw-na", "26 hours")   // التقني گال «الزبون ما رد» ← تقرير، بلا خصم

	notif := repository.NewNotificationRepository(db)
	s := NewBookingService(repository.NewBookingRepository(db), repository.NewEmployeeRepository(db), repository.NewCustomerRepository(db),
		repository.NewQualityFollowUpRepository(db), notif, repository.NewInventoryRepository(db))
	if _, err := s.TechWaiting("tcw-na", "tcw-t", "اتصلت ثلاث مرات وما رد", "NO_ANSWER"); err != nil {
		t.Fatal(err)
	}

	w := NewTechContactWatch(db, notif)
	w.remind()
	w.deduct()
	w.deduct() // مرة ثانية: ما تنعاد

	type ev struct {
		Points int     `db:"points"`
		Reason string  `db:"reason"`
		Amount float64 `db:"deductionAmount"`
	}
	var evs []ev
	db.Select(&evs, `SELECT points, reason, "deductionAmount" FROM "KpiEvaluation" WHERE "employeeId" = 'tcw-t'`)
	if len(evs) != 1 {
		t.Fatalf("لازم خصم واحد بس (TCW-LATE)، طلع %d: %+v", len(evs), evs)
	}
	if evs[0].Points != -1 || evs[0].Amount != 0 {
		t.Fatalf("الخصم لازم نقطة وحدة بلا فلوس: %+v", evs[0])
	}
	for _, want := range []string{"تنفيذ المهام", "عدم إكمال المهام الموجهة إليه", "TCW-LATE"} {
		if !strings.Contains(evs[0].Reason, want) {
			t.Fatalf("السبب ناقص %q: %s", want, evs[0].Reason)
		}
	}
	var deductNotes, reminders, watcher int
	db.Get(&deductNotes, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'tcw-t' AND type = 'kpi_deduction'`)
	db.Get(&reminders, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'tcw-t' AND type = 'tech_contact_reminder' AND strpos(message, 'TCW-NEW') > 0`)
	db.Get(&watcher, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'tcw-m' AND type IN ('tech_contact_kpi', 'tech_contact_report')`)
	if deductNotes != 1 {
		t.Fatalf("إشعار الخصم لازم مرة وحدة، طلع %d", deductNotes)
	}
	if reminders != 1 {
		t.Fatalf("لازم تذكير واحد للحجز الجديد، طلع %d", reminders)
	}
	if watcher != 2 { // تقرير «ما رد» + الخصم
		t.Fatalf("المراقب لازم يوصله التقرير والخصم، طلع %d", watcher)
	}
	w.remind()
	db.Get(&reminders, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'tcw-t' AND type = 'tech_contact_reminder'`)
	if reminders != 1 {
		t.Fatalf("التذكير ما يتكرر قبل ٤ ساعات، طلع %d", reminders)
	}
}
