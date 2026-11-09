package database

// ══════════════════════════════════════════════════════════════════
// صندوق المراقب — الشغل يجي له، مو هو يدور عليه
// ══════════════════════════════════════════════════════════════════
//
// المراقب عنده وصول لأغلب الشاشات، بس ماكو شي «يوصله». يعني لازم
// يفتح شاشة شاشة ويقلّب ويخمّن شنو تغيّر من آخر مرة دخل. النتيجة إنه
// يقعد ساكت وشغله — الي هو أهم شي بالنظام — ما ينعمل.
//
// هذا الجدول يقلب المعادلة: كل حدث لازم عين عليه يندزّ هنا كصف
// بانتظار قرار. المراقب يفتح صندوقه ويشوف بالضبط شنو جديد وشنو
// ينتظر منه، وكل صف إما «سليم» أو «عندي ملاحظة».
//
// المحطات (stage) — هاي الي طلبها صاحب العمل بالضبط:
//
//	INVOICE_BEFORE_AUDIT  الفاتورة قبل ما يدققها المحاسب
//	INVOICE_AFTER_AUDIT   الفاتورة بعد التدقيق
//	BOOKING_BEFORE_CONFIRM الحجز وموعده قبل التثبيت
//	BOOKING_AFTER_CONFIRM  بعد التثبيت (الكادر والموعد النهائي)
//	BOOKING_AFTER_COMPLETE بعد الإنجاز
//
// ownerRole = شغل منو: المحاسب، إداري الحجوزات، الفني... حتى المراقب
// يفلتر «وريني شغل المحاسب بس».
//
// ⚠️ ليش صف بالجدول مو إشعار؟ الإشعار ينقرا ويضيع. هنا كل صف يبقى
// بحالة PENDING حتى ينبتّ بيه، فما تنفوت وحدة، وينبقى سجل «هذا
// انتدقق من المراقب بهاي الملاحظة» تنفع لو انفتح خلاف بعدين.
func monitorReviewMigration() []Migration {
	return []Migration{
		{
			Version: "0234_monitor_review",
			SQL: `
				CREATE TABLE IF NOT EXISTS "MonitorReview" (
					id           TEXT PRIMARY KEY,
					stage        TEXT NOT NULL,
					-- BOOKING | LEADER_INVOICE — نوع الشي المراقَب
					"entityType" TEXT NOT NULL,
					"entityId"   TEXT NOT NULL,
					-- عنوان جاهز للعرض حتى ما نحتاج نجيب الكيان كله بالقائمة
					title        TEXT NOT NULL,
					summary      TEXT,
					"ownerRole"  TEXT,
					"ownerEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					-- PENDING | OK | FLAGGED
					status       TEXT NOT NULL DEFAULT 'PENDING',
					note         TEXT,
					"reviewedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"reviewedAt"   TIMESTAMPTZ,
					"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- المحطة الوحدة ما تتكرر لنفس الشي: لو الحجز انثبّت مرتين
				-- ما نريد صفين بالصندوق.
				CREATE UNIQUE INDEX IF NOT EXISTS "MonitorReview_unique_stage_idx"
					ON "MonitorReview" ("entityType", "entityId", stage);
				CREATE INDEX IF NOT EXISTS "MonitorReview_pending_idx"
					ON "MonitorReview" (status, "createdAt" DESC);
				CREATE INDEX IF NOT EXISTS "MonitorReview_stage_idx"
					ON "MonitorReview" (stage, status);
			`,
		},
	}
}
