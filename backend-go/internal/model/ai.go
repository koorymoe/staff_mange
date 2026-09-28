package model

import "time"

// ═══ نواة الذكاء الاصطناعي — النماذج ═══
//
// ⚠️ اقرا التعليق برأس schema_ai_core.go قبل ما تلمس هذا الملف:
// الفصل بين **الأدلة** (حقائق نحسبها) و**الحكم** (تفسير) هو أساس
// التصميم كله، وأي خلط بينهم يرجّعنا لتخمين بثقة.

// ═══ الإشارة ═══

type AiSignal struct {
	ID         string    `db:"id" json:"id"`
	Kind       string    `db:"kind" json:"kind"`
	EntityType string    `db:"entityType" json:"entityType"`
	EntityID   string    `db:"entityId" json:"entityId"`
	EmployeeID *string   `db:"employeeId" json:"employeeId,omitempty"`
	Payload    []byte    `db:"payload" json:"-"`
	Status     string    `db:"status" json:"status"`
	OccurredAt time.Time `db:"occurredAt" json:"occurredAt"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`

	EmployeeName *string     `db:"-" json:"employeeName,omitempty"`
	Evidence     *AiEvidence `db:"-" json:"evidence,omitempty"`
	Verdict      *AiVerdict  `db:"-" json:"verdict,omitempty"`
}

// أنواع الإشارات — اللحظات الي تستاهل تحليل.
const (
	// الموظف وقّف الشغل بنص الحجز. المثال الي وصفه صاحب العمل بالضبط.
	AiSignalWorkStopped = "WORK_STOPPED"
	// خرج للزبون متأخر عن الموعد.
	AiSignalLateStart = "LATE_START"
	// الحجز انأجّل أكثر من المعتاد.
	AiSignalRepeatPostpone = "REPEAT_POSTPONE"
	// المحاسب عدّل مبالغ فاتورة.
	AiSignalInvoiceAdjusted = "INVOICE_ADJUSTED"
	// إنجاز جزئي متكرر لنفس الحجز.
	AiSignalRepeatPartial = "REPEAT_PARTIAL"
	// تقرير إنجاز يدّعي شغلاً لحاله ("وحدي"/"لحالي") وكادر الحجز
	// الحقيقي المسجّل أكثر من واحد — تناقض رقمي قابل للإثبات، مو حكماً
	// على أسلوب الكتابة. شوف SelfReportMismatchEvidence تحت.
	AiSignalSelfReportMismatch = "SELF_REPORT_MISMATCH"
	// شذوذ تكلفة تعبئة وقود — الكاشف (`VehicleService.CheckFuelAnomaly`)
	// موجود من زمان، هذا وصله بخط أنابيب ماتركس.
	AiSignalFuelAnomaly = "VEHICLE_FUEL_ANOMALY"
	// زبون اشتكى مرتين أو أكثر خلال ٦٠ يوم على شغل نفس الليدر.
	AiSignalCustomerRepeatComplaint = "CUSTOMER_REPEAT_COMPLAINT"
	// فاتورة ليدر بكمية مادة أكثر من ١.٥× وسيط نفس الخدمة.
	AiSignalMaterialOveruse = "MATERIAL_OVERUSE"
	// فاتورة ليدر بكمية مادة أقل من ٠.٥× وسيط نفس الخدمة.
	AiSignalMaterialUnderuse = "MATERIAL_UNDERUSE"
	// فاتورة ليدر أرقامها ما تطابق بيانات الحجز المنظّمة (عدد الأجهزة
	// المسجّل بالحجز، أو المبلغ المقدّر).
	AiSignalInvoiceWorkMismatch = "INVOICE_WORK_MISMATCH"
	// زبون قرب يزعل: عاملين خطر أو أكثر (تأجيلات، تأخر، شكوى مفتوحة، تقييم واطي).
	AiSignalCustomerAtRisk = "CUSTOMER_AT_RISK"
	// صافي فاتورة ليدر شاذ عن وسيط فواتير نفس الخدمة (فوق ٢× أو تحت ٠.٤×).
	AiSignalPriceOutlier = "PRICE_OUTLIER"
	// حجوزات منجزة ناقصها فاتورة أو تقرير بعد ٤٨ ساعة — مجمّعة لكل ليدر بالأسبوع.
	AiSignalLatePaperwork = "LATE_PAPERWORK"
	// الحضور ما يطابق الشغل ليوم (حاضر بلا شغل، أو شغل بلا حضور) — ممكن نقص تسجيل.
	AiSignalAttendanceWorkGap = "ATTENDANCE_WORK_GAP"
)

// ⚠️ «مركبة صيانتها متأخرة ولسه تُرسل» **ما ينحط** بخط أنابيب
// AiSignal: هذا حالة مستمرة مو حدثاً لمرة وحدة، وإعادة الفحص الدوري
// كل مرة تسجّل إشارة جديدة بمعرّف/وقت مختلف تغرق الصندوق بتكرار نفس
// المشكلة. نفس سبب حماية الإرهاق ومؤشر الطاقة — تنبيه دوري مباشر
// للمالك ومدير النظام، مو إشارة تحتاج "أوافق/عندي ملاحظة".

func AiSignalLabel(kind string) string {
	switch kind {
	case AiSignalWorkStopped:
		return "توقف عمل"
	case AiSignalLateStart:
		return "تأخر بالخروج"
	case AiSignalRepeatPostpone:
		return "تأجيل متكرر"
	case AiSignalInvoiceAdjusted:
		return "تعديل مبالغ فاتورة"
	case AiSignalRepeatPartial:
		return "إنجاز جزئي متكرر"
	case AiSignalSelfReportMismatch:
		return "تقرير إنجاز يناقض كادر الحجز الحقيقي"
	case AiSignalFuelAnomaly:
		return "شذوذ بتكلفة تعبئة وقود"
	case AiSignalCustomerRepeatComplaint:
		return "شكاوى متكررة من نفس الزبون على نفس الليدر"
	case AiSignalMaterialOveruse:
		return "استهلاك مواد أعلى من المعتاد"
	case AiSignalMaterialUnderuse:
		return "استهلاك مواد أقل من المعتاد"
	case AiSignalInvoiceWorkMismatch:
		return "فاتورة ما تطابق بيانات الحجز"
	case AiSignalCustomerAtRisk:
		return "زبون قرب يزعل"
	case AiSignalPriceOutlier:
		return "تسعير شاذ"
	case AiSignalLatePaperwork:
		return "الورق المتأخر"
	case AiSignalAttendanceWorkGap:
		return "الحضور مقابل الشغل"
	}
	return kind
}

// ═══ الأدلة ═══

type AiEvidence struct {
	ID          string    `db:"id" json:"id"`
	SignalID    string    `db:"signalId" json:"signalId"`
	Facts       []byte    `db:"facts" json:"-"`
	Gaps        []byte    `db:"gaps" json:"-"`
	CollectedAt time.Time `db:"collectedAt" json:"collectedAt"`

	// مفكوكة للواجهة
	FactsMap map[string]any `db:"-" json:"facts"`
	GapsList []string       `db:"-" json:"gaps"`
}

// WorkStopEvidence الأدلة الي نجمعها لتوقف العمل.
//
// صاحب العمل وصف المسار بالضبط: «يروح يشوف سلة الزبون… نوب يروح
// لأبو الكميات يشوف هذا الموظف طالب شي… ترجع تشوف شوكت انضافت
// المادة الجديدة على السلة، قبل الحجز لو بأثناء الحجز؟».
//
// كل حقل هنا **ينجمع بالكود من الجداول**، ماكو ولا تخمين:
type WorkStopEvidence struct {
	// شنو گال الموظف
	StopReason string `json:"stopReason"`
	// وقت التوقف بتوقيت بغداد وكم باقي على نهاية الدوام.
	// ⚠️ هذا الي يفرّق بين «الوقت ما يكفي» صادقة وكاذبة.
	StoppedAtHour     int `json:"stoppedAtHour"`
	MinutesToShiftEnd int `json:"minutesToShiftEnd"`
	// شغل فعلي قبل التوقف — توقف بعد ١٠ دقايق غير توقف بعد ٥ ساعات
	WorkedMinutes int `json:"workedMinutes"`

	// ═══ خيط المواد ═══
	// هل طلب مادة من إداري الكميات أصلاً؟
	ProcurementRequests int    `json:"procurementRequests"`
	LastRequestStatus   string `json:"lastRequestStatus,omitempty"`
	// ⚠️ هذا المفتاح: طلب ووفّرها → المشكلة مو منه.
	//    طلب وما وفّرها → المسؤولية على إداري الكميات.
	//    ما طلب أصلاً → إما نسيان منه أو الزبون طلب شي جديد.
	RequestedBeforeStop bool `json:"requestedBeforeStop"`

	// ═══ خيط السلة ═══
	// انضافت مادة للسلة **بعد** ما بدأ الشغل؟ يعني الزبون طلب شي
	// جديد بالموقع — والموظف مو مقصّر.
	CartItemsTotal      int `json:"cartItemsTotal"`
	CartItemsAfterStart int `json:"cartItemsAfterStart"`

	// ═══ سجل الموظف ═══
	// نفس الموظف وقّف كم مرة بآخر ٣٠ يوم؟ مرة = ظرف، خمس مرات = نمط.
	StopsLast30Days int `json:"stopsLast30Days"`
	// ⚠️ نصف التصعيد الثاني — التساهل: شكد يوم مرّ من آخر توقف لنفس
	// الموظف قبل هذا. nil يعني أول مرة إطلاقاً (ماكو فترة نظيفة
	// نحچي عنها أصلاً).
	DaysSinceLastStop *int `json:"daysSinceLastStop,omitempty"`
	// TenureDays: عدالة الموظف الجديد — منحنى تعلّم. موظف بأول أسابيعه
	// معدل أخطائه الطبيعي أعلى من موظف خبير، فالمقارنة بميزان وحد
	// ظلم. nil يعني ماكو تاريخ تعيين مسجّل — «ما نعرف» مو «خبير».
	TenureDays *int `json:"tenureDays,omitempty"`
}

// LateStartEvidence الأدلة لتأخر الخروج للحجز.
//
// ⚠️ الحساب نفسه موجود أصلاً بجدول الخط الزمني (`DelayMetric` بمفتاح
// "DEPART") — هنا يوصل للتحليل بدل ما يبقى رقماً بشاشة وحدها.
type LateStartEvidence struct {
	// شكد دقيقة تأخر عن الموعد المجدول، والحد المعلن (نفس
	// DelayDepartMinutes — ساعة).
	MinutesLate      int `json:"minutesLate"`
	ThresholdMinutes int `json:"thresholdMinutes"`
	// نفس الموظف تأخر كم مرة بآخر ٣٠ يوم؟
	LateCountLast30Days int `json:"lateCountLast30Days"`
	// نفس فكرة `DaysSinceLastStop` أعلاه — بس لتأخر الخروج.
	DaysSinceLastLate *int `json:"daysSinceLastLate,omitempty"`
	// نفس `TenureDays` بـ`WorkStopEvidence` — عدالة الموظف الجديد.
	TenureDays *int `json:"tenureDays,omitempty"`
}

// RepeatPostponeEvidence الأدلة لتأجيل نفس الحجز أكثر من مرة.
type RepeatPostponeEvidence struct {
	PostponeCount int `json:"postponeCount"`
	// آخر سبب تأجيل مكتوب — قرار الأجّل بيد الزبون غير قرار داخلي.
	LastReason string `json:"lastReason,omitempty"`
	// انأجّل بلا موعد بديل؟ يعني الزبون نفسه ما محدّد وقته — نمط
	// مختلف عن حجز يتأجل داخلياً كل مرة.
	AwaitingReschedule bool `json:"awaitingReschedule"`
}

// InvoiceAdjustedEvidence الأدلة لتعديل مبالغ فاتورة بعد ما انسجّلت.
//
// ⚠️ الفاتورة مسجّلة باسم الليدر (معرّف الإشارة) والتعديل يصير من
// المحاسب — فالإشارة عن **الليدر** الي فاتورته احتاجت تصحيح، مو عن
// المحاسب الي صحّحها.
type InvoiceAdjustedEvidence struct {
	OldNetTotal float64 `json:"oldNetTotal"`
	NewNetTotal float64 `json:"newNetTotal"`
	// سالب = نزلت (الليدر بالغ بالتقدير)، موجب = زادت (نقّص).
	DifferenceAmount float64 `json:"differenceAmount"`
	DifferencePct    float64 `json:"differencePct"`
	Reason           string  `json:"reason"`
	// كم مرة انعدّلت فاتورة نفس هذا الليدر بآخر ٣٠ يوم؟
	AdjustCountLast30Days int `json:"adjustCountLast30Days"`
}

// RepeatPartialEvidence الأدلة لإنجاز جزئي متكرر بنفس الحجز.
type RepeatPartialEvidence struct {
	PartialCount int `json:"partialCount"`
	// آخر نسبة إنجاز مسجّلة، وشنو باقي وشنو المعوّقات حسب آخر تقرير.
	LastPercentDone int    `json:"lastPercentDone"`
	LastRemaining   string `json:"lastRemaining,omitempty"`
	LastBlockers    string `json:"lastBlockers,omitempty"`
}

// SelfReportMismatchEvidence الأدلة لتقرير إنجاز يدّعي شغلاً لحاله
// وكادر الحجز الحقيقي أكثر من واحد.
//
// 🔴 القيد الحاكم هنا: **ما نحكم على أسلوب الكتابة**. الدليل الوحيد
// المقبول جملة صريحة صادفناها بالتقرير ("وحدي"/"لحالي"/"بروحي"/
// "بمفردي") مقابل رقم حقيقي محسوب بالكود (كادر آخر طلعة). لو ماكو
// جملة صريحة أو الكادر فعلاً واحد، ماكو إشارة أصلاً — صمت مو تخمين.
type SelfReportMismatchEvidence struct {
	// الجملة الي التقطناها من التقرير — دليل حرفي مو استنتاج.
	ClaimPhrase string `json:"claimPhrase"`
	ReportText  string `json:"reportText"`
	// عدد الكادر الحقيقي المسجّل بآخر طلعة لهذا الحجز.
	ActualCrewSize int    `json:"actualCrewSize"`
	BookingCode    string `json:"bookingCode,omitempty"`
}

// FuelAnomalyEvidence أدلة شذوذ تكلفة تعبئة وقود — نفس منطق
// `VehicleService.CheckFuelAnomaly` الموجود من زمان (متوسط آخر ٥
// تعبئات)، بس محسوبة وقت الجمع للتحليل مو للعرض الفوري بس.
type FuelAnomalyEvidence struct {
	VehiclePlate string  `json:"vehiclePlate"`
	NewCost      float64 `json:"newCost"`
	AverageCost  float64 `json:"averageCost"`
	PercentAbove float64 `json:"percentAbove"`
}

// RepeatComplaintEvidence أدلة شكاوى متكررة — معرّفات وتواريخ وأكواد
// حجوزات بس، بلا اسم أو هاتف زبون.
type RepeatComplaintEvidence struct {
	Count        int                   `json:"count"`
	WindowDays   int                   `json:"windowDays"`
	CustomerCode string                `json:"customerCode,omitempty"`
	Complaints   []RepeatComplaintItem `json:"complaints"`
}

type RepeatComplaintItem struct {
	ComplaintID string    `json:"complaintId" db:"id"`
	CreatedAt   time.Time `json:"createdAt" db:"createdAt"`
	BookingCode string    `json:"bookingCode,omitempty" db:"bookingCode"`
	Type        string    `json:"type,omitempty" db:"type"`
}

// MaterialUsageEvidence أدلة استهلاك مواد شاذ بفاتورة ليدر — مقارنة
// بوسيط فواتير نفس الخدمة (آخر ١٨٠ يوم، بلا المسودات).
type MaterialUsageEvidence struct {
	Direction          string              `json:"direction"` // OVER | UNDER
	InvoiceID          string              `json:"invoiceId"`
	ServiceID          string              `json:"serviceId,omitempty"`
	Lines              []MaterialUsageLine `json:"lines"`
	SameKindLast30Days int                 `json:"sameKindLast30Days"`
}

type MaterialUsageLine struct {
	MaterialID string  `json:"materialId" db:"materialId"`
	Quantity   float64 `json:"quantity" db:"quantity"`
	Median     float64 `json:"median" db:"median"`
	Samples    int     `json:"samples" db:"samples"`
}

// InvoiceWorkMismatchEvidence أدلة عدم مطابقة فاتورة ليدر لبيانات
// حجزها المنظّمة — أرقام وأكواد بس، بلا اسم زبون أو هاتف.
type InvoiceWorkMismatchEvidence struct {
	InvoiceID          string               `json:"invoiceId"`
	AccountingCode     string               `json:"accountingCode,omitempty"`
	BookingCode        string               `json:"bookingCode,omitempty"`
	InvoiceDeviceCount int                  `json:"invoiceDeviceCount"`
	BookedDeviceCount  *int                 `json:"bookedDeviceCount,omitempty"`
	InvoiceNetTotal    float64              `json:"invoiceNetTotal"`
	QuotedPrice        *float64             `json:"quotedPrice,omitempty"`
	Findings           []InvoiceWorkFinding `json:"findings"`
}

// InvoiceWorkFinding فرق واحد: Kind = DEVICE_COUNT | QUOTED_PRICE.
type InvoiceWorkFinding struct {
	Kind     string  `json:"kind"`
	Invoiced float64 `json:"invoiced"`
	Booked   float64 `json:"booked"`
}

// CustomerAtRiskEvidence عوامل خطر زبون — أكواد وأرقام بس.
type CustomerAtRiskEvidence struct {
	CustomerCode      string               `json:"customerCode,omitempty"`
	LatestBookingCode string               `json:"latestBookingCode,omitempty"`
	Factors           []CustomerRiskFactor `json:"factors"`
}

// CustomerRiskFactor عامل واحد: Kind = POSTPONED | OVERDUE | OPEN_COMPLAINT | LOW_RATING.
type CustomerRiskFactor struct {
	Kind  string `json:"kind"`
	Value int    `json:"value"`
	Ref   string `json:"ref,omitempty"` // كود حجز لو ينطبق
}

// PriceOutlierEvidence صافي فاتورة مقابل وسيط نفس الخدمة.
type PriceOutlierEvidence struct {
	InvoiceID      string  `json:"invoiceId"`
	AccountingCode string  `json:"accountingCode,omitempty"`
	BookingCode    string  `json:"bookingCode,omitempty"`
	ServiceID      string  `json:"serviceId,omitempty"`
	NetTotal       float64 `json:"netTotal"`
	Median         float64 `json:"median"`
	Samples        int     `json:"samples"`
	Ratio          float64 `json:"ratio"`
	Direction      string  `json:"direction"` // HIGH | LOW | ""
}

// LatePaperworkEvidence حجوزات ليدر منجزة وورقها ناقص بعد ٤٨ ساعة.
type LatePaperworkEvidence struct {
	WeekStart string              `json:"weekStart"`
	LateCount int                 `json:"lateCount"`
	Bookings  []LatePaperworkItem `json:"bookings"`
}

type LatePaperworkItem struct {
	BookingCode    string `json:"bookingCode" db:"bookingCode"`
	MissingInvoice bool   `json:"missingInvoice" db:"missingInvoice"`
	MissingReport  bool   `json:"missingReport" db:"missingReport"`
	HoursSinceDone int    `json:"hoursSinceDone" db:"hoursSinceDone"`
}

// AttendanceWorkGapEvidence فرق حضور/شغل ليوم واحد.
// Kind = PRESENT_NO_WORK (حاضر ≥٦ ساعات بلا تكليف) أو WORK_NO_ATTENDANCE.
type AttendanceWorkGapEvidence struct {
	Day             string   `json:"day"`
	Kind            string   `json:"kind"`
	AttendedMinutes int      `json:"attendedMinutes"`
	BookingCodes    []string `json:"bookingCodes"`
}

// ═══ الحكم ═══

type AiVerdict struct {
	ID              string    `db:"id" json:"id"`
	SignalID        string    `db:"signalId" json:"signalId"`
	Source          string    `db:"source" json:"source"`
	ModelName       *string   `db:"modelName" json:"modelName,omitempty"`
	Headline        string    `db:"headline" json:"headline"`
	Reasoning       *string   `db:"reasoning" json:"reasoning,omitempty"`
	Confidence      int       `db:"confidence" json:"confidence"`
	Severity        string    `db:"severity" json:"severity"`
	BlameEmployeeID *string   `db:"blameEmployeeId" json:"blameEmployeeId,omitempty"`
	Suggestion      *string   `db:"suggestion" json:"suggestion,omitempty"`
	CreatedAt       time.Time `db:"createdAt" json:"createdAt"`

	BlameEmployeeName *string `db:"-" json:"blameEmployeeName,omitempty"`
}

const (
	AiSourceRules = "RULES" // محرّك القواعد عدنا — شغّال اليوم
	AiSourceModel = "MODEL" // منصّة خارجية — لما ننشترك

	AiSeverityInfo     = "INFO"
	AiSeverityWatch    = "WATCH"
	AiSeverityWarn     = "WARN"
	AiSeverityCritical = "CRITICAL"
)

// ⚠️ تحت هذا الحد الحكم يتأشر «مو متأكد» بالواجهة ولا ينبنى عليه
// قرار. رقم معلن أحسن من عتبة مخبّاية بالكود.
const AiConfidenceTrusted = 70

// MonitorFeedbackExample حكم سابق + قرار المراقب الحقيقي عليه —
// مادة التعلّم. ⚠️ ماكو جدول جديد لهذا: قرار المراقب مخزون أصلاً
// بصندوقه (`MonitorReview`، محطة `AI_VERDICT`) — هذا استعلام قراءة بس.
type MonitorFeedbackExample struct {
	Headline  string  `db:"headline" json:"headline"`
	Reasoning *string `db:"reasoning" json:"reasoning,omitempty"`
	Facts     []byte  `db:"facts" json:"-"`
	// MonitorStatus: OK (ما اكو مشكلة حقيقية) أو FLAGGED (فعلاً مشكلة).
	MonitorStatus string  `db:"monitorStatus" json:"monitorStatus"`
	MonitorNote   *string `db:"monitorNote" json:"monitorNote,omitempty"`
}

// ═══ المؤشرات ═══

type AiMetric struct {
	ID          string    `db:"id" json:"id"`
	MetricKey   string    `db:"metricKey" json:"metricKey"`
	Scope       string    `db:"scope" json:"scope"`
	ScopeID     *string   `db:"scopeId" json:"scopeId,omitempty"`
	PeriodStart time.Time `db:"periodStart" json:"periodStart"`
	PeriodEnd   time.Time `db:"periodEnd" json:"periodEnd"`
	Value       float64   `db:"value" json:"value"`
	SampleCount int       `db:"sampleCount" json:"sampleCount"`
	Details     []byte    `db:"details" json:"-"`
	ComputedAt  time.Time `db:"computedAt" json:"computedAt"`

	ScopeName *string `db:"-" json:"scopeName,omitempty"`
	// DetailsMap: نسخة مفكوكة من Details للواجهة — نفس نمط FactsMap
	// بـAiEvidence. تُملى بالمستودع (ListMetrics)، مو هنا.
	DetailsMap map[string]any `db:"-" json:"details,omitempty"`
}

// مفاتيح المؤشرات — تنحسب من الأدلة مو عدّادات خام.
const (
	AiMetricStopRate         = "STOP_RATE"          // نسبة الحجوزات الي وقّفت
	AiMetricStopMinutesAvg   = "STOP_MINUTES_AVG"   // متوسط الوقت الضايع بالتوقف
	AiMetricMaterialMissRate = "MATERIAL_MISS_RATE" // توقف بسبب مادة ما انطلبت
	AiMetricScopeCreepRate   = "SCOPE_CREEP_RATE"   // الزبون طلب زيادة بالموقع
	AiMetricProcurementDelay = "PROCUREMENT_DELAY"  // تأخر إداري الكميات
	AiMetricLateStartRate    = "LATE_START_RATE"    // نسبة التأخر بالخروج
	// AiMetricMonitorAgreement: شكد من أحكام ماتركس المراقب وافق
	// عليها (ما اكو مشكلة حقيقية) — هذا مقياس الدقة الي طلبه صاحب
	// النظام: «وافق المراقب على ٨٢٪ من أحكام الأسبوع».
	AiMetricMonitorAgreement = "MONITOR_AGREEMENT_RATE"
	// AiMetricEnergyUsed: مؤشر الطاقة المستخدمة الشهري — نسبة (حجوزات
	// + ساعات دوام + حجوزات صيانة) الموظف مقابل متوسط زملائه. ١ = نفس
	// المعدل، فوق ١ = طاقة أعلى من المعتاد.
	AiMetricEnergyUsed = "ENERGY_USED_SCORE"
)

func AiMetricLabel(key string) string {
	switch key {
	case AiMetricStopRate:
		return "نسبة توقف العمل"
	case AiMetricStopMinutesAvg:
		return "متوسط الوقت الضايع بالتوقف"
	case AiMetricMaterialMissRate:
		return "توقف بسبب مادة ما انطلبت"
	case AiMetricScopeCreepRate:
		return "زيادة طلبات الزبون بالموقع"
	case AiMetricProcurementDelay:
		return "تأخر توفير المواد"
	case AiMetricLateStartRate:
		return "نسبة التأخر بالخروج للزبون"
	case AiMetricMonitorAgreement:
		return "دقّة ماتركس — وافق المراقب على أحكامه"
	case AiMetricEnergyUsed:
		return "الطاقة المستخدمة الشهرية"
	}
	return AiDiscoveryMetricLabel(key)
}

// ═══ الاستكشاف الأسبوعي — محاور وشبكة أنماط ═══
//
// ⚠️ هذا يجاوب «ما أريد أشرحله بالمضبوط، أريده يكتشف» — بس بمحاور
// **موجودة أصلاً** بالجدول (يوم، منظومة، وردية، موظف مقابل زملائه)،
// مو نموذجاً يخترع أسئلة جديدة (خطر حقن SQL وبيانات حساسة). شوف
// خطة «الاستكشاف الأسبوعي» للتفصيل الكامل.
const (
	AiAxisDayOfWeek  = "DAY_OF_WEEK"
	AiAxisSystemType = "SYSTEM_TYPE"
	AiAxisShift      = "SHIFT"
	AiAxisEmployee   = "EMPLOYEE"
)

func AiAxisLabel(axis string) string {
	switch axis {
	case AiAxisDayOfWeek:
		return "يوم الأسبوع"
	case AiAxisSystemType:
		return "نوع المنظومة"
	case AiAxisShift:
		return "الوردية"
	case AiAxisEmployee:
		return "الموظف مقابل زملائه"
	}
	return axis
}

// AiDayOfWeekLabel يترجم ISODOW (١=الاثنين..٧=الأحد) لاسم عربي.
func AiDayOfWeekLabel(isodow string) string {
	switch isodow {
	case "1":
		return "الاثنين"
	case "2":
		return "الثلاثاء"
	case "3":
		return "الأربعاء"
	case "4":
		return "الخميس"
	case "5":
		return "الجمعة"
	case "6":
		return "السبت"
	case "7":
		return "الأحد"
	}
	return isodow
}

func AiShiftLabel(shift string) string {
	switch shift {
	case "MORNING":
		return "صباحي"
	case "EVENING":
		return "مسائي"
	}
	return shift
}

// مفاتيح مؤشرات الشبكة — ٥ مقاييس تُفحص ضد الـ٤ محاور فوق.
const (
	AiDiscoveryLateStartRate     = "DISC_LATE_START_RATE"
	AiDiscoveryWorkStopRate      = "DISC_WORK_STOP_RATE"
	AiDiscoveryPartialRate       = "DISC_PARTIAL_RATE"
	AiDiscoveryInvoiceAdjustRate = "DISC_INVOICE_ADJUST_RATE"
	AiDiscoveryDisciplineRate    = "DISC_DISCIPLINE_RATE"
)

func AiDiscoveryMetricLabel(key string) string {
	switch key {
	case AiDiscoveryLateStartRate:
		return "نسبة التأخر بالخروج"
	case AiDiscoveryWorkStopRate:
		return "نسبة توقف العمل"
	case AiDiscoveryPartialRate:
		return "نسبة الإنجاز الجزئي"
	case AiDiscoveryInvoiceAdjustRate:
		return "نسبة تعديل الفواتير"
	case AiDiscoveryDisciplineRate:
		return "نسبة المخالفات الانضباطية"
	}
	return key
}

// AiWorkWindow ساعات الدوام — مصدر واحد بدل ما تنبعثر بالكود.
type AiWorkWindow struct {
	ID        string    `db:"id" json:"id"`
	StartHour int       `db:"startHour" json:"startHour"`
	EndHour   int       `db:"endHour" json:"endHour"`
	UpdatedAt time.Time `db:"updatedAt" json:"updatedAt"`
}
