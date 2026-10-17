package database

// ═══ الإعلام والعلاقات العامة — طلب (ع) 10-05 ═══
// مرحلة «📸 الإعلام» بالمشاريع قبل التنفيذ: المشروع يتحوّل للإعلام بكل
// تفاصيله (المهندس المشرف، الموقع، المدة، متى يخلص) حتى يطلعون يصوّرون
// الشغل وينشرونه. ودور جديد «MEDIA» لموظفي الإعلام.
//
// ⚠️ ترحيلان: قيمة الـenum الجديدة ما تنستعمل بنفس المعاملة (نفس درس 0278).
func mediaMigrations() []Migration {
	return []Migration{
		{
			Version: "0319_employee_role_media",
			SQL:     `ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'MEDIA';`,
		},
		{
			Version: "0319_media_briefs",
			SQL: `
				INSERT INTO "Permission" (id, name, label) VALUES
					(gen_random_uuid()::text, 'media', 'الإعلام والعلاقات العامة')
				ON CONFLICT (name) DO NOTHING;

				CREATE TABLE IF NOT EXISTS "MediaBrief" (
					id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
					"projectId"     TEXT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
					"createdById"   TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"engineerId"    TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"startAt"       TIMESTAMPTZ,
					"expectedEndAt" TIMESTAMPTZ,
					duration        TEXT,
					notes           TEXT,
					status          TEXT NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'SCHEDULED', 'SHOT', 'PUBLISHED', 'CANCELLED')),
					"shootAt"       TIMESTAMPTZ,
					"mediaEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"publishedUrl"  TEXT,
					"mediaNotes"    TEXT,
					"createdAt"     TIMESTAMPTZ NOT NULL DEFAULT now(),
					"updatedAt"     TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "MediaBrief_status_idx" ON "MediaBrief" (status, "createdAt" DESC);
				CREATE INDEX IF NOT EXISTS "MediaBrief_project_idx" ON "MediaBrief" ("projectId");`,
		},
	}
}
