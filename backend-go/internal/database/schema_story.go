package database

// ══════════════════════════════════════════════════════════════════
// محرّك القصص — «المراقب الحي»
// ══════════════════════════════════════════════════════════════════
//
// الكيان ما يعرض تنبيهاً ويخلص: ينفّذ **قصة** بين واجهتين. المراقب
// يخصم نقطة ← شخصيته تحمل ورقة وتخرج من شاشته ← تدخل شاشة الموظف
// وتسلّمه إياها ← الأفتار يفتحها ويقراها.
//
// ⚠️ **الشخصية ما تنتقل مادياً بين جهازين.** الخادم يحفظ **قصة وحدة
// بمراحل**، وكل واجهة تعرض فصلها. التطابق بالمعرّف والتوقيت هو الي
// يخلّيها تبدو انتقالاً.
//
// ⚠️ **صفر جدول أحداث جديد**: الحدث الرسمي موجود أصلاً بجدوله
// (`KpiEvaluation` · `DisciplineEvent` · ...) ونخزن `id` مالته بس.
// جدول أحداث ثانٍ يعني **مصدرَي حقيقة يفترقان** — نفس العلّة الي
// نطاردها بكل جولة.
//
// ⚠️ **الحركة لا تسبق نجاح العملية**: القصة تنكتب **بعد** ما ينجح
// الإجراء الإداري. وفشل كتابة القصة **ما يلغي** الإجراء — بس ما
// ينبلع بصمت، ينسجّل بالسجل.
func storyMigrations() []Migration {
	return []Migration{
		{
			Version: "0272_story_instance",
			SQL: `CREATE TABLE IF NOT EXISTS "StoryInstance" (
			        id TEXT PRIMARY KEY,
			        "eventId"   TEXT NOT NULL,
			        "eventKind" TEXT NOT NULL,
			        "storyType" TEXT NOT NULL,
			        version     INT  NOT NULL DEFAULT 1,
			        "senderEmployeeId"    TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
			        "senderName"          TEXT NOT NULL DEFAULT '',
			        "recipientEmployeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
			        status      TEXT NOT NULL DEFAULT 'QUEUED',
			        priority    INT  NOT NULL DEFAULT 50,
			        physical    BOOLEAN NOT NULL DEFAULT true,
			        "groupKey"  TEXT,
			        "currentStep" INT NOT NULL DEFAULT 0,
			        payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
			        "deliveredAt"    TIMESTAMPTZ,
			        "seenAt"         TIMESTAMPTZ,
			        "openedAt"       TIMESTAMPTZ,
			        "acknowledgedAt" TIMESTAMPTZ,
			        "createdAt"      TIMESTAMPTZ NOT NULL DEFAULT now(),
			        "expiresAt"      TIMESTAMPTZ
			      )`,
		},
		{
			// ⚠️ **الـidempotency فهرس مو كود**: نفس الحدث لنفس المستلم
			// ما ينشئ قصة ثانية مهما انعاد النداء أو انطفى الخادم ورجع.
			// نفس نمط `discipline_event_unique_penalty` الموجود
			// (schema_inventory_features.go) — ننسخ القرار لا نخترعه.
			Version: "0272_story_event_recipient_unique",
			SQL: `CREATE UNIQUE INDEX IF NOT EXISTS "story_event_recipient_unique"
			        ON "StoryInstance" ("eventId", "recipientEmployeeId")`,
		},
		{
			// الواجهة تسأل سؤالاً واحداً كل استطلاع: «شيء ينتظرني؟».
			// فهرس جزئي على المعلّقة بس — الأعلى أولوية ثم الأقدم.
			Version: "0272_story_pending_idx",
			SQL: `CREATE INDEX IF NOT EXISTS "story_pending_idx"
			        ON "StoryInstance" ("recipientEmployeeId", priority DESC, "createdAt")
			        WHERE status IN ('QUEUED','DELIVERED','PLAYING')`,
		},
		{
			// ⚠️ **التجميع غير التكرار**: الفهرس فوق يمنع **نفس الحدث**
			// مرتين. وهذا يمنع **الازدحام**: ثلاث أوراق ناقصة بنفس الحجز
			// تندمج بمشهد واحد يذكر الثلاثة، بدل ثلاث ركضات ورا بعض.
			// محصور بالمعلّقة — قصة انلعبت خلاص، الجديدة تبدي طابورها.
			Version: "0272_story_group_unique",
			SQL: `CREATE UNIQUE INDEX IF NOT EXISTS "story_group_open_unique"
			        ON "StoryInstance" ("recipientEmployeeId", "groupKey")
			        WHERE "groupKey" IS NOT NULL AND status IN ('QUEUED','DELIVERED','PLAYING')`,
		},
	}
}
