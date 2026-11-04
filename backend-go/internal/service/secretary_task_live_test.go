package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// مهمة ايريس: الاسم المكرر أو الموعد الناقص = سؤال بالكلام بلا مسودة؛ الكاملة = مقترح ينرسل فوراً.
func TestSecretaryTaskDirect_Live(t *testing.T) {
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
		db.Exec(`DELETE FROM "SecretaryAction" WHERE "employeeId" = 'sec-t-owner'`)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ('sec-t-owner', 'sec-t-a1', 'sec-t-a2', 'sec-t-b')`)
	}
	cleanup()
	defer cleanup()
	for _, q := range []string{
		`INSERT INTO "Employee"(id, name, status) VALUES ('sec-t-owner', 'مالك تجربة', 'ACTIVE')`,
		`INSERT INTO "Employee"(id, name, status) VALUES ('sec-t-a1', 'زيدون كاظم تجربة', 'ACTIVE'), ('sec-t-a2', 'زيدون فلاح تجربة', 'ACTIVE'), ('sec-t-b', 'ياسر مهدي تجربة', 'ACTIVE')`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	repo := repository.NewSecretaryRepository(db)
	run := func(in map[string]any) (*SecretaryTurnCtx, string) {
		tc := &SecretaryTurnCtx{UserID: "sec-t-owner", Request: "وجّه"}
		a := NewMatrixAgent(db, nil, "")
		registerSecretaryTools(a, repo, tc)
		for _, tl := range a.tools {
			if tl.Name == "propose_task" {
				out, err := tl.Run(in, nil)
				if err != nil {
					t.Fatalf("run: %v", err)
				}
				return tc, out
			}
		}
		t.Fatal("no propose_task")
		return nil, ""
	}
	if tc, out := run(map[string]any{"employee": "زيدون", "title": "جرد", "dueAt": "2030-01-01T10:00"}); tc.Proposal != nil || !strings.Contains(out, "زيدون كاظم") {
		t.Fatalf("ambiguous should ask in text: %q %+v", out, tc.Proposal)
	}
	if tc, out := run(map[string]any{"employee": "ياسر مهدي", "title": "جرد"}); tc.Proposal != nil || !strings.Contains(out, "موعد") {
		t.Fatalf("missing due should ask: %q", out)
	}
	tc, _ := run(map[string]any{"employee": "ياسر مهدي", "title": "جرد المخزن", "dueAt": "2030-01-01T10:00"})
	if tc.Proposal == nil || tc.Proposal.Fields.EmployeeID != "sec-t-b" || tc.Proposal.Fields.DueAt == "" {
		t.Fatalf("complete task should be ready to send: %+v", tc.Proposal)
	}
}
