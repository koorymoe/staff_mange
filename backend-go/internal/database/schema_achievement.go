package database

// ══════════════════════════════════════════════════════════════════
// الإنجازات — تقرير يومي حر من أي موظف بأي دور (ماتركس)
// ══════════════════════════════════════════════════════════════════
//
// طلب صاحب النظام حرفياً: كل موظف يرفع تقريراً يومياً بشغله — الليدر
// شكد حجز طلّع، مهندس الجودة شكد زبون اتصل، المراقب شكد دقّق،
// المحاسب شكد طابق — حتى لو مو شغله يرفع تقرير. الهدف المعلن: يخلي
// ماتركس يتعرف أكثر على الموظفين وسلوكهم وطريقة تفكيرهم.
//
// ⚠️ الربط بحجز اختياري بالتصميم (`bookingId` NULLABLE): الفني يربطه
// بحجز، بس مهندس الجودة/المراقب/المحاسب يرفعون تقريراً عاماً بلا حجز.
//
// 🔴 الوجهة حصراً: مدير النظام والمالك (`requireAdmin` بالخادم) — لا
// المراقب ولا زميل ثانٍ. لهذا هذا مساره الخاص، مو صندوق المراقب
// (`MonitorReview`) — ذاك مسموح لدور `MONITOR` يشوفه (`requireMonitor`
// بمسارات `/api/monitor-reviews`)، وهذا بالضبط الي ما نريده هنا.
func achievementMigration() []Migration {
	return []Migration{
		{
			Version: "0295_achievement",
			SQL: `
				CREATE TABLE IF NOT EXISTS "Achievement" (
					id             TEXT PRIMARY KEY,
					"employeeId"   TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
					role           TEXT NOT NULL,
					"bookingId"    TEXT REFERENCES "Booking"(id) ON DELETE SET NULL,
					"reportText"   TEXT NOT NULL,
					"reviewStatus" TEXT NOT NULL DEFAULT 'PENDING'
						CHECK ("reviewStatus" IN ('PENDING', 'GOOD', 'NEEDS_REVIEW')),
					"reviewNote"   TEXT,
					"reviewedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"reviewedAt"   TIMESTAMPTZ,
					"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "Achievement_employeeId_idx"
					ON "Achievement" ("employeeId", "createdAt" DESC);
				CREATE INDEX IF NOT EXISTS "Achievement_reviewStatus_idx"
					ON "Achievement" ("reviewStatus");
				CREATE INDEX IF NOT EXISTS "Achievement_createdAt_idx"
					ON "Achievement" ("createdAt");
			`,
		},
	}
}
