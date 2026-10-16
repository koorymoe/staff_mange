package model

import "time"

// ═══ صندوق المراقب ═══
//
// صف واحد = «شي صار بالنظام ولازم عين المراقب عليه».
type MonitorReview struct {
	ID              string     `db:"id" json:"id"`
	Stage           string     `db:"stage" json:"stage"`
	EntityType      string     `db:"entityType" json:"entityType"`
	EntityID        string     `db:"entityId" json:"entityId"`
	Title           string     `db:"title" json:"title"`
	Summary         *string    `db:"summary" json:"summary"`
	OwnerRole       *string    `db:"ownerRole" json:"ownerRole"`
	OwnerEmployeeID *string    `db:"ownerEmployeeId" json:"ownerEmployeeId"`
	Status          string     `db:"status" json:"status"`
	Note            *string    `db:"note" json:"note"`
	ReviewedByID    *string    `db:"reviewedById" json:"reviewedById"`
	ReviewedAt      *time.Time `db:"reviewedAt" json:"reviewedAt"`
	CreatedAt       time.Time  `db:"createdAt" json:"createdAt"`

	OwnerEmployee *EmployeeBrief `db:"-" json:"ownerEmployee,omitempty"`
	ReviewedBy    *EmployeeBrief `db:"-" json:"reviewedBy,omitempty"`

	// هوية الشي الي ينتراجع. صندوق المراقب كان يعرض عنوان نصّي بس —
	// «فاتورة الليدر» بلا كود حجز ولا زبون — فالمراقب لازم يفتح كل صف
	// حتى يعرف عن منو يحچي. طلب صاحب العمل: «مامربوط شي بشي لازم ينربط».
	Identity *MonitorIdentity `db:"-" json:"identity,omitempty"`
	// Urgent صف عاجل — يطلع بأول الصندوق بشارة حمرا.
	Urgent bool `db:"urgent" json:"urgent"`
	// Late تفاصيل «تأخر بالخروج للزبون» — منو، والموعد، ووكت الطلعة، وشكد، وليش.
	Late *LateDetail `db:"-" json:"late,omitempty"`
	// (ع 10-16) «تجيني الفاتورة بكل تفاصيلها… الحجز وتفاصيله» — يتعبّى بالخادم
	Details *MonitorDetails `db:"-" json:"details,omitempty"`
}

// LateDetail لأحكام ماتركس عن التأخر بالخروج — (ع): «أريد يظهر الحجز واسم
// الليدر، وزر يحوّلني عليه، وشكد تأخر، وشوكت الموعد، وشنو سبب التأخير».
type LateDetail struct {
	VerdictID    string     `db:"verdictId" json:"-"`
	EmployeeID   *string    `db:"employeeId" json:"employeeId"`
	EmployeeName *string    `db:"employeeName" json:"employeeName"`
	BookingID    *string    `db:"bookingId" json:"bookingId"`
	BookingCode  *string    `db:"bookingCode" json:"bookingCode"`
	ScheduledAt  *time.Time `db:"scheduledAt" json:"scheduledAt"`
	DepartedAt   *time.Time `db:"departedAt" json:"departedAt"`
	MinutesLate  *int       `db:"minutesLate" json:"minutesLate"`
	Threshold    *int       `db:"threshold" json:"threshold"`
	Count30d     *int       `db:"count30d" json:"count30d"`
	Reason       *string    `db:"reason" json:"reason"`
	SignalID     string     `db:"signalId" json:"-"`
	// «صارت قبل؟» — مرات تأخر نفس الموظف السابقة (الأحدث أول).
	History []LateHistoryItem `db:"-" json:"history"`
}

// LateHistoryItem مرة تأخر سابقة لنفس الموظف، ويا حكم المراقب وقتها.
type LateHistoryItem struct {
	EmployeeID  string    `db:"employeeId" json:"-"`
	SignalID    string    `db:"signalId" json:"-"`
	OccurredAt  time.Time `db:"occurredAt" json:"occurredAt"`
	BookingID   *string   `db:"bookingId" json:"bookingId"`
	BookingCode *string   `db:"bookingCode" json:"bookingCode"`
	MinutesLate *int      `db:"minutesLate" json:"minutesLate"`
	Reason      *string   `db:"reason" json:"reason"`
	Review      *string   `db:"review" json:"review"` // PENDING | OK | FLAGGED
	ReviewNote  *string   `db:"reviewNote" json:"reviewNote"`
}

// MonitorIdentity هوية الحجز وراء صف الصندوق — نفس حقول رأس الهوية
// بالواجهة حتى ما تختلف التسمية بين المكانين.
//
// ⚠️ الحقل الفارغ يبقى فارغ ولا ينتخمّن: صف بلا حجز (جهاز جي بي اس
// مثلاً) يرجّع بلا هوية، ما نلزق بيه أول حجز نلگاه.
type MonitorIdentity struct {
	BookingID     string  `db:"bookingId" json:"bookingId"`
	BookingCode   string  `db:"bookingCode" json:"bookingCode"`
	CustomerCode  int     `db:"customerCode" json:"customerCode"`
	CustomerName  string  `db:"customerName" json:"customerName"`
	CustomerPhone *string `db:"customerPhone" json:"customerPhone"`
	Address       *string `db:"address" json:"address"`
	LeaderName    *string `db:"leaderName" json:"leaderName"`

	// رقم الفاتورة المحاسبية الي كتبه المحاسب من النظام الخارجي.
	// المراقب يدقّق **ورا** المحاسب، فلازم يشوف الرقم الي ثبّته —
	// بدونه ما يكدر يطابق فاتورتنا بفاتورة النظام الخارجي.
	// ⚠️ ينتعبى لصفوف الفواتير وتعديلاتها بس، ويبقى فاضي لغيرها.
	ExternalInvoiceNumber *string `db:"-" json:"externalInvoiceNumber,omitempty"`

	// مفتاح الربط الداخلي — entityType|entityId
	Key string `db:"key" json:"-"`
}

// محطات المراقبة — بالضبط الي طلبها صاحب العمل.
const (
	MonitorStageInvoiceBeforeAudit   = "INVOICE_BEFORE_AUDIT"
	MonitorStageInvoiceAfterAudit    = "INVOICE_AFTER_AUDIT"
	MonitorStageInvoiceAdjusted      = "INVOICE_ADJUSTED"
	MonitorStageBookingBeforeConfirm = "BOOKING_BEFORE_CONFIRM"
	MonitorStageBookingAfterConfirm  = "BOOKING_AFTER_CONFIRM"
	MonitorStageBookingAfterComplete = "BOOKING_AFTER_COMPLETE"
	// شغل بقية الأقسام — نفس الفكرة: اللحظة الي بيها القرار ينفّذ
	// ويصير صعب التراجع عنه.
	MonitorStageProcurementFulfilled = "PROCUREMENT_FULFILLED"
	MonitorStageQualityVerdict       = "QUALITY_VERDICT"
	MonitorStageGpsDeviceDone        = "GPS_DEVICE_DONE"
	// حجز طاقة شمسية انتسعّر: السعر ينحسب تلقائياً من مكوّنات المنظومة
	// بأسعار المخزن، وموظف المبيعات ما يشوف الأرقام ولا يقدر يعدّلها.
	// فلو تغيّر سعر مكوّن بالمخزن أو انختارت منظومة غلط، الرقم يوصل
	// الزبون بعرض سعر وما ينكشف الخطأ إلا بعد ما يوافق. المراقب يشوفه
	// بلحظة التسعير — قبل ما ينبني عليه اتفاق.
	MonitorStageSolarQuoted = "SOLAR_QUOTED"
	// أحكام ماتركس — حكم النموذج/القواعد على إشارة تحليل. entityId
	// هو معرّف الحكم (`AiVerdict.ID`) مو الحجز، نفس منطق تعديلات
	// الفاتورة (شوف تعليق `INVOICE_ADJUSTMENT` بمستودع الهوية).
	MonitorStageAiVerdict = "AI_VERDICT"
)

const (
	MonitorStatusPending = "PENDING"
	MonitorStatusOK      = "OK"
	MonitorStatusFlagged = "FLAGGED"
)

// MonitorStageLabel التسمية العربية — تنستعمل بالإشعارات وبالواجهة سوه
// حتى ما تختلف التسمية بين المكانين.
func MonitorStageLabel(stage string) string {
	switch stage {
	case MonitorStageInvoiceBeforeAudit:
		return "فاتورة قبل التدقيق"
	case MonitorStageInvoiceAfterAudit:
		return "فاتورة بعد التدقيق"
	case MonitorStageInvoiceAdjusted:
		return "مبالغ فاتورة انتعدّلت"
	case MonitorStageBookingBeforeConfirm:
		return "حجز قبل التثبيت"
	case MonitorStageBookingAfterConfirm:
		return "حجز بعد التثبيت"
	case MonitorStageBookingAfterComplete:
		return "حجز بعد الإنجاز"
	case MonitorStageProcurementFulfilled:
		return "مادة انشترت"
	case MonitorStageQualityVerdict:
		return "حكم الجودة"
	case MonitorStageGpsDeviceDone:
		return "جهاز جي بي اس انخلص"
	case MonitorStageSolarQuoted:
		return "منظومة شمسية انتسعّرت"
	case MonitorStageAiVerdict:
		return "حكم ماتركس"
	}
	return stage
}

// EnqueueMonitorReview مدخلات إضافة صف للصندوق.
type EnqueueMonitorReview struct {
	Stage           string
	EntityType      string
	EntityID        string
	Title           string
	Summary         string
	OwnerRole       string
	OwnerEmployeeID *string
	// Urgent صف يطلع بأول الصندوق بشارة حمرا — «غير مطابق» بالفواتير.
	// ⚠️ ما ينطفي بتحديث لاحق: صف صار عاجلاً مرة يبقى عاجلاً لحد ما
	// يبتّ فيه المراقب. وإلا تحديث روتيني يخفي التنبيه.
	Urgent bool
}

// DecideMonitorReviewRequest قرار المراقب على صف.
//
// الملاحظة إجبارية بالتأشير: «عندي ملاحظة» بدون نص ما تفيد أحد،
// والموظف الي توصله ما راح يعرف شنو يصلّح.
type DecideMonitorReviewRequest struct {
	Flag bool   `json:"flag"`
	Note string `json:"note"`
}

// MonitorInboxCounts عدّادات فوق التبويبات.
type MonitorInboxCount struct {
	Stage string `db:"stage" json:"stage"`
	Count int    `db:"count" json:"count"`
}

// MonitorDetails تفاصيل الصف الكاملة بجهة وحدة — قرار (ع) 10-16.
type MonitorDetails struct {
	Invoice *MonitorInvoiceDetail `json:"invoice,omitempty"`
	Booking *MonitorBookingDetail `json:"booking,omitempty"`
}

type MonitorInvoiceDetail struct {
	InvoiceID      string                   `db:"id" json:"invoiceId"`
	LeaderName     *string                  `db:"leaderName" json:"leaderName"`
	ExecutionCost  float64                  `db:"executionCost" json:"executionCost"`
	MaterialsTotal float64                  `db:"materialsTotal" json:"materialsTotal"`
	DiscountValue  float64                  `db:"discountValue" json:"discountValue"`
	NetTotal       float64                  `db:"netTotal" json:"netTotal"`
	IsFree         bool                     `db:"isFree" json:"isFree"`
	Status         string                   `db:"status" json:"status"`
	Collected      *float64                 `db:"collected" json:"collected"`
	ExternalNo     *string                  `db:"externalNo" json:"externalNo"`
	CreatedAt      time.Time                `db:"createdAt" json:"createdAt"`
	Adjustments    []MonitorInvoiceAdjustment `db:"-" json:"adjustments"`
}

type MonitorInvoiceAdjustment struct {
	OldNetTotal       float64   `db:"oldNetTotal" json:"oldNetTotal"`
	NewNetTotal       float64   `db:"newNetTotal" json:"newNetTotal"`
	OldExecutionCost  float64   `db:"oldExecutionCost" json:"oldExecutionCost"`
	NewExecutionCost  float64   `db:"newExecutionCost" json:"newExecutionCost"`
	OldMaterialsTotal float64   `db:"oldMaterialsTotal" json:"oldMaterialsTotal"`
	NewMaterialsTotal float64   `db:"newMaterialsTotal" json:"newMaterialsTotal"`
	Reason            string    `db:"reason" json:"reason"`
	ByName            *string   `db:"byName" json:"byName"`
	CreatedAt         time.Time `db:"createdAt" json:"createdAt"`
}

type MonitorBookingDetail struct {
	Status          string     `db:"status" json:"status"`
	Station         string     `db:"station" json:"station"`
	Service         *string    `db:"service" json:"service"`
	CreatedByName   *string    `db:"createdByName" json:"createdByName"`
	CreatedAt       time.Time  `db:"createdAt" json:"createdAt"`
	ContactedAt     *time.Time `db:"contactedAt" json:"contactedAt"`
	ConfirmedAt     *time.Time `db:"confirmedAt" json:"confirmedAt"`
	ConfirmedByName *string    `db:"confirmedByName" json:"confirmedByName"`
	ScheduledAt     *time.Time `db:"scheduledAt" json:"scheduledAt"`
	Shift           *string    `db:"shift" json:"shift"`
	ArrivedAt       *time.Time `db:"arrivedAt" json:"arrivedAt"`
	StartedAt       *time.Time `db:"startedAt" json:"startedAt"`
	CompletedAt     *time.Time `db:"completedAt" json:"completedAt"`
	LeaderName      *string    `db:"leaderName" json:"leaderName"`
	Crew            *string    `db:"crew" json:"crew"`
	Collected       *float64   `db:"collected" json:"collected"`
	Quality         *string    `db:"quality" json:"quality"`
}
