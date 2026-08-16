package model

import "time"

// ═══ نقاط الانضباط ═══
//
// كل موظف يبدي بـ١٠٠ نقطة، والنقطة الوحدة بعشر آلاف دينار. النظام هو
// الي يغرّم تلقائياً — مو المدير — حتى ما تصير محاباة ولا نسيان ولا
// «تعال بكرة نحچي».
//
// ⚠️ هاي نقاط منفصلة تماماً عن الكي بي اي: الكي بي اي تقييم أداء،
// وهاي انضباط إداري. ما تنخلط وياها ولا تنقص منها.
const (
	// DisciplineStartingPoints رصيد كل موظف بالبداية.
	DisciplineStartingPoints = 100
	// DisciplineDinarPerPoint قيمة النقطة الوحدة بالدينار.
	DisciplineDinarPerPoint = 10000
	// DisciplineLeaderPaperworkHours كم ساعة ننطي الليدر يخلّص فاتورته
	// وتقريره قبل ما **ينغرم هو**.
	//
	// «يتغرّم الليدر إذا ما سوّى تقرير وفاتورة للحجز خلال مدة أقصاها
	// ٢٤ ساعة، ويتغرّم الإداري إذا مرّت يومين والليدر ما مسوّي
	// الفاتورة والتقرير. هاي شغلة حتى ينجبرون يكملون الحجز».
	//
	// ⚠️ الترتيب مقصود: الي سوّى الشغل هو الي يوثّقه، فهو أول من
	// ينغرم. الغرامة كانت تروح **للإداري وحده** بعد ١٦ ساعة —
	// والليدر الي ما كتب تقريره ما يمسّه شي، فما إله سبب يستعجل.
	DisciplineLeaderPaperworkHours = 24
	// DisciplinePaperworkHours كم ساعة تمر قبل ما ينغرم **الإداري**
	// الي كلّف — لأنه المسؤول عن متابعة كادره لمن ما يلتزم.
	DisciplinePaperworkHours = 48
	// DisciplineCleanDaysToRestore كم يوم نظيف (بلا أي غرامة) لازم
	// يشتغلها الموظف حتى ترجع له نقطة وحدة.
	DisciplineCleanDaysToRestore = 3
	// DisciplineAuditHours كم ساعة ننطي المحاسب يدقّق مبلغ حجز منجز
	// قبل ما ينغرم. أطول من مهلة الورق (١٦ ساعة) لأن التدقيق يحتاج
	// الفاتورة تكون جاهزة أصلاً.
	DisciplineAuditHours = 36
)

// أنواع الأحداث
const (
	// DisciplineLatePaperwork الحجز انجز وعدّت **يومين** بلا فاتورة أو
	// تقرير. الغرامة تروح للإداري الي كلّف، لأنه المسؤول عن متابعة
	// كادره بعد ما تأخر الليدر ومرّت مهلته.
	DisciplineLatePaperwork = "LATE_PAPERWORK"
	// DisciplineLeaderLatePaperwork الحجز انجز وعدّت **٢٤ ساعة** بلا
	// فاتورة أو تقرير — والغرامة على الليدر نفسه: هو الي طلع وسوّى
	// الشغل، وهو الي يوثّقه.
	//
	// ⚠️ نوع منفصل مو نفس النوع: بدونه ما تكدر تفرّق بالسجل بين
	// «الليدر تأخر» و«الإداري ما تابع»، ولا تكدر تغرّم الاثنين على
	// نفس الحجز (الفهرس الفريد يمنع تكرار نفس النوع لنفس الحجز).
	DisciplineLeaderLatePaperwork = "LEADER_LATE_PAPERWORK"
	// DisciplineUnbalancedAssign الإداري كلّف ليدر عنده حجوزات وبنفس
	// الوقت أكو ليدر فاضي — توزيع غلط يتحاسب عليه.
	DisciplineUnbalancedAssign = "UNBALANCED_ASSIGNMENT"
	// DisciplineLateAudit الحجز انجز وعدّت ٣٦ ساعة بلا تدقيق المبلغ.
	// الغرامة على المحاسب — المبلغ الي ما ينتدقّق يبقى معلّق بالذمة،
	// وكل ما يتأخر التدقيق يصعب تتبّع الفلوس لوين راحت.
	DisciplineLateAudit = "LATE_AUDIT"
	// DisciplineRestore رجوع نقطة بعد شغل نظيف.
	DisciplineRestore = "RESTORE"
	// DisciplineManual تعديل يدوي من المالك أو مدير النظام.
	//
	// النظام يغرّم تلقائياً حتى ما تصير محاباة — بس الآلة ما تعرف كل شي:
	// الموظف ممكن يتأخر لأن الزبون ما كان بالبيت، أو ينغرم على شي مو
	// ذنبه. فلازم يكون بيد المالك مفتاح يصحّح.
	//
	// وحتى ما يصير هذا المفتاح باب خلفي: كل تعديل يدوي ينسجّل بنفس
	// السجل مع اسم الي عدّل والسبب — يعني ينشاف متل أي حركة ثانية.
	DisciplineManual = "MANUAL"
)

// DisciplinePoints رصيد نقاط موظف.
type DisciplinePoints struct {
	EmployeeID     string     `db:"employeeId" json:"employeeId"`
	Points         int        `db:"points" json:"points"`
	LastRestoredAt *time.Time `db:"lastRestoredAt" json:"lastRestoredAt"`
	UpdatedAt      time.Time  `db:"updatedAt" json:"updatedAt"`

	EmployeeName string `db:"-" json:"employeeName"`
	// المبلغ المكافئ للنقاط المخصومة بالدينار — حتى ما يحسبه أحد بيده
	DeductedDinar int `db:"-" json:"deductedDinar"`
}

// DisciplineEvent حركة وحدة على رصيد موظف (غرامة أو رجوع نقطة).
type DisciplineEvent struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	BookingID  *string   `db:"bookingId" json:"bookingId"`
	Kind       string    `db:"kind" json:"kind"`
	Delta      int       `db:"delta" json:"delta"`
	Reason     string    `db:"reason" json:"reason"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
	// منو سوّى التعديل — يمتلي بالتعديل اليدوي بس، والتلقائي يبقى فاضي.
	// ⚠️ عمود بالجدول → لازم حقل هنا (الجلب SELECT *).
	ByEmployeeID *string `db:"byEmployeeId" json:"byEmployeeId"`

	EmployeeName string `db:"-" json:"employeeName"`
}
