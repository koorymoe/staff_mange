package model

import "time"

type Booking struct {
	ID                     string     `db:"id" json:"id"`
	Code                   string     `db:"code" json:"code"`
	SequenceNumber         *int       `db:"sequenceNumber" json:"sequenceNumber"`
	ScheduledAt            *time.Time `db:"scheduledAt" json:"scheduledAt"`
	PendingScheduledAt     *time.Time `db:"pendingScheduledAt" json:"pendingScheduledAt"`
	CustomerID             string     `db:"customerId" json:"-"`
	ServiceID              *string    `db:"serviceId" json:"-"`
	TransferEmployeeID     *string    `db:"transferEmployeeId" json:"-"`
	ProjectSupervisorID    *string    `db:"projectSupervisorId" json:"-"`
	ExpenseResponsibleID   *string    `db:"expenseResponsibleId" json:"expenseResponsibleId"`
	ConfirmedByEmployeeID  *string    `db:"confirmedByEmployeeId" json:"-"`
	Notes                  *string    `db:"notes" json:"notes"`
	VehicleType            *string    `db:"vehicleType" json:"vehicleType"`
	Priority               string     `db:"priority" json:"priority"`
	Status                 string     `db:"status" json:"status"`
	TransferToProjects     bool       `db:"transferToProjects" json:"transferToProjects"`
	ConfirmedByName        *string    `db:"confirmedByName" json:"confirmedByName"`
	AdminNotes             *string    `db:"adminNotes" json:"adminNotes"`
	AssignedVehicle        *string    `db:"assignedVehicle" json:"assignedVehicle"`
	QuotedPrice            *float64   `db:"quotedPrice" json:"quotedPrice"`
	Address                *string    `db:"address" json:"address"`
	MapLocation            *string    `db:"mapLocation" json:"mapLocation"`
	MapLatitude            *float64   `db:"mapLatitude" json:"mapLatitude"`
	MapLongitude           *float64   `db:"mapLongitude" json:"mapLongitude"`
	CompletedAt            *time.Time `db:"completedAt" json:"completedAt"`
	CompletionNotes        *string    `db:"completionNotes" json:"completionNotes"`
	AmountCollected        *float64   `db:"amountCollected" json:"amountCollected"`
	AdvancePaid            *float64   `db:"advancePaid" json:"advancePaid"`
	AmountVerified         bool       `db:"amountVerified" json:"amountVerified"`
	EquipmentStatus        string     `db:"equipmentStatus" json:"equipmentStatus"`
	Shift                  *string    `db:"shift" json:"shift"`
	DeviceCount            *int       `db:"deviceCount" json:"deviceCount"`
	InspectionSupervisorID *string    `db:"inspectionSupervisorId" json:"inspectionSupervisorId"`
	ProjectCar             *string    `db:"projectCar" json:"projectCar"`
	CrewNotes              *string    `db:"crewNotes" json:"crewNotes"`
	BookingType            string     `db:"bookingType" json:"bookingType"`
	Urgency                *string    `db:"urgency" json:"urgency"`
	MaintenanceType        *string    `db:"maintenanceType" json:"maintenanceType"`
	RemembersExecutionCrew bool       `db:"remembersExecutionCrew" json:"remembersExecutionCrew"`
	SystemCount            *int       `db:"systemCount" json:"systemCount"`
	SystemType             *string    `db:"systemType" json:"systemType"`
	ProjectSpeed           *string    `db:"projectSpeed" json:"projectSpeed"`
	WorkType               *string    `db:"workType" json:"workType"`
	// WorkLocation وين انشتغل الشغل: عند الزبون لو داخل الشركة (بالورشة).
	// ⚠️ عمود بجدول Booking — لازم يضل إله حقل هنا لأن الاستعلام SELECT *.
	WorkLocation string `db:"workLocation" json:"workLocation"`
	// حجز داخل الشركة: الشغل لموظف من موظفينا مو لزبون خارجي.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	InternalEmployeeName  *string `db:"internalEmployeeName" json:"internalEmployeeName"`
	InternalEmployeePhone *string `db:"internalEmployeePhone" json:"internalEmployeePhone"`
	InternalDepartment    *string `db:"internalDepartment" json:"internalDepartment"`
	InternalApproved      *bool   `db:"internalApproved" json:"internalApproved"`
	// توقف العمل: الليدر بدأ وما كدر يكمّل. الحجز يضل شغّال ويكدر
	// يكمّله بعدين، بس يبين «متوقف» بتنسيق الحجوزات مع سببه.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	WorkStoppedAt   *time.Time `db:"workStoppedAt" json:"workStoppedAt"`
	WorkStopReason  *string    `db:"workStopReason" json:"workStopReason"`
	WorkStoppedByID *string    `db:"workStoppedById" json:"workStoppedById"`

	// ═══ اكتمال الحجز بعد الإنجاز ═══
	// الإنجاز لحاله ما يكفي: الليدر لازم يسوي فاتورة التكاليف المربوطة
	// بالحجز، وتقرير العمل. هذولا محسوبات وقت الجلب (مو أعمدة بالجدول)
	// من وجود LeaderInvoice و WorkReport مربوطات بهذا الحجز.
	HasInvoice bool `db:"-" json:"hasInvoice"`
	HasReport  bool `db:"-" json:"hasReport"`
	// CompletionState الحالة الي تنعرض بتنسيق الحجوزات:
	//   ASSIGNED         مكلّف (منترحّل لليدر، لسه ما انتهى)
	//   STOPPED          توقف العمل
	//   DONE_NO_BOTH     تم الإنجاز بدون فاتورة وتقرير
	//   DONE_NO_INVOICE  تم الإنجاز بدون فاتورة
	//   DONE_NO_REPORT   تم الإنجاز بدون تقرير
	//   DONE_FULL        تم الإنجاز بشكل كامل
	CompletionState string `db:"-" json:"completionState"`

	AddressDescription      *string    `db:"addressDescription" json:"addressDescription"`
	CreatedAt               time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt               time.Time  `db:"updatedAt" json:"updatedAt"`
	MaterialsReadyAt        *time.Time `db:"materialsReadyAt" json:"materialsReadyAt"`
	MaterialsReadyByID      *string    `db:"materialsReadyById" json:"-"`
	ResponseMinutes         *int       `db:"responseMinutes" json:"responseMinutes"`
	ArrivedAt               *time.Time `db:"arrivedAt" json:"arrivedAt"`
	StartedAt               *time.Time `db:"startedAt" json:"startedAt"`
	ConfirmationContactedAt *time.Time `db:"confirmationContactedAt" json:"confirmationContactedAt"`
	// وقت تحويل الحجز لتنسيق الحجوزات (التثبيت)
	ConfirmedAt               *time.Time     `db:"confirmedAt" json:"confirmedAt"`
	LocationUrl               *string        `db:"locationUrl" json:"locationUrl"`
	ConfirmationContactedByID *string        `db:"confirmationContactedById" json:"-"`
	ConfirmationContactedBy   *EmployeeBrief `db:"-" json:"confirmationContactedBy"`
	LastEditedByID            *string        `db:"lastEditedById" json:"-"`
	LastEditedAt              *time.Time     `db:"lastEditedAt" json:"lastEditedAt"`
	LastEditedBy              *EmployeeBrief `db:"-" json:"lastEditedBy"`

	Customer *Customer `db:"-" json:"customer"`
	// Service الخدمة الرئيسية (توافق مع الشاشات القديمة)، و Services كل
	// الخدمات المطلوبة بنفس الحجز — الزبون ممكن يطلب أكثر من منظومة سوة.
	Service             *Service            `db:"-" json:"service"`
	Services            []Service           `db:"-" json:"services"`
	TransferEmployee    *Employee           `db:"-" json:"transferEmployee"`
	ProjectSupervisor   *Employee           `db:"-" json:"projectSupervisor"`
	ConfirmedByEmployee *Employee           `db:"-" json:"confirmedByEmployee"`
	ExpenseResponsible  *Employee           `db:"-" json:"expenseResponsible"`
	MaterialsReadyBy    *EmployeeBrief      `db:"-" json:"materialsReadyBy"`
	Assignments         []BookingAssignment `db:"-" json:"assignments"`
	CartItems           []CartItem          `db:"-" json:"cartItems"`
	ScheduleLogs        []ScheduleChangeLog `db:"-" json:"scheduleLogs"`
}

type BookingAssignment struct {
	ID         string    `db:"id" json:"id"`
	BookingID  string    `db:"bookingId" json:"-"`
	EmployeeID string    `db:"employeeId" json:"-"`
	Role       string    `db:"role" json:"role"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
	Employee   Employee  `db:"-" json:"employee"`
}

type CartItem struct {
	ID          string    `db:"id" json:"id"`
	BookingID   string    `db:"bookingId" json:"bookingId"`
	ProductName string    `db:"productName" json:"productName"`
	Quantity    float64   `db:"quantity" json:"quantity"`
	UnitPrice   float64   `db:"unitPrice" json:"unitPrice"`
	TotalPrice  float64   `db:"totalPrice" json:"totalPrice"`
	Notes       *string   `db:"notes" json:"notes"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `db:"updatedAt" json:"updatedAt"`
}

type ScheduleChangeLog struct {
	ID          string     `db:"id" json:"id"`
	BookingID   string     `db:"bookingId" json:"-"`
	ChangedByID string     `db:"changedById" json:"changedById"`
	OldTime     *time.Time `db:"oldTime" json:"oldTime"`
	NewTime     time.Time  `db:"newTime" json:"newTime"`
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
	ChangedBy   *Employee  `db:"-" json:"changedBy"`
}

type CreateBookingRequest struct {
	CustomerID string  `json:"customerId"`
	ServiceID  *string `json:"serviceId"`
	// ServiceIDs كل الخدمات المطلوبة بنفس الحجز (الأولى تنعتبر الرئيسية).
	// لو انرسلت فاضية ننزل على serviceId المفرد حتى ما ننكسر مع أي شاشة قديمة.
	ServiceIDs         []string `json:"serviceIds"`
	Notes              *string  `json:"notes"`
	VehicleType        *string  `json:"vehicleType"`
	Priority           *string  `json:"priority"`
	TransferEmployeeID *string  `json:"transferEmployeeId"`
	Address            *string  `json:"address"`
	MapLatitude        *float64 `json:"mapLatitude"`
	MapLongitude       *float64 `json:"mapLongitude"`
	// رابط الموقع (كوكل ماب) — بديل عن التأشير على الخريطة، نفس فكرة الموردين
	LocationUrl *string `json:"locationUrl"`

	// حجز داخل الشركة: نوع الحجز INTERNAL مع معلومات الموظف الطالب.
	// الخدمات والموقع يبقون مثل أي حجز.
	BookingType           *string `json:"bookingType"`
	InternalEmployeeName  *string `json:"internalEmployeeName"`
	InternalEmployeePhone *string `json:"internalEmployeePhone"`
	InternalDepartment    *string `json:"internalDepartment"`
	InternalApproved      *bool   `json:"internalApproved"`
}

type ConfirmBookingRequest struct {
	ConfirmedByName       *string  `json:"confirmedByName"`
	ConfirmedByEmployeeID *string  `json:"confirmedByEmployeeId"`
	AdminNotes            *string  `json:"adminNotes"`
	TransferToProjects    bool     `json:"transferToProjects"`
	QuotedPrice           *float64 `json:"quotedPrice"`
	Address               *string  `json:"address"`
	ScheduledAt           *string  `json:"scheduledAt"`
}

type AssignBookingRequest struct {
	EmployeeID      string  `json:"employeeId"`
	Role            string  `json:"role"`
	AssignedVehicle *string `json:"assignedVehicle"`
}

type CreateCartItemRequest struct {
	ProductName string   `json:"productName"`
	Quantity    *float64 `json:"quantity"`
	UnitPrice   *float64 `json:"unitPrice"`
	Notes       *string  `json:"notes"`
}

type UpdateCartItemRequest struct {
	ProductName *string  `json:"productName"`
	Quantity    *float64 `json:"quantity"`
	UnitPrice   *float64 `json:"unitPrice"`
	Notes       *string  `json:"notes"`
}

type UpdateBookingDetailsRequest struct {
	QuotedPrice          *float64 `json:"quotedPrice"`
	Address              *string  `json:"address"`
	AssignedVehicle      *string  `json:"assignedVehicle"`
	MapLocation          *string  `json:"mapLocation"`
	MapLatitude          *float64 `json:"mapLatitude"`
	MapLongitude         *float64 `json:"mapLongitude"`
	ExpenseResponsibleID *string  `json:"expenseResponsibleId"`
	// رابط الموقع (بديل عن التحديد على الخريطة) — نفس فكرة الموردين
	LocationUrl *string `json:"locationUrl"`
	// قائمة الخدمات المطلوبة بالحجز (لو انرسلت، تستبدل القائمة الحالية)
	ServiceIDs []string `json:"serviceIds"`
}

type CompleteBookingRequest struct {
	CompletionNotes *string  `json:"completionNotes"`
	AmountCollected *float64 `json:"amountCollected"`
	AdvancePaid     *float64 `json:"advancePaid"`
	// WorkLocation وين انجز الشغل — ينسأل وقت الإنجاز لأن هذا الوقت
	// الوحيد الي نعرف بيه الجواب أكيد.
	WorkLocation *string `json:"workLocation"`
}

// وين انشتغل الشغل — أساس إحصائية «الأعمال داخل الشركة».
const (
	WorkOnSite  = "ON_SITE"  // عند الزبون
	WorkInHouse = "IN_HOUSE" // داخل الشركة (بالورشة)
)

var WorkLocationLabels = map[string]string{
	WorkOnSite:  "عند الزبون",
	WorkInHouse: "داخل الشركة",
}

func ValidWorkLocation(v string) bool {
	_, ok := WorkLocationLabels[v]
	return ok
}

// StopWorkRequest سبب توقف العمل — إجباري، لأن «توقف» بلا سبب ما تنفع
// لا للمتابعة ولا للتقرير.
type StopWorkRequest struct {
	Reason string `json:"reason"`
}
