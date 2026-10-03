package database

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// seedEngineeringSkills يزرع خدمة "الهندسة" ومهاراتها الأربع مرة وحدة (idempotent) —
// نفس هالمهارات تنشرط بالموظف قبل ما يوصل دور "مهندس".
func seedEngineeringSkills(db *sqlx.DB) error {
	if _, err := db.Exec(`
		INSERT INTO "Service" (id, name, category)
		VALUES ('svc_engineering', 'الهندسة', 'إدارية')
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		return err
	}
	for _, name := range model.EngineeringSkillNames {
		if _, err := db.Exec(`
			INSERT INTO "Skill" (id, name, "serviceId")
			VALUES ('sk_eng_' || $1, $1, 'svc_engineering')
			ON CONFLICT (id) DO NOTHING
		`, name); err != nil {
			return err
		}
	}
	return nil
}

// legacySkillsByService هي نفس مصفوفة المهارات التفصيلية الي كانت موجودة بالنظام
// القديم (قبل الترحيل لهذا النظام) — كل خدمة وتحتها المهارات الدقيقة الخاصة بيها،
// بالضبط متل ما كانت مقسمة بملف تتبع مهارات الكوادر القديم.
var legacySkillsByService = map[string][]string{
	"كاميرات انلوك":       {"تسليك كيبل انلوك", "تثبيت كاميرات", "اعدادت DVR", "صيانة كاميرات و كشف الصيانة"},
	"كاميرات IP":          {"اعدادت NVR IP", "ربط الكاميرات بالهاتف"},
	"كاميرات تخصصية":      {"تثبيت كاميرات سيارة", "تثبيت كاميرات الدراجة", "شد كامرات واي فاي", "اعداد كامرات واي فاي المتخصصة"},
	"الاقفال الالكترونية": {"برمجة قفل باب", "تسليك قفل باب", "تثبيت قفل باب"},
	"GPS":                 {"تثبيت GPS دراجة", "تثبيت GPS سيارة فيوز", "تثبيت GPS ستوتة", "تثبيت GPS تكتك"},
	"شبكات": {
		"تسليك كيبل لان", "تسليك كيبل ضوئي", "الراوترات والشبكات", "rack cable managment",
		"تثبيت اليونيفايات", "خدمة تصميم الشبكات حسب الموقع و الحاجة السلكية و اللاسلكية",
		"تنفيذ شبكات للمؤسسات Wi-fi و الفنادق و المطاعم و المولات", "انترنيت المنازل (تنصيب و صيانة)",
		"خدمة ربط المواقع بإستخدام VPN", "تنصيب و ربط الطابعات الشبكية",
		"ادارة الحسابات (خدمة active directory)", "خدمة Hotspot لشبكات الواي فاي",
		"تنصيب برامج مراقبة الشبكة و ادارة الشبكة عن بعد", "خدمة تصميم و تنفيذ ATS",
	},
	"بيوت ذكية":          {"اعداد المنزل الذكي", "تنصيب المنزل الذكي"},
	"انذار حريق":         {"تثبيت حساسات منظومات الحريق", "اعداد منظومات الحريق", "تسليك منظومات الحريق", "تثبيت الحساسات"},
	"انظمة الصوت":        {"تثييت سماعات واجهزة الصوت", "اعداد انظمة الصوت"},
	"التليفونات البدالة": {"اعداد تلفون بدالة", "تثبيت تليفون بدالة", "تسليك كيبل تلفون بدالة", "التليفونات البدالة IP"},
	"شاشات":              {"شد شاشة", "شد داتا شو"},
	"ستلايت":             {"صحن ستلايت", "اعداد ستلايت"},
	"كشوفات المشاريع":    {"كشوفات المشاريع"},
	"الحاكيات":           {"اعداد حاكيات الباب"},
	"طاقة الشمسية":       {"صيانة الطاقة الشمسية", "برمجة انفيرتر الواح الطاقة", "تثبيت الواح الطاقة الشمسية", "تسليك وتفييش MC4"},
	"اجهزة البصمة":       {"تصنيب و اعداد اجهزة البصمة", "برمجة اجهزة البصمة و ربط بالشبكة"},
	"انذار سرقة":         {"برمجة نظام انذار السرقة"},
	"اخر":                {"تسليك كيبل كهرباء", "القيادة"},
}

// seedLegacySkills يرجّع مهارات النظام القديم (قبل الترحيل) — لأي خدمة بهذي
// القائمة مو موجودة يسويها، ولأي مهارة تفصيلية بيها مو موجودة يضيفها. idempotent
// بالكامل (ON CONFLICT DO NOTHING بالخدمة والمهارة) — تكرار التشغيل ما يكرر شي.
func seedLegacySkills(db *sqlx.DB) error {
	for serviceName, skills := range legacySkillsByService {
		var serviceID string
		err := db.Get(&serviceID, `
			INSERT INTO "Service" (id, name)
			VALUES (gen_random_uuid()::text, $1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, serviceName)
		if err != nil {
			return err
		}
		for _, skillName := range skills {
			if _, err := db.Exec(`
				INSERT INTO "Skill" (id, name, "serviceId")
				VALUES (gen_random_uuid()::text, $1, $2)
				ON CONFLICT ("serviceId", name) DO NOTHING
			`, skillName, serviceID); err != nil {
				return err
			}
		}
	}
	return nil
}

// seedDefaultSkillForServices تضمن أي خدمة موجودة بدون أي مهارة محددة إلها تاخذ
// مهارة افتراضية وحدة بنفس اسم الخدمة — حتى تصير مطابقة الفني بالحجز (matching)
// شغالة فوراً من غير ما ننتظر الإداري يضيف مهارات تفصيلية لكل خدمة يدوياً.
// idempotent: ما تلمس أي خدمة عندها مهارات أصلاً (حتى لو مهارة وحدة بس).
func seedDefaultSkillForServices(db *sqlx.DB) error {
	type svcRow struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	var services []svcRow
	if err := db.Select(&services, `
		SELECT s.id, s.name FROM "Service" s
		WHERE NOT EXISTS (SELECT 1 FROM "Skill" sk WHERE sk."serviceId" = s.id)
	`); err != nil {
		return err
	}
	for _, s := range services {
		if _, err := db.Exec(`
			INSERT INTO "Skill" (id, name, "serviceId")
			VALUES (gen_random_uuid()::text, $1, $2)
			ON CONFLICT ("serviceId", name) DO NOTHING
		`, s.Name, s.ID); err != nil {
			return err
		}
	}
	return nil
}

// migrateGpsEngineersToSkill يحول أي موظف لسه بدور GPS_ENGINEER (القديم) إلى TECHNICIAN
// عادي، ويمنحه كل مهارات تركيب GPS تلقائياً حتى يضل يقدر يستلم نفس شغله بالضبط.
func migrateGpsEngineersToSkill(db *sqlx.DB) error {
	var ids []string
	if err := db.Select(&ids, `SELECT id FROM "Employee" WHERE role = 'GPS_ENGINEER'`); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	var gpsSkillIDs []string
	if err := db.Select(&gpsSkillIDs, `
		SELECT sk.id FROM "Skill" sk JOIN "Service" sv ON sv.id = sk."serviceId" WHERE sv.name = 'GPS'
	`); err != nil {
		return err
	}

	for _, empID := range ids {
		if _, err := db.Exec(`UPDATE "Employee" SET role = 'TECHNICIAN' WHERE id = $1`, empID); err != nil {
			return err
		}
		for _, skillID := range gpsSkillIDs {
			if _, err := db.Exec(`
				INSERT INTO "EmployeeSkill" (id, "employeeId", "skillId", "canPerform")
				VALUES (gen_random_uuid()::text, $1, $2, true)
				ON CONFLICT ("employeeId", "skillId") DO UPDATE SET "canPerform" = true
			`, empID, skillID); err != nil {
				return err
			}
		}
	}
	return nil
}
