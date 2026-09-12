package model

import "time"

type InventoryCheck struct {
	ID           string    `db:"id" json:"id"`
	EmployeeID   string    `db:"employeeId" json:"employeeId"`
	Complete     bool      `db:"complete" json:"complete"`
	MissingItems *string   `db:"missingItems" json:"missingItems"`
	CheckedAt    time.Time `db:"checkedAt" json:"checkedAt"`
	// الحجز الي انجرد قبله. فاضي بالجرود القديمة وبالجرد العام
	// الي ما ينربط بشغل معيّن.
	BookingID    *string    `db:"bookingId" json:"bookingId"`
	Resolved     bool       `db:"resolved" json:"resolved"`
	ResolvedByID *string    `db:"resolvedById" json:"resolvedById"`
	ResolvedAt   *time.Time `db:"resolvedAt" json:"resolvedAt"`

	Employee   *EmployeeBrief `db:"-" json:"employee"`
	ResolvedBy *EmployeeBrief `db:"-" json:"resolvedBy"`
}

type CreateInventoryCheckRequest struct {
	Complete     bool    `json:"complete"`
	MissingItems *string `json:"missingItems"`
	// اختياري: الجرد الي يصير قبل حجز معيّن ينربط بيه.
	BookingID *string `json:"bookingId"`
}

// ═══ حالة جرد كادر حجز واحد ═══
//
// الليدر ما يحتاج يسأل واحد واحد «جردت لو لا؟» — الشاشة تگله.
type BookingCrewInventoryState struct {
	EmployeeID string  `db:"employeeId" json:"employeeId"`
	Name       string  `db:"name" json:"name"`
	Position   *string `db:"position" json:"position"`
	IsLeader   bool    `db:"isLeader" json:"isLeader"`
	// فاضية = ما جرد لهذا الحجز بعد
	Checked      *time.Time `db:"checkedAt" json:"checkedAt"`
	Complete     *bool      `db:"complete" json:"complete"`
	MissingItems *string    `db:"missingItems" json:"missingItems"`
}

// BookingToolCheck لقطة (snapshot) للأدوات الشخصية الي كانت ناقصة عند الموظف
// باللحظة الي ضغط فيها "استلام" لحجز معيّن — بديل سريع عن جرد كامل منفصل،
// يسمح للإداري يشوف بالضبط شنو كان ناقص عند هذا الموظف بهذا الحجز بالذات.
type BookingToolCheck struct {
	ID           string    `db:"id" json:"id"`
	BookingID    string    `db:"bookingId" json:"bookingId"`
	EmployeeID   string    `db:"employeeId" json:"employeeId"`
	MissingItems *string   `db:"missingItems" json:"missingItems"`
	CheckedAt    time.Time `db:"checkedAt" json:"checkedAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

// AcceptBookingRequest جسم طلب اختياري لاستلام الحجز — لو الموظف عنده أدوات
// شخصية مسجّلة، الواجهة ترسل هون قائمة معرّفات الأدوات الي علّمها كناقصة (كل
// شي غير مذكور هون يُفترض إنه متوفر عنده، لأن الشيك يجي كله مؤشر افتراضياً).
type AcceptBookingRequest struct {
	MissingToolIDs []string `json:"missingToolIds"`
}

type PersonalTool struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	Name       string    `db:"name" json:"name"`
	Barcode    string    `db:"barcode" json:"barcode"`
	Status     string    `db:"status" json:"status"`
	CheckedOut bool      `db:"checkedOut" json:"checkedOut"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type CreatePersonalToolRequest struct {
	EmployeeID string `json:"employeeId"`
	Name       string `json:"name"`
	Barcode    string `json:"barcode"`
}

type UpdatePersonalToolRequest struct {
	Name       *string `json:"name"`
	Barcode    *string `json:"barcode"`
	Status     *string `json:"status"`
	CheckedOut *bool   `json:"checkedOut"`
	Note       *string `json:"note"`
}

// حالات الأداة الشخصية — "مفقودة" و"تالفة" هي الي تهم بسجل الحركة، لأن السؤال
// الأساسي هو "متى انفقدت هذي الأداة ومنو سجّلها".
const (
	PersonalToolStatusAvailable = "AVAILABLE"
	PersonalToolStatusLost      = "LOST"
	PersonalToolStatusDamaged   = "DAMAGED"
	PersonalToolStatusRepairing = "REPAIRING"
	PersonalToolStatusRetired   = "RETIRED"
	// موجودة أصلاً بالـenum من قبل — نبقيها حتى بيانات قديمة تنعرض صح
	PersonalToolStatusCheckedOut = "CHECKED_OUT"
)

var PersonalToolStatusLabels = map[string]string{
	PersonalToolStatusAvailable:  "موجودة",
	PersonalToolStatusLost:       "مفقودة",
	PersonalToolStatusDamaged:    "تالفة",
	PersonalToolStatusRepairing:  "بالتصليح",
	PersonalToolStatusRetired:    "خارج الخدمة",
	PersonalToolStatusCheckedOut: "مصروفة",
}

// أنواع أحداث سجل حركة الأداة
const (
	ToolEventCreated       = "CREATED"        // انضافت لعدة الموظف
	ToolEventStatusChanged = "STATUS_CHANGED" // تغيّرت حالتها (هنا يبين وقت الفقدان)
	ToolEventRenamed       = "RENAMED"        // انتعدّل اسمها أو باركودها
	ToolEventCheckedOut    = "CHECKED_OUT"    // انصرفت للموظف
	ToolEventReturned      = "RETURNED"       // انرجعت
	ToolEventDeleted       = "DELETED"        // انحذفت من العدة
)

var ToolEventLabels = map[string]string{
	ToolEventCreated:       "انضافت للعدة",
	ToolEventStatusChanged: "تغيّرت الحالة",
	ToolEventRenamed:       "انتعدّلت البيانات",
	ToolEventCheckedOut:    "انصرفت",
	ToolEventReturned:      "انرجعت",
	ToolEventDeleted:       "انحذفت من العدة",
}

// PersonalToolEvent سجل حركة الأداة الشخصية — كل تغيير ينكتب هنا حتى نقدر
// نجاوب "متى انفقدت هذي الأداة، ومنو سجّل الفقدان". السجل يبقى حتى بعد حذف
// الأداة (ما اكو قيد مفتاح خارجي)، لأن قيمته بالضبط إنه يوثّق الي راح.
type PersonalToolEvent struct {
	ID         string    `db:"id" json:"id"`
	ToolID     string    `db:"toolId" json:"toolId"`
	ToolName   string    `db:"toolName" json:"toolName"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	EventType  string    `db:"eventType" json:"eventType"`
	FromStatus *string   `db:"fromStatus" json:"fromStatus"`
	ToStatus   *string   `db:"toStatus" json:"toStatus"`
	Note       *string   `db:"note" json:"note"`
	ActorID    *string   `db:"actorId" json:"actorId"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`

	ActorName      *string `db:"actorName" json:"actorName"`
	EmployeeName   *string `db:"employeeName" json:"employeeName"`
	EventLabel     string  `db:"-" json:"eventLabel"`
	FromStatusText string  `db:"-" json:"fromStatusText"`
	ToStatusText   string  `db:"-" json:"toStatusText"`
}

type VehicleTool struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Barcode   *string   `db:"barcode" json:"barcode"`
	Quantity  int       `db:"quantity" json:"quantity"`
	VehicleID string    `db:"vehicleId" json:"vehicleId"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`

	// اسم/رقم السيارة، جايين بـJOIN حتى الجدول يعرض اسم مفهوم بدل معرّف.
	// كل استعلامات VehicleTool تمر بـvehicleToolSelect الي يجيب هذول العمودين.
	VehicleName  string `db:"vehicleName" json:"vehicleName"`
	VehiclePlate string `db:"vehiclePlate" json:"vehiclePlate"`
}

type CreateVehicleToolRequest struct {
	Name      string  `json:"name"`
	Barcode   *string `json:"barcode"`
	Quantity  *int    `json:"quantity"`
	VehicleID string  `json:"vehicleId"`
}

type UpdateVehicleToolRequest struct {
	Name      *string `json:"name"`
	Barcode   *string `json:"barcode"`
	Quantity  *int    `json:"quantity"`
	VehicleID *string `json:"vehicleId"`
	Status    *string `json:"status"`
}

type OnDemandTool struct {
	ID                string    `db:"id" json:"id"`
	Name              string    `db:"name" json:"name"`
	Barcode           string    `db:"barcode" json:"barcode"`
	TotalQuantity     int       `db:"totalQuantity" json:"totalQuantity"`
	AvailableQuantity int       `db:"availableQuantity" json:"availableQuantity"`
	Status            string    `db:"status" json:"status"`
	CreatedAt         time.Time `db:"createdAt" json:"createdAt"`
}

type CreateOnDemandToolRequest struct {
	Name              string `json:"name"`
	Barcode           string `json:"barcode"`
	TotalQuantity     int    `json:"totalQuantity"`
	AvailableQuantity int    `json:"availableQuantity"`
}

type UpdateOnDemandToolRequest struct {
	Name              *string `json:"name"`
	Barcode           *string `json:"barcode"`
	TotalQuantity     *int    `json:"totalQuantity"`
	AvailableQuantity *int    `json:"availableQuantity"`
	Status            *string `json:"status"`
}

// أسباب طلب أداة — الموظف لازم يختار وحد منها حتى إداري الكميات يعرف ليش
// ينطلب، ويقدر يقرأ الشرح ويقرر موافقة أو رفض على أساس واضح بدل طلب أعمى.
const (
	ToolRequestReasonDamaged   = "DAMAGED"   // الأداة الي عنده تالفة
	ToolRequestReasonLost      = "LOST"      // الأداة الي عنده ضايعة
	ToolRequestReasonWorn      = "WORN"      // مستهلكة من كثر الاستعمال
	ToolRequestReasonStolen    = "STOLEN"    // مسروقة
	ToolRequestReasonNeverHad  = "NEVER_HAD" // ما عنده الأداة أصلاً
	ToolRequestReasonExtraNeed = "EXTRA"     // يحتاج نسخة إضافية لطبيعة الشغل
	ToolRequestReasonOther     = "OTHER"     // سبب ثاني (يشرحه بالوصف)
)

var ToolRequestReasonLabels = map[string]string{
	ToolRequestReasonDamaged:   "الأداة الي عندي تالفة",
	ToolRequestReasonLost:      "الأداة الي عندي ضايعة",
	ToolRequestReasonWorn:      "الأداة مستهلكة من كثر الاستعمال",
	ToolRequestReasonStolen:    "الأداة مسروقة",
	ToolRequestReasonNeverHad:  "ما عندي هذي الأداة أصلاً",
	ToolRequestReasonExtraNeed: "أحتاج نسخة إضافية لطبيعة الشغل",
	ToolRequestReasonOther:     "سبب آخر",
}

func IsValidToolRequestReason(reason string) bool {
	_, ok := ToolRequestReasonLabels[reason]
	return ok
}

// تصنيف طلب الأداة — إداري الكميات يشوف الطلبات بثلاث سلال منفصلة.
const (
	ToolRequestKindSpecialized    = "SPECIALIZED"     // أداة تخصصية من أدوات حسب الحاجة
	ToolRequestKindReplaceLost    = "REPLACE_LOST"    // بدل مفقود
	ToolRequestKindReplaceDamaged = "REPLACE_DAMAGED" // بدل تالف
)

var ToolRequestKindLabels = map[string]string{
	ToolRequestKindSpecialized:    "أداة تخصصية",
	ToolRequestKindReplaceLost:    "بدل مفقود",
	ToolRequestKindReplaceDamaged: "بدل تالف",
}

// KindForReason يستنتج التصنيف من سبب الطلب — حتى لو الواجهة ما بعثت تصنيف.
func KindForReason(reason string) string {
	switch reason {
	case ToolRequestReasonLost, ToolRequestReasonStolen:
		return ToolRequestKindReplaceLost
	case ToolRequestReasonDamaged, ToolRequestReasonWorn:
		return ToolRequestKindReplaceDamaged
	default:
		return ToolRequestKindSpecialized
	}
}

// StockIntake إضافة كمية للمخزون — من يجي الفني يطلب أداة، الإداري ينطيه
// من النظام والكمية تنقص. لازم يكون اكو طريقة يزيد بيها الكمية بأثر واضح.
type StockIntake struct {
	ID          string    `db:"id" json:"id"`
	ToolID      string    `db:"toolId" json:"toolId"`
	Quantity    int       `db:"quantity" json:"quantity"`
	UnitPrice   *float64  `db:"unitPrice" json:"unitPrice"`
	Supplier    *string   `db:"supplier" json:"supplier"`
	Notes       *string   `db:"notes" json:"notes"`
	CreatedByID *string   `db:"createdById" json:"createdById"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`

	ToolName    string  `db:"toolName" json:"toolName"`
	CreatedName *string `db:"createdName" json:"createdName"`
}

type CreateStockIntakeRequest struct {
	ToolID    string   `json:"toolId"`
	Quantity  int      `json:"quantity"`
	UnitPrice *float64 `json:"unitPrice"`
	Supplier  *string  `json:"supplier"`
	Notes     *string  `json:"notes"`
}

type ToolRequest struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	ToolID       string     `db:"toolId" json:"toolId"`
	Status       string     `db:"status" json:"status"`
	Reason       *string    `db:"reason" json:"reason"`
	RequestKind  *string    `db:"requestKind" json:"requestKind"`
	Description  *string    `db:"description" json:"description"`
	ApprovedByID *string    `db:"approvedById" json:"approvedById"`
	RequestedAt  time.Time  `db:"requestedAt" json:"requestedAt"`
	ApprovedAt   *time.Time `db:"approvedAt" json:"approvedAt"`
	ReturnedAt   *time.Time `db:"returnedAt" json:"returnedAt"`
	// إذا الأداة ما كانت متوفرة بالشركة وقت الموافقة، إداري الكميات لازم يدخل
	// سعر الشراء، ويتفتح طلب مشتريات تلقائياً يوصل للمحاسب — هذول الحقلين
	// يربطون طلب الأداة بطلب المشتريات المتولّد منه.
	PurchasePrice        *float64 `db:"purchasePrice" json:"purchasePrice"`
	ProcurementRequestID *string  `db:"procurementRequestId" json:"procurementRequestId"`

	Employee   *EmployeeBrief `db:"-" json:"employee"`
	Tool       *OnDemandTool  `db:"-" json:"tool"`
	ApprovedBy *EmployeeBrief `db:"-" json:"approvedBy"`
	// ReasonLabel نص السبب بالعربي، محسوب بالسيرفر حتى الواجهة ما تعيد
	// تعريف نفس الخريطة وتنحرف عنها.
	ReasonLabel string `db:"-" json:"reasonLabel"`
	KindLabel   string `db:"-" json:"kindLabel"`
}

type CreateToolRequestRequest struct {
	EmployeeID  string  `json:"employeeId"`
	ToolID      string  `json:"toolId"`
	Reason      string  `json:"reason"`
	RequestKind *string `json:"requestKind"`
	Description *string `json:"description"`
}

type ApproveToolRequestRequest struct {
	ApprovedByID string `json:"approvedById"`
	// يُطلب فقط لما الأداة مو متوفرة بالمخزن — بغير هذي الحالة ينهمل.
	PurchasePrice *float64 `json:"purchasePrice"`
}

// PersonalToolTemplateItem هو "عدة قياسية" — قائمة رئيسية بأسماء الأدوات الشخصية
// الي المفروض كل موظف يكون عنده إياها. أي عنصر جديد ينضاف هون يتطبق فوراً على
// كل الموظفين الحاليين (ينشئ PersonalTool لكل واحد منهم)، وأي موظف جديد ينضاف
// بعدين ياخذ القائمة كاملة تلقائياً وقت إنشاء حسابه (انظر employee_service.go).
type PersonalToolTemplateItem struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type CreatePersonalToolTemplateItemRequest struct {
	Name string `json:"name"`
}

// ═══ استثناء أداة من عدة موظف بعينه ═══
//
// «الناقص» يتحسب = أسماء العدة القياسية ناقص الي يملكه الموظف. فحذف
// الأداة من عدته **هو الي يخلقها نقصاً** — اسمها لسه بالقالب. أبو
// الكميات يريدها تنشال من نواقص **هذا الموظف** بلا ما تنشال من
// القالب المشترك (وإلا ماكو ولا فني يتحاسب عليها).
type PersonalToolExemption struct {
	ID           string    `db:"id" json:"id"`
	EmployeeID   string    `db:"employeeId" json:"employeeId"`
	ToolName     string    `db:"toolName" json:"toolName"`
	Note         *string   `db:"note" json:"note"`
	ByEmployeeID *string   `db:"byEmployeeId" json:"byEmployeeId"`
	ByName       string    `db:"byName" json:"byName"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

type CreatePersonalToolExemptionRequest struct {
	EmployeeID string  `json:"employeeId"`
	ToolName   string  `json:"toolName"`
	Note       *string `json:"note"`
}

// VehicleToolCheck لقطة تسجّل الأدوات العامة الناقصة بمركبة معينة عند لحظة
// بدء مهمة من قبل ليدر — نفس فكرة BookingToolCheck بس لأدوات المركبة العامة
// وتُطلب فقط لما الموظف الي بادر المهمة يكون ليدر (isLeader=true فريش من قاعدة
// البيانات)، الموظف العادي يقدر يبدأ مهمة بدون هالخطوة إطلاقاً.
type VehicleToolCheck struct {
	ID               string    `db:"id" json:"id"`
	VehicleID        string    `db:"vehicleId" json:"vehicleId"`
	MissionID        string    `db:"missionId" json:"missionId"`
	EmployeeID       string    `db:"employeeId" json:"employeeId"`
	MissingToolNames *string   `db:"missingToolNames" json:"missingToolNames"`
	CreatedAt        time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type CreateVehicleToolCheckRequest struct {
	MissingToolNames []string `json:"missingToolNames"`
}
