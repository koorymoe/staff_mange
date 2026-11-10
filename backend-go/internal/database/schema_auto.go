package database

// ═══ تحديثات مولّدة من الموديلات (🏴‍☠️ ون بيس 11-10) ═══
// لا تكتب هنا بإيدك: عدّل الموديل بـinternal/models، وشغّل
//
//	DATABASE_URL=… go run ./cmd/makemigration -name وصف_قصير
//
// والأداة تضيف التحديث هنا. الحذف أو تغيير شي موجود ما يتولّد — يطلع تنبيه وينتظر موافقة.
func autoMigrations() []Migration {
	return []Migration{
		{
			Version: "0471_admin_edit_log",
			SQL: `
			CREATE SEQUENCE IF NOT EXISTS "AdminEditLog_id_seq";
			CREATE TABLE IF NOT EXISTS "AdminEditLog" (
	"id" bigint DEFAULT nextval('"AdminEditLog_id_seq"'::regclass) NOT NULL,
	"tableName" text NOT NULL,
	"action" text NOT NULL,
	"rowKey" jsonb NOT NULL,
	"before" jsonb,
	"after" jsonb,
	"byId" text NOT NULL,
	"createdAt" timestamp with time zone DEFAULT now() NOT NULL
);
			ALTER TABLE "AdminEditLog" ADD CONSTRAINT "AdminEditLog_pkey" PRIMARY KEY (id);
			CREATE INDEX IF NOT EXISTS "AdminEditLog_table_idx" ON public."AdminEditLog" USING btree ("tableName", "createdAt" DESC);
			ALTER SEQUENCE "AdminEditLog_id_seq" OWNED BY "AdminEditLog"."id";
		`,
		},
		// ‹makemigration› — التحديثات الجديدة تنضاف فوگ هالسطر
	}
}
