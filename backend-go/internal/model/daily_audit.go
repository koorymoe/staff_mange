package model

// التدقيق اليومي: المحاسب يحدد تاريخ ويشوف حجوزات ذاك اليوم كلها،
// بمبالغها وحالتها، مع أربع مجاميع تخليه يعرف وين واصل.

type DailyAuditRow struct {
	ID             string   `db:"id" json:"id"`
	Code           string   `db:"code" json:"code"`
	Status         string   `db:"status" json:"status"`
	CustomerName   string   `db:"customerName" json:"customerName"`
	CustomerPhone  string   `db:"customerPhone" json:"customerPhone"`
	ServiceName    string   `db:"serviceName" json:"serviceName"`
	AmountVerified bool     `db:"amountVerified" json:"amountVerified"`
	Collected      float64  `db:"collected" json:"collected"`
	QuotedPrice    float64  `db:"quotedPrice" json:"quotedPrice"`
	InvoiceTotal   *float64 `db:"invoiceTotal" json:"invoiceTotal"`
	InvoiceCode    *string  `db:"invoiceCode" json:"invoiceCode"`
	// ExpectedAmount المعتمد: فاتورة الليدر إذا موجودة، وإلا تقدير الإداري
	ExpectedAmount float64 `db:"expectedAmount" json:"expectedAmount"`
	OpenIssues     int     `db:"openIssues" json:"openIssues"`
}

type DailyAuditReport struct {
	Date string          `json:"date"`
	Rows []DailyAuditRow `json:"rows"`

	CompletedCount int `json:"completedCount"`
	PendingCount   int `json:"pendingCount"`
	IssuesCount    int `json:"issuesCount"`

	// ١) المبالغ — الي انجمعت فعلاً من الحجوزات المكتملة
	CollectedTotal float64 `json:"collectedTotal"`
	// ٢) ما تم تدقيقه
	NotVerifiedTotal float64 `json:"notVerifiedTotal"`
	// ٣) كل المبالغ (المدقق + غير المدقق)
	VerifiedTotal   float64 `json:"verifiedTotal"`
	AllAmountsTotal float64 `json:"allAmountsTotal"`
	// ٤) الإجمالي المتوقع لليوم — من فواتير الليدرز وتقديرات الإداري.
	// هذا الي يخلي المحاسب يعرف من الصبح شكد المفروض يجمع.
	ExpectedTotal float64 `json:"expectedTotal"`
}
