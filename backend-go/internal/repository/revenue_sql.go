package repository

// ═══ «الإيراد» — تعريف واحد لكل النظام (قرار (ع) 10-06) ═══
//
// قبل: الإحصائيات تجمع المحصّل حتى من الملغي والداخلي، ماتركس يجمع فواتير
// الليدر (حتى المسحوبة)، وتوزيع الخدمات يضيف العربون — فكل شاشة رقم.
// هسه: الإيراد = **الفلوس المستلمة فعلاً**:
//
//	المبلغ المستلم + الدفعة المقدّمة، لحجز منجز (تام أو جزئي)، مو داخلي،
//	مو مؤرشف، مو مطلوب حذفه — + دفعات المشاريع.
//
// قرار (ع) 10-09: دفعات المشروع **تنضاف** على مبلغ الحجز — قبل جانت تحل محله
// (أول دفعة تشيل مبلغ الحجز)، فالمشرف يرفع فلوس والمجموع ما يزيد.
// فواتير الليدر رقم ثاني «المفوتر»، والفرق بينهم يراقبه ماتركس.

// RevenueAmountSQL مبلغ الحجز المستلم.
func RevenueAmountSQL(q string) string {
	return `(COALESCE(` + q + `."amountCollected", 0) + COALESCE(` + q + `."advancePaid", 0))`
}

// RevenueBookingSQL شرط «حجز ينحسب بالإيراد» (بلا AND بادئة).
func RevenueBookingSQL(q string) string {
	return q + `.status::text IN ('COMPLETED', 'PARTIAL') AND ` + BookingCountableSQL(q) +
		` AND ` + q + `."bookingType" IS DISTINCT FROM 'INTERNAL'`
}
