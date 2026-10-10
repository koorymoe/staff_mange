package database

// ══════════════════════════════════════════════════════════════════
// معرض التصاميم
// ══════════════════════════════════════════════════════════════════
//
// «معرض تصاميم مستقل» — قراره. المصممة ترفع تصاميمها وتنعرض بمعرض
// **ما ينربط بحجز**، لأن قراره الصريح إنها **ما تشوف الحجوزات**.
//
// ⚠️ **`fileKey` مفتاح مو ملف**: الصورة تنخزن بـ`storage.Store` مثل
// صور الموظفين والوصولات. حشر الصور بقاعدة البيانات ينفخها بلا
// فايدة ويثقّل كل نسخة احتياطية.
//
// ⚠️ **`ON DELETE SET NULL` + `uploadedByName` منسوخ نصاً**: حذف
// الموظفة ما يمحي تصاميمها ولا يخلّي السطر مجهولاً.
//
// ⚠️ **`archivedAt` مو حذف فعلي**: التصميم ينزاح من المعرض ويبقى
// محفوظاً — نفس مبدأ أرشفة الحجوزات بكل النظام.
func designAssetMigrations() []Migration {
	return []Migration{
		{
			Version: "0271_design_asset",
			SQL: `CREATE TABLE IF NOT EXISTS "DesignAsset" (
			        id TEXT PRIMARY KEY,
			        title TEXT NOT NULL,
			        category TEXT NOT NULL,
			        notes TEXT,
			        "fileKey" TEXT NOT NULL,
			        "fileType" TEXT,
			        "uploadedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
			        "uploadedByName" TEXT NOT NULL DEFAULT '',
			        "createdAt" TIMESTAMP NOT NULL DEFAULT now(),
			        "archivedAt" TIMESTAMP
			      )`,
		},
		{
			// المعرض يُقرا دائماً «الأحدث أول وغير المؤرشف» — فهرس
			// جزئي يخدم هذا بالضبط.
			Version: "0271_design_asset_idx",
			SQL: `CREATE INDEX IF NOT EXISTS "DesignAsset_live_idx"
			      ON "DesignAsset"("createdAt" DESC) WHERE "archivedAt" IS NULL`,
		},
	}
}
