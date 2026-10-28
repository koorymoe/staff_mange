package config

import (
	"os"
	"strconv"
	"strings"
)

// Config يحمل كل إعدادات التطبيق المقروءة من متغيرات البيئة (ENV)
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	CORSOrigins []string
	R2Bucket    string
	R2AccessKey string
	R2SecretKey string
	R2Endpoint  string
	// UploadsDir مجلد التخزين المحلي — يُستخدم لما إعدادات R2 ناقصة.
	UploadsDir     string
	GeminiAPIKey   string
	GeminiDailyCap int
	// GeminiImageModel موديل توليد صور شخصية الكيان. من البيئة مو
	// مكتوب بالكود: أسماء موديلات الصور تتبدّل، وتثبيت الاسم يعني
	// ميزة تنكسر بلا ما نلمس سطر.
	GeminiImageModel string
	// ═══ حاكم التحليل (Claude) ═══
	// ⚠️ مفصول عن مفاتيح Gemini عمداً: ذاك مساعد محادثة للموظف، وهذا
	// محرّك تحليل بالخلفية. خلطهم بمفتاح واحد يعني إن إطفاء أحدهما
	// يطفي الثاني، وحد يومي مشترك يخلّي أسئلة الموظفين تاكل حصّة
	// التحليل (أو العكس) بلا ما ينتبه أحد.
	//
	// ⚠️ وبلا مفتاح النظام **ما يوقف**: يبقى على محرّك القواعد. الميزة
	// تنطفي، مو النظام.
	AnthropicAPIKey string
	// ═══ أربع مفاتيح لأربع وكلاء (دازاي 10-23) ═══
	// كل وكيل بمفتاحه — يتوكف أو ينحد صرفه لحاله، وبالكونسول ينعرف منو صرف شكد.
	// المفتاح الفاضي يرجع لـANTHROPIC_API_KEY حتى ما ينكسر شي.
	AIKeyMatrix    string
	AIKeySecretary string
	AIKeyQuality   string // محجوز — وكيل الجودة عالواتساب (ينتظر قرار الخصوصية)
	AIKeyBooking   string // محجوز — وكيل الحجوزات عالواتساب (ينتظر قرار الخصوصية)
	// موديل وسقف السكرتير — منفصلات عن ماتركس.
	AIModelChat         string
	AIInternalDailyUSD  float64
	AIExternalDailyUSD  float64
	AIModelSecretary    string
	AISecretaryDailyCap int
	AIIrisWebSearch     bool   // بحث الويب لايريس (افتراضياً شغّال)
	AIIrisVoice         string // gemini | whisper — تحويل فويس ايريس (افتراضياً gemini إذا اكو مفتاح)
	// حد الصرف اليومي بالدولار لكل وكيل — يتحسب من AiUsage فيبقى بعد إعادة التشغيل.
	AIMatrixDailyUSD float64
	AIIrisDailyUSD   float64
	// وكيلي الواتساب (دازاي 10-23) — حد يومي بالدولار، وحد محادثات جديدة باليوم.
	AIQualityDailyUSD float64
	AIBookingDailyUSD float64
	WaDailyNew        int
	// WhatsApp Business Cloud API (Meta). بلاهن الإرسال الحقيقي مطفي، والتجربة تشتغل.
	WaToken           string
	WaPhoneID         string
	WaAppSecret       string
	WaVerifyToken     string
	WaTemplateQuality string // قالب أول رسالة لابو الجودة (موافق عليه من Meta)
	WaTemplateBooking string // قالب أول رسالة لابو الحجوزات
	WaTemplateLng     string
	// WaPrivacyApproved قرار دازاي المكتوب (الخصوصية · الكلفة · محل المعالجة · مدة الحفظ) —
	// بلاه ملاحظات الحجز ما تروح للنموذج.
	WaPrivacyApproved bool
	// WhisperURL خدمة تحويل الصوت لنص **الداخلية** (حاوية whisper). فاضي = بلا تحليل فويس.
	WhisperURL string
	// AIModel من البيئة مو مكتوب بالكود: أسماء الموديلات تتبدّل،
	// والترقية لموديل أقوى لازم تصير بتبديل متغيّر بيئة مو بنشر كود.
	AIModel string
	// AIDailyCap سقف طلبات التحليل باليوم — حارس فاتورة. إشارات
	// تنفجر بيوم (خلل، استيراد، كنسة تتكرر) تصير كلفة ما حسبناها.
	AIDailyCap   int
	TutorialsDir string
	// بيانات دخول حساب المالك (OWNER) — لازم تنقرأ من البيئة، ممنوع أي قيمة افتراضية
	// ثابتة بالكود (كانت هذي ثغرة حرجة: username/password مكتوبين نص صريح بالكود
	// ومرفوعين لتاريخ Git). لو غير معرّفتين، seedOwnerAccount ما يغيّر شي — الحساب
	// الموجود بقاعدة البيانات يبقى كما هو (ما ينقفل الوصول للمالك الحالي بالغلط).
	OwnerUsername string
	OwnerPassword string
}

func Load() *Config {
	c := load()
	// 🧠 دماغين (دازاي 10-28): مفتاح داخلي (ماتركس + ايريس) ومفتاح خارجي (الجودة + الحجوزات).
	// إذا انحطّوا يغلبون على مفاتيح الوكلاء القديمة.
	if k := getEnv("AI_KEY_INTERNAL", ""); k != "" {
		c.AIKeyMatrix, c.AIKeySecretary = k, k
	}
	if k := getEnv("AI_KEY_EXTERNAL", ""); k != "" {
		c.AIKeyQuality, c.AIKeyBooking = k, k
	}
	// المفتاح الفاضي لوكيل يرجع للعام — حتى .env القديم يبقى يشتغل.
	for _, k := range []*string{&c.AIKeyMatrix, &c.AIKeySecretary, &c.AIKeyQuality, &c.AIKeyBooking} {
		if *k == "" {
			*k = c.AnthropicAPIKey
		}
	}
	// المحادثة المباشرة ويا المالك (ماتركس شات + ايريس) على موديل أذكى؛ الخلفية تبقى AI_MODEL.
	if c.AIModelChat == "" {
		c.AIModelChat = "claude-sonnet-5-5"
	}
	if c.AIModelSecretary == "" {
		c.AIModelSecretary = c.AIModelChat
	}
	return c
}

func load() *Config {
	return &Config{
		Port:           getEnv("PORT", "4000"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		CORSOrigins:    splitOrigins(getEnv("CORS_ORIGIN", "http://localhost:5173")),
		R2Bucket:       getEnv("R2_BUCKET", ""),
		R2AccessKey:    getEnv("R2_ACCESS_KEY", ""),
		R2SecretKey:    getEnv("R2_SECRET_KEY", ""),
		R2Endpoint:     getEnv("R2_ENDPOINT", ""),
		UploadsDir:     getEnv("UPLOADS_DIR", "data/uploads"),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		GeminiDailyCap: getEnvInt("GEMINI_DAILY_CAP", 300),
		// الافتراضي موديل الصور الحالي من نفس المزوّد — ويُضبط بالبيئة
		// لو تبدّل الاسم بلا ما نعدّل كود.
		GeminiImageModel: getEnv("GEMINI_IMAGE_MODEL", "gemini-2.5-flash-image"),
		AnthropicAPIKey:  getEnv("ANTHROPIC_API_KEY", ""),
		WhisperURL:       getEnv("WHISPER_URL", ""),
		AIModel:          getEnv("AI_MODEL", "claude-haiku-5-5"),
		// ٢٠٠ سقف واسع: حجم الشغل اليوم عشرات الإشارات باليوم، فالسقف
		// ما يعترض الشغل الطبيعي ويمسك الانفجار.
		AIDailyCap:       getEnvInt("AI_DAILY_CAP", 200),
		AIKeyMatrix:      getEnv("AI_KEY_MATRIX", ""),
		AIKeySecretary:   getEnv("AI_KEY_SECRETARY", ""),
		AIKeyQuality:     getEnv("AI_KEY_QUALITY", ""),
		AIKeyBooking:     getEnv("AI_KEY_BOOKING", ""),
		AIModelSecretary: getEnv("AI_MODEL_SECRETARY", ""),
		// موديل المحادثة المباشرة ويا المالك (ماتركس شات) — فارغ = AI_MODEL
		AIModelChat: getEnv("AI_MODEL_CHAT", ""),
		// سقف الدماغين (دازاي 10-28) — الافتراضي، والمالك يغيّره من صفحة الوكلاء
		AIInternalDailyUSD:  getEnvFloat("AI_INTERNAL_DAILY_USD", 2),
		AIExternalDailyUSD:  getEnvFloat("AI_EXTERNAL_DAILY_USD", 1),
		AISecretaryDailyCap: getEnvInt("AI_SECRETARY_DAILY_CAP", 300),
		AIIrisWebSearch:     getEnv("AI_IRIS_WEB_SEARCH", "true") != "false",
		AIIrisVoice:         getEnv("AI_IRIS_VOICE", "gemini"),
		AIMatrixDailyUSD:    getEnvFloat("AI_MATRIX_DAILY_USD", 0.5),
		AIIrisDailyUSD:      getEnvFloat("AI_IRIS_DAILY_USD", 0.5),
		AIQualityDailyUSD:   getEnvFloat("AI_QUALITY_DAILY_USD", 0.2),
		AIBookingDailyUSD:   getEnvFloat("AI_BOOKING_DAILY_USD", 0.2),
		WaDailyNew:          getEnvInt("WA_DAILY_NEW", 10),
		WaToken:             getEnv("WA_TOKEN", ""),
		WaPhoneID:           getEnv("WA_PHONE_ID", ""),
		WaAppSecret:         getEnv("WA_APP_SECRET", ""),
		WaVerifyToken:       getEnv("WA_VERIFY_TOKEN", ""),
		WaTemplateQuality:   getEnv("WA_TEMPLATE_QUALITY", ""),
		WaTemplateBooking:   getEnv("WA_TEMPLATE_BOOKING", ""),
		WaTemplateLng:       getEnv("WA_TEMPLATE_LANG", "ar"),
		WaPrivacyApproved:   getEnv("WA_PRIVACY_APPROVED", "") == "true",
		// TutorialsDir: مسار مجلد أدلة الاستخدام (tutorial-*.html و guide-*.txt) اللي
		// يقرأها المساعد الذكي حتى يعرف يشرح للموظف شلون يستخدم النظام فعلياً.
		// الافتراضي "../tutorials" يناسب التطوير المحلي (تشغيل go run من داخل backend-go
		// حيث مجلد tutorials أخو backend-go بجذر المستودع). بالإنتاج نضبط
		// TUTORIALS_DIR=/app/tutorials صراحة بـ docker-compose.yml مع bind-mount
		// (شوف backend service بـ docker-compose.yml).
		TutorialsDir:  getEnv("TUTORIALS_DIR", "../tutorials"),
		OwnerUsername: getEnv("OWNER_USERNAME", ""),
		OwnerPassword: getEnv("OWNER_PASSWORD", ""),
	}
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func splitOrigins(v string) []string {
	parts := strings.Split(v, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}
