package model

import "time"

// ServiceManager يعمم فكرة "أبو الجي بي اس" لأي مجموعة خدمات — موظف واحد يُمنح مسؤولية
// خدمة أو أكثر (مثال: GPS + منظومات الصوت + منظومات الحريق سوا).
type ServiceManager struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"-"`
	ServiceID  string    `db:"serviceId" json:"-"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
	Service  *Service       `db:"-" json:"service"`
}

type SetServiceManagerRequest struct {
	EmployeeID string   `json:"employeeId"`
	ServiceIDs []string `json:"serviceIds"`
}
