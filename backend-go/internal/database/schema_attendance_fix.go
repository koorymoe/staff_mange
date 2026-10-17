package database

// ═══ تصليحات فحص الحضور (10-07) ═══
//   - جلسة مفتوحة وحدة بس لكل موظف (ضغطتين سوه چانت تفتح جلستين).
//   - «approvedOut»: انصراف انصحّح بدليل أو بموافقة المراقب أو بتصحيح يدوي —
//     ينحسب مثل ما هو وما ينقطع بحد الساعات.
func attendanceFixMigrations() []Migration {
	return []Migration{{
		Version: "0328_attendance_fix",
		SQL: `
			ALTER TABLE "Attendance" ADD COLUMN IF NOT EXISTS "approvedOut" BOOLEAN NOT NULL DEFAULT false;

			-- إذا اكو جلستين مفتوحة لنفس الموظف: الأقدم تنسكر بنفس وقت حضورها.
			UPDATE "Attendance" a SET "checkOut" = a."checkIn"
			WHERE a."checkOut" IS NULL AND EXISTS (
				SELECT 1 FROM "Attendance" b WHERE b."employeeId" = a."employeeId" AND b."checkOut" IS NULL
				AND (b."checkIn" > a."checkIn" OR (b."checkIn" = a."checkIn" AND b.id > a.id)));
			CREATE UNIQUE INDEX IF NOT EXISTS "Attendance_one_open_idx" ON "Attendance" ("employeeId") WHERE "checkOut" IS NULL;

			CREATE OR REPLACE FUNCTION att_counted_end(emp TEXT, ci TIMESTAMP, co TIMESTAMP)
			RETURNS TIMESTAMP LANGUAGE sql STABLE AS $$
				WITH c AS (
					SELECT date_trunc('day', ci + interval '3 hours') - interval '3 hours'
					       + CASE WHEN extract(hour FROM ci + interval '3 hours') < 6 THEN interval '0'
					              WHEN extract(hour FROM ci + interval '3 hours') < 15 THEN interval '16 hours'
					              ELSE interval '24 hours' END AS cap,
					       NOW() AT TIME ZONE 'UTC' AS nw,
					       EXISTS (SELECT 1 FROM "Attendance" x WHERE x."employeeId" = emp AND x."checkIn" = ci
					               AND x."approvedOut") AS approved
				), l AS (
					SELECT GREATEST(c.cap, COALESCE(att_last_activity(emp, ci, co), c.cap), ci) AS lim, c.nw, c.approved FROM c
				)
				SELECT CASE
					WHEN co IS NOT NULL AND l.approved THEN GREATEST(co, ci)
					WHEN COALESCE(co, l.nw) <= l.lim THEN GREATEST(COALESCE(co, l.nw), ci)
					WHEN co IS NULL AND l.nw < l.lim + interval '3 hours' THEN l.nw
					ELSE l.lim END
				FROM l
			$$;
		`,
	}}
}
