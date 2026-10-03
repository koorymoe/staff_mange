package model

import "time"

// StaffRequest طلب كادر: مدير المشاريع يطلب موظفين محددين بوقت ومدة، وإدارة الكوادر تلبيه
type StaffRequest struct {
	ID            string     `db:"id" json:"id"`
	RequesterID   string     `db:"requesterId" json:"-"`
	ProjectID     *string    `db:"projectId" json:"projectId"`
	NeededAt      time.Time  `db:"neededAt" json:"neededAt"`
	DurationHours float64    `db:"durationHours" json:"durationHours"`
	Notes         *string    `db:"notes" json:"notes"`
	Status        string     `db:"status" json:"status"`
	HandledByID   *string    `db:"handledById" json:"-"`
	HandledAt     *time.Time `db:"handledAt" json:"handledAt"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`

	Requester   *EmployeeBrief  `db:"-" json:"requester"`
	HandledBy   *EmployeeBrief  `db:"-" json:"handledBy"`
	ProjectName *string         `db:"-" json:"projectName"`
	Employees   []EmployeeBrief `db:"-" json:"employees"`
}

type CreateStaffRequest struct {
	ProjectID     *string   `json:"projectId"`
	NeededAt      time.Time `json:"neededAt"`
	DurationHours float64   `json:"durationHours"`
	Notes         *string   `json:"notes"`
	EmployeeIDs   []string  `json:"employeeIds"`
}

type UpdateStaffRequestStatus struct {
	Status string `json:"status"` // APPROVED | REJECTED | FULFILLED
}
