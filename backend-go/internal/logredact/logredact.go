// Package logredact — 🙈 حجب الأسرار من السجلات (تعديلات ٦ز — 11-04).
//
// فلتر واحد على مخرج `log`: أي شي يشبه مفتاح، أو توكن، أو باسورد، أو JWT
// ينمسح قبل ما ينكتب — حتى لو مبرمج طبع الطلب كامل بالغلط.
package logredact

import (
	"io"
	"regexp"
)

const mask = "[محجوب]"

type rule struct {
	re   *regexp.Regexp
	repl string
}

var rules = []rule{
	// JWT: ثلاث قطع base64url يبدي بـ eyJ
	{regexp.MustCompile(`eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}`), mask},
	// Authorization: Bearer xxx
	{regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9._~+/=-]{8,}`), "${1}" + mask},
	// مفاتيح مزودين معروفة (Anthropic، OpenAI، Google، GitHub، Slack، AWS)
	{regexp.MustCompile(`sk-(?:ant-)?[A-Za-z0-9_-]{16,}`), mask},
	{regexp.MustCompile(`AIza[0-9A-Za-z_-]{30,}`), mask},
	{regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`), mask},
	{regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{10,}`), mask},
	{regexp.MustCompile(`AKIA[0-9A-Z]{16}`), mask},
	// "password":"…" / "token": "…" / "secret" / "apiKey" … بـ JSON
	{regexp.MustCompile(`(?i)("(?:password|passwd|pass|token|access_token|refresh_token|secret|api_?key|apikey|authorization|otp|code_?verifier|private_?key)"\s*:\s*")[^"]*(")`), "${1}" + mask + "${2}"},
	// key=value بالروابط أو النص: password=… token=… api_key=… secret=…
	{regexp.MustCompile(`(?i)\b(password|passwd|token|access_token|secret|api_?key|apikey|otp)=([^\s&"']+)`), "${1}=" + mask},
	// رابط قاعدة بيانات بيه باسورد: postgres://user:pass@
	{regexp.MustCompile(`(?i)([a-z][a-z0-9+.-]*://[^:/\s@]+:)[^@\s]+@`), "${1}" + mask + "@"},
}

// Redact يرجّع النص بلا أسرار.
func Redact(s string) string {
	for _, r := range rules {
		s = r.re.ReplaceAllString(s, r.repl)
	}
	return s
}

// Writer يلف مخرج السجل.
type Writer struct{ W io.Writer }

func (w Writer) Write(p []byte) (int, error) {
	if _, err := w.W.Write([]byte(Redact(string(p)))); err != nil {
		return 0, err
	}
	return len(p), nil // نرجّع الطول الأصلي — log ما يعتبره كتابة ناقصة
}
