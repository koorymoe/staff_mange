package model

import "time"

// تدقيق المحاسب على مبالغ الحجوزات.
//
// قبل هذا، «تأكيد التدقيق» كان مجرد علم بوليان ينضغط حتى لو المبلغ
// فاضي — يعني حجز بلا مبلغ يطلع «مدقق» وتطلع الأرباح غلط. صار
// المحاسب لازم يكتب المبلغ، أو يأشر إن أكو خطأ ويوجّهه للمعني.

const (
	// AuditVerify المبلغ مطابق للفاتورة — يأشر الحجز مدقق.
	AuditVerify = "VERIFY"
	// AuditMismatch المبلغ مو مطابق للفاتورة الي بيد المحاسب →
	// يروح للرقابة والجودة حتى يتابعون وين راح الفرق.
	AuditMismatch = "MISMATCH"
	// AuditPriceError السعر نفسه مكتوب غلط (الإداري قدّره غلط أو
	// الليدر كتبه غلط) → يروح للرقابة والإداري حتى يصلحون التسعير.
	AuditPriceError = "PRICE_ERROR"
)

var AuditIssueLabels = map[string]string{
	AuditMismatch:   "المبلغ غير مطابق",
	AuditPriceError: "خطأ بالسعر",
}

func ValidAuditIssue(v string) bool {
	_, ok := AuditIssueLabels[v]
	return ok
}

const (
	AuditIssueOpen     = "OPEN"
	AuditIssueResolved = "RESOLVED"
)

// BookingAuditIssue بلاغ خطأ سجّله المحاسب على حجز.
type BookingAuditIssue struct {
	ID         string     `db:"id" json:"id"`
	BookingID  string     `db:"bookingId" json:"bookingId"`
	Kind       string     `db:"kind" json:"kind"`
	Note       *string    `db:"note" json:"note"`
	ExpectedAt *float64   `db:"expectedAmount" json:"expectedAmount"`
	ActualAt   *float64   `db:"actualAmount" json:"actualAmount"`
	RaisedByID string     `db:"raisedById" json:"raisedById"`
	Status     string     `db:"status" json:"status"`
	ResolvedAt *time.Time `db:"resolvedAt" json:"resolvedAt"`
	CreatedAt  time.Time  `db:"createdAt" json:"createdAt"`

	BookingCode  string  `db:"bookingCode" json:"bookingCode"`
	CustomerName string  `db:"customerName" json:"customerName"`
	RaisedByName string  `db:"raisedByName" json:"raisedByName"`
	KindLabel    string  `db:"-" json:"kindLabel"`
	RoutedTo     string  `db:"-" json:"routedTo"`
}

// AuditBookingRequest قرار المحاسب على حجز واحد.
type AuditBookingRequest struct {
	Action string `json:"action"` // VERIFY | MISMATCH | PRICE_ERROR
	// المحاسب يقدر يصحّح المبلغ وهو يدقق — لازم للحجوزات القديمة
	// المستوردة بقيم صفر: يفتح فاتورة النظام القديم ويكتب سعرها.
	AmountCollected *float64 `json:"amountCollected"`
	AdvancePaid     *float64 `json:"advancePaid"`
	Note            *string  `json:"note"`
}

// AuditRoutedRoles وين يروح كل نوع خطأ.
//
// مو مجرد إشعار: الفرق بالمبلغ شغل رقابة وجودة (وين راح الفلوس
// وهل الزبون انخدع)، وخطأ السعر شغل رقابة وإداري (منو سعّر غلط).
func AuditRoutedRoles(kind string) []string {
	switch kind {
	case AuditMismatch:
		return []string{"MONITOR", "QUALITY_ENGINEER"}
	case AuditPriceError:
		return []string{"MONITOR", "HR_COORDINATOR", "ADMIN"}
	default:
		return nil
	}
}

func AuditRoutedLabel(kind string) string {
	switch kind {
	case AuditMismatch:
		return "الرقابة والجودة"
	case AuditPriceError:
		return "الرقابة والإداري"
	default:
		return ""
	}
}
