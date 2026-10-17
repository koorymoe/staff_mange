package model

import "time"

// Notification تنبيه لموظف معيّن — يظهر بجرس الإشعارات بالواجهة (مثال: تصدر ترتيب
// الفنيين، خصم نقطة كي بي اي). لا ترتبط بأي إجراء تلقائي، هي إعلام بس.
type Notification struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	Type       string    `db:"type" json:"type"`
	Message    string    `db:"message" json:"message"`
	Read       bool      `db:"read" json:"read"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}
