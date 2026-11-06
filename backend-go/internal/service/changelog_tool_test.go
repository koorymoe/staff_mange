package service

import (
	"strings"
	"testing"
	"time"
)

func TestChangelog(t *testing.T) {
	days := parseChangelog(changelogMD)
	if len(days) < 3 {
		t.Fatalf("السجل المضمّن لازم بيه أيام: %d", len(days))
	}
	for i := 1; i < len(days); i++ {
		if !days[i].Date.Before(days[i-1].Date) {
			t.Fatalf("الأحدث لازم فوق: %s بعد %s", days[i].Date, days[i-1].Date)
		}
	}
	for _, d := range days {
		if len(d.Items) == 0 {
			t.Fatalf("يوم بلا بنود: %s", d.Date)
		}
	}
	now := days[0].Date.Add(12 * time.Hour)
	got := ChangelogText(now, 1, "")
	if !strings.Contains(got, days[0].Date.Format("2006-01-02")) || strings.Contains(got, days[len(days)-1].Date.Format("2006-01-02")) {
		t.Fatalf("since_days=1 لازم يجيب آخر يوم بس:\n%s", got)
	}
	if s := ChangelogText(now, 1, "تلغرام"); !strings.Contains(s, "تلغرام") {
		t.Fatalf("البحث لازم يدوّر بكل السجل:\n%s", s)
	}
	if s := ChangelogText(now, 1, "كلمة-ما-موجودة"); !strings.Contains(s, "ماكو") {
		t.Fatal("بحث بلا نتيجة")
	}
	out, _ := changelogTool().Run(map[string]any{"since_days": float64(3)}, nil)
	if out == "" {
		t.Fatal("الأداة رجعت فارغ")
	}
}

func TestSuggestVerifyDue(t *testing.T) {
	at := func(s string) time.Time {
		v, err := time.ParseInLocation("2006-01-02 15:04", s, debriefLoc)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	for _, c := range []struct {
		now, last string
		want      bool
	}{
		{"2026-11-06 23:29", "", false},           // قبل الموعد
		{"2026-11-06 23:30", "", true},            // الموعد
		{"2026-11-06 23:45", "2026-11-06", false}, // اشتغل الليلة
		{"2026-11-07 00:10", "2026-11-06", false}, // بعد نص الليل: يوم جديد بس قبل موعده
		{"2026-11-07 23:31", "2026-11-06", true},  // الليلة الجاية
	} {
		if got := suggestVerifyDue(at(c.now), c.last); got != c.want {
			t.Errorf("%s (آخر %q): %v، المتوقع %v", c.now, c.last, got, c.want)
		}
	}
	// الوقت بالـ UTC ينحسب ببغداد: 20:30 UTC = 23:30 بغداد
	if !suggestVerifyDue(time.Date(2026, 11, 6, 20, 30, 0, 0, time.UTC), "") {
		t.Error("لازم يحسب بتوقيت بغداد")
	}
}
