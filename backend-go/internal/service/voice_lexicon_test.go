package service

import (
	"regexp"
	"testing"
	"time"
)

func TestVoiceLexiconApply(t *testing.T) {
	l := &VoiceLexicon{loaded: time.Now()}
	for _, f := range [][2]string{{"الدش كاب", "الداش كام"}, {"دش كام", "داش كام"}, {"ماتريكس", "ماتركس"}} {
		l.fixes = append(l.fixes, voiceFix{re: regexp.MustCompile(`(^|[^\p{L}])` + regexp.QuoteMeta(f[0]) + `($|[^\p{L}])`), right: f[1]})
	}
	cases := map[string]string{
		"ركبنا الدش كاب للزبون": "ركبنا الداش كام للزبون",
		"دش كام":                "داش كام",
		"اسأل ماتريكس، شكو":     "اسأل ماتركس، شكو",
		"ماتريكسات":             "ماتريكسات", // ما يلمس نص كلمة
	}
	for in, want := range cases {
		if got := l.Apply(in); got != want {
			t.Errorf("%q → %q, want %q", in, got, want)
		}
	}
}
