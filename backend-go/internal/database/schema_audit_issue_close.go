package database

// ═══ إغلاق بلاغ التدقيق: منو أغلقه وليش وشنو سوّى ═══
//
// ⚠️⚠️ **الزر چان ما يوصّل لشي.** «أغلق البلاغ» چان ينفّذ سطراً
// واحداً: `status='RESOLVED', resolvedAt=now()`. يعني:
//   - ماكو `resolvedById` — ما ينحفظ **منو** أغلقه
//   - ماكو سبب — ما ينحفظ **ليش**
//   - ماكو إشعار — المحاسب الي سجّل البلاغ ما يدري شنو صار بيه
//   - وماكو أي أثر على الليدر الي البلاغ عليه
//
// فالبلاغ يختفي من القائمة وينتهي. صاحب النظام سأل حرفياً «من
// أنطي أغلق البلاغ وين يروح؟» — والجواب چان: ما يروح لأي مكان.
//
// **قراره**: ما ينغلق إلا **بإجراء** — إما مخالفة انضباط على
// الليدر، أو تصريح صريح «تأكدت ماكو خطأ». والاثنان ينحفظان
// باسم من سوّاهن، ويوصل إشعار للمحاسب والليدر.
//
// ⚠️ **`actionKind` إجباري وقت الإغلاق** (يتفحّص بالخدمة): بدونه
// يرجع الإغلاق الروتيني — يضغط ويسكّر بلا ما يقرا، وهاي بالضبط
// العلّة الي نصلّحها.
//
// ⚠️ **و`resolvedById` بـ`SET NULL` مو CASCADE**: حذف الموظف ما
// يمحي **إن البلاغ انغلق**. القرار واقعة حصلت، وهوية صاحبه
// تنفقد أما الواقعة فتبقى. و`resolvedByName` منسوخ نصاً حتى
// يبقى السطر مقروءاً بعدها.
func auditIssueCloseMigrations() []Migration {
	return []Migration{
		{
			Version: "0264_audit_issue_close_action",
			SQL: `
				ALTER TABLE "BookingAuditIssue"
					ADD COLUMN IF NOT EXISTS "resolvedById"   TEXT
						REFERENCES "Employee"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "resolvedByName" TEXT,
					ADD COLUMN IF NOT EXISTS "resolveReason"  TEXT,
					ADD COLUMN IF NOT EXISTS "actionKind"     TEXT;

				-- المراقب يفتح «المفتوحة» بكل مرة، فالفهرس عليها.
				CREATE INDEX IF NOT EXISTS "BookingAuditIssue_status_idx"
					ON "BookingAuditIssue" (status, "createdAt" DESC);
			`,
		},
	}
}
