package service

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/whatsapp"
)

// scriptedBrain عقل وهمي يستدعي أدوات محددة — نفحص الأدوات بلا نموذج.
type scriptedBrain struct {
	calls []struct {
		name string
		in   map[string]any
	}
}

func (s *scriptedBrain) Reply(_ context.Context, t WaTurn) (string, error) {
	for _, c := range s.calls {
		for _, tool := range t.Tools {
			if tool.Name == c.name {
				if _, err := tool.Run(c.in, nil); err != nil {
					return "", err
				}
			}
		}
	}
	return "تمام 🙏", nil
}

// TestWaAgentTools_Live: التجربة ما تكتب على النظام، والحقيقي يكتب ويا AiAgentAction،
// والإلغاء يرجّع المتابعة، والرسالة المكررة من Meta تنتجاهل، والمرشّحين بس حجوزات آخر ٧ أيام.
func TestWaAgentTools_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	cleanup := func() {
		db.Exec(`DELETE FROM "AiAgentAction" WHERE "conversationId" IN (SELECT id FROM "WaConversation" WHERE "customerId" = 'wat-c')`)
		db.Exec(`DELETE FROM "WaConversation" WHERE "customerId" = 'wat-c'`)
		db.Exec(`DELETE FROM "QualityFollowUp" WHERE "customerId" = 'wat-c'`)
		db.Exec(`DELETE FROM "Booking" WHERE "customerId" = 'wat-c'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'wat-c'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'wat-owner'`)
	}
	cleanup()
	defer cleanup()
	for _, q := range []string{
		`INSERT INTO "Employee"(id, name) VALUES ('wat-owner', 'مالك اختبار')`,
		`INSERT INTO "Customer"(id, name, phone) VALUES ('wat-c', 'زبون', '07701112299')`,
		`INSERT INTO "Booking"(id, code, "customerId", status, "completedAt") VALUES ('wat-b1', 'WAT1', 'wat-c', 'COMPLETED', now() - interval '1 day')`,
		`INSERT INTO "Booking"(id, code, "customerId", status, "completedAt") VALUES ('wat-b0', 'WAT0', 'wat-c', 'COMPLETED', now() - interval '20 days')`,
		`INSERT INTO "QualityFollowUp"(id, "bookingId", "customerId") VALUES ('wat-f1', 'wat-b1', 'wat-c'), ('wat-f0', 'wat-b0', 'wat-c')`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	repo := repository.NewWaRepository(db)
	svc := NewWaAgentService(db, repo, whatsapp.New("", "", "", ""), nil, repository.NewQualityFollowUpRepository(db),
		repository.NewMatrixProposalRepository(db), repository.NewBookingDeleteRequestRepository(db), nil, nil, nil)
	brain := &scriptedBrain{}
	brain.calls = append(brain.calls, struct {
		name string
		in   map[string]any
	}{"record_contact", map[string]any{"outcome": "OK", "customer_rating": float64(5), "crew_rating": float64(4), "notes": "راضي"}})
	svc.SetBrain(WaAgentQuality, brain)
	ctx := context.Background()

	// ① التجربة: الإجراء ينسجل sandbox وما يمس المتابعة
	conv, msgs, err := svc.StartSandbox(WaAgentQuality, "wat-b1")
	if err != nil || len(msgs) != 1 {
		t.Fatalf("sandbox start: %v %d", err, len(msgs))
	}
	_, acts, err := svc.Say(ctx, conv.ID, "تمام", nil)
	if err != nil || len(acts) != 1 || !acts[0].Sandbox {
		t.Fatalf("sandbox say: %v %+v", err, acts)
	}
	var st string
	db.Get(&st, `SELECT status FROM "QualityFollowUp" WHERE id = 'wat-f1'`)
	if st != "PENDING" {
		t.Fatalf("sandbox wrote to the system: %s", st)
	}

	// ② المرشّحين: الجديد بس، القديم (٢٠ يوم) ما ينلمس
	cands, _ := repo.QualityCandidates(50)
	seen := map[string]bool{}
	for _, c := range cands {
		seen[c.FollowUpID] = true
	}
	if !seen["wat-f1"] || seen["wat-f0"] {
		t.Fatalf("candidates wrong: %+v", cands)
	}

	// ③ حقيقي (المفتاح مطفي = ما يرد؛ نحاكي مشتغل بمحادثة BOT ونستدعي الرد مباشرة)
	fu := "wat-f1"
	cust := "wat-c"
	b1 := "wat-b1"
	real, err := repo.CreateConversation(WaAgentQuality, &cust, "9647701112299", &b1, &fu, false)
	if err != nil {
		t.Fatalf("real conv: %v", err)
	}
	id := "wamid.test." + real.ID
	if _, inserted, _ := repo.AddMessage(real.ID, "IN", "CUSTOMER", nil, "تمام", &id); !inserted {
		t.Fatal("first message not inserted")
	}
	if _, inserted, _ := repo.AddMessage(real.ID, "IN", "CUSTOMER", nil, "تمام", &id); inserted {
		t.Fatal("duplicate meta message inserted")
	}
	real, _ = repo.Conversation(real.ID)
	_, acts, err = svc.reply(ctx, real)
	if err != nil || len(acts) != 1 || acts[0].Sandbox {
		t.Fatalf("real reply: %v %+v", err, acts)
	}
	var cr, kr int
	db.QueryRow(`SELECT status, "customerRating", "crewRating" FROM "QualityFollowUp" WHERE id = 'wat-f1'`).Scan(&st, &cr, &kr)
	if st != "CONTACTED_OK" || cr != 5 || kr != 4 {
		t.Fatalf("record_contact: %s %d %d", st, cr, kr)
	}

	// ④ الإلغاء يرجّع المتابعة PENDING
	if _, err := svc.Undo(acts[0].ID, "wat-owner", "غلط"); err != nil {
		t.Fatalf("undo: %v", err)
	}
	db.Get(&st, `SELECT status FROM "QualityFollowUp" WHERE id = 'wat-f1'`)
	if st != "PENDING" {
		t.Fatalf("undo did not reset: %s", st)
	}
	if _, err := svc.Undo(acts[0].ID, "wat-owner", "مرة ثانية"); err == nil {
		t.Fatal("double undo allowed")
	}

	// ⑤ المفتاح مطفي + الواتساب مو مربوط = المشغّل ما يسوي شي
	svc.Scan(ctx)
	if n := repo.NewToday(WaAgentQuality); n != 1 {
		t.Fatalf("scan opened conversations while off: %d", n)
	}
	if waPhone("07701112299") != "9647701112299" || waPhone("+964 770 111 2299") != "9647701112299" {
		t.Fatal("waPhone")
	}
}
