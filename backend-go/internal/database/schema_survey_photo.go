package database

// صور الكشف — للحجز (نوع SURVEY) وللمشروع (مرحلة الكشف) بنفس الجدول.
//
// الصفّ مربوط بواحد بس منهم (CHECK)، حتى ما تضيع صورة بلا صاحب ولا
// تنعرض بمكانين. التنزيل يمر من مسار محمي بالدور، مو من رابط
// /api/files العام.
func surveyPhotoMigration() []Migration {
	return []Migration{
		{
			Version: "0296_survey_photo",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SurveyPhoto" (
					id             TEXT PRIMARY KEY,
					"bookingId"    TEXT REFERENCES "Booking"(id) ON DELETE CASCADE,
					"projectId"    TEXT REFERENCES "Project"(id) ON DELETE CASCADE,
					"fileKey"      TEXT NOT NULL,
					"contentType"  TEXT NOT NULL,
					"fileName"     TEXT NOT NULL DEFAULT '',
					"uploadedById" TEXT NOT NULL REFERENCES "Employee"(id),
					"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT now(),
					CONSTRAINT "SurveyPhoto_one_owner" CHECK (("bookingId" IS NULL) <> ("projectId" IS NULL))
				);
				CREATE INDEX IF NOT EXISTS "SurveyPhoto_bookingId_idx" ON "SurveyPhoto" ("bookingId");
				CREATE INDEX IF NOT EXISTS "SurveyPhoto_projectId_idx" ON "SurveyPhoto" ("projectId");
			`,
		},
	}
}
