package model

import "time"

// PerformanceReview تقييم أداء منفصل تماماً عن KPI (الغرامات المالية) — يحدد فقط
// هل الموظف يستحق تدريب أو لا. سلسلة هرمية: التيم ليدر يقيّم فنييه، والإداري
// (HR_COORDINATOR) يقيّم التيم ليدر نفسه، والأدمن يقدر يقيّم أي أحد.
type PerformanceReview struct {
	ID          string `db:"id" json:"id"`
	EmployeeID  string `db:"employeeId" json:"employeeId"`
	EvaluatorID string `db:"evaluatorId" json:"evaluatorId"`
	Rating      string `db:"rating" json:"rating"` // POSITIVE | NEGATIVE
	Reason      string `db:"reason" json:"reason"`
	// الحجز الي انقيّم عليه. فاضي بالتقييمات القديمة الي انسجّلت قبل
	// ما يصير التقييم مربوط بشغل.
	BookingID *string   `db:"bookingId" json:"bookingId"`
	// درجات ١-٥ اختيارية — تفصّل «وين» بالضبط بدل حكم واحد عام.
	// NULL يعني الليدر ما نطّى نجوم، مو إنه نطّى صفر.
	CommitmentScore *int      `db:"commitmentScore" json:"commitmentScore"`
	SpeedScore      *int      `db:"speedScore" json:"speedScore"`
	QualityScore    *int      `db:"qualityScore" json:"qualityScore"`
	CreatedAt       time.Time `db:"createdAt" json:"createdAt"`

	Employee  *EmployeeBrief `db:"-" json:"employee"`
	Evaluator *EmployeeBrief `db:"-" json:"evaluator"`
}

// ═══ أنواع التقييم ═══
//
// «يحتاج تدريب» غير «مخالفة سلوك» — الأول نقص مهارة علاجه دورة،
// والثاني إجراء إداري. خلطهن بخانة وحدة يظلم الاثنين: صاحب الأسلوب
// السيّئ ينزل بدورة فنية ما تعالج شي، وناقص المهارة ينحسب مخالف.
const (
	ReviewPositive      = "POSITIVE"
	ReviewNeedsTraining = "NEEDS_TRAINING"
	ReviewMisconduct    = "MISCONDUCT"
	ReviewCommitment    = "COMMITMENT"
)

var ReviewRatingLabels = map[string]string{
	ReviewPositive:      "أداء إيجابي",
	ReviewNeedsTraining: "يحتاج تدريب",
	ReviewMisconduct:    "مخالفة سلوك",
	ReviewCommitment:    "خلل بالالتزام",
}

// ReviewNeedsAdminAction هل هذا التقييم يحتاج قرار إداري؟
//
// ⚠️ الليدر **ما يغرّم** — يبلّغ بس. الغرامة قرار الإدارة.
// إعطاء الليدر سلطة غرامة مباشرة على زملائه يخلي أي خلاف شخصي
// يتحوّل خصم من راتب، والنظام يصير سلاح مو أداة.
func ReviewNeedsAdminAction(rating string) bool {
	return rating == ReviewMisconduct || rating == ReviewCommitment
}

// ValidReviewRating فحص القيمة.
func ValidReviewRating(r string) bool {
	_, ok := ReviewRatingLabels[r]
	return ok
}

type CreatePerformanceReviewRequest struct {
	EmployeeID string  `json:"employeeId"`
	Rating     string  `json:"rating"`
	Reason     string  `json:"reason"`
	BookingID  *string `json:"bookingId"`

	CommitmentScore *int `json:"commitmentScore"`
	SpeedScore      *int `json:"speedScore"`
	QualityScore    *int `json:"qualityScore"`
}

// ValidReviewScore الدرجة إما فاضية (ما انطّى نجوم) أو ١-٥.
// ⚠️ الفحص هنا **زيادة** على قيد قاعدة البيانات مو بديل عنه: القيد
// يحمي من أي مسار ثاني، وهذا ينطي رسالة عربية مفهومة بدل خطأ SQL.
func ValidReviewScore(v *int) bool {
	return v == nil || (*v >= 1 && *v <= 5)
}

// ═══ حجز ينتظر تقييم كادره ═══
//
// الليدر ما يحتاج يدور على موظفيه بقائمة — النظام يگله «هذني
// الحجوزات الي خلّصتها، ومنو طلع وياك بكل وحدة».
type BookingAwaitingReview struct {
	BookingID    string     `db:"bookingId" json:"bookingId"`
	Code         string     `db:"code" json:"code"`
	CustomerName string     `db:"customerName" json:"customerName"`
	ServiceName  *string    `db:"serviceName" json:"serviceName"`
	// هوية الزبون بنفس البطاقة — الليدر يتذكّر الشغلة من العنوان
	// والرقم أسرع ما يتذكرها من كود الحجز.
	CustomerPhone   *string `db:"customerPhone" json:"customerPhone"`
	CustomerAddress *string `db:"customerAddress" json:"customerAddress"`
	CompletedAt  *time.Time `db:"completedAt" json:"completedAt"`

	// الكادر الي طلع بهذا الحجز، وحالة تقييم كل واحد
	Crew []CrewReviewState `db:"-" json:"crew"`
}

type CrewReviewState struct {
	EmployeeID string  `db:"employeeId" json:"employeeId"`
	Name       string  `db:"name" json:"name"`
	Position   *string `db:"position" json:"position"`
	// التقييم الي انسجّل إذا انقيّم — فاضي إذا لسه
	Rating *string `db:"rating" json:"rating"`
	Reason *string `db:"reason" json:"reason"`

	CommitmentScore *int `db:"commitmentScore" json:"commitmentScore"`
	SpeedScore      *int `db:"speedScore" json:"speedScore"`
	QualityScore    *int `db:"qualityScore" json:"qualityScore"`
}
