package database

// ═══ الحضور الإجباري + الانصراف التلقائي (قرار (ع) 10-06) ═══
// AttendanceAuto: جلسة سكّرها النظام لحاله لأن صاحبها ما سجّل انصراف —
// جدول منفصل حتى "Attendance" (SELECT *) ما يتغيّر.
func attendanceAutoMigrations() []Migration {
	return []Migration{{
		Version: "0322_attendance_auto",
		SQL: `CREATE TABLE IF NOT EXISTS "AttendanceAuto" (
			"attendanceId" TEXT PRIMARY KEY REFERENCES "Attendance"(id) ON DELETE CASCADE,
			"employeeId"   TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
			reason         TEXT NOT NULL,
			"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS "AttendanceAuto_emp_idx" ON "AttendanceAuto"("employeeId", "createdAt");`,
	}}
}
