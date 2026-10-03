package model

import "time"

type QualityIssue struct {
	ID                    string     `db:"id" json:"id"`
	Category              string     `db:"category" json:"category"` // EXECUTION | OVERSIGHT
	Title                 string     `db:"title" json:"title"`
	Description           *string    `db:"description" json:"description"`
	ResponsibleEmployeeID *string    `db:"responsibleEmployeeId" json:"-"`
	ReportedByID          *string    `db:"reportedById" json:"-"`
	BookingID             *string    `db:"bookingId" json:"bookingId"`
	Status                string     `db:"status" json:"status"` // OPEN | IN_PROGRESS | RESOLVED
	CreatedAt             time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt            *time.Time `db:"resolvedAt" json:"resolvedAt"`

	ResponsibleEmployee *EmployeeBrief `db:"-" json:"responsibleEmployee"`
	ReportedBy          *EmployeeBrief `db:"-" json:"reportedBy"`
}

type CreateQualityIssueRequest struct {
	Category              string  `json:"category"`
	Title                 string  `json:"title"`
	Description           *string `json:"description"`
	ResponsibleEmployeeID *string `json:"responsibleEmployeeId"`
	BookingID             *string `json:"bookingId"`
}

type UpdateQualityIssueRequest struct {
	Status *string `json:"status"`
}
