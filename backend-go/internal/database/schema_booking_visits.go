package database

// ═══ الطلعة وحدة القياس، مو الحجز ═══
//
// «الحجز الي يكتمل بشكل جزئي… وين ما أوصل بيه، حجز وحدة أريد أحسب
// إنتاجية الموظف. يعني شلون يطلع الموظف ليوم للحجز وينطي إنجاز
// جزئي؟ … أريد حتى لو الحجز نفسه طلعناله أربع أيام، كل مرة طلعناله
// تنحسب حجز للموظف، وكل مرة ينكتب بيها تاريخ وكادر طلع — لأن يجوز
// الكادر يتغيّر. المشكلة الي تصير هسه إن الطلعة الأولى تختفي ويُحسب
// بس الطلعة الثانية، يعني إنتاجية الموظف بالضيم».
//
// وهو محق تماماً، والسبب بالكود:
//
//	١) الإنتاجية تنحسب من `BookingAssignment` — وهذا جدول **الحالة
//	   الحالية** مو التاريخ: صف واحد لكل دور بالحجز. فلمن الإداري
//	   يبدّل الكادر للطلعة الثانية، صف «الفني الأول» ينكتب فوگ
//	   القديم — والكادر الأول ينمحي من الحجز وكأنه ما طلع.
//
//	٢) وحتى لو ما تبدّل الكادر: العدّ `COUNT(*) ... WHERE status =
//	   'COMPLETED'` يعدّ **الحجز** مرة وحدة. أربع طلعات = حجز واحد
//	   بالإنتاجية.
//
// `BookingProgressReport` يسجّل الطلعات الجزئية فعلاً — بس كادره
// محفوظ **كنص أسماء** (`crewSnapshot`)، والنص ما ينربط بموظف: ما
// تكدر تعدّه ولا تفلتره ولا تعرف «فلان شكد طلعة عنده هالشهر». وزيادة
// على هذا، الطلعة **الأخيرة** (الي خلّصت الحجز) ما إلها تقرير جزئي
// أصلاً، فما تنسجّل ولا بشكل نص.
//
// فصار جدولين:
//
//   - `BookingVisit` — طلعة وحدة: تاريخها، رقمها، ونتيجتها (جزئي لو
//     منجز)، ومربوطة بتقريرها الجزئي لو إله.
//   - `BookingVisitCrew` — منو طلع بهاي الطلعة **بالمعرّف** ودوره.
//
// ⚠️ ليش جدول كادر منفصل مو عمود نص؟ لأن السؤال الي نريد نجاوبه هو
// «شكد طلعة سوّاها هذا الموظف» — وهذا يحتاج `JOIN` على معرّف، والنص
// ما ينفع. وأي محاولة نطابق بالاسم تنكسر أول ما يصير موظفين بنفس
// الاسم أو ينتغيّر اسم.
//
// ⚠️ والطلعة **ما تنحذف** لمن يتبدّل الكادر: هي واقعة صارت. تبديل
// كادر اليوم الجاي يخلق طلعة جديدة، وما يلمس الي قبلها.
func bookingVisitMigrations() []Migration {
	return []Migration{
		{
			Version: "0247_booking_visits",
			SQL: `
				CREATE TABLE IF NOT EXISTS "BookingVisit" (
					id TEXT PRIMARY KEY,
					"bookingId" TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
					"visitNumber" INTEGER NOT NULL,
					-- 'PARTIAL' = يوم شغل انقفل والحجز ما خلص
					-- 'DONE'    = الطلعة الي خلّصت الحجز
					outcome TEXT NOT NULL,
					"percentDone" INTEGER,
					-- التقرير الجزئي الي طلعت منه (فاضي بطلعة الإنجاز)
					"progressReportId" TEXT REFERENCES "BookingProgressReport"(id) ON DELETE SET NULL,
					"scheduledAt" TIMESTAMPTZ,
					"occurredAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- رقم الطلعة فريد داخل الحجز: طلعتين برقم ٢ يعني عدّ مكرر
				CREATE UNIQUE INDEX IF NOT EXISTS "BookingVisit_booking_number_key"
					ON "BookingVisit" ("bookingId", "visitNumber");
				CREATE INDEX IF NOT EXISTS "BookingVisit_occurred_idx"
					ON "BookingVisit" ("occurredAt" DESC);

				CREATE TABLE IF NOT EXISTS "BookingVisitCrew" (
					id TEXT PRIMARY KEY,
					"visitId" TEXT NOT NULL REFERENCES "BookingVisit"(id) ON DELETE CASCADE,
					"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
					role TEXT NOT NULL,
					"isLeader" BOOLEAN NOT NULL DEFAULT false
				);

				CREATE UNIQUE INDEX IF NOT EXISTS "BookingVisitCrew_visit_employee_key"
					ON "BookingVisitCrew" ("visitId", "employeeId");
				-- الفهرس الي تشتغل عليه الإنتاجية: «شكد طلعة لهذا الموظف»
				CREATE INDEX IF NOT EXISTS "BookingVisitCrew_employee_idx"
					ON "BookingVisitCrew" ("employeeId");
			`,
		},
		{
			// ═══ نرجّع التاريخ الي انفقد ═══
			//
			// بدون هاي الخطوة، الإنتاجية تبدي من صفر ليوم التشغيل، وكل
			// شغل الكوادر السابق ينمحي من الحساب — وهذا بالضبط الظلم
			// الي نصلّحه، بس بالمقلوب.
			//
			// نبني الطلعات من الي موجود فعلاً:
			//   • كل تقرير إنجاز جزئي = طلعة جزئية بتاريخه.
			//   • كل حجز منجز = طلعة إنجاز بتاريخ إنجازه.
			// وكادر الطلعة المستعاد هو الكادر الحالي — أدق شي متوفر،
			// ⚠️ ومو دقيق تماماً للطلعات القديمة الي تبدّل كادرها
			// (بياناتهن ضاعت ولا تنرجع). من هنا وطالع الكادر ينلقط
			// بلحظته.
			Version: "0247_booking_visits_backfill",
			SQL: `
				INSERT INTO "BookingVisit" (id, "bookingId", "visitNumber", outcome, "percentDone", "progressReportId", "occurredAt")
				SELECT gen_random_uuid()::text, pr."bookingId", pr."dayNumber", 'PARTIAL', pr."percentDone", pr.id, pr."createdAt"
				FROM "BookingProgressReport" pr
				ON CONFLICT ("bookingId", "visitNumber") DO NOTHING;

				INSERT INTO "BookingVisit" (id, "bookingId", "visitNumber", outcome, "occurredAt")
				SELECT gen_random_uuid()::text, b.id,
				       COALESCE((SELECT MAX("visitNumber") FROM "BookingVisit" v WHERE v."bookingId" = b.id), 0) + 1,
				       'DONE', b."completedAt"
				FROM "Booking" b
				WHERE b.status = 'COMPLETED' AND b."completedAt" IS NOT NULL
				  AND NOT EXISTS (SELECT 1 FROM "BookingVisit" v WHERE v."bookingId" = b.id AND v.outcome = 'DONE')
				ON CONFLICT ("bookingId", "visitNumber") DO NOTHING;

				INSERT INTO "BookingVisitCrew" (id, "visitId", "employeeId", role, "isLeader")
				SELECT gen_random_uuid()::text, v.id, a."employeeId", a.role, COALESCE(e."isLeader", false)
				FROM "BookingVisit" v
				JOIN "BookingAssignment" a ON a."bookingId" = v."bookingId"
				JOIN "Employee" e ON e.id = a."employeeId"
				ON CONFLICT ("visitId", "employeeId") DO NOTHING;
			`,
		},
	}
}
