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
	}}
}
