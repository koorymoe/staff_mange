package model

// DailyStats — الفقرة الأولى (أرقام إجمالية) والفقرة الثانية (كل موظف
// وحجوزاته اليوم) — لتاريخ معيّن (اليوم افتراضياً، أو أي تاريخ سابق بالفلتر).
type DailyStats struct {
	Date                 string               `json:"date"`
	TotalBookings        int                  `json:"totalBookings"`
	MorningBookings      int                  `json:"morningBookings"`
	EveningBookings      int                  `json:"eveningBookings"`
	CrewOutCount         int                  `json:"crewOutCount"`         // عدد الموظفين الي طلعوا للحجوزات
	VehiclesOutCount     int                  `json:"vehiclesOutCount"`     // عدد السيارات المستخدمة
	TotalEmployeesCount  int                  `json:"totalEmployeesCount"`  // إجمالي عدد الموظفين بالنظام
	TotalSalesAmount     float64              `json:"totalSalesAmount"`     // إجمالي المبيعات (فواتير الليدر)
	TotalProfitAmount    float64              `json:"totalProfitAmount"`    // إجمالي الأرباح (مجموع العمولات)
	Employees            []DailyEmployeeStats `json:"employees"`
}

type DailyEmployeeStats struct {
	EmployeeID        string `json:"employeeId"`
	EmployeeName      string `json:"employeeName"`
	Role              string `json:"role"`
	BookingsAssigned  int    `json:"bookingsAssigned"`
	BookingsCompleted int    `json:"bookingsCompleted"`
	CheckedIn         bool   `json:"checkedIn"`
}

// WeeklyStats — إنتاجية الكوادر (عمل وتنفيذ فعلي) وحجم المبيعات (من فواتير
// الليدر) خلال آخر 7 أيام، وإحصائية موظفي المبيعات حسب عدد الحجوزات المدخلة.
type WeeklyStats struct {
	WeekStart string               `json:"weekStart"`
	Crew      []WeeklyCrewStats    `json:"crew"`
	Sales     []WeeklySalesStats   `json:"sales"`
}

type WeeklyCrewStats struct {
	EmployeeID           string  `json:"employeeId"`
	EmployeeName         string  `json:"employeeName"`
	Role                 string  `json:"role"`
	CompletedBookings    int     `json:"completedBookings"`
	SalesVolume          float64 `json:"salesVolume"` // نفس معادلة العمولة (ليدر/فني)
}

type WeeklySalesStats struct {
	EmployeeID      string `json:"employeeId"`
	EmployeeName    string `json:"employeeName"`
	BookingsEntered int    `json:"bookingsEntered"`
}

// ProjectStageStats — عدد المشاريع بكل مرحلة.
type ProjectStageStats struct {
	Stage string `json:"stage"`
	Count int    `json:"count"`
}
