package database

// telegramBotMigrations (دازاي 11-05): بوت تلغرام لبلاغات الموظفين عن النظام.
// مخطط بس (قرار ٠٠٠٩). الموظف يربط حسابه بزر بالنظام (رمز مرة وحدة، ١٠ دقايق)،
// والبلاغ = نص واضح + صورة وحدة على الأقل. الصور بمخزن الملفات، والجدول بيه روابطها.
func telegramBotMigrations() []Migration {
	return []Migration{{
		Version: "0442_telegram_bot",
		SQL: `
			CREATE TABLE IF NOT EXISTS "TelegramLink" (
				"employeeId" TEXT PRIMARY KEY REFERENCES "Employee"(id) ON DELETE CASCADE,
				"chatId"     BIGINT NOT NULL UNIQUE,
				"tgUsername" TEXT,
				"linkedAt"   TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE TABLE IF NOT EXISTS "TelegramLinkToken" (
				token        TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"expiresAt"  TIMESTAMP NOT NULL,
				"usedAt"     TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS "BotReportDraft" (
				"chatId"     BIGINT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				body         TEXT NOT NULL DEFAULT '',
				photos       TEXT[] NOT NULL DEFAULT '{}',
				"updatedAt"  TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE TABLE IF NOT EXISTS "BotReport" (
				id             TEXT PRIMARY KEY,
				number         SERIAL UNIQUE,
				"employeeId"   TEXT NOT NULL REFERENCES "Employee"(id),
				"chatId"       BIGINT NOT NULL,
				body           TEXT NOT NULL,
				photos         TEXT[] NOT NULL DEFAULT '{}',
				"mediaGroupId" TEXT,
				status         TEXT NOT NULL DEFAULT 'NEW',
				"adminNote"    TEXT,
				"createdAt"    TIMESTAMP NOT NULL DEFAULT now(),
				"updatedAt"    TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "BotReport_status_idx" ON "BotReport"(status, "createdAt" DESC);
			CREATE INDEX IF NOT EXISTS "BotReport_employee_idx" ON "BotReport"("employeeId", "createdAt" DESC);
		`,
	}, {
		// (دازاي 11-05) البلاغ ينكتب من النظام همين (مو بس من البوت). chatId = 0 للي من النظام.
		Version: "0443_bot_report_source",
		SQL: `
			ALTER TABLE "BotReport" ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'telegram';
			ALTER TABLE "BotReport" ADD COLUMN IF NOT EXISTS page TEXT;
		`,
	}, {
		// (دازاي 11-05) صور البلاغات تنمسح من السيرفر بعد أسبوع — نص البلاغ يبقى، وهذا العمود يأشّر متى انمسحت.
		Version: "0444_bot_report_photos_purged",
		SQL:     `ALTER TABLE "BotReport" ADD COLUMN IF NOT EXISTS "photosPurgedAt" TIMESTAMP;`,
	}}
}
