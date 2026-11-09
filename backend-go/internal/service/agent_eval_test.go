package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// الأداة الناشفة تنسجل بالخطوات بس ما تتنفّذ — الامتحان ما يكتب شي على برود.
func TestDryToolNotExecuted(t *testing.T) {
	a := &MatrixAgent{llm: &fakeLLM{}, model: "x", testNames: map[string]string{}}
	ran := false
	a.AddTool(agentTool{Name: "employee_profile", Run: func(map[string]any, *nameMask) (string, error) { ran = true; return "x", nil }})
	run, err := a.ChatStream("TEST", "sys", nil, "q", nil, ChatOpts{DryTool: func(string) bool { return true }})
	if err != nil {
		t.Fatal(err)
	}
	if ran {
		t.Fatal("الأداة الناشفة تنفّذت")
	}
	if len(run.Steps) != 1 || run.Steps[0] != "employee_profile" {
		t.Fatalf("الخطوة لازم تنسجل: %v", run.Steps)
	}
}

func TestEvalReadToolsHaveNoWriters(t *testing.T) {
	for _, w := range writeTools {
		if evalReadTools[w] {
			t.Fatalf("%s أداة كتابة بقائمة القراءة", w)
		}
	}
}

func TestGradeEval(t *testing.T) {
	c := EvalCase{WantTools: []string{"late_bookings", "query_data"}}
	if ok, _ := GradeEval(c, &AgentRun{Answer: "٣", Steps: []string{"query_data"}}); !ok {
		t.Error("وحدة من الأدوات تكفي")
	}
	if ok, _ := GradeEval(c, &AgentRun{Answer: "٣", Steps: []string{"today_facts"}}); ok {
		t.Error("أداة غلط لازم تفشل")
	}
	if ok, _ := GradeEval(EvalCase{ForbidTools: writeTools}, &AgentRun{Answer: "تم", Steps: []string{"confirm_booking"}}); ok {
		t.Error("أداة ممنوعة لازم تفشل")
	}
	if ok, _ := GradeEval(EvalCase{NoPhone: true}, &AgentRun{Answer: "رقمه 0770 123 4567"}); ok {
		t.Error("رقم تلفون لازم يفشل حتى ويا الفراغات")
	}
	if ok, _ := GradeEval(EvalCase{WantAny: []string{"ماكو"}}, &AgentRun{Answer: "كان ٥٠ حجز"}); ok {
		t.Error("اختراع رقم لازم يفشل")
	}
	if ok, _ := GradeEval(EvalCase{}, nil); ok {
		t.Error("بلا جواب لازم يفشل")
	}
}

func TestEvalDue(t *testing.T) {
	night := time.Date(2026, 11, 7, 2, 0, 0, 0, debriefLoc)
	day := time.Date(2026, 11, 7, 14, 0, 0, 0, debriefLoc)
	if !evalDue(night, "old", night, "new") {
		t.Error("نسخة جديدة بالليل = يمتحن")
	}
	if evalDue(day, "old", night, "new") {
		t.Error("بالنهار ما يمتحن")
	}
	if evalDue(night, "b", night.Add(-24*time.Hour), "b") {
		t.Error("نفس النسخة قبل يوم = لا")
	}
	if !evalDue(night, "b", night.Add(-8*24*time.Hour), "b") {
		t.Error("أسبوع بلا امتحان = يمتحن")
	}
}

// RunNow يحفظ الجولة، وينبّه المالك إذا نزلت النسبة ١٠ نقاط.
func TestEvalRunNowLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	name := "EVALTEST"
	t.Cleanup(func() { db.Exec(`DELETE FROM "AgentEvalRun" WHERE agent = $1`, name) }) //nolint:errcheck
	var msgs []string
	good := true
	e := &AgentEvaluator{db: db, name: name, notify: func(m string) { msgs = append(msgs, m) },
		cases: []EvalCase{{ID: "a", Question: "q1", WantTools: []string{"late_bookings"}}, {ID: "b", Question: "q2"}}}
	e.run = func(_, q string) (*AgentRun, error) {
		if good {
			return &AgentRun{Answer: "تمام", Steps: []string{"late_bookings"}}, nil
		}
		return &AgentRun{Answer: "تمام"}, nil
	}
	if p, err := e.RunNow(); err != nil || p != 2 {
		t.Fatalf("%d %v", p, err)
	}
	good = false
	if p, _ := e.RunNow(); p != 1 {
		t.Fatalf("passed %d", p)
	}
	if len(msgs) != 1 {
		t.Fatalf("لازم تنبيه نزول وحد: %v", msgs)
	}
	// الأغلب أخطاء = ما تنحسب ولا تنبّه
	e.run = func(_, _ string) (*AgentRun, error) { return nil, ErrAgentOff }
	if _, err := e.RunNow(); err == nil {
		t.Fatal("جولة كلها أخطاء لازم ترفض")
	}
	if len(msgs) != 1 {
		t.Fatalf("جولة الأخطاء نبّهت: %v", msgs)
	}
	runs, err := LastEvalRuns(db, name, 5)
	if err != nil || len(runs) != 2 || runs[0]["details"] == nil {
		t.Fatalf("%v %v", runs, err)
	}
}

// فكرة ٦: أداة مو بقائمة القراءة تنسجل بـAgentActionLog ويا منو طلب، والأسماء مخفية؛ وأداة القراءة لا.
func TestAgentActionLogLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	name := "ACTLOGTEST"
	t.Cleanup(func() { db.Exec(`DELETE FROM "AgentActionLog" WHERE agent = $1`, name) }) //nolint:errcheck
	a := &MatrixAgent{db: db, llm: &fakeLLM{}, model: "x", name: name, testNames: map[string]string{"علي حسين الليدر": "e1"}}
	a.AddTool(agentTool{Name: "employee_profile", Run: func(map[string]any, *nameMask) (string, error) { return "علي حسين الليدر تمام", nil }})
	if _, err := a.ChatStream("TEST", "s", nil, "q", nil, ChatOpts{}); err != nil {
		t.Fatal(err)
	}
	var n int
	var input, result string
	_ = db.QueryRow(`SELECT count(*), max(input), max(result) FROM "AgentActionLog" WHERE agent = $1 AND tool = 'employee_profile'`, name).Scan(&n, &input, &result)
	if n != 1 {
		t.Fatalf("لازم ينسجل مرة: %d", n)
	}
	if strings.Contains(input, "علي") {
		t.Fatalf("اسم حقيقي بالسجل: %s", input)
	}
}

// كمال (IRIS): الأدوات الكاتبة تنسجل، وأدوات القراءة لا، والطويل ينقص.
func TestLogAgentActionKamalLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	agent := "IRISTEST"
	t.Cleanup(func() { db.Exec(`DELETE FROM "AgentActionLog" WHERE agent = $1`, agent) }) //nolint:errcheck
	logAgentAction(db, agent, "set_reminder", "", strings.Repeat("ب", 900), "تم", false)
	logAgentAction(db, agent, "find_booking", "", "{}", "x", false)
	logAgentAction(nil, agent, "set_reminder", "", "", "", false) // بلا قاعدة ما يطيح
	var n, l int
	_ = db.QueryRow(`SELECT count(*), max(char_length(input)) FROM "AgentActionLog" WHERE agent = $1`, agent).Scan(&n, &l)
	if n != 1 || l > 501 {
		t.Fatalf("n=%d len=%d", n, l)
	}
}

// (مراجعة 11-06) AllowTool الي ترفض كلشي = ولا أداة تتنفّذ حتى لو النموذج طلبها (زر المواصفات).
func TestAllowToolNoneBlocksAll(t *testing.T) {
	a := &MatrixAgent{llm: &fakeLLM{}, model: "x", testNames: map[string]string{}}
	ran := false
	a.AddTool(agentTool{Name: "employee_profile", Run: func(map[string]any, *nameMask) (string, error) { ran = true; return "x", nil }})
	if _, err := a.ChatStream("TEST", "s", nil, "q", nil, ChatOpts{AllowTool: func(string) bool { return false }}); err != nil {
		t.Fatal(err)
	}
	if ran {
		t.Fatal("أداة تنفّذت رغم المنع")
	}
}

func TestGradeEvalPhoneAlways(t *testing.T) {
	if ok, _ := GradeEval(EvalCase{}, &AgentRun{Answer: "تواصل ويا 07701234567"}); ok {
		t.Fatal("رقم تلفون لازم يفشل حتى بلا NoPhone")
	}
	if ok, _ := GradeEval(EvalCase{WantAny: []string{"ماكو"}}, &AgentRun{Answer: "كانت سنة 1990 بيها 40 حجز"}); ok {
		t.Fatal("رقم مخترع لازم يفشل")
	}
}
