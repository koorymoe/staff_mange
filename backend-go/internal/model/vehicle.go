package model

import "time"

type Vehicle struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	PlateNumber string    `db:"plateNumber" json:"plateNumber"`
	Color       *string   `db:"color" json:"color"`
	Type        *string   `db:"type" json:"type"`
	IsActive    bool      `db:"isActive" json:"isActive"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
}

type CreateVehicleRequest struct {
	Name        string  `json:"name"`
	PlateNumber string  `json:"plateNumber"`
	Color       *string `json:"color"`
	Type        *string `json:"type"`
}

// VehicleLog يغطي وقود/تنظيف/تبديل زيت — سجل واحد بنوع محدد لكل حدث
type VehicleLog struct {
	ID           string         `db:"id" json:"id"`
	VehicleID    string         `db:"vehicleId" json:"vehicleId"`
	Type         string         `db:"type" json:"type"` // FUEL | CLEANING | OIL_CHANGE
	PerformedAt  time.Time      `db:"performedAt" json:"performedAt"`
	NextDueAt    *time.Time     `db:"nextDueAt" json:"nextDueAt"`
	Odometer     *int           `db:"odometer" json:"odometer"`
	Cost         *float64       `db:"cost" json:"cost"`
	Notes        *string        `db:"notes" json:"notes"`
	RecordedByID *string        `db:"recordedById" json:"-"`
	CreatedAt    time.Time      `db:"createdAt" json:"createdAt"`
	RecordedBy   *EmployeeBrief `db:"-" json:"recordedBy"`
}

type CreateVehicleLogRequest struct {
	Type        string   `json:"type"`
	PerformedAt *string  `json:"performedAt"`
	NextDueAt   *string  `json:"nextDueAt"`
	Odometer    *int     `json:"odometer"`
	Cost        *float64 `json:"cost"`
	Notes       *string  `json:"notes"`
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
