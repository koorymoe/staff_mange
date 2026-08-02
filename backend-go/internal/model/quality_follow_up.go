package model

import "time"

type QualityFollowUp struct {
	ID                    string     `db:"id" json:"id"`
	BookingID             string     `db:"bookingId" json:"-"`
	CustomerID            string     `db:"customerId" json:"-"`
	Status                string     `db:"status" json:"status"` // PENDING | CONTACTED_OK | CONTACTED_ISSUE | CONVERTED | CLOSED
	ContactNotes          *string    `db:"contactNotes" json:"contactNotes"`
	ContactedByEmployeeID *string    `db:"contactedByEmployeeId" json:"-"`
	ContactedAt           *time.Time `db:"contactedAt" json:"contactedAt"`
	CreatedAt             time.Time  `db:"createdAt" json:"createdAt"`

	Booking             *Booking       `db:"-" json:"booking"`
	Customer            *Customer      `db:"-" json:"customer"`
	ContactedByEmployee *EmployeeBrief `db:"-" json:"contactedByEmployee"`

	// تفاصيل المشروع والمبالغ — مهندس الجودة يحتاجها وهو يتصل بالزبون حتى
	// يعرف شنو انتفق عليه وشكد انستلم فعلاً، ويقدر يكتب تفاصيل الفارق.
	Financials *QualityFollowUpFinancials `db:"-" json:"financials"`
}

// QualityFollowUpFinancials صورة كاملة عن مشروع/حجز المتابعة وأمواله.
type QualityFollowUpFinancials struct {
	BookingCode  string  `db:"bookingCode" json:"bookingCode"`
	ServiceName  *string `db:"serviceName" json:"serviceName"`
	Location     *string `db:"location" json:"location"`
	WorkDetails  *string `db:"workDetails" json:"workDetails"` // ملاحظات الحجز = تفاصيل العمل
	ProjectID    *string `db:"projectId" json:"projectId"`
	ProjectCode  *string `db:"projectCode" json:"projectCode"`
	ProjectName  *string `db:"projectName" json:"projectName"`
	ProjectStage *string `db:"projectStage" json:"projectStage"`

	QuotedPrice     *float64 `db:"quotedPrice" json:"quotedPrice"`         // السعر المتفق عليه
	ProjectPrice    *float64 `db:"projectPrice" json:"projectPrice"`       // سعر المشروع لو مرحّل
	AdvancePaid     *float64 `db:"advancePaid" json:"advancePaid"`         // العربون
	AmountCollected *float64 `db:"amountCollected" json:"amountCollected"` // المستلم فعلاً

	// Difference = المتفق عليه − (العربون + المستلم). موجب يعني باقي بذمة
	// الزبون، سالب يعني انستلم أكثر من المتفق.
	AgreedTotal   float64 `db:"-" json:"agreedTotal"`
	ReceivedTotal float64 `db:"-" json:"receivedTotal"`
	Difference    float64 `db:"-" json:"difference"`
}

type UpdateQualityFollowUpRequest struct {
	Status       string  `json:"status"`
	ContactNotes *string `json:"contactNotes"`
}
