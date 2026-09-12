package database

// jobDurationVersionedMigrations يرجّع الترحيلات المرقّمة الخاصة بميزة "تعلّم زمن
// تنفيذ العمل": طابعا وصول/بدء العمل الحقيقيان على الحجز، وجدول عيّنات مدة العمل
// (JobDurationSample) الذي يُغذّى تلقائياً من فواتير الليدر (تركيب) وتذاكر صيانة
// الأجهزة (صيانة) — بدون أي رقم مفروض يدوياً، النظام يتعلّم من البيانات الحقيقية فقط.
func jobDurationVersionedMigrations() []Migration {
	return []Migration{
		{
			Version: "0127_add_booking_arrived_started_at",
			SQL: `ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "arrivedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "startedAt" TIMESTAMP`,
		},
		{
			Version: "0128_create_job_duration_sample",
			SQL: `CREATE TABLE IF NOT EXISTS "JobDurationSample" (
				id TEXT PRIMARY KEY,
				"systemName" TEXT NOT NULL,
				"jobType" TEXT NOT NULL CHECK ("jobType" IN ('INSTALL', 'MAINTENANCE')),
				"itemCount" INTEGER NOT NULL,
				"crewSize" INTEGER NOT NULL,
				"durationMinutes" INTEGER NOT NULL,
				"bookingId" TEXT REFERENCES "Booking"(id),
				"deviceMaintenanceTicketId" TEXT REFERENCES "DeviceMaintenanceTicket"(id),
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "JobDurationSample_system_jobtype_idx" ON "JobDurationSample"("systemName", "jobType")`,
		},
	}
}
