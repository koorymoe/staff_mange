package database

// ══════════════════════════════════════════════════════════════════
// صف عاجل بصندوق المراقب
// ══════════════════════════════════════════════════════════════════
//
// «من يضغط غير مطابق، الفاتورة تنرفع للمراقب بصورة **عاجلة** — إنو
// هاي بيها مشكلة. أما خطأ بالسعر عادي تصير».
//
// ⚠️ الأحكام الثلاثة كلها تروح «بانتظار الاعتماد» — هذا شغّال أصلاً
// وما يتغيّر. الي ينضاف: **علامة عجلة** على صف المراقب لحالة «غير
// مطابق» وحدها، حتى تطلع بأول الصندوق بشارة حمرا.
//
// ⚠️ ليش عمود مو نص «🔴 عاجل» بالملخّص؟ لأن النص ما ينفرز ولا ينفلتر
// ولا ينعدّ — والمراقب لازم يشوف العاجل أول، مو يدوّر عليه بالقائمة.
func monitorUrgentMigrations() []Migration {
	return []Migration{
		{
			Version: "0269_monitor_review_urgent",
			SQL: `ALTER TABLE "MonitorReview"
			      ADD COLUMN IF NOT EXISTS urgent BOOLEAN NOT NULL DEFAULT false`,
		},
		{
			Version: "0269_monitor_review_urgent_idx",
			// الفرز «العاجل أول» يمر على هذا الفهرس بدل مسح الجدول.
			SQL: `CREATE INDEX IF NOT EXISTS "MonitorReview_urgent_idx"
			      ON "MonitorReview"(urgent, "createdAt" DESC) WHERE status = 'PENDING'`,
		},
	}
}
