package model

import (
	"encoding/json"
	"time"
)

// ══════════════════════════════════════════════════════════════════
// محرّك القصص — الكيان ينفّذ مشهداً بدل ما يعرض تنبيهاً
// ══════════════════════════════════════════════════════════════════

// حالات القصة. ⚠️ **أربع حقائق مختلفة مو حالة وحدة**: «وصل الجهاز»
// غير «انعرض المشهد» غير «فتح التفاصيل» غير «أقرّ بالاطلاع». عمود
// حالة واحد يبلع التفريق، ولذلك كل وحدة إلها طابع زمني مستقل.
const (
	StoryStatusQueued       = "QUEUED"
	StoryStatusDelivered    = "DELIVERED"
	StoryStatusPlaying      = "PLAYING"
	StoryStatusSeen         = "SEEN"
	StoryStatusOpened       = "OPENED"
	StoryStatusAcknowledged = "ACKNOWLEDGED"
	StoryStatusFailed       = "FAILED"
)

// أنواع الأحداث الي تولّد قصة — **السبعة الي قررهم (ع)**.
const (
	StoryEventPointDeducted  = "DISCIPLINE_POINT_DEDUCTED"
	StoryEventPointRestored  = "DISCIPLINE_POINT_RESTORED"
	StoryEventPaperMissing   = "PAPERWORK_MISSING"
	StoryEventTaskDue        = "TASK_DUE"
	StoryEventWorkCompleted  = "WORK_COMPLETED"
	StoryEventAdminMessage   = "ADMIN_MESSAGE"
	StoryEventFieldMission   = "FIELD_MISSION"
	StoryEventPraise         = "PRAISE"
)

// StoryPriority أولوية كل نوع.
//
// ⚠️ **العقوبة والأمان قبل الاحتفال** — نفس ترتيب `EntityMoodPositive`
// الي انبنى: **التحذير يسبق الفرح**. موظف عليه خصم ما يستلم تهنئة
// قبل ما يستلم تحذيره.
var StoryPriority = map[string]int{
	StoryEventPointDeducted: 100,
	StoryEventPaperMissing:  90,
	StoryEventTaskDue:       80,
	StoryEventAdminMessage:  70,
	StoryEventFieldMission:  60,
	StoryEventPointRestored: 40,
	StoryEventWorkCompleted: 30,
	StoryEventPraise:        20,
}

// StoryEventLabel تسمية عربية لكل حدث — بيانات لا كود، التوسعة بسطر.
var StoryEventLabel = map[string]string{
	StoryEventPointDeducted: "خصم نقطة",
	StoryEventPointRestored: "رجوع نقطة",
	StoryEventPaperMissing:  "ورق ناقص",
	StoryEventTaskDue:       "موعد قرب",
	StoryEventWorkCompleted: "شغل انخلص",
	StoryEventAdminMessage:  "رسالة إدارية",
	StoryEventFieldMission:  "مهمة ميدانية",
	StoryEventPraise:        "مدح موثّق",
}

// قاموس الأفعال — الحركات الي يعرفها الـrig.
//
// ⚠️ **بالكود مو بالقاعدة**: نفس مبدأ `MonitorStageLabel` و
// `DesignCategoryLabels`. إضافة فعل سطر بيانات، مو ترحيل.
//
// ⚠️ **وكل فعل إله fallback نصي**: لو فشل Rive أو المتصفح على
// `reduced-motion`، المعنى يوصل كامل بالنص. الحركة زينة، **والمعنى
// إلزامي** — كيان ما يشتغل يعني موظف ما يعرف إنه انخصم.
const (
	ActionEnterFromEdge   = "ENTER_FROM_EDGE"
	ActionWalkToTarget    = "WALK_TO_TARGET"
	ActionRunToEdge       = "RUN_TO_EDGE"
	ActionExitToRecipient = "EXIT_TO_RECIPIENT"
	ActionPickUpDocument  = "PICK_UP_DOCUMENT"
	ActionCarryDocument   = "CARRY_DOCUMENT"
	ActionDeliverDocument = "DELIVER_DOCUMENT"
	ActionOpenDocument    = "OPEN_DOCUMENT"
	ActionReadDocument    = "READ_DOCUMENT"
	ActionPointAtUI       = "POINT_AT_UI"
	ActionSpeak           = "SPEAK"
	ActionWaitForAck      = "WAIT_FOR_ACK"
	ActionCelebrate       = "CELEBRATE"
	ActionShowWarning     = "SHOW_WARNING"
	ActionReturnToIdle    = "RETURN_TO_IDLE"
)

// StoryStep خطوة وحدة بمشهد.
//
// ⚠️ **`Checkpoint` هو الي يخلّي الاستئناف ممكناً**: لو انقطع الاتصال
// وسط المشهد، نرجع لآخر خطوة checkpoint — **ما نعيد المشهد من أوله
// ولا نضيّع الرسالة**. والخصم نفسه ما ينعاد أصلاً لأنه بجدوله المنفصل.
type StoryStep struct {
	Action     string `json:"action"`
	Actor      string `json:"actor"` // MESSENGER | SELF
	DurationMs int    `json:"durationMs"`
	Checkpoint bool   `json:"checkpoint"`
}

// أدوار الممثلين بالمشهد.
//
// ⚠️ **`ActorMessenger` = شخصية المراقب نفسه** — قرار (ع) الصريح:
// «الموظف يعرف منو خصمه». هذا **تغيير سياسة**: اليوم إشعار الخصم
// يوصل بلا اسم. واسم المرسِل ينحفظ نصاً بـ`senderName` حتى لو
// انحذف الموظف يبقى السطر مقروءاً.
const (
	ActorMessenger = "MESSENGER"
	ActorSelf      = "SELF"
)

// StoryScene مشهد جاهز لكل نوع حدث.
//
// ⚠️ **بيانات لا كود**: قصة جديدة = سطر بهالخريطة، مو كتلة كود
// جديدة. هذا الي يخلّي «آلاف القصص» ممكنة بلا آلاف الملفات.
var StoryScenes = map[string][]StoryStep{
	StoryEventPointDeducted: {
		{Action: ActionEnterFromEdge, Actor: ActorMessenger, DurationMs: 900, Checkpoint: true},
		{Action: ActionWalkToTarget, Actor: ActorMessenger, DurationMs: 800},
		{Action: ActionDeliverDocument, Actor: ActorMessenger, DurationMs: 600, Checkpoint: true},
		{Action: ActionOpenDocument, Actor: ActorSelf, DurationMs: 500},
		{Action: ActionShowWarning, Actor: ActorSelf, DurationMs: 400},
		{Action: ActionSpeak, Actor: ActorSelf, DurationMs: 0, Checkpoint: true},
		{Action: ActionWaitForAck, Actor: ActorSelf, DurationMs: 0, Checkpoint: true},
		{Action: ActionRunToEdge, Actor: ActorMessenger, DurationMs: 700},
	},
	StoryEventPaperMissing: {
		{Action: ActionEnterFromEdge, Actor: ActorMessenger, DurationMs: 900, Checkpoint: true},
		{Action: ActionDeliverDocument, Actor: ActorMessenger, DurationMs: 600, Checkpoint: true},
		{Action: ActionOpenDocument, Actor: ActorSelf, DurationMs: 500},
		{Action: ActionSpeak, Actor: ActorSelf, DurationMs: 0, Checkpoint: true},
		{Action: ActionRunToEdge, Actor: ActorMessenger, DurationMs: 700},
	},
	StoryEventTaskDue: {
		{Action: ActionEnterFromEdge, Actor: ActorMessenger, DurationMs: 900, Checkpoint: true},
		{Action: ActionPointAtUI, Actor: ActorMessenger, DurationMs: 600},
		{Action: ActionSpeak, Actor: ActorSelf, DurationMs: 0, Checkpoint: true},
		{Action: ActionRunToEdge, Actor: ActorMessenger, DurationMs: 700},
	},
	StoryEventPointRestored: {
		{Action: ActionEnterFromEdge, Actor: ActorMessenger, DurationMs: 900, Checkpoint: true},
		{Action: ActionDeliverDocument, Actor: ActorMessenger, DurationMs: 600, Checkpoint: true},
		{Action: ActionCelebrate, Actor: ActorSelf, DurationMs: 900},
		{Action: ActionSpeak, Actor: ActorSelf, DurationMs: 0, Checkpoint: true},
		{Action: ActionRunToEdge, Actor: ActorMessenger, DurationMs: 700},
	},
}

// scenesFallback مشهد بسيط لأي نوع ما إله مشهد مخصّص.
//
// ⚠️ **ما نرجّع مشهداً فارغاً**: قصة بلا خطوات تعني رسالة ما تنعرض،
// ورسالة ما تنعرض أسوأ من ماكو رسالة.
var storyFallbackScene = []StoryStep{
	{Action: ActionEnterFromEdge, Actor: ActorMessenger, DurationMs: 900, Checkpoint: true},
	{Action: ActionSpeak, Actor: ActorSelf, DurationMs: 0, Checkpoint: true},
	{Action: ActionRunToEdge, Actor: ActorMessenger, DurationMs: 700},
}

// SceneFor يرجّع خطوات المشهد لنوع الحدث — ومشهداً بديلاً لو ما إله واحد.
func SceneFor(eventKind string) []StoryStep {
	if steps, ok := StoryScenes[eventKind]; ok {
		return steps
	}
	return storyFallbackScene
}

// StoryDailyPhysicalCap سقف المشاهد الجسدية لكل موظف باليوم.
//
// ⚠️⚠️ **شرط بقاء مو تحسين**: (ع) قرر إن **الأحداث السبعة كلها**
// تاخذ قصة. سبعة مصادر تعني عدة مشاهد باليوم الواحد — وموظف يستلم
// ركضاً وورقاً عشر مرات يبطّل ينتبه، **ووقتها العقوبة نفسها تفقد
// أثرها**. بعد السقف القصة **تبقى محفوظة** وتنعرض هادئة بالصندوق،
// **ما تنلغى ولا تضيع**.
const StoryDailyPhysicalCap = 4

// StoryInstance قصة وحدة بمراحلها.
type StoryInstance struct {
	ID                  string          `db:"id" json:"id"`
	EventID             string          `db:"eventId" json:"eventId"`
	EventKind           string          `db:"eventKind" json:"eventKind"`
	StoryType           string          `db:"storyType" json:"storyType"`
	Version             int             `db:"version" json:"version"`
	SenderEmployeeID    *string         `db:"senderEmployeeId" json:"senderEmployeeId,omitempty"`
	SenderName          string          `db:"senderName" json:"senderName"`
	RecipientEmployeeID string          `db:"recipientEmployeeId" json:"recipientEmployeeId"`
	Status              string          `db:"status" json:"status"`
	Priority            int             `db:"priority" json:"priority"`
	Physical            bool            `db:"physical" json:"physical"`
	GroupKey            *string         `db:"groupKey" json:"groupKey,omitempty"`
	CurrentStep         int             `db:"currentStep" json:"currentStep"`
	Payload             json.RawMessage `db:"payload" json:"payload"`
	DeliveredAt         *time.Time      `db:"deliveredAt" json:"deliveredAt,omitempty"`
	SeenAt              *time.Time      `db:"seenAt" json:"seenAt,omitempty"`
	OpenedAt            *time.Time      `db:"openedAt" json:"openedAt,omitempty"`
	AcknowledgedAt      *time.Time      `db:"acknowledgedAt" json:"acknowledgedAt,omitempty"`
	CreatedAt           time.Time       `db:"createdAt" json:"createdAt"`
	ExpiresAt           *time.Time      `db:"expiresAt" json:"expiresAt,omitempty"`
}

// StoryWithScene القصة ومعها خطواتها الجاهزة — الواجهة ما تبني مشهداً
// من عندها، تنفّذ الي يجي من الخادم. **مصدر واحد للمشهد.**
type StoryWithScene struct {
	StoryInstance
	Scene []StoryStep `json:"scene"`
	Label string      `json:"label"`
}

// EmitStoryRequest طلب إنشاء قصة من حدث رسمي **صار فعلاً**.
//
// ⚠️ **`EventID` إجباري ويجي من الحدث الموجود** (`KpiEvaluation.id`،
// `DisciplineEvent.id` ...) — ما ننشئ معرّفاً جديداً، لأن معرّفاً
// جديداً كل نداء يعني **الفهرس الفريد ما يمنع التكرار**، وهذا يفرّغ
// كل ضمانة idempotency من معناها.
type EmitStoryRequest struct {
	EventID     string
	EventKind   string
	SenderID    *string
	SenderName  string
	RecipientID string
	GroupKey    *string
	Payload     map[string]any
}
