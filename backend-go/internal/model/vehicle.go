package model

import "time"

type Vehicle struct {
	ID              string    `db:"id" json:"id"`
	Name            string    `db:"name" json:"name"`
	PlateNumber     string    `db:"plateNumber" json:"plateNumber"`
	Color           *string   `db:"color" json:"color"`
	Type            *string   `db:"type" json:"type"`
	Model           *string   `db:"model" json:"model"`
	Year            *int      `db:"year" json:"year"`
	ChassisNumber   *string   `db:"chassisNumber" json:"chassisNumber"`
	EngineNumber    *string   `db:"engineNumber" json:"engineNumber"`
	FuelType        *string   `db:"fuelType" json:"fuelType"`
	CurrentOdometer int       `db:"currentOdometer" json:"currentOdometer"`
	Condition       *string   `db:"condition" json:"condition"`
	IsActive        bool      `db:"isActive" json:"isActive"`
	CreatedAt       time.Time `db:"createdAt" json:"createdAt"`
}

type CreateVehicleRequest struct {
	Name        string  `json:"name"`
	PlateNumber string  `json:"plateNumber"`
	Color       *string `json:"color"`
	Type        *string `json:"type"`
}

type UpdateVehicleRequest struct {
	Name            *string `json:"name"`
	PlateNumber     *string `json:"plateNumber"`
	Color           *string `json:"color"`
	Type            *string `json:"type"`
	Model           *string `json:"model"`
	Year            *int    `json:"year"`
	ChassisNumber   *string `json:"chassisNumber"`
	EngineNumber    *string `json:"engineNumber"`
	FuelType        *string `json:"fuelType"`
	CurrentOdometer *int    `json:"currentOdometer"`
	Condition       *string `json:"condition"`
	IsActive        *bool   `json:"isActive"`
}

// VehicleDocument وثائق السيارة: تأمين، إجازة سنوية، فحص دوري... إلخ.
type VehicleDocument struct {
	ID             string     `db:"id" json:"id"`
	VehicleID      string     `db:"vehicleId" json:"vehicleId"`
	DocumentType   string     `db:"documentType" json:"documentType"` // INSURANCE | ANNUAL_LICENSE | INSPECTION | OTHER
	DocumentNumber *string    `db:"documentNumber" json:"documentNumber"`
	IssueDate      *time.Time `db:"issueDate" json:"issueDate"`
	ExpiryDate     *time.Time `db:"expiryDate" json:"expiryDate"`
	FileURL        *string    `db:"fileUrl" json:"fileUrl"`
	Notes          *string    `db:"notes" json:"notes"`
	CreatedAt      time.Time  `db:"createdAt" json:"createdAt"`
}

type CreateVehicleDocumentRequest struct {
	DocumentType   string  `json:"documentType"`
	DocumentNumber *string `json:"documentNumber"`
	IssueDate      *string `json:"issueDate"`
	ExpiryDate     *string `json:"expiryDate"`
	FileURL        *string `json:"fileUrl"`
	Notes          *string `json:"notes"`
}

// VehiclePhoto صور السيارة (معرض عام).
type VehiclePhoto struct {
	ID        string    `db:"id" json:"id"`
	VehicleID string    `db:"vehicleId" json:"vehicleId"`
	URL       string    `db:"url" json:"url"`
	Caption   *string   `db:"caption" json:"caption"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type CreateVehiclePhotoRequest struct {
	URL     string  `json:"url"`
	Caption *string `json:"caption"`
}

// VehicleMission سجل مهمة/خروج سيارة: سبب، وجهة، عداد بداية/نهاية، مسافة محسوبة.
type VehicleMission struct {
	ID            string     `db:"id" json:"id"`
	VehicleID     string     `db:"vehicleId" json:"vehicleId"`
	DriverID      string     `db:"driverId" json:"driverId"`
	Purpose       string     `db:"purpose" json:"purpose"`
	Destination   string     `db:"destination" json:"destination"`
	StartedAt     time.Time  `db:"startedAt" json:"startedAt"`
	EndedAt       *time.Time `db:"endedAt" json:"endedAt"`
	StartOdometer int        `db:"startOdometer" json:"startOdometer"`
	EndOdometer   *int       `db:"endOdometer" json:"endOdometer"`
	DistanceKm    *int       `db:"distanceKm" json:"distanceKm"`
	Notes         *string    `db:"notes" json:"notes"`
	Status        string     `db:"status" json:"status"` // IN_PROGRESS | COMPLETED
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`

	Vehicle    *Vehicle                  `db:"-" json:"vehicle,omitempty"`
	Driver     *EmployeeBrief            `db:"-" json:"driver,omitempty"`
	Passengers []VehicleMissionPassenger `db:"-" json:"passengers,omitempty"`
}

// VehicleMissionPassenger الموظفين المرافقين بالمهمة (يُعلَنون عند بدء المهمة).
type VehicleMissionPassenger struct {
	ID         string `db:"id" json:"id"`
	MissionID  string `db:"missionId" json:"missionId"`
	EmployeeID string `db:"employeeId" json:"employeeId"`

	Employee *EmployeeBrief `db:"-" json:"employee,omitempty"`
}

type StartVehicleMissionRequest struct {
	VehicleID     string   `json:"vehicleId"`
	DriverID      *string  `json:"driverId"` // لو فاضي = الموظف الحالي نفسه
	Purpose       string   `json:"purpose"`
	Destination   string   `json:"destination"`
	StartOdometer int      `json:"startOdometer"`
	PassengerIDs  []string `json:"passengerIds"`
}

type EndVehicleMissionRequest struct {
	EndOdometer int     `json:"endOdometer"`
	Notes       *string `json:"notes"`
}

type VehicleMissionFilters struct {
	VehicleID *string
	DriverID  *string
	Status    *string
	From      *string
	To        *string
}

// VehicleLog يغطي وقود/تنظيف/تبديل زيت/صيانة عامة — سجل واحد بنوع محدد لكل حدث
type VehicleLog struct {
	ID              string         `db:"id" json:"id"`
	VehicleID       string         `db:"vehicleId" json:"vehicleId"`
	Type            string         `db:"type" json:"type"` // FUEL | CLEANING | OIL_CHANGE | MAINTENANCE
	PerformedAt     time.Time      `db:"performedAt" json:"performedAt"`
	NextDueAt       *time.Time     `db:"nextDueAt" json:"nextDueAt"`
	NextDueOdometer *int           `db:"nextDueOdometer" json:"nextDueOdometer"`
	Odometer        *int           `db:"odometer" json:"odometer"`
	Cost            *float64       `db:"cost" json:"cost"`
	Notes           *string        `db:"notes" json:"notes"`
	RecordedByID    *string        `db:"recordedById" json:"-"`
	CreatedAt       time.Time      `db:"createdAt" json:"createdAt"`
	RecordedBy      *EmployeeBrief `db:"-" json:"recordedBy"`
}

type CreateVehicleLogRequest struct {
	Type            string   `json:"type"`
	PerformedAt     *string  `json:"performedAt"`
	NextDueAt       *string  `json:"nextDueAt"`
	NextDueOdometer *int     `json:"nextDueOdometer"`
	Odometer        *int     `json:"odometer"`
	Cost            *float64 `json:"cost"`
	Notes           *string  `json:"notes"`
}

// VehicleIncident يغطي الأعطال والأضرار (صدمات) مع تحديد المسبب والتكلفة
type VehicleIncident struct {
	ID                    string     `db:"id" json:"id"`
	VehicleID             string     `db:"vehicleId" json:"vehicleId"`
	Type                  string     `db:"type" json:"type"` // FAULT | DAMAGE
	Description           string     `db:"description" json:"description"`
	ResponsibleEmployeeID *string    `db:"responsibleEmployeeId" json:"-"`
	Cost                  *float64   `db:"cost" json:"cost"`
	Status                string     `db:"status" json:"status"` // OPEN | RESOLVED
	ReportedByID          *string    `db:"reportedById" json:"-"`
	CreatedAt             time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt            *time.Time `db:"resolvedAt" json:"resolvedAt"`

	ResponsibleEmployee *EmployeeBrief `db:"-" json:"responsibleEmployee"`
	ReportedBy          *EmployeeBrief `db:"-" json:"reportedBy"`
}

type CreateVehicleIncidentRequest struct {
	Type                  string   `json:"type"`
	Description           string   `json:"description"`
	ResponsibleEmployeeID *string  `json:"responsibleEmployeeId"`
	Cost                  *float64 `json:"cost"`
}

type UpdateVehicleIncidentRequest struct {
	Status *string  `json:"status"`
	Cost   *float64 `json:"cost"`
}

// VehicleIncidentAttachment صور/فيديو مرفقة بعطل أو ضرر (تُخزَّن base64 مثل الوثائق والصور).
type VehicleIncidentAttachment struct {
	ID         string    `db:"id" json:"id"`
	IncidentID string    `db:"incidentId" json:"incidentId"`
	URL        string    `db:"url" json:"url"`
	MediaType  string    `db:"mediaType" json:"mediaType"` // IMAGE | VIDEO
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

type CreateVehicleIncidentAttachmentRequest struct {
	URL       string `json:"url"`
	MediaType string `json:"mediaType"`
}

// VehiclePart قطعة استهلاك (إطار/بطارية) مع تاريخ وعداد التركيب وعمر متوقع
// بالمسافة و/أو الزمن — أيهما أقرب يحدد الاستحقاق.
type VehiclePart struct {
	ID                     string     `db:"id" json:"id"`
	VehicleID              string     `db:"vehicleId" json:"vehicleId"`
	PartType               string     `db:"partType" json:"partType"` // TIRE | BATTERY
	InstalledAt            time.Time  `db:"installedAt" json:"installedAt"`
	InstalledOdometer      int        `db:"installedOdometer" json:"installedOdometer"`
	ExpectedLifespanKm     *int       `db:"expectedLifespanKm" json:"expectedLifespanKm"`
	ExpectedLifespanMonths *int       `db:"expectedLifespanMonths" json:"expectedLifespanMonths"`
	Notes                  *string    `db:"notes" json:"notes"`
	ReplacedAt             *time.Time `db:"replacedAt" json:"replacedAt"`
	CreatedAt              time.Time  `db:"createdAt" json:"createdAt"`

	DueSoon bool `db:"-" json:"dueSoon"`
}

type CreateVehiclePartRequest struct {
	PartType               string  `json:"partType"`
	InstalledAt            *string `json:"installedAt"`
	InstalledOdometer      int     `json:"installedOdometer"`
	ExpectedLifespanKm     *int    `json:"expectedLifespanKm"`
	ExpectedLifespanMonths *int    `json:"expectedLifespanMonths"`
	Notes                  *string `json:"notes"`
}

// VehicleAlert تنبيه استحقاق صيانة/وثيقة/قطعة — لوحة تذكير سريعة أعلى صفحة السيارات.
type VehicleAlert struct {
	VehicleID   string `json:"vehicleId"`
	VehicleName string `json:"vehicleName"`
	AlertType   string `json:"alertType"` // MAINTENANCE | PART | DOCUMENT
	Message     string `json:"message"`
	Severity    string `json:"severity"` // warning | danger
}

// VehicleMonthlyStatus يوثّق حالة كل سيارة شهرياً
type VehicleMonthlyStatus struct {
	ID               string    `db:"id" json:"id"`
	VehicleID        string    `db:"vehicleId" json:"vehicleId"`
	Month            string    `db:"month" json:"month"`
	HasIssue         bool      `db:"hasIssue" json:"hasIssue"`
	IssueDescription *string   `db:"issueDescription" json:"issueDescription"`
	Resolved         bool      `db:"resolved" json:"resolved"`
	Notes            *string   `db:"notes" json:"notes"`
	RecordedByID     *string   `db:"recordedById" json:"-"`
	CreatedAt        time.Time `db:"createdAt" json:"createdAt"`
}

type SetVehicleMonthlyStatusRequest struct {
	Month            string  `json:"month"`
	HasIssue         bool    `json:"hasIssue"`
	IssueDescription *string `json:"issueDescription"`
	Resolved         bool    `json:"resolved"`
	Notes            *string `json:"notes"`
}

// VehicleDailyRating تقييم يومي لنظافة وحالة السيارة (11 بند من 0 إلى 4)، مع تقييم
// جودة غسيل الفني/الفنيين المسؤولين (0 إلى 2) — نظام مأخوذ من ملف إكسل الشركة.
// النتيجة الموزونة والراتب المقترح للفني معلومات فقط تُعرض للمراقب كتذكير — لا
// يوجد أي ربط تلقائي براتب الموظف الفعلي بقاعدة البيانات.
type VehicleDailyRating struct {
	ID                string    `db:"id" json:"id"`
	VehicleID         string    `db:"vehicleId" json:"vehicleId"`
	RatedDate         time.Time `db:"ratedDate" json:"ratedDate"`
	Wash              *int      `db:"wash" json:"wash"`
	ExteriorClean     *int      `db:"exteriorClean" json:"exteriorClean"`
	ExteriorCondition *int      `db:"exteriorCondition" json:"exteriorCondition"`
	TireCondition     *int      `db:"tireCondition" json:"tireCondition"`
	GlassClean        *int      `db:"glassClean" json:"glassClean"`
	LightsCondition   *int      `db:"lightsCondition" json:"lightsCondition"`
	TechnicalFaults   *int      `db:"technicalFaults" json:"technicalFaults"`
	FaultDescription  *string   `db:"faultDescription" json:"faultDescription"`
	InteriorClean     *int      `db:"interiorClean" json:"interiorClean"`
	SeatsCondition    *int      `db:"seatsCondition" json:"seatsCondition"`
	InteriorDirt      *int      `db:"interiorDirt" json:"interiorDirt"`
	Smell             *int      `db:"smell" json:"smell"`
	Notes             *string   `db:"notes" json:"notes"`
	RecordedByID      *string   `db:"recordedById" json:"-"`
	CreatedAt         time.Time `db:"createdAt" json:"createdAt"`

	WeightedScore *float64            `db:"-" json:"weightedScore"`
	Vehicle       *Vehicle            `db:"-" json:"vehicle"`
	RecordedBy    *EmployeeBrief      `db:"-" json:"recordedBy"`
	WashRatings   []VehicleWashRating `db:"-" json:"washRatings"`
}

// VehicleWashRating تقييم فني واحد لجودة غسله لسيارة معيّنة بيوم معيّن (0-2).
type VehicleWashRating struct {
	ID            string    `db:"id" json:"id"`
	DailyRatingID string    `db:"dailyRatingId" json:"dailyRatingId"`
	EmployeeID    string    `db:"employeeId" json:"employeeId"`
	Score         int       `db:"score" json:"score"`
	CreatedAt     time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type TechnicianWashRatingInput struct {
	EmployeeID string `json:"employeeId"`
	Score      int    `json:"score"`
}

type CreateVehicleDailyRatingRequest struct {
	VehicleID         string                      `json:"vehicleId"`
	RatedDate         *string                     `json:"ratedDate"`
	Wash              *int                        `json:"wash"`
	ExteriorClean     *int                        `json:"exteriorClean"`
	ExteriorCondition *int                        `json:"exteriorCondition"`
	TireCondition     *int                        `json:"tireCondition"`
	GlassClean        *int                        `json:"glassClean"`
	LightsCondition   *int                        `json:"lightsCondition"`
	TechnicalFaults   *int                        `json:"technicalFaults"`
	FaultDescription  *string                     `json:"faultDescription"`
	InteriorClean     *int                        `json:"interiorClean"`
	SeatsCondition    *int                        `json:"seatsCondition"`
	InteriorDirt      *int                        `json:"interiorDirt"`
	Smell             *int                        `json:"smell"`
	Notes             *string                     `json:"notes"`
	TechnicianRatings []TechnicianWashRatingInput `json:"technicianRatings"`
}

// VehicleScoreSummary متوسط النتيجة الموزونة لسيارة خلال فترة معيّنة — للوحة تذكير المراقب.
type VehicleScoreSummary struct {
	VehicleID    string  `db:"vehicleId" json:"vehicleId"`
	VehicleName  string  `db:"vehicleName" json:"vehicleName"`
	RatingsCount int     `db:"ratingsCount" json:"ratingsCount"`
	AverageScore float64 `db:"averageScore" json:"averageScore"`
}

// TechnicianWashSummary مجموع نقاط غسيل الفني خلال فترة، والراتب المقترح (تذكير
// بس — ما ينحفظ ولا يؤثر براتب الموظف الفعلي إلا إذا المراقب عدّله يدوياً).
type TechnicianWashSummary struct {
	EmployeeID     string  `json:"employeeId"`
	EmployeeName   string  `json:"employeeName"`
	VehiclesWashed int     `json:"vehiclesWashed"`
	TotalPoints    int     `json:"totalPoints"`
	SuggestedWage  float64 `json:"suggestedWage"`
	MonthlyCap     float64 `json:"monthlyCap"`
}
