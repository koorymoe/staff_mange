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
	// دوام مكتوب بنظام ١٢ ساعة (مثل 04:00–12:00): ماكو دوام يبدي الفجر،
	// فالمقصود عصر وليل (16:00–24:00).
	if okS && okE && sh >= 1 && sh <= 6 && eh <= 12 {
		sh += 12
		if eh < 12 {
			eh += 12
		}
	}
	// «12:00» كنهاية لشفت يبدي الظهر أو بعده = ١٢ بالليل (غلط إدخال شائع).
	if okE && eh == 12 && em == 0 && sh >= 12 {
		eh = 24
	}
	from := base.Add(time.Duration(sh)*time.Hour + time.Duration(sm)*time.Minute)
	to := base.Add(time.Duration(eh)*time.Hour + time.Duration(em)*time.Minute)
	if !to.After(from) {
		to = to.Add(24 * time.Hour)
	}
	// شفت أطول من ١٦ ساعة = غلط إدخال: نرجع للافتراضي حتى ما يقفل على أحد.
	if to.Sub(from) > 16*time.Hour {
		return ShiftWindow(shift, nil, nil, day)
	}
	return from, to
}

// ═══ الساعات المحسوبة — قرار (ع) 10-07 ═══
// «إذا بدا من ٨ الصبح ينحسبله الى ٤، وإذا من ٣ مساءً أو الأربعة ينحسبله الى
// ١٢، وبقية الساعات يلغن». الحد حسب ساعة الحضور (بتوقيت بغداد):
//   - قبل ٦ الصبح (بعد نص الليل): ما ينحسب شي إلا إذا خلّص حجز.
//   - قبل ٣ العصر: لحد ٤ العصر.
//   - من ٣ العصر وطالع: لحد ١٢ بالليل.
// وإذا خلّص حجز بعد الحد، ينحسبله لحد خلوص الحجز (تأخّر بالشغل).
// نفس القاعدة بالـSQL: attendance_counted_end (0327).

// CountCap حد الساعات المحسوبة لحضور معيّن (قبل تمديد الحجز).
func CountCap(checkIn time.Time) time.Time {
	l := checkIn.In(debriefLoc)
	base := time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, debriefLoc)
	switch {
	case l.Hour() < 6:
		return base
	case l.Hour() < 15:
		return base.Add(16 * time.Hour)
	default:
		return base.Add(24 * time.Hour)
	}
}

// countLimit الحد ويا تمديد آخر حجز خلّصه.
func countLimit(checkIn time.Time, lastActivity *time.Time) time.Time {
	c := CountCap(checkIn)
	if lastActivity != nil && lastActivity.After(c) {
		c = *lastActivity
	}
	if c.Before(checkIn) {
		c = checkIn
	}
	return c
}

// CountedEnd نهاية الجلسة المحسوبة. assumed=true إذا انقطع من الانصراف
// (ما سجّل انصراف، أو سجّله بعد الحد)، ويا السبب.
func CountedEnd(checkIn time.Time, checkOut *time.Time, lastActivity *time.Time, now time.Time) (time.Time, bool, string) {
	end := now
	if checkOut != nil {
		end = *checkOut
	}
	lim := countLimit(checkIn, lastActivity)
	if !end.After(lim) {
		if end.Before(checkIn) {
			return checkIn, false, ""
		}
		return end, false, ""
	}
	if checkOut == nil && now.Before(lim.Add(autoCheckoutGrace)) {
		return now, false, "" // بعده بالدوام
	}
	why := "سجّل انصراف بعد وقت الدوام"
	if checkOut == nil {
		why = "ما سجّل انصراف"
	}
	if lastActivity != nil && lim.Equal(*lastActivity) {
		return lim, true, why + " — انحسبله لحد آخر حجز خلّصه (" + clockLabel(lim) + ")."
	}
	if lim.Equal(checkIn) {
		return lim, true, why + " — حضور بعد نص الليل ما ينحسب."
	}
	return lim, true, why + " — انحسبله لحد " + clockLabel(lim) + "."
}

// AutoCheckoutAt وقت الانصراف التلقائي: حد الساعات المحسوبة. ok=false لحد
// ما تفوت ٣ ساعات سماح بعد الحد.
func AutoCheckoutAt(checkIn time.Time, lastActivity *time.Time, now time.Time) (time.Time, bool) {
	lim := countLimit(checkIn, lastActivity)
	if now.Before(lim.Add(autoCheckoutGrace)) {
		return time.Time{}, false
	}
	return lim, true
}

func clockLabel(t time.Time) string {
	t = t.In(debriefLoc)
	if t.Hour() == 0 && t.Minute() == 0 {
		return "١٢ بالليل"
	}
	return t.Format("15:04")
}

// InShift الوقت ضمن الدوام (من ساعة قبل البداية لحد النهاية).
func InShift(from, to, now time.Time) bool {
	return !now.Before(from.Add(-time.Hour)) && now.Before(to)
}

// IsEveningShift الشفت يبدي من ١٢ الظهر وطالع.
func IsEveningShift(from time.Time) bool {
	return from.In(debriefLoc).Hour() >= 12
}
