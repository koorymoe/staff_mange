package model

type SmartKpiBreakdownEntry struct {
	Count  int `json:"count"`
	Points int `json:"points"`
}

type CompletionSpeedEntry struct {
	AvgMinutes int `json:"avgMinutes"`
	Points     int `json:"points"`
}

type WorkReportsEntry struct {
	Count       int `json:"count"`
	FullReports int `json:"fullReports"`
	Points      int `json:"points"`
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
