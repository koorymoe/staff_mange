package database

// ══════════════════════════════════════════════════════════════════
// «أي يوم؟» — لازم يكون يوم بغداد مو يوم غرينتش
// ══════════════════════════════════════════════════════════════════
//
// النظام يخزن الأوقات بالتوقيت العالمي (شوف internal/timeutil). هذا
// صحيح ومقصود. المشكلة إن الاستعلامات كانت تسأل «هذا الحجز بأي يوم؟»
// بكتابة "scheduledAt"::date — وهذا يعطي **يوم غرينتش** مو يوم بغداد.
//
// وبغداد تسبق غرينتش بثلاث ساعات. يعني كل شي يصير ببغداد من ١٢:٠٠
// منتصف الليل لحد ٣:٠٠ الفجر، غرينتش لسه باليوم الي قبله:
//
//   الإداري رحّل الحجز:  الجمعة ١٤ آب، ١:٣٠ فجراً
//   انخزن (عالمي):        الخميس ١٣ آب، ٢٢:٣٠
//   الواجهة تعرضه:        الجمعة ١٤  ✅ (تحوّل لبغداد)
//   الفلتر بالسيرفر:      الخميس ١٣  ❌ (ما يحوّل)
//
// فالحجز ينعرض «الجمعة» بس يطلع بعمود الخميس بجدول التنسيق، ويخربط
// ترتيب بقية اليوم. ونفس الخلل بكل الإحصاءات اليومية: شغل الساعة ١
// فجراً ينحسب على اليوم الي فات.
//
// الحل: دالة وحدة تحوّل لبغداد قبل ما تاخذ التاريخ، ونستعملها بكل
// مكان بدل ::date. العراق ما عنده توقيت صيفي من ٢٠١٥، فالفرق ثابت
// +٣ دائماً — وهذا يخلي الدالة IMMUTABLE يعني تنفع بالفهارس بعد.
func baghdadDateMigration() []Migration {
	return []Migration{
		{
			Version: "0218_baghdad_date_functions",
			SQL: `
				CREATE OR REPLACE FUNCTION baghdad_date(ts timestamp without time zone)
				RETURNS date LANGUAGE sql IMMUTABLE PARALLEL SAFE AS
				$$ SELECT (ts + interval '3 hours')::date $$;

				CREATE OR REPLACE FUNCTION baghdad_date(ts timestamp with time zone)
				RETURNS date LANGUAGE sql STABLE PARALLEL SAFE AS
				$$ SELECT (ts AT TIME ZONE 'UTC' + interval '3 hours')::date $$;

				-- «اليوم» بنظر بغداد مو بنظر السيرفر
				CREATE OR REPLACE FUNCTION baghdad_today()
				RETURNS date LANGUAGE sql STABLE PARALLEL SAFE AS
				$$ SELECT (NOW() AT TIME ZONE 'UTC' + interval '3 hours')::date $$;
			`,
		},
		{
			// فهرس على يوم بغداد — بدونه فلترة يوم وحد تقرا كل جدول
			// الحجوزات. الفهرس القديم على "scheduledAt" ما ينفع لأن
			// الشرط صار على دالة مو على العمود نفسه.
			Version: "0219_baghdad_date_indexes",
			SQL: `
				CREATE INDEX IF NOT EXISTS "Booking_scheduledAt_baghdad_idx"
					ON "Booking" (baghdad_date("scheduledAt"))
					WHERE "scheduledAt" IS NOT NULL;

				CREATE INDEX IF NOT EXISTS "Booking_completedAt_baghdad_idx"
					ON "Booking" (baghdad_date("completedAt"))
					WHERE "completedAt" IS NOT NULL;
			`,
		},
	}
}
