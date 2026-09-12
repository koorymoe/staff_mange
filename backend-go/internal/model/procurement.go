package model

import "time"

// أنواع طلبات المشتريات — كل نوع له صلاحية مستقلة يمنحها الأدمن يدوياً للموظف
// (procurement_personal / procurement_customer)، ما تنجر تلقائياً مع أي دور.
const (
	RequestTypePersonalSupply  = "PERSONAL_SUPPLY"  // طلب احتياجات شخصية للموظف نفسه
	RequestTypeCustomerProduct = "CUSTOMER_PRODUCT" // طلب منتج للزبون (الفلو الأصلي، مربوط بحجز)
	// طلب مادة يدوي: الإداري أو الليدر يكتب المادة بالتفصيل — موديل،
	// نوعية، عدد، سعر تقريبي — بدون ما يكون مربوط بحجز ولا بكتالوك.
	RequestTypeManualSupply = "MANUAL_SUPPLY"
)

var RequestTypeLabels = map[string]string{
	RequestTypePersonalSupply:  "احتياجات شخصية",
	RequestTypeCustomerProduct: "منتج للزبون",
	RequestTypeManualSupply:    "طلب مادة يدوي",
}

type ProcurementRequest struct {
	ID               string   `db:"id" json:"id"`
	Code             string   `db:"code" json:"code"`
	RequestedByID    string   `db:"requestedById" json:"-"`
	BookingID        *string  `db:"bookingId" json:"-"`
	RequestType      string   `db:"requestType" json:"requestType"`
	Notes            *string  `db:"notes" json:"notes"`
	Status           string   `db:"status" json:"status"`
	FulfilledByID    *string  `db:"fulfilledById" json:"-"`
	TotalCost        *float64 `db:"totalCost" json:"totalCost"`
	FulfillmentNotes *string  `db:"fulfillmentNotes" json:"fulfillmentNotes"`
	// المورد الي انجابت منه المادة — يتحدد وقت التجهيز، إلزامي حتى
	// نعرف من وين انجابت ونكدر نحاسب على السعر.
	SupplierID  *string    `db:"supplierId" json:"supplierId"`
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
	FulfilledAt *time.Time `db:"fulfilledAt" json:"fulfilledAt"`

	Items       []ProcurementItem          `db:"-" json:"items"`
	RequestedBy *ProcurementRequesterBrief `db:"-" json:"requestedBy"`
	FulfilledBy *EmployeeIDNameBrief       `db:"-" json:"fulfilledBy"`
	Booking     *ProcurementBookingBrief   `db:"-" json:"booking"`
	Supplier    *EmployeeIDNameBrief       `db:"-" json:"supplier"`
	TypeLabel   string                     `db:"-" json:"typeLabel"`
}

// ProcurementRequesterBrief يماثل { select: { id: true, name: true, role: true } } بالباك إند القديم
type ProcurementRequesterBrief struct {
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Role string `db:"role" json:"role"`
}

// EmployeeIDNameBrief يماثل { select: { id: true, name: true } } بالباك إند القديم
type EmployeeIDNameBrief struct {
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// ProcurementBookingBrief يماثل include: { booking: { include: { customer: { select ... } } } } بالباك إند القديم
type ProcurementBookingBrief struct {
	ID string `db:"id" json:"id"`
	// كود الحجز — إداري الكميات يحتاجه حتى يعرف أي حجز يخص الطلب.
	// كانت الواجهة تعرضه وهو مو موجود بالرد أصلاً، فيطلع «undefined».
	Code     string         `db:"code" json:"code"`
	Customer *CustomerBrief `db:"-" json:"customer"`
}

type CustomerBrief struct {
	ID    string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Phone string `db:"phone" json:"phone"`
}

type ProcurementItem struct {
	ID          string   `db:"id" json:"id"`
	RequestID   string   `db:"requestId" json:"requestId"`
	ProductName string   `db:"productName" json:"productName"`
	Quantity    int      `db:"quantity" json:"quantity"`
	UnitPrice   *float64 `db:"unitPrice" json:"unitPrice"`
	TotalPrice  *float64 `db:"totalPrice" json:"totalPrice"`
	Fulfilled   bool     `db:"fulfilled" json:"fulfilled"`
	// تفاصيل الطلب اليدوي — الموديل والنوعية وملاحظات المادة
	Model     *string `db:"model" json:"model"`
	Spec      *string `db:"spec" json:"spec"`
	ItemNotes *string `db:"itemNotes" json:"itemNotes"`
}

type CreateProcurementItemRequest struct {
	ProductName string   `json:"productName"`
	Quantity    int      `json:"quantity"`
	UnitPrice   *float64 `json:"unitPrice"`
	TotalPrice  *float64 `json:"totalPrice"`
	Model       *string  `json:"model"`
	Spec        *string  `json:"spec"`
	ItemNotes   *string  `json:"itemNotes"`
}

type CreateProcurementRequestRequest struct {
	RequestedByID string                         `json:"requestedById"`
	BookingID     *string                        `json:"bookingId"`
	RequestType   string                         `json:"requestType"`
	Notes         *string                        `json:"notes"`
	Items         []CreateProcurementItemRequest `json:"items"`
}

type UpdateProcurementStatusRequest struct {
	Status string `json:"status"`
}

type FulfillProcurementItemRequest struct {
	ID         string   `json:"id"`
	UnitPrice  *float64 `json:"unitPrice"`
	TotalPrice *float64 `json:"totalPrice"`
	Fulfilled  *bool    `json:"fulfilled"`
}

type FulfillProcurementRequestRequest struct {
	FulfilledByID    string                          `json:"fulfilledById"`
	SupplierID       *string                         `json:"supplierId"`
	TotalCost        *float64                        `json:"totalCost"`
	FulfillmentNotes *string                         `json:"fulfillmentNotes"`
	Items            []FulfillProcurementItemRequest `json:"items"`
}

type ProcurementStats struct {
	TotalSpent     float64            `json:"totalSpent"`
	TotalItems     int                `json:"totalItems"`
	PendingCount   int                `json:"pendingCount"`
	MonthlySpent   float64            `json:"monthlySpent"`
	FulfilledCount int                `json:"fulfilledCount"`
	ByMonth        map[string]float64 `json:"byMonth"`
}
