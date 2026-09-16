package model

// ── إحصائيات المشاريع ──
// ثلاث طبقات: نظرة عامة على المشاريع كلها، سطر لكل مشروع بقيمته المالية،
// وسطر لكل موظف يبيّن دوره الفعلي بالمشاريع (منو أضاف، منو طلع كشف، منو
// كان مسؤول، ومنو استلم مشروع مُسلَّم).

// ProjectValueRow مشروع واحد بقيمته المالية والأشخاص المرتبطين بيه.
type ProjectValueRow struct {
	ID              string   `db:"id" json:"id"`
	Code            string   `db:"code" json:"code"`
	Name            string   `db:"name" json:"name"`
	Stage           string   `db:"stage" json:"stage"`
	WorkType        *string  `db:"workType" json:"workType"`
	Priority        string   `db:"priority" json:"priority"`
	PriceRaw        *string  `db:"price" json:"priceRaw"`
	PriceValue      *float64 `db:"priceValue" json:"priceValue"`
	CreatedByName   *string  `db:"createdByName" json:"createdByName"`
	ResponsibleName *string  `db:"responsibleName" json:"responsibleName"`
	SurveyorName    *string  `db:"surveyorName" json:"surveyorName"`
	DelegatedToName *string  `db:"delegatedToName" json:"delegatedToName"`
	HasSurvey       bool     `db:"hasSurvey" json:"hasSurvey"`
	CreatedAt       string   `db:"createdAt" json:"createdAt"`
}

// ProjectEmployeeStatRow إحصائية موظف واحد داخل المشاريع.
type ProjectEmployeeStatRow struct {
	EmployeeID string `db:"employeeId" json:"employeeId"`
	Name       string `db:"name" json:"name"`
	Role       string `db:"role" json:"role"`

	// أضاف المشروع أو رحّل الحجز لإدارة المشاريع
	AddedCount int `db:"addedCount" json:"addedCount"`
	// طلع كشف (منفّذ الكشف على المشروع)
	SurveyAssignedCount int `db:"surveyAssignedCount" json:"surveyAssignedCount"`
	// من هذي الكشوفات، شكد واحد انملت استمارته فعلاً
	SurveyFilledCount int `db:"surveyFilledCount" json:"surveyFilledCount"`
	// كان المسؤول عن المشروع
	ResponsibleCount int `db:"responsibleCount" json:"responsibleCount"`
	// استلم مشروع مُسلَّم (مرات التسليم بالسجل)
	DelegationsReceived int `db:"delegationsReceived" json:"delegationsReceived"`
	// مشاريع مُسلَّمة إله حالياً
	CurrentlyDelegated int `db:"currentlyDelegated" json:"currentlyDelegated"`
	// مجموع قيمة المشاريع الي هو مسؤول عنها
	ResponsibleValue float64 `db:"responsibleValue" json:"responsibleValue"`
	// مشاريع مكتملة من الي هو مسؤول عنها
	CompletedCount int `db:"completedCount" json:"completedCount"`
}

// ProjectStatisticsOverview أرقام عامة فوق الصفحة.
type ProjectStatisticsOverview struct {
	TotalProjects   int          `json:"totalProjects"`
	TotalValue      float64      `json:"totalValue"`
	PricedProjects  int          `json:"pricedProjects"`
	AverageValue    float64      `json:"averageValue"`
	CompletedCount  int          `json:"completedCount"`
	RejectedCount   int          `json:"rejectedCount"`
	ActiveCount     int          `json:"activeCount"`
	DelegatedCount  int          `json:"delegatedCount"`
	SurveysFilled   int          `json:"surveysFilled"`
	StageBreakdown  ProjectStats `json:"stageBreakdown"`
	CompletedValue  float64      `json:"completedValue"`
	InProgressValue float64      `json:"inProgressValue"`
}

// ProjectStatisticsResponse الرد الكامل لصفحة الإحصائيات.
type ProjectStatisticsResponse struct {
	Overview  ProjectStatisticsOverview `json:"overview"`
	Projects  []ProjectValueRow         `json:"projects"`
	Employees []ProjectEmployeeStatRow  `json:"employees"`
}
