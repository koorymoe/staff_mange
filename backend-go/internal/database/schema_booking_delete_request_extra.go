package database

// ═══ طلب حذف الحجز: القناة ونوع الطلب و«معلقة» ═══
//
// التصميم يطلب بطاقتين وعمودين ماكو وراهن بيانة بالنظام:
//   - «القناة» (موقع ويب / تطبيق الجوال / مركز الاتصال)
//   - «نوع الطلب» (حجز متكرر / إلغاء من الزبون / تصحيح بيانات)
//   - بطاقة «معلقة» بجنب «بانتظار الموافقة» — وبالنظام حالة
//     وحدة بس اسمها PENDING.
//
// **قراره**: القناة ونوع الطلب حقلان فعليان يُدخلان وقت تسجيل
// الطلب، مو استخلاص من نص السبب. و«معلقة» تصنيف فرعي داخل
// PENDING يحدّده المعتمِد بزر «اطلب معلومات إضافية» — مو حالة
// رابعة، لأن الطلب لسه ما انبتّ فيه.
//
// ⚠️ `needsInfoById` بـ`SET NULL` مثل `decidedById` الموجود:
// حذف الموظف ما يمحي إن الطلب صار ناقصه معلومات ومتى.
func bookingDeleteRequestExtraMigrations() []Migration {
	return []Migration{
		{
			Version: "0266_booking_delete_request_extra",
			SQL: `
				ALTER TABLE "BookingDeleteRequest"
					ADD COLUMN IF NOT EXISTS "channel"       TEXT,
					ADD COLUMN IF NOT EXISTS "requestType"    TEXT,
					ADD COLUMN IF NOT EXISTS "needsInfo"      BOOLEAN NOT NULL DEFAULT false,
					ADD COLUMN IF NOT EXISTS "needsInfoNote"  TEXT,
					ADD COLUMN IF NOT EXISTS "needsInfoAt"    TIMESTAMP,
					ADD COLUMN IF NOT EXISTS "needsInfoById"  TEXT
						REFERENCES "Employee"(id) ON DELETE SET NULL;

				-- الشاشة تفتح دائماً على غير المبتوت فيها، والفرز فيها فرعي بـneedsInfo
				CREATE INDEX IF NOT EXISTS "BookingDeleteRequest_status_needsinfo_idx"
					ON "BookingDeleteRequest" (status, "needsInfo", "createdAt" DESC);
			`,
		},
	}
}
