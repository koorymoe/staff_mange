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
}

type UpdateQualityFollowUpRequest struct {
	Status       string  `json:"status"`
	ContactNotes *string `json:"contactNotes"`
}
