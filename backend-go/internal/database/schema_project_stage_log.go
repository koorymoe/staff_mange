package database

// عين ماتركس على المشاريع — طلب (ع) 10-05: «أي أحد يتوجهله مشروع… نريد نشوف
// شنو ترتيب العمل بالمشروع». المرحلة چانت تنكتب فوق القديمة بلا أثر، فما
// نعرف شكد بقى المشروع بكل مرحلة. المشغّل يسجّل كل تغيير مرحلة لحاله، بأي
// طريق صار. ومنو غيّرها ينعرف من سجل نشاط الموظفين (0305) بنفس اللحظة.
func projectStageLogMigrations() []Migration {
	return []Migration{{
		Version: "0313_project_stage_log",
		SQL: `
			CREATE TABLE IF NOT EXISTS "ProjectStageLog" (
				id BIGSERIAL PRIMARY KEY,
				"projectId" TEXT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
				"fromStage" TEXT,
				"toStage"   TEXT NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "ProjectStageLog_project_idx" ON "ProjectStageLog" ("projectId", "createdAt");

			CREATE OR REPLACE FUNCTION project_stage_log() RETURNS trigger AS $$
			BEGIN
				IF TG_OP = 'INSERT' THEN
					INSERT INTO "ProjectStageLog" ("projectId", "fromStage", "toStage") VALUES (NEW.id, NULL, NEW.stage);
				ELSIF NEW.stage IS DISTINCT FROM OLD.stage THEN
					INSERT INTO "ProjectStageLog" ("projectId", "fromStage", "toStage") VALUES (NEW.id, OLD.stage, NEW.stage);
				END IF;
				RETURN NEW;
			END $$ LANGUAGE plpgsql;

			DROP TRIGGER IF EXISTS project_stage_log_trg ON "Project";
			CREATE TRIGGER project_stage_log_trg AFTER INSERT OR UPDATE OF stage ON "Project"
				FOR EACH ROW EXECUTE FUNCTION project_stage_log();

			-- المشاريع الموجودة: نقطة بداية بمرحلتها الحالية (وقت آخر تعديل).
			INSERT INTO "ProjectStageLog" ("projectId", "fromStage", "toStage", "createdAt")
			SELECT p.id, NULL, p.stage, p."updatedAt" FROM "Project" p
			WHERE NOT EXISTS (SELECT 1 FROM "ProjectStageLog" l WHERE l."projectId" = p.id);
		`,
	}}
}
