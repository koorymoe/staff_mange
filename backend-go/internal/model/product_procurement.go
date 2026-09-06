package model

import "time"

// ProductProcurement تجهيز طلب منتج من الدوار.
//
// المسار كامل: التقني يطلب منتج → أبو الحسابات (نفسه أبو الكميات) يجهّزه:
// يحدد المورد من الموردين المضافين مسبقاً، يأخذ جزء من الدوار ويشتري،
// ويسجّل المبلغ المصروف والسبب وصورة الوصل.
//
// النقطة الي تخلي هذا نظام مو سجل: الطلب يبقى PENDING — يعني المبلغ
// ناقص من الدوار — لحد ما المحاسب يرجّع الفلوس فعلاً. ومنو يرجّعهن
// يعتمد على payerKind: للشركة يعني الشركة تعوّض، للزبون يعني الزبون.
const (
	ProcurementPending = "PENDING" // المبلغ لسّه ناقص من الدوار
	ProcurementSettled = "SETTLED" // انرجع المبلغ للدوار
)

var ProcurementStatusLabels = map[string]string{
	ProcurementPending: "معلّق — المبلغ ما رجع للدوار",
	ProcurementSettled: "انتسوّى — المبلغ رجع للدوار",
}

// منو يعوّض الدوار
const (
	PayerCompany  = "COMPANY"
	PayerCustomer = "CUSTOMER"
)

var PayerKindLabels = map[string]string{
	PayerCompany:  "للشركة — الشركة تعوّض الدوار",
	PayerCustomer: "للزبون — الزبون يعوّض الدوار",
}

func ValidPayerKind(v string) bool {
	_, ok := PayerKindLabels[v]
	return ok
}

type ProductProcurement struct {
	ID            string     `db:"id" json:"id"`
	RequestID     string     `db:"requestId" json:"requestId"`
	FundID        string     `db:"fundId" json:"fundId"`
	SupplierID    string     `db:"supplierId" json:"supplierId"`
	ProductName   string     `db:"productName" json:"productName"`
	SpentAmount   float64    `db:"spentAmount" json:"spentAmount"`
	Reason        string     `db:"reason" json:"reason"`
	ReceiptImage  *string    `db:"receiptImage" json:"receiptImage"`
	PayerKind     string     `db:"payerKind" json:"payerKind"`
	CustomerNote  *string    `db:"customerNote" json:"customerNote"`
	BookingID     *string    `db:"bookingId" json:"bookingId"`
	Status        string     `db:"status" json:"status"`
	PurchasedByID string     `db:"purchasedById" json:"purchasedById"`
	SettledByID   *string    `db:"settledById" json:"settledById"`
	SettledAt     *time.Time `db:"settledAt" json:"settledAt"`
	SettleNote    *string    `db:"settleNote" json:"settleNote"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`

	// محسوبة بالاستعلام — الواجهة ما تسوّي طلب ثاني لكل صف
	FundName        string  `db:"fundName" json:"fundName"`
	SupplierName    string  `db:"supplierName" json:"supplierName"`
	PurchasedByName string  `db:"purchasedByName" json:"purchasedByName"`
	SettledByName   *string `db:"settledByName" json:"settledByName"`
	StatusLabel     string  `db:"-" json:"statusLabel"`
	PayerLabel      string  `db:"-" json:"payerLabel"`
}

// FulfillProductRequest شنو يعبّيه أبو الحسابات وقت ما يجهّز الطلب.
type FulfillProductRequest struct {
	FundID       string  `json:"fundId"`
	SupplierID   string  `json:"supplierId"`
	SpentAmount  float64 `json:"spentAmount"`
	Reason       string  `json:"reason"`
	ReceiptImage *string `json:"receiptImage"`
	PayerKind    string  `json:"payerKind"`
	CustomerNote *string `json:"customerNote"`
	BookingID    *string `json:"bookingId"`
}

type SettleProcurementRequest struct {
	Note *string `json:"note"`
}
