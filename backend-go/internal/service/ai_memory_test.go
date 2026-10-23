package service

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

type fakeMemLLM struct{ got string }

func (f *fakeMemLLM) New(_ context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	f.got = p.Messages[0].Content[0].OfText.Text
	var m anthropic.Message
	err := json.Unmarshal([]byte(`{"id":"m","type":"message","role":"assistant","model":"m","content":[{"type":"text","text":"- طلب مهمة جرد لياسر"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`), &m)
	return &m, err
}

func TestRememberMergesPrevSummary(t *testing.T) {
	llm := &fakeMemLLM{}
	out, err := Remember(context.Background(), llm, "m", "TEST", "- ملخص قديم", []AgentTurn{{Role: "USER", Text: "وجّه جرد لياسر"}, {Role: "ASSISTANT", Text: "انرسلت"}})
	if err != nil || out != "- طلب مهمة جرد لياسر" {
		t.Fatalf("got %q %v", out, err)
	}
	if !strings.Contains(llm.got, "ملخص قديم") || !strings.Contains(llm.got, "المستخدم: وجّه جرد لياسر") || !strings.Contains(llm.got, "المساعد: انرسلت") {
		t.Fatalf("prompt missing parts: %q", llm.got)
	}
}

// المحادثات: كل وحدة برسائلها، والذاكرة تاخذ القديم بس بعد ما يتجمّع.
func TestSecretaryConversations_Live(t *testing.T) {
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
		db.Exec(`DELETE FROM "SecretaryConversation" WHERE "employeeId" IN ('conv-o', 'conv-x')`)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ('conv-o', 'conv-x')`)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status) VALUES ('conv-o', 'مالك محادثات', 'ACTIVE'), ('conv-x', 'غريب', 'ACTIVE')`)
	r := repository.NewSecretaryRepository(db)
	a, err := r.CreateConversation("conv-o", "")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := r.CreateConversation("conv-o", "ثانية")
	if _, err := r.Conversation("conv-x", a.ID); err == nil {
		t.Fatal("other user can open my conversation")
	}
	for i := 0; i < MemoryKeep+MemoryThreshold; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		if _, err := r.SaveMessageIn(&a.ID, "conv-o", role, "رسالة", nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond)
	}
	r.SaveMessageIn(&b.ID, "conv-o", "user", "بالثانية", nil, nil, nil)
	if h, _ := r.ConversationHistory(b.ID, 50); len(h) != 1 {
		t.Fatalf("conversation b leaked messages: %d", len(h))
	}
	r.TitleIfNew(a.ID, "أول سؤال")
	a, _ = r.Conversation("conv-o", a.ID)
	if a.Title != "أول سؤال" || a.Messages != MemoryKeep+MemoryThreshold {
		t.Fatalf("title/count: %q %d", a.Title, a.Messages)
	}
	old, err := r.ToSummarize(a, MemoryKeep, MemoryThreshold)
	if err != nil || len(old) != MemoryThreshold {
		t.Fatalf("to summarize: %d %v", len(old), err)
	}
	if err := r.SetMemory(a.ID, "- ملخص", old[len(old)-1].At); err != nil {
		t.Fatal(err)
	}
	a, _ = r.Conversation("conv-o", a.ID)
	if more, _ := r.ToSummarize(a, MemoryKeep, MemoryThreshold); len(more) != 0 || a.Summary != "- ملخص" {
		t.Fatalf("after memory: %d %q", len(more), a.Summary)
	}
	if err := r.DeleteConversation("conv-x", a.ID); err == nil {
		t.Fatal("other user deleted my conversation")
	}
}
