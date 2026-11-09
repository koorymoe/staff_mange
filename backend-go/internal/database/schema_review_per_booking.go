package database

// ═══ التقييم يصير لكل حجز ═══
//
// كان التقييم **حر**: الليدر يفتح الشاشة، يختار موظف، ويكتب سبب.
// بلا ربط بشغل معيّن.
//
// وهاي تخلي التقييم بلا سياق: بعد أسبوع تقرا «محمد — يحتاج تدريب»
// وما تعرف بأي شغلة، ولا منو كان وياه، ولا شنو صار بالضبط. والليدر
// نفسه ينسى، فيصير التقييم انطباع عام مو ملاحظة على شغل حقيقي.
//
// «الليدر يكدر يقيّم فريقه لكل حجز يطلعوله، مو يقيمه مرة وحدة
// باليوم».
//
// هسه كل تقييم مربوط بحجز. ومنه:
//   - الملاحظة إلها سياق: أي زبون، أي خدمة، أي يوم
//   - نعرف منو ما انقيّم بأي حجز
//   - ونمنع تقييمين لنفس الموظف بنفس الحجز
func reviewPerBookingMigrations() []Migration {
	return []Migration{
		{
			Version: "0243_review_per_booking",
			SQL: `
				ALTER TABLE "PerformanceReview"
					ADD COLUMN IF NOT EXISTS "bookingId" TEXT
					REFERENCES "Booking"(id) ON DELETE CASCADE;

				-- ⚠️ فهرس فريد **جزئي**: يمنع تقييمين لنفس الموظف بنفس
				-- الحجز، بس ما يلمس التقييمات القديمة الي بلا حجز
				-- (bookingId فاضي) — هذني تاريخ انسجّل قبل التغيير وما
				-- يصير نخسره ولا نمنع تكراره بأثر رجعي.
				CREATE UNIQUE INDEX IF NOT EXISTS "PerformanceReview_booking_employee_uniq"
					ON "PerformanceReview" ("bookingId", "employeeId")
					WHERE "bookingId" IS NOT NULL;

				CREATE INDEX IF NOT EXISTS "PerformanceReview_booking_idx"
					ON "PerformanceReview" ("bookingId");
			`,
		},
	}
}
