package model

import "time"

// الدوار: مبلغ دوّار للعمل يصرفه المحاسب للموظفين، ويرجع لما يتسوّى.
//
// الفكرة: الموظف ياخذ مبلغ، يشتري بيه، يرفع صورة الوصل، ويرجّع الباقي.
// المبلغ يبقى دَين برقبته لحد ما المحاسب يدقّق ويوافق — الموافقة بس هي
// الي تصفّر رصيده.

type RevolvingFund struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Balance   float64   `db:"balance" json:"balance"`
	IsActive  bool      `db:"isActive" json:"isActive"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `db:"updatedAt" json:"updatedAt"`

	// OutstandingTotal مجموع الي بيد الموظفين من هذا الدوار وما انتسوّى بعد
	OutstandingTotal float64 `db:"-" json:"outstandingTotal"`
	// AwaitingDischargeTotal المبلغ الي الدوار ناطره من المحاسب: مصروف
	// مدقّق وما انخرّج بعد. هذا الي ينقص من رأس مال الدوار فعلياً.
	AwaitingDischargeTotal float64 `db:"-" json:"awaitingDischargeTotal"`
}

// أنواع حركة الدوار
const (
	FundTxnDisburse   = "DISBURSE"   // المحاسب يسلّم موظف
	FundTxnSettlement = "SETTLEMENT" // الموظف يسوّي (صرف + مرتجع + وصل)
	FundTxnTopup      = "TOPUP"      // تغذية الدوار نفسه
	FundTxnDischarge  = "DISCHARGE"  // تخريج المادة → تعويض المبلغ المصروف
)

var FundTxnKindLabels = map[string]string{
	FundTxnDisburse:   "تسليم للموظف",
	FundTxnSettlement: "تسوية من الموظف",
	FundTxnTopup:      "تغذية الدوار",
	FundTxnDischarge:  "تخريج مادة — تعويض الدوار",
}

// حالات التسوية
const (
	FundTxnPending  = "PENDING"  // بانتظار تدقيق المحاسب
	FundTxnApproved = "APPROVED" // انوافق عليها — وقتها بس تأثر على الأرصدة
	FundTxnRejected = "REJECTED" // انرفضت — الرصيد يبقى برقبة الموظف
)

var FundTxnStatusLabels = map[string]string{
	FundTxnPending:  "بانتظار التدقيق",
	FundTxnApproved: "مدقّقة ومقبولة",
	FundTxnRejected: "مرفوضة",
}

type RevolvingFundTxn struct {
	ID             string     `db:"id" json:"id"`
	FundID         string     `db:"fundId" json:"fundId"`
	EmployeeID     *string    `db:"employeeId" json:"employeeId"`
	Kind           string     `db:"kind" json:"kind"`
	Amount         float64    `db:"amount" json:"amount"`
	SpentAmount    float64    `db:"spentAmount" json:"spentAmount"`
	ReturnedAmount float64    `db:"returnedAmount" json:"returnedAmount"`
	BookingID      *string    `db:"bookingId" json:"bookingId"`
	ReceiptImage   *string    `db:"receiptImage" json:"receiptImage"`
	Notes          *string    `db:"notes" json:"notes"`
	Status         string     `db:"status" json:"status"`
	CreatedByID    *string    `db:"createdById" json:"createdById"`
	ReviewedByID   *string    `db:"reviewedById" json:"reviewedById"`
	ReviewedAt     *time.Time `db:"reviewedAt" json:"reviewedAt"`
	ReviewNote     *string    `db:"reviewNote" json:"reviewNote"`
	CreatedAt      time.Time  `db:"createdAt" json:"createdAt"`

	// التخريج — الخطوة الي ترجّع المبلغ المصروف للدوار.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا وإلا ينكسر SELECT *.
	RequestID        *string    `db:"requestId" json:"requestId"`
	DischargedAt     *time.Time `db:"dischargedAt" json:"dischargedAt"`
	DischargedByID   *string    `db:"dischargedById" json:"dischargedById"`
	DischargeAccount *string    `db:"dischargeAccount" json:"dischargeAccount"`
	DischargeNote    *string    `db:"dischargeNote" json:"dischargeNote"`

	FundName     string  `db:"fundName" json:"fundName"`
	EmployeeName *string `db:"employeeName" json:"employeeName"`
	ReviewedBy   *string `db:"reviewedByName" json:"reviewedByName"`
	DischargedBy *string `db:"dischargedByName" json:"dischargedByName"`
	KindLabel    string  `db:"-" json:"kindLabel"`
	StatusLabel  string  `db:"-" json:"statusLabel"`
	// AwaitingDischarge = تسوية مقبولة، بيها مبلغ مصروف، وما انخرجت بعد.
	// يعني الدوار ناطر هذا المبلغ من المحاسب.
	AwaitingDischarge bool `db:"-" json:"awaitingDischarge"`
}

// DischargeRequest المحاسب يأشر إن المادة انخرجت وعلى أي حساب انحسبت.
type DischargeRequest struct {
	Account string  `json:"account"`
	Note    *string `json:"note"`
}

// EmployeeFundBalance رصيد موظف واحد: شكد بيده وما انتسوّى.
type EmployeeFundBalance struct {
	EmployeeID   string  `db:"employeeId" json:"employeeId"`
	EmployeeName string  `db:"employeeName" json:"employeeName"`
	JobTitle     *string `db:"jobTitle" json:"jobTitle"`
	// TotalTaken كل الي استلمه، Settled الي انتسوّى وانوافق عليه
	TotalTaken   float64 `db:"totalTaken" json:"totalTaken"`
	TotalSettled float64 `db:"totalSettled" json:"totalSettled"`
	Outstanding  float64 `db:"outstanding" json:"outstanding"`
	// PendingSettlements عدد التسويات المرفوعة وناطرة تدقيق المحاسب
	PendingSettlements int `db:"pendingSettlements" json:"pendingSettlements"`
}

type UpsertFundRequest struct {
	Name     *string  `json:"name"`
	Balance  *float64 `json:"balance"`
	IsActive *bool    `json:"isActive"`
}

type DisburseRequest struct {
	FundID     string  `json:"fundId"`
	EmployeeID string  `json:"employeeId"`
	Amount     float64 `json:"amount"`
	BookingID  *string `json:"bookingId"`
	// RequestID طلب المشتريات الي انصرف عليه — اختياري، بس لمن ينوجد
	// يبين اسم الموظف الطالب والمادة كدام المحاسب.
	RequestID *string `json:"requestId"`
	Notes     *string `json:"notes"`
}

type SettlementRequest struct {
	FundID         string  `json:"fundId"`
	SpentAmount    float64 `json:"spentAmount"`
	ReturnedAmount float64 `json:"returnedAmount"`
	ReceiptImage   *string `json:"receiptImage"`
	BookingID      *string `json:"bookingId"`
	Notes          *string `json:"notes"`
}

type ReviewSettlementRequest struct {
	Approve    bool    `json:"approve"`
	ReviewNote *string `json:"reviewNote"`
}
