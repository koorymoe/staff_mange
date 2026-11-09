package handler

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// botBodyClear (دازاي 11-06): هل شرح البلاغ واضح كفاية؟ قبل كان يكفي ٢٥ حرف و٥ كلمات،
// فعبر «Ggggg تجريبي جدا جدا جدا لازم». هسه، بعد ما نشيل الكلمات الي هي حرف واحد مكرر
// (ggggg، ههههه) ونختصر التكرار داخل الكلمة:
//   - ٢٥ حرف و٥ كلمات على الأقل (مثل قبل)
//   - ٤ كلمات مختلفة على الأقل (كل وحدة حرفين أو أكثر) — «جدا جدا جدا» = كلمة وحدة
//   - ٦٠٪ من النص حروف، مو رموز وأرقام
func botBodyClear(body string) bool {
	var words []string
	for _, w := range strings.Fields(strings.ToLower(body)) {
		if w = squashRepeats(w); w != "" {
			words = append(words, w)
		}
	}
	clean := strings.Join(words, " ")
	if utf8.RuneCountInString(clean) < botMinBodyRunes || len(words) < botMinBodyWords {
		return false
	}
	distinct := map[string]bool{}
	letters, total := 0, 0
	for _, w := range words {
		n := 0
		for _, r := range w {
			total++
			if unicode.IsLetter(r) {
				letters++
				n++
			}
		}
		if n >= 2 {
			distinct[w] = true
		}
	}
	return len(distinct) >= 4 && letters*10 >= total*6
}

// squashRepeats يختصر أي حرف متكرر ٣ مرات أو أكثر لحرف واحد، ويرجّع "" إذا الكلمة كلها
// حرف واحد مكرر (ggggg، ههههه، !!!!).
func squashRepeats(w string) string {
	var b strings.Builder
	var prev rune
	run, distinct := 0, map[rune]bool{}
	for _, r := range w {
		distinct[r] = true
		if r == prev {
			run++
			if run >= 2 { // الثالث وما بعده يتشال
				continue
			}
		} else {
			prev, run = r, 0
		}
		b.WriteRune(r)
	}
	if len(distinct) == 1 && utf8.RuneCountInString(w) >= 3 {
		return ""
	}
	return b.String()
}
