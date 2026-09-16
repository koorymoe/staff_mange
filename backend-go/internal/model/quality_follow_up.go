package model

import "time"

type QualityFollowUp struct {
	ID                    string     `db:"id" json:"id"`
	BookingID             string     `db:"bookingId" json:"-"`
	CustomerID            string     `db:"customerId" json:"-"`
	Status                string     `db:"status" json:"status"` // PENDING | CONTACTED_OK | CONTACTED_ISSUE | CONVERTED | CLOSED
	ContactNotes          *string    `db:"contactNotes" json:"contactNotes"`
	ContactedByEmployeeID *string    `db:"contactedByEmployeeId" json:"-"`
	ContactedAt           *time.Time `db:"contactedAt" json:"contactedAt"`
	CreatedAt             time.Time  `db:"createdAt" json:"createdAt"`

	// ── حكم الجودة ──
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (الجلب SELECT *).
	ReportType          *string    `db:"reportType" json:"reportType"`             // POSITIVE | NEGATIVE
	InspectionStatus    string     `db:"inspectionStatus" json:"inspectionStatus"` // NONE | PENDING | DONE
	InspectionResult    *string    `db:"inspectionResult" json:"inspectionResult"` // CUSTOMER_RIGHT | CUSTOMER_WRONG
	InspectionNotes     *string    `db:"inspectionNotes" json:"inspectionNotes"`
	InspectedByID       *string    `db:"inspectedById" json:"-"`
	InspectedAt         *time.Time `db:"inspectedAt" json:"inspectedAt"`
	PenalizedEmployeeID *string    `db:"penalizedEmployeeId" json:"-"`
	KpiEvaluationID     *string    `db:"kpiEvaluationId" json:"-"`

	Booking             *Booking       `db:"-" json:"booking"`
	Customer            *Customer      `db:"-" json:"customer"`
	ContactedByEmployee *EmployeeBrief `db:"-" json:"contactedByEmployee"`
	InspectedBy         *EmployeeBrief `db:"-" json:"inspectedBy"`
	// منو انغرم (أو منو راح ينغرم) — الليدر المسؤول عن الحجز
	PenalizedEmployee *EmployeeBrief `db:"-" json:"penalizedEmployee"`
	// تفاصيل التنفيذ: منو طلع، ومتى بدا وخلّص، وشكد استغرق
	Execution *BookingExecutionDetail `db:"-" json:"execution"`

	// تفاصيل المشروع والمبالغ — مهندس الجودة يحتاجها وهو يتصل بالزبون حتى
	// يعرف شنو انتفق عليه وشكد انستلم فعلاً، ويقدر يكتب تفاصيل الفارق.
	Financials *QualityFollowUpFinancials `db:"-" json:"financials"`
}

// QualityFollowUpFinancials صورة كاملة عن مشروع/حجز المتابعة وأمواله.
type QualityFollowUpFinancials struct {
	BookingCode  string  `db:"bookingCode" json:"bookingCode"`
	ServiceName  *string `db:"serviceName" json:"serviceName"`
	Location     *string `db:"location" json:"location"`
	WorkDetails  *string `db:"workDetails" json:"workDetails"` // ملاحظات الحجز = تفاصيل العمل
	ProjectID    *string `db:"projectId" json:"projectId"`
	ProjectCode  *string `db:"projectCode" json:"projectCode"`
	ProjectName  *string `db:"projectName" json:"projectName"`
	ProjectStage *string `db:"projectStage" json:"projectStage"`

	QuotedPrice     *float64 `db:"quotedPrice" json:"quotedPrice"`         // السعر المتفق عليه
	ProjectPrice    *float64 `db:"projectPrice" json:"projectPrice"`       // سعر المشروع لو مرحّل
	AdvancePaid     *float64 `db:"advancePaid" json:"advancePaid"`         // العربون
	AmountCollected *float64 `db:"amountCollected" json:"amountCollected"` // المستلم فعلاً

	// Difference = المتفق عليه − (العربون + المستلم). موجب يعني باقي بذمة
	// الزبون، سالب يعني انستلم أكثر من المتفق.
	AgreedTotal   float64 `db:"-" json:"agreedTotal"`
	ReceivedTotal float64 `db:"-" json:"receivedTotal"`
	Difference    float64 `db:"-" json:"difference"`
}

type UpdateQualityFollowUpRequest struct {
	Status       string  `json:"status"`
	ContactNotes *string `json:"contactNotes"`
}

// BookingExecutionDetail «شنو صار بهذا الحجز فعلاً» — يقراها مهندس
// الجودة قبل ما يتصل بالزبون.
//
// بدونها هو يتصل وهو ما يعرف منو طلع ولا شكد استغرقوا، فيسأل الزبون
// أسئلة عامة ويكتب ملاحظة عامة. ومع التفاصيل يقدر يسأل سؤال محدد:
// «الفريق وصلكم الساعة ٩ وخلّص ١١، هل هذا صحيح؟».
type BookingExecutionDetail struct {
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	// المدة بالدقائق — نحسبها بالسيرفر حتى تكون وحدة بكل مكان
	DurationMinutes *int                `json:"durationMinutes"`
	CompletionNotes *string             `json:"completionNotes"`
	WorkStoppedAt   *time.Time          `json:"workStoppedAt"`
	WorkStopReason  *string             `json:"workStopReason"`
	Crew            []BookingCrewMember `json:"crew"`
	// تقارير الإنجاز الجزئي إذا الحجز أخذ أكثر من يوم
	ProgressReports []BookingProgressReport `json:"progressReports"`
}

// BookingCrewMember عضو من الكادر الي طلع للحجز.
type BookingCrewMember struct {
	EmployeeID string `db:"employeeId" json:"employeeId"`
	Name       string `db:"name" json:"name"`
	Role       string `db:"role" json:"role"`
	IsLeader   bool   `db:"isLeader" json:"isLeader"`
}

// QualityVerdictRequest حكم مهندس الجودة بعد ما يتصل بالزبون.
type QualityVerdictRequest struct {
	// POSITIVE | NEGATIVE
	ReportType string `json:"reportType"`
	Notes      string `json:"notes"`
	// بالتقرير السلبي: هل يوقف الغرامة لحد الكشف؟
	NeedsInspection bool `json:"needsInspection"`
}

// QualityInspectionRequest نتيجة الكشف الميداني.
type QualityInspectionRequest struct {
	// CUSTOMER_RIGHT | CUSTOMER_WRONG
	Result string `json:"result"`
	Notes  string `json:"notes"`
}

// QualityKpiCriterion اسم معيار الكي بي اي الي تنخصم منه شكوى الزبون.
// موجود بجدول KpiCriterion ضمن المعايير الثمانية.
const QualityKpiCriterion = "شكوى الزبائن"
