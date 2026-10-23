package database

// ══════════════════════════════════════════════════════════════════
// التعديل اليدوي على نقاط الانضباط
// ══════════════════════════════════════════════════════════════════
//
// نقاط الانضباط النظام يخصمها تلقائياً — وهذا مقصود حتى ما تصير محاباة.
// بس الآلة ما تعرف كل شي: الموظف ممكن يتأخر لأن الزبون ما كان بالبيت،
// أو ينغرم على شغلة مو ذنبه أصلاً. فلازم يكون بيد المالك مفتاح يصحّح.
//
// والمفتاح هذا خطر لو انفتح بلا حساب — لهذا كل تعديل يدوي ينسجّل بنفس
// سجل الحركات مع **اسم الي عدّل والسبب**، فينشاف متل أي حركة ثانية.
// byEmployeeId هو الي يخلي التعديل اليدوي مسؤولية مو باب خلفي.
func disciplineManualMigration() []Migration {
	return []Migration{
		{
			Version: "0220_discipline_event_by_employee",
			SQL: `
				ALTER TABLE "DisciplineEvent"
					ADD COLUMN IF NOT EXISTS "byEmployeeId" TEXT
						REFERENCES "Employee"(id) ON DELETE SET NULL;

				CREATE INDEX IF NOT EXISTS "DisciplineEvent_byEmployee_idx"
					ON "DisciplineEvent" ("byEmployeeId")
					WHERE "byEmployeeId" IS NOT NULL;
			`,
		},
	}
}
