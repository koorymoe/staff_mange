package model

import "time"

// طلب حذف حجز.
//
// الحجوزات التجريبية والملغاة لازم تنشال، بس الحذف ما يترد — فالإداري
// يطلب، والمراقب أو مدير النظام يوافق أو يرفض. بلا هالدورة، أي غلطة
// تروح بيانات ما ترجع.

const (
	BookingDeleteStatusPending  = "PENDING"
	BookingDeleteStatusApproved = "APPROVED"
	BookingDeleteStatusRejected = "REJECTED"
)

var BookingDeleteStatusLabels = map[string]string{
	BookingDeleteStatusPending:  "بانتظار الموافقة",
	BookingDeleteStatusApproved: "انحذف",
	BookingDeleteStatusRejected: "انرفض الطلب",
}

// القناة الي انطلب منها الحذف — حقل فعلي يُدخل وقت التسجيل، مو استخلاص من النص.
const (
	BookingDeleteChannelWebsite    = "WEBSITE"
	BookingDeleteChannelMobileApp  = "MOBILE_APP"
	BookingDeleteChannelCallCenter = "CALL_CENTER"
)

var BookingDeleteChannelLabels = map[string]string{
	BookingDeleteChannelWebsite:    "موقع ويب",
	BookingDeleteChannelMobileApp:  "تطبيق الجوال",
	BookingDeleteChannelCallCenter: "مركز الاتصال",
}

// نوع طلب الحذف — نفس المبدأ: حقل فعلي، والقيم المسموحة تتفحّص بالخادم.
const (
	BookingDeleteTypeRecurringDuplicate = "RECURRING_DUPLICATE"
	BookingDeleteTypeCustomerCancel     = "CUSTOMER_CANCEL"
	BookingDeleteTypeDataCorrection     = "DATA_CORRECTION"
)

var BookingDeleteTypeLabels = map[string]string{
	BookingDeleteTypeRecurringDuplicate: "حجز متكرر",
	BookingDeleteTypeCustomerCancel:     "إلغاء من الزبون",
	BookingDeleteTypeDataCorrection:     "تصحيح بيانات",
}

type BookingDeleteRequest struct {
	ID            string     `db:"id" json:"id"`
	BookingID     string     `db:"bookingId" json:"bookingId"`
	RequestedByID string     `db:"requestedById" json:"requestedById"`
	Reason        string     `db:"reason" json:"reason"`
	Status        string     `db:"status" json:"status"`
	DecidedByID   *string    `db:"decidedById" json:"decidedById"`
	DecidedAt     *time.Time `db:"decidedAt" json:"decidedAt"`
	DecisionNote  *string    `db:"decisionNote" json:"decisionNote"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`

	Channel     *string `db:"channel" json:"channel"`
	RequestType *string `db:"requestType" json:"requestType"`

	// «معلقة»: تصنيف فرعي داخل PENDING — الطلب ناقصه معلومات، ولسه ما انبتّ فيه.
	NeedsInfo     bool       `db:"needsInfo" json:"needsInfo"`
	NeedsInfoNote *string    `db:"needsInfoNote" json:"needsInfoNote"`
	NeedsInfoAt   *time.Time `db:"needsInfoAt" json:"needsInfoAt"`
	NeedsInfoByID *string    `db:"needsInfoById" json:"needsInfoById"`

	BookingCode     string  `db:"bookingCode" json:"bookingCode"`
	CustomerName    string  `db:"customerName" json:"customerName"`
	BookingStatus   string  `db:"bookingStatus" json:"bookingStatus"`
	RequestedByName string  `db:"requestedByName" json:"requestedByName"`
	DecidedByName   *string `db:"decidedByName" json:"decidedByName"`
	NeedsInfoByName *string `db:"needsInfoByName" json:"needsInfoByName"`

	StatusLabel      string `db:"-" json:"statusLabel"`
	ChannelLabel     string `db:"-" json:"channelLabel"`
	RequestTypeLabel string `db:"-" json:"requestTypeLabel"`
}

type CreateBookingDeleteRequest struct {
	Reason      string `json:"reason"`
	Channel     string `json:"channel"`
	RequestType string `json:"requestType"`
}

type DecideBookingDeleteRequest struct {
	Approve bool    `json:"approve"`
	Note    *string `json:"note"`
}

type NeedsInfoBookingDeleteRequest struct {
	Note string `json:"note"`
}

// BookingDeleteRequestCounts — عدّ مجمَّع واحد يخدم البطاقات الخمس
// بالشاشة، بدل حساب العدّ من القائمة المصفّاة بالواجهة.
type BookingDeleteRequestCounts struct {
	Approved       int `db:"approved" json:"approved"`
	NeedsInfo      int `db:"needsInfo" json:"needsInfo"`
	AwaitingReview int `db:"awaitingReview" json:"awaitingReview"`
	Rejected       int `db:"rejected" json:"rejected"`
	Total          int `db:"total" json:"total"`
}
