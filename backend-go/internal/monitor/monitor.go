// Package monitor ربط Sentry (دازاي 11-05) — يلزم أخطاء السيرفر ويدزها لـ Sentry.
//
// ⚠️ قاعدة صاحب النظام: أسماء الموظفين والزبائن والرواتب ما تطلع لأي جهة خارجية.
// فكل نص يمر بـ Scrub قبل الإرسال: الأرقام الطويلة (تلفونات، مبالغ، معرّفات)، والإيميلات،
// وأي نص عربي (الأسماء بالنظام عربية) تتبدّل برموز. وماكو بيانات طلب ولا مستخدم ولا كوكيز.
// بلا SENTRY_DSN_BACKEND = مطفي بالكامل، وكل الدوال ما تسوي شي.
package monitor

import (
	"log"
	"regexp"
	"time"

	"github.com/getsentry/sentry-go"
)

var enabled bool

var (
	reEmail  = regexp.MustCompile(`[\w.+-]+@[\w-]+(\.[\w-]+)+`)
	reDigits = regexp.MustCompile(`\d{4,}`)
	reArabic = regexp.MustCompile(`[\p{Arabic}][\p{Arabic}\s،]*`)
	reUUID   = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	reQuoted = regexp.MustCompile(`=\([^)]*\)`) // تفاصيل pq: Key (phone)=(0790…)
)

// Scrub يشيل أي شي ممكن يكون اسم أو رقم أو إيميل من النص قبل ما يطلع.
func Scrub(s string) string {
	s = reEmail.ReplaceAllString(s, "[email]")
	s = reUUID.ReplaceAllString(s, "[id]")
	s = reQuoted.ReplaceAllString(s, "=(…)")
	s = reDigits.ReplaceAllString(s, "[n]")
	s = reArabic.ReplaceAllString(s, "[ar] ")
	return s
}

// Init يشغّل Sentry إذا انحدد الـ DSN. release = نسخة الكود (اختياري).
func Init(dsn, environment, release string) {
	if dsn == "" {
		log.Printf("[monitor] Sentry مطفي (ماكو SENTRY_DSN_BACKEND)")
		return
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Release:          release,
		SendDefaultPII:   false,
		AttachStacktrace: true,
		MaxBreadcrumbs:   0,
		BeforeSend: func(e *sentry.Event, _ *sentry.EventHint) *sentry.Event {
			e.User = sentry.User{}
			e.Request = nil
			e.ServerName = ""
			e.Message = Scrub(e.Message)
			for i := range e.Exception {
				e.Exception[i].Value = Scrub(e.Exception[i].Value)
				noSource(e.Exception[i].Stacktrace)
			}
			for i := range e.Threads {
				noSource(e.Threads[i].Stacktrace)
			}
			e.Extra = nil
			return e
		},
	})
	if err != nil {
		log.Printf("[monitor] Sentry ما اشتغل: %v", err)
		return
	}
	enabled = true
	log.Printf("[monitor] Sentry شغّال (%s)", environment)
}

// Error يدز خطأ ويا مكانه (where = ملف:سطر أو اسم العملية).
func Error(err error, where string) {
	if !enabled || err == nil {
		return
	}
	sentry.WithScope(func(s *sentry.Scope) {
		s.SetTag("where", where)
		sentry.CaptureException(err)
	})
}

// Message يدز رسالة (مثلاً خطأ داخلي ويا رقمه المرجعي).
func Message(msg, ref string, status int) {
	if !enabled {
		return
	}
	sentry.WithScope(func(s *sentry.Scope) {
		if ref != "" {
			s.SetTag("ref", ref)
		}
		if status > 0 {
			s.SetTag("status", statusClass(status))
		}
		sentry.CaptureMessage(msg)
	})
}

// Panic يدز panic انلزم بالـ recover.
func Panic(v any, where string) {
	if !enabled {
		return
	}
	sentry.WithScope(func(s *sentry.Scope) {
		s.SetTag("where", where)
		s.SetLevel(sentry.LevelFatal)
		sentry.CurrentHub().Recover(v)
	})
}

// Flush يدز الباقي قبل ما يطفي السيرفر.
func Flush() {
	if enabled {
		sentry.Flush(3 * time.Second)
	}
}

func statusClass(s int) string {
	switch {
	case s >= 500:
		return "5xx"
	case s >= 400:
		return "4xx"
	}
	return "other"
}

// noSource يشيل أسطر الكود المرفقة ويا الإطارات (الملف والسطر يبقون) — السطر ممكن بيه نص أو قيمة.
func noSource(st *sentry.Stacktrace) {
	if st == nil {
		return
	}
	for i := range st.Frames {
		st.Frames[i].ContextLine = ""
		st.Frames[i].PreContext = nil
		st.Frames[i].PostContext = nil
		st.Frames[i].Vars = nil
	}
}
