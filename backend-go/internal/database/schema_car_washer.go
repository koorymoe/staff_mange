package database

// ═══ دور «عامل الغسل» + سجل الغسل (قرار (ع) 10-09) ═══
// عامل الغسل يفتح النظام، يلگه السيارات، والي يغسلها يضغط عليها «تم الغسل».
// بعدها أبو الكميات يقيّم السيارات بمتابعة السيارات ويشوف منو غسلها.
// سطر واحد لكل سيارة باليوم (يوم بغداد).
func carWasherMigrations() []Migration {
	return []Migration{
		{Version: "0338_employee_role_car_washer", SQL: `ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'CAR_WASHER';`},
		{Version: "0339_vehicle_wash_log", SQL: `
			CREATE TABLE IF NOT EXISTS "VehicleWashLog" (
				id TEXT PRIMARY KEY,
				"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
				"employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"washDate" DATE NOT NULL,
				"washedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE UNIQUE INDEX IF NOT EXISTS "VehicleWashLog_vehicle_day" ON "VehicleWashLog"("vehicleId", "washDate");
		`},
	}
}
