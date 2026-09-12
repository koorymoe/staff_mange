package handler

import (
	"strings"
	"unicode/utf8"
)

// TextLen عدد **الحروف** مو البايتات.
//
// ⚠️ len() بالـGo يرجّع البايتات، والحرف العربي بايتين. يعني فحص
// len(s) < 5 على «سيء» (٣ حروف = ٦ بايتات) يمر — والحارس الي كتبناه
// حتى نمنع تقرير بلا سبب صار بلا فايدة. انكشفت بفحص حي.
//
// كل فحص على طول نص يكتبه المستخدم لازم يمر من هنا.
func TextLen(s string) int {
	return utf8.RuneCountInString(strings.TrimSpace(s))
}
