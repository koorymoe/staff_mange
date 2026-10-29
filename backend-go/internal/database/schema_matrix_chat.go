package database

// ═══ دردشة ماتركس (قرار (ع) 10-10) ═══
// المالك والمدير يسولفون ويا ماتركس (هايكو). الأسماء تنحفظ حقيقية هنا،
// وتتبدّل برموز بس وقت ما تطلع للنموذج.
func matrixChatMigrations() []Migration {
	return []Migration{{
		Version: "0344_matrix_talk",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixTalk" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				title TEXT NOT NULL DEFAULT 'محادثة جديدة',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "MatrixTalk_employee" ON "MatrixTalk"("employeeId", "updatedAt" DESC);
			CREATE TABLE IF NOT EXISTS "MatrixTalkMessage" (
				id TEXT PRIMARY KEY,
				"chatId" TEXT NOT NULL REFERENCES "MatrixTalk"(id) ON DELETE CASCADE,
				role TEXT NOT NULL CHECK (role IN ('USER', 'ASSISTANT')),
				text TEXT NOT NULL,
				steps TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS "MatrixTalkMessage_chat" ON "MatrixTalkMessage"("chatId", "createdAt");
		`,
	}, {
		// قرار (ع) 10-10: فواتير الشغل داخل الشركة تنعتمد بلا رقم فاتورة محاسبية
		Version: "0345_internal_invoice_no_number",
		SQL: `
			ALTER TABLE "LeaderInvoice" DROP CONSTRAINT IF EXISTS leader_invoice_approved_needs_number;
			ALTER TABLE "LeaderInvoice" ADD CONSTRAINT leader_invoice_approved_needs_number CHECK (
				status <> 'APPROVED'
				OR ("externalInvoiceNumber" IS NOT NULL AND btrim("externalInvoiceNumber") <> '')
				OR "isFree" OR COALESCE("auditVerdict", '') = 'FREE'
				OR systems::text LIKE '%شغل داخل الشركة%'
			) NOT VALID;
		`,
	}, {
		// شكوى (ع) 10-10: حجوزات انثبّتت وهي بالانتظار وبقى بيها waitingSince،
		// فعالقة بـ«بانتظار موافقة الزبون/ما رد» ومخفية عن الحجوزات
		Version: "0346_clear_stale_waiting",
		SQL: `
			UPDATE "Booking" SET "waitingSince" = NULL, "waitingNote" = NULL, "waitingById" = NULL,
				"waitingKind" = NULL, "lastWaitingReminderAt" = NULL, "waitingReminderCount" = 0
			WHERE "waitingSince" IS NOT NULL AND status NOT IN ('WAITING', 'CANCELLED');
		`,
	}, {
		// (ع) 10-10 «النظام صار ثكيل»: فهارس للاستعلامات الي تنضرب كل دقيقة
		Version: "0347_hot_path_indexes",
		SQL: `
			CREATE INDEX IF NOT EXISTS "Notification_unread" ON "Notification"("employeeId") WHERE read = false;
			CREATE INDEX IF NOT EXISTS "Booking_waitingSince" ON "Booking"("waitingSince") WHERE "waitingSince" IS NOT NULL;
			CREATE INDEX IF NOT EXISTS "Booking_status_scheduledAt" ON "Booking"(status, "scheduledAt");
			CREATE INDEX IF NOT EXISTS "Booking_status_createdAt" ON "Booking"(status, "createdAt" DESC);
			ANALYZE "Booking";
			ANALYZE "Notification";
		`,
	}, {
		// (ع) 10-11: سكرين بالدردشة + ماتركس يتعلّم من تصحيح المالك
		Version: "0348_matrix_chat_image_lessons",
		SQL: `
			ALTER TABLE "MatrixTalkMessage" ADD COLUMN IF NOT EXISTS image TEXT;
			CREATE TABLE IF NOT EXISTS "MatrixLesson" (
				id TEXT PRIMARY KEY,
				text TEXT NOT NULL,
				context TEXT,
				"createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				active BOOLEAN NOT NULL DEFAULT true,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`,
	}, {
		// (ع) 10-11: «راقب فلان» — مهمة مراقبة بيومية وخلاصة
		Version: "0349_matrix_missions",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixMission" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"createdById" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				note TEXT,
				"startsOn" DATE NOT NULL,
				"endsOn" DATE NOT NULL,
				active BOOLEAN NOT NULL DEFAULT true,
				summary TEXT,
				"summaryAt" TIMESTAMP,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS "MatrixMissionLog" (
				id TEXT PRIMARY KEY,
				"missionId" TEXT NOT NULL REFERENCES "MatrixMission"(id) ON DELETE CASCADE,
				day DATE NOT NULL,
				text TEXT NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UNIQUE ("missionId", day)
			);
		`,
	}, {
		// (ع) 10-11: صاحب «إضافة موظف»/«تعديل كل الموظفين» يشتغل بس على
		// الأدوار الي المالك يحددها إله — بلاها يكدر يفتح حساب محاسب
		Version: "0350_employee_admin_scope",
		SQL: `
			CREATE TABLE IF NOT EXISTS "EmployeeAdminScope" (
				"employeeId" TEXT PRIMARY KEY REFERENCES "Employee"(id) ON DELETE CASCADE,
				roles TEXT[] NOT NULL DEFAULT '{}',
				"updatedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`,
	}, {
		// فحص 10-11: فواتير انسحب اعتمادها ورجعت انعتمدت كانت مختفية من
		// المجاميع (revokedAt بقى) — نرجّعها؛ ومبالغ تسوية الدوار ما تكون سالبة
		Version: "0351_reapproved_invoices_fund_nonneg",
		SQL: `
			UPDATE "LeaderInvoice" SET "revokedAt" = NULL WHERE status = 'APPROVED' AND "revokedAt" IS NOT NULL;
			ALTER TABLE "RevolvingFundTxn" DROP CONSTRAINT IF EXISTS fund_txn_nonneg;
			ALTER TABLE "RevolvingFundTxn" ADD CONSTRAINT fund_txn_nonneg
				CHECK (COALESCE("spentAmount", 0) >= 0 AND COALESCE("returnedAmount", 0) >= 0) NOT VALID;
		`,
	}, {
		// (ع) 10-13: «رسول ماتركس» — المالك/المدير يوجّه شغل لموظف، وعين
		// الشارنغان تروح تسلّمه الرسالة على شاشته وترجع. السطر يبقى حتى
		// الموظف يرجع يقراها من عينه، و"seenAt" يصفّر رقم البند بالقائمة.
		Version: "0352_matrix_courier",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MatrixCourier" (
				id TEXT PRIMARY KEY,
				kind TEXT NOT NULL,
				"senderId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"recipientId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				message TEXT NOT NULL,
				link TEXT NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"deliveredAt" TIMESTAMP,
				"seenAt" TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS matrix_courier_recipient ON "MatrixCourier" ("recipientId", "createdAt" DESC);
			CREATE INDEX IF NOT EXISTS matrix_courier_sender ON "MatrixCourier" ("senderId", "createdAt" DESC);
		`,
	}, {
		// (ع) 10-13: «طلب تبرير» من المراقب يروح للموظف نفسه، يكتب عذره،
		// ويرجع للمراقب يحكم (مقبول/مرفوض) — والنتيجة تنحفظ وتوصل للموظف.
		Version: "0353_monitor_justification",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MonitorJustification" (
				id TEXT PRIMARY KEY,
				"reviewId" TEXT,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"monitorId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				question TEXT NOT NULL,
				answer TEXT,
				"answeredAt" TIMESTAMP,
				verdict TEXT,
				"verdictNote" TEXT,
				"decidedAt" TIMESTAMP,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS monitor_justification_emp ON "MonitorJustification" ("employeeId", "createdAt" DESC);
		`,
	}, {
		// (ع) 10-13: مركز التدريب — المنهج لكل دور، بنك أسئلة (يدوي + ماتركس
		// يقترح والمدرب يعتمد)، اختبار شهري لكل دور يسويه الموظف من حسابه،
		// خطة شهرية ويا تفتيش، ومشرف لكل موظف جديد. بلا خصم نقاط.
		Version: "0354_training_center",
		SQL: `
			CREATE TABLE IF NOT EXISTS "TrainingCurriculum" (
				id TEXT PRIMARY KEY, role TEXT NOT NULL, title TEXT NOT NULL, details TEXT,
				ord INTEGER NOT NULL DEFAULT 0, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE TABLE IF NOT EXISTS "TrainingQuestion" (
				id TEXT PRIMARY KEY, role TEXT NOT NULL, "curriculumId" TEXT REFERENCES "TrainingCurriculum"(id) ON DELETE SET NULL,
				text TEXT NOT NULL, options TEXT[] NOT NULL, correct INTEGER NOT NULL,
				approved BOOLEAN NOT NULL DEFAULT false, source TEXT NOT NULL DEFAULT 'MANUAL',
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE TABLE IF NOT EXISTS "TrainingExam" (
				id TEXT PRIMARY KEY, role TEXT NOT NULL, month TEXT NOT NULL, title TEXT NOT NULL,
				"questionIds" TEXT[] NOT NULL, minutes INTEGER NOT NULL DEFAULT 15, "passPct" INTEGER NOT NULL DEFAULT 60,
				"createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE (role, month));
			CREATE TABLE IF NOT EXISTS "TrainingAttempt" (
				id TEXT PRIMARY KEY, "examId" TEXT NOT NULL REFERENCES "TrainingExam"(id) ON DELETE CASCADE,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				answers INTEGER[], score INTEGER, total INTEGER,
				"startedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, "finishedAt" TIMESTAMP,
				UNIQUE ("examId", "employeeId"));
			CREATE TABLE IF NOT EXISTS "TrainingPlan" (
				id TEXT PRIMARY KEY, month TEXT NOT NULL, title TEXT NOT NULL, role TEXT,
				"employeeIds" TEXT[] NOT NULL DEFAULT '{}', "plannedDate" DATE,
				"trainerId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"inspectorId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"inspectionResult" TEXT, "inspectionNotes" TEXT, "inspectedAt" TIMESTAMP,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE TABLE IF NOT EXISTS "TrainingMentor" (
				"employeeId" TEXT PRIMARY KEY REFERENCES "Employee"(id) ON DELETE CASCADE,
				"mentorId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				notes TEXT, "updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
		`,
	}, {
		// (ع) 10-14: كل عرض سعر بشروطه وأحكامه — الفارغة (القديمة) تاخذ الافتراضية
		Version: "0355_quotation_terms",
		SQL:     `ALTER TABLE "Quotation" ADD COLUMN IF NOT EXISTS terms JSONB;`,
	}, {
		// (ع) 10-14: تطبيق الزبائن — حساب يسويه الزبون بنفسه (اسم مستخدم + تأكيد إيميل)،
		// محادثة ماتركس، محتوى ترويجي، ومصدر الحجز (ماتركس من التطبيق).
		Version: "0356_customer_app",
		SQL: `
			CREATE TABLE IF NOT EXISTS "CustomerAccount" (
				id TEXT PRIMARY KEY, username TEXT NOT NULL UNIQUE, email TEXT NOT NULL UNIQUE,
				"passwordHash" TEXT NOT NULL, name TEXT NOT NULL, phone TEXT NOT NULL,
				"customerId" TEXT REFERENCES "Customer"(id) ON DELETE SET NULL,
				"emailVerifiedAt" TIMESTAMP, "sessionsInvalidatedAt" TIMESTAMP,
				"lastSeenPromoAt" TIMESTAMP, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE TABLE IF NOT EXISTS "CustomerOtp" (
				id TEXT PRIMARY KEY, "accountId" TEXT NOT NULL REFERENCES "CustomerAccount"(id) ON DELETE CASCADE,
				purpose TEXT NOT NULL, "codeHash" TEXT NOT NULL, attempts INTEGER NOT NULL DEFAULT 0,
				"expiresAt" TIMESTAMP NOT NULL, "usedAt" TIMESTAMP, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE INDEX IF NOT EXISTS customer_otp_acc ON "CustomerOtp" ("accountId", "createdAt" DESC);
			CREATE TABLE IF NOT EXISTS "CustomerChatMessage" (
				id TEXT PRIMARY KEY, "accountId" TEXT NOT NULL REFERENCES "CustomerAccount"(id) ON DELETE CASCADE,
				role TEXT NOT NULL, content TEXT NOT NULL, "bookingId" TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE INDEX IF NOT EXISTS customer_chat_acc ON "CustomerChatMessage" ("accountId", "createdAt");
			CREATE TABLE IF NOT EXISTS "PromoItem" (
				id TEXT PRIMARY KEY, kind TEXT NOT NULL DEFAULT 'offer', title TEXT NOT NULL, body TEXT,
				"imageUrl" TEXT, "serviceId" TEXT REFERENCES "Service"(id) ON DELETE SET NULL,
				active BOOLEAN NOT NULL DEFAULT false, ord INTEGER NOT NULL DEFAULT 0,
				source TEXT NOT NULL DEFAULT 'manual', "approvedAt" TIMESTAMP,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS source TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "customerAccountId" TEXT;
		`,
	}, {
		// (ع) 10-14: تطبيق الزبائن بالتصميم الجديد — بطاقة كل خدمة (صورة، وصف،
		// مميزات، يبدأ من)، خصم العروض، المفضلة، تقييم الزبون، صور الطلب، والفروع.
		Version: "0357_customer_app_catalog",
		SQL: `
			CREATE TABLE IF NOT EXISTS "ServiceApp" (
				"serviceId" TEXT PRIMARY KEY REFERENCES "Service"(id) ON DELETE CASCADE,
				"imageUrl" TEXT, description TEXT, features JSONB NOT NULL DEFAULT '[]',
				"priceFrom" NUMERIC, visible BOOLEAN NOT NULL DEFAULT true, ord INTEGER NOT NULL DEFAULT 100,
				color TEXT, "updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			ALTER TABLE "PromoItem" ADD COLUMN IF NOT EXISTS "discountPercent" INTEGER;
			ALTER TABLE "CustomerAccount" ADD COLUMN IF NOT EXISTS city TEXT;
			CREATE TABLE IF NOT EXISTS "CustomerFavorite" (
				"accountId" TEXT NOT NULL REFERENCES "CustomerAccount"(id) ON DELETE CASCADE,
				"serviceId" TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("accountId", "serviceId"));
			CREATE TABLE IF NOT EXISTS "CustomerRating" (
				id TEXT PRIMARY KEY, "bookingId" TEXT NOT NULL UNIQUE REFERENCES "Booking"(id) ON DELETE CASCADE,
				"accountId" TEXT NOT NULL REFERENCES "CustomerAccount"(id) ON DELETE CASCADE,
				stars INTEGER NOT NULL CHECK (stars BETWEEN 1 AND 5), comment TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE TABLE IF NOT EXISTS "CustomerBookingPhoto" (
				id TEXT PRIMARY KEY, "bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
				key TEXT NOT NULL, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE TABLE IF NOT EXISTS "AppBranch" (
				id TEXT PRIMARY KEY, name TEXT NOT NULL, address TEXT, phone TEXT, "mapUrl" TEXT,
				ord INTEGER NOT NULL DEFAULT 0, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
		`,
	}, {
		// (ع) 10-14: «الآجل» — حجز أو مشروع خلص والفلوس مؤجلة (دفعات أو مبلغ كامل)
		Version: "0358_credit_sales",
		SQL: `
			CREATE TABLE IF NOT EXISTS "CreditSale" (
				id TEXT PRIMARY KEY, kind TEXT NOT NULL DEFAULT 'BOOKING',
				"bookingId" TEXT UNIQUE REFERENCES "Booking"(id) ON DELETE SET NULL,
				"projectId" TEXT REFERENCES "Project"(id) ON DELETE SET NULL,
				title TEXT NOT NULL, "customerName" TEXT, phone TEXT,
				amount NUMERIC NOT NULL CHECK (amount > 0), note TEXT, "dueDate" DATE,
				"createdById" TEXT, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, "settledAt" TIMESTAMP);
			CREATE TABLE IF NOT EXISTS "CreditPayment" (
				id TEXT PRIMARY KEY, "creditId" TEXT NOT NULL REFERENCES "CreditSale"(id) ON DELETE CASCADE,
				amount NUMERIC NOT NULL CHECK (amount > 0), "paidAt" DATE NOT NULL DEFAULT CURRENT_DATE,
				note TEXT, "byId" TEXT, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE INDEX IF NOT EXISTS credit_payment_credit ON "CreditPayment" ("creditId");
		`,
	}, {
		// 🔴 (ع) 10-14: اعتماد «الآجل» بلا رقم چان ينرفض من قيد القاعدة — نضيف CREDIT
		Version: "0359_credit_approve_no_number",
		SQL: `
			ALTER TABLE "LeaderInvoice" DROP CONSTRAINT IF EXISTS leader_invoice_approved_needs_number;
			ALTER TABLE "LeaderInvoice" ADD CONSTRAINT leader_invoice_approved_needs_number CHECK (
				status <> 'APPROVED'
				OR ("externalInvoiceNumber" IS NOT NULL AND btrim("externalInvoiceNumber") <> '')
				OR "isFree" OR COALESCE("auditVerdict", '') IN ('FREE', 'CREDIT')
				OR systems::text LIKE '%شغل داخل الشركة%'
			) NOT VALID;
		`,
	}, {
		// (ع) 10-14: مراحل الحجز بلا نت — وقت الضغطة الحقيقي + موقعها كدليل
		Version: "0360_booking_stage_proof",
		SQL: `
			CREATE TABLE IF NOT EXISTS "BookingStageProof" (
				id TEXT PRIMARY KEY, "bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
				stage TEXT NOT NULL, "employeeId" TEXT, "pressedAt" TIMESTAMP NOT NULL,
				"receivedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, offline BOOLEAN NOT NULL DEFAULT false,
				lat DOUBLE PRECISION, lng DOUBLE PRECISION, accuracy DOUBLE PRECISION, "distanceM" DOUBLE PRECISION,
				verdict TEXT NOT NULL DEFAULT 'OK');
			CREATE INDEX IF NOT EXISTS stage_proof_booking ON "BookingStageProof" ("bookingId");
			CREATE INDEX IF NOT EXISTS stage_proof_emp ON "BookingStageProof" ("employeeId", "pressedAt");
		`,
	}, {
		// (ع) 10-14: مراحل الحجز الست + موقع الشركة + تتبع المسار والخوارزمية التلقائية
		Version: "0361_booking_field_track",
		SQL: `
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "acceptedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "departedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "departOrigin" TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "autoDepartedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "autoArrivedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "autoCompletedAt" TIMESTAMP;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "autoArrivedLat" DOUBLE PRECISION;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "autoArrivedLng" DOUBLE PRECISION;
			CREATE TABLE IF NOT EXISTS "CompanyLocation" (
				id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
				lat DOUBLE PRECISION NOT NULL, lng DOUBLE PRECISION NOT NULL,
				"radiusM" INT NOT NULL DEFAULT 300, "updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			ALTER TABLE "LocationPing" ADD COLUMN IF NOT EXISTS accuracy DOUBLE PRECISION;
			CREATE INDEX IF NOT EXISTS "LocationPing_createdAt_idx" ON "LocationPing" ("createdAt");
			CREATE TABLE IF NOT EXISTS "BookingRouteStop" (
				id TEXT PRIMARY KEY, "bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
				"employeeId" TEXT, kind TEXT NOT NULL, lat DOUBLE PRECISION NOT NULL, lng DOUBLE PRECISION NOT NULL,
				"startAt" TIMESTAMP NOT NULL, "endAt" TIMESTAMP NOT NULL, minutes INT NOT NULL);
			CREATE INDEX IF NOT EXISTS route_stop_booking ON "BookingRouteStop" ("bookingId");
			ALTER TABLE "Customer" ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION;
			ALTER TABLE "Customer" ADD COLUMN IF NOT EXISTS lng DOUBLE PRECISION;
			ALTER TABLE "Customer" ADD COLUMN IF NOT EXISTS "locationSource" TEXT;
		`,
	}, {
		// (ع) 10-15: تكملة المراحل — مدة شغل موثوقة للمتوسطات + تذكير «وصلنا»
		Version: "0362_reliable_work_minutes",
		SQL: `
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "arriveRemindedAt" TIMESTAMP;
			-- مدة الشغل بالدقايق: أوقات النظام (من الـGPS) أولاً، وإلا ضغطات الموظف.
			-- NULL (يعني ما تدخل بالمتوسط) إذا البدء/الإنجاز تسجيل جماعي أو بعيد
			-- بلا وقت تلقائي يعوّضه، أو المدة أقل من ٥ دقايق (تسجيل بنفس اللحظة).
			CREATE OR REPLACE FUNCTION booking_work_minutes(bid TEXT, s TIMESTAMP, c TIMESTAMP, aa TIMESTAMP, ac TIMESTAMP)
			RETURNS DOUBLE PRECISION LANGUAGE plpgsql STABLE AS $$
			DECLARE st TIMESTAMP := COALESCE(aa, s); en TIMESTAMP := COALESCE(ac, c); m DOUBLE PRECISION;
			BEGIN
				IF st IS NULL OR en IS NULL OR en <= st THEN RETURN NULL; END IF;
				IF (aa IS NULL OR ac IS NULL) AND EXISTS (SELECT 1 FROM "BookingStageProof" p WHERE p."bookingId" = bid
					AND p.stage IN ('STARTED', 'COMPLETED') AND p.verdict IN ('BATCH', 'FAR')) THEN RETURN NULL; END IF;
				m := EXTRACT(EPOCH FROM (en - st)) / 60;
				IF m < 5 OR m > 16 * 60 THEN RETURN NULL; END IF;
				RETURN m;
			END $$;
		`,
	}, {
		// (ع) 10-15: رقابة المشتريات — تقييم المورد ببنود · تثبيت المحاسب · تنبيهات الأسعار
		Version: "0363_purchase_control",
		SQL: `
			CREATE TABLE IF NOT EXISTS "SupplierReview" (
				id TEXT PRIMARY KEY, "supplierId" TEXT NOT NULL REFERENCES "Supplier"(id) ON DELETE CASCADE,
				"procurementId" TEXT REFERENCES "ProcurementRequest"(id) ON DELETE SET NULL,
				price INT NOT NULL, quality INT NOT NULL, speed INT NOT NULL, punctuality INT NOT NULL, dealing INT NOT NULL,
				note TEXT, "ratedById" TEXT NOT NULL, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE UNIQUE INDEX IF NOT EXISTS supplier_review_once ON "SupplierReview" ("procurementId", "ratedById") WHERE "procurementId" IS NOT NULL;
			CREATE INDEX IF NOT EXISTS supplier_review_sup ON "SupplierReview" ("supplierId");
			CREATE TABLE IF NOT EXISTS "ProcurementVerify" (
				"procurementId" TEXT PRIMARY KEY REFERENCES "ProcurementRequest"(id) ON DELETE CASCADE,
				"verifiedById" TEXT NOT NULL, "verifiedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, note TEXT);
			CREATE TABLE IF NOT EXISTS "ProcurementPriceAlert" (
				id TEXT PRIMARY KEY, "procurementId" TEXT NOT NULL REFERENCES "ProcurementRequest"(id) ON DELETE CASCADE,
				"itemId" TEXT NOT NULL, kind TEXT NOT NULL, "productName" TEXT NOT NULL,
				"supplierId" TEXT, "unitPrice" DOUBLE PRECISION NOT NULL, quantity INT NOT NULL,
				"refSupplierId" TEXT, "refPrice" DOUBLE PRECISION NOT NULL, "lossIqd" DOUBLE PRECISION NOT NULL,
				"buyerId" TEXT, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, "reviewedAt" TIMESTAMP, "reviewNote" TEXT);
			CREATE UNIQUE INDEX IF NOT EXISTS price_alert_once ON "ProcurementPriceAlert" ("itemId", kind);
			-- اسم المادة الموحّد للمقارنة: حروف صغيرة ومسافة وحدة، ويّا الموديل إذا موجود
			CREATE OR REPLACE FUNCTION product_key(n TEXT, m TEXT) RETURNS TEXT LANGUAGE sql IMMUTABLE AS $$
				SELECT lower(btrim(regexp_replace(COALESCE(n, '') || CASE WHEN COALESCE(btrim(m), '') <> '' THEN ' ' || m ELSE '' END, '\s+', ' ', 'g')))
			$$;
		`,
	}, {
		// (ع) 10-15: رقابة المشتريات ٢ — شرط العرضين للمشتريات الكبيرة
		Version: "0364_purchase_quotes",
		SQL: `
			CREATE TABLE IF NOT EXISTS "PurchasePolicy" (
				id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1), "quoteThreshold" DOUBLE PRECISION NOT NULL DEFAULT 500000,
				"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			INSERT INTO "PurchasePolicy" (id) VALUES (1) ON CONFLICT DO NOTHING;
			CREATE TABLE IF NOT EXISTS "ProcurementQuote" (
				id TEXT PRIMARY KEY, "procurementId" TEXT NOT NULL REFERENCES "ProcurementRequest"(id) ON DELETE CASCADE,
				"supplierId" TEXT NOT NULL REFERENCES "Supplier"(id), "totalPrice" DOUBLE PRECISION NOT NULL,
				chosen BOOLEAN NOT NULL DEFAULT false, note TEXT, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE INDEX IF NOT EXISTS procurement_quote_req ON "ProcurementQuote" ("procurementId");
		`,
	}, {
		// (ع) 10-15: رقابة المشتريات ٣ — الأسعار الستاندرد + مشتريات الفنيين من موردينهم
		Version: "0365_standard_prices",
		SQL: `
			ALTER TABLE "PurchasePolicy" ADD COLUMN IF NOT EXISTS "ceilingPct" DOUBLE PRECISION NOT NULL DEFAULT 10;
			CREATE TABLE IF NOT EXISTS "StandardPrice" (
				key TEXT PRIMARY KEY, product TEXT NOT NULL, price DOUBLE PRECISION NOT NULL,
				source TEXT NOT NULL DEFAULT 'AUTO', "setById" TEXT, "updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE TABLE IF NOT EXISTS "StandardPriceSuggestion" (
				key TEXT PRIMARY KEY, product TEXT NOT NULL, current DOUBLE PRECISION NOT NULL, suggested DOUBLE PRECISION NOT NULL,
				basis TEXT NOT NULL, "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			-- شراء مواد من مصروف الفني: من أي مورد فنيين، شنو المادة، الكمية والسعر
			CREATE TABLE IF NOT EXISTS "ExpensePurchase" (
				"expenseId" TEXT PRIMARY KEY REFERENCES "Expense"(id) ON DELETE CASCADE,
				"techSupplierId" TEXT REFERENCES "TechSupplier"(id) ON DELETE SET NULL,
				"productName" TEXT NOT NULL, quantity INT NOT NULL, "unitPrice" DOUBLE PRECISION NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			ALTER TABLE "ProcurementPriceAlert" ALTER COLUMN "procurementId" DROP NOT NULL;
			ALTER TABLE "ProcurementPriceAlert" ADD COLUMN IF NOT EXISTS "expenseId" TEXT REFERENCES "Expense"(id) ON DELETE CASCADE;
			ALTER TABLE "ProcurementPriceAlert" ADD COLUMN IF NOT EXISTS "techSupplierId" TEXT;
			-- كل المشتريات بمكان واحد: طلبات المواد (موردين المشتريات) + مصاريف الفنيين (موردين الفنيين)
			CREATE OR REPLACE VIEW "PurchaseLine" AS
				SELECT 'PROC' AS source, r.id AS "refId", r.code, product_key(i."productName", i.model) AS key, i."productName" AS product,
					i.quantity AS qty, i."unitPrice" AS price, r."fulfilledAt" AS at, r."supplierId" AS "supplierId",
					s."companyName" AS supplier, r."fulfilledById" AS "buyerId"
				FROM "ProcurementItem" i JOIN "ProcurementRequest" r ON r.id = i."requestId" LEFT JOIN "Supplier" s ON s.id = r."supplierId"
				WHERE i."unitPrice" > 0 AND r."fulfilledAt" IS NOT NULL
				UNION ALL
				SELECT 'TECH', e.id, 'مصروف', product_key(p."productName", NULL), p."productName", p.quantity, p."unitPrice", p."createdAt",
					p."techSupplierId", COALESCE(t."companyName", 'مورد غير مسجّل'), e."employeeId"
				FROM "ExpensePurchase" p JOIN "Expense" e ON e.id = p."expenseId" LEFT JOIN "TechSupplier" t ON t.id = p."techSupplierId";
		`,
	}, {
		// (ع) 10-15: توقعات ماتركس — حفظ توقع الحجوزات + الحكم صح/غلط آخر الشهر
		Version: "0366_forecast_outcome",
		SQL: `
			ALTER TABLE "MatrixForecastLog" ADD COLUMN IF NOT EXISTS "expectedJobs" INT;
			ALTER TABLE "MatrixForecastLog" ADD COLUMN IF NOT EXISTS "mtdJobs" INT;
			CREATE TABLE IF NOT EXISTS "MatrixForecastOutcome" (
				month TEXT NOT NULL, kind TEXT NOT NULL, forecast DOUBLE PRECISION NOT NULL, actual DOUBLE PRECISION NOT NULL,
				"errorPct" DOUBLE PRECISION NOT NULL, verdict TEXT NOT NULL, "snapshotDay" DATE NOT NULL,
				"evaluatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (month, kind));
		`,
	}, {
		// (ع) 10-15: علاقات الموردين — سجل التواصل ويّا موردين المشتريات والتقنيين + تذكير
		Version: "0367_supplier_relations",
		SQL: `
			ALTER TABLE "PurchasePolicy" ADD COLUMN IF NOT EXISTS "contactDays" INT NOT NULL DEFAULT 30;
			CREATE TABLE IF NOT EXISTS "SupplierContact" (
				id TEXT PRIMARY KEY, "supplierKind" TEXT NOT NULL, "supplierId" TEXT NOT NULL,
				kind TEXT NOT NULL, note TEXT, "byId" TEXT, at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
			CREATE INDEX IF NOT EXISTS supplier_contact_sup ON "SupplierContact" ("supplierKind", "supplierId", at DESC);
			INSERT INTO "Permission" (id, name, label) VALUES
				(gen_random_uuid()::text, 'supplier_relations', 'علاقات الموردين (التواصل والتذكير)')
			ON CONFLICT (name) DO NOTHING;
		`,
	}, {
		// (ع) 10-15: الشكاوى المفتوحة الي ما مكلّف بيها أحد تروح لمهندس الجودة
		Version: "0368_complaints_to_quality",
		SQL: `
			UPDATE "Complaint" c SET "assignedToEmployeeId" = (
				SELECT e.id FROM "Employee" e WHERE e.role::text = 'QUALITY_ENGINEER' AND e.status::text = 'ACTIVE' ORDER BY e."createdAt" LIMIT 1)
			WHERE c."assignedToEmployeeId" IS NULL AND c.status::text IN ('NEW', 'IN_PROGRESS');
		`,
	}, {
		// 🔴 (ع) 10-16: «دققوهن وشافو مشكلتهن بس الى الان يضهرن» — قيد الحالة
		// من 0293 يقبل PENDING/DISMISSED بس، فزر «اطلب حذفه»/الدمج (RESOLVED)
		// چان يفشل بصمت والزوج يبقى معلّق ينعدّ متأخر على المراقب.
		Version: "0369_duplicate_resolved_status",
		SQL: `
			ALTER TABLE "DuplicateCandidate" DROP CONSTRAINT IF EXISTS "DuplicateCandidate_status_check";
			ALTER TABLE "DuplicateCandidate" ADD CONSTRAINT "DuplicateCandidate_status_check"
				CHECK (status IN ('PENDING', 'DISMISSED', 'RESOLVED'));
		`,
	}, {
		// 🔴 (ع) 10-16: الطلعات القديمة الي فاتها ليدر المهمة (استلمها بلا تكليف)
		// — ينضاف لكل طلعة حتى يطلع اسمه وتنحسبله.
		Version: "0370_visit_crew_leader",
		SQL: `
			INSERT INTO "BookingVisitCrew" (id, "visitId", "employeeId", role, "isLeader")
			SELECT gen_random_uuid()::text, v.id, l.id, 'LEADER', true
			FROM "BookingVisit" v
			JOIN LATERAL (
			  SELECT DISTINCT m."leaderId" AS id FROM "Mission" m WHERE m."bookingId" = v."bookingId" AND m."leaderId" IS NOT NULL
			) l ON true
			JOIN "Employee" e ON e.id = l.id
			ON CONFLICT ("visitId", "employeeId") DO NOTHING;
		`,
	}, {
		// 🔴 (ع) 10-16 «اكو هواي هيج اخطاء… دور بكل انحاء النظام»: «ليدر الحجز»
		// چان إله أكثر من ٨ تعريفات (كل شاشة بترتيب)، فنفس الحجز يطلع بليدر
		// بشاشة و«بلا ليدر» بثانية، والليدر الي مو بالتكليف ما تنحسبله طلعاته.
		// هسه تعريف واحد بقاعدة البيانات، وكل النظام يناديه.
		//
		// booking_leader_id: مشرف الحجز (إذا هو ليدر) ← ليدر بالتكليف (الأقدم)
		//                    ← ليدر المهمة الميدانية (الأحدث).
		// booking_crew_ids : كل من اشتغل على الحجز — التكليف + المشرف الليدر +
		//                    ليدر المهمة وأعضاؤها.
		Version: "0371_booking_leader_fn",
		SQL: `
			CREATE OR REPLACE FUNCTION booking_leader_id(bid TEXT) RETURNS TEXT
			LANGUAGE sql STABLE AS $f$
				SELECT COALESCE(
				  (SELECT b."projectSupervisorId" FROM "Booking" b JOIN "Employee" e ON e.id = b."projectSupervisorId"
				    WHERE b.id = bid AND e."isLeader"),
				  (SELECT ba."employeeId" FROM "BookingAssignment" ba JOIN "Employee" e ON e.id = ba."employeeId"
				    WHERE ba."bookingId" = bid AND e."isLeader" ORDER BY ba."createdAt", ba.role LIMIT 1),
				  (SELECT m."leaderId" FROM "Mission" m WHERE m."bookingId" = bid AND m."leaderId" IS NOT NULL
				    ORDER BY m."createdAt" DESC LIMIT 1))
			$f$;
			CREATE OR REPLACE FUNCTION booking_crew_ids(bid TEXT) RETURNS SETOF TEXT
			LANGUAGE sql STABLE AS $f$
				SELECT ba."employeeId" FROM "BookingAssignment" ba WHERE ba."bookingId" = bid
				UNION
				SELECT b."projectSupervisorId" FROM "Booking" b JOIN "Employee" e ON e.id = b."projectSupervisorId"
				 WHERE b.id = bid AND e."isLeader"
				UNION
				SELECT m."leaderId" FROM "Mission" m WHERE m."bookingId" = bid AND m."leaderId" IS NOT NULL
				UNION
				SELECT unnest(m."memberIds") FROM "Mission" m WHERE m."bookingId" = bid
			$f$;
			-- الطلعات القديمة: ليدر الحجز (بالتعريف الموحّد) ينضاف لكادرها لو فاته
			INSERT INTO "BookingVisitCrew" (id, "visitId", "employeeId", role, "isLeader")
			SELECT gen_random_uuid()::text, v.id, l.id, 'LEADER', true
			FROM "BookingVisit" v
			JOIN LATERAL (SELECT booking_leader_id(v."bookingId") AS id) l ON l.id IS NOT NULL
			ON CONFLICT ("visitId", "employeeId") DO NOTHING;
			-- الدوام المحسوب يتمدد لآخر حجز خلّصه الموظف — والليدر الي مو بالتكليف هم
			CREATE OR REPLACE FUNCTION att_last_activity(emp TEXT, ci TIMESTAMP, co TIMESTAMP)
			RETURNS TIMESTAMP LANGUAGE sql STABLE AS $$
				SELECT max(b."completedAt") FROM "Booking" b
				WHERE b."completedAt" > ci
				  AND b."completedAt" <= LEAST(COALESCE(co, NOW() AT TIME ZONE 'UTC'),
				      date_trunc('day', ci + interval '3 hours') + interval '27 hours')
				  AND (EXISTS (SELECT 1 FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id AND ba."employeeId" = emp)
				    OR (b."projectSupervisorId" = emp AND EXISTS (SELECT 1 FROM "Employee" se WHERE se.id = emp AND se."isLeader"))
				    OR EXISTS (SELECT 1 FROM "Mission" m WHERE m."bookingId" = b.id AND (m."leaderId" = emp OR emp = ANY(m."memberIds"))))
			$$;
			UPDATE "BookingVisitCrew" vc SET "isLeader" = true
			FROM "BookingVisit" v
			WHERE v.id = vc."visitId" AND vc."employeeId" = booking_leader_id(v."bookingId") AND NOT vc."isLeader";
		`,
	}, {
		// (ع) 10-16: «تقرير تنفيذ المشروع» بنموذج المشرف الأصلي — للحجز وللمشروع اليومي
		Version: "0372_execution_sheet",
		SQL: `
			ALTER TABLE "WorkReport" ADD COLUMN IF NOT EXISTS sheet JSONB;
			ALTER TABLE "BookingProgressReport" ADD COLUMN IF NOT EXISTS sheet JSONB;
		`,
	}, {
		// (ع) 10-16: «شوكت استلم المشروع، شوكت تواصل ويا الزبون، شوكت سواله عرض سعر»
		Version: "0373_project_customer_contact",
		SQL: `
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "customerContactedAt" TIMESTAMP;
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "customerContactedById" TEXT;
			ALTER TABLE "Project" ADD COLUMN IF NOT EXISTS "customerContactNote" TEXT;
		`,
	}, {
		// (ع) 10-16: «كل شكوى تاريخها قبل الشهر الثامن من هاي السنة تعتبر محلولة» —
		// من النظام القديم. التسوية مكشوفة بعمودها، مو مخلوطة بالحل الحقيقي.
		Version: "0374_complaint_legacy_settle",
		SQL: `
			ALTER TABLE "Complaint" ADD COLUMN IF NOT EXISTS "settledLegacyAt" TIMESTAMP;
			UPDATE "Complaint"
			   SET status = 'RESOLVED', "resolvedAt" = COALESCE("resolvedAt", now()),
			       "settledLegacyAt" = now(),
			       resolution = COALESCE(NULLIF(resolution, ''), 'محلولة — من النظام القديم')
			 WHERE "createdAt" < '2026-08-01' AND status NOT IN ('RESOLVED', 'CLOSED');
		`,
	}, {
		// (ع) 10-16: «من يتسجل الحجز بعد الشكوى لازم يتسجل حجز شكوى مو حجز عادي»
		Version: "0375_complaint_booking",
		SQL: `
			ALTER TYPE "BookingType" ADD VALUE IF NOT EXISTS 'COMPLAINT';
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "complaintId" TEXT REFERENCES "Complaint"(id) ON DELETE SET NULL;
			CREATE INDEX IF NOT EXISTS "Booking_complaint_idx" ON "Booking" ("complaintId") WHERE "complaintId" IS NOT NULL;
		`,
	}, {
		// (ع) 10-16: «الشكاوى القديمة يطلعلي كدامهن تم التواصل ومغلقات، والجديدة
		// مطلوب متابعة». كل شكوى قبل ٢٠٢٦-٠٨-٠١: مغلقة، تواصل مؤشّر، والتدقيق
		// معتمد بملاحظة «من النظام القديم» — حتى ما تبقى «✕ لا» و«مطلوب متابعة».
		Version: "0376_complaint_legacy_closed",
		SQL: `
			UPDATE "Complaint"
			   SET status = 'CLOSED',
			       "contactedAt" = COALESCE("contactedAt", "resolvedAt", "createdAt"),
			       "resolvedAt" = COALESCE("resolvedAt", now()),
			       "settledLegacyAt" = COALESCE("settledLegacyAt", now()),
			       "auditVerdict" = COALESCE("auditVerdict", 'APPROVED'),
			       "auditNote" = COALESCE(NULLIF("auditNote", ''), 'مغلقة — من النظام القديم'),
			       "auditedAt" = COALESCE("auditedAt", now())
			 WHERE "createdAt" < '2026-08-01';
		`,
	}, {
		// (ع) 10-16: المراقب يقيّم شغل مهندس الجودة («تمام / مو تمام»)، ويسجّل
		// تقصير على الحالات الي ما انتصل بيها — بلا غرامة فلوس.
		Version: "0377_quality_monitor_verdict",
		SQL: `
			ALTER TABLE "QualityFollowUp" ADD COLUMN IF NOT EXISTS "monitorVerdict" TEXT;
			ALTER TABLE "QualityFollowUp" ADD COLUMN IF NOT EXISTS "monitorNote" TEXT;
			ALTER TABLE "QualityFollowUp" ADD COLUMN IF NOT EXISTS "monitorById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
			ALTER TABLE "QualityFollowUp" ADD COLUMN IF NOT EXISTS "monitorAt" TIMESTAMP;
			ALTER TABLE "QualityFollowUp" ADD COLUMN IF NOT EXISTS "monitorTargetId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
		`,
	}, {
		// (ع) 10-16: «صار عدنا ٣ مدخلات لأوقات العمل — ماتركس يشوفهن كلهن، وإذا غابن
		// ثنين يعتمد الموجودة». وقت الشغل الموثوق لكل حجز، ومن وين أخذه:
		//   البداية: GPS (وصل الموقع) ← ضغطة «بدأ العمل» (إذا مو جماعية/بعيدة)
		//            ← وقت البدء بورقة التنفيذ ← الانطلاق + نص ساعة.
		//   النهاية: GPS (خلص) ← ضغطة «تم» (إذا مو جماعية/بعيدة) ← وقت الانتهاء بالورقة.
		// الأعمدة بلا منطقة زمنية = UTC؛ ساعة الورقة بتوقيت بغداد.
		Version: "0378_booking_work_window",
		SQL: `
			CREATE OR REPLACE FUNCTION booking_sheet_time(d TEXT, h TEXT, m TEXT) RETURNS TIMESTAMP
			LANGUAGE plpgsql IMMUTABLE AS $$
			BEGIN
				IF d !~ '^\d{4}-\d{2}-\d{2}$' OR h !~ '^\d{1,2}$' OR m !~ '^\d{1,2}$' THEN RETURN NULL; END IF;
				RETURN ((d || ' ' || h || ':' || m)::timestamp AT TIME ZONE 'Asia/Baghdad') AT TIME ZONE 'UTC';
			EXCEPTION WHEN others THEN RETURN NULL;
			END $$;

			CREATE OR REPLACE FUNCTION booking_work_window(bid TEXT)
			RETURNS TABLE("startAt" TIMESTAMP, "endAt" TIMESTAMP, "startSrc" TEXT, "endSrc" TEXT)
			LANGUAGE plpgsql STABLE AS $$
			DECLARE b RECORD; sh JSONB; bad_s BOOLEAN; bad_e BOOLEAN; ss TIMESTAMP; se TIMESTAMP; dep TIMESTAMP;
			BEGIN
				SELECT "autoArrivedAt", "autoCompletedAt", "startedAt", "completedAt", "autoDepartedAt", "departedAt"
				  INTO b FROM "Booking" WHERE id = bid;
				IF NOT FOUND THEN RETURN; END IF;
				bad_s := EXISTS (SELECT 1 FROM "BookingStageProof" p WHERE p."bookingId" = bid AND p.stage = 'STARTED' AND p.verdict IN ('BATCH', 'FAR'));
				bad_e := EXISTS (SELECT 1 FROM "BookingStageProof" p WHERE p."bookingId" = bid AND p.stage = 'COMPLETED' AND p.verdict IN ('BATCH', 'FAR'));
				SELECT w.sheet INTO sh FROM "WorkReport" w WHERE w."bookingId" = bid AND w.sheet IS NOT NULL ORDER BY w."createdAt" DESC LIMIT 1;
				IF sh IS NOT NULL THEN
					ss := booking_sheet_time(COALESCE(NULLIF(sh->'header'->>'date', ''), sh->'data'->>'startDate'), sh->'time'->>'startH', sh->'time'->>'startM');
					se := booking_sheet_time(COALESCE(NULLIF(sh->'data'->>'endDate', ''), sh->'header'->>'date'), sh->'time'->>'endH', sh->'time'->>'endM');
				END IF;
				dep := COALESCE(b."autoDepartedAt", b."departedAt");
				"startAt" := CASE WHEN b."autoArrivedAt" IS NOT NULL THEN b."autoArrivedAt"
				                  WHEN b."startedAt" IS NOT NULL AND NOT bad_s THEN b."startedAt"
				                  WHEN ss IS NOT NULL THEN ss
				                  WHEN dep IS NOT NULL THEN dep + interval '30 minutes' END;
				"startSrc" := CASE WHEN b."autoArrivedAt" IS NOT NULL THEN 'GPS'
				                   WHEN b."startedAt" IS NOT NULL AND NOT bad_s THEN 'TAP'
				                   WHEN ss IS NOT NULL THEN 'SHEET'
				                   WHEN dep IS NOT NULL THEN 'DEPART_30' END;
				"endAt" := CASE WHEN b."autoCompletedAt" IS NOT NULL THEN b."autoCompletedAt"
				                WHEN b."completedAt" IS NOT NULL AND NOT bad_e THEN b."completedAt"
				                WHEN se IS NOT NULL THEN se END;
				"endSrc" := CASE WHEN b."autoCompletedAt" IS NOT NULL THEN 'GPS'
				                 WHEN b."completedAt" IS NOT NULL AND NOT bad_e THEN 'TAP'
				                 WHEN se IS NOT NULL THEN 'SHEET' END;
				RETURN NEXT;
			END $$;

			-- نفس التوقيع القديم حتى كل استعلامات ماتركس تستفاد بلا تغيير.
			CREATE OR REPLACE FUNCTION booking_work_minutes(bid TEXT, s TIMESTAMP, c TIMESTAMP, aa TIMESTAMP, ac TIMESTAMP)
			RETURNS DOUBLE PRECISION LANGUAGE plpgsql STABLE AS $$
			DECLARE w RECORD; m DOUBLE PRECISION;
			BEGIN
				SELECT * INTO w FROM booking_work_window(bid);
				IF w."startAt" IS NULL OR w."endAt" IS NULL OR w."endAt" <= w."startAt" THEN RETURN NULL; END IF;
				m := EXTRACT(EPOCH FROM (w."endAt" - w."startAt")) / 60;
				IF m < 5 OR m > 16 * 60 THEN RETURN NULL; END IF;
				RETURN m;
			END $$;
		`,
	}, {
		// (ع) 10-16: «حجز ديكور مربوط» — الديكور يطلعون ويا الشعبة الهندسية أو قبلها
		// بيوم يكملون شغل (حداد يسوي قاعدة داتا شو…). حجز مستقل مربوط بالأب.
		Version: "0379_decor_booking",
		SQL: `
			ALTER TYPE "BookingType" ADD VALUE IF NOT EXISTS 'DECOR';
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "parentBookingId" TEXT REFERENCES "Booking"(id) ON DELETE SET NULL;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "decorTiming" TEXT;
			CREATE INDEX IF NOT EXISTS "Booking_parent_idx" ON "Booking" ("parentBookingId") WHERE "parentBookingId" IS NOT NULL;
		`,
	}, {
		// (ع) 10-16: «جدول أسعار للديكور» — لكل مهنة بنود بسعرها (متر/قطعة/يوم)،
		// والفاتورة تنحسب منه ومن القياسات. ماكو سعر بالإيد، نفس حماية الشعبة.
		Version: "0380_decor_rates",
		SQL: `
			CREATE TABLE IF NOT EXISTS "DecorRate" (
				id TEXT PRIMARY KEY,
				"serviceId" TEXT NOT NULL REFERENCES "Service"(id) ON DELETE CASCADE,
				name TEXT NOT NULL,
				unit TEXT NOT NULL,
				price NUMERIC NOT NULL CHECK (price >= 0),
				active BOOLEAN NOT NULL DEFAULT true,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now(),
				"updatedAt" TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "DecorRate_service_idx" ON "DecorRate" ("serviceId");
		`,
	}, {
		// (ع) 10-16: «لازم أكو خيار ينطي تم التدقيق حتى ما يضلن معلّكات» —
		// تقارير الحجوزات والتقارير اليومية للمشاريع.
		Version: "0381_report_review",
		SQL: `
			ALTER TABLE "WorkReport" ADD COLUMN IF NOT EXISTS "reviewedAt" TIMESTAMP;
			ALTER TABLE "WorkReport" ADD COLUMN IF NOT EXISTS "reviewedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
			ALTER TABLE "WorkReport" ADD COLUMN IF NOT EXISTS "reviewNote" TEXT;
			ALTER TABLE "BookingProgressReport" ADD COLUMN IF NOT EXISTS "reviewedAt" TIMESTAMP;
			ALTER TABLE "BookingProgressReport" ADD COLUMN IF NOT EXISTS "reviewedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL;
			ALTER TABLE "BookingProgressReport" ADD COLUMN IF NOT EXISTS "reviewNote" TEXT;
		`,
	}, {
		// (ع) 10-17: «من يطلع موظف، نختار موظف ويترحّل كلشي إله» — سجل التسليم.
		Version: "0382_employee_handover",
		SQL: `
			CREATE TABLE IF NOT EXISTS "EmployeeHandover" (
				id TEXT PRIMARY KEY,
				"fromId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"toId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"byId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				summary JSONB NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now()
			);
		`,
	}, {
		// (ع) 10-17: إدارة المركبات تشمل «إجازات سوق الموظفين، والتعهدات وتخويل
		// السياقة، وبطاقات البنزين» — الناقصات من الموجود (الوقود والعطلات والتنظيف).
		Version: "0383_vehicle_drivers_fuel_cards",
		SQL: `
			CREATE TABLE IF NOT EXISTS "VehicleDriver" (
				"employeeId" TEXT PRIMARY KEY REFERENCES "Employee"(id) ON DELETE CASCADE,
				"licenseNumber" TEXT,
				"licenseExpiry" DATE,
				authorized BOOLEAN NOT NULL DEFAULT false,
				"authorizedVehicles" TEXT,
				"pledgeSignedAt" DATE,
				"pledgeNote" TEXT,
				"updatedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"updatedAt" TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE TABLE IF NOT EXISTS "FuelCard" (
				id TEXT PRIMARY KEY,
				number TEXT NOT NULL,
				provider TEXT,
				"vehicleId" TEXT REFERENCES "Vehicle"(id) ON DELETE SET NULL,
				"holderEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"monthlyLimit" NUMERIC,
				active BOOLEAN NOT NULL DEFAULT true,
				notes TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now(),
				"updatedAt" TIMESTAMP NOT NULL DEFAULT now()
			);
		`,
	}, {
		// ═══ «السكرتير» — أول عميل ذكاء داخل النظام (قرار (ع) 10-22، docs/secretary.md) ═══
		// للمالك ومدير النظام، كل واحد بذاكرته. المحادثة، وكل مقترح تنفيذ
		// ونتيجته (السجل)، والتفضيلات المعتمدة بس.
		Version: "0386_secretary",
		SQL: `
			CREATE TABLE IF NOT EXISTS "SecretaryMessage" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
				text TEXT NOT NULL DEFAULT '',
				transcript TEXT,
				cards JSONB NOT NULL DEFAULT '[]'::jsonb,
				"actionId" TEXT,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "SecretaryMessage_emp_idx" ON "SecretaryMessage" ("employeeId", "createdAt" DESC);
			CREATE TABLE IF NOT EXISTS "SecretaryAction" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				kind TEXT NOT NULL,
				request TEXT NOT NULL DEFAULT '',
				summary TEXT NOT NULL DEFAULT '',
				fields JSONB NOT NULL DEFAULT '{}'::jsonb,
				status TEXT NOT NULL DEFAULT 'PROPOSED' CHECK (status IN ('PROPOSED', 'SENT', 'CANCELLED', 'FAILED')),
				result TEXT,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"decidedAt" TIMESTAMPTZ
			);
			CREATE INDEX IF NOT EXISTS "SecretaryAction_emp_idx" ON "SecretaryAction" ("employeeId", "createdAt" DESC);
			CREATE TABLE IF NOT EXISTS "SecretaryPref" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				text TEXT NOT NULL,
				active BOOLEAN NOT NULL DEFAULT true,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
		`,
	}, {
		// «تذكيراتي» (دازاي 10-22): الموظف يطلب مقابلة المدير أو يجيب ضيف،
		// والطلب يوصل لتذكيرات السكرتير — المالك أو المدير يحدد موعد أو يرفض بسبب.
		Version: "0387_meeting_request",
		SQL: `
			CREATE TABLE IF NOT EXISTS "MeetingRequest" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				kind TEXT NOT NULL CHECK (kind IN ('MEETING', 'GUEST')),
				reason TEXT NOT NULL,
				"guestName" TEXT,
				"preferredAt" TIMESTAMPTZ,
				status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SCHEDULED', 'DECLINED')),
				"scheduledAt" TIMESTAMPTZ,
				note TEXT,
				"decidedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"decidedAt" TIMESTAMPTZ
			);
			CREATE INDEX IF NOT EXISTS "MeetingRequest_emp_idx" ON "MeetingRequest" ("employeeId", "createdAt" DESC);
			CREATE INDEX IF NOT EXISTS "MeetingRequest_pending_idx" ON "MeetingRequest" ("createdAt") WHERE status = 'PENDING';
		`,
	}, {
		// «لقطة الواجبات اليومية» (دازاي 10-22): الواجبات چانت تنحسب لحظياً بس،
		// فماكو تاريخ يحكم عليه ماتركس («الي يكرر ما يسوي واجباته»). كل ليلة
		// ١١:٥٠ بغداد ينحفظ لكل موظف شنو كمّل — وحدة لكل موظف باليوم.
		Version: "0388_duty_snapshot",
		SQL: `
			CREATE TABLE IF NOT EXISTS "DutySnapshot" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				day DATE NOT NULL,
				position TEXT NOT NULL,
				"doneCount" INT NOT NULL,
				total INT NOT NULL,
				verdict TEXT NOT NULL,
				off TEXT,
				duties JSONB NOT NULL DEFAULT '[]'::jsonb,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				UNIQUE ("employeeId", day)
			);
			CREATE INDEX IF NOT EXISTS "DutySnapshot_day_idx" ON "DutySnapshot" (day);
		`,
	}, {
		// «الحجوزات المرجوعة من إدارة المشاريع وين تروح؟» (دازاي 10-23): چانت
		// ترجع لمحطتها بلا أثر، والإشعار لدور «إداري الكوادر» بس — فتضيع. هسه
		// تنعلّم، وتطلع بمحطة «↩️ رجعت من المشاريع» لحد ما المنسّق يستلمها.
		Version: "0389_booking_returned_from_projects",
		SQL: `
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "returnedFromProjectsAt" TIMESTAMPTZ;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "returnedFromProjectsNote" TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "returnedFromProjectsById" TEXT;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "returnHandledAt" TIMESTAMPTZ;
			ALTER TABLE "Booking" ADD COLUMN IF NOT EXISTS "returnHandledById" TEXT;
			CREATE INDEX IF NOT EXISTS "Booking_returned_open_idx" ON "Booking" ("returnedFromProjectsAt")
				WHERE "returnedFromProjectsAt" IS NOT NULL AND "returnHandledAt" IS NULL;
		`,
	}, {
		// 🕶️ «الدخول المتخفي» (دازاي 10-23): المالك يدخل بحساب موظف يشتغل بيه.
		// كل بداية جلسة وكل كتابة تنسجل هنا — الي سوّاه المالك بحساب الموظف
		// يبقى مثبوت عليه، مو على الموظف.
		Version: "0390_stealth_log",
		SQL: `
			CREATE TABLE IF NOT EXISTS "StealthLog" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"ownerId" TEXT NOT NULL,
				"employeeId" TEXT NOT NULL,
				method TEXT NOT NULL,
				path TEXT NOT NULL,
				status INT NOT NULL DEFAULT 0,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "StealthLog_employee_idx" ON "StealthLog" ("employeeId", "createdAt" DESC);
		`,
	}, {
		// 📔 «مذكراتي» و«جدول أعمالي» (دازاي 10-23): ملاحظات المالك/المدير الخاصة —
		// كل واحد يشوف ملاحظاته بس. DIARY مذكرة حرة، AGENDA شي بجدوله بوقت.
		Version: "0391_secretary_note",
		SQL: `
			CREATE TABLE IF NOT EXISTS "SecretaryNote" (
				id TEXT PRIMARY KEY,
				"ownerId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				kind TEXT NOT NULL CHECK (kind IN ('DIARY', 'AGENDA')),
				body TEXT NOT NULL,
				"at" TIMESTAMPTZ,
				"doneAt" TIMESTAMPTZ,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "SecretaryNote_owner_idx" ON "SecretaryNote" ("ownerId", kind, "createdAt" DESC);
		`,
	}, {
		// 🤖 أسماء الوكلاء (دازاي 10-23): «ابو الجودة» و«ابو الحجوزات» اسمهم يحدده
		// المالك من الإعدادات بعدين. ماتركس وايريس ثابتين بالكود.
		Version: "0392_ai_agent_name",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AiAgentName" (
				agent TEXT PRIMARY KEY CHECK (agent IN ('QUALITY', 'BOOKING')),
				name TEXT NOT NULL,
				"updatedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
		`,
	}, {
		// 💬 وكيلي الجودة والحجوزات عالواتساب (دازاي 10-23 — docs/ai-quality-booking-plan.md).
		// المحادثة والرسائل، وكل إجراء سواه الوكيل ويا سببه حتى ينراجع وينلغى.
		// ومتابعة الجودة تاخذ تقييم الزبون للشغل وللكادر.
		Version: "0393_wa_agents",
		SQL: `
			CREATE TABLE IF NOT EXISTS "WaConversation" (
				id TEXT PRIMARY KEY,
				agent TEXT NOT NULL CHECK (agent IN ('QUALITY', 'BOOKING')),
				"customerId" TEXT REFERENCES "Customer"(id) ON DELETE SET NULL,
				phone TEXT NOT NULL,
				"bookingId" TEXT REFERENCES "Booking"(id) ON DELETE SET NULL,
				"followUpId" TEXT REFERENCES "QualityFollowUp"(id) ON DELETE SET NULL,
				status TEXT NOT NULL DEFAULT 'BOT' CHECK (status IN ('BOT', 'HUMAN', 'CLOSED')),
				sandbox BOOLEAN NOT NULL DEFAULT false,
				"humanById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"lastMessageAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"lastReadAt" TIMESTAMPTZ,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE UNIQUE INDEX IF NOT EXISTS "WaConversation_open_uq" ON "WaConversation" (agent, phone)
				WHERE status <> 'CLOSED' AND NOT sandbox;
			CREATE INDEX IF NOT EXISTS "WaConversation_list_idx" ON "WaConversation" (agent, status, "lastMessageAt" DESC);
			CREATE TABLE IF NOT EXISTS "WaMessage" (
				id TEXT PRIMARY KEY,
				"conversationId" TEXT NOT NULL REFERENCES "WaConversation"(id) ON DELETE CASCADE,
				direction TEXT NOT NULL CHECK (direction IN ('IN', 'OUT')),
				author TEXT NOT NULL CHECK (author IN ('CUSTOMER', 'AGENT', 'HUMAN')),
				"authorId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				text TEXT NOT NULL,
				"metaId" TEXT UNIQUE,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "WaMessage_conv_idx" ON "WaMessage" ("conversationId", "createdAt");
			CREATE TABLE IF NOT EXISTS "AiAgentAction" (
				id TEXT PRIMARY KEY,
				agent TEXT NOT NULL,
				"conversationId" TEXT REFERENCES "WaConversation"(id) ON DELETE SET NULL,
				kind TEXT NOT NULL,
				summary TEXT NOT NULL,
				reason TEXT NOT NULL DEFAULT '',
				"targetType" TEXT,
				"targetId" TEXT,
				payload JSONB,
				sandbox BOOLEAN NOT NULL DEFAULT false,
				"undoneAt" TIMESTAMPTZ,
				"undoneById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"undoReason" TEXT,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "AiAgentAction_target_idx" ON "AiAgentAction" ("targetId");
			CREATE INDEX IF NOT EXISTS "AiAgentAction_conv_idx" ON "AiAgentAction" ("conversationId");
			ALTER TABLE "QualityFollowUp" ADD COLUMN IF NOT EXISTS "customerRating" INT CHECK ("customerRating" BETWEEN 1 AND 5);
			ALTER TABLE "QualityFollowUp" ADD COLUMN IF NOT EXISTS "crewRating" INT CHECK ("crewRating" BETWEEN 1 AND 5);
			ALTER TABLE "QualityFollowUp" ADD COLUMN IF NOT EXISTS "contactedByAgent" TEXT;
		`,
	}, {
		// 💬 محادثات ايريس المتعددة + ذاكرة لكل محادثة (دازاي 10-23: «اريد محادثات مو محادثة
		// وحدة… كل محادثة إلها ذاكرة حتى ما يصرف أكثر»). الرسائل القديمة تنضم لمحادثة
		// «المحادثات السابقة». وماتركس عنده محادثات أصلاً (MatrixTalk) — ياخذ الذاكرة بس.
		Version: "0394_ai_conversations_memory",
		SQL: `
			CREATE TABLE IF NOT EXISTS "SecretaryConversation" (
				id TEXT PRIMARY KEY,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				title TEXT NOT NULL DEFAULT 'محادثة جديدة',
				summary TEXT NOT NULL DEFAULT '',
				"summarizedUpTo" TIMESTAMPTZ,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "SecretaryConversation_emp_idx" ON "SecretaryConversation" ("employeeId", "updatedAt" DESC);
			ALTER TABLE "SecretaryMessage" ADD COLUMN IF NOT EXISTS "conversationId" TEXT REFERENCES "SecretaryConversation"(id) ON DELETE CASCADE;
			INSERT INTO "SecretaryConversation" (id, "employeeId", title, "createdAt", "updatedAt")
				SELECT 'old-' || m."employeeId", m."employeeId", 'المحادثات السابقة', min(m."createdAt"), max(m."createdAt")
				FROM "SecretaryMessage" m WHERE m."conversationId" IS NULL GROUP BY m."employeeId"
				ON CONFLICT (id) DO NOTHING;
			UPDATE "SecretaryMessage" SET "conversationId" = 'old-' || "employeeId" WHERE "conversationId" IS NULL;
			CREATE INDEX IF NOT EXISTS "SecretaryMessage_conv_idx" ON "SecretaryMessage" ("conversationId", "createdAt");
			ALTER TABLE "MatrixTalk" ADD COLUMN IF NOT EXISTS summary TEXT NOT NULL DEFAULT '';
			ALTER TABLE "MatrixTalk" ADD COLUMN IF NOT EXISTS "summarizedUpTo" TIMESTAMPTZ;
		`,
	}, {
		// (دازاي 10-25) كادر الديكور: صورة «قبل» و«بعد» لكل شغلة بدل التقرير المكتوب
		// (أغلبهم ما يكتبون). الصورة تنحفظ بالتخزين نفسه، وهنا ربطها بالحجز والمرحلة.
		Version: "0395_work_photo",
		SQL: `
			CREATE TABLE IF NOT EXISTS "WorkPhoto" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
				"employeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				phase TEXT NOT NULL CHECK (phase IN ('BEFORE', 'AFTER')),
				"fileKey" TEXT NOT NULL,
				"contentType" TEXT NOT NULL,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "WorkPhoto_booking" ON "WorkPhoto"("bookingId", phase);
		`,
	}, {
		// (دازاي 10-25) «اريد خيار للمحاسب انو يطلب حذف الفاتورة… طلبها يوصل للمالك».
		// الفاتورة ما تنحذف إلا بموافقة المالك، وقبل الحذف تنحفظ نسختها كاملة هنا.
		Version: "0396_invoice_delete_request",
		SQL: `
			CREATE TABLE IF NOT EXISTS "LeaderInvoiceDeleteRequest" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"invoiceId" TEXT NOT NULL,
				"accountingCode" TEXT,
				"bookingCode" TEXT,
				"netTotal" NUMERIC,
				"requestedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				reason TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
				"decidedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"decidedAt" TIMESTAMP,
				"decisionNote" TEXT,
				snapshot JSONB,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE UNIQUE INDEX IF NOT EXISTS "LeaderInvoiceDeleteRequest_pending"
				ON "LeaderInvoiceDeleteRequest"("invoiceId") WHERE status = 'PENDING';
		`,
	}, {
		// (دازاي 10-26) «ما اريد نفس اليوزر يتكرر» — قفل بالقاعدة بلا فرق حروف أو فراغات.
		// ⚠️ إذا اكو تكرار قديم، ما نوكّف الإقلاع: نتركه (الكود يمنع الجديد) ونطلّع ملاحظة.
		Version: "0397_username_ci_unique",
		SQL: `
			DO $$
			BEGIN
				IF EXISTS (SELECT 1 FROM "Employee" WHERE username IS NOT NULL
				           GROUP BY lower(btrim(username)) HAVING count(*) > 1) THEN
					RAISE NOTICE 'Employee: يوزرات مكررة قديمة — القفل ما انضاف، صلّحها أول';
				ELSE
					CREATE UNIQUE INDEX IF NOT EXISTS "Employee_username_ci_key"
						ON "Employee" (lower(btrim(username))) WHERE username IS NOT NULL;
				END IF;
			END $$;
		`,
	}, {
		// (دازاي 10-26) ترتيب الصلاحيات: استثناءات معتمدة (تبقى وكت المطابقة، وممكن
		// مؤقتة بموعد)، وسجل كل تغيير (منو، شنو، ليش)، ونسخة لكل موظف حتى تعديلين
		// بنفس الوقت ما يكتب واحد فوق الثاني.
		Version: "0398_permission_access",
		SQL: `
			ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "accessVersion" INT NOT NULL DEFAULT 0;
			CREATE TABLE IF NOT EXISTS "PermissionException" (
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				permission TEXT NOT NULL,
				reason TEXT NOT NULL,
				"approvedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"expiresAt" TIMESTAMP,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now(),
				PRIMARY KEY ("employeeId", permission)
			);
			CREATE TABLE IF NOT EXISTS "PermissionAuditLog" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"employeeId" TEXT NOT NULL,
				"actorId" TEXT,
				action TEXT NOT NULL,
				permission TEXT NOT NULL,
				reason TEXT,
				"createdAt" TIMESTAMP NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "PermissionAuditLog_emp" ON "PermissionAuditLog"("employeeId", "createdAt" DESC);
		`,
	}, {
		// (دازاي 10-26) «أي شي من النظام القديم أريده يعتبر تم» — حجوزات الاستيراد
		// التاريخي (OLD-) تنسوّى بنفس علامة «تسوية قديم» (settledLegacyAt) الي يستثنيها
		// كل طابور ورق وانضباط وماتركس. ما نلمس مبالغ ولا نسوّي فواتير ولا نحذف.
		Version: "0399_settle_old_imports",
		SQL: `
			UPDATE "Booking" SET
				status = 'COMPLETED',
				"completedAt" = COALESCE("completedAt", "createdAt"),
				"settledLegacyAt" = now(),
				"settledLegacyNote" = 'من النظام القديم — يعتبر تم (قرار دازاي 10-26)',
				"awaitingReschedule" = false,
				"waitingSince" = NULL,
				"waitingKind" = NULL,
				"updatedAt" = now()
			WHERE code LIKE 'OLD-%' AND "settledLegacyAt" IS NULL
			  AND "archivedAt" IS NULL AND status <> 'CANCELLED';
		`,
	}, {
		Version: "0400_agent_names_all",
		SQL: `
			ALTER TABLE "AiAgentName" DROP CONSTRAINT IF EXISTS "AiAgentName_agent_check";
			ALTER TABLE "AiAgentName" ADD CONSTRAINT "AiAgentName_agent_check" CHECK (agent IN ('QUALITY', 'BOOKING', 'MATRIX', 'IRIS'));
		`,
	}, {
		Version: "0401_ai_brain_budget",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AiBrainBudget" (
				brain TEXT PRIMARY KEY CHECK (brain IN ('internal', 'external')),
				"dailyUSD" DOUBLE PRECISION NOT NULL CHECK ("dailyUSD" > 0),
				"updatedById" TEXT,
				"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
		`,
	}, {
		Version: "0402_ai_suggestion",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AiSuggestion" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				agent TEXT NOT NULL,
				title TEXT NOT NULL,
				details TEXT NOT NULL DEFAULT '',
				"needsCode" BOOLEAN NOT NULL DEFAULT false,
				status TEXT NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'ACCEPTED', 'REJECTED', 'DONE')),
				"decidedById" TEXT,
				"decidedAt" TIMESTAMPTZ,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "AiSuggestion_status_idx" ON "AiSuggestion" (status, "createdAt" DESC);
		`,
	}, {
		Version: "0403_ai_file",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AiFile" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				name TEXT NOT NULL,
				data BYTEA NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
		`,
	}, {
		Version: "0404_ai_reminder",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AiReminder" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				text TEXT NOT NULL,
				"dueAt" TIMESTAMPTZ NOT NULL,
				"sentAt" TIMESTAMPTZ,
				"cancelledAt" TIMESTAMPTZ,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "AiReminder_due_idx" ON "AiReminder" ("dueAt") WHERE "sentAt" IS NULL AND "cancelledAt" IS NULL;
		`,
	}, {
		Version: "0405_ai_watch",
		SQL: `
			CREATE TABLE IF NOT EXISTS "AiWatch" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"ownerId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"noteId" TEXT,
				concern TEXT NOT NULL DEFAULT '',
				baseline JSONB NOT NULL,
				status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'IMPROVED', 'WORSE', 'CLOSED')),
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"checkedAt" TIMESTAMPTZ,
				"closedAt" TIMESTAMPTZ
			);
			CREATE INDEX IF NOT EXISTS "AiWatch_open_idx" ON "AiWatch" (status, "createdAt");
		`,
	}, {
		Version: "0406_voice_lexicon",
		SQL: `
			CREATE TABLE IF NOT EXISTS "VoiceTerm" (
				term TEXT PRIMARY KEY,
				meaning TEXT NOT NULL DEFAULT '',
				category TEXT NOT NULL DEFAULT '',
				source TEXT NOT NULL DEFAULT 'SEED' CHECK (source IN ('SEED', 'SYSTEM', 'LEARNED', 'OWNER')),
				uses INT NOT NULL DEFAULT 0,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE TABLE IF NOT EXISTS "VoiceFix" (
				wrong TEXT PRIMARY KEY,
				"right" TEXT NOT NULL,
				source TEXT NOT NULL DEFAULT 'SEED',
				uses INT NOT NULL DEFAULT 0,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
		`,
	}, {
		Version: "0407_secretary_audio",
		SQL: `
			CREATE TABLE IF NOT EXISTS "SecretaryAudio" (
				"messageId" TEXT PRIMARY KEY REFERENCES "SecretaryMessage"(id) ON DELETE CASCADE,
				"contentType" TEXT NOT NULL,
				data BYTEA NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
		`,
	}, {
		// 🏛️ الهيكل الجديد (دازاي 10-29): إدارات ووحدات ومناصب. طبقة فوك الأدوار —
		// المنصب ينطي حزمة صلاحيات وحدته (مصدر ORG)، والسجل ما ينحذف.
		Version: "0408_org_structure",
		SQL: `
			CREATE TABLE IF NOT EXISTS "OrgDivision" (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				"parentId" TEXT REFERENCES "OrgDivision"(id),
				sort INT NOT NULL DEFAULT 0,
				summary TEXT NOT NULL DEFAULT ''
			);
			CREATE TABLE IF NOT EXISTS "OrgUnit" (
				id TEXT PRIMARY KEY,
				"divisionId" TEXT NOT NULL REFERENCES "OrgDivision"(id),
				name TEXT NOT NULL,
				sort INT NOT NULL DEFAULT 0
			);
			-- حزمة الصلاحيات: scope = id الإدارة (لمنصب الإدارة نفسه) أو id الوحدة
			CREATE TABLE IF NOT EXISTS "OrgBundle" (
				scope TEXT NOT NULL,
				permission TEXT NOT NULL,
				PRIMARY KEY (scope, permission)
			);
			CREATE TABLE IF NOT EXISTS "OrgAssignment" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"divisionId" TEXT NOT NULL REFERENCES "OrgDivision"(id),
				"unitId" TEXT REFERENCES "OrgUnit"(id),
				position TEXT NOT NULL CHECK (position IN ('DIRECTOR', 'DEPUTY', 'MEMBER')),
				"allowanceIQD" NUMERIC NOT NULL DEFAULT 0,
				acting BOOLEAN NOT NULL DEFAULT false,
				"actingUntil" TIMESTAMPTZ,
				"replacesId" TEXT REFERENCES "OrgAssignment"(id),
				"startedAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"endedAt" TIMESTAMPTZ,
				"endedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"endReason" TEXT,
				note TEXT,
				"createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL
			);
			-- مدير واحد ونائب واحد (أصليين) لكل إدارة/وحدة بنفس الوكت
			CREATE UNIQUE INDEX IF NOT EXISTS "OrgAssignment_head_uq" ON "OrgAssignment"
				("divisionId", COALESCE("unitId", ''), position)
				WHERE "endedAt" IS NULL AND NOT acting AND position IN ('DIRECTOR', 'DEPUTY');
			-- موظف ما يتكرر بنفس المكان
			CREATE UNIQUE INDEX IF NOT EXISTS "OrgAssignment_emp_uq" ON "OrgAssignment"
				("employeeId", "divisionId", COALESCE("unitId", ''))
				WHERE "endedAt" IS NULL;
			CREATE INDEX IF NOT EXISTS "OrgAssignment_emp" ON "OrgAssignment" ("employeeId") WHERE "endedAt" IS NULL;
			-- الصلاحيات الي انطت بسبب المنصب بس (حتى السحب ما يمس اليدوي)
			CREATE TABLE IF NOT EXISTS "OrgPermissionGrant" (
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				permission TEXT NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				PRIMARY KEY ("employeeId", permission)
			);
			-- مفاتيح وحدات بالقائمة ما چانت موجودة بالخادم (فما چان أحد يگدر ياخذها)
			INSERT INTO "Permission" (id, name, label) VALUES
				(gen_random_uuid()::text, 'unit_training', 'وحدة التدريب'),
				(gen_random_uuid()::text, 'unit_accounts', 'وحدة الحسابات'),
				(gen_random_uuid()::text, 'unit_fleet', 'وحدة العُدد والمركبات'),
				(gen_random_uuid()::text, 'unit_matrix', 'وحدة ماتركس')
			ON CONFLICT (name) DO NOTHING;
		`,
	}, {
		// 🏛️ مكتب المدير (دازاي 10-29): تقرير المشاكل ومضارها، خطة التحسين بنسبة إنجاز
		// ويا ماتركس (حل المشاكل + تطوير الأقسام)، ملاحظات النائب، وجدول الصلاحيات DOA.
		Version: "0409_org_desk",
		SQL: `
			CREATE TABLE IF NOT EXISTS "OrgIssueReport" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"divisionId" TEXT NOT NULL REFERENCES "OrgDivision"(id),
				"unitId" TEXT REFERENCES "OrgUnit"(id),
				"authorId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				title TEXT NOT NULL,
				problem TEXT NOT NULL,
				impact TEXT NOT NULL DEFAULT '',
				scope TEXT NOT NULL DEFAULT 'SPECIFIC' CHECK (scope IN ('SPECIFIC', 'GENERAL')),
				principle INT CHECK (principle BETWEEN 1 AND 20),
				severity TEXT NOT NULL DEFAULT 'MEDIUM' CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH')),
				status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'IN_PROGRESS', 'RESOLVED')),
				resolution TEXT,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"resolvedAt" TIMESTAMPTZ
			);
			CREATE INDEX IF NOT EXISTS "OrgIssueReport_div" ON "OrgIssueReport" ("divisionId", status, "createdAt" DESC);
			CREATE TABLE IF NOT EXISTS "OrgPlan" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"divisionId" TEXT NOT NULL REFERENCES "OrgDivision"(id),
				period TEXT NOT NULL,
				title TEXT NOT NULL DEFAULT '',
				"createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				UNIQUE ("divisionId", period)
			);
			CREATE TABLE IF NOT EXISTS "OrgPlanItem" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"planId" TEXT NOT NULL REFERENCES "OrgPlan"(id) ON DELETE CASCADE,
				kind TEXT NOT NULL CHECK (kind IN ('PROBLEM', 'DEVELOPMENT')),
				title TEXT NOT NULL,
				"issueId" TEXT REFERENCES "OrgIssueReport"(id) ON DELETE SET NULL,
				metric TEXT,
				baseline NUMERIC,
				target NUMERIC,
				"manualProgress" INT NOT NULL DEFAULT 0 CHECK ("manualProgress" BETWEEN 0 AND 100),
				"createdById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"doneAt" TIMESTAMPTZ
			);
			CREATE TABLE IF NOT EXISTS "OrgComment" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"targetType" TEXT NOT NULL CHECK ("targetType" IN ('ISSUE', 'PLAN_ITEM')),
				"targetId" TEXT NOT NULL,
				"authorId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				text TEXT NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "OrgComment_target" ON "OrgComment" ("targetType", "targetId", "createdAt");
			CREATE TABLE IF NOT EXISTS "OrgAuthority" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"divisionId" TEXT NOT NULL REFERENCES "OrgDivision"(id),
				decision TEXT NOT NULL,
				mode TEXT NOT NULL DEFAULT 'DECIDE' CHECK (mode IN ('DECIDE', 'ESCALATE')),
				"limitIQD" NUMERIC,
				sort INT NOT NULL DEFAULT 0,
				"updatedById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "OrgAuthority_div" ON "OrgAuthority" ("divisionId", sort);
		`,
	}, {
		// 📜 المبادئ العشرين (دازاي 10-29): تنزرع بسياسة الخصوصية والعمل، وكل موظف يوافق
		// بـ«قرأت وألتزم». الإصدار صار بصمة النص (تعديل نقطة يطلب موافقة جديدة)، وكل موافقة
		// تنحفظ بتاريخها بسجل ما ينمسح — ماتركس يرجعله وقت التقييم.
		Version: "0410_principles_acceptance",
		SQL: `
			ALTER TABLE "PrivacyPolicyPoint" ADD COLUMN IF NOT EXISTS "principleNo" INT;
			CREATE UNIQUE INDEX IF NOT EXISTS "PrivacyPolicyPoint_principle_uq" ON "PrivacyPolicyPoint" ("principleNo") WHERE "principleNo" IS NOT NULL;
			ALTER TABLE "PrivacyPolicyAcceptance" ADD COLUMN IF NOT EXISTS "pointsHash" TEXT;
			CREATE TABLE IF NOT EXISTS "PrivacyPolicyAcceptanceLog" (
				id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"acceptedAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
				"pointsHash" TEXT NOT NULL,
				"pointsCount" INT NOT NULL
			);
			CREATE INDEX IF NOT EXISTS "PrivacyPolicyAcceptanceLog_emp" ON "PrivacyPolicyAcceptanceLog" ("employeeId", "acceptedAt" DESC);
			INSERT INTO "PrivacyPolicyPoint" (id, content, "order", "principleNo")
			-- الترتيب سالب: المبادئ كتلة وحدة بأول السياسة، قبل أي نقطة موجودة أو تنضاف بعدين
			SELECT gen_random_uuid()::text, 'مبدأ ' || x.n || ': ' || x.t, x.n - 1000, x.n FROM (VALUES
				(1, 'عمل الكوادر بصورة دقيقة، أفضل من أي منافس بالسوق.'),
				(2, 'استغلال الكوادر للعمل في جميع المهام الموكلة إليهم.'),
				(3, 'الحفاظ على العُدد: أعدادها وكمياتها وتنوعها.'),
				(4, 'جلب أكبر عدد من الحجوزات للشركة.'),
				(5, 'إكمال أعمال الشركة بصورة آنية.'),
				(6, 'العمل على أن يكون العمل بأسرع ما يمكن.'),
				(7, 'العمل على أن تكون كل أدوات ومركبات الشركة في أبهى حلّة.'),
				(8, 'عدم التأخر في حجوزات الكادر.'),
				(9, 'المتابعة لكل حجوزات واتصالات الكادر في كل شيء.'),
				(10, 'العمل وفق المعايير الهندسية والتصميمية وحسب آليات الشركة.'),
				(11, 'الاحترام المتبادل للعمل ولجميع الكوادر بغض النظر عن المناصب.'),
				(12, 'العمل وفق نظام المؤسسات ووفق آليات العمل الرسمية.'),
				(13, 'تكلفة الكوادر للعمل ضمن ضوابط شركة الأماني.'),
				(14, 'العمل وفق الهيكلية التنفيذية.'),
				(15, 'الالتزام بالساعات والدوام.'),
				(16, 'كل شخص له مسؤولية العمل الخاصة به ويُحاسَب عليها.'),
				(17, 'أخذ جميع توجيهات المدير على محمل الجد.'),
				(18, 'الحفاظ على سرية العمل في كل مرحلة.'),
				(19, 'العمل كخلية نحل: كلٌّ عارف بدوره ومؤدٍّ لمهامه.'),
				(20, 'عند الإجازات تتغير المناصب مؤقتاً وفقاً لمصلحة العمل وحسب رؤية المدير.')
			) AS x(n, t)
			WHERE NOT EXISTS (SELECT 1 FROM "PrivacyPolicyPoint" pp WHERE pp."principleNo" = x.n);
		`,
	}}
}
