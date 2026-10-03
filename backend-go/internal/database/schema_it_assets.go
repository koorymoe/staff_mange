package database

// ═══ جرد أجهزة تقنية المعلومات ═══
//
// دور الـIT انبنى (0278) بصلاحياته وما انبنتله ولا شاشة — فالموظف
// يدخل وما يلگه شغله. هنا الأجهزة نفسها (حاسبات، مخدّمات، شبكة…)
// وسجل صيانة لكل جهاز: متى خرب، شنو انسوّى، منو سوّاه.
func itAssetMigrations() []Migration {
	return []Migration{
		{
			Version: "0297_it_assets",
			SQL: `
				CREATE TABLE IF NOT EXISTS "ItAsset" (
					id                   TEXT PRIMARY KEY,
					name                 TEXT NOT NULL,
					kind                 TEXT NOT NULL DEFAULT 'COMPUTER',
					status               TEXT NOT NULL DEFAULT 'ACTIVE',
					brand                TEXT,
					model                TEXT,
					"serialNumber"       TEXT,
					"ipAddress"          TEXT,
					location             TEXT,
					"assignedEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"purchaseDate"       DATE,
					"warrantyUntil"      DATE,
					notes                TEXT,
					"createdById"        TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"createdAt"          TIMESTAMPTZ NOT NULL DEFAULT now(),
					"updatedAt"          TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "ItAsset_kind_idx" ON "ItAsset" (kind);
				CREATE INDEX IF NOT EXISTS "ItAsset_status_idx" ON "ItAsset" (status);

				CREATE TABLE IF NOT EXISTS "ItAssetLog" (
					id           TEXT PRIMARY KEY,
					"assetId"    TEXT NOT NULL REFERENCES "ItAsset"(id) ON DELETE CASCADE,
					kind         TEXT NOT NULL DEFAULT 'NOTE',
					note         TEXT NOT NULL,
					cost         DOUBLE PRECISION,
					"employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "ItAssetLog_assetId_idx" ON "ItAssetLog" ("assetId", "createdAt" DESC);
			`,
		},
	}
}
