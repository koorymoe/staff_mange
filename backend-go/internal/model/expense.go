package model

import "time"

type Expense struct {
	ID          string    `db:"id" json:"id"`
	EmployeeID  string    `db:"employeeId" json:"employeeId"`
	Amount      float64   `db:"amount" json:"amount"`
	Description *string   `db:"description" json:"description"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

// EmployeeBrief يطابق الحقول المُقتصرة اللي يرجّعها الباك إند القديم (id/name/position) بدون بقية بيانات الموظف الحساسة
type EmployeeBrief struct {
	ID       string  `db:"id" json:"id"`
	Name     string  `db:"name" json:"name"`
	Position *string `db:"position" json:"position"`
}

type CreateExpenseRequest struct {
	EmployeeID  string   `json:"employeeId"`
	Amount      *float64 `json:"amount"`
	Description *string  `json:"description"`
}

type UpdateExpenseStatusRequest struct {
	Status string `json:"status"`
}
