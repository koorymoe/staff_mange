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
// الوجهة: المدير والمالك والمراقب (ع) 10-23 — «تطلع للمدير والمالك والمراقب».
// المراقب ما يشوف ولا يراجع إنجازات المدير/المالك (ما ينقيّمون) ولا إنجازه هو.
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
	// فويس مسج مرفق (جدول AchievementVoice) — يتعبّى بالمعالج.
	VoiceSeconds *int `db:"-" json:"voiceSeconds,omitempty"`
	// رد ماتركس المقترح (ع) 10-23 — بالقواعد محلياً، والي يراجع يعدّله قبل الإرسال.
	SuggestedStatus string `db:"-" json:"suggestedStatus,omitempty"`
	SuggestedReply  string `db:"-" json:"suggestedReply,omitempty"`
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
