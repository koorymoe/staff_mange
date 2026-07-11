package model

import "time"

type Attendance struct {
	ID         string     `db:"id" json:"id"`
	EmployeeID string     `db:"employeeId" json:"employeeId"`
	CheckIn    time.Time  `db:"checkIn" json:"checkIn"`
	CheckOut   *time.Time `db:"checkOut" json:"checkOut"`
	Date       time.Time  `db:"date" json:"date"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type SetAttendanceCorrectionRequest struct {
	CheckIn  *time.Time `json:"checkIn"`
	CheckOut *time.Time `json:"checkOut"`
}

type MonthlyAttendanceReport struct {
	EmployeeID   string       `json:"employeeId"`
	Month        string       `json:"month"`
	Days         []Attendance `json:"days"`
	DaysPresent  int          `json:"daysPresent"`
	TotalMinutes int          `json:"totalMinutes"`
}
