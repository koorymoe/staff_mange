package database

// ══════════════════════════════════════════════════════════════════
// الإنجاز الجزئي — الحجز الي ما يخلص بيوم واحد
// ══════════════════════════════════════════════════════════════════
//
// المشكلة: عدنا حالتين بس — «تم الإنجاز» و«توقف العمل». وأكو حجوزات
// تاخذ يومين وثلاثة. فالليدر آخر اليوم يوكف بين خيارين غلط:
//
//   يأشّر «تم الإنجاز»  → الحجز ينحسب مكتمل وهو ناقص، وتطلع فاتورة
//                          على شغل ما انخلص، والغرامات تنزل عليه.
//   يأشّر «توقف العمل»   → يبين وكأن الشغل فشل أو انلغى.
//
// وبالحالتين المعلومة الأهم تضيع: **وين وصلوا؟** فالكادر الي يجي ثاني
// يوم يبدي من الصفر، يسأل الزبون «شنو سووا أمس؟»، ويعيد شغل منجز.
//
// الحل ثلاث قطع مربوطة:
//
//  ١) حالة PARTIAL — لا مكتمل ولا متوقف. الحجز يرجع لإداري الحجوزات
//     حتى ينسّقه لليوم الجاي، ويبقى بالحجوزات الفعالة مو بالأرشيف.
//
//  ٢) BookingProgressReport — تقرير كل يوم: شنو انخلص، شنو باقي، شكد
//     النسبة، وشنو العوائق. هذا الي يخلي الكادر الجاي يعرف وين يبدي.
//     نخزن معاه **لقطة الكادر** الي اشتغل ذاك اليوم، لأن التكليفات
//     تتغير لما الإداري يبدّل الكادر — ولو ما خزناها، تقرير أمس يطلع
//     منسوب للكادر الجديد.
//
//  ٣) الكادر المقترح — النظام يقترح نفس الي طلعوا أول يوم (همّه يعرفون
//     الشغل والزبون والطريق)، بس الإداري إله الحق الكامل يبدّل: ممكن
//     واحد منهم بإجازة أو مكلّف بشغل ثاني.
func partialCompletionMigration() []Migration {
	return []Migration{
		{
			// منفصلة لحالها: ALTER TYPE ADD VALUE ما يصير تُستعمل قيمته
			// بنفس المعاملة الي أضافته.
			Version: "0221_booking_status_partial",
			SQL:     `ALTER TYPE "BookingStatus" ADD VALUE IF NOT EXISTS 'PARTIAL'`,
		},
		{
			Version: "0222_booking_progress_report",
			SQL: `
				CREATE TABLE IF NOT EXISTS "BookingProgressReport" (
					id              TEXT PRIMARY KEY,
					"bookingId"     TEXT NOT NULL REFERENCES "Booking"(id) ON DELETE CASCADE,
					"dayNumber"     INTEGER NOT NULL DEFAULT 1,
					"reportedById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"workDone"      TEXT NOT NULL,
					"remainingWork" TEXT NOT NULL,
					"percentDone"   INTEGER NOT NULL DEFAULT 0,
					blockers        TEXT,
					"materialsUsed" TEXT,
					-- لقطة أسماء الكادر الي اشتغل هذا اليوم. نص جاهز مو
					-- ربط بالتكليفات: التكليفات تتبدل بالأيام الجاية،
					-- وتقرير أمس لازم يبقى منسوب لأهله.
					"crewSnapshot"  TEXT,
					"createdAt"     TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				CREATE INDEX IF NOT EXISTS "BookingProgressReport_booking_idx"
					ON "BookingProgressReport" ("bookingId", "dayNumber");
			`,
		},
		{
			Version: "0223_booking_partial_fields",
			SQL: `
				ALTER TABLE "Booking"
					-- كم مرة انأجّل لليوم الجاي. الرقم العالي إشارة: إما
					-- الشغل انقدّر غلط أو أكو مشكلة تتكرر.
					ADD COLUMN IF NOT EXISTS "partialCount" INTEGER NOT NULL DEFAULT 0,
					ADD COLUMN IF NOT EXISTS "lastPartialAt" TIMESTAMPTZ;
			`,
		},
	}
}
