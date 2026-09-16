package database

import "github.com/jmoiron/sqlx"

// decorSkillsByProfession هي كتالوج مهن الديكور السبعة وفروعها بالضبط متل ما
// راجعها ووافق عليها صاحب العمل — كل مهنة (profession) تصير Service بـ
// division='DECOR'، وكل فرع (branch) تصير Skill تحتها. نفس نمط legacySkillsByService
// بالضبط بس لشعبة الديكور.
var decorSkillsByProfession = map[string][]string{
	"حدادة": {
		"أبواب وشبابيك", "أبواب متحركة", "سقوف عازلة", "هياكل حديد",
		"دعامة طاقة شمسية", "درابزين وسلالم حديد",
	},
	"نجارة": {
		"أبواب وشبابيك خشب", "مطابخ", "غرف نوم ودولاب", "ديكورات جبس خشبية", "أرضيات باركيه",
	},
	"صباغة": {
		"صبغ داخلي", "صبغ خارجي", "ديكورات ورشات فنية", "عزل حراري بالصبغ",
	},
	"تطبيق سيراميك": {
		"أرضيات", "حوائط وجدران", "حمامات ومطابخ", "واجهات خارجية",
	},
	"لبخ": {
		"لبخ حوائط داخلية", "لبخ واجهات خارجية", "عزل رطوبة قبل اللبخ",
	},
	"تأسيس ماء ومجاري": {
		"تمديد مواسير ماء", "تمديد مجاري وصرف صحي", "تركيب سخانات وخزانات", "صيانة تسريبات",
	},
	"جبس بورد": {
		"أسقف معلقة", "حوائط قواطع", "ديكورات إضاءة مخفية", "عزل صوتي",
	},
}

// seedDecorSkills يزرع مهن الديكور السبعة وفروعها (idempotent بالكامل، نفس نمط
// seedLegacySkills) — كل مهنة Service بـ division='DECOR"، وكل فرع Skill تحتها.
// إذا كانت خدمة بنفس اسم إحدى المهن السبعة موجودة أصلاً (نادر جداً) يتحدّث
// عمود division لها فقط. ما تلمس أي خدمة هندسية موجودة أصلاً لأن أسماء المهن
// السبعة مختلفة تماماً عن أي اسم خدمة هندسية حالي.
func seedDecorSkills(db *sqlx.DB) error {
	for profession, branches := range decorSkillsByProfession {
		var serviceID string
		err := db.Get(&serviceID, `
			INSERT INTO "Service" (id, name, division)
			VALUES (gen_random_uuid()::text, $1, 'DECOR')
			ON CONFLICT (name) DO UPDATE SET division = 'DECOR'
			RETURNING id
		`, profession)
		if err != nil {
			return err
		}
		for _, branch := range branches {
			if _, err := db.Exec(`
				INSERT INTO "Skill" (id, name, "serviceId")
				VALUES (gen_random_uuid()::text, $1, $2)
				ON CONFLICT ("serviceId", name) DO NOTHING
			`, branch, serviceID); err != nil {
				return err
			}
		}
	}
	return nil
}
