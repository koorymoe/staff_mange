package model

import "time"

// ═══ الكيان — مراقب ومساعد شخصي لكل موظف ═══
//
// «كيان يهابه ويخافه الموظف، أول ما يفتح النظام يرحّب بيه: آني
// المراقب عليك والمساعد بنفس الوقت… كأنما بشر يراقب بشر».
//
// ⚠️ كل رقم يطلع من الكيان **مأخوذ من بيانات النظام الحقيقية** —
// مهلة الورق وغرامتها بالدينار من `discipline.go`، والمهام من
// `ExtraTask`، والحجوزات من `Booking`. الكيان ما يخترع رقماً ولا
// يخوّف بغرامة ما راح تنزل فعلاً: تحذير كاذب مرة وحدة يخلّي الموظف
// يتجاهل كل تحذير بعدها.

// حالات توليد الشخصية
const (
	CharacterPending = "PENDING"
	CharacterReady   = "READY"
	CharacterFailed  = "FAILED"
)

// EmployeeCharacter شخصية الكيان المولّدة لموظف واحد.
type EmployeeCharacter struct {
	ID            string     `db:"id" json:"id"`
	EmployeeID    string     `db:"employeeId" json:"employeeId"`
	Persona       *string    `db:"persona" json:"persona"`
	CalmKey       *string    `db:"calmKey" json:"calmKey"`
	HappyKey      *string    `db:"happyKey" json:"happyKey"`
	AngryKey      *string    `db:"angryKey" json:"angryKey"`
	Prompt        *string    `db:"prompt" json:"prompt"`
	Status        string     `db:"status" json:"status"`
	Error         *string    `db:"error" json:"error"`
	GeneratedAt   *time.Time `db:"generatedAt" json:"generatedAt"`
	GeneratedByID *string    `db:"generatedById" json:"generatedById"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time  `db:"updatedAt" json:"updatedAt"`
}

// مزاج الكيان — يُشتق من بيانات الانضباط والمعلّقات، مو عشوائي.
const (
	// EntityMoodHappy نظيف: ماكو تأخير ولا خصم جديد.
	EntityMoodHappy = "HAPPY"
	// EntityMoodWatching عنده معلّقات بس لسه ما تأخر — مرحلة المراقبة.
	EntityMoodWatching = "WATCHING"
	// EntityMoodAngry تأخر فعلاً أو انخصم منه بآخر ٢٤ ساعة.
	EntityMoodAngry = "ANGRY"
	// EntityMoodPositive نظيف **وصار شي إيجابي** بآخر ٢٤ ساعة:
	// نقطة انرجعتله، أو ورق حجز انخلص. الكيان يفرح له بسبب.
	//
	// ⚠️⚠️ **بلا سبب حقيقي يبقى `HAPPY`** — احتفال بلا سبب كذبة
	// بواجهة المستخدم، ونفس قاعدة «رقم غلط أسوأ من ماكو رقم».
	// وهذا الي يخلي الفرح يعني شي لمن يطلع.
	EntityMoodPositive = "POSITIVE"
)

// أنواع سطور الكيان — تحدد الأيقونة واللون بالواجهة.
const (
	EntityLinePaperwork  = "PAPERWORK"
	EntityLineExtraTask  = "EXTRA_TASK"
	EntityLineBooking    = "BOOKING"
	EntityLineDiscipline = "DISCIPLINE"
)

// EntityLine سطر واحد يقوله الكيان — جملة جاهزة بأرقامها ورابطها.
type EntityLine struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
	// Link وين يودّي الضغط على الفقاعة (مسار بالواجهة).
	Link string `json:"link,omitempty"`
	// Urgent الغرامة نزلت فعلاً أو المهلة خلصت — يهزّ الكيان ويقلبه غاضب.
	Urgent bool `json:"urgent"`
}

// EntityBriefing كل الي يحتاجه الكيان حتى يتكلم بصدق مع صاحبه.
type EntityBriefing struct {
	Mood     string `json:"mood"`
	Greeting string `json:"greeting"`
	// Persona وصف شخصية الموظف المشتق — يظهر بلوحة الكيان.
	Persona string `json:"persona,omitempty"`
	// Points رصيد الانضباط من ١٠٠.
	Points int `json:"points"`
	// DinarAtRisk شكد راح يخسر بالدينار لو ما تحرّك بالمعلّقات المتأخرة.
	DinarAtRisk float64      `json:"dinarAtRisk"`
	Lines       []EntityLine `json:"lines"`
	// صور الملامح الثلاث (روابط /api/files/…) — فارغة لو ما انولدت بعد،
	// ووقتها الواجهة تستعمل إيموجي الموظف الموجود أصلاً.
	CalmURL       string `json:"calmUrl,omitempty"`
	HappyURL      string `json:"happyUrl,omitempty"`
	AngryURL      string `json:"angryUrl,omitempty"`
	CharacterState string `json:"characterState"` // PENDING/READY/FAILED/NONE
}
