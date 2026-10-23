package database

// ═══ تنبيهات تقصير الإداري بتثبيت الحجز ═══
//
// المراقب يشوف حجوزات واقفة «بانتظار تواصل الإداري» وما عنده وسيلة
// يسجّل بيها إن الإداري قصّر — الشاشة عرض بحت بلا ولا زر. صار
// عنده «تسجيل تقصير»، وبالعاشر ينشر إعلان يسمّي الإداري والحجز.
//
// ⚠️⚠️ **جدول أحداث مو عمود عدّاد.** العدّاد يطلع `count(*)` منه،
// وهذا يعطي **مجاناً**: «آخر تنبيه» (MAX createdAt) و«إجمالي
// التنبيهات» و«عرض السجل» — كلهن من نفس الجدول. عمود رقم وحده
// يعطي الرقم ويضيّع **السبب ومنو سجّله ومتى**، وبعد شهر يصير
// عندك «٧/١٠» وماكو طريقة تعرف ليش.
//
// ⚠️ **و`byName` منسوخ نصاً بقصد**: لو انحذف الموظف يبقى السطر
// مقروءاً «سجّله فلان» بدل «—». سجل نصفه فاضي ما ينفع بمحاسبة.
//
// ⚠️ **يُكتب ولا يُعدّل**: «تمت المعالجة» تحطّ `resolvedAt` على
// تنبيهات الحجز، **ما تمحيهن**. المعالجة واقعة تنضاف للسجل مو
// ممحاة له — وإلا يقدر أحد ينظّف تاريخه بضغطة.
//
// ⚠️ **`bookingId` بـCASCADE و`coordinatorId` بـSET NULL**: تنبيه
// على حجز محذوف ماكو إله معنى، أما هوية الإداري فتنفقد ويبقى
// التنبيه — نفس المبدأ الي مشينا عليه بسجل الشكاوى وبلاغ التدقيق.
func coordinationAlertMigrations() []Migration {
	return []Migration{
		{
			Version: "0265_coordination_alert",
			SQL: `
				CREATE TABLE IF NOT EXISTS "CoordinationAlert" (
					id              TEXT PRIMARY KEY,
					"bookingId"     TEXT NOT NULL
						REFERENCES "Booking"(id) ON DELETE CASCADE,
					"coordinatorId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"coordinatorName" TEXT,
					reason          TEXT,
					"byEmployeeId"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"byName"        TEXT,
					"createdAt"     TIMESTAMP NOT NULL DEFAULT now(),
					"resolvedAt"    TIMESTAMP,
					"resolvedById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"resolvedByName" TEXT,
					"resolveNote"   TEXT
				);

				-- القراءة دائماً «تنبيهات هذا الحجز بالترتيب».
				CREATE INDEX IF NOT EXISTS "CoordinationAlert_booking_time_idx"
					ON "CoordinationAlert" ("bookingId", "createdAt" DESC);
			`,
		},
	}
}
