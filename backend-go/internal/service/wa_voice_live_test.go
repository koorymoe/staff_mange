package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/whatsapp"
)

func TestSpeakable(t *testing.T) {
	got := speakable("تمام 🌹 **موعدك**\nباچر «الساعة عشرة» ✅")
	if got != "تمام موعدك باچر الساعة عشرة" {
		t.Fatalf("speakable = %q", got)
	}
	if speakable("مَرحَباً") != "مَرحَباً" {
		t.Fatal("الحركات العربية لازم تبقى — تساعد النطق")
	}
}

// سياسة أي وكيل تاخذ قواعد الصوت لمّا الدور مكالمة — وما تاخذها بالكتابة.
func TestVoiceRulesOnlyOnCalls(t *testing.T) {
	for _, voice := range []bool{true, false} {
		s := &waScript{steps: []string{"text:تمام"}}
		tr := turn("📞 هلو")
		tr.Voice = voice
		newTestReception(s, &waFakeTools{}).Reply(context.Background(), tr)
		if len(s.calls) != 1 || strings.Contains(s.calls[0], "قواعد المكالمة الصوتية") != voice {
			t.Fatalf("voice=%v: rules presence wrong", voice)
		}
	}
}

// TestWaVoice_Live (دازاي 11-15): المكالمة الحية — نفس عقل الاستقبال، بس الرد ينگال وما ينرسل واتساب.
func TestWaVoice_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	const phone = "9647709997711"
	switches := repository.NewSystemSwitchRepository(db)
	cleanup := func() {
		db.Exec(`DELETE FROM "AiAgentAction" WHERE "conversationId" IN (SELECT id FROM "WaConversation" WHERE phone = $1 OR phone LIKE 'voice-test:wv-%')`, phone)
		db.Exec(`DELETE FROM "WaConversation" WHERE phone = $1 OR phone LIKE 'voice-test:wv-%'`, phone)
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" IN (SELECT id FROM "Customer" WHERE phone = '07709997711')`)
		db.Exec(`DELETE FROM "Customer" WHERE phone = '07709997711'`)
		db.Exec(`DELETE FROM "Service" WHERE id = 'wv-s'`)
		db.Exec(`DELETE FROM "SystemSwitch" WHERE key = $1`, model.SwitchAIVoiceCalls)
	}
	cleanup()
	defer cleanup()
	if _, err := db.Exec(`INSERT INTO "Service"(id, name) VALUES ('wv-s', 'صيانة منظومة اختبار صوت')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var sends int32
	meta := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&sends, 1)
		_, _ = w.Write([]byte(`{"messages":[{"id":"x"}]}`))
	}))
	defer meta.Close()
	repo := repository.NewWaRepository(db)
	customers := repository.NewCustomerRepository(db)
	followups := repository.NewQualityFollowUpRepository(db)
	bookings := NewBookingService(repository.NewBookingRepository(db), repository.NewEmployeeRepository(db), customers,
		followups, repository.NewNotificationRepository(db), repository.NewInventoryRepository(db))
	svc := NewWaAgentService(db, repo, whatsapp.New("tok", "123", "secret", "v").WithBase(meta.URL), bookings, followups,
		repository.NewMatrixProposalRepository(db), repository.NewBookingDeleteRequestRepository(db), switches, nil, nil)
	svc.SetCustomers(customers)
	svc.VoiceReady = true
	ctx := context.Background()
	turns, book, handoff := 0, false, false
	svc.SetBrain(WaAgentReception, funcBrain(func(wt WaTurn) (string, error) {
		turns++
		if !wt.Voice || !strings.HasPrefix(wt.Message, "📞 ") {
			t.Errorf("turn not marked voice: voice=%v msg=%q", wt.Voice, wt.Message)
		}
		if book {
			if _, err := waTool(wt, "create_booking", map[string]any{"kind": "NEW", "service": "صيانة منظومة اختبار صوت",
				"details": "المنظومة تطفي بالليل", "customerName": "زبون مكالمة"}); err != nil {
				return "", err
			}
			return "انسجّل طلبك ✅", nil
		}
		if handoff {
			_, _ = waTool(wt, "handoff", map[string]any{"reason": "يسأل عن السعر"})
			return "زميلنا يتصل بيك", nil
		}
		return "هلا بيك 🌹 **شلون** أكدر أخدمك؟", nil
	}))

	// ① تجربة (بلا رقم): الرد نظيف للنطق، والجملة تنحفظ بـ📞، والترحيب ينحفظ مرة.
	r := svc.VoiceTurn(ctx, "wv-1", "", 1, "هلو", "هلا بيك وياك المساعد")
	if r.End || r.Text != "هلا بيك شلون أكدر أخدمك؟" || turns != 1 {
		t.Fatalf("sandbox turn: %+v turns=%d", r, turns)
	}
	sb, err := repo.SandboxByPhone("voice-test:wv-1")
	if err != nil || !sb.Sandbox || !sb.Inbound {
		t.Fatalf("sandbox conv: %+v %v", sb, err)
	}
	if msgs, _ := repo.Messages(sb.ID); len(msgs) != 3 || msgs[0].Direction != "OUT" || msgs[1].Text != "📞 هلو" {
		t.Fatalf("sandbox messages: %+v", msgs)
	}
	// المنصة تعيد نفس الجملة: نفس الرد، والعقل ما ينادى مرة ثانية.
	if r2 := svc.VoiceTurn(ctx, "wv-1", "", 1, "هلو", "هلا بيك وياك المساعد"); r2.Text != r.Text || turns != 1 {
		t.Fatalf("retry re-ran the brain: %+v turns=%d", r2, turns)
	}
	// جملة فاضية: يطلب الإعادة بلا عقل.
	if r3 := svc.VoiceTurn(ctx, "wv-1", "", 2, "  ", ""); r3.End || turns != 1 || !strings.Contains(r3.Text, "تعيد") {
		t.Fatalf("empty utterance: %+v", r3)
	}

	// ② رقم حقيقي والمفتاح مطفي: اعتذار وسد المكالمة، بلا محادثة وبلا عقل.
	if off := svc.VoiceTurn(ctx, "wv-2", "+"+phone, 1, "هلو", ""); !off.End || turns != 1 {
		t.Fatalf("voice off: %+v", off)
	}
	if _, err := repo.OpenByPhone(phone); err == nil {
		t.Fatal("conversation opened while voice is off")
	}
	// المفتاح ما يشتغل بلا سر الربط.
	svc.VoiceReady = false
	if err := svc.SetEnabled(WaAgentVoice, true, ""); err != ErrVoiceOff {
		t.Fatalf("enable without secret: %v", err)
	}
	svc.VoiceReady = true
	if err := svc.SetEnabled(WaAgentVoice, true, ""); err != nil {
		t.Fatalf("enable: %v", err)
	}

	// ③ رقم حقيقي: يسوي حجز فعلي بالمكالمة، وما ينرسل ولا رسالة واتساب.
	book = true
	rb := svc.VoiceTurn(ctx, "wv-3", phone, 1, "اي سجّلها", "")
	if rb.End || rb.Text != "انسجّل طلبك" {
		t.Fatalf("booking call: %+v", rb)
	}
	conv, err := repo.OpenByPhone(phone)
	if err != nil || !conv.Inbound || conv.Sandbox || conv.BookingID == nil {
		t.Fatalf("call conversation: %+v %v", conv, err)
	}
	var notes string
	if db.Get(&notes, `SELECT notes FROM "Booking" WHERE id = $1`, *conv.BookingID); !strings.Contains(notes, "المنظومة تطفي") {
		t.Fatalf("booking notes: %q", notes)
	}
	if n := atomic.LoadInt32(&sends); n != 0 {
		t.Fatalf("voice turn sent %d WhatsApp messages", n)
	}

	// ④ تحويل لإنسان بالمكالمة: المكالمة تنسد، والجملة الجاية ما توصل العقل.
	book, handoff = false, true
	if rh := svc.VoiceTurn(ctx, "wv-3", phone, 2, "بيش السعر", ""); !rh.End {
		t.Fatalf("handoff should end the call: %+v", rh)
	}
	before := turns
	if rn := svc.VoiceTurn(ctx, "wv-4", phone, 1, "الو", ""); !rn.End || turns != before || !strings.Contains(rn.Text, "زميلنا") {
		t.Fatalf("human-held conversation: %+v turns %d→%d", rn, before, turns)
	}
}
