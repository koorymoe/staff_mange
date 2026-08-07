package database

// solarVersionedMigrations يرجّع ترحيلات نظام الطاقة الشمسية (Solar Expert).
//
// ═══ من وين جاي هذا النظام ═══
// جان نظام منفصل على Google Sheets: ملف Apps Script + واجهة HTML، وبيه ٨
// شيتات. أربعة منهن (موظفين، مهارات، تدريب، زبائن) عدنا مثلهن بالنظام
// أصلاً — فما ننقلهن، ننقل الي ما موجود ونربطه بالموجود:
//
//   Solar_Data      → "SolarSystem"      (جديد)
//   Inventory       → "SolarComponent"   (جديد — مكوّنات بمواصفات تخصصية)
//   Customers       → "SolarInstallation" + "Customer" الموجود
//   Employees       → أعمدة إضافية على "Employee" الموجود
//   Skills_Catalog  → "Skill" الموجود + عمود تصنيف
//   Training        → "TrainingProgram" (جديد، على موظفينا ومهاراتنا)
//   System_Logs     → "SolarLog"
//   Processed_Count → ما يحتاج جدول: عدّة صفوف "SolarInstallation"
//
// ليش المكوّنات جدول لحاله ومو بالمخزن الموجود؟ لأن مخزننا نوعين: عدة
// الفنيين ("OnDemandTool") ومواد التنفيذ ("Material"). لوح شمسي بمواصفات
// Vmp/Imp/Voc/Isc وحد أدنى للمخزون مو هذا ولا ذاك — بضاعة تنباع بمواصفات
// فنية يعتمد عليها اختيار المنظومة.
func solarVersionedMigrations() []Migration {
	return []Migration{
		{
			// ═══ مكوّنات المنظومات ═══
			// النظام القديم جان يخزنهن بطريقة موجعة: خمس أعمدة أسماء
			// (لوح/إنفيرتر/بطارية/بورد/حديد) والصف يملأ عمود واحد بس،
			// والفارغ هو الي «يحدد» التصنيف. هنا التصنيف عمود صريح.
			//
			// المواصفات ١١ خانة لأن كل تصنيف إله مواصفاته: اللوح Vmp/Imp/
			// Voc/Isc والكفاءة، الإنفيرتر MPPT وجهد البطارية، البطارية Ah
			// وعدد الدورات وDOD. نخزنهن JSONB بمفاتيح واضحة بدل specA..specK
			// حتى يبقى معناهن بالبيانات نفسها مو بذاكرة الي كتب الكود.
			Version: "0203_solar_component",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SolarComponent" (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					category TEXT NOT NULL CHECK (category IN ('PANEL','INVERTER','BATTERY','BOARD','IRON')),
					quantity INTEGER NOT NULL DEFAULT 0,
					price DOUBLE PRECISION NOT NULL DEFAULT 0,
					"minStock" INTEGER NOT NULL DEFAULT 0,
					specs JSONB NOT NULL DEFAULT '{}'::jsonb,
					notes TEXT,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
					"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				-- الاسم مفتاح الاختيار بالمنظومة، وما ينعاد بنفس التصنيف
				CREATE UNIQUE INDEX IF NOT EXISTS "SolarComponent_name_category_key"
					ON "SolarComponent" (lower(btrim(name)), category);
				CREATE INDEX IF NOT EXISTS "SolarComponent_category_idx"
					ON "SolarComponent" (category);
				-- شاشة التنبيهات تسأل «منو تحت الحد الأدنى» بكل فتحة
				CREATE INDEX IF NOT EXISTS "SolarComponent_low_stock_idx"
					ON "SolarComponent" (quantity) WHERE quantity <= 0;
			`,
		},
		{
			// ═══ كتالوك المنظومات الجاهزة ═══
			// المنظومة = ماركة وموديل وسعة + قائمة مكوّنات بكميّاتها +
			// تكاليف تسليك وحديد وتنصيب وبرمجة وضمان.
			//
			// المكوّنات مربوطة بـ id مو باسم نصّي — بالنظام القديم جان
			// الاسم مكتوب بالخلية، فأي تعديل على اسم المادة بالمخزن يكسر
			// كل منظومة تستعملها بالسكوت.
			//
			// ON DELETE RESTRICT مقصودة: ما نخلي أحد يمحي مكوّن مستعمل
			// بمنظومة، لأن المحو يخلي المنظومة بلا لوح ولا يحچي.
			//
			// التسليك والحديد سطور متغيّرة العدد (نوع الكابل وطوله وسعر
			// المتر / نوع الحديد ومقاسه ووحدته) — JSONB بدل جدول لأنهن
			// ما ينبحث بيهن، ينقرون ويتكتبون ويّا المنظومة دائماً.
			Version: "0204_solar_system",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SolarSystem" (
					id TEXT PRIMARY KEY,
					brand TEXT NOT NULL,
					model TEXT NOT NULL,
					capacity TEXT NOT NULL,

					"panelId"    TEXT REFERENCES "SolarComponent"(id) ON DELETE RESTRICT,
					"panelQty"   INTEGER NOT NULL DEFAULT 0,
					"inverterId" TEXT REFERENCES "SolarComponent"(id) ON DELETE RESTRICT,
					"inverterQty" INTEGER NOT NULL DEFAULT 0,
					"batteryId"  TEXT REFERENCES "SolarComponent"(id) ON DELETE RESTRICT,
					"batteryQty" INTEGER NOT NULL DEFAULT 0,
					"boardId"    TEXT REFERENCES "SolarComponent"(id) ON DELETE RESTRICT,

					"wiringDetails" JSONB NOT NULL DEFAULT '[]'::jsonb,
					"wiringTotalCost" DOUBLE PRECISION NOT NULL DEFAULT 0,
					"ironDetails" JSONB NOT NULL DEFAULT '[]'::jsonb,
					"ironTotalCost" DOUBLE PRECISION NOT NULL DEFAULT 0,

					"installPrice"  DOUBLE PRECISION NOT NULL DEFAULT 0,
					"programPrice"  DOUBLE PRECISION NOT NULL DEFAULT 0,
					"warrantyPrice" DOUBLE PRECISION NOT NULL DEFAULT 0,

					notes TEXT,
					"createdById" TEXT REFERENCES "Employee"(id),
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
					"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "SolarSystem_brand_idx" ON "SolarSystem" (brand);
				CREATE INDEX IF NOT EXISTS "SolarSystem_panel_idx" ON "SolarSystem" ("panelId");
				CREATE INDEX IF NOT EXISTS "SolarSystem_inverter_idx" ON "SolarSystem" ("inverterId");
				CREATE INDEX IF NOT EXISTS "SolarSystem_battery_idx" ON "SolarSystem" ("batteryId");
				CREATE INDEX IF NOT EXISTS "SolarSystem_board_idx" ON "SolarSystem" ("boardId");
			`,
		},
		{
			// ═══ تجهيز منظومة لزبون ═══
			// هاي أقوى شغلة بالنظام القديم: تجهيز منظومة = خصم مكوّناتها
			// من المخزن + تسجيل الزبون + متابعته بعد ٣٠ يوم.
			//
			// الزبون يربط بجدول "Customer" الموجود مو بجدول زبائن ثاني —
			// حتى زبون الطاقة الشمسية يضل نفس زبون الحجوزات والفواتير.
			//
			// نخزن "followUpAt" محسوبة وقت التجهيز بدل ما نحسبها كل مرة:
			// مدة المتابعة ممكن تتغير بكرة، والي انجهز على ٣٠ يوم لازم
			// يضل على ٣٠ يوم مو ينقلب أثراً رجعياً.
			//
			// وأسعار وقت البيع تنخزن نسخة (snapshot): سعر اللوح بالمخزن
			// يتغير كل شهر، ولو نحسب سعر تركيب قديم من السعر الحالي يطلع
			// رقم ما إله علاقة بالي انباع فعلاً.
			Version: "0205_solar_installation",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SolarInstallation" (
					id TEXT PRIMARY KEY,
					"systemId" TEXT NOT NULL REFERENCES "SolarSystem"(id) ON DELETE RESTRICT,
					"customerId" TEXT NOT NULL REFERENCES "Customer"(id) ON DELETE RESTRICT,
					"installDate" DATE NOT NULL,
					"followUpAt" DATE NOT NULL,
					"contactedAt" TIMESTAMPTZ,
					"contactedById" TEXT REFERENCES "Employee"(id),
					"contactNotes" TEXT,
					status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','CONTACTED')),

					"totalPrice" DOUBLE PRECISION NOT NULL DEFAULT 0,
					"priceBreakdown" JSONB NOT NULL DEFAULT '{}'::jsonb,

					notes TEXT,
					"createdById" TEXT REFERENCES "Employee"(id),
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "SolarInstallation_customer_idx"
					ON "SolarInstallation" ("customerId");
				CREATE INDEX IF NOT EXISTS "SolarInstallation_system_idx"
					ON "SolarInstallation" ("systemId");
				-- شاشة «يستحقون الاتصال» تسأل هذا السؤال بالضبط
				CREATE INDEX IF NOT EXISTS "SolarInstallation_due_idx"
					ON "SolarInstallation" ("followUpAt") WHERE status = 'PENDING';
			`,
		},
		{
			// ═══ سجل عمليات الطاقة الشمسية ═══
			// النظام القديم جان يسجّل كل عملية بشيت System_Logs بمنو
			// سواها. نفس الفكرة: خصم مخزن أو تعديل منظومة لازم يبقى إله
			// أثر يعرف منو ومتى.
			Version: "0206_solar_log",
			SQL: `
				CREATE TABLE IF NOT EXISTS "SolarLog" (
					id TEXT PRIMARY KEY,
					kind TEXT NOT NULL,
					details TEXT NOT NULL,
					payload JSONB,
					"employeeId" TEXT REFERENCES "Employee"(id),
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "SolarLog_createdAt_idx"
					ON "SolarLog" ("createdAt" DESC);
			`,
		},
		{
			// ═══ حقول الموارد البشرية من النظام القديم ═══
			// شيت Employees جان بيه قسم وتاريخ تعيين وخبرة وتقييم وحالة
			// وظيفية و«الوظيفة القادمة» و«الاحتياجات التدريبية». موظفينا
			// عندهم الاسم والهاتف والراتب والشهادة والمسمى — هذي الناقصة.
			//
			// نضيفهن على "Employee" الموجود مو على جدول موظفين ثاني: موظف
			// واحد بقائمة وحدة، والي عنده مهارة طاقة شمسية يطلع بشغلها
			// تلقائياً بلا ما ينضاف مرتين.
			//
			// كلها NULL-able بلا افتراضي إجباري — الموظفين الموجودين ما
			// يتأثرون، والحقول تنملى بالتدريج.
			Version: "0207_employee_hr_fields",
			SQL: `
				ALTER TABLE "Employee"
					ADD COLUMN IF NOT EXISTS department TEXT,
					ADD COLUMN IF NOT EXISTS "hireDate" DATE,
					ADD COLUMN IF NOT EXISTS "experienceYears" DOUBLE PRECISION,
					ADD COLUMN IF NOT EXISTS "lastReview" TEXT,
					ADD COLUMN IF NOT EXISTS "careerStatus" TEXT NOT NULL DEFAULT 'مستقر',
					ADD COLUMN IF NOT EXISTS "nextRole" TEXT,
					ADD COLUMN IF NOT EXISTS "trainingNeeds" TEXT;

				CREATE INDEX IF NOT EXISTS "Employee_department_idx"
					ON "Employee" (department) WHERE department IS NOT NULL;
				-- لوحة «منو يحتاج ترقية أو تدريب» تفلتر على هذا العمود
				CREATE INDEX IF NOT EXISTS "Employee_careerStatus_idx"
					ON "Employee" ("careerStatus");
			`,
		},
		{
			// ═══ تصنيف المهارة ═══
			// مهاراتنا مربوطة بخدمة (كاميرات، شبكات، طاقة شمسية...). النظام
			// القديم جان يصنّفهن بمحور ثاني: فنية / سلامة / إدارية — وهذا
			// محور مفيد ومستقل عن الخدمة (السلامة المهنية مو خدمة تنباع،
			// وخدمة العملاء تنفع لكل الخدمات).
			Version: "0208_skill_category",
			SQL: `
				ALTER TABLE "Skill"
					ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'فنية',
					ADD COLUMN IF NOT EXISTS description TEXT;
				CREATE INDEX IF NOT EXISTS "Skill_category_idx" ON "Skill" (category);
			`,
		},
		{
			// ═══ برامج التدريب ═══
			// عدنا "TrainingMaterial" (مواد تعليمية لكل خدمة) و
			// "EmployeeTrainingAssignment" (منو محوّل على تدريب خدمة).
			// ماكو عدنا «برنامج تدريبي»: دورة إلها مستوى ومدرّب ومشاركين
			// ومهارات مستهدفة ومدة وتكلفة ونسبة نجاح وتقدّم بخمس مراحل.
			// هذا الي ننقله، وينربط بموظفينا ومهاراتنا الموجودة.
			Version: "0209_training_program",
			SQL: `
				CREATE TABLE IF NOT EXISTS "TrainingProgram" (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					level TEXT NOT NULL DEFAULT 'مبتدئ' CHECK (level IN ('مبتدئ','متوسط','متقدم')),
					"durationDays" INTEGER NOT NULL DEFAULT 1,
					"startDate" DATE,
					"endDate" DATE,
					"targetDepartment" TEXT,
					"instructorId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					objectives TEXT,
					content TEXT,
					"passRate" INTEGER NOT NULL DEFAULT 80,
					cost DOUBLE PRECISION NOT NULL DEFAULT 0,
					status TEXT NOT NULL DEFAULT 'قيد التخطيط'
						CHECK (status IN ('قيد التخطيط','جاري التنفيذ','مكتمل')),
					progress INTEGER NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
					"createdById" TEXT REFERENCES "Employee"(id),
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
					"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);

				-- المشاركين والمهارات جداول ربط مو JSON: النظام القديم جان
				-- يخزنهن نص JSON بخلية، فسؤال بسيط مثل «هذا الموظف شنو
				-- تدرّب عليه؟» ما جان ينسأل — لازم تقرا كل الصفوف وتفكّهن.
				CREATE TABLE IF NOT EXISTS "TrainingProgramParticipant" (
					"programId" TEXT NOT NULL REFERENCES "TrainingProgram"(id) ON DELETE CASCADE,
					"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
					passed BOOLEAN,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
					PRIMARY KEY ("programId", "employeeId")
				);
				CREATE INDEX IF NOT EXISTS "TrainingProgramParticipant_employee_idx"
					ON "TrainingProgramParticipant" ("employeeId");

				CREATE TABLE IF NOT EXISTS "TrainingProgramSkill" (
					"programId" TEXT NOT NULL REFERENCES "TrainingProgram"(id) ON DELETE CASCADE,
					"skillId" TEXT NOT NULL REFERENCES "Skill"(id) ON DELETE CASCADE,
					PRIMARY KEY ("programId", "skillId")
				);
				CREATE INDEX IF NOT EXISTS "TrainingProgramSkill_skill_idx"
					ON "TrainingProgramSkill" ("skillId");

				CREATE INDEX IF NOT EXISTS "TrainingProgram_status_idx"
					ON "TrainingProgram" (status);
			`,
		},
	}
}

// employeeJobLevelMigration المستوى الوظيفي (١-١٠) من نظام الطاقة
// الشمسية.
//
// أول ما ربطته بـ"leaderSkillLevel" الموجود — وهذا غلط: ذاك درجة
// مهارة الليدر، وافتراضيته صفر لكل الموظفين. يعني قاعدة «مستوى <٤
// ← يحتاج تدريب» جانت تأشّر **كل موظف بالشركة** إنه يحتاج تدريب
// عاجل، وهذا إنذار كاذب بحجم الشركة كلها.
//
// المستوى الوظيفي شي ثاني: تقييم إداري من ١ لـ ١٠. عمود مستقل،
// وافتراضيته ٥ (وسط) حتى الموظفين الحاليين ما ينوسمون بشي ما
// انقيّموا بيه أصلاً.
func employeeJobLevelMigration() []Migration {
	return []Migration{
		{
			Version: "0214_employee_job_level",
			SQL: `
				ALTER TABLE "Employee"
					ADD COLUMN IF NOT EXISTS "jobLevel" INTEGER NOT NULL DEFAULT 5;
			`,
		},
	}
}

// solarBookingMigration نوع حجز «طاقة شمسية».
//
// موظف المبيعات يستلم زبون يريد منظومة شمسية — وهذا شغل مختلف عن
// التركيب العادي والصيانة: إله كتالوك منظومات وسعر يطلع من المخزن،
// ولازم ينفرز بالإحصاءات لحاله حتى نعرف شكد بعنا منظومات.
//
// والأهم: الحجز ينربط بمنظومة من الكتالوك، فيوصل للمنسّق ومعاه سعرها
// ومكوّناتها — بدل ما يتصل بالمبيعات يسأل «شنو المنظومة الي اتفقتوا
// عليها؟».
func solarBookingMigration() []Migration {
	return []Migration{
		{
			Version: "0215_booking_type_solar",
			SQL:     `ALTER TYPE "BookingType" ADD VALUE IF NOT EXISTS 'SOLAR'`,
		},
		{
			// المنظومة المطلوبة — اختيارية: الزبون أحياناً يريد منظومة
			// ولسه ما قرر أي وحدة، فالمبيعات يسجّل الحجز والمنسّق يحدد
			// المنظومة بعد المعاينة.
			//
			// ON DELETE SET NULL مو RESTRICT: محو منظومة من الكتالوك ما
			// يصير يقفل حجز قديم انسجّل عليها.
			Version: "0216_booking_solar_system",
			SQL: `
				ALTER TABLE "Booking"
					ADD COLUMN IF NOT EXISTS "solarSystemId" TEXT
						REFERENCES "SolarSystem"(id) ON DELETE SET NULL,
					ADD COLUMN IF NOT EXISTS "solarMonthlyKwh" DOUBLE PRECISION;

				CREATE INDEX IF NOT EXISTS "Booking_solarSystem_idx"
					ON "Booking" ("solarSystemId") WHERE "solarSystemId" IS NOT NULL;
			`,
		},
	}
}
