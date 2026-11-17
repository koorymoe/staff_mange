package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/whatsapp"
)

// تقييم الزبائن يدخل باختيار الكادر بس بعيّنة كافية — وبلاها صفر، مو عقوبة.
func TestCustomerRatingScore(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	if customerRatingScore(nil, 10) != 0 || customerRatingScore(f(1), custRatingMinN-1) != 0 {
		t.Fatal("no sample must be neutral")
	}
	if hi, lo := customerRatingScore(f(5), 5), customerRatingScore(f(2), 5); hi != 6 || lo != -3 {
		t.Fatalf("score: %v %v", hi, lo)
	}
}

// funcBrain عقل وهمي بدالة — كل دور يقرر شنو يستدعي من الأدوات.
type funcBrain func(t WaTurn) (string, error)

func (f funcBrain) Reply(_ context.Context, t WaTurn) (string, error) { return f(t) }

func waTool(t WaTurn, name string, in map[string]any) (string, error) {
	for _, tool := range t.Tools {
		if tool.Name == name {
			return tool.Run(in, nil)
		}
	}
	return "", errors.New("أداة مو موجودة: " + name)
}

// TestWaReception_Live (دازاي 11-15): الزبون يراسل/يدز فويس/يتصل والوكيل يرد.
// رقم جديد ← محادثة استقبال ← حجز ← وقت فاضي ← تثبيت. الفويس ينكتب نص. المكالمة الفايتة تنرد برسالة.
// والمفتاح المطفي = ولا شي ينفتح. والتجربة ما تكتب على النظام.
func TestWaReception_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const phone, phone2 = "9647709998877", "9647709998866"
	switches := repository.NewSystemSwitchRepository(db)
	cleanup := func() {
		db.Exec(`DELETE FROM "AiAgentAction" WHERE "conversationId" IN (SELECT id FROM "WaConversation" WHERE phone IN ($1, $2) OR (sandbox AND inbound))`, phone, phone2)
		db.Exec(`DELETE FROM "WaConversation" WHERE phone IN ($1, $2) OR (sandbox AND inbound)`, phone, phone2)
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" IN (SELECT id FROM "Customer" WHERE phone IN ('07709998877', '07709998866'))`)
		db.Exec(`DELETE FROM "Customer" WHERE phone IN ('07709998877', '07709998866')`)
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = 'war-c'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'war-c'`)
		db.Exec(`DELETE FROM "Service" WHERE id = 'war-s'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'war-l'`)
		db.Exec(`DELETE FROM "SystemSwitch" WHERE key = $1`, model.SwitchAIReceptionAgent)
	}
	cleanup()
	defer cleanup()
	for _, q := range []string{
		`INSERT INTO "Service"(id, name) VALUES ('war-s', 'تنصيب كاميرات اختبار')`,
		`INSERT INTO "Employee"(id, name, "isLeader") VALUES ('war-l', 'ليدر اختبار', true)`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	// التعلّم: تقييم الزبائن (من متابعة الجودة) لشغل الليدر يوصل لمحرك اختيار الكادر.
	for _, q := range []string{
		`INSERT INTO "Customer"(id, name, phone) VALUES ('war-c', 'زبون قديم', '07709998855')`,
		`INSERT INTO "Booking"(id, code, "customerId", status, "completedAt", "projectSupervisorId")
			SELECT 'war-b' || g, 'WAR' || g, 'war-c', 'COMPLETED', now() - interval '3 days', 'war-l' FROM generate_series(1, 3) g`,
		`INSERT INTO "QualityFollowUp"(id, "bookingId", "customerId", status, "contactedAt", "crewRating", "customerRating")
			SELECT 'war-f' || g, 'war-b' || g, 'war-c', 'CONTACTED_OK', now(), 5, 3 FROM generate_series(1, 3) g`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("seed ratings: %v", err)
		}
	}
	people, err := repository.NewMatrixSuggestRepository(db).People(nil, time.Now())
	if err != nil {
		t.Fatalf("people: %v", err)
	}
	learned := false
	for _, p := range people {
		if p.ID == "war-l" {
			learned = p.CustN == 3 && p.CustAvg != nil && *p.CustAvg == 5
		}
	}
	if !learned {
		t.Fatalf("customer ratings did not reach crew selection: %+v", people)
	}

	// Meta وهمي: يستلم الإرسال ويخدم الفويس.
	var mu sync.Mutex
	var sent []string
	var meta *httptest.Server
	meta = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/messages"):
			var body struct {
				To   string `json:"to"`
				Text struct {
					Body string `json:"body"`
				} `json:"text"`
			}
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
			mu.Lock()
			sent = append(sent, body.To+"|"+body.Text.Body)
			mu.Unlock()
			_, _ = w.Write([]byte(`{"messages":[{"id":"wamid.out"}]}`))
		case r.URL.Path == "/media-ok" || r.URL.Path == "/media-bad":
			_ = json.NewEncoder(w).Encode(map[string]string{"url": meta.URL + "/file", "mime_type": "audio/ogg"})
		case r.URL.Path == "/file":
			if r.Header.Get("Authorization") != "Bearer tok" {
				http.Error(w, "no auth", http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte("OggS-fake-audio"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer meta.Close()
	sentCount := func() int { mu.Lock(); defer mu.Unlock(); return len(sent) }
	lastSent := func() string { mu.Lock(); defer mu.Unlock(); return sent[len(sent)-1] }

	repo := repository.NewWaRepository(db)
	customers := repository.NewCustomerRepository(db)
	followups := repository.NewQualityFollowUpRepository(db)
	bookings := NewBookingService(repository.NewBookingRepository(db), repository.NewEmployeeRepository(db), customers,
		followups, repository.NewNotificationRepository(db), repository.NewInventoryRepository(db))
	wa := whatsapp.New("tok", "123", "secret", "verify").WithBase(meta.URL)
	svc := NewWaAgentService(db, repo, wa, bookings, followups, repository.NewMatrixProposalRepository(db),
		repository.NewBookingDeleteRequestRepository(db), switches, nil, nil)
	svc.SetCustomers(customers)
	svc.SetSlots(NewMatrixSuggestService(repository.NewMatrixSuggestRepository(db)))
	heard := 0
	svc.SetTranscriber(func(_ context.Context, audio []byte, ct string) (string, error) {
		heard++
		if string(audio) != "OggS-fake-audio" || ct != "audio/ogg" {
			return "", errors.New("الفويس ما وصل صح")
		}
		if heard == 2 {
			return "", errors.New("ضجيج")
		}
		return "اي سجّلها", nil
	})
	ctx := context.Background()
	slotRe := regexp.MustCompile(`للأداة: (\d{4}-\d{2}-\d{2}T\d{2}:\d{2})`)
	turn := 0
	var confirmedAt string
	svc.SetBrain(WaAgentReception, funcBrain(func(wt WaTurn) (string, error) {
		turn++
		if wt.Agent != WaAgentReception {
			t.Errorf("turn agent = %s", wt.Agent)
		}
		switch turn {
		case 1:
			if !strings.Contains(wt.Context, "رقم جديد") {
				t.Errorf("context should say new number: %s", wt.Context)
			}
			if _, err := waTool(wt, "confirm_booking", map[string]any{"at": "2099-01-01T10:00"}); err == nil {
				t.Error("confirm_booking بلا حجز لازم يرفض")
			}
			return "شنو المنطقة؟", nil
		case 2:
			if wt.Message != "🎤 اي سجّلها" {
				t.Errorf("voice text = %q", wt.Message)
			}
			if _, err := waTool(wt, "create_booking", map[string]any{"kind": "NEW", "service": "كاميرات اختبار", "details": "يريد ٤ كاميرات للبيت"}); err == nil {
				t.Error("رقم جديد بلا اسم لازم يرفض")
			}
			if _, err := waTool(wt, "create_booking", map[string]any{"kind": "NEW", "service": "خدمة مو موجودة", "details": "يريد ٤ كاميرات للبيت", "customerName": "زبون استقبال"}); err == nil {
				t.Error("خدمة مو موجودة لازم ترفض")
			}
			if _, err := waTool(wt, "create_booking", map[string]any{"kind": "NEW", "service": "كاميرات اختبار", "details": "يريد ٤ كاميرات للبيت",
				"address": "المنصور", "customerName": "زبون استقبال"}); err != nil {
				return "", err
			}
			slots, err := waTool(wt, "free_slots", nil)
			if err != nil {
				return "", err
			}
			m := slotRe.FindStringSubmatch(slots)
			if m == nil {
				t.Fatalf("ماكو أوقات فاضية: %s", slots)
			}
			confirmedAt = m[1]
			if _, err := waTool(wt, "confirm_booking", map[string]any{"at": confirmedAt}); err != nil {
				return "", err
			}
			return "انثبّت موعدك 🌹", nil
		}
		return "تدلل", nil
	}))

	// ① المفتاح مطفي: الرسالة ما تفتح محادثة ولا ينرد عليها.
	svc.Inbound(ctx, whatsapp.Inbound{MetaID: "wamid.r0", From: phone, Text: "هلو"})
	if _, err := repo.OpenByPhone(phone); err == nil || sentCount() != 0 || turn != 0 {
		t.Fatal("الاستقبال مطفي وانفتحت محادثة")
	}
	if err := switches.Set(model.SwitchAIReceptionAgent, true, ""); err != nil {
		t.Fatalf("switch: %v", err)
	}

	// ② رسالة من رقم جديد: تنفتح محادثة استقبال، والرد الأول يعرّف إنه مساعد ذكي.
	svc.Inbound(ctx, whatsapp.Inbound{MetaID: "wamid.r1", From: phone, Text: "اريد كاميرات للبيت"})
	conv, err := repo.OpenByPhone(phone)
	if err != nil || !conv.Inbound || conv.Agent != WaAgentBooking || conv.CustomerID != nil || conv.BookingID != nil {
		t.Fatalf("inbound conv: %+v %v", conv, err)
	}
	if sentCount() != 1 || !strings.HasPrefix(lastSent(), phone+"|") || !strings.Contains(lastSent(), "مساعد ذكي") || !strings.Contains(lastSent(), "شنو المنطقة؟") {
		t.Fatalf("first reply: %v", sent)
	}
	// المكرر من Meta ما ينرد عليه مرتين.
	svc.Inbound(ctx, whatsapp.Inbound{MetaID: "wamid.r1", From: phone, Text: "اريد كاميرات للبيت"})
	if sentCount() != 1 || turn != 1 {
		t.Fatal("duplicate inbound answered twice")
	}

	// ③ فويس: ينسحب من Meta، ينكتب نص، والوكيل يسوي الحجز ويثبّته بوقت فاضي.
	svc.Inbound(ctx, whatsapp.Inbound{MetaID: "wamid.r2", From: phone, Audio: &whatsapp.Audio{MediaID: "media-ok"}})
	conv, _ = repo.Conversation(conv.ID)
	if conv.BookingID == nil || conv.CustomerID == nil {
		t.Fatalf("conversation not bound: %+v", conv)
	}
	var b struct {
		Status      string  `db:"status"`
		Sched       *string `db:"sched"`
		ServiceID   *string `db:"serviceId"`
		Address     *string `db:"address"`
		Notes       *string `db:"notes"`
		Name, Phone string
	}
	if err := db.QueryRowx(`SELECT b.status::text, to_char((b."scheduledAt" AT TIME ZONE 'UTC') AT TIME ZONE 'Asia/Baghdad', 'YYYY-MM-DD"T"HH24:MI'), b."serviceId", b.address, b.notes, c.name, c.phone
		FROM "Booking" b JOIN "Customer" c ON c.id = b."customerId" WHERE b.id = $1`, *conv.BookingID).
		Scan(&b.Status, &b.Sched, &b.ServiceID, &b.Address, &b.Notes, &b.Name, &b.Phone); err != nil {
		t.Fatalf("booking: %v", err)
	}
	if b.Status != "CONFIRMED" || b.Sched == nil || *b.Sched != confirmedAt || b.ServiceID == nil || *b.ServiceID != "war-s" ||
		b.Address == nil || *b.Address != "المنصور" || b.Name != "زبون استقبال" || b.Phone != "07709998877" ||
		b.Notes == nil || !strings.Contains(*b.Notes, "🤖") {
		t.Fatalf("booking wrong: %s sched=%v svc=%v addr=%v notes=%v (want at %s)", b.Status, *b.Sched, *b.ServiceID, *b.Address, *b.Notes, confirmedAt)
	}
	acts, _ := svc.Actions("", conv.ID)
	kinds := map[string]bool{}
	for _, a := range acts {
		kinds[a.Kind] = true
	}
	if !kinds["CREATE_BOOKING"] || !kinds["CONFIRM_BOOKING"] {
		t.Fatalf("actions not recorded: %+v", acts)
	}
	if sentCount() != 2 || !strings.Contains(lastSent(), "انثبّت") {
		t.Fatalf("voice reply: %v", sent)
	}

	// ④ فويس ما انسمع: رد ثابت يطلب الإعادة — بلا عقل.
	before := turn
	svc.Inbound(ctx, whatsapp.Inbound{MetaID: "wamid.r3", From: phone, Audio: &whatsapp.Audio{MediaID: "media-bad"}})
	if turn != before || sentCount() != 3 || !strings.Contains(lastSent(), "ما كدرت أسمع") {
		t.Fatalf("failed voice: turn %d→%d sent %v", before, turn, sent)
	}

	// ⑤ مكالمة فايتة من رقم ثاني: محادثة استقبال ورسالة متابعة، والمكرر ما ينعاد.
	call := whatsapp.Inbound{MetaID: "wacid.1", From: phone2, Call: &whatsapp.Call{}}
	svc.Inbound(ctx, call)
	svc.Inbound(ctx, call)
	c2, err := repo.OpenByPhone(phone2)
	if err != nil || !c2.Inbound || sentCount() != 4 || !strings.HasPrefix(lastSent(), phone2+"|") || !strings.Contains(lastSent(), "شفنا اتصالك") {
		t.Fatalf("missed call: %+v %v %v", c2, err, sent)
	}
	msgs, _ := repo.Messages(c2.ID)
	if len(msgs) != 2 || !strings.Contains(msgs[0].Text, "📞") || turn != before {
		t.Fatalf("missed call messages: %+v", msgs)
	}
	// مكالمة انردّ عليها: تنسجّل بس، بلا رسالة.
	svc.Inbound(ctx, whatsapp.Inbound{MetaID: "wacid.2", From: phone2, Call: &whatsapp.Call{Answered: true, Duration: 95}})
	if msgs, _ = repo.Messages(c2.ID); len(msgs) != 3 || !strings.Contains(msgs[2].Text, "(2 د)") || sentCount() != 4 {
		t.Fatalf("answered call: %+v", msgs)
	}

	// ⑥ إنسان استلم المحادثة: الوكيل يسكت حتى بالفويس.
	if err := svc.Takeover(c2.ID, ""); err == nil {
		svc.Inbound(ctx, whatsapp.Inbound{MetaID: "wamid.r4", From: phone2, Text: "وينكم"})
		if sentCount() != 4 || turn != before {
			t.Fatal("agent answered a conversation a human took over")
		}
	}

	// ⑦ التجربة: حجز الاستقبال ما ينكتب بالنظام، والدور الجاي يكمل على الحجز الوهمي.
	sb, first, err := svc.StartSandbox(WaAgentReception, "")
	if err != nil || len(first) != 0 || !sb.Sandbox || !sb.Inbound {
		t.Fatalf("sandbox start: %+v %v", sb, err)
	}
	var nBefore, nAfter int
	db.Get(&nBefore, `SELECT count(*) FROM "Booking"`)
	svc.SetBrain(WaAgentReception, funcBrain(func(wt WaTurn) (string, error) {
		if _, err := waTool(wt, "confirm_booking", map[string]any{"at": confirmedAt}); err == nil {
			// الدور الثاني: الحجز الوهمي موجود فالتثبيت يعبر
			return "انثبّت (تجربة)", nil
		}
		_, err := waTool(wt, "create_booking", map[string]any{"kind": "MAINTENANCE", "details": "الكاميرا تطفي بالليل"})
		return "انسجّل (تجربة)", err
	}))
	if _, a1, err := svc.Say(ctx, sb.ID, "الكاميرا تطفي", nil); err != nil || len(a1) != 1 || !a1[0].Sandbox || a1[0].Kind != "CREATE_BOOKING" {
		t.Fatalf("sandbox create: %v %+v", err, a1)
	}
	if out, a2, err := svc.Say(ctx, sb.ID, "اي", nil); err != nil || len(a2) != 1 || a2[0].Kind != "CONFIRM_BOOKING" || !strings.Contains(out[len(out)-1].Text, "انثبّت") {
		t.Fatalf("sandbox confirm: %v %+v", err, a2)
	}
	if db.Get(&nAfter, `SELECT count(*) FROM "Booking"`); nAfter != nBefore {
		t.Fatalf("sandbox wrote bookings: %d → %d", nBefore, nAfter)
	}

	// ⑨ الجودة من أي تواصل: زبون عنده متابعة معلّقة وراسلنا — السياق يذكرها، والوكيل يسجّل رأيه.
	db.Exec(`UPDATE "QualityFollowUp" SET status = 'PENDING', "contactedAt" = NULL, "crewRating" = NULL, "customerRating" = NULL WHERE id = 'war-f1'`)
	db.Exec(`DELETE FROM "WaConversation" WHERE phone = '9647709998855'`)
	sawQuality := false
	svc.SetBrain(WaAgentReception, funcBrain(func(wt WaTurn) (string, error) {
		sawQuality = strings.Contains(wt.Context, "متابعة جودة معلّقة") && strings.Contains(wt.Context, "WAR1")
		_, err := waTool(wt, "record_contact", map[string]any{"outcome": "OK", "customerRating": float64(5), "crewRating": float64(4), "notes": "راضي"})
		return "شكراً إلك", err
	}))
	svc.Inbound(ctx, whatsapp.Inbound{MetaID: "wamid.q1", From: "9647709998855", Text: "هلو"})
	var qst string
	var qcr int
	db.QueryRow(`SELECT status, "crewRating" FROM "QualityFollowUp" WHERE id = 'war-f1'`).Scan(&qst, &qcr)
	if !sawQuality || qst != "CONTACTED_OK" || qcr != 4 {
		t.Fatalf("quality from inbound contact: saw=%v status=%s crew=%d", sawQuality, qst, qcr)
	}
	db.Exec(`DELETE FROM "AiAgentAction" WHERE "conversationId" IN (SELECT id FROM "WaConversation" WHERE phone = '9647709998855')`)
	db.Exec(`DELETE FROM "WaConversation" WHERE phone = '9647709998855'`)

	// ⑧ محادثة الاستقبال الساكتة أكثر من يوم تنسد (حتى ما تبقى ماسكة الرقم).
	db.Exec(`UPDATE "WaConversation" SET status = 'BOT', "lastMessageAt" = now() - interval '2 days' WHERE id = $1`, c2.ID)
	if n := repo.CloseIdleInbound(); n < 1 {
		t.Fatal("idle inbound conversation not closed")
	}
	if _, err := repo.OpenByPhone(phone2); err == nil {
		t.Fatal("idle conversation still open")
	}
}
