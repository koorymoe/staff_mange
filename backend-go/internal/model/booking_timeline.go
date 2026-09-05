package model

import "time"

// ═══ الخط الزمني للحجز ═══
//
// قصة الحجز كاملة بالترتيب: انسجّل ← انثبّت ← انكلّف كادر ← طلعوا ←
// انجز ← انفوتر ← انتدقّق.
//
// ⚠️ ماكو بيانات جديدة تنجمع. كل حدث هنا موجود أصلاً بجدوله — بس
// متفرق بسبع شاشات، فالي يريد يعرف «شنو صار بهذا الحجز؟» لازم يفتحهن
// كلهن ويرتّب بمخه. هنا نجمعهن ونرتّبهن بمكان واحد.

// TimelineEvent حدث واحد بالخط الزمني.
type TimelineEvent struct {
	At    time.Time `json:"at"`
	Kind  string    `json:"kind"`
	Title string    `json:"title"`
	// Detail تفصيل اختياري — فاضي ما ينعرض.
	Detail string `json:"detail,omitempty"`
	// Actor منو سوّى الحدث. ⚠️ ممكن يكون فاضي: بعض الأحداث ما تسجّل
	// فاعلها (أحداث قديمة)، و«مجهول» أصدق من اسم نخمّنه.
	Actor string `json:"actor,omitempty"`
}

// أنواع الأحداث — تحدد الأيقونة واللون بالواجهة.
const (
	TimelineCreated   = "CREATED"
	TimelineContacted = "CONTACTED"
	TimelineConfirmed = "CONFIRMED"
	TimelineAssigned  = "ASSIGNED"
	TimelineSchedule  = "SCHEDULE_CHANGE"
	TimelineStarted   = "STARTED"
	TimelineStopped   = "WORK_STOPPED"
	TimelinePartial   = "PARTIAL"
	TimelineCompleted = "COMPLETED"
	TimelineInvoiced  = "INVOICED"
	TimelineApproved  = "INVOICE_APPROVED"
	TimelineQuality   = "QUALITY"
	TimelineMonitor   = "MONITOR"
	TimelineCancelled = "CANCELLED"
	TimelinePostponed = "POSTPONED"
	TimelineWaiting   = "WAITING"
)

// DelayMetric قياس تأخير واحد.
//
// ⚠️ Minutes مؤشر (pointer) مقصود: nil = «ما ينطبق» مو «صفر». حجز ما
// وصل الفوترة ما إله «تأخر فوترة»، وعرضه صفراً يعني إنه انفوتر فوراً —
// كذبة. الواجهة تخفي الـnil كلياً.
type DelayMetric struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Minutes   *int    `json:"minutes"`
	Owner     string  `json:"owner"`
	// ThresholdMinutes الحد المعلن — تجاوزه يتلوّن بالواجهة.
	// ⚠️ معلن ومكتوب هنا، مو رقم مخبّى بالكود: صاحب العمل لازم يعرف
	// على أي أساس انتأشّر موظفه «متأخر».
	ThresholdMinutes int  `json:"thresholdMinutes"`
	Breached         bool `json:"breached"`
}

// BookingTimeline الرد الكامل.
type BookingTimeline struct {
	BookingID string          `json:"bookingId"`
	Code      string          `json:"code"`
	Events    []TimelineEvent `json:"events"`
	Delays    []DelayMetric   `json:"delays"`
}

// ═══ حدود التأخير ═══
//
// أرقام معلنة — نفس منطق غرامة الـ٣٦ ساعة الموجودة بالنظام.
// ⚠️ **عرض بس بهاي المرحلة**: ماكو غرامة تلقائية تنبني عليها. تجاوز
// الحد يتلوّن ويُقرا، والقرار يبقى بيد الإدارة.
const (
	DelayConfirmMinutes  = 24 * 60      // من التسجيل للتثبيت: يوم
	DelayAssignMinutes   = 12 * 60      // من التثبيت للتكليف: نص يوم
	DelayDepartMinutes   = 60           // من الموعد للخروج: ساعة
	DelayExecuteMinutes  = 8 * 60       // مدة التنفيذ: يوم عمل
	DelayInvoiceMinutes  = 36 * 60      // من الإنجاز للفوترة: ٣٦ ساعة
	DelayAuditMinutes    = 36 * 60      // من الفوترة للتدقيق: ٣٦ ساعة
)
