package database

// «الزبون ما رد» على الشكوى — طلب (ع) 10-19: «يدوس زبون مارد ويتحول لمكان
// ثاني مو عاجل، بعدها يرجع يتصل عليه ورا فترة، ٣ مرات، إذا ما رد الشكوى
// تسقط وتصير بإحصائية إنو تواصلنا والزبون ما رد».
func complaintNoAnswerMigrations() []Migration {
	return []Migration{{
		Version: "0385_complaint_no_answer",
		SQL: `
			ALTER TABLE "Complaint" ADD COLUMN IF NOT EXISTS "noAnswerCount" INT NOT NULL DEFAULT 0;
			ALTER TABLE "Complaint" ADD COLUMN IF NOT EXISTS "noAnswerAt" TIMESTAMP;
			ALTER TABLE "Complaint" ADD COLUMN IF NOT EXISTS "nextCallAt" TIMESTAMP;
			ALTER TABLE "Complaint" ADD COLUMN IF NOT EXISTS "droppedNoAnswerAt" TIMESTAMP;
		`,
	}}
}
