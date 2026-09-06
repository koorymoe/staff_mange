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
	EmployeeID   string            `json:"employeeId"`
	Month        string            `json:"month"`
	Days         []DailyAttendance `json:"days"`
	DaysPresent  int               `json:"daysPresent"`
	TotalMinutes int               `json:"totalMinutes"`
}

// DailyAttendance يمثل يوم واحد (تاريخ تقويمي) مع كل جلسات الحضور فيه —
// موظف ممكن يسجل دخول/خروج أكثر من مرة بنفس اليوم.
type DailyAttendance struct {
	Date         string       `json:"date"`
	Sessions     []Attendance `json:"sessions"`
	FirstCheckIn time.Time    `json:"firstCheckIn"`
	LastCheckOut *time.Time   `json:"lastCheckOut"`
	StillOpen    bool         `json:"stillOpen"`
	TotalMinutes int          `json:"totalMinutes"`
}

// EmployeeDailyAttendanceSummary يمثل ملخص حضور موظف واحد ليوم واحد (لجدول
// المراقب "كل الموظفين اليوم") — مجمّع من كل جلسات ذلك اليوم.
type EmployeeDailyAttendanceSummary struct {
	EmployeeID      string         `db:"employeeId" json:"employeeId"`
	Employee        *EmployeeBrief `db:"-" json:"employee"`
	SessionsCount   int            `db:"sessionsCount" json:"sessionsCount"`
	FirstCheckIn    time.Time      `db:"firstCheckIn" json:"firstCheckIn"`
	LastCheckOut    *time.Time     `db:"lastCheckOut" json:"lastCheckOut"`
	CurrentlyActive bool           `db:"currentlyActive" json:"currentlyActive"`
	TotalMinutes    int            `db:"totalMinutes" json:"totalMinutes"`
}
