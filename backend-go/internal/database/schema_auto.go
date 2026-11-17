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
		{
			Version: "0491_project_waiting_customer",
			SQL: `
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "waitingCustomerSince" timestamp without time zone;
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "waitingCustomerNote" text;
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "waitingCustomerById" text;
		`,
		},
		{
			Version: "0492_seed_once",
			SQL: `
			CREATE TABLE IF NOT EXISTS "SeedOnce" (
	"key" text NOT NULL,
	"doneAt" timestamp with time zone DEFAULT now() NOT NULL
);
			ALTER TABLE "SeedOnce" ADD CONSTRAINT "SeedOnce_pkey" PRIMARY KEY (key);
		`,
		},
		{
			Version: "0493_delete_log",
			SQL: `
			CREATE TABLE IF NOT EXISTS "DeleteLog" (
	"id" text NOT NULL,
	"tableName" text NOT NULL,
	"rowId" text NOT NULL,
	"actorId" text,
	"route" text NOT NULL,
	"data" jsonb NOT NULL,
	"createdAt" timestamp with time zone DEFAULT now() NOT NULL
);
			ALTER TABLE "DeleteLog" ADD CONSTRAINT "DeleteLog_pkey" PRIMARY KEY (id);
			CREATE INDEX IF NOT EXISTS "DeleteLog_table_row_idx" ON public."DeleteLog" USING btree ("tableName", "rowId");
		`,
		},
		{
			Version: "0494_booking_assignment_seen",
			SQL: `
			ALTER TABLE "BookingAssignment" ADD COLUMN IF NOT EXISTS "seenAt" timestamp with time zone;
		`,
		},
		{
			Version: "0495_training_curriculum_audiences",
			SQL: `
			ALTER TABLE "TrainingCurriculum" ADD COLUMN IF NOT EXISTS "audiences" text[];
		`,
		},
		{
			Version: "0496_training_curriculum_assignees",
			SQL: `
			ALTER TABLE "TrainingCurriculum" ADD COLUMN IF NOT EXISTS "assignees" text[];
		`,
		},
		{
			Version: "0500_wa_conversation_inbound",
			SQL: `
			ALTER TABLE "WaConversation" ADD COLUMN IF NOT EXISTS "inbound" boolean DEFAULT false NOT NULL;
		`,
		},
		{
			// (دازاي 11-16) «ما أتقيّم من قبل البشر — بس ماتركس يقرا تقاريري ويقيّمني»
			Version: "0501_employee_ai_only_rating",
			SQL: `
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "aiOnlyRating" boolean DEFAULT false NOT NULL;
		`,
		},
		{
			// (دازاي 11-16) «غيرنه دوره… الاشعارات القديمه بعدها ضاهره اله — لازم تضهرله بس
			// الاشعارات الجديدة، وهاي طبقها ع كل واحد يتغير دوره». تريگر على Employee: أي تغيير
			// بالدور أو علامة الليدر — من أي شاشة — يختم roleChangedAt، وقائمة الإشعارات
			// وعدّادها يعرضن بس الي بعده. القديم ما ينحذف (يبقى بالقاعدة)، بس ما يطلع.
			Version: "0502_employee_role_changed_at",
			SQL: `
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "roleChangedAt" timestamp without time zone;
			CREATE OR REPLACE FUNCTION employee_role_changed() RETURNS trigger AS $fn$
			BEGIN
				IF NEW.role IS DISTINCT FROM OLD.role OR NEW."isLeader" IS DISTINCT FROM OLD."isLeader" THEN
					NEW."roleChangedAt" := CURRENT_TIMESTAMP;
				END IF;
				RETURN NEW;
			END
			$fn$ LANGUAGE plpgsql;
			DROP TRIGGER IF EXISTS employee_role_changed_trg ON "Employee";
			CREATE TRIGGER employee_role_changed_trg BEFORE UPDATE ON "Employee"
				FOR EACH ROW EXECUTE FUNCTION employee_role_changed();
		`,
		},
		{
			// (دازاي 11-16) «التقييم حسب كل دور»: التريگر يحفظ الدور القبلي ويا تاريخ التغيير، حتى
			// تقييم الشهر ينقسم: الدور القديم لحد roleChangedAt، والجديد من بعده.
			Version: "0503_employee_prev_role",
			SQL: `
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "prevRole" text;
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "prevIsLeader" boolean;
			CREATE OR REPLACE FUNCTION employee_role_changed() RETURNS trigger AS $fn$
			BEGIN
				IF NEW.role IS DISTINCT FROM OLD.role OR NEW."isLeader" IS DISTINCT FROM OLD."isLeader" THEN
					NEW."roleChangedAt" := CURRENT_TIMESTAMP;
					NEW."prevRole" := OLD.role::text;
					NEW."prevIsLeader" := OLD."isLeader";
				END IF;
				RETURN NEW;
			END
			$fn$ LANGUAGE plpgsql;
		`,
		},
		{
			// (دازاي 11-16) تقي تغيّر دوره (إداري الكوادر ← فني ليدر) قبل ما ينبني التريگر —
			// دازاي: «اعتمد تاريخ اليوم». بداية ١٦ تشرين الثاني بتوقيت بغداد = 21:00 UTC من ١٥.
			Version: "0504_data_role_change_taqi",
			SQL: `
			UPDATE "Employee" SET "roleChangedAt" = '2026-11-15 21:00:00', "prevRole" = 'HR_COORDINATOR', "prevIsLeader" = false
			WHERE name = 'تقي يوسف عبيد' AND "roleChangedAt" IS NULL AND role::text = 'TECHNICIAN';
		`,
		},
		{
			// (11-17 — طلب ماتركس لإداري الكميات) «تذكير على طلبات المواد المعلّقة الي طال عمرها»
			Version: "0505_procurement_reminded_at",
			SQL: `
			ALTER TABLE "ProcurementRequest" ADD COLUMN IF NOT EXISTS "remindedAt" timestamp without time zone;
		`,
		},
		{
			// (دازاي 11-17) «أسعار اشتراك GPS تدخل بالتجديد»: كل طلب تجديد ياخذ سعر نوعه وقت الطلب.
			Version: "0506_gps_renewal_price",
			SQL: `
			ALTER TABLE "GpsRenewalRequest" ADD COLUMN IF NOT EXISTS price double precision;
		`,
		},
		// ‹makemigration› — التحديثات الجديدة تنضاف فوگ هالسطر
	}
}
