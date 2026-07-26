package model

// EmployeeMonthlyStats يجمع كل مؤشرات الموظف الواحد خلال شهر واحد بنداء واحد —
// صفحة إحصائيات الموظفين الشهرية (OWNER/ADMIN فقط).
type EmployeeMonthlyStats struct {
	EmployeeID   string `json:"employeeId"`
	EmployeeName string `json:"employeeName"`
	Role         string `json:"role"`
	Month        string `json:"month"` // "YYYY-MM"

	// نقاط الكي بي اي — نفس آلية تسجيل النقاط الموجودة أصلاً (KpiEvaluation)،
	// مجموع النقاط غير الملغاة خلال الشهر.
	KpiPoints int `json:"kpiPoints"`

	// WorkSpeedScore: TODO — يُملأ فعلياً بعد اكتمال ميزة تقدير مدة تنفيذ العمل
	// (job-duration-estimation) التي يبنيها فريق موازي بنفس الجلسة. حالياً
	// دايماً nil عمداً (placeholder صريح، مو رقم مختلق).
	WorkSpeedScore *float64 `json:"workSpeedScore"`

	// VehicleCleanlinessScore: متوسط بند "cleanliness" من "تقييم السائقين بعد
	// المهمة" (VehicleMissionRating) خلال الشهر — nil لو ما عنده أي مهام مقيَّمة
	// بهذا الشهر (مو صفر، حتى ما نوهم إنه تقييم سيء وهو ببساطة ما سافر أصلاً).
	VehicleCleanlinessScore *float64 `json:"vehicleCleanlinessScore"`
	VehicleRatingsCount     int      `json:"vehicleRatingsCount"`

	// عدد الشكاوى المرتبطة بالموظف خلال الشهر (Complaint.relatedEmployeeId).
	ComplaintsCount int `json:"complaintsCount"`

	// عدد فواتير الليدر (المبيعات) التي أنشأها الموظف خلال الشهر.
	SalesCount int `json:"salesCount"`

	// عدد الحجوزات المكتملة التي شارك بها الموظف (أي دور بـBookingAssignment)
	// خلال الشهر.
	CompletedBookingsCount int `json:"completedBookingsCount"`

	// مجموع عمولات الموظف (ليدر أو فني) المحسوبة تلقائياً خلال الشهر.
	TotalCommission float64 `json:"totalCommission"`
}
