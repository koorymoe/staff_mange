package model

// DailyStats — عدد حجوزات اليوم إجمالاً، وكل كادر وإحصائيته اليوم.
type DailyStats struct {
	Date              string                `json:"date"`
	TotalBookingsToday int                  `json:"totalBookingsToday"`
	Employees         []DailyEmployeeStats   `json:"employees"`
}

type DailyEmployeeStats struct {
	EmployeeID   string `json:"employeeId"`
	EmployeeName string `json:"employeeName"`
	Role         string `json:"role"`
	BookingsToday int   `json:"bookingsToday"`
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
