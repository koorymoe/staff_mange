package model

// DailyStats — الفقرة الأولى (أرقام إجمالية) والفقرة الثانية (كل موظف
// وحجوزاته اليوم) — لتاريخ معيّن (اليوم افتراضياً، أو أي تاريخ سابق بالفلتر).
type DailyStats struct {
	Date                string               `json:"date"`
	TotalBookings       int                  `json:"totalBookings"`
	MorningBookings     int                  `json:"morningBookings"`
	EveningBookings     int                  `json:"eveningBookings"`
	CrewOutCount        int                  `json:"crewOutCount"`        // عدد الموظفين الي طلعوا للحجوزات
	VehiclesOutCount    int                  `json:"vehiclesOutCount"`    // عدد السيارات المستخدمة
	TotalEmployeesCount int                  `json:"totalEmployeesCount"` // إجمالي عدد الموظفين بالنظام
	TotalSalesAmount    float64              `json:"totalSalesAmount"`    // إجمالي المبيعات (فواتير الليدر)
	TotalProfitAmount   float64              `json:"totalProfitAmount"`   // إجمالي الأرباح (مجموع العمولات)
	Employees           []DailyEmployeeStats `json:"employees"`
}

type DailyEmployeeStats struct {
	EmployeeID        string `json:"employeeId"`
	EmployeeName      string `json:"employeeName"`
	Role              string `json:"role"`
	BookingsAssigned  int    `json:"bookingsAssigned"`
	BookingsCompleted int    `json:"bookingsCompleted"`
	CheckedIn         bool   `json:"checkedIn"`
}

// WeeklyStats — إحصائية أسبوعية بمدى تاريخ حر (من/إلى، يحدده المستخدم بالفلتر).
// الفقرة الأولى: إجمالي حجم المبيعات مقسوم صباحي/مسائي/مجموع. الفقرة الثانية:
// نفس أعمدة الإحصائية الشهرية بالضبط لكن محسوبة على مدى الأسبوع بدل الشهر
// (نقاط الكي بي اي، الشكاوى، المبيعات، الحجوزات، والعمولة) — عمداً بدون تعديل
// "عدد الخدمات التي يعرفها الموظف" (هذا التعديل خاص بالشهرية فقط).
type WeeklyStats struct {
	From               string                 `json:"from"`
	To                 string                 `json:"to"`
	MorningSalesAmount float64                `json:"morningSalesAmount"`
	EveningSalesAmount float64                `json:"eveningSalesAmount"`
	TotalSalesAmount   float64                `json:"totalSalesAmount"`
	Employees          []EmployeeMonthlyStats `json:"employees"`
}

// ProjectStageStats — عدد المشاريع بكل مرحلة.
type ProjectStageStats struct {
	Stage string `json:"stage"`
	Count int    `json:"count"`
}
