package database

// ═══ نوع الإجازة ═══
//
// طلب الإجازة كان يحمل تواريخ وسبب مكتوب بس — بلا نوع. يعني الإجازة
// المرضية والاعتيادية والطارئة كلهن نفس الشي بنظر النظام.
//
// ليش هذا مهم إدارياً: الرصيد الشهري للإجازات ما ينطبق على المرضية
// (الموظف المريض ما ينحاسب على رصيده)، والطارئة إلها أولوية مختلفة
// بالموافقة. بلا النوع، المدير لازم يقرا السبب المكتوب ويخمّن — وأكو
// موظفين ما يكتبون سبب أصلاً.
//
// ⚠️ الافتراضي REGULAR للصفوف القديمة: كل الطلبات الي انقدّمت قبل هذا
// التغيير ما نعرف نوعها فعلاً، بس «اعتيادية» هي الغالب ومو ادّعاء
// خطير — بعكس ما لو حطّينا «مرضية» على طلبات ما كانت مرضية.
func leaveKindMigrations() []Migration {
	return []Migration{
		{
			Version: "0240_leave_kind",
			SQL: `
				ALTER TABLE "LeaveRequest"
					ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'REGULAR';

				-- الفحص بقاعدة البيانات مو بالكود بس: قيمة غريبة تنكتب
				-- بأمر SQL مباشر تخرب كل شاشة تعرض التسمية العربية.
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'leave_request_kind_valid'
					) THEN
						ALTER TABLE "LeaveRequest"
							ADD CONSTRAINT leave_request_kind_valid
							CHECK (kind IN ('REGULAR','SICK','URGENT','UNPAID'));
					END IF;
				END $$;
			`,
		},
	}
}
