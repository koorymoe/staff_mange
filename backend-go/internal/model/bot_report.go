package model

import (
	"time"

	"github.com/lib/pq"
)

// BotReport بلاغ موظف عن النظام وصل من بوت التلغرام (دازاي 11-05).
type BotReport struct {
	ID           string         `db:"id" json:"id"`
	Number       int            `db:"number" json:"number"`
	EmployeeID   string         `db:"employeeId" json:"employeeId"`
	ChatID       int64          `db:"chatId" json:"-"`
	Body         string         `db:"body" json:"body"`
	Photos       pq.StringArray `db:"photos" json:"photos"`
	MediaGroupID *string        `db:"mediaGroupId" json:"-"`
	Status       string         `db:"status" json:"status"`
	AdminNote    *string        `db:"adminNote" json:"adminNote"`
	CreatedAt    time.Time      `db:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time      `db:"updatedAt" json:"updatedAt"`
	// Source من وين وصل: telegram | system. Page الشاشة الي كان بيها (للي من النظام).
	Source string  `db:"source" json:"source"`
	Page   *string `db:"page" json:"page"`
	// PhotosPurgedAt متى انمسحت الصور من السيرفر (قرار دازاي: كل صورة عمرها أكثر من أسبوع).
	PhotosPurgedAt *time.Time `db:"photosPurgedAt" json:"photosPurgedAt"`
	// للعرض — من جدول الموظف
	EmployeeName string `db:"employeeName" json:"employeeName"`
	EmployeeRole string `db:"employeeRole" json:"employeeRole"`
}

// حالات البلاغ. كل تغيير يوصل للموظف بالبوت.
var BotReportStatusLabel = map[string]string{
	"NEW":         "🆕 جديد",
	"NEEDS_INFO":  "❓ يحتاج توضيح",
	"IN_PROGRESS": "🔧 قيد العمل",
	"DONE":        "✅ انحل",
	"NOT_A_BUG":   "💡 مو خلل — انشرحله",
	"REJECTED":    "✖️ ما ينسوّى",
}

type TelegramLink struct {
	EmployeeID string    `db:"employeeId" json:"-"`
	ChatID     int64     `db:"chatId" json:"-"`
	TgUsername *string   `db:"tgUsername" json:"tgUsername"`
	LinkedAt   time.Time `db:"linkedAt" json:"linkedAt"`
}

type BotReportDraft struct {
	ChatID     int64          `db:"chatId"`
	EmployeeID string         `db:"employeeId"`
	Body       string         `db:"body"`
	Photos     pq.StringArray `db:"photos"`
	UpdatedAt  time.Time      `db:"updatedAt"`
}
