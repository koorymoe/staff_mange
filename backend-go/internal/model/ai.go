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

	EmployeeName *string    `db:"-" json:"employeeName,omitempty"`
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
)

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
}

// مفاتيح المؤشرات — تنحسب من الأدلة مو عدّادات خام.
const (
	AiMetricStopRate          = "STOP_RATE"           // نسبة الحجوزات الي وقّفت
	AiMetricStopMinutesAvg    = "STOP_MINUTES_AVG"    // متوسط الوقت الضايع بالتوقف
	AiMetricMaterialMissRate  = "MATERIAL_MISS_RATE"  // توقف بسبب مادة ما انطلبت
	AiMetricScopeCreepRate    = "SCOPE_CREEP_RATE"    // الزبون طلب زيادة بالموقع
	AiMetricProcurementDelay  = "PROCUREMENT_DELAY"   // تأخر إداري الكميات
	AiMetricLateStartRate     = "LATE_START_RATE"     // نسبة التأخر بالخروج
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
