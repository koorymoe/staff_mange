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
	}}
}
