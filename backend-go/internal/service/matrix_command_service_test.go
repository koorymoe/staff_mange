package service

import (
	"strings"
	"testing"
)

// الأسماء ما لازم توصل النموذج، وترجع صح بالجواب.
func TestAskAnonymizeRestore(t *testing.T) {
	c := &askContext{names: []string{"علي", "علي حسين", "سارة"}}
	in := "علي حسين تأخر، وسارة مفتوح عندها ٢، وعلي تمام"
	anon := c.anonymize(in)
	for _, n := range c.names {
		if strings.Contains(anon, n) {
			t.Fatalf("الاسم %q وصل للنموذج: %s", n, anon)
		}
	}
	if got := c.restore(anon); got != in {
		t.Fatalf("الإرجاع غلط:\n%s\n%s", got, in)
	}
}

func TestRulesAnswerPicksLate(t *testing.T) {
	c := &askContext{lines: []string{"الحجوزات آخر ٣ أيام: 10، المتأخر 3", "هالشهر لحد اليوم: 5 حجز"}}
	if a := rulesAnswer(c, "منو متأخر اليوم؟"); !strings.Contains(a, "المتأخر 3") || strings.Contains(a, "هالشهر") {
		t.Fatalf("جواب غلط: %s", a)
	}
}
