package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// أنواع أفعال ماتركس — كلها تذكيرات وتنبيهات، ماكو فعل يلمس فلوس أو
// عقوبة أو تكليف كادر.
const (
	AiActionPaperworkReminder = "PAPERWORK_REMINDER" // الليدر: ورق متأخر
	AiActionDelayWarning      = "DELAY_WARNING"      // الليدر: حجز اليوم متوقع يطوّل
	AiActionCustomerFollowUp  = "CUSTOMER_FOLLOWUP"  // الجودة: زبون قريب يزعل
	AiActionUnstaffedAlert    = "UNSTAFFED_ALERT"    // التنسيق: حجز باچر بلا كادر
	AiActionReplacementAlert  = "REPLACEMENT_ALERT"  // الآيتي/الأسطول: جهاز أو سيارة تكلّف
	AiActionGpsExpiry         = "GPS_EXPIRY"         // البائع: اشتراك جي بي اس قرب يخلص
	AiActionVehicleDocExpiry  = "VEHICLE_DOC_EXPIRY" // الأسطول: وثيقة سيارة قربت تخلص
	AiActionExtraTaskOverdue  = "EXTRA_TASK_OVERDUE" // الموظف: مهمة إضافية فات موعدها
	AiActionAttendanceNudge   = "ATTENDANCE_NUDGE"   // الموظف: عنده شغل اليوم وما سجّل حضور
	AiActionLowStock          = "LOW_STOCK"          // المخزن: أداة قربت تخلص
	AiActionInvoiceApproval   = "INVOICE_APPROVAL"   // المحاسبة: فاتورة تنتظر اعتماد
	AiActionPredictionNudge   = "PREDICTION_NUDGE"   // الموظف: توقّع ماتركس وافق عليه المدير
	AiActionCrewRating        = "CREW_RATING"        // الليدر: خلّص حجز وما قيّم فنيّيه
)

// AiActionEscalateAfter تذكير ما انحل بهالمدة يصعد للمدير.
const AiActionEscalateAfter = 3 * 24 * time.Hour

// AiActionPauseAfterRejects المدير رفض نفس النوع هالعدد بآخر ٣٠ يوم؟ يوقف.
const AiActionPauseAfterRejects = 3

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
	AiActionGpsExpiry:         "اشتراك جي بي اس يخلص",
	AiActionVehicleDocExpiry:  "وثيقة سيارة تخلص",
	AiActionExtraTaskOverdue:  "مهمة إضافية متأخرة",
	AiActionAttendanceNudge:   "تذكير تسجيل الحضور",
	AiActionLowStock:          "أداة قربت تخلص بالمخزن",
	AiActionInvoiceApproval:   "فاتورة تنتظر اعتماد",
	AiActionPredictionNudge:   "تذكير بعد توقّع",
	AiActionCrewRating:        "تقييم الفنيين بعد الحجز",
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
	Details          NullJSON   `db:"details" json:"details"`
	ResolvedAt       *time.Time `db:"resolvedAt" json:"resolvedAt"`
	EscalatedAt      *time.Time `db:"escalatedAt" json:"escalatedAt"`
}

// AiActionKindPause نوع وقّفه ماتركس لحاله.
type AiActionKindPause struct {
	Kind      string     `db:"kind" json:"kind"`
	Reason    string     `db:"reason" json:"reason"`
	PausedAt  time.Time  `db:"pausedAt" json:"pausedAt"`
	ResumedAt *time.Time `db:"resumedAt" json:"resumedAt"`
}

// MatrixAccuracy دقة ماتركس لآخر ٣٠ يوم.
type MatrixAccuracy struct {
	Actions        int `db:"actions" json:"actions"`
	Resolved       int `db:"resolved" json:"resolved"`
	Escalated      int `db:"escalated" json:"escalated"`
	Rejected       int `db:"rejected" json:"rejected"`
	DelayPredicted int `db:"delayPredicted" json:"delayPredicted"`
	DelayChecked   int `db:"delayChecked" json:"delayChecked"`
	DelayCorrect   int `db:"delayCorrect" json:"delayCorrect"`
}

// PendingAiDecision حكم ماتركس ينتظر قرار بشر (من صندوق المراقب).
type PendingAiDecision struct {
	ID         string    `db:"id" json:"id"`
	EntityType string    `db:"entityType" json:"entityType"`
	EntityID   string    `db:"entityId" json:"entityId"`
	Title      string    `db:"title" json:"title"`
	Summary    *string   `db:"summary" json:"summary"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
	// منو وأي حجز — من الإشارة الأصلية (للمدير والمالك بس).
	EmployeeID   *string    `db:"employeeId" json:"employeeId"`
	EmployeeName *string    `db:"employeeName" json:"employeeName"`
	BookingID    *string    `db:"bookingId" json:"bookingId"`
	BookingCode  *string    `db:"bookingCode" json:"bookingCode"`
	OccurredAt   *time.Time `db:"occurredAt" json:"occurredAt"`
}

// UnstaffedBooking حجز مثبّت بلا أي كادر مكلّف.
type UnstaffedBooking struct {
	ID          string    `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	ScheduledAt time.Time `db:"scheduledAt" json:"scheduledAt"`
}

// NullJSON عمود JSONB ممكن يكون NULL — json.RawMessage لحاله ما يقرا NULL.
type NullJSON json.RawMessage

func (j *NullJSON) Scan(v any) error {
	switch x := v.(type) {
	case nil:
		*j = nil
	case []byte:
		*j = append((*j)[:0], x...)
	case string:
		*j = NullJSON(x)
	default:
		return fmt.Errorf("NullJSON: نوع غير متوقع %T", v)
	}
	return nil
}

func (j NullJSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}
