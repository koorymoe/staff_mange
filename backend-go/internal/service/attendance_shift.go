package service

import (
	"fmt"
	"strings"
	"time"
)

// ═══ الشفتات — قرار (ع) 10-06 ═══
// الصباحي يخلص ٤ العصر، والمسائي من ٤ لـ١٢ بالليل. إذا للموظف وقت شفت
// مسجّل (shiftStart/shiftEnd) نعتمده؛ غير هيچ الافتراضي حسب نوع شفته.

const autoCheckoutGrace = 3 * time.Hour

// requiredWork ساعات الشغل المطلوبة باليوم — قرار (ع) 10-07: الانصراف التلقائي
// ما ينسجّل قبل ما يكمّل الموظف ٨ ساعات من وقت حضوره.
const requiredWork = 8 * time.Hour

// RequiredEnd الأبعد بين نهاية الشفت وحضوره + ٨ ساعات.
func RequiredEnd(checkIn, shiftEnd time.Time) time.Time {
	if e := checkIn.Add(requiredWork); e.After(shiftEnd) {
		return e
	}
	return shiftEnd
}

func parseClock(s *string) (int, int, bool) {
	if s == nil {
		return 0, 0, false
	}
	var h, m int
	if _, err := fmt.Sscanf(strings.TrimSpace(*s), "%d:%d", &h, &m); err != nil || h < 0 || h > 24 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}

// ShiftWindow بداية ونهاية الشفت ليوم معيّن (بتوقيت بغداد). المسائي يعبر نص الليل.
func ShiftWindow(shift, start, end *string, day time.Time) (time.Time, time.Time) {
	d := day.In(debriefLoc)
	base := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, debriefLoc)
	sh, sm, okS := parseClock(start)
	eh, em, okE := parseClock(end)
	evening := shift != nil && *shift == "EVENING"
	if !okS {
		sh, sm = 8, 0
		if evening {
			sh = 16
		}
	}
	if !okE {
		eh, em = 16, 0
		if evening {
			eh = 24
		}
	}
	from := base.Add(time.Duration(sh)*time.Hour + time.Duration(sm)*time.Minute)
	to := base.Add(time.Duration(eh)*time.Hour + time.Duration(em)*time.Minute)
	if !to.After(from) {
		to = to.Add(24 * time.Hour)
	}
	return from, to
}

// AutoCheckoutAt وقت الانصراف التلقائي لجلسة: نهاية الشفت (أو حضوره + ٨ ساعات إذا أبعد)، أو آخر نشاط إذا
// بعدها. ok=false: بعد ما فات وقت السماح (٣ ساعات بعد الشفت).
func AutoCheckoutAt(checkIn time.Time, shiftEnd time.Time, lastActivity *time.Time, now time.Time) (time.Time, bool) {
	shiftEnd = RequiredEnd(checkIn, shiftEnd)
	if now.Before(shiftEnd.Add(autoCheckoutGrace)) {
		return time.Time{}, false
	}
	at := shiftEnd
	if lastActivity != nil && lastActivity.After(at) && lastActivity.Before(now) {
		at = *lastActivity
	}
	if at.Before(checkIn) {
		at = checkIn
	}
	return at, true
}

func clockLabel(t time.Time) string {
	t = t.In(debriefLoc)
	if t.Hour() == 0 && t.Minute() == 0 {
		return "١٢ بالليل"
	}
	return t.Format("15:04")
}
