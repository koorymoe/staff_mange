package database

// ═══ تقييم مهارات القيادة ═══
//
// ⚠️⚠️ **الأشرطة الخمسة چانت كذباً كاملاً.** بشاشة إدارة الكوادر
// اكو خمس أشرطة (القيادة · إدارة الفريق · حل المشكلات · التواصل ·
// اتخاذ القرار) يسحبها المدير ويشوف الرقم يتغيّر — وهي
// `useState` محلية **ما تنحفظ ولا تنجلب أبداً**.
//
// وأسوأ: تصفير الحالة عند تبديل الموظف **ما چان يشملها**، فالمدير
// يقيّم موظفاً، يضغط على الي بعده، **ويشوف نفس درجاته عليه**.
// يعني مو بس ما تنحفظ — تعرض معلومة غلط عن موظف ثاني.
//
// ⚠️ **`CHECK` على ١..١٠**: الدرجة تدخل بمتوسط يتحاسب عليه الموظف،
// ورقم برّا المدى يسمّمه. نفس مبدأ تقييم الزبون بالشكاوى.
//
// ⚠️⚠️ **مفتاح فريد على (employeeId, skill)**: التقييم **حالة** مو
// سجل أحداث — «كم درجته بالقيادة الآن؟» سؤال له جواب واحد. بدون
// القيد كل سحبة شريط تضيف صفاً، وبعد شهر يصير عندنا مئة صف لنفس
// المهارة وماكو طريقة نعرف أيّهن الحالي.
//
// ⚠️ **`ratedById` بـSET NULL مو CASCADE**: حذف الموظف الي قيّم ما
// يمحي **إن التقييم صار**. الدرجة قرار إداري، وهوية صاحبه تنفقد
// أما القرار فيبقى. أما `employeeId` فـCASCADE: تقييم موظف محذوف
// ماكو إله معنى.
func leaderSkillMigrations() []Migration {
	return []Migration{
		{
			Version: "0262_leader_skill_rating",
			SQL: `
				CREATE TABLE IF NOT EXISTS "LeaderSkillRating" (
					id           TEXT PRIMARY KEY,
					"employeeId" TEXT NOT NULL
						REFERENCES "Employee"(id) ON DELETE CASCADE,
					skill        TEXT NOT NULL,
					score        INT  NOT NULL CHECK (score BETWEEN 1 AND 10),
					"ratedById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"ratedAt"    TIMESTAMP NOT NULL DEFAULT now()
				);

				CREATE UNIQUE INDEX IF NOT EXISTS "LeaderSkillRating_employee_skill_key"
					ON "LeaderSkillRating" ("employeeId", skill);
			`,
		},
	}
}
