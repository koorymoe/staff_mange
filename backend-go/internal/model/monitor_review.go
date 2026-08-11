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
	MonitorStageInvoiceBeforeAudit  = "INVOICE_BEFORE_AUDIT"
	MonitorStageInvoiceAfterAudit   = "INVOICE_AFTER_AUDIT"
	MonitorStageInvoiceAdjusted     = "INVOICE_ADJUSTED"
	MonitorStageBookingBeforeConfirm = "BOOKING_BEFORE_CONFIRM"
	MonitorStageBookingAfterConfirm  = "BOOKING_AFTER_CONFIRM"
	MonitorStageBookingAfterComplete = "BOOKING_AFTER_COMPLETE"
	// شغل بقية الأقسام — نفس الفكرة: اللحظة الي بيها القرار ينفّذ
	// ويصير صعب التراجع عنه.
	MonitorStageProcurementFulfilled = "PROCUREMENT_FULFILLED"
	MonitorStageQualityVerdict       = "QUALITY_VERDICT"
	MonitorStageGpsDeviceDone        = "GPS_DEVICE_DONE"
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
