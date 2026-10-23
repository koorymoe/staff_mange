package model

import "time"

type LoginAudit struct {
	ID         string    `db:"id" json:"id"`
	Username   string    `db:"username" json:"username"`
	EmployeeID *string   `db:"employeeId" json:"-"`
	Success    bool      `db:"success" json:"success"`
	IPAddress  *string   `db:"ipAddress" json:"ipAddress"`
	UserAgent  *string   `db:"userAgent" json:"userAgent"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}
