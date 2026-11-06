package database

// aiSuggestionVerifyMigrations (دازاي 11-06، بلاغ #10): من المالك يضغط «تم» على اقتراح،
// الاقتراح يصير VERIFYING، والوكيل صاحبه يتأكد إن الشي صدگ انضاف ويكتب النتيجة.
// مخطط بس (قرار ٠٠٠٩).
func aiSuggestionVerifyMigrations() []Migration {
	return []Migration{{
		Version: "0445_ai_suggestion_verify",
		SQL: `
			ALTER TABLE "AiSuggestion" DROP CONSTRAINT IF EXISTS "AiSuggestion_status_check";
			ALTER TABLE "AiSuggestion" ADD CONSTRAINT "AiSuggestion_status_check"
				CHECK (status IN ('NEW', 'ACCEPTED', 'REJECTED', 'DONE', 'VERIFYING'));
			ALTER TABLE "AiSuggestion" ADD COLUMN IF NOT EXISTS "verifyResult" TEXT;
			ALTER TABLE "AiSuggestion" ADD COLUMN IF NOT EXISTS "verifyNote" TEXT;
			ALTER TABLE "AiSuggestion" ADD COLUMN IF NOT EXISTS "verifiedAt" TIMESTAMPTZ;
		`,
	}}
}
