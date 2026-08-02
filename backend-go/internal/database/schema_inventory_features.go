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
		{
			// فهارس ناقصة على أعمدة تُستخدم بالربط/الفلترة/الترتيب — كانت تخلي
			// الأنظمة المدموجة (المشاريع، الجي بي اس، عروض الأسعار) تسوي مسح
			// كامل للجدول (Seq Scan) بكل استعلام.
			Version: "0150_perf_indexes_merged_modules",
			SQL: `CREATE INDEX IF NOT EXISTS "Project_createdAt_idx" ON "Project"("createdAt" DESC);
			CREATE INDEX IF NOT EXISTS "Project_bookingId_idx" ON "Project"("bookingId");
			CREATE INDEX IF NOT EXISTS "Quotation_createdByEmployeeId_idx" ON "Quotation"("createdByEmployeeId");
			CREATE INDEX IF NOT EXISTS "GpsRenewalRequest_deviceRequestId_idx" ON "GpsRenewalRequest"("deviceRequestId");
			CREATE INDEX IF NOT EXISTS "GpsDeviceRequest_simCardId_idx" ON "GpsDeviceRequest"("simCardId");
			CREATE INDEX IF NOT EXISTS "GpsDeviceRequest_employeeId_idx" ON "GpsDeviceRequest"("employeeId");
			CREATE INDEX IF NOT EXISTS "GpsDeviceRequest_assignedTechnicianId_idx" ON "GpsDeviceRequest"("assignedTechnicianId");
			CREATE INDEX IF NOT EXISTS "GpsMaintenanceRequest_employeeId_idx" ON "GpsMaintenanceRequest"("employeeId")`,
		},
		{
			// "شخصية مهمة" (VIP): أي موظف يثبّت حجز أو يتعامل مع زبون يقدر يعلّمه
			// بضغطة زر. السجل يحفظ مين علّمه وشنو طلب الزبون، والقائمة تُعرض
			// لمدير النظام فقط (GET محمي بـrequireAdmin).
			Version: "0151_create_vip_customer",
			SQL: `CREATE TABLE IF NOT EXISTS "VipCustomer" (
				id TEXT PRIMARY KEY,
				"customerId" TEXT NOT NULL UNIQUE REFERENCES "Customer"(id) ON DELETE CASCADE,
				"bookingId" TEXT REFERENCES "Booking"(id) ON DELETE SET NULL,
				"requestSummary" TEXT,
				note TEXT,
				"markedByEmployeeId" TEXT NOT NULL REFERENCES "Employee"(id),
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "VipCustomer_markedByEmployeeId_idx" ON "VipCustomer"("markedByEmployeeId")`,
		},
		{
			// رابط موقع المورد (بديل عن التحديد اليدوي على الخريطة) + هل هو منافس.
			Version: "0152_add_supplier_location_url_and_competitor",
			SQL: `ALTER TABLE "Supplier" ADD COLUMN IF NOT EXISTS "locationUrl" TEXT;
			ALTER TABLE "Supplier" ADD COLUMN IF NOT EXISTS "isCompetitor" BOOLEAN NOT NULL DEFAULT false`,
		},
		{
			// فصل صلاحية الموردين عن صلاحية التقني: نمنح suppliers_management
			// تلقائياً لكل موظف عنده content_technician حالياً، حتى ما يخسر أحد
			// وصوله الي كان موجود قبل الفصل.
			Version: "0153_split_suppliers_permission",
			SQL: `INSERT INTO "Permission" (id, name, label)
			VALUES (gen_random_uuid()::text, 'suppliers_management', 'إدارة الموردين (إضافة وتعديل)')
			ON CONFLICT (name) DO NOTHING;
			INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			SELECT gen_random_uuid()::text, ep."employeeId", (SELECT id FROM "Permission" WHERE name = 'suppliers_management')
			FROM "EmployeePermission" ep
			JOIN "Permission" p ON p.id = ep."permissionId"
			WHERE p.name = 'content_technician'
			ON CONFLICT ("employeeId", "permissionId") DO NOTHING`,
		},
		{
			// سبب الطلب وشرحه، وسعر الشراء + ربط طلب المشتريات المتولّد لما
			// تكون الأداة مو متوفرة بالمخزن.
			Version: "0154_add_tool_request_reason_and_purchase",
			SQL: `ALTER TABLE "ToolRequest" ADD COLUMN IF NOT EXISTS reason TEXT;
			ALTER TABLE "ToolRequest" ADD COLUMN IF NOT EXISTS description TEXT;
			ALTER TABLE "ToolRequest" ADD COLUMN IF NOT EXISTS "purchasePrice" DOUBLE PRECISION;
			ALTER TABLE "ToolRequest" ADD COLUMN IF NOT EXISTS "procurementRequestId" TEXT`,
		},
		{
			// فصل موافقة طلبات الأدوات عن الدور الوظيفي الصارم. نمنحها تلقائياً
			// لكل إداري كميات (وهو المعني بالشغلة) ولكل موظف بدور كان يقدر
			// يوافق قبل الفصل، حتى ما ينقطع وصول أحد.
			Version: "0155_tool_requests_approve_permission",
			SQL: `INSERT INTO "Permission" (id, name, label)
			VALUES (gen_random_uuid()::text, 'tool_requests_approve', 'موافقة/رفض طلبات الأدوات')
			ON CONFLICT (name) DO NOTHING;
			INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			SELECT gen_random_uuid()::text, e.id, (SELECT id FROM "Permission" WHERE name = 'tool_requests_approve')
			FROM "Employee" e
			WHERE e.role IN ('PROCUREMENT_ADMIN', 'ADMIN', 'HR_COORDINATOR', 'MONITOR')
			ON CONFLICT ("employeeId", "permissionId") DO NOTHING`,
		},
		{
			// أدوات المركبات: كمية بدل باركود. مرات تكون نفس الأداة موجودة
			// مرتين بنفس السيارة، والباركود الفريد كان يمنع هذا أصلاً — فنخليه
			// اختياري ونشيل قيد التفرّد.
			Version: "0156_vehicle_tool_quantity",
			SQL: `ALTER TABLE "VehicleTool" ADD COLUMN IF NOT EXISTS quantity INTEGER NOT NULL DEFAULT 1;
			ALTER TABLE "VehicleTool" ALTER COLUMN barcode DROP NOT NULL;
			ALTER TABLE "VehicleTool" DROP CONSTRAINT IF EXISTS "VehicleTool_barcode_key"`,
		},
		{
			// تفاصيل تعبئة الوقود: اللترات، منو عبّأ، رقم الوصل، المحطة، وصورة
			// الوصل. الفهرس على (filledByEmployeeId, performedAt) يخدم إحصائية
			// "كم مرة عبّأ كل موظف بالشهر".
			Version: "0157_vehicle_log_fuel_details",
			SQL: `ALTER TABLE "VehicleLog" ADD COLUMN IF NOT EXISTS liters DOUBLE PRECISION;
			ALTER TABLE "VehicleLog" ADD COLUMN IF NOT EXISTS "filledByEmployeeId" TEXT;
			ALTER TABLE "VehicleLog" ADD COLUMN IF NOT EXISTS "receiptNumber" TEXT;
			ALTER TABLE "VehicleLog" ADD COLUMN IF NOT EXISTS "stationName" TEXT;
			ALTER TABLE "VehicleLog" ADD COLUMN IF NOT EXISTS "receiptPhotoBase64" TEXT;
			CREATE INDEX IF NOT EXISTS "VehicleLog_filledBy_performedAt_idx"
				ON "VehicleLog"("filledByEmployeeId", "performedAt")`,
		},
		{
			// سجل حركة الأداة الشخصية — يجاوب "متى انفقدت الأداة ومنو سجّلها".
			// بدون مفتاح خارجي على toolId عن قصد: السجل لازم يبقى حتى بعد حذف
			// الأداة، لأن توثيق الي راح هو بالضبط قيمته.
			Version: "0158_create_personal_tool_event",
			SQL: `CREATE TABLE IF NOT EXISTS "PersonalToolEvent" (
				id TEXT PRIMARY KEY,
				"toolId" TEXT NOT NULL,
				"toolName" TEXT NOT NULL,
				"employeeId" TEXT NOT NULL,
				"eventType" TEXT NOT NULL,
				"fromStatus" TEXT,
				"toStatus" TEXT,
				note TEXT,
				"actorId" TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "PersonalToolEvent_toolId_idx" ON "PersonalToolEvent"("toolId", "createdAt" DESC);
			CREATE INDEX IF NOT EXISTS "PersonalToolEvent_employeeId_idx" ON "PersonalToolEvent"("employeeId", "createdAt" DESC);
			CREATE INDEX IF NOT EXISTS "PersonalToolEvent_type_idx" ON "PersonalToolEvent"("eventType", "createdAt" DESC)`,
		},
		{
			// حالة الأداة enum بقيم AVAILABLE/CHECKED_OUT/DAMAGED فقط — نحتاج
			// LOST خصوصاً (السؤال الأساسي "متى انفقدت") و REPAIRING و RETIRED.
			// ADD VALUE ما ينفع داخل معاملة بنسخ قديمة، فنعملها وحدة وحدة
			// بـIF NOT EXISTS الي يخليها آمنة للإعادة.
			Version: "0159_extend_tool_status_enum",
			SQL: `ALTER TYPE "ToolStatus" ADD VALUE IF NOT EXISTS 'LOST';
			ALTER TYPE "ToolStatus" ADD VALUE IF NOT EXISTS 'REPAIRING';
			ALTER TYPE "ToolStatus" ADD VALUE IF NOT EXISTS 'RETIRED'`,
		},
		{
			// وقت تحويل الحجز لتنسيق الحجوزات (التثبيت) — كان ما ينحفظ إطلاقاً،
			// فما نقدر نجاوب "شوكت انحول هذا الحجز للتنسيق".
			Version: "0160_add_booking_confirmed_at",
			SQL: `ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "confirmedAt" TIMESTAMP;
			CREATE INDEX IF NOT EXISTS "Booking_confirmedAt_idx" ON "Booking"("confirmedAt" DESC)`,
		},
		{
			// الموقع بنفس آلية الموردين: عنوان كلامي + نقطة على الخريطة + رابط.
			// الرابط يغني عن التحديد اليدوي (السيرفر يفكّه لإحداثيات).
			Version: "0161_add_location_url_project_booking",
			SQL: `ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "locationUrl" TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "locationUrl" TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "mapLatitude" DOUBLE PRECISION;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "mapLongitude" DOUBLE PRECISION;
			-- ملاحظة: العمودين فوق موجودين أصلاً بأغلب النسخ، IF NOT EXISTS يخليها آمنة`,
		},
		{
			// سياسة الخصوصية: نقاط يضيفها صاحب الصلاحية، وكل موظف يوافق عليها
			// أول دخول. نحتفظ بمنو أضاف كل نقطة (يشوفه المالك ومدير النظام).
			Version: "0162_create_privacy_policy",
			SQL: `INSERT INTO "Permission" (id, name, label)
			VALUES (gen_random_uuid()::text, 'privacy_policy_manage', 'إضافة وتعديل سياسات الخصوصية')
			ON CONFLICT (name) DO NOTHING;

			CREATE TABLE IF NOT EXISTS "PrivacyPolicyPoint" (
				id TEXT PRIMARY KEY,
				content TEXT NOT NULL,
				"order" INTEGER NOT NULL DEFAULT 0,
				"isActive" BOOLEAN NOT NULL DEFAULT true,
				"createdByEmployeeId" TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now(),
				"updatedAt" TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "PrivacyPolicyPoint_order_idx" ON "PrivacyPolicyPoint"("order", "createdAt");

			-- موافقة الموظف: نخزن عدد النقاط وقت الموافقة حتى إذا انضافت نقاط
			-- جديدة بعدين تنطلب موافقة جديدة تلقائياً.
			CREATE TABLE IF NOT EXISTS "PrivacyPolicyAcceptance" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL UNIQUE,
				"acceptedAt" TIMESTAMP NOT NULL DEFAULT now(),
				"pointsVersion" INTEGER NOT NULL DEFAULT 0
			)`,
		},
		{
			// صلاحية ظهور مستقلة لكل وحدة — تُمنح لأي موظف من صفحة الصلاحيات.
			Version: "0163_unit_visibility_permissions",
			SQL: `INSERT INTO "Permission" (id, name, label) VALUES
			(gen_random_uuid()::text, 'unit_service', 'وحدة الخدمة'),
			(gen_random_uuid()::text, 'unit_design', 'وحدة التصميم'),
			(gen_random_uuid()::text, 'unit_pr', 'وحدة الإعلام والعلاقات العامة'),
			(gen_random_uuid()::text, 'unit_quality', 'وحدة الجودة والسلامة المهنية'),
			(gen_random_uuid()::text, 'unit_monitoring', 'وحدة الرقابة'),
			(gen_random_uuid()::text, 'unit_procurement', 'وحدة المشتريات والمخازن'),
			(gen_random_uuid()::text, 'unit_finance', 'وحدة الحسابات'),
			(gen_random_uuid()::text, 'unit_hr', 'وحدة الكوادر التنفيذية'),
			(gen_random_uuid()::text, 'unit_projects', 'وحدة إدارة المشاريع')
			ON CONFLICT (name) DO NOTHING`,
		},
		{
			// منو أضاف المشروع (أو رحّل الحجز لإدارة المشاريع) — يظهر ببطاقة
			// المشروع مكان المرحلة، والمرحلة انتقلت لأعلى البطاقة.
			Version: "0164_add_project_created_by",
			SQL:     `ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "createdByEmployeeId" TEXT`,
		},
		{
			// تسليم المشروع لموظف: الموظف المُسلَّم إله يشوف المشروع كامل
			// بكل مراحله ويتحكم بيه — كأنه عنده إدارة مشاريع بس على هذا
			// المشروع لحاله، بدون ما ننطيه الصلاحية العامة.
			Version: "0165_project_delegation",
			SQL: `
				ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "delegatedToEmployeeId" TEXT;
				ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "delegatedByEmployeeId" TEXT;
				ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "delegatedAt" TIMESTAMPTZ;
				CREATE INDEX IF NOT EXISTS "Project_delegatedTo_idx" ON "Project" ("delegatedToEmployeeId");
				CREATE TABLE IF NOT EXISTS "ProjectDelegationLog" (
					id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
					"projectId" TEXT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
					"employeeId" TEXT NOT NULL,
					"delegatedByEmployeeId" TEXT,
					action TEXT NOT NULL DEFAULT 'ASSIGN',
					note TEXT,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "ProjectDelegationLog_project_idx" ON "ProjectDelegationLog" ("projectId");
				CREATE INDEX IF NOT EXISTS "ProjectDelegationLog_employee_idx" ON "ProjectDelegationLog" ("employeeId");
			`,
		},
		{
			// خدمات متعددة للحجز الواحد: الزبون ممكن يطلب منظومة صوت وكاميرات
			// بنفس الوقت. عمود serviceId القديم يبقى للخدمة الرئيسية (توافق مع
			// كل الشاشات والتقارير الحالية)، وهذا الجدول يضيف البقية.
			Version: "0166_booking_multiple_services",
			SQL: `
				CREATE TABLE IF NOT EXISTS "BookingService" (
					id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
					"bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
					"serviceId" TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
					UNIQUE ("bookingId", "serviceId")
				);
				CREATE INDEX IF NOT EXISTS "BookingService_booking_idx" ON "BookingService" ("bookingId");
				-- نملي الجدول من الخدمة المفردة الموجودة حالياً حتى الحجوزات
				-- القديمة تطلع بنفس الشكل الجديد بدون فرق
				INSERT INTO "BookingService" ("bookingId", "serviceId")
				SELECT id, "serviceId" FROM "Booking"
				WHERE "serviceId" IS NOT NULL
				ON CONFLICT DO NOTHING;
			`,
		},
		{
			// الحظر التلقائي: الحساب ينحظر تلقائياً عند تكرار كلمة سر غلط أو
			// محاولة الوصول لعملية مو مخوّل لها، ويضل محظور لحد ما المالك
			// يفعّله. نخزن السبب والوقت حتى تطلع بلوحة المراقبة.
			Version: "0167_auto_lockout",
			SQL: `
				ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "failedLoginStreak" INTEGER NOT NULL DEFAULT 0;
				ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "lockedAt" TIMESTAMPTZ;
				ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "lockedReason" TEXT;
				ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "lockedDetail" TEXT;
				CREATE TABLE IF NOT EXISTS "SecurityEvent" (
					id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
					"employeeId" TEXT,
					"employeeName" TEXT,
					kind TEXT NOT NULL,
					detail TEXT,
					ip TEXT,
					"userAgent" TEXT,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "SecurityEvent_createdAt_idx" ON "SecurityEvent" ("createdAt" DESC);
				CREATE INDEX IF NOT EXISTS "SecurityEvent_employee_idx" ON "SecurityEvent" ("employeeId");
			`,
		},
		{
			// إبطال الجلسات: نخزن لحظة "آخر إبطال" لكل موظف. أي توكن صدر قبلها
			// يصير غير صالح فوراً — بدونها كان التوكن المسروق يضل شغّال 12
			// ساعة حتى لو الموظف غيّر كلمة سره أو المالك حظره.
			Version: "0168_session_invalidation",
			SQL: `
				ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "sessionsInvalidatedAt" TIMESTAMPTZ;
			`,
		},
		{
			// العدة القياسية كانت تنطبق على "كل" موظف بلا استثناء، فطلع موظف
			// مبيعات معلّق برقبته ٣٩ أداة وإداري كوادر ٤١ ومدقق ٤١ — وهذول
			// أصلاً ما عندهم عدة وما يتحاسبون عليها.
			//
			// هنا ننظّف الموجود: نشيل عدة كل موظف مو مستحق (مو فني ولا ليدر).
			// نسجّل حدث حذف لكل أداة قبل ما نشيلها، لأن PersonalToolEvent مصمّم
			// أصلاً يبقى بعد حذف الأداة حتى يوثّق الي راح — فيضل أثر للمراجعة.
			Version: "0169_tool_kit_role_scope",
			SQL: `
				INSERT INTO "PersonalToolEvent" (id, "toolId", "toolName", "employeeId", "eventType", note)
				SELECT gen_random_uuid()::text, p.id, p.name, p."employeeId", 'DELETED',
					'انشالت بترحيل 0169: دور الموظف ما يستحق عدة قياسية'
				FROM "PersonalTool" p
				JOIN "Employee" e ON e.id = p."employeeId"
				WHERE NOT (e.role = 'TECHNICIAN' OR e."isLeader" = true);

				DELETE FROM "PersonalTool" p
				USING "Employee" e
				WHERE e.id = p."employeeId"
				  AND NOT (e.role = 'TECHNICIAN' OR e."isLeader" = true);
			`,
		},
		{
			// دورة حياة شريحة الجي بي اس + متابعة التجديد.
			//
			// الشريحة كان عندها حالتين بس (متوفرة/مستخدمة) وما اكو طريقة
			// نعرف بيها شريحة تحتاج حرق ولا نحرّرها ترجع للمتوفر. أضفنا:
			//   NEEDS_BURN — الزبون رفض التجديد وخلصت مهلة الـ٨٠ يوم
			//   BURNED     — انحرقت فعلاً
			// والتحرير يرجّعها AVAILABLE ويفك ارتباطها بالزبون.
			//
			// وجدول متابعة الاتصالات: مهندس الجودة يتصل بالزبون بعد ٤٠ يوم
			// من انتهاء الاشتراك ويسجّل النتيجة، والنتيجة هي الي تقرر شنو
			// يصير بالشريحة بعدين.
			Version: "0170_gps_sim_lifecycle",
			SQL: `
				ALTER TYPE "SimStatus" ADD VALUE IF NOT EXISTS 'NEEDS_BURN';
				ALTER TYPE "SimStatus" ADD VALUE IF NOT EXISTS 'BURNED';

				ALTER TABLE "SimCard" ADD COLUMN IF NOT EXISTS "assignedAt" TIMESTAMPTZ;
				ALTER TABLE "SimCard" ADD COLUMN IF NOT EXISTS "releasedAt" TIMESTAMPTZ;
				ALTER TABLE "SimCard" ADD COLUMN IF NOT EXISTS "burnedAt" TIMESTAMPTZ;

				CREATE TABLE IF NOT EXISTS "GpsRenewalFollowUp" (
					id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
					"deviceRequestId" TEXT NOT NULL,
					"customerId" TEXT,
					"calledById" TEXT,
					outcome TEXT NOT NULL,
					notes TEXT,
					"daysSinceExpiry" INT,
					"calledAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "GpsRenewalFollowUp_device_idx"
					ON "GpsRenewalFollowUp" ("deviceRequestId", "calledAt" DESC);
			`,
		},
		{
			// الدوار: مبلغ دوّار للعمل يصرفه المحاسب للموظفين.
			//
			// الدورة: المحاسب يسلّم الموظف مبلغ → ينزل من رصيد الدوار ويطلع
			// دَين برقبة الموظف → الموظف يشتري ويرفع صورة الوصل ويرجّع الباقي
			// → المحاسب يدقّق ويوافق → وقتها بس رصيد الموظف يتصفّر ويرجع
			// المبلغ المرتجع للدوار.
			//
			// ليش ما نصفّر إلا بموافقة المحاسب؟ لأن التصفير بدون تدقيق يعني
			// أي موظف يقدر يعلن إنه صرف الفلوس ويطلع نظيف بدون ما أحد يشوف
			// الوصل. الموافقة هي كل قيمة النظام هنا.
			Version: "0171_revolving_fund",
			SQL: `
				CREATE TABLE IF NOT EXISTS "RevolvingFund" (
					id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
					name TEXT NOT NULL UNIQUE,
					balance NUMERIC(14,2) NOT NULL DEFAULT 0,
					"isActive" BOOLEAN NOT NULL DEFAULT true,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
					"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);

				INSERT INTO "RevolvingFund" (id, name) VALUES
					(gen_random_uuid()::text, 'دوار الطاقة الشمسية'),
					(gen_random_uuid()::text, 'دوار الشعبة الهندسية')
				ON CONFLICT (name) DO NOTHING;

				-- حركة الدوار: صرف للموظف، أو تسوية يرفعها الموظف، أو
				-- تغذية للدوار نفسه من المحاسب.
				CREATE TABLE IF NOT EXISTS "RevolvingFundTxn" (
					id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
					"fundId" TEXT NOT NULL REFERENCES "RevolvingFund"(id),
					"employeeId" TEXT,
					kind TEXT NOT NULL,              -- DISBURSE | SETTLEMENT | TOPUP
					amount NUMERIC(14,2) NOT NULL DEFAULT 0,        -- المبلغ المسلَّم (بالصرف)
					"spentAmount" NUMERIC(14,2) NOT NULL DEFAULT 0, -- المصروف بالتسوية
					"returnedAmount" NUMERIC(14,2) NOT NULL DEFAULT 0, -- المرتجع بالتسوية
					"bookingId" TEXT,
					"receiptImage" TEXT,             -- صورة الوصل (base64)
					notes TEXT,
					status TEXT NOT NULL DEFAULT 'APPROVED', -- التسوية تبدي PENDING
					"createdById" TEXT,
					"reviewedById" TEXT,
					"reviewedAt" TIMESTAMPTZ,
					"reviewNote" TEXT,
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX IF NOT EXISTS "RevolvingFundTxn_employee_idx"
					ON "RevolvingFundTxn" ("employeeId", "createdAt" DESC);
				CREATE INDEX IF NOT EXISTS "RevolvingFundTxn_status_idx"
					ON "RevolvingFundTxn" (status, "createdAt" DESC);

				INSERT INTO "Permission" (id, name, label) VALUES
					(gen_random_uuid()::text, 'revolving_fund', 'الدوار (صرف وتدقيق المبالغ الدوّارة)')
				ON CONFLICT (name) DO NOTHING;
			`,
		},
	}
}
