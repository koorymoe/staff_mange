package model

import "time"

// CoordinationAlertThreshold عدد التنبيهات الي عندها ينشر إعلان.
//
// ⚠️ رقم واحد بمكان واحد: الشاشة تعرض «٣/١٠» والخادم يقرر النشر —
// ولو انكتب الرقم بمكانين يصير عندنا شاشة تكول «وصل العشرة»
// وخادم ما ينشر (أو العكس).
const CoordinationAlertThreshold = 10

// CoordinationAlert تنبيه سجّله المراقب على إداري ما ثبّت حجزاً.
type CoordinationAlert struct {
	ID              string     `db:"id" json:"id"`
	BookingID       string     `db:"bookingId" json:"bookingId"`
	CoordinatorID   *string    `db:"coordinatorId" json:"-"`
	CoordinatorName *string    `db:"coordinatorName" json:"coordinatorName"`
	Reason          *string    `db:"reason" json:"reason"`
	ByEmployeeID    *string    `db:"byEmployeeId" json:"-"`
	ByName          *string    `db:"byName" json:"byName"`
	CreatedAt       time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt      *time.Time `db:"resolvedAt" json:"resolvedAt"`
	ResolvedByID    *string    `db:"resolvedById" json:"-"`
	ResolvedByName  *string    `db:"resolvedByName" json:"resolvedByName"`
	ResolveNote     *string    `db:"resolveNote" json:"resolveNote"`
}

// CoordinationAlertSummary ملخص حجز واحد — الي تعرضه البطاقة.
//
// ⚠️ العدّاد `openCount` **يعدّ المفتوحة بس**: «تمت المعالجة»
// تصفّر العدّاد الظاهر، والسجل الكامل يبقى بـ`totalCount`.
type CoordinationAlertSummary struct {
	BookingID   string     `db:"bookingId" json:"bookingId"`
	OpenCount   int        `db:"openCount" json:"openCount"`
	TotalCount  int        `db:"totalCount" json:"totalCount"`
	LastAlertAt *time.Time `db:"lastAlertAt" json:"lastAlertAt"`
}

// AddCoordinationAlertRequest تسجيل تقصير.
type AddCoordinationAlertRequest struct {
	Reason string `json:"reason"`
}

// ResolveCoordinationAlertsRequest «تمت المعالجة».
type ResolveCoordinationAlertsRequest struct {
	Note string `json:"note"`
}
