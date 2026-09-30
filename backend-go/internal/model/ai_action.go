package model

import "time"

// أنواع أفعال ماتركس — كلها تذكيرات وتنبيهات، ماكو فعل يلمس فلوس أو
// عقوبة أو تكليف كادر.
const (
	AiActionPaperworkReminder = "PAPERWORK_REMINDER" // الليدر: ورق متأخر
	AiActionDelayWarning      = "DELAY_WARNING"      // الليدر: حجز اليوم متوقع يطوّل
	AiActionCustomerFollowUp  = "CUSTOMER_FOLLOWUP"  // الجودة: زبون قريب يزعل
	AiActionUnstaffedAlert    = "UNSTAFFED_ALERT"    // التنسيق: حجز باچر بلا كادر
	AiActionReplacementAlert  = "REPLACEMENT_ALERT"  // الآيتي/الأسطول: جهاز أو سيارة تكلّف
)

const (
	AiActionDone   = "DONE"
	AiActionUndone = "UNDONE"
)

// AiActionLabels اسم الفعل للعرض.
var AiActionLabels = map[string]string{
	AiActionPaperworkReminder: "تذكير بالورق المتأخر",
	AiActionDelayWarning:      "تنبيه تأخير لليدر",
	AiActionCustomerFollowUp:  "متابعة زبون قريب يزعل",
	AiActionUnstaffedAlert:    "حجز باچر بلا كادر",
	AiActionReplacementAlert:  "اقتراح استبدال",
}

// AiAction فعل واحد نفّذه ماتركس لحاله.
type AiAction struct {
	ID               string     `db:"id" json:"id"`
	Kind             string     `db:"kind" json:"kind"`
	EntityType       string     `db:"entityType" json:"entityType"`
	EntityID         string     `db:"entityId" json:"entityId"`
	Period           string     `db:"period" json:"period"`
	TargetEmployeeID *string    `db:"targetEmployeeId" json:"targetEmployeeId"`
	TargetLabel      string     `db:"targetLabel" json:"targetLabel"`
	Summary          string     `db:"summary" json:"summary"`
	Status           string     `db:"status" json:"status"`
	CreatedAt        time.Time  `db:"createdAt" json:"createdAt"`
	UndoneByID       *string    `db:"undoneById" json:"undoneById"`
	UndoneAt         *time.Time `db:"undoneAt" json:"undoneAt"`
}

// PendingAiDecision حكم ماتركس ينتظر قرار بشر (من صندوق المراقب).
type PendingAiDecision struct {
	ID         string    `db:"id" json:"id"`
	EntityType string    `db:"entityType" json:"entityType"`
	EntityID   string    `db:"entityId" json:"entityId"`
	Title      string    `db:"title" json:"title"`
	Summary    *string   `db:"summary" json:"summary"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

// UnstaffedBooking حجز مثبّت بلا أي كادر مكلّف.
type UnstaffedBooking struct {
	ID          string    `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	ScheduledAt time.Time `db:"scheduledAt" json:"scheduledAt"`
}
