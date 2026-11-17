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

// TestMatrixOrdersBooking_Live (دازاي 11-16): ماتركس يوجّه الكوادر ويثبّت الحجوزات وينسّق —
// يقترح من المحرك، وما ينفّذ إلا بـ«نفّذ»، وبنفس مسارات المنسّق.
func TestMatrixOrdersBooking_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	clean := func() {
		db.Exec(`DELETE FROM "MatrixOrder" WHERE "actorId" = 'mob-boss'`)
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" IN ('mob-lead', 'mob-tech', 'mob-boss') OR message LIKE '%MOB_ %'`)
		// اختبارات ثانية تمسح كل الحجوزات وتولّدلها مهام ميدان — تنشال أول، وإلا الحجز ما ينحذف ويبقى ماسك الموظفين.
		db.Exec(`DELETE FROM "Mission" WHERE "bookingId" IN (SELECT id FROM "Booking" WHERE "customerId" = 'mob-cu')`)
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = 'mob-cu'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'mob-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ('mob-lead', 'mob-tech', 'mob-boss')`)
	}
	clean()
	t.Cleanup(func() { clean(); db.Close() })
	// موعد باچر ١٠ الصبح بغداد — scheduledAt ينخزن عالمي بلا منطقة.
	tomorrow := time.Now().In(baghdad).AddDate(0, 0, 1)
	at := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, baghdad)
	utc := at.UTC().Format("2006-01-02 15:04:05")
	db.MustExec(`INSERT INTO "Employee"(id, name, role, "isLeader") VALUES ('mob-boss', 'mob boss', 'OWNER', false),
		('mob-lead', 'ليدر تنسيق', 'TECHNICIAN', true), ('mob-tech', 'فني تنسيق', 'TECHNICIAN', false)`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('mob-cu', 'زبون تنسيق', '07700009922')`)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "scheduledAt") VALUES
		('mob-b1', 'MOB1', 'mob-cu', 'CONFIRMED', $1), ('mob-b2', 'MOB2', 'mob-cu', 'CONFIRMED', $1), ('mob-b3', 'MOB3', 'mob-cu', 'PENDING', NULL)`, utc)

	notif := repository.NewNotificationRepository(db)
	bookings := NewBookingService(repository.NewBookingRepository(db), repository.NewEmployeeRepository(db),
		repository.NewCustomerRepository(db), repository.NewQualityFollowUpRepository(db), notif, repository.NewInventoryRepository(db))
	suggest := NewMatrixSuggestService(repository.NewMatrixSuggestRepository(db))
	s := NewMatrixOrderService(db, nil, OrderDeps{Notif: notif, Bookings: bookings, Suggest: suggest, Actions: repository.NewAiActionRepository(db)})

	// ① حجز بلا موعد ما ينكلّفله كادر — لازم موعد أول.
	if _, err := s.proposeCrew("mob3", "", ""); err == nil || !strings.Contains(err.Error(), "بلا موعد") {
		t.Fatalf("crew for unscheduled booking: %v", err)
	}
	if _, err := s.proposeCrew("NOPE999", "", ""); err == nil {
		t.Fatal("unknown booking code accepted")
	}

	// ② «نسّق حجوزات باچر»: حجزين بنفس الوقت — ولا شخص ينقترح للاثنين (بلا تعارض)، والي ما إله كادر يطلع «يحتاج المنسّق».
	// (قاعدة الفحص ممكن بيها ليدرية من اختبارات ثانية، فما نفترض منو ينختار — بس إن التعارض ممنوع.)
	acts, msg, err := s.proposeCoordinate(at.Format("2006-01-02"))
	if err != nil {
		t.Fatal(err)
	}
	mine := []OrderAction{}
	seen := map[string]string{}
	for _, a := range acts {
		if !strings.HasPrefix(a.BookingCode, "MOB") {
			continue
		}
		mine = append(mine, a)
		for _, id := range crewIDs(a) {
			if other, dup := seen[id]; dup {
				t.Fatalf("%s انقترح لحجزين بنفس الوقت: %s و%s", id, other, a.BookingCode)
			}
			seen[id] = a.BookingCode
		}
	}
	if len(mine) == 0 || mine[0].LeaderID == "" || msg == "" {
		t.Fatalf("coordinate staffed nothing: %+v | %s", acts, msg)
	}

	// ③ الاقتراح ما ينفّذ شي؛ «نفّذ» يكلّف فعلاً ويبلّغ الكادر.
	crew := mine[0]
	s.propose = func(string, string) (string, []OrderAction, error) {
		return "فهمت", []OrderAction{crew, {Kind: orderKindCrew}, {Kind: orderKindSchedule, BookingID: "mob-b3", BookingCode: "MOB3", At: "غلط"}}, nil
	}
	o, err := s.Propose("mob-boss", "كلّف كادر")
	if err != nil || len(o.Actions) != 1 {
		t.Fatalf("validate should keep only the complete action: %+v %v", o, err)
	}
	var sup *string
	db.Get(&sup, `SELECT "projectSupervisorId" FROM "Booking" WHERE id = $1`, crew.BookingID)
	if sup != nil {
		t.Fatal("proposal executed before approval")
	}
	done, err := s.Run(o.ID, "mob-boss", "OWNER", nil)
	if err != nil || done.Status != OrderDone || !done.Actions[0].Done {
		t.Fatalf("run: %+v %v", done, err)
	}
	db.Get(&sup, `SELECT "projectSupervisorId" FROM "Booking" WHERE id = $1`, crew.BookingID)
	var techs, notes int
	db.Get(&techs, `SELECT count(*) FROM "BookingAssignment" WHERE "bookingId" = $1`, crew.BookingID)
	db.Get(&notes, `SELECT count(*) FROM "Notification" WHERE "employeeId" = $2 AND message LIKE '%'||$1||'%'`, crew.BookingCode, crew.LeaderID)
	if sup == nil || *sup != crew.LeaderID || techs != len(crew.Techs) || notes < 1 {
		t.Fatalf("crew not applied: sup=%v techs=%d notes=%d", sup, techs, notes)
	}
	// الحجز صار بيه كادر — اقتراح ثاني ينرفض، بس تسمية موظف بالاسم تمشي.
	if _, err := s.proposeCrew(crew.BookingCode, "", ""); err == nil {
		t.Fatal("re-staffing a staffed booking accepted")
	}

	// ④ تثبيت حجز معلّق: بلا وقت = أقرب وقت فاضي، وبعد «نفّذ» يصير مثبّت بيه.
	sch, err := s.proposeSchedule("MOB3", "")
	if err != nil || sch.At == "" {
		t.Fatalf("schedule: %+v %v", sch, err)
	}
	if _, err := s.proposeSchedule("MOB3", "2020-01-01T10:00"); err == nil {
		t.Fatal("past time accepted")
	}
	s.propose = func(string, string) (string, []OrderAction, error) { return "فهمت", []OrderAction{sch}, nil }
	o2, _ := s.Propose("mob-boss", "ثبّت MOB3")
	if d2, err := s.Run(o2.ID, "mob-boss", "OWNER", nil); err != nil || !d2.Actions[0].Done {
		t.Fatalf("run schedule: %+v %v", d2, err)
	}
	var st, got string
	db.QueryRow(`SELECT status::text, to_char(("scheduledAt" AT TIME ZONE 'UTC') AT TIME ZONE 'Asia/Baghdad', 'YYYY-MM-DD"T"HH24:MI') FROM "Booking" WHERE id = 'mob-b3'`).Scan(&st, &got)
	if st != "CONFIRMED" || got != sch.At {
		t.Fatalf("booking not confirmed at the proposed time: %s %s (want %s)", st, got, sch.At)
	}
}
