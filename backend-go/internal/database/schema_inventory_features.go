package database

// inventoryFeaturesVersionedMigrations يرجّع الترحيلات المرقّمة لميزات المخزون
// المضافة بهذه الجلسة: (1) "العدة القياسية" (PersonalToolTemplateItem) — قائمة
// رئيسية بأسماء الأدوات الشخصية الي كل موظف لازم يكون عنده إياها، تُطبَّق تلقائياً
// على الموظفين الحاليين والجدد، و(2) VehicleToolCheck — لقطة الأدوات العامة
// الناقصة بالمركبة عند بدء مهمة من قبل ليدر.
func inventoryFeaturesVersionedMigrations() []Migration {
	return []Migration{
		{
			Version: "0133_create_personal_tool_template_item",
			SQL: `CREATE TABLE IF NOT EXISTS "PersonalToolTemplateItem" (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		{
			Version: "0134_create_vehicle_tool_check",
			SQL: `CREATE TABLE IF NOT EXISTS "VehicleToolCheck" (
				id TEXT PRIMARY KEY,
				"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
				"missionId" TEXT NOT NULL REFERENCES "VehicleMission"(id) ON DELETE CASCADE,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				"missingToolNames" TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "VehicleToolCheck_vehicleId_idx" ON "VehicleToolCheck"("vehicleId");
			CREATE INDEX IF NOT EXISTS "VehicleToolCheck_missionId_idx" ON "VehicleToolCheck"("missionId")`,
		},
		{
			// رسالة بيانات دخول الزبون (يوزر + باسورد بنص حر) — لازم تنكتب قبل
			// ما يكدر الإداري يفعّل الجهاز، مع تاريخ التفعيل.
			Version: "0135_add_gps_device_request_credentials_message",
			SQL:     `ALTER TABLE "GpsDeviceRequest" ADD COLUMN IF NOT EXISTS "credentialsMessage" TEXT`,
		},
		{
			// تتبع آخر موظف عدّل تفاصيل/تكليف الحجز (منفصل عن "من أكّد الحجز") —
			// حتى يطلع بصفحة تفاصيل الحجز "من عدّله" بعد "من أكّده".
			Version: "0136_add_booking_last_edited_by",
			SQL: `ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "lastEditedById" TEXT REFERENCES "Employee"(id);
				ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "lastEditedAt" TIMESTAMP`,
		},
		{
			// الكشوفات: فورمات فارغة يطبعها المهندس، يمليها يدوياً بالموقع، وبعدين
			// يرجع يرفع صور الفورمة المالية للنظام — مربوطة بمشروع اختياري.
			Version: "0137_create_project_checklist",
			SQL: `CREATE TABLE IF NOT EXISTS "ProjectChecklist" (
				id TEXT PRIMARY KEY,
				"projectId" TEXT REFERENCES "Project"(id),
				title TEXT NOT NULL,
				"createdById" TEXT NOT NULL REFERENCES "Employee"(id),
				"photoUrls" TEXT[] NOT NULL DEFAULT '{}',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "ProjectChecklist_projectId_idx" ON "ProjectChecklist"("projectId")`,
		},
	}
}
