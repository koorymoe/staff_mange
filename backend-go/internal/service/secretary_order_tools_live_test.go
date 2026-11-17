package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// السكرتير ينفّذ عن المدير (دازاي 11-16): يجهّز تكليف الكادر ببطاقة، وما ينفّذه بنفس الدور —
// بس بعد موافقة المدير برسالة بعدها.
func TestSecretaryOrders_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	clean := func() {
		db.Exec(`DELETE FROM "MatrixOrder" WHERE "actorId" = 'so-boss'`)
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" IN ('so-lead', 'so-boss')`)
		// اختبارات ثانية تمسح كل الحجوزات وتولّدلها مهام ميدان — تنشال أول، وإلا الحجز ما ينحذف ويبقى ماسك الموظفين.
		db.Exec(`DELETE FROM "Mission" WHERE "bookingId" IN (SELECT id FROM "Booking" WHERE "customerId" = 'so-cu')`)
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = 'so-cu'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'so-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ('so-lead', 'so-boss')`)
	}
	clean()
	t.Cleanup(func() { clean(); db.Close() })
	tm := time.Now().In(baghdad).AddDate(0, 0, 2)
	at := time.Date(tm.Year(), tm.Month(), tm.Day(), 11, 0, 0, 0, baghdad).UTC().Format("2006-01-02 15:04:05")
	db.MustExec(`INSERT INTO "Employee"(id, name, role, "isLeader", status) VALUES ('so-boss', 'so boss', 'OWNER', false, 'ACTIVE'),
		('so-lead', 'سجاد سكرتير تجربة', 'TECHNICIAN', true, 'ACTIVE')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('so-cu', 'زبون سكرتير', '07700009933')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "scheduledAt") VALUES ('so-b1', 'SOB1', 'so-cu', 'CONFIRMED', $1)`, at)

	notif := repository.NewNotificationRepository(db)
	orders := NewMatrixOrderService(db, nil, OrderDeps{Notif: notif,
		Bookings: NewBookingService(repository.NewBookingRepository(db), repository.NewEmployeeRepository(db),
			repository.NewCustomerRepository(db), repository.NewQualityFollowUpRepository(db), notif, repository.NewInventoryRepository(db)),
		Suggest: NewMatrixSuggestService(repository.NewMatrixSuggestRepository(db)), Actions: repository.NewAiActionRepository(db)})
	repo := repository.NewSecretaryRepository(db)
	call := func(tc *SecretaryTurnCtx, name string, in map[string]any) (string, error) {
		a := NewMatrixAgent(db, nil, "")
		registerOrderTools(a, orders, repo, tc)
		for _, tl := range a.tools {
			if tl.Name == name {
				return tl.Run(in, nil)
			}
		}
		t.Fatalf("no tool %s", name)
		return "", nil
	}
	sup := func() *string {
		var s *string
		db.Get(&s, `SELECT "projectSupervisorId" FROM "Booking" WHERE id = 'so-b1'`)
		return s
	}

	// الدور الأول: «كلّف سجاد بحجز SOB1» — بطاقة بالاسم، وولا شي يتنفّذ.
	turn1 := &SecretaryTurnCtx{UserID: "so-boss", Request: "كلّف سجاد بحجز SOB1"}
	out, err := call(turn1, "assign_crew", map[string]any{"booking": "SOB1", "employee": "سجاد سكرتير"})
	if err != nil || !strings.Contains(out, "orderId=") || len(turn1.Cards) != 1 || !strings.Contains(strings.Join(turn1.Cards[0].Lines, " "), "سجاد") {
		t.Fatalf("assign_crew: %q %v %+v", out, err, turn1.Cards)
	}
	if sup() != nil {
		t.Fatal("prepared order executed without approval")
	}
	// التنفيذ بنفس الدور ممنوع بالكود — حتى لو النموذج حاول.
	if _, err := call(turn1, "run_order", map[string]any{}); err == nil || sup() != nil {
		t.Fatalf("run_order in the same turn must be refused: %v", err)
	}
	// موظف مو واضح: سؤال، بلا أمر.
	if out, _ := call(&SecretaryTurnCtx{UserID: "so-boss"}, "assign_crew", map[string]any{"booking": "SOB1", "employee": "اسم ما موجود قطعاً"}); !strings.Contains(out, "اسأله") {
		t.Fatalf("unknown employee: %q", out)
	}

	// الدور الثاني (المدير گال «اي نفّذ»): يتنفّذ آخر أمر ينتظر.
	turn2 := &SecretaryTurnCtx{UserID: "so-boss", Request: "اي نفّذ"}
	res, err := call(turn2, "run_order", map[string]any{})
	if err != nil || !strings.Contains(res, "✅") {
		t.Fatalf("run_order: %q %v", res, err)
	}
	if s := sup(); s == nil || *s != "so-lead" {
		t.Fatalf("leader not assigned: %v", s)
	}
	// مرة ثانية: ماكو شي ينتظر — ما ينعاد.
	if again, _ := call(&SecretaryTurnCtx{UserID: "so-boss"}, "run_order", map[string]any{}); !strings.Contains(again, "ماكو أمر") {
		t.Fatalf("second run: %q", again)
	}
	// مدير ثاني ما يكدر ينفّذ أمر غيره.
	turn3 := &SecretaryTurnCtx{UserID: "so-boss", Request: "ثبّت"}
	out3, err := call(turn3, "schedule_booking", map[string]any{"booking": "SOB1", "at": time.Now().In(baghdad).AddDate(0, 0, 3).Format("2006-01-02") + "T10:00"})
	if err != nil {
		t.Fatalf("schedule_booking: %v", err)
	}
	id := out3[strings.Index(out3, "orderId=")+8:]
	id = id[:strings.Index(id, ")")]
	if other, _ := call(&SecretaryTurnCtx{UserID: "someone-else"}, "run_order", map[string]any{"orderId": id}); !strings.Contains(other, "ما لگيت") {
		t.Fatalf("another user ran the order: %q", other)
	}
}
