package repository

// ═══ الحجز المطلوب حذفه — تعريف واحد لكل النظام ═══
//
// ⚠️ العلّة الي انصلحت هنا: الشرط چان معرَّفاً بمكان واحد بس
// (`booking_page_repository.go`) ومطبَّقاً على شاشة «الحجوزات» وحدها.
// كل مصدر ثاني (التنسيق، الرئيسية، «شنو صاير اليوم»، تتبع المهام)
// يفلتر `archivedAt IS NULL` بس — **بلا أي ذكر لطلبات الحذف**.
//
// فالحجز المطلوب حذفه چان **يختفي من شاشة الحجوزات ويضل ظاهراً
// بالتنسيق وبالرئيسية**: الإداري يشتغل على حجز يمكن ينحذف بعد ساعة،
// وبطاقة «بانتظار التثبيت» تعطي **رقمين مختلفين** بشاشتين — نفس علّة
// «رقمان بنفس الاسم» الي نطاردها بكل النظام.
//
// هسه التعريف هنا، ويُستدعى من كل مكان. نسخة وحدة ما تفترق.
//
// qualifier = اسم الجدول أو اللقب الي ينشار بيه لعمود id
// (مثلاً `b` لو الاستعلام يستعمل لقباً، أو `"Booking"` لو بلا لقب).
func BookingDeletePendingSQL(qualifier string) string {
	return `EXISTS (SELECT 1 FROM "BookingDeleteRequest" dr
		WHERE dr."bookingId" = ` + qualifier + `.id AND dr.status = 'PENDING')`
}

// NotDeletePendingSQL نفس الشرط بالنفي — للاستعمال المباشر بـWHERE.
func NotDeletePendingSQL(qualifier string) string {
	return ` AND NOT ` + BookingDeletePendingSQL(qualifier)
}

// BookingCountableSQL نطاق «الحجز الي ينعدّ» — تعريف **واحد** لكل
// النظام، ومطابق حرفياً للنطاق الي ترجّعه `BookingRepository.List`
// (يعني الي تعرضه الشاشات فعلاً).
//
// 🔴 ليش لازم يكون واحداً: الشاشات تستثني المؤرشف والمطلوب حذفه،
// وطبقة الإحصائيات چانت ما تستثني ولا واحد منهم. فنفس المفهوم يطلع
// برقمين: اللوحة تگول «٦٢٠ حجز» والشاشة تعرض ٥٤٠، والفرق حجوزات
// محذوفة يعدّها العدّاد وما يعرضها الضغط عليه. ورقم غلط أسوأ من
// ماكو رقم.
//
// يُستعمل بعد `WHERE` مباشرة (يرجّع شرطاً موجباً بلا AND بادئة).
func BookingCountableSQL(qualifier string) string {
	return qualifier + `."archivedAt" IS NULL AND NOT ` + BookingDeletePendingSQL(qualifier)
}

// BookingCountableAndSQL نفس الشرط بـAND بادئة — للاستعمال بعد شرط
// موجود بـWHERE.
func BookingCountableAndSQL(qualifier string) string {
	return ` AND ` + BookingCountableSQL(qualifier)
}
