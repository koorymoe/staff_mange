package model

import (
	"encoding/json"
	"time"
)

// ═══ مختبر المحاكاة ═══
//
// «أريد محاكيات… يجي يطبّق هنا». اقرا رأس schema_sim_lab.go للفكرة
// كاملة.
//
// ⚠️ الحقول الي شكلها يتغيّر من جهاز لجهاز (`spec`, `terminals`, `ui`,
// `scene`, `steps`) نوعها `json.RawMessage`: تمرّ من قاعدة البيانات
// للواجهة بلا ما Go يفهم محتواها. هذا مقصود — المحرّك بالواجهة هو الي
// يفهمها، ولو عرّفناها بأنواع Go صار كل جهاز جديد بشكل مختلف يحتاج
// تعديل كود وترحيل. الهدف إن الفني يضيف جهازاً بلا ما يلمس الكود.

// أنواع المحرّكات.
const (
	SimEngineWiring = "WIRING" // توصيل أسلاك على صورة الجهاز
	SimEngineCLI    = "CLI"    // طرفية أوامر
	SimEnginePanel  = "PANEL"  // كيباد أو واجهة برمجة
)

// حالات المحتوى.
const (
	SimStatusDraft     = "DRAFT"
	SimStatusInReview  = "IN_REVIEW"
	SimStatusPublished = "PUBLISHED"
	SimStatusRetired   = "RETIRED"
)

// حالات المحاولة.
const (
	SimAttemptInProgress = "IN_PROGRESS"
	SimAttemptPassed     = "PASSED"
	SimAttemptFailed     = "FAILED"
	SimAttemptAbandoned  = "ABANDONED"
)

// SimCategory فئة أجهزة داخل خدمة — «كاميرات IP»، «أقفال ذكية».
type SimCategory struct {
	ID          string    `db:"id" json:"id"`
	ServiceID   *string   `db:"serviceId" json:"serviceId,omitempty"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	ImagePath   *string   `db:"imagePath" json:"imagePath,omitempty"`
	SortOrder   int       `db:"sortOrder" json:"sortOrder"`
	Archived    bool      `db:"archived" json:"archived"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `db:"updatedAt" json:"updatedAt"`

	// محسوبة وقت الجلب — مو أعمدة.
	ServiceName   *string `db:"-" json:"serviceName,omitempty"`
	ExerciseCount int     `db:"-" json:"exerciseCount"`
}

// SimDevice تعريف موديل جهاز بعينه.
type SimDevice struct {
	ID         string  `db:"id" json:"id"`
	CategoryID string  `db:"categoryId" json:"categoryId"`
	Brand      string  `db:"brand" json:"brand"`
	Model      string  `db:"model" json:"model"`
	Name       string  `db:"name" json:"name"`
	Summary    *string `db:"summary" json:"summary,omitempty"`
	ImagePath  *string `db:"imagePath" json:"imagePath,omitempty"`
	EngineKind string  `db:"engineKind" json:"engineKind"`

	Spec      json.RawMessage `db:"spec" json:"spec"`
	Terminals json.RawMessage `db:"terminals" json:"terminals"`
	UI        json.RawMessage `db:"ui" json:"ui"`

	// Geometry هندسة الجهاز بالمتر — منها يتولّد الجسم ثلاثي الأبعاد بالكود.
	// الشكل مجرد View لنفس الجهاز، فالهندسة تعيش وياه مو بالواجهة.
	Geometry json.RawMessage `db:"geometry" json:"geometry"`

	Status  string `db:"status" json:"status"`
	Version int    `db:"version" json:"version"`

	// ⚠️ مصداقية المحتوى — أهم ثلاثة حقول بالجدول كله.
	// `SourceRef` من وين إجت المعلومة، و`LocalPractice` الطريقة
	// المعتمدة عدنا (تختلف من بلد لبلد)، و`Verified` يعني تأكّدنا منها
	// على جهاز حقيقي. غير المحقّق ما يوصل متدرّباً أبداً.
	SourceRef     *string    `db:"sourceRef" json:"sourceRef,omitempty"`
	LocalPractice *string    `db:"localPractice" json:"localPractice,omitempty"`
	Verified      bool       `db:"verified" json:"verified"`
	VerifiedByID  *string    `db:"verifiedById" json:"verifiedById,omitempty"`
	VerifiedAt    *time.Time `db:"verifiedAt" json:"verifiedAt,omitempty"`

	AuthorID     *string    `db:"authorId" json:"authorId,omitempty"`
	ReviewedByID *string    `db:"reviewedById" json:"reviewedById,omitempty"`
	ReviewedAt   *time.Time `db:"reviewedAt" json:"reviewedAt,omitempty"`
	ReviewNote   *string    `db:"reviewNote" json:"reviewNote,omitempty"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updatedAt" json:"updatedAt"`
}

// SimLesson درس نظري قصير يسبق التمرين.
type SimLesson struct {
	ID         string          `db:"id" json:"id"`
	CategoryID *string         `db:"categoryId" json:"categoryId,omitempty"`
	DeviceID   *string         `db:"deviceId" json:"deviceId,omitempty"`
	Title      string          `db:"title" json:"title"`
	Blocks     json.RawMessage `db:"blocks" json:"blocks"`
	SortOrder  int             `db:"sortOrder" json:"sortOrder"`
	Status     string          `db:"status" json:"status"`
	CreatedAt  time.Time       `db:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time       `db:"updatedAt" json:"updatedAt"`
}

// SimExercise تمرين على مشهد فيه جهاز أو أكثر.
type SimExercise struct {
	ID           string  `db:"id" json:"id"`
	CategoryID   string  `db:"categoryId" json:"categoryId"`
	Title        string  `db:"title" json:"title"`
	Brief        *string `db:"brief" json:"brief,omitempty"`
	EngineKind   string  `db:"engineKind" json:"engineKind"`
	Difficulty   int     `db:"difficulty" json:"difficulty"`
	TimeLimitSec *int    `db:"timeLimitSec" json:"timeLimitSec,omitempty"`
	PassScore    int     `db:"passScore" json:"passScore"`
	MaxAttempts  *int    `db:"maxAttempts" json:"maxAttempts,omitempty"`

	Scene json.RawMessage `db:"scene" json:"scene"`
	Steps json.RawMessage `db:"steps" json:"steps"`

	SkillID *string `db:"skillId" json:"skillId,omitempty"`
	Status  string  `db:"status" json:"status"`
	Version int     `db:"version" json:"version"`

	SourceRef     *string    `db:"sourceRef" json:"sourceRef,omitempty"`
	LocalPractice *string    `db:"localPractice" json:"localPractice,omitempty"`
	Verified      bool       `db:"verified" json:"verified"`
	VerifiedByID  *string    `db:"verifiedById" json:"verifiedById,omitempty"`
	VerifiedAt    *time.Time `db:"verifiedAt" json:"verifiedAt,omitempty"`

	AuthorID     *string    `db:"authorId" json:"authorId,omitempty"`
	ReviewedByID *string    `db:"reviewedById" json:"reviewedById,omitempty"`
	ReviewedAt   *time.Time `db:"reviewedAt" json:"reviewedAt,omitempty"`
	SortOrder    int        `db:"sortOrder" json:"sortOrder"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updatedAt" json:"updatedAt"`

	// محسوبة وقت الجلب: أجهزة المشهد كاملة، حتى الواجهة ما تنادي
	// نداءً لكل جهاز.
	Devices []SimDevice `db:"-" json:"devices,omitempty"`
	// أفضل نتيجة للموظف الحالي على هذا التمرين.
	BestScore *int `db:"-" json:"bestScore,omitempty"`
	Passed    bool `db:"-" json:"passed"`
}

// SimAttempt محاولة موظف على تمرين.
type SimAttempt struct {
	ID              string          `db:"id" json:"id"`
	ExerciseID      string          `db:"exerciseId" json:"exerciseId"`
	ExerciseVersion int             `db:"exerciseVersion" json:"exerciseVersion"`
	EmployeeID      string          `db:"employeeId" json:"employeeId"`
	Status          string          `db:"status" json:"status"`
	Score           *int            `db:"score" json:"score,omitempty"`
	StepsTotal      int             `db:"stepsTotal" json:"stepsTotal"`
	StepsPassed     int             `db:"stepsPassed" json:"stepsPassed"`
	HintsUsed       int             `db:"hintsUsed" json:"hintsUsed"`
	WrongCount      int             `db:"wrongCount" json:"wrongCount"`
	DurationSec     *int            `db:"durationSec" json:"durationSec,omitempty"`
	State           json.RawMessage `db:"state" json:"state"`
	StartedAt       time.Time       `db:"startedAt" json:"startedAt"`
	FinishedAt      *time.Time      `db:"finishedAt" json:"finishedAt,omitempty"`
	CreatedAt       time.Time       `db:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time       `db:"updatedAt" json:"updatedAt"`
}

// SimAttemptEvent حركة وحدة داخل محاولة — للتحليل وإعادة العرض.
type SimAttemptEvent struct {
	ID        string          `db:"id" json:"id"`
	AttemptID string          `db:"attemptId" json:"attemptId"`
	StepIndex *int            `db:"stepIndex" json:"stepIndex,omitempty"`
	Kind      string          `db:"kind" json:"kind"`
	Payload   json.RawMessage `db:"payload" json:"payload"`
	AtMs      int             `db:"atMs" json:"atMs"`
	CreatedAt time.Time       `db:"createdAt" json:"createdAt"`
}

// ═══ طلبات الواجهة ═══

// SaveAttemptProgressRequest حفظ تقدّم بالنص — حتى يوقّف ويكمّل باچر.
type SaveAttemptProgressRequest struct {
	State       json.RawMessage        `json:"state"`
	StepsPassed int                    `json:"stepsPassed"`
	HintsUsed   int                    `json:"hintsUsed"`
	WrongCount  int                    `json:"wrongCount"`
	Events      []SaveAttemptEventItem `json:"events"`
}

// SaveAttemptEventItem حركة وحدة ترسلها الواجهة بالدفعة.
type SaveAttemptEventItem struct {
	StepIndex *int            `json:"stepIndex,omitempty"`
	Kind      string          `json:"kind"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	AtMs      int             `json:"atMs"`
}

// FinishAttemptRequest إنهاء المحاولة.
//
// ⚠️ الدرجة الي ترسلها الواجهة **ما تنقبل كما هي**: السيرفر يعيد
// حسابها من أوزان الخطوات وسجل الأحداث. التقييم يصير بالمتصفح لأن
// إعادة كتابة قواعد المحاكاة بـGo تعني نسختين تنحرفن عن بعض — بس
// خزن الدرجة بلا تحقّق يعني أي واحد يفتح أدوات المطوّر ينجح.
type FinishAttemptRequest struct {
	State       json.RawMessage        `json:"state"`
	StepsPassed int                    `json:"stepsPassed"`
	HintsUsed   int                    `json:"hintsUsed"`
	WrongCount  int                    `json:"wrongCount"`
	DurationSec int                    `json:"durationSec"`
	Events      []SaveAttemptEventItem `json:"events"`
}
