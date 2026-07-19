package model

import "time"

type Complaint struct {
	ID                   string     `db:"id" json:"id"`
	CustomerID           string     `db:"customerId" json:"-"`
	BookingID            *string    `db:"bookingId" json:"-"`
	Description          string     `db:"description" json:"description"`
	Status               string     `db:"status" json:"status"`
	CreatedByEmployeeID  string     `db:"createdByEmployeeId" json:"-"`
	AssignedToEmployeeID *string    `db:"assignedToEmployeeId" json:"-"`
	Resolution           *string    `db:"resolution" json:"resolution"`
	CreatedAt            time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt           *time.Time `db:"resolvedAt" json:"resolvedAt"`

	Customer           *Customer      `db:"-" json:"customer"`
	Booking            *Booking       `db:"-" json:"booking"`
	CreatedByEmployee  *EmployeeBrief `db:"-" json:"createdByEmployee"`
	AssignedToEmployee *EmployeeBrief `db:"-" json:"assignedToEmployee"`
}

type CreateComplaintRequest struct {
	CustomerID          string  `json:"customerId"`
	BookingID           *string `json:"bookingId"`
	Description         string  `json:"description"`
	CreatedByEmployeeID string  `json:"createdByEmployeeId"`
}

type UpdateComplaintRequest struct {
	Status               *string `json:"status"`
	AssignedToEmployeeID *string `json:"assignedToEmployeeId"`
	Resolution           *string `json:"resolution"`
}

type ResolveComplaintRequest struct {
	Resolution *string `json:"resolution"`
}

// ComplaintCustomerStat يلخص كم مرة اشتكى زبون معيّن — تقرير منفصل عن إحصائيات الحجوزات.
type ComplaintCustomerStat struct {
	CustomerID     string `db:"customerId" json:"customerId"`
	CustomerName   string `db:"customerName" json:"customerName"`
	CustomerPhone  string `db:"customerPhone" json:"customerPhone"`
	ComplaintCount int    `db:"complaintCount" json:"complaintCount"`
	OpenCount      int    `db:"openCount" json:"openCount"`
}
