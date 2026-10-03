package model

import "time"

// DuplicateCandidate زوج (حجزين أو زبونين) اكتشفهم فحص التكرار
// الدوري — يكتشف بالكود، يعرض للمراجعة، بلا حذف أو دمج تلقائي.
type DuplicateCandidate struct {
	ID           string     `db:"id" json:"id"`
	Kind         string     `db:"kind" json:"kind"` // BOOKING | CUSTOMER
	EntityAID    string     `db:"entityAId" json:"-"`
	EntityBID    string     `db:"entityBId" json:"-"`
	MatchReason  string     `db:"matchReason" json:"matchReason"`
	Status       string     `db:"status" json:"status"` // PENDING | DISMISSED
	ReviewedByID *string    `db:"reviewedById" json:"-"`
	ReviewedAt   *time.Time `db:"reviewedAt" json:"reviewedAt,omitempty"`
	DetectedAt   time.Time  `db:"detectedAt" json:"detectedAt"`

	ReviewedByName *string                 `db:"-" json:"reviewedByName,omitempty"`
	BookingA       *DuplicateBookingBrief  `db:"-" json:"bookingA,omitempty"`
	BookingB       *DuplicateBookingBrief  `db:"-" json:"bookingB,omitempty"`
	CustomerA      *DuplicateCustomerBrief `db:"-" json:"customerA,omitempty"`
	CustomerB      *DuplicateCustomerBrief `db:"-" json:"customerB,omitempty"`
}

const (
	DuplicateKindBooking  = "BOOKING"
	DuplicateKindCustomer = "CUSTOMER"

	DuplicateStatusPending   = "PENDING"
	DuplicateStatusDismissed = "DISMISSED"
)

// DuplicateBookingBrief هوية حجز كافية لعرضه بجانب شقيقه المشتبه به
// (نفس حقول EntityIdentity بالواجهة).
type DuplicateBookingBrief struct {
	ID            string     `db:"id" json:"id"`
	Code          string     `db:"code" json:"code"`
	CustomerName  string     `db:"customerName" json:"customerName"`
	CustomerPhone string     `db:"customerPhone" json:"customerPhone"`
	CustomerCode  int        `db:"customerCode" json:"customerCode"`
	Address       *string    `db:"address" json:"address"`
	ServiceName   *string    `db:"serviceName" json:"serviceName"`
	ScheduledAt   *time.Time `db:"scheduledAt" json:"scheduledAt"`
}

// DuplicateCustomerBrief هوية زبون كافية لعرضه بجانب شقيقه المشتبه به.
type DuplicateCustomerBrief struct {
	ID           string `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	Phone        string `db:"phone" json:"phone"`
	CustomerCode int    `db:"customerCode" json:"customerCode"`
}
