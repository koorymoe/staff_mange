package model

type SmartKpiBreakdownEntry struct {
	Count  int `json:"count"`
	Points int `json:"points"`
}

type CompletionSpeedEntry struct {
	AvgMinutes int `json:"avgMinutes"`
	Points     int `json:"points"`
	Entries    int `json:"entries"` // كم شغلة انحسب منها المعدل
	// Basis منين ينحسب: MISSION = من تكليف المهمة لإنجازها · BOOKING = من تسجيل الحجز لإنجازه
	Basis string `json:"basis"`
}

type WorkReportsEntry struct {
	Count       int `json:"count"`
	FullReports int `json:"fullReports"`
	Points      int `json:"points"`
	// تفسير الفرق بين التقارير والحجوزات (ع 10-23)
	Bookings      int `json:"bookings"`
	ExtraVisits   int `json:"extraVisits"`
	MultiBookings int `json:"multiBookings"`
	OutsideMonth  int `json:"outsideMonth"`
	NoReport      int `json:"noReport"`
}

type AttendanceEntry struct {
	DaysPresent int `json:"daysPresent"`
	TotalDays   int `json:"totalDays"`
	Points      int `json:"points"`
}

type SmartKpiBreakdown struct {
	CompletedBookings SmartKpiBreakdownEntry `json:"completedBookings"`
	CompletionSpeed   CompletionSpeedEntry   `json:"completionSpeed"`
	WorkReports       WorkReportsEntry       `json:"workReports"`
	Attendance        AttendanceEntry        `json:"attendance"`
	Complaints        SmartKpiBreakdownEntry `json:"complaints"`
	ManualDeductions  SmartKpiBreakdownEntry `json:"manualDeductions"`
}

type SmartKpiResult struct {
	EmployeeID   string            `json:"employeeId"`
	EmployeeName string            `json:"employeeName"`
	Period       string            `json:"period"`
	Breakdown    SmartKpiBreakdown `json:"breakdown"`
	TotalPoints  int               `json:"totalPoints"`
}
