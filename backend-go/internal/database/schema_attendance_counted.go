package database

// ═══ الساعات المحسوبة (قرار (ع) 10-07) ═══
// نفس قاعدة CountedEnd بالخدمة: الحد حسب ساعة الحضور (بغداد +٣ ثابت) —
// قبل ٦ الصبح ما ينحسب، قبل ٣ العصر لحد ٤ العصر، وبعدها لحد ١٢ بالليل —
// ويتمدد لآخر حجز خلّصه. الأعمدة naive = UTC. السجل الأصلي ما يتغيّر.
func attendanceCountedMigrations() []Migration {
	return []Migration{{
		Version: "0327_attendance_counted",
		SQL: `
			CREATE OR REPLACE FUNCTION att_last_activity(emp TEXT, ci TIMESTAMP, co TIMESTAMP)
			RETURNS TIMESTAMP LANGUAGE sql STABLE AS $$
				SELECT max(b."completedAt") FROM "BookingAssignment" ba JOIN "Booking" b ON b.id = ba."bookingId"
				WHERE ba."employeeId" = emp AND b."completedAt" > ci
				  AND b."completedAt" <= LEAST(COALESCE(co, NOW() AT TIME ZONE 'UTC'),
				      -- نفس يوم الشغل: لحد ٦ الصبح اليوم الي بعده (بغداد).
				      date_trunc('day', ci + interval '3 hours') + interval '27 hours')
			$$;

			CREATE OR REPLACE FUNCTION att_counted_end(emp TEXT, ci TIMESTAMP, co TIMESTAMP)
			RETURNS TIMESTAMP LANGUAGE sql STABLE AS $$
				WITH c AS (
					SELECT date_trunc('day', ci + interval '3 hours') - interval '3 hours'
					       + CASE WHEN extract(hour FROM ci + interval '3 hours') < 6 THEN interval '0'
					              WHEN extract(hour FROM ci + interval '3 hours') < 15 THEN interval '16 hours'
					              ELSE interval '24 hours' END AS cap,
					       NOW() AT TIME ZONE 'UTC' AS nw
				), l AS (
					SELECT GREATEST(c.cap, COALESCE(att_last_activity(emp, ci, co), c.cap), ci) AS lim, c.nw FROM c
				)
				SELECT CASE
					WHEN COALESCE(co, l.nw) <= l.lim THEN GREATEST(COALESCE(co, l.nw), ci)
					WHEN co IS NULL AND l.nw < l.lim + interval '3 hours' THEN l.nw
					ELSE l.lim END
				FROM l
			$$;

			CREATE OR REPLACE FUNCTION att_counted_minutes(emp TEXT, ci TIMESTAMP, co TIMESTAMP)
			RETURNS DOUBLE PRECISION LANGUAGE sql STABLE AS $$
				SELECT EXTRACT(EPOCH FROM (att_counted_end(emp, ci, co) - ci)) / 60
			$$;
		`,
	}}
}
