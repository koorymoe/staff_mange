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
	Rating     *VehicleMissionRating     `db:"-" json:"rating,omitempty"`
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

	// حقول خاصة بتعبئة الوقود
	Liters             *float64 `db:"liters" json:"liters"`
	FilledByEmployeeID *string  `db:"filledByEmployeeId" json:"filledByEmployeeId"`
	ReceiptNumber      *string  `db:"receiptNumber" json:"receiptNumber"`
	StationName        *string  `db:"stationName" json:"stationName"`
	// اسم الي عبّأ، جاي بـJOIN حتى الجدول يعرضه بدون استعلام لكل صف
	FilledByName *string `db:"filledByName" json:"filledByName"`
	// صورة الوصل تنحفظ base64 وهيه ثقيلة — القوائم ترجع العلم فقط،
	// والصورة نفسها تنجلب بطلب مستقل لما المستخدم يضغط عليها.
	HasReceiptPhoto    bool    `db:"hasReceiptPhoto" json:"hasReceiptPhoto"`
	ReceiptPhotoBase64 *string `db:"-" json:"receiptPhotoBase64,omitempty"`
}

// VehicleLogCreateResult نتيجة إنشاء سجل سيارة، مع نتيجة فحص شذوذ الوقود إن كان النوع FUEL.
type VehicleLogCreateResult struct {
	*VehicleLog
	FuelAnomaly *FuelAnomalyResult `json:"fuelAnomaly,omitempty"`
}

type CreateVehicleLogRequest struct {
	Type            string   `json:"type"`
	PerformedAt     *string  `json:"performedAt"`
	NextDueAt       *string  `json:"nextDueAt"`
	NextDueOdometer *int     `json:"nextDueOdometer"`
	Odometer        *int     `json:"odometer"`
	Cost            *float64 `json:"cost"`
	Notes           *string  `json:"notes"`

	Liters             *float64 `json:"liters"`
	FilledByEmployeeID *string  `json:"filledByEmployeeId"`
	ReceiptNumber      *string  `json:"receiptNumber"`
	StationName        *string  `json:"stationName"`
	ReceiptPhotoBase64 *string  `json:"receiptPhotoBase64"`
}

// UpdateVehicleLogRequest تعديل سجل موجود — كل الحقول اختيارية، الي ينرسل
// فقط ينتحدث. ClearReceiptPhoto لازم علم مستقل لأن COALESCE ما يقدر يفرّغ عمود.
type UpdateVehicleLogRequest struct {
	PerformedAt        *string  `json:"performedAt"`
	Odometer           *int     `json:"odometer"`
	Cost               *float64 `json:"cost"`
	Notes              *string  `json:"notes"`
	NextDueAt          *string  `json:"nextDueAt"`
	NextDueOdometer    *int     `json:"nextDueOdometer"`
	Liters             *float64 `json:"liters"`
	FilledByEmployeeID *string  `json:"filledByEmployeeId"`
	ReceiptNumber      *string  `json:"receiptNumber"`
	StationName        *string  `json:"stationName"`
	ReceiptPhotoBase64 *string  `json:"receiptPhotoBase64"`
	ClearReceiptPhoto  bool     `json:"clearReceiptPhoto"`
}

// EmployeeFuelStat عدد مرات التعبئة وإجمالي اللترات والكلفة لكل موظف بفترة —
// يجاوب سؤال "منو عبّأ وكم مرة" بالإحصائية الشهرية.
type EmployeeFuelStat struct {
	EmployeeID   string  `db:"employeeId" json:"employeeId"`
	EmployeeName string  `db:"employeeName" json:"employeeName"`
	FillCount    int     `db:"fillCount" json:"fillCount"`
	TotalLiters  float64 `db:"totalLiters" json:"totalLiters"`
	TotalCost    float64 `db:"totalCost" json:"totalCost"`
}

// VehicleIncident يغطي الأعطال والأضرار (صدمات) والحوادث الموثّقة بالكامل مع تحديد
// المسبب والتكلفة. الحقول الإضافية (location..repairCost) اختيارية وتُستخدم عملياً
// فقط لنوع ACCIDENT (حادث) — تبقى فارغة لأنواع FAULT/DAMAGE البسيطة.
type VehicleIncident struct {
	ID                    string     `db:"id" json:"id"`
	VehicleID             string     `db:"vehicleId" json:"vehicleId"`
	Type                  string     `db:"type" json:"type"` // FAULT | DAMAGE | ACCIDENT
	Description           string     `db:"description" json:"description"`
	ResponsibleEmployeeID *string    `db:"responsibleEmployeeId" json:"-"`
	Cost                  *float64   `db:"cost" json:"cost"`
	Status                string     `db:"status" json:"status"` // OPEN | RESOLVED
	ReportedByID          *string    `db:"reportedById" json:"-"`
	CreatedAt             time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt            *time.Time `db:"resolvedAt" json:"resolvedAt"`

	// حقول توثيق الحوادث (ACCIDENT) فقط
	Location           *string  `db:"location" json:"location"`
	DriverID           *string  `db:"driverId" json:"-"`
	PeoplePresent      *string  `db:"peoplePresent" json:"peoplePresent"`
	PoliceReportNumber *string  `db:"policeReportNumber" json:"policeReportNumber"`
	RepairCost         *float64 `db:"repairCost" json:"repairCost"`

	ResponsibleEmployee *EmployeeBrief `db:"-" json:"responsibleEmployee"`
	ReportedBy          *EmployeeBrief `db:"-" json:"reportedBy"`
	Driver              *EmployeeBrief `db:"-" json:"driver"`
}

type CreateVehicleIncidentRequest struct {
	Type                  string   `json:"type"`
	Description           string   `json:"description"`
	ResponsibleEmployeeID *string  `json:"responsibleEmployeeId"`
	Cost                  *float64 `json:"cost"`

	// حقول اختيارية لنوع ACCIDENT
	Location           *string  `json:"location"`
	DriverID           *string  `json:"driverId"`
	PeoplePresent      *string  `json:"peoplePresent"`
	PoliceReportNumber *string  `json:"policeReportNumber"`
	RepairCost         *float64 `json:"repairCost"`
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
	Cost                   *float64   `db:"cost" json:"cost"`
	CreatedAt              time.Time  `db:"createdAt" json:"createdAt"`

	DueSoon bool `db:"-" json:"dueSoon"`
}

type CreateVehiclePartRequest struct {
	PartType               string   `json:"partType"`
	InstalledAt            *string  `json:"installedAt"`
	InstalledOdometer      int      `json:"installedOdometer"`
	ExpectedLifespanKm     *int     `json:"expectedLifespanKm"`
	ExpectedLifespanMonths *int     `json:"expectedLifespanMonths"`
	Notes                  *string  `json:"notes"`
	Cost                   *float64 `json:"cost"`
}

// VehicleAlert تنبيه استحقاق صيانة/وثيقة/قطعة أو شذوذ بمصروف الوقود — لوحة تذكير
// سريعة أعلى صفحة السيارات.
type VehicleAlert struct {
	VehicleID   string `json:"vehicleId"`
	VehicleName string `json:"vehicleName"`
	AlertType   string `json:"alertType"` // MAINTENANCE | PART | DOCUMENT | FUEL_ANOMALY
	Message     string `json:"message"`
	Severity    string `json:"severity"` // warning | danger
}

// VehicleExpenseSummary تجميع كل مصاريف سيارة (وقود/صيانة/قطع/حوادث/تنظيف) خلال
// فترة (شهر أو سنة)، مع متوسط التكلفة لكل كم إن أمكن حساب المسافة المقطوعة بالفترة.
type VehicleExpenseSummary struct {
	VehicleID       string   `json:"vehicleId"`
	Period          string   `json:"period"` // مثلاً "2026-07" أو "2026"
	FuelCost        float64  `json:"fuelCost"`
	MaintenanceCost float64  `json:"maintenanceCost"`
	PartsCost       float64  `json:"partsCost"`
	IncidentCost    float64  `json:"incidentCost"`
	CleaningCost    float64  `json:"cleaningCost"`
	TotalCost       float64  `json:"totalCost"`
	DistanceKm      *int     `json:"distanceKm"`
	AvgCostPerKm    *float64 `json:"avgCostPerKm"`
}

// FuelAnomalyResult نتيجة فحص شذوذ تكلفة تعبئة وقود جديدة مقارنة بمتوسط السيارة.
type FuelAnomalyResult struct {
	IsAnomaly       bool    `json:"isAnomaly"`
	AverageCost     float64 `json:"averageCost"`
	NewCost         float64 `json:"newCost"`
	PercentAboveAvg float64 `json:"percentAboveAvg"`
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

// VehicleBooking حجز مسبق لسيارة بفترة زمنية مستقبلية — منفصل عن VehicleMission
// (الذي يمثل بدء استخدام فعلي حالاً). يحتاج اعتماد من مسؤول قبل التنفيذ.
type VehicleBooking struct {
	ID              string     `db:"id" json:"id"`
	VehicleID       string     `db:"vehicleId" json:"vehicleId"`
	RequestedByID   string     `db:"requestedById" json:"requestedById"`
	Purpose         string     `db:"purpose" json:"purpose"`
	StartAt         time.Time  `db:"startAt" json:"startAt"`
	EndAt           time.Time  `db:"endAt" json:"endAt"`
	Status          string     `db:"status" json:"status"` // PENDING | APPROVED | REJECTED | CANCELLED
	ApprovedByID    *string    `db:"approvedById" json:"approvedById"`
	RejectionReason *string    `db:"rejectionReason" json:"rejectionReason"`
	CreatedAt       time.Time  `db:"createdAt" json:"createdAt"`
	DecidedAt       *time.Time `db:"decidedAt" json:"decidedAt"`

	Vehicle     *Vehicle       `db:"-" json:"vehicle,omitempty"`
	RequestedBy *EmployeeBrief `db:"-" json:"requestedBy,omitempty"`
	ApprovedBy  *EmployeeBrief `db:"-" json:"approvedBy,omitempty"`
}

type CreateVehicleBookingRequest struct {
	VehicleID string `json:"vehicleId"`
	Purpose   string `json:"purpose"`
	StartAt   string `json:"startAt"`
	EndAt     string `json:"endAt"`
}

type DecideVehicleBookingRequest struct {
	Approve         bool    `json:"approve"`
	RejectionReason *string `json:"rejectionReason"`
}

type VehicleBookingFilters struct {
	VehicleID     *string
	RequestedByID *string
	Status        *string
	From          *string
	To            *string
}

// VehicleMissionRating تقييم السائق بعد إنهاء المهمة (تقييم واحد لكل مهمة).
type VehicleMissionRating struct {
	ID          string    `db:"id" json:"id"`
	MissionID   string    `db:"missionId" json:"missionId"`
	RatedByID   string    `db:"ratedById" json:"ratedById"`
	Commitment  int       `db:"commitment" json:"commitment"`
	VehicleCare int       `db:"vehicleCare" json:"vehicleCare"`
	Driving     int       `db:"driving" json:"driving"`
	Cleanliness int       `db:"cleanliness" json:"cleanliness"`
	Notes       *string   `db:"notes" json:"notes"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`

	RatedBy *EmployeeBrief `db:"-" json:"ratedBy,omitempty"`
}

type CreateVehicleMissionRatingRequest struct {
	Commitment  int     `json:"commitment"`
	VehicleCare int     `json:"vehicleCare"`
	Driving     int     `json:"driving"`
	Cleanliness int     `json:"cleanliness"`
	Notes       *string `json:"notes"`
}

// DriverRatingSummary متوسط تقييمات سائق عبر كل مهامه المقيَّمة.
type DriverRatingSummary struct {
	EmployeeID     string  `db:"-" json:"employeeId"`
	RatingsCount   int     `db:"ratingsCount" json:"ratingsCount"`
	AvgCommitment  float64 `db:"avgCommitment" json:"avgCommitment"`
	AvgVehicleCare float64 `db:"avgVehicleCare" json:"avgVehicleCare"`
	AvgDriving     float64 `db:"avgDriving" json:"avgDriving"`
	AvgCleanliness float64 `db:"avgCleanliness" json:"avgCleanliness"`
	AvgOverall     float64 `db:"avgOverall" json:"avgOverall"`
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

// ═══ إحصاء الغسل: كل سيارة × كل شهر ═══
//
// طلب أبو الكميات: «نظام غسل السيارات جيد ولكن يفتقر للإحصائيات
// المفصّلة — مثلاً السيارة الأربيلية كم مرة انغسلت بالشهر ومنو الي
// غسلها».
//
// ⚠️ البيانة موجودة أصلاً: `VehicleWashRating.employeeId` هو **منو
// غسل** (منفصل عن `VehicleDailyRating.recordedById` = منو قيّم).
// الناقص چان التجميع: `TechnicianWashSummaries` تجمع بالموظف بس
// وترمي `vehicleId` كلياً، وماكو ولا معامل شهر بأي مسار.
//
// ⚠️⚠️ **تعريف «انغسلت» مكتوب صراحةً**: عمود `wash` هو **درجة جودة
// ٠–٤** وممكن يكون فارغاً، وصف التقييم اليومي ينكتب حتى بيوم ماكو
// غسل أصلاً. فعدّه كـ«مرة غسل» يخترع رقماً.
//
// التعريف المعتمد: **انغسلت = اكو صف `VehicleWashRating`** — يعني
// اكو شخص مسمّى غسلها. أي تعريف ثاني يعطي رقماً ما وراه بيانة.
type VehicleWashMonthly struct {
	VehicleID   string `json:"vehicleId"`
	VehicleName string `json:"vehicleName"`
	PlateNumber string `json:"plateNumber"`
	// Month بصيغة YYYY-MM بتوقيت بغداد
	Month string `json:"month"`
	// WashCount عدد أيام الغسل (صفوف تقييم يومي فيها غاسل مسمّى)
	WashCount int `json:"washCount"`
	// AvgQuality متوسط درجة الغسل ٠–٤ — nil لو ماكو ولا درجة مسجّلة.
	// ⚠️ nil مو صفر: «ما انقيّم» غير «تقييمه صفر».
	AvgQuality *float64             `json:"avgQuality"`
	Washers    []VehicleWasherCount `json:"washers"`
}

// VehicleWasherCount منو غسل هالسيارة وكم مرة بهذا الشهر
// ⚠️ `db` tags إجبارية: sqlx يطابق العمود باسم الحقل بحروف صغيرة لو
// ماكو tag، فـ`EmployeeID` يصير `employeeid` وما يطابق `"employeeId"`
// — والنتيجة قائمة فاضية بلا أي خطأ ظاهر.
type VehicleWasherCount struct {
	EmployeeID   string `db:"employeeId" json:"employeeId"`
	EmployeeName string `db:"employeeName" json:"employeeName"`
	Times        int    `db:"times" json:"times"`
	TotalPoints  int    `db:"totalPoints" json:"totalPoints"`
}

// VehicleExpenseRow مصروف سيارة واحدة ضمن ملخص لوحة التحكم الشاملة (اسم السيارة +
// إجمالي التكلفة للفترة الحالية).
type VehicleExpenseRow struct {
	VehicleID   string  `json:"vehicleId"`
	VehicleName string  `json:"vehicleName"`
	PlateNumber string  `json:"plateNumber"`
	TotalCost   float64 `json:"totalCost"`
	FuelCost    float64 `json:"fuelCost"`
}

// VehicleUsageRankRow ترتيب سيارة ضمن قائمة "الأكثر استخداماً أو تكلفة".
type VehicleUsageRankRow struct {
	VehicleID    string  `json:"vehicleId"`
	VehicleName  string  `json:"vehicleName"`
	PlateNumber  string  `json:"plateNumber"`
	MissionCount int     `json:"missionCount"`
	DistanceKm   int     `json:"distanceKm"`
	TotalCost    float64 `json:"totalCost"`
}

// FleetDashboardSummary لوحة التحكم الشاملة للأسطول — تجميع كل المؤشرات المطلوبة
// بنداء واحد (GET /api/vehicles/dashboard) بدل جلبها من عدة نقاط.
type FleetDashboardSummary struct {
	Period string `json:"period"` // الشهر الحالي بصيغة YYYY-MM

	TotalVehicles       int `json:"totalVehicles"`
	ActiveVehiclesCount int `json:"activeVehiclesCount"` // عاملة (isActive وبدون عطل مفتوح)
	InMaintenanceCount  int `json:"inMaintenanceCount"`  // بها عطل مفتوح (FAULT)
	OnMissionCount      int `json:"onMissionCount"`      // بمهمة قيد التنفيذ حالياً
	NeedsServiceCount   int `json:"needsServiceCount"`   // تنبيه صيانة/قطعة مستحقة
	ExpiringDocsCount   int `json:"expiringDocsCount"`   // وثيقة قريبة/منتهية الصلاحية

	Alerts []VehicleAlert `json:"alerts"`

	FleetFuelCostThisMonth  float64 `json:"fleetFuelCostThisMonth"`
	FleetTotalCostThisMonth float64 `json:"fleetTotalCostThisMonth"`

	VehicleExpenses []VehicleExpenseRow `json:"vehicleExpenses"` // مصاريف كل سيارة هذا الشهر

	TopByUsage []VehicleUsageRankRow `json:"topByUsage"` // أعلى 5 حسب عدد المهام/المسافة
	TopByCost  []VehicleUsageRankRow `json:"topByCost"`  // أعلى 5 حسب التكلفة
}
