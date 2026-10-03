package model

import "time"

// ═══ الإنجازات (ماتركس) ═══
//
// تقرير يومي حر من أي موظف بأي دور — شنو سوا اليوم. الهدف المعلن من
// صاحب النظام: يخلي ماتركس يتعرف أكثر على الموظفين وسلوكهم وطريقة
// تفكيرهم — مادة تغذية للتعلّم والقرار، أبعد من الحجز والفاتورة بس.
//
// ⚠️ الربط بحجز اختياري بالتصميم: الفني يربطه بحجز، بس مهندس الجودة/
// المراقب/المحاسب يرفعون تقريراً عاماً بلا حجز (شغلهم اليومي ما
// يرتبط بحجز بالضرورة).
//
// 🔴 الوجهة حصراً: مدير النظام والمالك — لا المراقب ولا زميل ثانٍ.
// مساره مستقل عن صندوق المراقب (`MonitorReview`) عمداً — ذاك يوصله
// دور MONITOR كمان (`requireMonitor`)، وهذا بالضبط الممنوع هنا.
type Achievement struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	Role         string     `db:"role" json:"role"`
	BookingID    *string    `db:"bookingId" json:"bookingId"`
	ReportText   string     `db:"reportText" json:"reportText"`
	ReviewStatus string     `db:"reviewStatus" json:"reviewStatus"` // PENDING | GOOD | NEEDS_REVIEW
	ReviewNote   *string    `db:"reviewNote" json:"reviewNote"`
	ReviewedByID *string    `db:"reviewedById" json:"reviewedById"`
	ReviewedAt   *time.Time `db:"reviewedAt" json:"reviewedAt"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`

	EmployeeName   *string `db:"-" json:"employeeName,omitempty"`
	BookingCode    *string `db:"-" json:"bookingCode,omitempty"`
	ReviewedByName *string `db:"-" json:"reviewedByName,omitempty"`
}

const (
	AchievementReviewPending = "PENDING"
	AchievementReviewGood    = "GOOD"
	AchievementReviewNeeds   = "NEEDS_REVIEW"
)

type CreateAchievementRequest struct {
	BookingID  *string `json:"bookingId"`
	ReportText string  `json:"reportText"`
}

type ReviewAchievementRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}
