package model

import "time"

type Expense struct {
	ID          string    `db:"id" json:"id"`
	EmployeeID  string    `db:"employeeId" json:"employeeId"`
	Amount      float64   `db:"amount" json:"amount"`
	Description *string   `db:"description" json:"description"`
	Status      string    `db:"status" json:"status"`
	// ═══ الحجز الي انصرف عليه ═══
	//
	// ⚠️ **قابل للفراغ**: المصاريف المسجّلة قبل هالتعديل ما إلها حجز،
	// وتنعرض بجدول منفصل «ما تنحسب» بدل ما تنحسب غلط أو تختفي بصمت.
	//
	// ⚠️ ولازم يكون بالـstruct لأن المستودع يستعمل `SELECT *` —
	// بدونه sqlx يفشل بـ«missing destination name bookingId».
	BookingID   *string   `db:"bookingId" json:"bookingId"`
	// BookingCode كود الحجز للعرض — ما ينخزن.
	BookingCode *string   `db:"-" json:"bookingCode,omitempty"`
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
	// BookingID الحجز الي انصرف عليه — الليدر يختاره من حجوزاته.
	BookingID *string `json:"bookingId"`
	EmployeeID  string   `json:"employeeId"`
	Amount      *float64 `json:"amount"`
	Description *string  `json:"description"`
}

type UpdateExpenseStatusRequest struct {
	Status string `json:"status"`
}
