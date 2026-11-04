// Package timeutil يوحّد تفسير الأوقات الي يكتبها الموظف بيده.
//
// المشكلة: النظام يخزن نوعين وقت بنفس الأعمدة:
//   - الأوقات الي يكتبها النظام لحاله (now() بقاعدة البيانات) تنخزن
//     بالتوقيت العالمي — وقتها المتصفح ببغداد يزيد ٣ ساعات ويطلع صحيح.
//   - الأوقات الي يكتبها الموظف بحقل datetime-local تجي «٢:٠٠» وهو
//     يقصد ٢:٠٠ بغداد، وتنخزن ٢:٠٠ حرفياً — فالمتصفح يزيد ٣ ساعات
//     ويعرضها ٥:٠٠. يعني كل موعد حجز يتقدم لوحده.
//
// الحل: الوقت الي يجي من الموظف ينتفهم على أنه بغداد وينتحول للتوقيت
// العالمي قبل التخزين، فيتساوى مع باقي أوقات النظام.
package timeutil

import "time"

const CompanyTimeZone = "Asia/Baghdad"

// نحمّلها مرة وحدة. إذا الصورة ما بيها قاعدة مناطق زمنية، نرجع لفرق
// بغداد الثابت (+3) — العراق ما عنده توقيت صيفي، فالثابت صحيح دائماً.
var companyLoc = func() *time.Location {
	if loc, err := time.LoadLocation(CompanyTimeZone); err == nil {
		return loc
	}
	return time.FixedZone("+03", 3*60*60)
}()

// صيغ حقل datetime-local بالمتصفح: بثواني أو بدونها.
var localLayouts = []string{"2006-01-02T15:04", "2006-01-02T15:04:05"}

// ParseCompanyLocal يفسّر وقتاً كتبه الموظف على أنه توقيت بغداد ويرجعه
// بالتوقيت العالمي. الوقت الي جاي أصلاً بمنطقة زمنية (فيه Z أو +03:00)
// ينمر مثل ما هو — يعني الواجهات الي ترسل ISO كامل ما تتأثر.
func ParseCompanyLocal(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC(), nil
	}
	var lastErr error
	for _, layout := range localLayouts {
		t, err := time.ParseInLocation(layout, value, companyLoc)
		if err == nil {
			return t.UTC(), nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

// NormalizeCompanyLocal نفس الي فوق بس يرجع نص جاهز للتخزين. إذا ما
// عرف يفسّر القيمة يرجعها مثل ما هي بدل ما يفشّل العملية كلها.
func NormalizeCompanyLocal(value string) string {
	if value == "" {
		return value
	}
	t, err := ParseCompanyLocal(value)
	if err != nil {
		return value
	}
	return t.Format("2006-01-02T15:04:05")
}

// CompanyLocation توقيت الشركة (بغداد) — لكل «اليوم/الساعة» الي تنعرض للناس.
func CompanyLocation() *time.Location { return companyLoc }

// MonthStart (DRY 11-02) أول الشهر «YYYY-MM» بتوقيت الشركة، ونهايته (أول الشهر الجاي).
// ok=false لو الصيغة غلط. مصدر واحد لـ«فترة الشهر» — چانت مكتوبة بأكثر من مكان.
func MonthStart(period string) (from, to time.Time, ok bool) {
	t, err := time.ParseInLocation("2006-01", period, companyLoc)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	return t, t.AddDate(0, 1, 0), true
}

// Today تاريخ اليوم YYYY-MM-DD بتوقيت الشركة (بغداد). ⚠️ السيرفر يشتغل UTC:
// time.Now().Format("2006-01-02") يطلع **أمس** من ١٢ الليل لـ ٣ الفجر.
func Today() string { return time.Now().In(companyLoc).Format("2006-01-02") }

// ThisMonth الشهر YYYY-MM بتوقيت الشركة.
func ThisMonth() string { return time.Now().In(companyLoc).Format("2006-01") }
