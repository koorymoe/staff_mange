package model

import "time"

// ═══ المهمة الإضافية ═══
//
// شغل موجّه من المدير لموظف، مو مربوط بحجز: «خرّج فواتير الشهر»،
// «رتّب المخزن»، «راجع عقود الزبائن».
//
// ⚠️ اقرا رأس schema_extra_task.go لسبب فصلها عن الحجوزات.
type ExtraTask struct {
	ID           string     `db:"id" json:"id"`
	Title        string     `db:"title" json:"title"`
	Description  *string    `db:"description" json:"description,omitempty"`
	AssignedToID string     `db:"assignedToId" json:"assignedToId"`
	AssignedByID *string    `db:"assignedById" json:"assignedById,omitempty"`
	Priority     string     `db:"priority" json:"priority"`
	DueAt        *time.Time `db:"dueAt" json:"dueAt,omitempty"`
	Status       string     `db:"status" json:"status"`
	SeenAt       *time.Time `db:"seenAt" json:"seenAt,omitempty"`
	StartedAt    *time.Time `db:"startedAt" json:"startedAt,omitempty"`
	DoneAt       *time.Time `db:"doneAt" json:"doneAt,omitempty"`
	DoneNote     *string    `db:"doneNote" json:"doneNote,omitempty"`
	CancelledAt  *time.Time `db:"cancelledAt" json:"cancelledAt,omitempty"`
	CancelReason *string    `db:"cancelReason" json:"cancelReason,omitempty"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updatedAt" json:"updatedAt"`

	// أسماء محسوبة وقت الجلب — مو أعمدة.
	AssignedToName *string `db:"-" json:"assignedToName,omitempty"`
	AssignedByName *string `db:"-" json:"assignedByName,omitempty"`
	// Overdue محسوبة: فات موعدها وما انخلصت.
	Overdue bool `db:"-" json:"overdue"`
}

const (
	ExtraTaskNew        = "NEW"
	ExtraTaskInProgress = "IN_PROGRESS"
	ExtraTaskDone       = "DONE"
	ExtraTaskCancelled  = "CANCELLED"

	ExtraTaskNormal = "NORMAL"
	ExtraTaskUrgent = "URGENT"
)

func ExtraTaskStatusLabel(s string) string {
	switch s {
	case ExtraTaskNew:
		return "جديدة"
	case ExtraTaskInProgress:
		return "قيد التنفيذ"
	case ExtraTaskDone:
		return "منجزة"
	case ExtraTaskCancelled:
		return "ملغاة"
	}
	return s
}

// CreateExtraTaskRequest توجيه مهمة لموظف.
type CreateExtraTaskRequest struct {
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	AssignedToID string  `json:"assignedToId"`
	Priority     string  `json:"priority"`
	// DueAt نص ISO أو فاضي — «بلا موعد» حالة مشروعة، فما نفرض تاريخاً
	// وهمياً يخلي المهمة تطلع «متأخرة» وهي أصلاً بلا موعد.
	DueAt *string `json:"dueAt"`
}

// CompleteExtraTaskRequest إنجاز المهمة.
//
// ⚠️ الوصف إجباري: «تم» بلا شرح ما تفيد لا بالمتابعة ولا بالتقييم،
// والمدير يرجع يسأل «تم شنو بالضبط؟» — وهاي نفس المكالمة الي بنينا
// الميزة حتى نلغيها.
type CompleteExtraTaskRequest struct {
	DoneNote string `json:"doneNote"`
}
