package database

// ═══ الجرد يصير قبل كل حجز مو مرة باليوم ═══
//
// «الموظف يحتاج يجرد أدواته **قبل كل حجز**، والليدر يجرد أدواته
// ويشوف منو من الموظفين الي راح يطلعون وياه بهذا الحجز جرد».
//
// الجرد اليومي ما يجاوب السؤال المهم: الفني جرد الصبح، وطلع بثلاث
// حجوزات، ونسى الميتر بالحجز الثاني — الجرد يگول «كامل» وهو مو كامل
// وقت الحجز الثالث. والأهم: لما تنقص أداة عند الزبون، ماكو شي يربط
// النقص بالحجز الي صار بيه.
//
// الحل: عمود واحد. الجرد ينربط بالحجز.
//
// ⚠️ العمود **يقبل الفراغ** عن قصد:
//   - كل الجرود القديمة ما إلها حجز، وتصفيرها يعني نمحي تاريخ الشركة.
//   - والجرد اليومي العام يبقى مشروع لمن يريده (فني ما عنده حجز
//     اليوم بس يريد يأشّر عدته كاملة).
//
// ⚠️ والفهرس الفريد **جزئي** (WHERE bookingId IS NOT NULL): بدونه
// الجرود القديمة الي كلها NULL تتعارض مع بعضها ويفشل إنشاء الفهرس
// على قاعدة فيها بيانات.
func inventoryPerBookingMigrations() []Migration {
	return []Migration{
		{
			Version: "0246_inventory_check_booking",
			SQL: `
				ALTER TABLE "InventoryCheck"
					ADD COLUMN IF NOT EXISTS "bookingId" TEXT
					REFERENCES "Booking"(id) ON DELETE CASCADE;

				-- جرد واحد لكل موظف بكل حجز — إعادة الجرد تحدّث الموجود
				-- بدل ما تكدّس صفوف وتخلي «منو جرد؟» جواب متعدد.
				CREATE UNIQUE INDEX IF NOT EXISTS "inventory_check_booking_employee_key"
					ON "InventoryCheck" ("bookingId", "employeeId")
					WHERE "bookingId" IS NOT NULL;

				-- شاشة الليدر تسأل: «منو جرد بهذا الحجز؟» — بلا فهرس
				-- تصير مسحة كاملة للجدول كل ما يفتح الشاشة.
				CREATE INDEX IF NOT EXISTS "inventory_check_booking_idx"
					ON "InventoryCheck" ("bookingId");
			`,
		},
	}
}
