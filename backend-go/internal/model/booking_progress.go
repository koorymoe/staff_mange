package model

import "time"

// BookingProgressReport تقرير يوم واحد على حجز ما انخلص بيوم واحد.
//
// ⚠️ هذي أعمدة الجدول كاملة — الجلب SELECT *، وأي عمود بلا حقل يفشّل
// الاستعلام كله بالسكوت.
type BookingProgressReport struct {
	ID            string    `db:"id" json:"id"`
	BookingID     string    `db:"bookingId" json:"bookingId"`
	DayNumber     int       `db:"dayNumber" json:"dayNumber"`
	ReportedByID  *string   `db:"reportedById" json:"reportedById"`
	WorkDone      string    `db:"workDone" json:"workDone"`
	RemainingWork string    `db:"remainingWork" json:"remainingWork"`
	PercentDone   int       `db:"percentDone" json:"percentDone"`
	Blockers      *string   `db:"blockers" json:"blockers"`
	MaterialsUsed *string   `db:"materialsUsed" json:"materialsUsed"`
	CrewSnapshot  *string   `db:"crewSnapshot" json:"crewSnapshot"`
	CreatedAt     time.Time `db:"createdAt" json:"createdAt"`

	ReportedBy *EmployeeBrief `db:"-" json:"reportedBy"`
}

// PartialCompleteRequest طلب «خلصنا جزء والباقي باچر».
//
// workDone وremainingWork إلزاميين: تقرير جزئي بلا «وين وصلنا» ما ينفع
// الكادر الجاي بشي، وهو كل سبب وجود الميزة.
type PartialCompleteRequest struct {
	WorkDone      string  `json:"workDone"`
	RemainingWork string  `json:"remainingWork"`
	PercentDone   int     `json:"percentDone"`
	Blockers      *string `json:"blockers"`
	MaterialsUsed *string `json:"materialsUsed"`
	AmountCollected *float64 `json:"amountCollected"`
}

// SuggestedCrewMember عضو كادر يقترحه النظام لإكمال الحجز — هم الي
// طلعوا بالأيام الفائتة.
type SuggestedCrewMember struct {
	EmployeeID string `db:"employeeId" json:"employeeId"`
	Name       string `db:"name" json:"name"`
	Role       string `db:"role" json:"role"`
	// كم يوم اشتغل على هذا الحجز — الي اشتغل أكثر يعرف أكثر
	DaysWorked int `db:"daysWorked" json:"daysWorked"`
	// متوفر اليوم المقترح لو لا — الإداري لازم يعرف قبل ما يختار
	Available bool `db:"-" json:"available"`
	Note      string `db:"-" json:"note"`
}
