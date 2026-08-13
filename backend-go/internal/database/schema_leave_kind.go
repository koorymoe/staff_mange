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
		{
			// ═══ رجوع لحالتين بس ═══
			//
			// أربع أنواع كانت أكثر من اللازم: كل نوع إضافي يخلي الموظف
			// يوقف يفكّر «أي وحدة أختار؟»، ويخلي المدير يقرا تصنيف بدل
			// ما يقرا السبب المكتوب.
			//
			// بقت حالتان، والفرق بينهن **مهلة التقديم** مو تصنيف إداري:
			//   REGULAR → قبل يومين (حتى يرتبون الشفت)
			//   URGENT  → بلا مهلة
			//
			// ⚠️ الترتيب مهم: نحوّل الصفوف القديمة **قبل** ما نضيّق
			// القيد. بالعكس، القيد يفشل على أول صف SICK موجود ويوقف
			// الترحيل كله ويمنع السيرفر من الإقلاع.
			//
			// ⚠️ SICK يصير REGULAR مو URGENT: المرضية كانت مستثناة من
			// المهلة، بس تحويلها لـURGENT يخلي طلبات قديمة تنقرا «طارئة»
			// وهي ما كانت. الطلبات القديمة انبتّ بيها أصلاً، فالتسمية
			// الدقيقة أهم من المهلة الي ما عاد إلها معنى.
			Version: "0241_leave_kind_two_states",
			SQL: `
				UPDATE "LeaveRequest" SET kind = 'REGULAR' WHERE kind <> 'URGENT';

				ALTER TABLE "LeaveRequest"
					DROP CONSTRAINT IF EXISTS leave_request_kind_valid;

				ALTER TABLE "LeaveRequest"
					ADD CONSTRAINT leave_request_kind_valid
					CHECK (kind IN ('REGULAR','URGENT'));
			`,
		},
	}
}
