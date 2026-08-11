package database

// ══════════════════════════════════════════════════════════════════
// هيكلية الذكاء الاصطناعي داخل النظام
// ══════════════════════════════════════════════════════════════════
//
// طلب صاحب العمل: «مو محادثة مع الذكاء — أريده امتدادات لجميع أنحاء
// النظام، ضمن الباك إند، يحلل ويقيس ويؤشّر ويتوقع الخطأ قبل ليصير».
// وحالياً: **تهيكل بس** لحد ما ننشترك بمنصّة.
//
// ⚠️ الفصل الأساسي بهذا التصميم — وهو أهم قرار هنا:
//
//   الأدلة (Evidence)  =  حقائق منتزعة من قاعدة البيانات. حتمية،
//                         تنحسب بالكود، تنعاد نفسها كل مرة.
//   الحكم   (Verdict)  =  تفسير وسبب ورأي. هذا شغل النموذج.
//
// ليش الفصل؟ لأن الحكم بلا أدلة = تخمين، والمنصّة تكذب بثقة. لما
// نجمع الأدلة بأنفسنا: (١) نكدر نعرض للمالك «هاي الحقائق» حتى لو
// النموذج غلط، (٢) الحساب ما يعتمد على شي خارجي ممكن يطيح، (٣) لما
// نشترك بالمنصّة تشتغل على أدلة جاهزة فتصير أرخص وأدق.
//
// يعني اليوم النظام **يجمع ويقيس** بلا اشتراك، وبكرة النموذج يجي
// يفسّر. الي انبنى هسه ما ينرمي.
//
// ⚠️ التقارير للمالك ومدير النظام حصراً — طلب صريح. تحليل «ليش هذا
// الموظف وقّف الشغل» بيد زميله يصير سلاح مو إدارة.
func aiCoreMigration() []Migration {
	return []Migration{
		{
			Version: "0243_ai_core",
			SQL: `
				-- ═══ ١. الإشارة ═══
				-- «شي صار بالنظام يستاهل تحليل». الإشارة تنسجّل لحظة
				-- الحدث، والتحليل يصير بعدها — حتى ما نأخّر شغل الموظف
				-- بانتظار تحليل.
				CREATE TABLE IF NOT EXISTS "AiSignal" (
					id TEXT PRIMARY KEY,
					kind TEXT NOT NULL,              -- WORK_STOPPED, LATE_START, ...
					"entityType" TEXT NOT NULL,      -- BOOKING, EMPLOYEE, INVOICE
					"entityId" TEXT NOT NULL,
					"employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					payload JSONB NOT NULL DEFAULT '{}'::jsonb,
					status TEXT NOT NULL DEFAULT 'PENDING',  -- PENDING|COLLECTED|ANALYZED|SKIPPED
					"occurredAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				-- نفس الحدث ما ينسجّل مرتين: توقف عمل واحد = إشارة وحدة.
				CREATE UNIQUE INDEX IF NOT EXISTS "AiSignal_unique_event"
					ON "AiSignal" (kind, "entityType", "entityId", "occurredAt");
				CREATE INDEX IF NOT EXISTS "AiSignal_pending_idx"
					ON "AiSignal" (status, "occurredAt" DESC);

				-- ═══ ٢. ملف الأدلة ═══
				-- حقائق منتزعة من الجداول: هل طلب مادة؟ متى انضافت
				-- للسلة؟ الدوام خالص لو لا؟ كلها **محسوبة** مو مخمّنة.
				CREATE TABLE IF NOT EXISTS "AiEvidence" (
					id TEXT PRIMARY KEY,
					"signalId" TEXT NOT NULL REFERENCES "AiSignal"(id) ON DELETE CASCADE,
					facts JSONB NOT NULL DEFAULT '{}'::jsonb,
					-- الي ما قدرنا نجمعه ونعرف إنه ناقص. الفراغ المعلن
					-- أأمن من الفراغ الصامت: النموذج لازم يعرف شنو ما
					-- شافه بدل ما يفترض إنه ماكو.
					gaps JSONB NOT NULL DEFAULT '[]'::jsonb,
					"collectedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS "AiEvidence_signal_unique"
					ON "AiEvidence" ("signalId");

				-- ═══ ٣. الحكم ═══
				-- تفسير الأدلة. اليوم ينتعبى بمحرّك قواعد بسيط، وبكرة
				-- بالمنصّة — ونفس الجدول يخدم الاثنين.
				CREATE TABLE IF NOT EXISTS "AiVerdict" (
					id TEXT PRIMARY KEY,
					"signalId" TEXT NOT NULL REFERENCES "AiSignal"(id) ON DELETE CASCADE,
					-- منو حكم: RULES = محرّك القواعد عدنا، MODEL = منصّة
					-- خارجية. لازم ينعرف حتى المالك يعرف وزن الكلام.
					source TEXT NOT NULL DEFAULT 'RULES',
					"modelName" TEXT,
					headline TEXT NOT NULL,
					reasoning TEXT,
					-- 0..100. ⚠️ ما ينعرض كـ«حقيقة» تحت ٧٠ — يتأشر «مو متأكد».
					confidence INT NOT NULL DEFAULT 0,
					severity TEXT NOT NULL DEFAULT 'INFO',   -- INFO|WATCH|WARN|CRITICAL
					-- منو المسؤول حسب التحليل. ⚠️ ممكن يكون NULL: «ماكو
					-- مسؤول واضح» جواب مشروع، وإجبار اسم يخلق ظالم.
					"blameEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					suggestion TEXT,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS "AiVerdict_signal_idx" ON "AiVerdict" ("signalId");
				CREATE INDEX IF NOT EXISTS "AiVerdict_severity_idx"
					ON "AiVerdict" (severity, "createdAt" DESC);

				-- ═══ ٤. المؤشرات ═══
				-- «إحصائيات ومؤشرات الذكاء الاصطناعي» — رقم بفترة
				-- محسوب من الأدلة، مو عدّاد خام.
				CREATE TABLE IF NOT EXISTS "AiMetric" (
					id TEXT PRIMARY KEY,
					"metricKey" TEXT NOT NULL,
					scope TEXT NOT NULL DEFAULT 'COMPANY',   -- COMPANY|EMPLOYEE|SERVICE
					"scopeId" TEXT,
					"periodStart" DATE NOT NULL,
					"periodEnd" DATE NOT NULL,
					value NUMERIC(14,3) NOT NULL DEFAULT 0,
					"sampleCount" INT NOT NULL DEFAULT 0,
					details JSONB NOT NULL DEFAULT '{}'::jsonb,
					"computedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				-- إعادة الحساب تحدّث الصف مو تضيف صف ثاني.
				CREATE UNIQUE INDEX IF NOT EXISTS "AiMetric_unique"
					ON "AiMetric" ("metricKey", scope, COALESCE("scopeId",''), "periodStart", "periodEnd");

				-- ═══ ٥. ساعات الدوام ═══
				-- الأدلة تحتاجها: «وقّف الشغل ٥ دقايق قبل نهاية الدوام»
				-- تفسير مختلف تماماً عن «وقّف الساعة ١١ صباحاً».
				-- كانت مبعثرة بالكود، وهسه مصدر واحد يعدّله المالك.
				CREATE TABLE IF NOT EXISTS "AiWorkWindow" (
					id TEXT PRIMARY KEY,
					"startHour" INT NOT NULL DEFAULT 9,
					"endHour" INT NOT NULL DEFAULT 24,
					"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				-- ٩ صباحاً لـ١٢ ليلاً — بالضبط الي گاله صاحب العمل.
				-- ⚠️ ١٢ ليلاً = ٢٤ بالعد، مو ٠: الصفر يعني «الدوام خالص
				-- قبل ما يبدي» وكل حساب بعده يطلع بالسالب.
				INSERT INTO "AiWorkWindow" (id, "startHour", "endHour")
				SELECT 'default', 9, 24
				WHERE NOT EXISTS (SELECT 1 FROM "AiWorkWindow" WHERE id = 'default');
			`,
		},
	}
}
