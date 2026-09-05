package database

// ══════════════════════════════════════════════════════════════════
// سجل النسخ الاحتياطية — للمالك وحده
// ══════════════════════════════════════════════════════════════════
//
// ليش جدول أصلاً؟ لأن سكربت النسخ يشتغل على السيرفر نفسه ويكتب بمجلد
// backups/، وحاوية الباك إند ما تشوف هذا المجلد. فلو ما سجّلنا نتيجة كل
// نسخة بقاعدة البيانات، ما اكو طريقة النظام يعرض للمالك حالة النسخ —
// يبقى لازم يفتح SSH كل يوم حتى يطمّن، وهذا الي يخلي الناس تنسى وتكتشف
// إن النسخ واقفة من شهرين بأسوأ لحظة ممكنة.
//
// ⚠️ هذا الجدول ما يُقرأ إلا من مسار /api/owner/backups المحمي بـ
// RequireOwner. مدير النظام (ADMIN) ما يشوفه ولا يعرف بوجوده — هذا شرط
// صريح من المالك: يسلّم النظام ويبقى الإشراف على النسخ عنده هو بس.
func backupRunMigration() []Migration {
	return []Migration{
		{
			Version: "0217_backup_run",
			SQL: `
				CREATE TABLE IF NOT EXISTS "BackupRun" (
					id              TEXT PRIMARY KEY,
					"startedAt"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"finishedAt"    TIMESTAMPTZ,
					ok              BOOLEAN NOT NULL DEFAULT FALSE,
					"fileName"      TEXT,
					"sizeBytes"     BIGINT NOT NULL DEFAULT 0,
					"tableCount"    INTEGER NOT NULL DEFAULT 0,
					encrypted       BOOLEAN NOT NULL DEFAULT FALSE,
					offsite         BOOLEAN NOT NULL DEFAULT FALSE,
					"offsiteTarget" TEXT,
					"hasUploads"    BOOLEAN NOT NULL DEFAULT FALSE,
					"hasEnv"        BOOLEAN NOT NULL DEFAULT FALSE,
					warnings        TEXT,
					error           TEXT,
					"keptCount"     INTEGER NOT NULL DEFAULT 0
				);

				CREATE INDEX IF NOT EXISTS "BackupRun_startedAt_idx"
					ON "BackupRun" ("startedAt" DESC);
			`,
		},
	}
}
