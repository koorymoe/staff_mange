package model

import "time"

// ══════════════════════════════════════════════════════════════════
// نتائج الجي بي اس للمراقب — قراءة فقط
// ══════════════════════════════════════════════════════════════════
//
// «ماريدها تضهر بهاي الطريقة للمراقب. أريد تضهر كنتائج: يشوف أسماء
// زبائن الجي بي اس ويشوف المشاكل الي عنده والاشتراكات وشوكت تخلص.
// ما تخلي وحدة هيج — ماريد أزيد ازدحامها».
//
// ⚠️ ثلاثة أقسام بتبويب واحد بمكتب المراقب، **بدل ست بنود بالقائمة
// وعشر محطات بصندوقه**. الصندوق مبني على «الشغل يجيك» — بس إغراقه
// بكل حدث جي بي اس يخلّيه ضجيجاً وينتجاهل، وهذا الي منعه صراحةً.
//
// ⚠️ وقراءة فقط: المراقب يشوف ويحاسب، **ما ينفّذ** — التنفيذ (موافقة
// على طلب، حرق شريحة، تجديد) يبقى بشاشات مسؤول الجي بي اس. نفس مبدأ
// «المراقب يشوف الجودة كاملة بلا ما يتصل بالزبون».

// GpsMonitorSnapshot الصورة الي يشوفها المراقب بلمحة وحدة.
type GpsMonitorSnapshot struct {
	// Expiring اشتراكات **قربت تنتهي** ولسه ما انتهت (نافذة قادمة).
	// ⚠️ ما چان اكو استعلام لهذي إطلاقاً — الموجود يجيب المنتهية فعلاً.
	Expiring []GpsMonitorSubscription `json:"expiring"`
	// Expired انتهت وما انجدّدت
	Expired []GpsMonitorSubscription `json:"expired"`
	// Problems مشاكل مفتوحة (صيانة PENDING أو IN_PROGRESS)
	Problems []GpsMonitorProblem `json:"problems"`
	// ExpiringWindowDays نافذة «قربت تنتهي» — تنكتب بالشاشة صراحةً
	// حتى المراقب يعرف شنو يعني «قربت».
	ExpiringWindowDays int `json:"expiringWindowDays"`
}

type GpsMonitorSubscription struct {
	DeviceRequestID  string     `db:"deviceRequestId" json:"deviceRequestId"`
	CustomerName     string     `db:"customerName" json:"customerName"`
	CustomerPhone    string     `db:"customerPhone" json:"customerPhone"`
	GpsNumber        *string    `db:"gpsNumber" json:"gpsNumber"`
	SubscriptionEnd  *time.Time `db:"subscriptionEnd" json:"subscriptionEnd"`
	SubscriptionType *string    `db:"subscriptionType" json:"subscriptionType"`
	// DaysLeft موجب = باقي أيام · سالب = فاتت أيام
	DaysLeft int `db:"daysLeft" json:"daysLeft"`
}

type GpsMonitorProblem struct {
	ID                 string     `db:"id" json:"id"`
	CustomerName       string     `db:"customerName" json:"customerName"`
	CustomerPhone      string     `db:"customerPhone" json:"customerPhone"`
	ProblemDescription string     `db:"problemDescription" json:"problemDescription"`
	Status             string     `db:"status" json:"status"`
	CreatedAt          time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt         *time.Time `db:"resolvedAt" json:"resolvedAt"`
	// AgeDays عمر المشكلة بالأيام — المراقب يحاسب على التأخير
	AgeDays int `db:"ageDays" json:"ageDays"`
}

var GpsMaintenanceStatusLabels = map[string]string{
	"PENDING":     "معلّق",
	"IN_PROGRESS": "قيد المعالجة",
	"COMPLETED":   "مكتمل",
}
