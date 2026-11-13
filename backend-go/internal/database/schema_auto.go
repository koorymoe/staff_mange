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
		{
			Version: "0487_permission_revoke",
			SQL: `
			CREATE TABLE IF NOT EXISTS "PermissionRevoke" (
	"employeeId" text NOT NULL,
	"permission" text NOT NULL,
	"revokedById" text,
	"createdAt" timestamp with time zone DEFAULT now() NOT NULL
);
			ALTER TABLE "PermissionRevoke" ADD CONSTRAINT "PermissionRevoke_pkey" PRIMARY KEY ("employeeId", permission);
			ALTER TABLE "PermissionRevoke" ADD CONSTRAINT "PermissionRevoke_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"(id) ON DELETE CASCADE;
		`,
		},
		{
			Version: "0488_add_matrix_order",
			SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixOrder" (
	"id" text NOT NULL,
	"actorId" text NOT NULL,
	"text" text NOT NULL,
	"status" text DEFAULT 'PROPOSED'::text NOT NULL,
	"actions" jsonb DEFAULT '[]'::jsonb NOT NULL,
	"reply" text,
	"createdAt" timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
	"doneAt" timestamp without time zone
);
			ALTER TABLE "MatrixOrder" ADD CONSTRAINT "MatrixOrder_pkey" PRIMARY KEY (id);
			ALTER TABLE "MatrixOrder" ADD CONSTRAINT "MatrixOrder_actorId_fkey" FOREIGN KEY ("actorId") REFERENCES "Employee"(id) ON DELETE CASCADE;
		`,
		},
		{
			Version: "0489_employee_off_system_work",
			SQL: `
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "offSystemWork" boolean DEFAULT false NOT NULL;
		`,
		},
		// ‹makemigration› — التحديثات الجديدة تنضاف فوگ هالسطر
	}
}
