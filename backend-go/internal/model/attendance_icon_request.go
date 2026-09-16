package model

import "time"

type AttendanceIconRequest struct {
	ID            string     `db:"id" json:"id"`
	EmployeeID    string     `db:"employeeId" json:"-"`
	RequestedIcon string     `db:"requestedIcon" json:"requestedIcon"`
	Status        string     `db:"status" json:"status"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt    *time.Time `db:"resolvedAt" json:"resolvedAt"`
	ResolvedByID  *string    `db:"resolvedById" json:"-"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type CreateAttendanceIconRequest struct {
	RequestedIcon string `json:"requestedIcon"`
}
