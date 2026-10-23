package model

// حساب تكاليف الشد — تفصيلي لكل الكوادر، بخانة الحسابات.
//
// كان بشيت إكسل منفصل يتحدّث يدوياً؛ صار يتحسب من بيانات الشد نفسها.

type GpsInstallCostRow struct {
	Month        string  `db:"month" json:"month"`
	EmployeeName string  `db:"employeeName" json:"employeeName"`
	Installs     int     `db:"installs" json:"installs"`
	Total        float64 `db:"total" json:"total"`
}

type GpsInstallCostByEmployee struct {
	EmployeeName string  `json:"employeeName"`
	Total        float64 `json:"total"`
}

type GpsInstallCostSummary struct {
	Rows          []GpsInstallCostRow        `json:"rows"`
	ByEmployee    []GpsInstallCostByEmployee `json:"byEmployee"`
	GrandTotal    float64                    `json:"grandTotal"`
	TotalInstalls int                        `json:"totalInstalls"`
	MonthCount    int                        `json:"monthCount"`
}
