package database

// ═══ مختبر المحاكاة ═══
//
// «أريد محاكيات… الموظف من يجي أنطي كورس يدرسه وبعدها يجي يطبّق هنا.
// تعرف بكت تريسر شلون يطبقون عليه شبكات؟ أريد مثله… القفل الإلكتروني
// بيه ١٥ واير ملوّنات كل لون شنو يعني وشلون يربطهنة بكهرباء وشلون
// يبرمجه».
//
// الفكرة: الموظف ياخذ الدرس النظري، وبعدها **يطبّق بإيده** على محاكي —
// يسحب سلكاً ويربطه، والنظام يرد عليه فوراً: صح، أو غلط وليش.
//
// ⚠️ كلشي **بيانات مو كود**. الجهاز والدرس والتمرين كلهن صفوف بقاعدة
// البيانات، والكود ينزّل **محرّكات** بس. لأن لو انبنى كل جهاز بملف كود،
// يوم يتوقف الي يكتب الكود يموت المشروع — والشركة عندها تسع خدمات
// وعشرات الموديلات. هيج فنيو الشركة يكمّلونه بأنفسهم.
//
// ⚠️ ثلاثة محرّكات تغطّي كل الخدمات:
//
//	WIRING — توصيل أسلاك على صورة الجهاز (أقفال، كاميرات، حريق، شمسية)
//	CLI    — طرفية أوامر (سويچات وراوترات هواوي/سيسكو)
//	PANEL  — كيباد أو واجهة برمجة الجهاز (قفل، كاميرا، إنفرتر)
//
// و`engineKind` عمود **نصّي مو enum** حتى إضافة نوع رابع تبقى ممكنة
// بلا ترحيل يعدّل نوعاً بقاعدة البيانات.
func simLabMigrations() []Migration {
	return []Migration{
		{
			// ═══ الفهرس والمحتوى ═══
			Version: "0251_sim_lab_core",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SimCategory" (
					id            TEXT PRIMARY KEY,
					"serviceId"   TEXT REFERENCES "Service"(id) ON DELETE SET NULL,
					name          TEXT NOT NULL,
					description   TEXT,
					"imagePath"   TEXT,
					"sortOrder"   INTEGER NOT NULL DEFAULT 0,
					archived      BOOLEAN NOT NULL DEFAULT FALSE,
					"createdAt"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"updatedAt"   TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS "SimCategory_service_idx"
					ON "SimCategory"("serviceId", archived, "sortOrder");

				CREATE TABLE IF NOT EXISTS "SimDevice" (
					id             TEXT PRIMARY KEY,
					"categoryId"   TEXT NOT NULL REFERENCES "SimCategory"(id) ON DELETE CASCADE,
					brand          TEXT NOT NULL,
					model          TEXT NOT NULL,
					name           TEXT NOT NULL,
					summary        TEXT,
					"imagePath"    TEXT,
					"engineKind"   TEXT NOT NULL DEFAULT 'WIRING',
					spec           JSONB NOT NULL DEFAULT '{}'::jsonb,
					terminals      JSONB NOT NULL DEFAULT '[]'::jsonb,
					ui             JSONB NOT NULL DEFAULT '{}'::jsonb,
					status         TEXT NOT NULL DEFAULT 'DRAFT',
					version        INTEGER NOT NULL DEFAULT 1,
					"sourceRef"    TEXT,
					"localPractice" TEXT,
					verified       BOOLEAN NOT NULL DEFAULT FALSE,
					"verifiedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"verifiedAt"   TIMESTAMPTZ,
					"authorId"     TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"reviewedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"reviewedAt"   TIMESTAMPTZ,
					"reviewNote"   TEXT,
					"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"updatedAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS "SimDevice_category_idx"
					ON "SimDevice"("categoryId", status, verified);

				CREATE TABLE IF NOT EXISTS "SimCliGrammar" (
					id          TEXT PRIMARY KEY,
					name        TEXT NOT NULL,
					os          TEXT NOT NULL,
					tree        JSONB NOT NULL DEFAULT '{}'::jsonb,
					status      TEXT NOT NULL DEFAULT 'DRAFT',
					version     INTEGER NOT NULL DEFAULT 1,
					"sourceRef" TEXT,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				CREATE TABLE IF NOT EXISTS "SimLesson" (
					id           TEXT PRIMARY KEY,
					"categoryId" TEXT REFERENCES "SimCategory"(id) ON DELETE CASCADE,
					"deviceId"   TEXT REFERENCES "SimDevice"(id) ON DELETE CASCADE,
					title        TEXT NOT NULL,
					blocks       JSONB NOT NULL DEFAULT '[]'::jsonb,
					"sortOrder"  INTEGER NOT NULL DEFAULT 0,
					status       TEXT NOT NULL DEFAULT 'DRAFT',
					"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"updatedAt"  TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS "SimLesson_cat_idx"
					ON "SimLesson"("categoryId", "sortOrder");

				CREATE TABLE IF NOT EXISTS "SimExercise" (
					id             TEXT PRIMARY KEY,
					"categoryId"   TEXT NOT NULL REFERENCES "SimCategory"(id) ON DELETE CASCADE,
					title          TEXT NOT NULL,
					brief          TEXT,
					"engineKind"   TEXT NOT NULL DEFAULT 'WIRING',
					difficulty     INTEGER NOT NULL DEFAULT 1,
					"timeLimitSec" INTEGER,
					"passScore"    INTEGER NOT NULL DEFAULT 80,
					"maxAttempts"  INTEGER,
					scene          JSONB NOT NULL DEFAULT '{}'::jsonb,
					steps          JSONB NOT NULL DEFAULT '[]'::jsonb,
					"skillId"      TEXT REFERENCES "Skill"(id) ON DELETE SET NULL,
					status         TEXT NOT NULL DEFAULT 'DRAFT',
					version        INTEGER NOT NULL DEFAULT 1,
					"sourceRef"    TEXT,
					"localPractice" TEXT,
					verified       BOOLEAN NOT NULL DEFAULT FALSE,
					"verifiedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"verifiedAt"   TIMESTAMPTZ,
					"authorId"     TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"reviewedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"reviewedAt"   TIMESTAMPTZ,
					"sortOrder"    INTEGER NOT NULL DEFAULT 0,
					"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"updatedAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS "SimExercise_category_idx"
					ON "SimExercise"("categoryId", status, verified, "sortOrder");

				CREATE TABLE IF NOT EXISTS "SimContentVersion" (
					id              TEXT PRIMARY KEY,
					"entityKind"    TEXT NOT NULL,
					"entityId"      TEXT NOT NULL,
					version         INTEGER NOT NULL,
					snapshot        JSONB NOT NULL,
					"publishedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"publishedAt"   TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS "SimContentVersion_uniq"
					ON "SimContentVersion"("entityKind", "entityId", version);
			`,
		},
		{
			// ═══ المحاولات ═══
			// ⚠️ `state` تخزّن المشهد لحظة بلحظة: الفني يوقّف ويكمّل
			// باچر بلا ما يعيد من الأول. تمرين توصيل ١٥ سلك مو شي
			// ينخلص بخمس دقائق.
			Version: "0252_sim_lab_attempts",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SimAttempt" (
					id                TEXT PRIMARY KEY,
					"exerciseId"      TEXT NOT NULL REFERENCES "SimExercise"(id) ON DELETE CASCADE,
					"exerciseVersion" INTEGER NOT NULL DEFAULT 1,
					"employeeId"      TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
					status            TEXT NOT NULL DEFAULT 'IN_PROGRESS',
					score             INTEGER,
					"stepsTotal"      INTEGER NOT NULL DEFAULT 0,
					"stepsPassed"     INTEGER NOT NULL DEFAULT 0,
					"hintsUsed"       INTEGER NOT NULL DEFAULT 0,
					"wrongCount"      INTEGER NOT NULL DEFAULT 0,
					"durationSec"     INTEGER,
					state             JSONB NOT NULL DEFAULT '{}'::jsonb,
					"startedAt"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"finishedAt"      TIMESTAMPTZ,
					"createdAt"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"updatedAt"       TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS "SimAttempt_emp_idx"
					ON "SimAttempt"("employeeId", "startedAt" DESC);
				CREATE INDEX IF NOT EXISTS "SimAttempt_ex_idx"
					ON "SimAttempt"("exerciseId", status);

				CREATE TABLE IF NOT EXISTS "SimAttemptEvent" (
					id          TEXT PRIMARY KEY,
					"attemptId" TEXT NOT NULL REFERENCES "SimAttempt"(id) ON DELETE CASCADE,
					"stepIndex" INTEGER,
					kind        TEXT NOT NULL,
					payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
					"atMs"      INTEGER NOT NULL DEFAULT 0,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS "SimAttemptEvent_att_idx"
					ON "SimAttemptEvent"("attemptId", "atMs");

				CREATE TABLE IF NOT EXISTS "SimMastery" (
					id            TEXT PRIMARY KEY,
					"employeeId"  TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
					"exerciseId"  TEXT NOT NULL REFERENCES "SimExercise"(id) ON DELETE CASCADE,
					"bestScore"   INTEGER NOT NULL DEFAULT 0,
					attempts      INTEGER NOT NULL DEFAULT 0,
					passed        BOOLEAN NOT NULL DEFAULT FALSE,
					"firstPassAt" TIMESTAMPTZ,
					"lastAt"      TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS "SimMastery_uniq"
					ON "SimMastery"("employeeId", "exerciseId");
			`,
		},
		{
			// ═══ إعدادات النشر ═══
			// ⚠️ ينبنى ولا يُقرأ بهالمرحلة. المختبر مخفي بـRequireOwner
			// ثابتة بالكود — والبوابة أخطر ما تكون بالضبط بفترة الإخفاء،
			// فما نضيف علماً بقاعدة بيانات يگدر ينكشف بالغلط قبل أوانه.
			Version: "0253_sim_lab_config",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SimLabConfig" (
					id              TEXT PRIMARY KEY DEFAULT 'singleton',
					published       BOOLEAN NOT NULL DEFAULT FALSE,
					"audienceRoles" TEXT[] NOT NULL DEFAULT '{}',
					"safetyNotice"  TEXT NOT NULL DEFAULT 'هذا تدريب محاكاة. لا تعتمد عليه بالميدان — الرجوع لكتالوگ الشركة المصنّعة إلزامي قبل أي توصيل حقيقي.',
					"updatedById"   TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"updatedAt"     TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				INSERT INTO "SimLabConfig" (id) VALUES ('singleton') ON CONFLICT (id) DO NOTHING;
			`,
		},
		{
			// ═══ تصحيح مواقع المشهد ═══
			//
			// بالبذرة الأولى القفل چان **يسار** المشهد وأطرافه على حافّته
			// **اليسرى**، والمغذّي **يمين** وأطرافه على حافّته **اليمنى**.
			// يعني كل سلك لازم يلف حول الجهازين بدل ما يمشي بالفراغ بينهما.
			//
			// الإصلاح بالمواقع بس (الأطراف تبقى مثل ما هي): القفل يمين
			// والمغذّي يسار — فأطراف القفل اليسرى تقابل أطراف المغذّي
			// اليمنى مباشرة. وهذا يقرا صح بواجهة عربية (الأساسي يمين).
			//
			// ⚠️ ترحيل مستقل مو تعديل البذرة: البذرة `ON CONFLICT DO NOTHING`
			// فما تكتب فوق الموجود، والترحيلات للأمام بس.
			Version: "0254_sim_lock_scene_fix",
			SQL: `
				UPDATE "SimExercise"
				SET scene = jsonb_set(
				      jsonb_set(scene, '{devices,0,x}', '0.68'::jsonb),
				      '{devices,1,x}', '0.22'::jsonb),
				    "updatedAt" = NOW()
				WHERE id = 'simex_ac_keypad_wiring'
				  AND scene #>> '{devices,0,x}' = '0.30';
			`,
		},
		{
			// ═══ الهندسة ثلاثية الأبعاد للجهاز ═══
			//
			// المخطط الرئيسي (القسم ٣) يعتبر الجهاز **توأماً رقمياً** له
			// Geometry، والشكل ثلاثي الأبعاد مجرد View لنفس الكائن. فالهندسة
			// تعيش بالخلفية مع الجهاز — مو مكتوبة بالواجهة.
			//
			// ⚠️ الوحدة **متر دائماً** ومقياس ١:١ (القسم ٧٫٢). الأبعاد
			// أدناه تقريب عام لحجم لوحة كيباد جدارية ومحوّل صغير — مثل
			// باقي محتوى هذا الجهاز **غير محقّقة** من كتالوگ موديل بعينه.
			//
			// ⚠️ ماكو ملف موديل مصنّع: الأجسام **تتولّد بالكود** من هذي
			// المواصفة. ولمن تجي موديلات حقيقية بعدين تنركّب فوگ نفس
			// المراسي الدلالية بلا إعادة عمل (القسم ٧٫٣).
			//
			// ⚠️ ماكو إحداثيات ثلاثية للأطراف: الطرف يبقى (x,y) نسبة من
			// **وجه** الجهاز — نفس الأرقام الي يستعملها المنظر المنطقي.
			// منظر واحد ما يقدر يزيح الثاني (القسم ١٩).
			Version: "0255_sim_device_geometry",
			SQL: `
				ALTER TABLE "SimDevice"
				  ADD COLUMN IF NOT EXISTS geometry jsonb NOT NULL DEFAULT '{}'::jsonb;

				UPDATE "SimDevice" SET geometry = '{
				  "shape": "wall_box",
				  "sizeM": {"w": 0.145, "h": 0.145, "d": 0.030},
				  "bodyColorHex": "#3f4756",
				  "faceColorHex": "#232a36",
				  "terminalPost": {"radiusM": 0.0035, "heightM": 0.0055},
				  "features": [
				    {"kind": "keypad", "x": 0.62, "y": 0.52, "w": 0.30, "h": 0.62, "cols": 3, "rows": 4},
				    {"kind": "statusLed", "x": 0.62, "y": 0.10, "channel": "status_led"},
				    {"kind": "terminalPlate", "x0": 0.03, "y0": 0.08, "x1": 0.26, "y1": 0.94}
				  ]
				}'::jsonb, "updatedAt" = NOW()
				WHERE id = 'simdev_ac_keypad_15w' AND geometry = '{}'::jsonb;

				UPDATE "SimDevice" SET geometry = '{
				  "shape": "psu_brick",
				  "sizeM": {"w": 0.110, "h": 0.075, "d": 0.045},
				  "bodyColorHex": "#4b5563",
				  "faceColorHex": "#2b3242",
				  "terminalPost": {"radiusM": 0.0035, "heightM": 0.0055},
				  "features": [
				    {"kind": "statusLed", "x": 0.35, "y": 0.72, "channel": "status_led"},
				    {"kind": "terminalPlate", "x0": 0.68, "y0": 0.16, "x1": 0.94, "y1": 0.60}
				  ]
				}'::jsonb, "updatedAt" = NOW()
				WHERE id = 'simdev_psu_12v' AND geometry = '{}'::jsonb;
			`,
		},
		{
			// ═══ مخططات مساحة العمل ═══
			//
			// اللوح بلا حفظ يعني شغل يروح مع تسكير الصفحة. والمخطط مو
			// رسمة: هو **تصميم مشروع** — الفني يبني توبولوجي مشروع
			// حقيقي، يفحصه، ويرجعله بعدين.
			//
			// ⚠️ `employeeId` مو `ownerId`: المرحلة الحالية للمالك وحده
			// بس الجدول ما ينبنى على هالافتراض. لمن ينفتح للفنيين، كل
			// واحد يشوف مخططاته بلا ترحيل جديد.
			//
			// ⚠️ المستند كله `jsonb` عمداً: شكله يتغيّر مع كل قطعة
			// جديدة تنضاف للكتالوگ، وأعمدة مفصّلة تعني ترحيلاً بكل
			// إضافة.
			Version: "0256_sim_project",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SimProject" (
					id           TEXT PRIMARY KEY,
					"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
					name         TEXT NOT NULL,
					domain       TEXT NOT NULL,
					doc          JSONB NOT NULL DEFAULT '{}'::jsonb,
					notes        TEXT,
					"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"updatedAt"  TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS "SimProject_emp_idx"
					ON "SimProject"("employeeId", "updatedAt" DESC);
			`,
		},
	}
}
