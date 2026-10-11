package database

// ═══ دردشة ماتركس (قرار (ع) 10-10) ═══
// المالك والمدير يسولفون ويا ماتركس (هايكو). الأسماء تنحفظ حقيقية هنا،
// وتتبدّل برموز بس وقت ما تطلع للنموذج.
func matrixChatMigrations() []Migration {
	return []Migration{{
		Version: "0344_matrix_talk",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixTalk" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				title TEXT NOT NULL DEFAULT 'محادثة جديدة',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "MatrixTalk_employee" ON "MatrixTalk"("employeeId", "updatedAt" DESC);
			CREATE TABLE IF NOT EXISTS "MatrixTalkMessage" (
				id TEXT PRIMARY KEY,
				"chatId" TEXT NOT NULL REFERENCES "MatrixTalk"(id) ON DELETE CASCADE,
				role TEXT NOT NULL CHECK (role IN ('USER', 'ASSISTANT')),
				text TEXT NOT NULL,
				steps TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "MatrixTalkMessage_chat" ON "MatrixTalkMessage"("chatId", "createdAt");
		`,
	}, {
		// قرار (ع) 10-10: فواتير الشغل داخل الشركة تنعتمد بلا رقم فاتورة محاسبية
		Version: "0345_internal_invoice_no_number",
		SQL: `
			ALTER TABLE "LeaderInvoice" DROP CONSTRAINT IF EXISTS leader_invoice_approved_needs_number;
			ALTER TABLE "LeaderInvoice" ADD CONSTRAINT leader_invoice_approved_needs_number CHECK (
				status <> 'APPROVED'
				OR ("externalInvoiceNumber" IS NOT NULL AND btrim("externalInvoiceNumber") <> '')
				OR "isFree" OR COALESCE("auditVerdict", '') = 'FREE'
				OR systems::text LIKE '%شغل داخل الشركة%'
			) NOT VALID;
		`,
	}, {
		// شكوى (ع) 10-10: حجوزات انثبّتت وهي بالانتظار وبقى بيها waitingSince،
		// فعالقة بـ«بانتظار موافقة الزبون/ما رد» ومخفية عن الحجوزات
		Version: "0346_clear_stale_waiting",
		SQL: `
			UPDATE "Booking" SET "waitingSince" = NULL, "waitingNote" = NULL, "waitingById" = NULL,
				"waitingKind" = NULL, "lastWaitingReminderAt" = NULL, "waitingReminderCount" = 0
			WHERE "waitingSince" IS NOT NULL AND status NOT IN ('WAITING', 'CANCELLED');
		`,
	}, {
		// (ع) 10-10 «النظام صار ثكيل»: فهارس للاستعلامات الي تنضرب كل دقيقة
		Version: "0347_hot_path_indexes",
		SQL: `
			CREATE INDEX IF NOT EXISTS "Notification_unread" ON "Notification"("employeeId") WHERE read = false;
			CREATE INDEX IF NOT EXISTS "Booking_waitingSince" ON "Booking"("waitingSince") WHERE "waitingSince" IS NOT NULL;
			CREATE INDEX IF NOT EXISTS "Booking_status_scheduledAt" ON "Booking"(status, "scheduledAt");
			CREATE INDEX IF NOT EXISTS "Booking_status_createdAt" ON "Booking"(status, "createdAt" DESC);
			ANALYZE "Booking";
			ANALYZE "Notification";
		`,
	}, {
		// (ع) 10-11: سكرين بالدردشة + ماتركس يتعلّم من تصحيح المالك
		Version: "0348_matrix_chat_image_lessons",
		SQL: `
			ALTER TABLE "MatrixTalkMessage" ADD COLUMN IF NOT EXISTS image TEXT;
			CREATE TABLE IF NOT EXISTS "MatrixLesson" (
				id TEXT PRIMARY KEY,
				text TEXT NOT NULL,
				context TEXT,
				"createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				active BOOLEAN NOT NULL DEFAULT true,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`,
	}, {
		// (ع) 10-11: «راقب فلان» — مهمة مراقبة بيومية وخلاصة
		Version: "0349_matrix_missions",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixMission" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"createdById" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				note TEXT,
				"startsOn" DATE NOT NULL,
				"endsOn" DATE NOT NULL,
				active BOOLEAN NOT NULL DEFAULT true,
				summary TEXT,
				"summaryAt" TIMESTAMP,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS "MatrixMissionLog" (
				id TEXT PRIMARY KEY,
				"missionId" TEXT NOT NULL REFERENCES "MatrixMission"(id) ON DELETE CASCADE,
				day DATE NOT NULL,
				text TEXT NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UNIQUE ("missionId", day)
			);
		`,
	}, {
		// (ع) 10-11: صاحب «إضافة موظف»/«تعديل كل الموظفين» يشتغل بس على
		// الأدوار الي المالك يحددها إله — بلاها يكدر يفتح حساب محاسب
		Version: "0350_employee_admin_scope",
		SQL: `
			CREATE TABLE IF NOT EXISTS "EmployeeAdminScope" (
				"employeeId" TEXT PRIMARY KEY REFERENCES "Employee"(id) ON DELETE CASCADE,
				roles TEXT[] NOT NULL DEFAULT '{}',
				"updatedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`,
	}, {
		// فحص 10-11: فواتير انسحب اعتمادها ورجعت انعتمدت كانت مختفية من
		// المجاميع (revokedAt بقى) — نرجّعها؛ ومبالغ تسوية الدوار ما تكون سالبة
		Version: "0351_reapproved_invoices_fund_nonneg",
		SQL: `
			UPDATE "LeaderInvoice" SET "revokedAt" = NULL WHERE status = 'APPROVED' AND "revokedAt" IS NOT NULL;
			ALTER TABLE "RevolvingFundTxn" DROP CONSTRAINT IF EXISTS fund_txn_nonneg;
			ALTER TABLE "RevolvingFundTxn" ADD CONSTRAINT fund_txn_nonneg
				CHECK (COALESCE("spentAmount", 0) >= 0 AND COALESCE("returnedAmount", 0) >= 0) NOT VALID;
		`,
	}}
}
