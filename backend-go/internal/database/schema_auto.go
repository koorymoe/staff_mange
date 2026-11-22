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
		{
			// (بلاغ #50 — دازاي 11-18، خيار B) الحجز ينفتح للمنسّق من «تجهيز المنتجات» —
			// المشاريع الي هسه بالتجهيز ينفتح حجزها هم (الجديدة يفتحها الكود).
			Version: "0507_data_project_prepare_unlock",
			SQL: `
			UPDATE "Booking" b SET "projectExecutionAt" = now(), "updatedAt" = now()
			FROM "Project" p
			WHERE p."bookingId" = b.id AND p.stage = '5. تجهيز المنتجات' AND b."projectExecutionAt" IS NULL;
		`,
		},
		{
			// (دازاي 11-18) التدريب ما يحبس الموظف عن شغله: المدير يجدول وقت يومي (بغداد HH:MM)
			// يطلعله بيه تنبيه التدريب، وبقية الوقت يشتغل عادي.
			Version: "0508_employee_training_window",
			SQL: `
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "trainingFrom" text;
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "trainingTo" text;
		`,
		},
		{
			// (دازاي 11-18 — التقييم الجديد للفني) السلبي ما ينحسب إلا بعد تأكيد المسؤول:
			// تقييم الليدر السيء ← مدير الإدارة/نائبه؛ المشكلة الوظيفية ← منو الغلطان؛
			// نقص العدّة بعد الشغل ← إداري الكميات يگول «مبرَّر».
			Version: "0509_score_review",
			SQL: `
			ALTER TABLE "CrewRating" ADD COLUMN IF NOT EXISTS "reviewStatus" text;
			ALTER TABLE "CrewRating" ADD COLUMN IF NOT EXISTS "reviewedById" text;
			ALTER TABLE "CrewRating" ADD COLUMN IF NOT EXISTS "reviewedAt" timestamp with time zone;
			ALTER TABLE "WorkplaceIssue" ADD COLUMN IF NOT EXISTS "faultEmployeeId" text;
			ALTER TABLE "WorkplaceIssue" ADD COLUMN IF NOT EXISTS "faultDecided" boolean NOT NULL DEFAULT false;
			ALTER TABLE "WorkplaceIssue" ADD COLUMN IF NOT EXISTS "faultById" text;
			ALTER TABLE "WorkplaceIssue" ADD COLUMN IF NOT EXISTS "faultAt" timestamp with time zone;
			ALTER TABLE "BookingAfterInventory" ADD COLUMN IF NOT EXISTS "justifiedById" text;
			ALTER TABLE "BookingAfterInventory" ADD COLUMN IF NOT EXISTS "justifiedAt" timestamp with time zone;
		`,
		},
		{
			// (دازاي 11-18 — تقييم الليدر) الوقت القياسي لكل خدمة: ماتركس يحسبه من الشغل الفعلي،
			// والمالك/مدير النظام يگدر يعدّله. صف هنا = تعديل يدوي يغلب المحسوب.
			Version: "0510_service_standard_time",
			SQL: `
			CREATE TABLE IF NOT EXISTS "ServiceStandardTime" (
				"serviceId" text PRIMARY KEY REFERENCES "Service"(id) ON DELETE CASCADE,
				minutes integer NOT NULL CHECK (minutes BETWEEN 5 AND 1440),
				"setById" text,
				"setAt" timestamp with time zone NOT NULL DEFAULT now()
			);
		`,
		},
		{
			Version: "0511_training_match_items",
			SQL: `
			ALTER TABLE "TrainingMatch" ADD COLUMN IF NOT EXISTS "items" text[] DEFAULT '{}'::text[] NOT NULL;
		`,
		},
		{
			// (دازاي 11-19) «هوه يومية ينعاد الشغل مالته نفسه شلون تأخر يوم؟» — واجب يومي 🔁
			// بالخطة ما إله موعد ولا «متأخر»: درجته من نقاط ماتركس لشغل صاحبه بالشهر.
			Version: "0511_org_plan_item_recurring",
			SQL: `
			ALTER TABLE "OrgPlanItem" ADD COLUMN IF NOT EXISTS recurring boolean NOT NULL DEFAULT false;
		`,
		},
		{
			// البنود الحالية الي مكتوب بيها إنها كل يوم، وبنود مدير سلاسل التوريد لهالشهر (دازاي 11-19 «كلها يومية»)
			Version: "0512_data_plan_items_recurring",
			SQL: `
			UPDATE "OrgPlanItem" SET recurring = true
			WHERE NOT recurring AND (title ~ '(كل يوم|يومي|يومياً|يومية|اليوم واليوم التالي)'
			   OR "ownerId" IN (SELECT id FROM "Employee" WHERE name = 'علي اركان حسين')
			      AND "planId" IN (SELECT id FROM "OrgPlan" WHERE period = '2026-11'));
		`,
		},
		{
			// (دازاي 11-19) «البشر ميكدرون يقيموني لان شغلي برمجة… تأكد انت وأشّرهن»:
			// 💻 شغله برّا النظام + 🤖 تقييم ماتركس بس.
			Version: "0513_data_programmer_flags",
			SQL: `
			UPDATE "Employee" SET "offSystemWork" = true, "aiOnlyRating" = true
			WHERE name = 'مصطفى محمد يونس' AND status = 'ACTIVE';
		`,
		},
		{
			// (دازاي 11-19) «ما اتحدد بجدول دوام، أداوم براحتي — المفروض ما أتحاسب ع دوام»
			Version: "0514_data_programmer_flexible",
			SQL: `
			UPDATE "Employee" SET "flexibleHours" = true WHERE name = 'مصطفى محمد يونس' AND status = 'ACTIVE' AND NOT "flexibleHours";
		`,
		},
		{
			Version: "0515_station_excuse",
			SQL: `
			CREATE TABLE IF NOT EXISTS "StationExcuse" (
	"id" text NOT NULL,
	"bookingId" text NOT NULL,
	"stationKey" text NOT NULL,
	"employeeId" text NOT NULL,
	"reason" text NOT NULL,
	"status" text DEFAULT 'PENDING'::text NOT NULL,
	"decidedById" text,
	"decidedAt" timestamp with time zone,
	"createdAt" timestamp with time zone DEFAULT now() NOT NULL
);
			ALTER TABLE "StationExcuse" ADD CONSTRAINT "StationExcuse_pkey" PRIMARY KEY (id);
			ALTER TABLE "StationExcuse" ADD CONSTRAINT "StationExcuse_bookingId_stationKey_employeeId_key" UNIQUE ("bookingId", "stationKey", "employeeId");
			ALTER TABLE "StationExcuse" ADD CONSTRAINT "StationExcuse_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"(id) ON DELETE CASCADE;
			ALTER TABLE "StationExcuse" ADD CONSTRAINT "StationExcuse_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"(id) ON DELETE CASCADE;
		`,
		},
		{
			Version: "0516_maintenance_decision",
			SQL: `
			CREATE TABLE IF NOT EXISTS "MaintenanceDecision" (
	"id" text NOT NULL,
	"customerId" text NOT NULL,
	"decision" text NOT NULL,
	"note" text,
	"hiddenUntil" timestamp with time zone NOT NULL,
	"byId" text NOT NULL,
	"createdAt" timestamp with time zone DEFAULT now() NOT NULL
);
			ALTER TABLE "MaintenanceDecision" ADD CONSTRAINT "MaintenanceDecision_pkey" PRIMARY KEY (id);
			CREATE INDEX IF NOT EXISTS "MaintenanceDecision_customer_idx" ON public."MaintenanceDecision" USING btree ("customerId", "hiddenUntil" DESC);
			ALTER TABLE "MaintenanceDecision" ADD CONSTRAINT "MaintenanceDecision_customerId_fkey" FOREIGN KEY ("customerId") REFERENCES "Customer"(id) ON DELETE CASCADE;
		`,
		},
		{
			Version: "0517_employee_idea",
			SQL: `
			CREATE TABLE IF NOT EXISTS "EmployeeIdea" (
	"id" text NOT NULL,
	"employeeId" text NOT NULL,
	"text" text NOT NULL,
	"category" text DEFAULT 'OTHER'::text NOT NULL,
	"status" text DEFAULT 'NEW'::text NOT NULL,
	"reply" text,
	"decidedById" text,
	"decidedAt" timestamp with time zone,
	"escalatedAt" timestamp with time zone,
	"createdAt" timestamp with time zone DEFAULT now() NOT NULL
);
			ALTER TABLE "EmployeeIdea" ADD CONSTRAINT "EmployeeIdea_pkey" PRIMARY KEY (id);
			CREATE INDEX IF NOT EXISTS "EmployeeIdea_status_idx" ON public."EmployeeIdea" USING btree (status, "createdAt" DESC);
			ALTER TABLE "EmployeeIdea" ADD CONSTRAINT "EmployeeIdea_employeeId_fkey" FOREIGN KEY ("employeeId") REFERENCES "Employee"(id) ON DELETE CASCADE;
		`,
		},
		{
			// (دازاي 11-21) «فواتير الجي بي اس… تترحل كبل لمسؤول الخدمة» — الخدمة الي اسمها
			// GPS تتأشّر نوعها جي بي اس وورقها على مسؤولها (إذا ما متأشّرة قبل).
			Version: "0518_data_gps_service_kind",
			SQL: `
			UPDATE "Service" SET "serviceKind" = 'GPS', "managerHandlesPaperwork" = true
				WHERE upper(trim(name)) = 'GPS' AND "serviceKind" IS NULL;
		`,
		},
		{
			// (ل) 11-22 دازاي: حزمة العرض صارت نوع واحد (10.1) — ملفات 10.2/10.3/10.4 تنضم إله.
			Version: "0519_data_plan_render_single_type",
			SQL: `
			UPDATE "PlanFile" SET
				title = CASE WHEN title = "typeCode" THEN '10.1' ELSE title END,
				"typeCode" = '10.1'
			WHERE "typeCode" IN ('10.2', '10.3', '10.4');
		`,
		},
		{
			Version: "0520_plan_sheet_extra",
			SQL: `
			ALTER TABLE "PlanFile" ADD COLUMN IF NOT EXISTS "sheetExtra" jsonb DEFAULT '[]'::jsonb NOT NULL;
		`,
		},
		// ‹makemigration› — التحديثات الجديدة تنضاف فوگ هالسطر
	}
}
