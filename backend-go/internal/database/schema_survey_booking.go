package database

// ══════════════════════════════════════════════════════════════════
// حجز «كشف» — زيارة معاينة بلا فاتورة ولا عُدّة
// ══════════════════════════════════════════════════════════════════
//
// الإداري مال الكوادر يأشّر الحجز «كشف» وقت ما يرسل الكادر: الكادر
// ما يحتاج ياخذ عدّة/أدوات، وما يحتاج يسوي فاتورة ليدر — بس يسلّم
// تقرير معاينة بسيط (شنو يريد الزبون، شنو المساحة، تفاصيل حرة).
//
// ⚠️ منفصل كلياً عن حالة PARTIAL الحية (`schema_partial_completion.go`):
// ذيچ حجز واحد شغّال يرجع لليوم الجاي، وهذا نوع حجز (كشف مقابل شغل
// حقيقي) يتقرر وقت التكليف ويبقى طول عمر الحجز.
func surveyBookingMigration() []Migration {
	return []Migration{
		{
			// منفصلة لحالها: ALTER TYPE ADD VALUE ما يصير تُستعمل قيمته
			// بنفس المعاملة الي أضافته (نفس قيد 0221 بالضبط).
			Version: "0290_booking_type_survey",
			SQL:     `ALTER TYPE "BookingType" ADD VALUE IF NOT EXISTS 'SURVEY'`,
		},
		{
			Version: "0291_booking_survey_report",
			SQL: `
				CREATE TABLE IF NOT EXISTS "BookingSurveyReport" (
					id              TEXT PRIMARY KEY,
					"bookingId"     TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
					"employeeId"    TEXT NOT NULL REFERENCES "Employee"(id),
					"customerWants" TEXT NOT NULL,
					"siteAreaSqm"   DOUBLE PRECISION,
					"siteDetails"   TEXT,
					"otherNotes"    TEXT,
					"createdAt"     TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "BookingSurveyReport_bookingId_idx"
					ON "BookingSurveyReport" ("bookingId");
			`,
		},
	}
}
