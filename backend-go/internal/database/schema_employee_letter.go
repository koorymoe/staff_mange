package database

// ══════════════════════════════════════════════════════════════════
// الطلبات — كتاب رسمي من الموظف للإدارة
// ══════════════════════════════════════════════════════════════════
//
// الموظف الي يريد شي من الإدارة (سلفة، نقل، شكوى، اقتراح، إجازة
// استثنائية) يفتح الوورد بالموبايل أو يكتب بورقة ويوصلها بيده. وبعدها:
// الورقة تضيع، أو المدير ينساها، أو الموظف ما يعرف إذا وصلت أصلاً.
// وماكو سجل يجاوب «شكد طلب سلفة انطلب هالسنة؟».
//
// هذا الجدول يخلي الطلب:
//   - ينكتب بالنظام بنفس صيغة الكتاب الرسمي (إلى السيد المدير...)
//   - يوصل الإدارة فوراً بإشعار
//   - ينطبع بورقة الشركة الرسمية (نفس إطار عرض السعر)
//   - يبقى بالسجل بجوابه — منو رد ومتى وشنو كال
//
// وحالة الطلب مو «مقروء/غير مقروء» — إما موافقة أو رفض بجواب مكتوب.
// الطلب الي يبقى معلق للأبد أسوأ من الرفض.
func employeeLetterMigration() []Migration {
	return []Migration{
		{
			Version: "0231_employee_letter",
			SQL: `
				CREATE TABLE IF NOT EXISTS "EmployeeLetter" (
					id            TEXT PRIMARY KEY,
					"employeeId"  TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
					-- إلى منو: المدير، المالك، الموارد البشرية...
					"addressedTo" TEXT NOT NULL DEFAULT 'السيد المدير المحترم',
					subject       TEXT NOT NULL,
					body          TEXT NOT NULL,
					-- PENDING | APPROVED | REJECTED
					status        TEXT NOT NULL DEFAULT 'PENDING',
					"decisionNote" TEXT,
					"decidedById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"decidedAt"    TIMESTAMPTZ,
					"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				CREATE INDEX IF NOT EXISTS "EmployeeLetter_employee_idx"
					ON "EmployeeLetter" ("employeeId", "createdAt" DESC);
				CREATE INDEX IF NOT EXISTS "EmployeeLetter_pending_idx"
					ON "EmployeeLetter" (status) WHERE status = 'PENDING';
			`,
		},
	}
}
