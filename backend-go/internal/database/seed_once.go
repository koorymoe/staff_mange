package database

import "github.com/jmoiron/sqlx"

// ═══ البذور مرة وحدة (دازاي 11-15) ═══
// «لازم كلشي يتسجل… ميصير بكل تشغيل يرجع الي شلته»: چانت أسعار المنظومات والمواد تنكتب
// فوك تعديل المدير بكل تشغيل للسيرفر، ومعايير التقييم والمهارات المحذوفة ترجع. هسه كل بذرة
// تنزرع مرة وحدة وتنسجل بـ"SeedOnce"، وبعدها السيرفر ما يلمسها.

// legacySeeds البذور الي چانت تنعاد بكل تشغيل. بسيرفر شغّال أصلاً (بيه موظفين) تنعلّم
// «انزرعت» بدون ما تنعاد — حتى أول تشغيل بعد التحديث ما يرجّع المحذوف ولا يغيّر الأسعار.
var legacySeeds = []string{"engineering_skills", "legacy_skills", "decor_skills", "solar_skills", "kpi_criteria", "system_price_catalog", "material_archive"}

func adoptLegacySeeds(db *sqlx.DB) error {
	var n int
	if err := db.Get(&n, `SELECT count(*) FROM "SeedOnce"`); err != nil || n > 0 {
		return err
	}
	var existing bool
	if err := db.Get(&existing, `SELECT EXISTS (SELECT 1 FROM "Employee")`); err != nil || !existing {
		return err
	}
	for _, k := range legacySeeds {
		if _, err := db.Exec(`INSERT INTO "SeedOnce"(key) VALUES ($1) ON CONFLICT DO NOTHING`, k); err != nil {
			return err
		}
	}
	return nil
}

// seedOnce يشغّل البذرة إذا ما انزرعت قبل، ويسجّلها.
func seedOnce(db *sqlx.DB, key string, fn func(*sqlx.DB) error) error {
	var done bool
	if err := db.Get(&done, `SELECT EXISTS (SELECT 1 FROM "SeedOnce" WHERE key = $1)`, key); err != nil {
		return err
	}
	if done {
		return nil
	}
	if err := fn(db); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT INTO "SeedOnce"(key) VALUES ($1) ON CONFLICT DO NOTHING`, key)
	return err
}
