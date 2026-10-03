package database

import "github.com/jmoiron/sqlx"

// ═══ مهارات نظام الطاقة الشمسية المنقول ═══
//
// نظام Solar Expert القديم جان عنده ١٤ مهارة بشيت لحاله. عدنا أربعة منهن
// موجودات أصلاً تحت خدمة «طاقة الشمسية» بأسماء مختلفة شوية:
//
//	تركيب الألواح الشمسية  ←  «تثبيت الواح الطاقة الشمسية»
//	صيانة ما بعد التركيب   ←  «صيانة الطاقة الشمسية»
//	تسليك كابلات DC/AC     ←  «تسليك وتفييش MC4»
//
// فما نضيفهن مرة ثانية — لو أضفناهن جان صار عدنا مهارتين بنفس المعنى،
// والفني الي عنده وحدة ما يطلع بفلترة الثانية، ويصير عندك كادر «ما يعرف»
// شغلة هو يشتغلها كل يوم.
//
// و«برمجة انفيرتر الواح الطاقة» عدنا وما موجودة بالنظام القديم — تضل.
//
// المهارات الجديدة تنزرع تحت نفس خدمة «طاقة الشمسية» الموجودة (مو خدمة
// جديدة)، إلا مهارات السلامة والإدارة — هذي محورها ثاني: السلامة المهنية
// وخدمة العملاء وإدارة المخزن تنفع لكل الخدمات مو للطاقة الشمسية بس، فما
// نحبسها بخدمة وحدة. تنزرع تحت خدمة «عام» ويميّزها التصنيف.

type solarSkillSeed struct {
	Name        string
	Category    string // فنية / سلامة / إدارية
	Description string
	Service     string // اسم الخدمة الي تنربط بيها
}

const (
	solarServiceName  = "طاقة الشمسية"
	commonServiceName = "عام"
)

// solarSkillSeeds المهارات الي ما موجودة عدنا من نظام Solar Expert.
// المهارات الأربع الموجودة أصلاً مو هنا عمداً — انظر شرح الملف فوق.
var solarSkillSeeds = []solarSkillSeed{
	{"صيانة البطاريات الليثيوم", "فنية", "فحص وصيانة بطاريات LiFePO4 والأنظمة الكهروكيميائية", solarServiceName},
	{"تصميم المنظومات الشمسية", "فنية", "حساب الأحمال وتصميم المنظومات المناسبة للعملاء", solarServiceName},
	{"فحص واختبار المنظومات الشمسية", "فنية", "اختبار الأداء والتأكد من سلامة التوصيلات والإعدادات", solarServiceName},
	{"أعمال الحدادة والتشكيل (Iron Work)", "فنية", "تصنيع وتركيب الهياكل الحديدية للألواح والمعدات", solarServiceName},
	{"اللحام الكهربائي", "فنية", "لحام الهياكل الحديدية والتعديلات الميكانيكية", solarServiceName},
	{"قراءة المخططات الكهربائية", "فنية", "فهم وتحليل المخططات والدائرة الكهربائية", solarServiceName},
	{"حل المشاكل التقنية (Troubleshooting)", "فنية", "اكتشاف الأعطال وإصلاحها", solarServiceName},

	{"السلامة المهنية (OSHA)", "سلامة", "إجراءات السلامة أثناء التركيب والصيانة", commonServiceName},

	{"خدمة العملاء والمبيعات", "إدارية", "التعامل مع العملاء وإغلاق الصفقات", commonServiceName},
	{"إدارة المخزن والتوريد", "إدارية", "متابعة المخزون وطلب التوريدات", commonServiceName},
	{"التدريب والإشراف", "إدارية", "تدريب الموظفين الجدد والإشراف على الفرق", commonServiceName},
}

// existingSolarSkillDescriptions المهارات الأربع الموجودة عدنا — ما نضيفها،
// بس نكمّل عليها الوصف والتصنيف الي جان بالنظام القديم حتى ما تضيع المعلومة.
var existingSolarSkillDescriptions = map[string]string{
	"تثبيت الواح الطاقة الشمسية": "تركيب وتثبيت الألواح الشمسية على مختلف أنواع الأسطح",
	"صيانة الطاقة الشمسية":       "الدعم الفني والصيانة الدورية للمنظومات بعد التركيب",
	"تسليك وتفييش MC4":           "تمديد وتوصيل كابلات DC/AC وتفييش موصلات MC4",
	"برمجة انفيرتر الواح الطاقة": "برمجة الإنفيرتر وضبط إعدادات الشحن والتفريغ",
}

// seedSolarSkills يزرع مهارات الطاقة الشمسية المنقولة. idempotent بالكامل:
// تشغيله مية مرة ما يكرر ولا مهارة ولا يدوس على وصف مكتوب بالإيد.
func seedSolarSkills(db *sqlx.DB) error {
	// نتأكد من وجود خدمة «عام» للمهارات العابرة للخدمات (سلامة/إدارية)
	for _, svc := range []string{solarServiceName, commonServiceName} {
		if _, err := db.Exec(`
			INSERT INTO "Service" (id, name)
			VALUES (gen_random_uuid()::text, $1)
			ON CONFLICT (name) DO NOTHING
		`, svc); err != nil {
			return err
		}
	}

	for _, s := range solarSkillSeeds {
		var serviceID string
		if err := db.Get(&serviceID, `SELECT id FROM "Service" WHERE name = $1`, s.Service); err != nil {
			return err
		}
		// المهارة تنعرف باسمها داخل الخدمة — لو موجودة ما نلمسها
		var exists bool
		if err := db.Get(&exists, `
			SELECT EXISTS (SELECT 1 FROM "Skill" WHERE name = $1)
		`, s.Name); err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := db.Exec(`
			INSERT INTO "Skill" (id, name, "serviceId", category, description)
			VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		`, s.Name, serviceID, s.Category, s.Description); err != nil {
			return err
		}
	}

	// الوصف ينكتب مرة وحدة بس: لو أحد كتب وصف بالإيد ما ندوس عليه
	for name, desc := range existingSolarSkillDescriptions {
		if _, err := db.Exec(`
			UPDATE "Skill" SET description = $2
			WHERE name = $1 AND (description IS NULL OR btrim(description) = '')
		`, name, desc); err != nil {
			return err
		}
	}
	return nil
}
