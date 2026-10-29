package model

import (
	"fmt"
	"strconv"
)

// ═══ 💵 المبالغ بالدينار (دازاي 10-29) ═══
// «الليدر يكتب 50 بدل 50000» وماتركس يطلّعها «غير مطابق». الدينار بالآلاف —
// أي مبلغ بين ١ و٩٩٩ أكيد غلط كتابة، فينرفض برسالة توضح الصح. الصفر يبقى
// مسموح (مجاني / ماكو دفعة).

// MinIQD أصغر مبلغ منطقي بالدينار.
const MinIQD = 1000

// CheckIQD يرجّع خطأ إذا المبلغ صغير بشكل يدل على نسيان الأصفار.
func CheckIQD(label string, v float64) error {
	if v > 0 && v < MinIQD {
		n := strconv.FormatFloat(v, 'f', -1, 64)
		return fmt.Errorf("%s «%s» صغير كلش — المبالغ بالدينار بالآلاف. إذا تقصد %s ألف اكتبها %s",
			label, n, n, strconv.FormatFloat(v*1000, 'f', -1, 64))
	}
	return nil
}

// CheckIQDPtr نفس CheckIQD للحقول الاختيارية.
func CheckIQDPtr(label string, v *float64) error {
	if v == nil {
		return nil
	}
	return CheckIQD(label, *v)
}
