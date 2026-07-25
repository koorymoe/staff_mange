package model

import "time"

// AssistantConversation يسجّل كل تبادل رسالة/جواب بين موظف عادي والمساعد الذكي
// (نقطة Ask فقط) — يخلي المالك يراجع كل محادثات كل الموظفين مع الذكاء الاصطناعي.
type AssistantConversation struct {
	ID         string         `db:"id" json:"id"`
	EmployeeID string         `db:"employeeId" json:"employeeId"`
	Message    string         `db:"message" json:"message"`
	Reply      string         `db:"reply" json:"reply"`
	CreatedAt  time.Time      `db:"createdAt" json:"createdAt"`
	Employee   *EmployeeBrief `db:"-" json:"employee"`
}
