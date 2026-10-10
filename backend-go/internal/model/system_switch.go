package model

import "time"

// SystemSwitch مفتاح ميزة على مستوى النظام كله.
type SystemSwitch struct {
	Key       string    `db:"key" json:"key"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	UpdatedAt time.Time `db:"updatedAt" json:"updatedAt"`
	UpdatedBy *string   `db:"updatedBy" json:"updatedBy"`
}

// المفاتيح المعروفة — **قائمة بيضاء مقصودة**.
//
// ⚠️ بلاها أي واحد عنده حساب مالك يكتب مفتاحاً بأي اسم، فيتجمّع
// بالجدول صفوف ما تقرأها ولا شاشة — و(ع) يحسب إنه أطفى شي وهو ما
// انطفى، لأن الاسم انكتب غلط بحرف.
const (
	// SwitchEntity شخصية الكائن — الودجة العائمة وورقة القصة.
	SwitchEntity = "entity_enabled"
	// SwitchAnnouncements شريط الإعلانات فوق كل الشاشات.
	SwitchAnnouncements = "announcements_enabled"
	// SwitchMatrixAutopilot ماتركس ينفّذ التذكيرات البسيطة لحاله.
	SwitchMatrixAutopilot = "matrix_autopilot_enabled"
	// SwitchMatrixStaffEye عين ماتركس بشاشات الموظفين (المدير والمالك تبقى عندهم).
	SwitchMatrixStaffEye = "matrix_staff_eye_enabled"
	// SwitchMatrixAutoCrew ماتركس يكلّف كادر لحجز باچر ما انحدد (المرحلة الثالثة).
	SwitchMatrixAutoCrew = "matrix_auto_crew"
	// SwitchMatrixMonitorEscalate ماتركس يصعّد بنود المراقب المتأخرة للمدير.
	SwitchMatrixMonitorEscalate = "matrix_monitor_escalate"
	// SwitchMatrixScoring ماتركس يقيّم الموظفين بنقاط (درجة تقييم بس، بلا فلوس).
	SwitchMatrixScoring = "matrix_scoring"
)

// مفاتيح توفير الكلفة — كل ميزة تستعمل هايكو إلها مفتاح. المطفي يرجع للقواعد
// (أو يحفظ النص بلا تحليل بالفويس)، فالكلفة تنزل بلا ما يوقف النظام.
const (
	SwitchAIGuide          = "ai_guide"
	SwitchAIJudge          = "ai_judge"
	SwitchAIDiscovery      = "ai_discovery"
	SwitchAIVoice          = "ai_voice"
	SwitchAIAsk            = "ai_ask"
	SwitchAILearning       = "ai_learning"
	SwitchAIEmployeeReport = "ai_employee_report"
	SwitchAIPeer           = "ai_peer"
	SwitchAITone           = "ai_tone"
	SwitchAIChat           = "ai_chat"
)

// AIFeatureSwitch مفتاح كل ميزة (اسم الميزة بعدّاد الكلفة ← المفتاح).
var AIFeatureSwitch = map[string]string{
	"GUIDE": SwitchAIGuide, "JUDGE": SwitchAIJudge, "DISCOVERY": SwitchAIDiscovery, "VOICE": SwitchAIVoice,
	"ASK": SwitchAIAsk, "LEARNING": SwitchAILearning, "EMPLOYEE_REPORT": SwitchAIEmployeeReport, "PEER": SwitchAIPeer,
	"TONE": SwitchAITone, "CHAT": SwitchAIChat,
}

// systemSwitchDefaultOff مفاتيح **مطفية** لحد ما المالك يشغّلها بنفسه —
// أفعال ماتركس التنفيذية (قرار (ع) 10-05: بأقفال أمان). الباقي شغّال افتراضياً.
var systemSwitchDefaultOff = map[string]bool{
	SwitchMatrixAutoCrew:        true,
	SwitchMatrixMonitorEscalate: true,
	// قرار (ع) 10-08: «ماتركس لازم يقيّم كل موظف» — التقييم درجة بس بلا فلوس، فصار شغّال افتراضياً.
}

// SystemSwitchDefault الحالة لمّا ماكو صف بالجدول.
func SystemSwitchDefault(key string) bool { return !systemSwitchDefaultOff[key] }

// SystemSwitchLabels الاسم العربي لكل مفتاح — للعرض وللسجل.
var SystemSwitchLabels = map[string]string{
	SwitchEntity:                "شخصية الكائن",
	SwitchAnnouncements:         "شريط الإعلانات",
	SwitchMatrixAutopilot:       "ماتركس ينفّذ التذكيرات لحاله",
	SwitchMatrixStaffEye:        "عين ماتركس عند الموظفين",
	SwitchMatrixAutoCrew:        "ماتركس يكلّف كادر حجز باچر لحاله",
	SwitchMatrixMonitorEscalate: "ماتركس يصعّد بنود المراقب المتأخرة",
	SwitchMatrixScoring:         "ماتركس يقيّم الموظفين بالنقاط",
	SwitchAIGuide:               "هايكو: توجيهات ماتركس للموظفين",
	SwitchAIJudge:               "هايكو: أحكام ماتركس",
	SwitchAIDiscovery:           "هايكو: ماتركس اكتشف",
	SwitchAIVoice:               "هايكو: تحليل الفويس",
	SwitchAIAsk:                 "هايكو: اسأل ماتركس",
	SwitchAILearning:            "هايكو: توقعات ماتركس",
	SwitchAIEmployeeReport:      "هايكو: تقارير الموظفين",
	SwitchAIPeer:                "هايكو: صوت الموظفين والمشاكل الوظيفية",
	SwitchAITone:                "هايكو: صياغة تنبيهات ماتركس بأسلوب متجدد",
	SwitchAIChat:                "هايكو: دردشة ماتركس",
}

// KnownSystemSwitch هل هذا مفتاح نعرفه؟
func KnownSystemSwitch(key string) bool {
	_, ok := SystemSwitchLabels[key]
	return ok
}

// systemSwitchAdminAllowed المفاتيح الي مدير النظام (ADMIN) يقدر
// يبدّلها هو الثاني — مو المالك حصراً.
//
// ⚠️ **قائمة بيضاء بالاسم مو قاعدة عامة**: شريط الإعلانات قرار
// تشغيلي يومي، بينما شخصية الكائن مرتبطة بقرارات (ع) الشخصية
// (الصور، الترخيص) وتبقى له حصراً حتى لو مدير النظام موجود.
var systemSwitchAdminAllowed = map[string]bool{
	SwitchAnnouncements:   true,
	SwitchMatrixAutopilot: true,
	SwitchMatrixStaffEye:  true,
	// مفاتيح الكلفة: قرار تشغيلي — المدير كمان.
	SwitchAIGuide: true, SwitchAIJudge: true, SwitchAIDiscovery: true, SwitchAIVoice: true,
	SwitchAIAsk: true, SwitchAILearning: true, SwitchAIEmployeeReport: true, SwitchAIPeer: true,
	SwitchAITone: true, SwitchAIChat: true,
}

// SystemSwitchAdminAllowed هل مدير النظام (ADMIN) مسموح له يبدّل
// هذا المفتاح، إضافة على المالك الي مسموح له كل المفاتيح دايماً؟
func SystemSwitchAdminAllowed(key string) bool {
	return systemSwitchAdminAllowed[key]
}
