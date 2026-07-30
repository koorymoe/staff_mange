package database

// inventoryFeaturesVersionedMigrations يرجّع الترحيلات المرقّمة لميزات المخزون
// المضافة بهذه الجلسة: (1) "العدة القياسية" (PersonalToolTemplateItem) — قائمة
// رئيسية بأسماء الأدوات الشخصية الي كل موظف لازم يكون عنده إياها، تُطبَّق تلقائياً
// على الموظفين الحاليين والجدد، و(2) VehicleToolCheck — لقطة الأدوات العامة
// الناقصة بالمركبة عند بدء مهمة من قبل ليدر.
func inventoryFeaturesVersionedMigrations() []Migration {
	return []Migration{
		{
			Version: "0133_create_personal_tool_template_item",
			SQL: `CREATE TABLE IF NOT EXISTS "PersonalToolTemplateItem" (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		{
			Version: "0134_create_vehicle_tool_check",
			SQL: `CREATE TABLE IF NOT EXISTS "VehicleToolCheck" (
				id TEXT PRIMARY KEY,
				"vehicleId" TEXT NOT NULL REFERENCES "Vehicle"(id) ON DELETE CASCADE,
				"missionId" TEXT NOT NULL REFERENCES "VehicleMission"(id) ON DELETE CASCADE,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				"missingToolNames" TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "VehicleToolCheck_vehicleId_idx" ON "VehicleToolCheck"("vehicleId");
			CREATE INDEX IF NOT EXISTS "VehicleToolCheck_missionId_idx" ON "VehicleToolCheck"("missionId")`,
		},
		{
			// رسالة بيانات دخول الزبون (يوزر + باسورد بنص حر) — لازم تنكتب قبل
			// ما يكدر الإداري يفعّل الجهاز، مع تاريخ التفعيل.
			Version: "0135_add_gps_device_request_credentials_message",
			SQL:     `ALTER TABLE "GpsDeviceRequest" ADD COLUMN IF NOT EXISTS "credentialsMessage" TEXT`,
		},
		{
			// تتبع آخر موظف عدّل تفاصيل/تكليف الحجز (منفصل عن "من أكّد الحجز") —
			// حتى يطلع بصفحة تفاصيل الحجز "من عدّله" بعد "من أكّده".
			Version: "0136_add_booking_last_edited_by",
			SQL: `ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "lastEditedById" TEXT REFERENCES "Employee"(id);
				ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "lastEditedAt" TIMESTAMP`,
		},
		{
			// الكشوفات: فورمات فارغة يطبعها المهندس، يمليها يدوياً بالموقع، وبعدين
			// يرجع يرفع صور الفورمة المالية للنظام — مربوطة بمشروع اختياري.
			Version: "0137_create_project_checklist",
			SQL: `CREATE TABLE IF NOT EXISTS "ProjectChecklist" (
				id TEXT PRIMARY KEY,
				"projectId" TEXT REFERENCES "Project"(id),
				title TEXT NOT NULL,
				"createdById" TEXT NOT NULL REFERENCES "Employee"(id),
				"photoUrls" TEXT[] NOT NULL DEFAULT '{}',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "ProjectChecklist_projectId_idx" ON "ProjectChecklist"("projectId")`,
		},
		{
			// أيقونة الحضور الشخصية لكل موظف — تبقى فاضية (يستخدم النظام أول حرف
			// من الاسم كافتراضي) لين يطلب الموظف تغييرها ويوافق الإداري.
			Version: "0138b_add_employee_attendance_icon",
			SQL:     `ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "attendanceIcon" TEXT`,
		},
		{
			// طلبات تغيير أيقونة الحضور — الموظف يطلب رمز جديد، ومدير النظام يوافق
			// أو يرفض.
			Version: "0138c_create_attendance_icon_request",
			SQL: `CREATE TABLE IF NOT EXISTS "AttendanceIconRequest" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				"requestedIcon" TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'PENDING',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"resolvedAt" TIMESTAMP,
				"resolvedById" TEXT REFERENCES "Employee"(id)
			);
			CREATE INDEX IF NOT EXISTS "AttendanceIconRequest_status_idx" ON "AttendanceIconRequest"(status)`,
		},
		{
			// معرض أعمال التقنيين: نماذج أعمال وأفكار وتصاميم جديدة يرفعها الفنيون
			// بأنفسهم (منفصل عن مواد التدريب الي يديرها الإداري للمتدربين).
			Version: "0138_create_tech_showcase_item",
			SQL: `CREATE TABLE IF NOT EXISTS "TechShowcaseItem" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				title TEXT NOT NULL,
				description TEXT,
				"mediaUrls" TEXT[] NOT NULL DEFAULT '{}',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "TechShowcaseItem_employeeId_idx" ON "TechShowcaseItem"("employeeId")`,
		},
		{
			// إدارة المعارض (وحدة التقنيين) — معرض تجاري تحضره الشركة: شركات
			// حاضرة، منتجات معروضة، ترشيح المدير لمين يروح، صور كارتات بزنس،
			// وتقرير زيارة يُولَّد بالذكاء الصناعي بعد الأرشفة.
			Version: "0139_create_exhibition",
			SQL: `CREATE TABLE IF NOT EXISTS "Exhibition" (
				id TEXT PRIMARY KEY,
				title TEXT NOT NULL,
				location TEXT NOT NULL,
				"startDate" TEXT NOT NULL,
				"endDate" TEXT NOT NULL,
				companies TEXT[] NOT NULL DEFAULT '{}',
				"productsToShow" TEXT[] NOT NULL DEFAULT '{}',
				"nominatedEmployeeIds" TEXT[] NOT NULL DEFAULT '{}',
				"businessCardPhotos" TEXT[] NOT NULL DEFAULT '{}',
				"keyFindings" TEXT,
				"visitReport" TEXT,
				archived BOOLEAN NOT NULL DEFAULT false,
				"createdById" TEXT NOT NULL REFERENCES "Employee"(id),
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "Exhibition_archived_idx" ON "Exhibition"(archived)`,
		},
		{
			// إدارة المنتجات (وحدة التقنيين) — اقتراح منتج جديد يُضاف للكتالوج،
			// يفتحه المدير أو التقني أو مسؤول المشتريات، ويوافق/يرفض المدير.
			Version: "0140_create_product_request",
			SQL: `CREATE TABLE IF NOT EXISTS "ProductRequest" (
				id TEXT PRIMARY KEY,
				"requestedById" TEXT NOT NULL REFERENCES "Employee"(id),
				"productName" TEXT NOT NULL,
				specs TEXT,
				source TEXT,
				model TEXT,
				category TEXT,
				price DOUBLE PRECISION,
				status TEXT NOT NULL DEFAULT 'PENDING',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"resolvedAt" TIMESTAMP,
				"resolvedById" TEXT REFERENCES "Employee"(id)
			);
			CREATE INDEX IF NOT EXISTS "ProductRequest_status_idx" ON "ProductRequest"(status)`,
		},
		{
			// إدارة الخدمات (وحدة التقنيين) — خدمة جديدة مقترحة تحتاج دراسة: المدير
			// يوكّل تقني/تقنيين محددين، وكل موكَّل يرفع تقارير/دراسات تُؤرشف.
			Version: "0141_create_service_study",
			SQL: `CREATE TABLE IF NOT EXISTS "ServiceStudy" (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				"createdById" TEXT NOT NULL REFERENCES "Employee"(id),
				archived BOOLEAN NOT NULL DEFAULT false,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS "ServiceStudyAssignment" (
				id TEXT PRIMARY KEY,
				"serviceStudyId" TEXT NOT NULL REFERENCES "ServiceStudy"(id),
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UNIQUE("serviceStudyId", "employeeId")
			);
			CREATE TABLE IF NOT EXISTS "ServiceStudyReport" (
				id TEXT PRIMARY KEY,
				"serviceStudyId" TEXT NOT NULL REFERENCES "ServiceStudy"(id),
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				content TEXT NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "ServiceStudyAssignment_serviceStudyId_idx" ON "ServiceStudyAssignment"("serviceStudyId");
			CREATE INDEX IF NOT EXISTS "ServiceStudyReport_serviceStudyId_idx" ON "ServiceStudyReport"("serviceStudyId")`,
		},
		{
			// وحدة التصميم: بنّاء أسئلة قابل للتخصيص كلياً — المدير يضيف أسئلة
			// استمارة طلب التصميم بنفسه يدوياً (نص/رقم/تاريخ/اختيار من متعدد/
			// خيارات متعددة/ملف)، النظام بوفّر آلية الإضافة/الترتيب بس، مو محتوى
			// استمارة جاهزة مكتوبة بالكود.
			Version: "0142_create_design_form_question",
			SQL: `CREATE TABLE IF NOT EXISTS "DesignFormQuestion" (
				id TEXT PRIMARY KEY,
				label TEXT NOT NULL,
				type TEXT NOT NULL,
				options TEXT[] NOT NULL DEFAULT '{}',
				required BOOLEAN NOT NULL DEFAULT false,
				"order" INT NOT NULL DEFAULT 0,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "DesignFormQuestion_order_idx" ON "DesignFormQuestion"("order")`,
		},
		{
			// عدة استمارات تصميم منفصلة (بدل استمارة وحدة عامة) — كل استمارة اسمها
			// الخاص ورابط عام (publicToken) نرسله للزبون مباشرة بدون تسجيل دخول،
			// وأسئلتها منفصلة عن استمارات ثانية عبر formId. الأجوبة المستلمة
			// تنخزن بـDesignFormSubmission كـJSON (سؤال → جواب).
			Version: "0143_create_design_form_and_submissions",
			SQL: `CREATE TABLE IF NOT EXISTS "DesignForm" (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				"publicToken" TEXT NOT NULL UNIQUE,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			ALTER TABLE "DesignFormQuestion" ADD COLUMN IF NOT EXISTS "formId" TEXT REFERENCES "DesignForm"(id) ON DELETE CASCADE;
			CREATE INDEX IF NOT EXISTS "DesignFormQuestion_formId_idx" ON "DesignFormQuestion"("formId");
			CREATE TABLE IF NOT EXISTS "DesignFormSubmission" (
				id TEXT PRIMARY KEY,
				"formId" TEXT NOT NULL REFERENCES "DesignForm"(id) ON DELETE CASCADE,
				answers JSONB NOT NULL DEFAULT '{}',
				"submittedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "DesignFormSubmission_formId_idx" ON "DesignFormSubmission"("formId")`,
		},
		{
			// اعتماد فاتورة الليدر: تبقى الفاتورة ظاهرة عند الليدر بحالة SUBMITTED
			// لين مدير/محاسب (requireFinance) يعتمدها لـAPPROVED — الليدر نفسه ما
			// يقدر يعتمد فاتورته بنفسه (الراوت محمي بـrequireFinance بالباك اند).
			Version: "0144_add_leader_invoice_approval",
			SQL: `ALTER TABLE "LeaderInvoice" ADD COLUMN IF NOT EXISTS "approvedByEmployeeId" TEXT REFERENCES "Employee"(id);
			ALTER TABLE "LeaderInvoice" ADD COLUMN IF NOT EXISTS "approvedAt" TIMESTAMP`,
		},
		{
			// أنواع الأعمال ("نوع العمل" بحقل المشروع) كانت قائمة ثابتة مكتوبة
			// بالكود (WORK_TYPES بـProjectsPage.tsx) — صارت جدول قابل للإضافة/الحذف
			// من إعدادات وحدة إدارة المشاريع، مع زرع نفس القيم القديمة كبداية حتى
			// ما ينكسر أي مشروع موجود يشاور عليهن.
			Version: "0145_create_project_work_type",
			SQL: `CREATE TABLE IF NOT EXISTS "ProjectWorkType" (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL UNIQUE,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			INSERT INTO "ProjectWorkType" (id, name) VALUES
				(gen_random_uuid()::text, 'طاقة شمسية'),
				(gen_random_uuid()::text, 'كاميرات'),
				(gen_random_uuid()::text, 'بيت ذكي'),
				(gen_random_uuid()::text, 'شبكات'),
				(gen_random_uuid()::text, 'إنذار حريق'),
				(gen_random_uuid()::text, 'أقفال وحاكيات'),
				(gen_random_uuid()::text, 'ستلايت'),
				(gen_random_uuid()::text, 'منظومة صوت'),
				(gen_random_uuid()::text, 'أخرى')
			ON CONFLICT (name) DO NOTHING`,
		},
		{
			// موقع المشروع صار نص عنوان + إحداثيات دقيقة (نفس أسلوب حجز جديد
			// LocationPicker)، وأضفنا مرحلة "العقد" بين عرض السعر والتنفيذ — العقد
			// يترفع كـPDF قبل التوقيع وبعده، ويتخزن مع المشروع نفسه (نفس أسلوب
			// تخزين الصور base64 بالمشروع/المركبة).
			Version: "0146_add_project_map_and_contract",
			SQL: `ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "mapLatitude" DOUBLE PRECISION;
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "mapLongitude" DOUBLE PRECISION;
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "contractPdfBase64" TEXT;
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "signedContractPdfBase64" TEXT`,
		},
		{
			// تحديد المسؤول عن المشروع ومنفّذ الكشف من قائمة منسدلة (موظفين عندهم
			// صلاحية إدارة المشاريع) بدل كتابة الأسماء يدوياً — ويجوز يكون نفس
			// الشخص للاثنين.
			Version: "0147_add_project_responsible_and_surveyor",
			SQL: `ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "responsibleEmployeeId" TEXT REFERENCES "Employee"(id);
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "surveyorEmployeeId" TEXT REFERENCES "Employee"(id)`,
		},
		{
			// إلغاء خاصية "نسخ البيانات للجروب" بالكامل — ما كانت مربوطة بأي خدمة
			// خارجية (كانت مجرد نسخ نص للحافظة + علامة بقاعدة البيانات)، فحذفناها
			// حتى ما تبقى قناة تسريب بيانات المشاريع خارج النظام.
			Version: "0148_drop_project_sent_to_group",
			SQL:     `ALTER TABLE "Project" DROP COLUMN IF EXISTS "sentToGroup"`,
		},
		{
			// عنوان المورد كلامي بجانب إحداثيات الخريطة — نفس أسلوب الحجوزات:
			// خانة عنوان نصي + نقطة محددة على الخريطة (lat/lng موجودين أصلاً).
			Version: "0149_add_supplier_address",
			SQL:     `ALTER TABLE "Supplier" ADD COLUMN IF NOT EXISTS address TEXT`,
		},
	}
}
