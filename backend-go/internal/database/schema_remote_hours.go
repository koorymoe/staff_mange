package database

// ساعات العمل من البيت — المسؤول يضيفها يومياً، والنظام يجمعها بالثانية.
func remoteHoursMigrations() []Migration {
	return []Migration{{
		Version: "0306_remote_work_entry",
		SQL: `
			CREATE TABLE IF NOT EXISTS "RemoteWorkEntry" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"workDate" DATE NOT NULL,
				seconds INT NOT NULL CHECK (seconds > 0),
				note TEXT,
				"addedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "RemoteWorkEntry_emp_date_idx" ON "RemoteWorkEntry"("employeeId", "workDate");
		`,
	}}
}
