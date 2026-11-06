package service

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// TestAgentGrowth_Live (دازاي 11-06): القياس، والفجوة تصير اقتراح بالمرة الثالثة بس، والدرس المقترح
// ما يدخل البرومبت قبل الموافقة.
func TestAgentGrowth_Live(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	agent := "TEST_" + uuid.NewString()[:6]
	t.Cleanup(func() {
		db.MustExec(`DELETE FROM "AgentToolStat" WHERE agent = $1`, agent)
		db.MustExec(`DELETE FROM "AgentGap" WHERE agent = $1`, agent)
		db.MustExec(`DELETE FROM "AiSuggestion" WHERE agent = $1`, agent)
		db.MustExec(`DELETE FROM "MatrixLesson" WHERE agent = $1`, agent)
	})

	// ١) القياس: نداءين، واحد فاشل بخطأ بيه رقم يتنظّف
	RecordToolStat(db, agent, "money_month", 30*time.Millisecond, nil)
	RecordToolStat(db, agent, "money_month", 50*time.Millisecond, errors.New("فشل للزبون 07701234567"))
	st, err := AgentToolStats(db, agent, 7)
	if err != nil || len(st) != 1 || st[0].Calls != 2 || st[0].Failures != 1 || st[0].AvgMs != 40 {
		t.Fatalf("القياس: %v %+v", err, st)
	}
	if st[0].LastError == nil || strings.Contains(*st[0].LastError, "07701234567") {
		t.Fatalf("الخطأ لازم يتنظّف من الأرقام: %v", st[0].LastError)
	}

	// ٢) الفجوة: المرة الثالثة تصنع اقتراح واحد، والرابعة ما تكرر
	var notes []string
	notify := func(m string) { notes = append(notes, m) }
	for i := 0; i < 4; i++ {
		if _, err := recordGap(db, agent, "رصيد مخزن مادة", "ماكو أداة للرصيد", notify); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "AiSuggestion" WHERE agent = $1 AND "needsCode"`, agent)
	if n != 1 || len(notes) != 1 {
		t.Fatalf("الفجوة المتكررة لازم تصير اقتراح واحد: %d اقتراح، %d إشعار", n, len(notes))
	}

	// ٣) الدرس المقترح ما يدخل قبل الموافقة
	if _, err := proposeLesson(db, agent, "التاريخ دائماً بتوقيت بغداد", "غلطت بيوم", nil); err != nil {
		t.Fatal(err)
	}
	if LessonsBlock(db, agent) != "" {
		t.Fatal("الدرس المقترح دخل البرومبت قبل الموافقة")
	}
	db.MustExec(`UPDATE "MatrixLesson" SET status = 'ACTIVE', active = true WHERE agent = $1`, agent)
	if !strings.Contains(LessonsBlock(db, agent), "بتوقيت بغداد") {
		t.Fatal("بعد الموافقة لازم يدخل")
	}
	if out := performanceText(db, agent, 7); !strings.Contains(out, "money_month") || !strings.Contains(out, "رصيد مخزن مادة") {
		t.Fatalf("my_performance ناقص:\n%s", out)
	}
}

func TestSelfReviewDue(t *testing.T) {
	at := func(s string) time.Time {
		v, _ := time.ParseInLocation("2006-01-02 15:04", s, debriefLoc)
		return v
	}
	// 2026-11-06 جمعة
	for _, c := range []struct {
		now, last string
		want      bool
	}{
		{"2026-11-06 23:44", "", false},
		{"2026-11-06 23:45", "", true},
		{"2026-11-06 23:50", "2026-11-06", false},
		{"2026-11-07 23:50", "", false}, // سبت
		{"2026-11-13 23:46", "2026-11-06", true},
	} {
		if got := selfReviewDue(at(c.now), c.last); got != c.want {
			t.Errorf("%s: %v، المتوقع %v", c.now, got, c.want)
		}
	}
}
