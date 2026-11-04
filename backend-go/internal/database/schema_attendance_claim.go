package database

// ═══ الانصراف التلقائي: ماتركس يسأل ويتأكد بالدليل (قرار (ع) 10-06) ═══
// «يجوز طالع من النظام بس جاي يشتغل وره اربع ساعات… لازم تفكرلي بطريقه
// قويه». بعد الانصراف التلقائي ماتركس يسأل الموظف شنو صار:
//   - ACK: خلص وطلع — الانصراف التلقائي صحيح.
//   - WORKED: چان يشتغل لحد ساعة معيّنة — ماتركس يدوّر دليل بالنظام (حجز
//     خلّصه، فاتورة، تقرير، مهمة) بعد نهاية الشفت؛ الدليل يصحّح الوقت لحاله،
//     والباقي بلا دليل يروح للمراقب يقرر.
//   - BACK: رجع يشتغل هسه — جلسة جديدة (دوام إضافي).
func attendanceClaimMigrations() []Migration {
	return []Migration{{
		Version: "0324_attendance_claim",
		SQL: `CREATE TABLE IF NOT EXISTS "AttendanceClaim" (
			id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			"attendanceId"  TEXT NOT NULL UNIQUE REFERENCES "Attendance"(id) ON DELETE CASCADE,
			"employeeId"    TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
			kind            TEXT NOT NULL,              -- ACK | WORKED | BACK
			"autoAt"        TIMESTAMPTZ NOT NULL,       -- وقت الانصراف التلقائي
			"claimedUntil"  TIMESTAMPTZ,
			"evidenceAt"    TIMESTAMPTZ,
			evidence        TEXT,
			note            TEXT,
			status          TEXT NOT NULL,              -- OK | PENDING | APPROVED | REJECTED
			"decidedById"   TEXT REFERENCES "Employee"(id),
			"decidedAt"     TIMESTAMPTZ,
			"createdAt"     TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS "AttendanceClaim_status_idx" ON "AttendanceClaim"(status, "createdAt");`,
	}}
}
