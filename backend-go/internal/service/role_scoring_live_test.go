package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// (ل) 11-12: استعلام أقسام التصاميم للتقييم، وعدّاد الخدمات للشهر الحالي.
func TestRoleScoringLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clean := func() {
		db.Exec(`DELETE FROM "DesignSection" WHERE "setId" = 'rs-set'`)
		db.Exec(`DELETE FROM "DesignSet" WHERE id = 'rs-set'`)
		db.Exec(`DELETE FROM "Booking" WHERE id LIKE 'rs-b%'`)
		db.Exec(`DELETE FROM "Service" WHERE id = 'rs-svc'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'rs-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'rs-des'`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('rs-des', 'مصممة', 'DESIGNER')`)
	db.MustExec(`INSERT INTO "DesignSet"(id, title) VALUES ('rs-set', 'فيلا')`)
	db.MustExec(`INSERT INTO "DesignSection"(id, "setId", discipline, status, "assigneeId", "dueDate", "submittedAt") VALUES
		('rs-s1', 'rs-set', 'معماري', 'REVIEW', 'rs-des', CURRENT_DATE - 3, now() - interval '5 days'),
		('rs-s2', 'rs-set', 'كهرباء', 'IN_PROGRESS', 'rs-des', CURRENT_DATE - 2, NULL)`)
	repo := repository.NewMatrixScoreRepository(db)
	rows, err := repo.DesignSectionsDue(time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, r := range rows {
		if r.AssigneeID == "rs-des" {
			p, _, ok := DesignSectionPoints(r.Status, r.DueEnd, r.SubmittedAt, time.Now())
			if ok {
				got[r.ID] = p
			}
		}
	}
	if got["rs-s1"] != 2 || got["rs-s2"] != 0 || len(got) != 2 {
		t.Fatalf("سلّم بموعده = ٢، فات وما سلّم = ٠: %+v", got)
	}

	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('rs-cu', 'زبون', '07700000009')`)
	db.MustExec(`INSERT INTO "Service"(id, name) VALUES ('rs-svc', 'خدمة اختبار الشهر')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "serviceId", "createdAt") VALUES
		('rs-b1', 'RS-1', 'rs-cu', 'PENDING', 'rs-svc', now()),
		('rs-b2', 'RS-2', 'rs-cu', 'PENDING', 'rs-svc', now() - interval '70 days')`)
	entries, err := repository.NewStatsRepository(db).ServiceBreakdown()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.ServiceID != nil && *e.ServiceID == "rs-svc" {
			if e.Count != 2 || e.MonthCount != 1 {
				t.Fatalf("الكلي ٢، هذا الشهر ١: %+v", e)
			}
			return
		}
	}
	t.Fatal("الخدمة ما طلعت")
}
