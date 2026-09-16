package database

// ═══ سجل إجراءات الشكوى ═══
//
// التصميم بيه زر «عرض السجل» بلوحة تفاصيل الشكوى. **وماكو سجل
// أصلاً**: الأثر الوحيد على الشكوى هو الصف نفسه بعد ما ينتعدّل
// (`contactedAt` · `resolvedAt` · `resolution`)، يعني نعرف **الحالة
// الأخيرة** بس ولا نعرف شنو صار قبلها ولا منو سوّاه ولا متى.
//
// وهاي بالضبط الي يحتاجها المدقق: مو «هل انتصل؟» بل «شنو صار بهذي
// الشكوى من يوم ما انفتحت».
//
// ⚠️⚠️ **يُكتب ولا يُعدّل**: ماكو مسار تعديل ولا حذف على هذا الجدول،
// بقصد. سجل يكدر أي أحد يعدّله **مو سجل** — يصير رواية. والمدقق
// يعتمد عليه بمحاسبة موظفين، فقيمته كلها بكونه ما ينلمس.
//
// ⚠️ **`byName` منسوخ نصاً وهذا تكرار مقصود**: لو انحذف الموظف
// (`byEmployeeId` تصير NULL) يبقى السطر مقروءاً «اتصل: فلان». سجل
// يصير نصفه «—» بعد سنة ما ينفع بتدقيق.
//
// ⚠️ **`ON DELETE CASCADE` على الشكوى** — هنا مقصودة عكس `auditedById`:
// سجل شكوى محذوفة ماكو إله معنى، بينما هوية المدقق تنفقد ويبقى الحكم.
//
// ⚠️⚠️ **والشكاوى القديمة ماكو إلها سجل**: هذا الجدول يبدي فاضياً،
// والأحداث تنكتب من اليوم فصاعداً. الشاشة **لازم** تكول هذا صراحةً
// — «السجل يبدي من تاريخ التفعيل» — وإلا الموظف يفتح شكوى عمرها
// شهرين، يشوف سجلاً فارغاً، ويستنتج إنه **ماكو أحد اشتغل عليها**.
// وهذا استنتاج غلط يوصّل لمحاسبة غلط.
func complaintEventMigrations() []Migration {
	return []Migration{
		{
			Version: "0261_complaint_event_log",
			SQL: `
				CREATE TABLE IF NOT EXISTS "ComplaintEvent" (
					id             TEXT PRIMARY KEY,
					"complaintId"  TEXT NOT NULL
						REFERENCES "Complaint"(id) ON DELETE CASCADE,
					kind           TEXT NOT NULL,
					detail         TEXT,
					"byEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
					"byName"       TEXT,
					"createdAt"    TIMESTAMP NOT NULL DEFAULT now()
				);

				-- القراءة دائماً «أحداث هذي الشكوى بالترتيب» — فالفهرس
				-- مركّب على الاثنين مو على المعرّف وحده.
				CREATE INDEX IF NOT EXISTS "ComplaintEvent_complaint_time_idx"
					ON "ComplaintEvent" ("complaintId", "createdAt" DESC);
			`,
		},
	}
}
