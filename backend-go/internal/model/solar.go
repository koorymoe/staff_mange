package model

import "time"

// ═══ نظام الطاقة الشمسية ═══
// منقول من نظام Solar Expert الي جان على Google Sheets.

// تصنيفات مكوّنات المنظومة. النظام القديم جان يحددها بطريقة ملتوية —
// خمس أعمدة أسماء والفارغ منهن يحدد التصنيف — وهنا صارت قيمة صريحة.
const (
	SolarPanel    = "PANEL"
	SolarInverter = "INVERTER"
	SolarBattery  = "BATTERY"
	SolarBoard    = "BOARD"
	SolarIron     = "IRON"
)

// SolarFollowUpDays مدة متابعة الزبون بعد التركيب. النظام القديم جان
// يحسبها بالواجهة بكل رسمة؛ هنا رقم واحد، والتاريخ ينخزن وقت التجهيز حتى
// تغييره بكرة ما ينقلب أثراً رجعياً على تركيبات قديمة.
const SolarFollowUpDays = 30

// SolarComponent مادة بالمخزن تدخل بمنظومة (لوح، إنفيرتر، بطارية، بورد،
// حديد) — بسعرها وكميتها وحدها الأدنى ومواصفاتها الفنية.
type SolarComponent struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Category  string    `db:"category" json:"category"`
	Quantity  int       `db:"quantity" json:"quantity"`
	Price     float64   `db:"price" json:"price"`
	MinStock  int       `db:"minStock" json:"minStock"`
	Specs     JSONMap   `db:"specs" json:"specs"`
	Notes     *string   `db:"notes" json:"notes"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `db:"updatedAt" json:"updatedAt"`
}

// SolarWiringLine سطر تسليك: نوع الكابل وطوله بالمتر وسعر المتر.
type SolarWiringLine struct {
	Type   string  `json:"type"`
	Length float64 `json:"length"`
	Price  float64 `json:"price"`
}

// SolarIronLine سطر حدادة: النوع والعدد والوحدة وسعر الوحدة.
type SolarIronLine struct {
	Type  string  `json:"type"`
	Qty   float64 `json:"qty"`
	Unit  string  `json:"unit"`
	Price float64 `json:"price"`
}

// SolarSystem منظومة جاهزة بالكتالوك.
type SolarSystem struct {
	ID       string `db:"id" json:"id"`
	Brand    string `db:"brand" json:"brand"`
	Model    string `db:"model" json:"model"`
	Capacity string `db:"capacity" json:"capacity"`

	PanelID     *string `db:"panelId" json:"panelId"`
	PanelQty    int     `db:"panelQty" json:"panelQty"`
	InverterID  *string `db:"inverterId" json:"inverterId"`
	InverterQty int     `db:"inverterQty" json:"inverterQty"`
	BatteryID   *string `db:"batteryId" json:"batteryId"`
	BatteryQty  int     `db:"batteryQty" json:"batteryQty"`
	BoardID     *string `db:"boardId" json:"boardId"`

	WiringDetails   JSONRaw `db:"wiringDetails" json:"wiringDetails"`
	WiringTotalCost float64 `db:"wiringTotalCost" json:"wiringTotalCost"`
	IronDetails     JSONRaw `db:"ironDetails" json:"ironDetails"`
	IronTotalCost   float64 `db:"ironTotalCost" json:"ironTotalCost"`

	InstallPrice  float64 `db:"installPrice" json:"installPrice"`
	ProgramPrice  float64 `db:"programPrice" json:"programPrice"`
	WarrantyPrice float64 `db:"warrantyPrice" json:"warrantyPrice"`

	Notes       *string   `db:"notes" json:"notes"`
	CreatedByID *string   `db:"createdById" json:"-"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `db:"updatedAt" json:"updatedAt"`

	// محسوبة وقت الجلب — مو أعمدة بالجدول
	Panel     *SolarComponent `db:"-" json:"panel"`
	Inverter  *SolarComponent `db:"-" json:"inverter"`
	Battery   *SolarComponent `db:"-" json:"battery"`
	Board     *SolarComponent `db:"-" json:"board"`
	Price     SolarPrice      `db:"-" json:"price"`
	Shortages []SolarShortage `db:"-" json:"shortages"`
}

// SolarPrice تفصيل سعر المنظومة. ينحسب من أسعار المخزن الحالية وقت
// العرض — سعر اللوح يتغير، والكتالوك لازم يعكس السعر اليوم مو سعر يوم
// ما انكتبت المنظومة.
type SolarPrice struct {
	Panels     float64 `json:"panels"`
	Inverters  float64 `json:"inverters"`
	Batteries  float64 `json:"batteries"`
	Board      float64 `json:"board"`
	Wiring     float64 `json:"wiring"`
	Iron       float64 `json:"iron"`
	Install    float64 `json:"install"`
	Program    float64 `json:"program"`
	Warranty   float64 `json:"warranty"`
	Components float64 `json:"components"`
	Total      float64 `json:"total"`
}

// SolarShortage مكوّن ما يكفي بالمخزن لتجهيز المنظومة.
type SolarShortage struct {
	ComponentID string `json:"componentId"`
	Name        string `json:"name"`
	Required    int    `json:"required"`
	Available   int    `json:"available"`
	Missing     bool   `json:"missing"` // مو موجود بالمخزن أصلاً
}

// SolarInstallation منظومة انجهزت لزبون: انخصمت مكوّناتها من المخزن،
// والزبون دخل دورة متابعة بعد ٣٠ يوم.
type SolarInstallation struct {
	ID            string     `db:"id" json:"id"`
	SystemID      string     `db:"systemId" json:"systemId"`
	CustomerID    string     `db:"customerId" json:"customerId"`
	InstallDate   time.Time  `db:"installDate" json:"installDate"`
	FollowUpAt    time.Time  `db:"followUpAt" json:"followUpAt"`
	ContactedAt   *time.Time `db:"contactedAt" json:"contactedAt"`
	ContactedByID *string    `db:"contactedById" json:"-"`
	ContactNotes  *string    `db:"contactNotes" json:"contactNotes"`
	Status        string     `db:"status" json:"status"`

	TotalPrice     float64 `db:"totalPrice" json:"totalPrice"`
	PriceBreakdown JSONRaw `db:"priceBreakdown" json:"priceBreakdown"`

	Notes       *string   `db:"notes" json:"notes"`
	CreatedByID *string   `db:"createdById" json:"-"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`

	// محسوبة وقت الجلب
	Customer        *Customer    `db:"-" json:"customer"`
	System          *SolarSystem `db:"-" json:"system"`
	ContactedByName *string      `db:"-" json:"contactedByName"`
	DueForFollowUp  bool         `db:"-" json:"dueForFollowUp"`
	DaysOverdue     int          `db:"-" json:"daysOverdue"`
}

// ═══ الطلبات ═══

type SaveSolarComponentRequest struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	MinStock int     `json:"minStock"`
	Specs    JSONMap `json:"specs"`
	Notes    *string `json:"notes"`
}

type SaveSolarSystemRequest struct {
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Capacity string `json:"capacity"`

	PanelID     *string `json:"panelId"`
	PanelQty    int     `json:"panelQty"`
	InverterID  *string `json:"inverterId"`
	InverterQty int     `json:"inverterQty"`
	BatteryID   *string `json:"batteryId"`
	BatteryQty  int     `json:"batteryQty"`
	BoardID     *string `json:"boardId"`

	WiringDetails []SolarWiringLine `json:"wiringDetails"`
	IronDetails   []SolarIronLine   `json:"ironDetails"`

	InstallPrice  float64 `json:"installPrice"`
	ProgramPrice  float64 `json:"programPrice"`
	WarrantyPrice float64 `json:"warrantyPrice"`

	Notes *string `json:"notes"`
}

// ProcessSolarSystemRequest تجهيز منظومة لزبون.
//
// CustomerID اختياري: لو الزبون موجود عدنا نربط بيه، وإلا نسوي زبون
// جديد من الاسم والهاتف والعنوان — لأن زبون الطاقة الشمسية هو نفسه زبون
// الشركة، مو دفتر زبائن ثاني.
type ProcessSolarSystemRequest struct {
	CustomerID      *string `json:"customerId"`
	CustomerName    string  `json:"customerName"`
	CustomerPhone   string  `json:"customerPhone"`
	CustomerAddress string  `json:"customerAddress"`
	InstallDate     string  `json:"installDate"`
	Notes           *string `json:"notes"`
}

// SolarStats أرقام لوحة الطاقة الشمسية — كلها تنحسب بقاعدة البيانات.
type SolarStats struct {
	SystemCount        int     `db:"systemCount" json:"systemCount"`
	ComponentCount     int     `db:"componentCount" json:"componentCount"`
	InventoryValue     float64 `db:"inventoryValue" json:"inventoryValue"`
	LowStockCount      int     `db:"lowStockCount" json:"lowStockCount"`
	OutOfStockCount    int     `db:"outOfStockCount" json:"outOfStockCount"`
	ProcessedCount     int     `db:"processedCount" json:"processedCount"`
	CustomerCount      int     `db:"customerCount" json:"customerCount"`
	DueFollowUpCount   int     `db:"dueFollowUpCount" json:"dueFollowUpCount"`
	ContactedCount     int     `db:"contactedCount" json:"contactedCount"`
	InstalledThisMonth int     `db:"installedThisMonth" json:"installedThisMonth"`
	TotalWiring        float64 `db:"totalWiring" json:"totalWiring"`
	TotalIron          float64 `db:"totalIron" json:"totalIron"`
	TotalInstall       float64 `db:"totalInstall" json:"totalInstall"`
	TotalProgram       float64 `db:"totalProgram" json:"totalProgram"`
}

// ═══ برامج التدريب ═══
// منقولة من نظام الطاقة الشمسية، بس على موظفينا ومهاراتنا الموجودة.

type TrainingProgram struct {
	ID               string     `db:"id" json:"id"`
	Name             string     `db:"name" json:"name"`
	Level            string     `db:"level" json:"level"`
	DurationDays     int        `db:"durationDays" json:"durationDays"`
	StartDate        *time.Time `db:"startDate" json:"startDate"`
	EndDate          *time.Time `db:"endDate" json:"endDate"`
	TargetDepartment *string    `db:"targetDepartment" json:"targetDepartment"`
	InstructorID     *string    `db:"instructorId" json:"instructorId"`
	Objectives       *string    `db:"objectives" json:"objectives"`
	Content          *string    `db:"content" json:"content"`
	PassRate         int        `db:"passRate" json:"passRate"`
	Cost             float64    `db:"cost" json:"cost"`
	Status           string     `db:"status" json:"status"`
	Progress         int        `db:"progress" json:"progress"`
	CreatedByID      *string    `db:"createdById" json:"-"`
	CreatedAt        time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time  `db:"updatedAt" json:"updatedAt"`

	// محسوبة وقت الجلب
	InstructorName *string               `db:"-" json:"instructorName"`
	Participants   []TrainingParticipant `db:"-" json:"participants"`
	Skills         []TrainingSkill       `db:"-" json:"skills"`
}

type TrainingParticipant struct {
	EmployeeID string  `json:"employeeId"`
	Name       string  `json:"name"`
	Department *string `json:"department"`
	JobTitle   *string `json:"jobTitle"`
	Passed     *bool   `json:"passed"`
}

type TrainingSkill struct {
	SkillID  string `json:"skillId"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type SaveTrainingProgramRequest struct {
	Name             string   `json:"name"`
	Level            string   `json:"level"`
	DurationDays     int      `json:"durationDays"`
	StartDate        string   `json:"startDate"`
	TargetDepartment string   `json:"targetDepartment"`
	InstructorID     string   `json:"instructorId"`
	Objectives       string   `json:"objectives"`
	Content          string   `json:"content"`
	PassRate         int      `json:"passRate"`
	Cost             float64  `json:"cost"`
	Status           string   `json:"status"`
	Progress         int      `json:"progress"`
	ParticipantIDs   []string `json:"participantIds"`
	SkillIDs         []string `json:"skillIds"`
}
