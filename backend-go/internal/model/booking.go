package model

import "time"

type Booking struct {
	ID             string     `db:"id" json:"id"`
	Code           string     `db:"code" json:"code"`
	SequenceNumber *int       `db:"sequenceNumber" json:"sequenceNumber,omitempty"`
	ScheduledAt    *time.Time `db:"scheduledAt" json:"scheduledAt,omitempty"`
	// ScheduledEndAt نهاية المدى المتفق عليه مع الزبون. الوقت الواحد ما
	// يصير وعد يلتزم بيه (الطريق والشغل الي قبله ما ينحسبون بالدقيقة)،
	// فالمتفق عليه مدى ساعة: «نجيك بين ٧ و٨». ينحسب تلقائياً ساعة بعد
	// البداية بكل مكان ينتحدد بيه الموعد.
	// ⚠️ عمود بالجدول → لازم حقل هنا (SELECT *).
	ScheduledEndAt        *time.Time `db:"scheduledEndAt" json:"scheduledEndAt,omitempty"`
	PendingScheduledAt    *time.Time `db:"pendingScheduledAt" json:"pendingScheduledAt,omitempty"`
	CustomerID            string     `db:"customerId" json:"-"`
	ServiceID             *string    `db:"serviceId" json:"-"`
	TransferEmployeeID    *string    `db:"transferEmployeeId" json:"-"`
	ProjectSupervisorID   *string    `db:"projectSupervisorId" json:"-"`
	ExpenseResponsibleID  *string    `db:"expenseResponsibleId" json:"expenseResponsibleId,omitempty"`
	ConfirmedByEmployeeID *string    `db:"confirmedByEmployeeId" json:"-"`
	Notes                 *string    `db:"notes" json:"notes,omitempty"`
	VehicleType           *string    `db:"vehicleType" json:"vehicleType,omitempty"`
	Priority              string     `db:"priority" json:"priority"`
	Status                string     `db:"status" json:"status"`
	TransferToProjects    bool       `db:"transferToProjects" json:"transferToProjects"`
	// وقت وصول المشروع لمرحلة «٥. البدء بالتنفيذ» — هو الي يفتح الحجز
	// عند إداري الحجوزات حتى ينسّقه بكادر الشد. فارغ = لسه بالإجراءات.
	// ⚠️ عمود بالجدول → لازم حقل هنا (الجلب SELECT *).
	ProjectExecutionAt *time.Time `db:"projectExecutionAt" json:"projectExecutionAt,omitempty"`
	// ProjectLocked محسوب مو عمود: عند المشاريع وما وصل التنفيذ.
	// المنسّق يشوفه بس ما يكدر يلمسه.
	ProjectLocked          bool       `db:"-" json:"projectLocked"`
	ConfirmedByName        *string    `db:"confirmedByName" json:"confirmedByName,omitempty"`
	AdminNotes             *string    `db:"adminNotes" json:"adminNotes,omitempty"`
	AssignedVehicle        *string    `db:"assignedVehicle" json:"assignedVehicle,omitempty"`
	QuotedPrice            *float64   `db:"quotedPrice" json:"quotedPrice,omitempty"`
	Address                *string    `db:"address" json:"address,omitempty"`
	MapLocation            *string    `db:"mapLocation" json:"mapLocation,omitempty"`
	MapLatitude            *float64   `db:"mapLatitude" json:"mapLatitude,omitempty"`
	MapLongitude           *float64   `db:"mapLongitude" json:"mapLongitude,omitempty"`
	CompletedAt            *time.Time `db:"completedAt" json:"completedAt,omitempty"`
	CompletionNotes        *string    `db:"completionNotes" json:"completionNotes,omitempty"`
	AmountCollected        *float64   `db:"amountCollected" json:"amountCollected,omitempty"`
	AdvancePaid            *float64   `db:"advancePaid" json:"advancePaid,omitempty"`
	AmountVerified         bool       `db:"amountVerified" json:"amountVerified"`
	EquipmentStatus        string     `db:"equipmentStatus" json:"equipmentStatus"`
	Shift                  *string    `db:"shift" json:"shift,omitempty"`
	DeviceCount            *int       `db:"deviceCount" json:"deviceCount,omitempty"`
	InspectionSupervisorID *string    `db:"inspectionSupervisorId" json:"inspectionSupervisorId,omitempty"`
	ProjectCar             *string    `db:"projectCar" json:"projectCar,omitempty"`
	CrewNotes              *string    `db:"crewNotes" json:"crewNotes,omitempty"`
	BookingType            string     `db:"bookingType" json:"bookingType"`
	Urgency                *string    `db:"urgency" json:"urgency,omitempty"`
	MaintenanceType        *string    `db:"maintenanceType" json:"maintenanceType,omitempty"`
	RemembersExecutionCrew bool       `db:"remembersExecutionCrew" json:"remembersExecutionCrew"`
	SystemCount            *int       `db:"systemCount" json:"systemCount,omitempty"`
	SystemType             *string    `db:"systemType" json:"systemType,omitempty"`
	ProjectSpeed           *string    `db:"projectSpeed" json:"projectSpeed,omitempty"`
	WorkType               *string    `db:"workType" json:"workType,omitempty"`
	// WorkLocation وين انشتغل الشغل: عند الزبون لو داخل الشركة (بالورشة).
	// ⚠️ عمود بجدول Booking — لازم يضل إله حقل هنا لأن الاستعلام SELECT *.
	WorkLocation string `db:"workLocation" json:"workLocation"`
	// حجز داخل الشركة: الشغل لموظف من موظفينا مو لزبون خارجي.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	InternalEmployeeName  *string `db:"internalEmployeeName" json:"internalEmployeeName,omitempty"`
	InternalEmployeePhone *string `db:"internalEmployeePhone" json:"internalEmployeePhone,omitempty"`
	InternalDepartment    *string `db:"internalDepartment" json:"internalDepartment,omitempty"`
	// ⚠️ القسم صار سجلاً: الاسم يبقى منسوخاً نصاً (فوق) حتى الحجوزات
	// القديمة تبقى مقروءة، والمفتاح ينضاف حتى التقارير تجمّع صح.
	InternalDepartmentID *string `db:"internalDepartmentId" json:"internalDepartmentId,omitempty"`
	InternalHeadID       *string `db:"internalHeadId" json:"internalHeadId,omitempty"`
	// ملاحظات إداري الكوادر على الحجز الداخلي — منفصلة عن ملاحظات
	// الحجز العامة حتى ما يضيع منو كتب شنو.
	InternalHrNote *string `db:"internalHrNote" json:"internalHrNote,omitempty"`
	InternalApproved      *bool   `db:"internalApproved" json:"internalApproved,omitempty"`
	// حجز طاقة شمسية: المنظومة الي اتفق عليها المبيعات مع الزبون (اختيارية —
	// ممكن تتحدد بعد المعاينة)، واستهلاك الزبون الشهري لحساب السعة المناسبة.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	SolarSystemID   *string  `db:"solarSystemId" json:"solarSystemId,omitempty"`
	SolarMonthlyKwh *float64 `db:"solarMonthlyKwh" json:"solarMonthlyKwh,omitempty"`
	// توقف العمل: الليدر بدأ وما كدر يكمّل. الحجز يضل شغّال ويكدر
	// يكمّله بعدين، بس يبين «متوقف» بتنسيق الحجوزات مع سببه.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	WorkStoppedAt   *time.Time `db:"workStoppedAt" json:"workStoppedAt,omitempty"`
	WorkStopReason  *string    `db:"workStopReason" json:"workStopReason,omitempty"`
	WorkStoppedByID *string    `db:"workStoppedById" json:"workStoppedById,omitempty"`

	// ═══ الأرشفة ═══
	// الحجز المحذوف ما يروح: يختفي من الحجوزات ومن تنسيق الحجوزات ويضل
	// بالأرشيف بسبب حذفه ومنو حذفه. قبل كان DELETE حقيقي — الحجز وسببه
	// ومنو وافق على حذفه كلهم ينمحون سوه ولا يبقى أثر.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	ArchivedAt *time.Time `db:"archivedAt" json:"archivedAt,omitempty"`
	// الإنجاز الجزئي: كم مرة انأجّل الحجز لليوم الجاي وآخر مرة متى.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	PartialCount  int        `db:"partialCount" json:"partialCount"`
	LastPartialAt *time.Time `db:"lastPartialAt" json:"lastPartialAt,omitempty"`
	ArchivedByID  *string    `db:"archivedById" json:"archivedById,omitempty"`
	ArchiveReason *string    `db:"archiveReason" json:"archiveReason,omitempty"`

	// ═══ في الانتظار ═══
	// اتصلنا بالزبون بعد التثبيت حتى نطلعله وما رد. الحجز ينزاح من طابور
	// الشغل ويضل محفوظ لحد ما يرد. عدد المحاولات يفرّق بين زبون ما رد
	// مرة وزبون ما رد خمس مرات — الثاني قرار مو انتظار.
	WaitingSince         *time.Time `db:"waitingSince" json:"waitingSince,omitempty"`
	WaitingNote          *string    `db:"waitingNote" json:"waitingNote,omitempty"`
	WaitingByID          *string    `db:"waitingById" json:"waitingById,omitempty"`
	ContactAttempts      int        `db:"contactAttempts" json:"contactAttempts"`
	LastContactAttemptAt *time.Time `db:"lastContactAttemptAt" json:"lastContactAttemptAt,omitempty"`
	// تذكير المعاودة: آخر مرة ذكّرنا الإداري وكم مرة. للحد من الإزعاج
	// مو للتذكير نفسه — تذكير ينتجاهل أسوأ من ماكو تذكير.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	LastWaitingReminderAt *time.Time `db:"lastWaitingReminderAt" json:"lastWaitingReminderAt,omitempty"`
	WaitingReminderCount  int        `db:"waitingReminderCount" json:"waitingReminderCount"`

	// ═══ التأجيل ═══
	// الزبون يأجّل الموعد. الموعد إله سجل تغييرات أصلاً، بس ماكو فرق بين
	// «الإداري رتّب الجدول» و«الزبون أجّل» — وحجز تأجل أربع مرات علامة
	// على شي غلط لازم ينشاف.
	PostponeCount   int        `db:"postponeCount" json:"postponeCount"`
	LastPostponedAt *time.Time `db:"lastPostponedAt" json:"lastPostponedAt,omitempty"`
	PostponeReason  *string    `db:"postponeReason" json:"postponeReason,omitempty"`
	// انأجّل بلا موعد: الزبون ما محدّد متى يناسبه. الحجز ينزاح من جدول
	// اليوم ويروح لقائمة «الحجوزات المؤجلة» لحد ما ينحدد له موعد.
	// ⚠️ عمود بالجدول → لازم حقل هنا (الجلب SELECT *).
	AwaitingReschedule bool `db:"awaitingReschedule" json:"awaitingReschedule"`
	// ═══ تسوية إدارية لحجز قديم ═══
	// «تم الإنجاز بدون تفاصيل» — شغل صار قبل النظام وما نعرف كادره
	// ولا تكلفته. معلَّم حتى ينستثنى من الغرامات وما ينخلط بالمنجز
	// الحقيقي.
	SettledLegacyAt   *time.Time `db:"settledLegacyAt" json:"settledLegacyAt"`
	SettledLegacyByID *string    `db:"settledLegacyById" json:"settledLegacyById"`
	SettledLegacyNote *string    `db:"settledLegacyNote" json:"settledLegacyNote"`

	// ═══ تتبّع المراحل ═══
	// منو أدخل/رحّل الحجز — «بانتظار التثبيت» كانت تعرض حجوزات بلا
	// ما تقول لمنو ترجع لو المعلومة ناقصة.
	// ملاحظات موجّهة: وحدة للكادر المنفّذ ووحدة لمدير المشاريع —
	// منفصلات لأن الاثنين يقرون أشياء مختلفة.
	// والإلغاء بوقته حتى نفرّق «انلغى قبل التثبيت» عن «بعده».
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	CreatedByID      *string    `db:"createdById" json:"createdById,omitempty"`
	CrewNotesByID    *string    `db:"crewNotesById" json:"crewNotesById,omitempty"`
	CrewNotesAt      *time.Time `db:"crewNotesAt" json:"crewNotesAt,omitempty"`
	ProjectNotes     *string    `db:"projectNotes" json:"projectNotes,omitempty"`
	ProjectNotesByID *string    `db:"projectNotesById" json:"projectNotesById,omitempty"`
	ProjectNotesAt   *time.Time `db:"projectNotesAt" json:"projectNotesAt,omitempty"`
	CancelledAt      *time.Time `db:"cancelledAt" json:"cancelledAt,omitempty"`
	CancelledByID    *string    `db:"cancelledById" json:"cancelledById,omitempty"`
	CancelReason     *string    `db:"cancelReason" json:"cancelReason,omitempty"`

	// أسماء محسوبة وقت الجلب — مو أعمدة.
	CreatedByName      *string `db:"-" json:"createdByName,omitempty"`
	CrewNotesByName    *string `db:"-" json:"crewNotesByName,omitempty"`
	ProjectNotesByName *string `db:"-" json:"projectNotesByName,omitempty"`
	CancelledByName    *string `db:"-" json:"cancelledByName,omitempty"`

	// ═══ سلّة المرحلة ═══
	// «مؤجّل قبل التثبيت» غير «مؤجّل بعده»، ونفس الشي للملغى وللزبون
	// الي ما رد. محسوبة وقت الجلب من confirmedAt — مو عمود، حتى ما
	// يصير مصدرين للحقيقة ينفرزون عن بعض.
	StageBucket string `db:"-" json:"stageBucket,omitempty"`

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

	AddressDescription      *string    `db:"addressDescription" json:"addressDescription,omitempty"`
	CreatedAt               time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt               time.Time  `db:"updatedAt" json:"updatedAt"`
	MaterialsReadyAt        *time.Time `db:"materialsReadyAt" json:"materialsReadyAt,omitempty"`
	MaterialsReadyByID      *string    `db:"materialsReadyById" json:"-"`
	ResponseMinutes         *int       `db:"responseMinutes" json:"responseMinutes,omitempty"`
	ArrivedAt               *time.Time `db:"arrivedAt" json:"arrivedAt,omitempty"`
	StartedAt               *time.Time `db:"startedAt" json:"startedAt,omitempty"`
	ConfirmationContactedAt *time.Time `db:"confirmationContactedAt" json:"confirmationContactedAt,omitempty"`
	// وقت تحويل الحجز لتنسيق الحجوزات (التثبيت)
	ConfirmedAt               *time.Time     `db:"confirmedAt" json:"confirmedAt,omitempty"`
	LocationUrl               *string        `db:"locationUrl" json:"locationUrl,omitempty"`
	ConfirmationContactedByID *string        `db:"confirmationContactedById" json:"-"`
	ConfirmationContactedBy   *EmployeeBrief `db:"-" json:"confirmationContactedBy,omitempty"`
	LastEditedByID            *string        `db:"lastEditedById" json:"-"`
	LastEditedAt              *time.Time     `db:"lastEditedAt" json:"lastEditedAt,omitempty"`
	LastEditedBy              *EmployeeBrief `db:"-" json:"lastEditedBy,omitempty"`

	Customer *Customer `db:"-" json:"customer,omitempty"`
	// Service الخدمة الرئيسية (توافق مع الشاشات القديمة)، و Services كل
	// الخدمات المطلوبة بنفس الحجز — الزبون ممكن يطلب أكثر من منظومة سوة.
	Service             *Service            `db:"-" json:"service,omitempty"`
	Services            []Service           `db:"-" json:"services"`
	TransferEmployee    *Employee           `db:"-" json:"transferEmployee,omitempty"`
	ProjectSupervisor   *Employee           `db:"-" json:"projectSupervisor,omitempty"`
	ConfirmedByEmployee *Employee           `db:"-" json:"confirmedByEmployee,omitempty"`
	ExpenseResponsible  *Employee           `db:"-" json:"expenseResponsible,omitempty"`
	MaterialsReadyBy    *EmployeeBrief      `db:"-" json:"materialsReadyBy,omitempty"`
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
	// AssignedByID منو الإداري الي كلّف هذا الكادر — أساس المحاسبة لو
	// تأخر ورق الحجز.
	// ⚠️ عمود بالجدول → لازم حقل هنا، لأن الجلب يستعمل SELECT * وأي
	// عمود بلا حقل يفشل الاستعلام كله بالسكوت وترجع التعيينات فاضية.
	AssignedByID *string  `db:"assignedById" json:"-"`
	Employee     Employee `db:"-" json:"employee"`
}

type CartItem struct {
	ID          string    `db:"id" json:"id"`
	BookingID   string    `db:"bookingId" json:"bookingId"`
	ProductName string    `db:"productName" json:"productName"`
	Quantity    float64   `db:"quantity" json:"quantity"`
	UnitPrice   float64   `db:"unitPrice" json:"unitPrice"`
	TotalPrice  float64   `db:"totalPrice" json:"totalPrice"`
	Notes       *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `db:"updatedAt" json:"updatedAt"`
}

type ScheduleChangeLog struct {
	ID          string     `db:"id" json:"id"`
	BookingID   string     `db:"bookingId" json:"-"`
	ChangedByID string     `db:"changedById" json:"changedById"`
	OldTime     *time.Time `db:"oldTime" json:"oldTime,omitempty"`
	// NewTime فارغ يعني «انأجّل بلا موعد» — ما ينشال الحقل من السجل،
	// التأجيل بلا موعد حدث لازم ينتسجّل مثل غيره.
	NewTime *time.Time `db:"newTime" json:"newTime,omitempty"`
	// Kind يفرّق بين تغيير جدولة عادي (SCHEDULE) وتأجيل من الزبون
	// (POSTPONE)، وReason سبب التأجيل.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	Kind      string    `db:"kind" json:"kind"`
	Reason    *string   `db:"reason" json:"reason,omitempty"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
	ChangedBy *Employee `db:"-" json:"changedBy,omitempty"`
}

type CreateBookingRequest struct {
	CustomerID string  `json:"customerId"`
	ServiceID  *string `json:"serviceId,omitempty"`
	// ServiceIDs كل الخدمات المطلوبة بنفس الحجز (الأولى تنعتبر الرئيسية).
	// لو انرسلت فاضية ننزل على serviceId المفرد حتى ما ننكسر مع أي شاشة قديمة.
	ServiceIDs         []string `json:"serviceIds"`
	Notes              *string  `json:"notes,omitempty"`
	VehicleType        *string  `json:"vehicleType,omitempty"`
	Priority           *string  `json:"priority,omitempty"`
	TransferEmployeeID *string  `json:"transferEmployeeId,omitempty"`
	Address            *string  `json:"address,omitempty"`
	MapLatitude        *float64 `json:"mapLatitude,omitempty"`
	MapLongitude       *float64 `json:"mapLongitude,omitempty"`
	// رابط الموقع (كوكل ماب) — بديل عن التأشير على الخريطة، نفس فكرة الموردين
	LocationUrl *string `json:"locationUrl,omitempty"`
	// تفاصيل الأجهزة — إجبارية للخدمات المؤشّرة requiresDeviceInfo
	// (جي بي اس). عمودين موجودين أصلاً بجدول الحجز.
	DeviceCount *int `json:"deviceCount,omitempty"`

	// حجز داخل الشركة: نوع الحجز INTERNAL مع معلومات الموظف الطالب.
	// الخدمات والموقع يبقون مثل أي حجز.
	BookingType           *string  `json:"bookingType,omitempty"`
	SolarSystemID         *string  `json:"solarSystemId,omitempty"`
	SolarMonthlyKwh       *float64 `json:"solarMonthlyKwh,omitempty"`
	InternalEmployeeName  *string  `json:"internalEmployeeName,omitempty"`
	InternalEmployeePhone *string  `json:"internalEmployeePhone,omitempty"`
	InternalDepartment    *string  `json:"internalDepartment,omitempty"`
	InternalDepartmentID  *string  `json:"internalDepartmentId,omitempty"`
	InternalHeadID        *string  `json:"internalHeadId,omitempty"`
	InternalHrNote        *string  `json:"internalHrNote,omitempty"`
	InternalApproved      *bool    `json:"internalApproved,omitempty"`
}

type ConfirmBookingRequest struct {
	ConfirmedByName       *string  `json:"confirmedByName,omitempty"`
	ConfirmedByEmployeeID *string  `json:"confirmedByEmployeeId,omitempty"`
	AdminNotes            *string  `json:"adminNotes,omitempty"`
	TransferToProjects    bool     `json:"transferToProjects"`
	QuotedPrice           *float64 `json:"quotedPrice,omitempty"`
	Address               *string  `json:"address,omitempty"`
	ScheduledAt           *string  `json:"scheduledAt,omitempty"`
}

type AssignBookingRequest struct {
	EmployeeID      string  `json:"employeeId"`
	Role            string  `json:"role"`
	AssignedVehicle *string `json:"assignedVehicle,omitempty"`
}

type CreateCartItemRequest struct {
	ProductName string   `json:"productName"`
	Quantity    *float64 `json:"quantity,omitempty"`
	UnitPrice   *float64 `json:"unitPrice,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
}

type UpdateCartItemRequest struct {
	ProductName *string  `json:"productName,omitempty"`
	Quantity    *float64 `json:"quantity,omitempty"`
	UnitPrice   *float64 `json:"unitPrice,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
}

type UpdateBookingDetailsRequest struct {
	QuotedPrice          *float64 `json:"quotedPrice,omitempty"`
	Address              *string  `json:"address,omitempty"`
	AssignedVehicle      *string  `json:"assignedVehicle,omitempty"`
	MapLocation          *string  `json:"mapLocation,omitempty"`
	MapLatitude          *float64 `json:"mapLatitude,omitempty"`
	MapLongitude         *float64 `json:"mapLongitude,omitempty"`
	ExpenseResponsibleID *string  `json:"expenseResponsibleId,omitempty"`
	// رابط الموقع (بديل عن التحديد على الخريطة) — نفس فكرة الموردين
	LocationUrl *string `json:"locationUrl,omitempty"`
	// قائمة الخدمات المطلوبة بالحجز (لو انرسلت، تستبدل القائمة الحالية)
	ServiceIDs []string `json:"serviceIds"`
	// ملاحظات إداري الكوادر على الحجز الداخلي — تتعدّل بعد الإنشاء،
	// لأن الملاحظة تجي وقت ما تجي مو لحظة تسجيل الحجز.
	InternalHrNote *string `json:"internalHrNote,omitempty"`
}

type CompleteBookingRequest struct {
	CompletionNotes *string  `json:"completionNotes,omitempty"`
	AmountCollected *float64 `json:"amountCollected,omitempty"`
	AdvancePaid     *float64 `json:"advancePaid,omitempty"`
	// WorkLocation وين انجز الشغل — ينسأل وقت الإنجاز لأن هذا الوقت
	// الوحيد الي نعرف بيه الجواب أكيد.
	WorkLocation *string `json:"workLocation,omitempty"`
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

// ═══ سلال المراحل ═══
//
// صاحب العمل: «هاي تصير بيها حالتين — قبل التثبيت وبعد التثبيت».
// وهو محق: زبون ألغى **قبل** ما نثبتله موعد شي، وزبون ألغى **بعد**
// ما وعدناه وحضّرنا كادر شي ثاني تماماً. نفس الفرق بالتأجيل وبعدم
// الرد. دمجهن بسلّة وحدة يخبّي فرقاً بالمسؤولية وبالخسارة.
const (
	// ⚠️ ماكو «مؤجّل قبل التثبيت»: صاحب العمل — «ما عدي هيج شي، يعني
	// شلون أأجّل موعد وأني أصلاً ما محددله موعد؟». والتأجيل بالكود
	// نفسه ينقل من موعد قديم لموعد جديد، فالحجز بلا موعد ما يوصله.
	StageBucketPostponed       = "POSTPONED_AFTER_CONFIRM"
	StageBucketNoAnswerBefore  = "NO_ANSWER_BEFORE_CONFIRM"
	StageBucketNoAnswerAfter   = "NO_ANSWER_AFTER_CONFIRM"
	StageBucketCancelledBefore = "CANCELLED_BEFORE_CONFIRM"
	StageBucketCancelledAfter  = "CANCELLED_AFTER_CONFIRM"
)

// StageBucketLabel التسمية العربية — مصدر واحد للسيرفر والواجهة.
func StageBucketLabel(bucket string) string {
	switch bucket {
	case StageBucketPostponed:
		return "مؤجّلة"
	case StageBucketNoAnswerBefore:
		return "الزبون ما رد — قبل التثبيت"
	case StageBucketNoAnswerAfter:
		return "الزبون ما رد — بعد التثبيت"
	case StageBucketCancelledBefore:
		return "ملغى قبل التثبيت"
	case StageBucketCancelledAfter:
		return "ملغى بعد التثبيت"
	}
	return ""
}

// ComputeStageBucket يحدد سلّة الحجز — أو نص فاضي لو ما ينتمي لولا وحدة.
//
// ⚠️ الترتيب مقصود: الإلغاء يغلب التأجيل وعدم الرد. حجز الزبون ما رد
// بيه وبعدين ألغاه **ملغى** — الانتظار انتهى وصار قرار. لو عكسنا
// الترتيب چان الحجز الملغى ظل يطلع بقائمة «ما رد» وننتظر رداً من واحد
// خلاص انسحب.
//
// ⚠️ الفرز «قبل/بعد» يعتمد على confirmedAt مو على الحالة الحالية:
// الحالة تتغيّر (الملغى صار CANCELLED)، بس confirmedAt يبقى شاهد إنه
// كان مثبتاً يوم انلغى.
func (b *Booking) ComputeStageBucket() string {
	after := b.ConfirmedAt != nil
	switch {
	case b.Status == "CANCELLED":
		if after {
			return StageBucketCancelledAfter
		}
		return StageBucketCancelledBefore
	case b.WaitingSince != nil:
		if after {
			return StageBucketNoAnswerAfter
		}
		return StageBucketNoAnswerBefore
	case b.AwaitingReschedule:
		return StageBucketPostponed
	}
	return ""
}

// ═══ الطلعة ═══
//
// «كل مرة طلعناله تنحسب حجز للموظف، وكل مرة ينكتب بيها تاريخ وكادر
// طلع». الحجز ممكن ياخذ أربع طلعات بأربع كوادر مختلفة — والإنتاجية
// تنحسب من هذول مو من الحجز الواحد.
type BookingVisit struct {
	ID               string                   `db:"id" json:"id"`
	BookingID        string                   `db:"bookingId" json:"bookingId"`
	VisitNumber      int                      `db:"visitNumber" json:"visitNumber"`
	Outcome          string                   `db:"outcome" json:"outcome"`
	PercentDone      *int                     `db:"percentDone" json:"percentDone"`
	ProgressReportID *string                  `db:"progressReportId" json:"progressReportId"`
	ScheduledAt      *time.Time               `db:"scheduledAt" json:"scheduledAt"`
	OccurredAt       time.Time                `db:"occurredAt" json:"occurredAt"`
	CreatedAt        time.Time                `db:"createdAt" json:"createdAt"`
	Crew             []BookingVisitCrewMember `db:"-" json:"crew"`
}

// BookingVisitCrewMember منو طلع بهاي الطلعة — بالمعرّف مو بالاسم،
// حتى الإنتاجية تنعدّ بـJOIN مو بمطابقة نصوص تنكسر أول تشابه أسماء.
type BookingVisitCrewMember struct {
	EmployeeID string `db:"employeeId" json:"employeeId"`
	Name       string `db:"name" json:"name"`
	Role       string `db:"role" json:"role"`
	IsLeader   bool   `db:"isLeader" json:"isLeader"`
}
