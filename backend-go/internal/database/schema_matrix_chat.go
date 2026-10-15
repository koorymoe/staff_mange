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
	}}
}
