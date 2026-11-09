package database

// حذف المشروع المرفوض نهائياً وإرجاعه — طلب (ع) 10-19: «اريد خيار حذف نهائي… مع
// طلب السبب». المشروع ينشال من كل الشاشات، بس نسخته الكاملة والسبب ومنو
// حذفه يبقون هنا — الحذف بلا أثر يخلّي أي غلطة ما تنصلّح ولا تنعرف.
func deletedProjectMigrations() []Migration {
	return []Migration{{
		Version: "0384_deleted_project",
		SQL: `
			CREATE TABLE IF NOT EXISTS "DeletedProject" (
				id BIGSERIAL PRIMARY KEY,
				"projectId" TEXT NOT NULL,
				code TEXT,
				name TEXT,
				snapshot JSONB NOT NULL,
				reason TEXT NOT NULL,
				"deletedById" TEXT,
				"deletedByName" TEXT,
				"deletedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			-- وإرجاع المرفوض للمشاريع: منو، وليش، ولأي مرحلة.
			CREATE TABLE IF NOT EXISTS "ProjectRestoreLog" (
				id BIGSERIAL PRIMARY KEY,
				"projectId" TEXT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
				"toStage" TEXT NOT NULL,
				reason TEXT NOT NULL,
				"byId" TEXT,
				"byName" TEXT,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
		`,
	}}
}
