package database

// ═══ تسوية الحجوزات القديمة ═══
//
// «هذني حجوزات قديمة احنا مشتغّليهن وما نعرف الكادر الي طلع ولا
// التكلفة، فنريده ينكتب عليه (تم الإنجاز بشكل كامل) بدون تفاصيل،
// وبعدين نكمل المحتاجيهن. هذا الخيار يكون مؤقت فقط للمالك».
//
// شغل صار قبل النظام، وبياناته مو موجودة ولا راح تنوجد. فيبقى
// بالطابور للأبد يزاحم الشغل الحقيقي، والعدّادات تكذب.
//
// ⚠️ وليش عمود مخصص مو «نأشّره منجز وخلص»؟ لأن التأشير العادي يشغّل
// **الغرامات**: النظام يشوف حجز منجز بلا فاتورة ولا تقرير، فيغرّم
// الليدر بعد ٢٤ ساعة والإداري بعد ٤٨ — على شغل صار قبل ما يوجد
// النظام ومحد يكدر يوثّقه. يعني تنظيف الطابور يتحوّل لعقوبات ظالمة
// على ناس ما إلهم ذنب.
//
// والعمود يخلّي التسوية **مكشوفة**: منو سوّاها ومتى. حجز ينقفل بلا
// تفاصيل شي مقبول لمرة وحدة بالترحيل، وشي خطير لو صار عادة — والعمود
// يخلّيه ينعدّ وينراقب بدل ما ينخلط بالمنجز الحقيقي.
func legacySettleMigrations() []Migration {
	return []Migration{
		{
			Version: "0248_legacy_settlement",
			SQL: `
				ALTER TABLE "Booking"
					ADD COLUMN IF NOT EXISTS "settledLegacyAt" TIMESTAMPTZ,
					ADD COLUMN IF NOT EXISTS "settledLegacyById" TEXT
						REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "settledLegacyNote" TEXT;

				CREATE INDEX IF NOT EXISTS "Booking_settled_legacy_idx"
					ON "Booking" ("settledLegacyAt") WHERE "settledLegacyAt" IS NOT NULL;
			`,
		},
	}
}
