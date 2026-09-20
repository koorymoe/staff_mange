package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"staffmange-api/internal/model"
)

// ═══ الحاكم النموذجي ═══
//
// التنفيذ الثاني لواجهة `Judge` — الي چان مكتوب برأس ai_brain_service.go
// إنه «ينضاف جنبه بلا ما ينتغيّر ولا سطر بالمنادين». هذا هو.
//
// شغله **تفسير الأدلة بس**: يقرا حقائق محسوبة بالكود ويطلّع عنوان وسبب
// واقتراح بلغة يقراها بني آدم. ما يقرر غرامة، وما يوصل لقاعدة البيانات،
// وما يشوف اسم زبون ولا رقم هاتف.
//
// ⚠️⚠️ **الخط الأحمر** — من AGENTS.md حرفياً:
//
//	«الغرامات والنقاط بالقواعد حصراً، وكل غرامة بسبب ودليل.»
//
// فحكم هذا الملف **اقتراح للمراقب**، والمراقب هو الي يقرر. ولا مسار
// من هنا يوصل لـ`Penalize()` — لا اليوم ولا بعدين. نموذج لغوي يغلط،
// وغرامة غلط تعني فلوس من جيب موظف بريء وثقة ما ترجع.
//
// ⚠️ وليش نموذج أصلاً ومحرّك القواعد شغّال؟ القواعد تجاوب على الحتمي
// («باقي شكد على نهاية الدوام») وتطلّع نصاً جامداً. النموذج يجي للي
// القواعد ما تقدر عليه: يربط أنماط متفرقة، ويصيغ كلاماً يفهمه المراقب
// ويقنع الموظف. الاثنان يشتغلان سوا مو واحد بدل الثاني.

// FeedbackSource يجيب أمثلة حقيقية: حكم سابق + قرار المراقب الحقيقي
// عليه — مادة التعلّم. نفس فصل `AiSignalRecorder` بـbooking_service.go:
// واجهة ضيّقة حتى `ModelJudge` ما يعتمد على مستودع كامل.
type FeedbackSource interface {
	RecentJudgedFeedback(kind string, limit int) ([]model.MonitorFeedbackExample, error)
}

// ModelJudge يفسّر الأدلة بنموذج خارجي، ويرجع للقواعد عند أي عثرة.
type ModelJudge struct {
	client   anthropic.Client
	model    string
	fallback Judge
	// feedback: أمثلة حقيقية من قرارات المراقب — اختياري. بدونه
	// النموذج يشتغل عادي بس بلا تعلّم من التاريخ.
	feedback FeedbackSource

	// ═══ الحد اليومي ═══
	// ⚠️ حارس فاتورة مو حارس جودة: إشارات تنفجر بيوم واحد (خلل بالنظام،
	// استيراد، كنسة تتكرر) تصير فاتورة ما حسبناها. نفس نمط
	// assistant_service.go — عدّاد بالذاكرة ينصفّر كل يوم.
	mu        sync.Mutex
	dailyCap  int
	usedToday int
	resetDate string
}

// NewModelJudge يبني الحاكم. `fallback` لازم يكون موجود — بدونه أي
// عثرة بالشبكة تعني إشارة بلا حكم.
func NewModelJudge(apiKey, modelName string, dailyCap int, fallback Judge) *ModelJudge {
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}
	if fallback == nil {
		fallback = RulesJudge{}
	}
	return &ModelJudge{
		client:    anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:     modelName,
		fallback:  fallback,
		dailyCap:  dailyCap,
		resetDate: time.Now().Format("2006-01-02"),
	}
}

func (j *ModelJudge) Name() string { return "model:" + j.model }

// SetFeedbackSource يربط حلقة التعلّم بعد البناء — نفس نمط بقية
// الروابط الاختيارية بالنظام.
func (j *ModelJudge) SetFeedbackSource(f FeedbackSource) { j.feedback = f }

// UsageToday شكد انستهلك اليوم وشكد الحد — للعرض بشاشة المالك.
func (j *ModelJudge) UsageToday() (used, cap_ int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.usedToday, j.dailyCap
}

func (j *ModelJudge) takeQuota() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != j.resetDate {
		j.resetDate = today
		j.usedToday = 0
	}
	if j.usedToday >= j.dailyCap {
		return false
	}
	j.usedToday++
	return true
}

// ═══ ما يطلبه النموذج ═══
//
// ⚠️ ماكو حقل لمعرّف الموظف عمداً. لو خلّيناه يكتب معرّفاً، يكدر يخترع
// واحداً ما موجود أو يلصق التهمة بموظف ثاني — والمعرّف المخترع يعدي
// أي فحص شكلي. بدالها سؤال نعم/لا، وإحنا نحط المعرّف من الإشارة نفسها.
type modelVerdictOut struct {
	Headline              string `json:"headline"`
	Reasoning             string `json:"reasoning"`
	Suggestion            string `json:"suggestion"`
	Severity              string `json:"severity"`
	Confidence            int    `json:"confidence"`
	BlameOnSignalEmployee bool   `json:"blameOnSignalEmployee"`
}

var modelVerdictSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"headline":   map[string]any{"type": "string", "description": "عنوان قصير بالعربية الفصحى المبسّطة، أقل من ٨٠ حرف"},
		"reasoning":  map[string]any{"type": "string", "description": "السبب مشروحاً من الأدلة المعطاة حصراً، جملتين لثلاث"},
		"suggestion": map[string]any{"type": "string", "description": "إجراء عملي واحد يسويه المراقب. فارغ إذا ماكو إجراء واضح"},
		"severity": map[string]any{
			"type":        "string",
			"enum":        []string{model.AiSeverityInfo, model.AiSeverityWatch, model.AiSeverityWarn, model.AiSeverityCritical},
			"description": "INFO=للعلم، WATCH=يستاهل متابعة، WARN=خلل مؤكد، CRITICAL=نمط متكرر",
		},
		"confidence": map[string]any{"type": "integer", "minimum": 0, "maximum": 100, "description": "شكد واثق بهذا التفسير من الأدلة المتوفرة"},
		"blameOnSignalEmployee": map[string]any{
			"type":        "boolean",
			"description": "هل الأدلة تدل إن موظف الإشارة هو المسؤول؟ false إذا السبب خارج عنه أو غير واضح",
		},
	},
	"required":             []string{"headline", "reasoning", "suggestion", "severity", "confidence", "blameOnSignalEmployee"},
	"additionalProperties": false,
}

// ⚠️ التعليمات ثابتة نصاً — وهذا مقصود: الجزء الثابت ينخزن بالسياق
// المؤقت عند المزوّد فينحسب مرة وحدة بدل كل طلب. أي شي متغيّر (تاريخ،
// معرّف) يروح بالرسالة مو هنا، وإلا التخزين ينكسر وتزيد الكلفة.
const modelJudgeSystemPrompt = `أنت مدقّق عمليات بشركة تركيب وصيانة أنظمة (كاميرات، بصمة، ستلايت) بمدينة كربلاء.

مهمتك: تقرا حقائق منتزعة من قاعدة بيانات الشركة، وتفسّرها للمراقب.

قواعد ما تنكسر:
١. اعتمد على الحقائق المعطاة حصراً. ما تفترض شي ما موجود بالأدلة.
٢. قائمة "الفجوات" تعني بيانات ما انجمعت — لا تعتبرها دليل نفي. إذا فجوة مهمة موجودة، نزّل ثقتك واذكرها بالسبب.
٣. إذا الأدلة تبرّئ الموظف، قلها بوضوح قبل أي لوم.
٤. "ما نعرف" جواب مقبول. الثقة الواطية أشرف من تفسير مخترع.
٥. ما تقترح غرامة ولا عقوبة — هذا قرار مو مالك. اقترح إجراء إداري: راجع، اسأل، تابع.
٦. اكتب بالعربية المبسّطة الي يفهمها موظف عراقي. بلا مصطلحات تقنية وبلا إنشاء.`

// Judge يفسّر الأدلة بالنموذج، ويرجع للقواعد عند أي عثرة.
//
// ⚠️ **الرجوع للقواعد مو معالجة أخطاء — هو التصميم.** مزوّد خارجي
// يطيح، والشبكة تنقطع، والحد اليومي يخلص. لو ربطنا التحليل بالنموذج
// وحده، أول عثرة تعني إشارات تتكدس بلا حكم ومراقب يشوف صندوقاً فاضياً
// ويظن ماكو شغل. القواعد تبقى شبكة الأمان دائماً.
func (j *ModelJudge) Judge(sig model.AiSignal, ev model.AiEvidence) (*model.AiVerdict, error) {
	if !j.takeQuota() {
		log.Printf("[ai] الحد اليومي للنموذج خلص (%d) — رجعنا للقواعد", j.dailyCap)
		return j.fallback.Judge(sig, ev)
	}

	out, err := j.ask(sig, ev)
	if err != nil {
		log.Printf("[ai] النموذج ما جاوب لإشارة %s (%v) — رجعنا للقواعد", sig.ID, err)
		return j.fallback.Judge(sig, ev)
	}

	v := j.toVerdict(sig, *out)
	if v.Headline == "" {
		log.Printf("[ai] النموذج رجّع حكماً فارغاً لإشارة %s — رجعنا للقواعد", sig.ID)
		return j.fallback.Judge(sig, ev)
	}
	return v, nil
}

func (j *ModelJudge) ask(sig model.AiSignal, ev model.AiEvidence) (*modelVerdictOut, error) {
	// ⚠️ مهلة قصيرة: هاي كنسة خلفية مو طلب مستخدم ينتظر. طلب معلّق
	// دقيقتين يوقف بقية الإشارات بالدور.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// أمثلة حقيقية من قرارات المراقب — بأفضل جهد: فشل جلبها ما يوقف
	// التحليل، بس النموذج يشتغل بلا تعلّم من التاريخ.
	var examples []model.MonitorFeedbackExample
	if j.feedback != nil {
		if ex, err := j.feedback.RecentJudgedFeedback(sig.Kind, 5); err == nil {
			examples = ex
		}
	}

	resp, err := j.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(j.model),
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{{
			Text:         modelJudgeSystemPrompt,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}},
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{Schema: modelVerdictSchema},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(buildEvidencePrompt(sig, ev, examples))),
		},
	})
	if err != nil {
		return nil, err
	}

	var text strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			text.WriteString(tb.Text)
		}
	}
	if text.Len() == 0 {
		return nil, fmt.Errorf("رد فاضي")
	}

	var out modelVerdictOut
	if err := json.Unmarshal([]byte(text.String()), &out); err != nil {
		return nil, fmt.Errorf("رد مو مقروء: %w", err)
	}
	return &out, nil
}

// buildEvidencePrompt يبني نص الطلب من الإشارة والأدلة.
//
// ⚠️⚠️ **ولا اسم زبون ولا رقم هاتف ولا اسم موظف يطلع من هنا.** ينداز
// صنف الإشارة والحقائق المحسوبة والفجوات بس. سببان:
//
//   - **خصوصية**: البيانات تمر بمزوّد خارجي، والتحليل ما يحتاج أسماء —
//     «وقّف بعد ٢٠ دقيقة وما طلب مادة» تنفهم بلا ما نعرف منو.
//   - **دقة**: الاسم يجرّ النموذج لأحكام مسبقة بدل الأدلة.
func buildEvidencePrompt(sig model.AiSignal, ev model.AiEvidence, examples []model.MonitorFeedbackExample) string {
	var b strings.Builder
	b.WriteString("صنف الحدث: ")
	b.WriteString(signalKindLabel(sig.Kind))
	b.WriteString("\nوقت الحدث: ")
	b.WriteString(sig.OccurredAt.Format("2006-01-02 15:04"))

	b.WriteString("\n\nالحقائق المنتزعة من قاعدة البيانات:\n")
	if len(ev.Facts) > 0 {
		b.Write(ev.Facts)
	} else {
		b.WriteString("{}")
	}

	// الفجوات تنقال صراحة — «الفراغ المعلن أأمن من الفراغ الصامت».
	gaps := decodeGaps(ev.Gaps)
	b.WriteString("\n\nفجوات بالأدلة (بيانات ما قدرنا نجمعها):\n")
	if len(gaps) == 0 {
		b.WriteString("- ماكو فجوات، الأدلة كاملة\n")
	} else {
		for _, g := range gaps {
			b.WriteString("- ")
			b.WriteString(g)
			b.WriteString("\n")
		}
	}

	// ═══ حلقة التعلّم ═══
	// ⚠️ عقل النموذج نفسه ما يتغيّر — الي يتراكم هو الأمثلة الحقيقية
	// من قرارات المراقب الي ننطيها إياه كل مرة. تتغيّر كل طلب، فتروح
	// بالرسالة مو بالتعليمات الثابتة (وإلا التخزين المؤقت ينكسر).
	if len(examples) > 0 {
		b.WriteString("\n\nأمثلة حقيقية من قرارات المراقب على نفس صنف الحدث سابقاً:\n")
		for _, ex := range examples {
			status := "سليم — ما اكو مشكلة حقيقية"
			if ex.MonitorStatus == model.MonitorStatusFlagged {
				status = "فعلاً مشكلة"
			}
			b.WriteString("- حكم سابق «" + ex.Headline + "» بحقائق: ")
			if len(ex.Facts) > 0 {
				b.Write(ex.Facts)
			} else {
				b.WriteString("{}")
			}
			b.WriteString(" ← قرار المراقب: " + status)
			if ex.MonitorNote != nil && *ex.MonitorNote != "" {
				b.WriteString(" — ملاحظته: " + *ex.MonitorNote)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\nفسّر شنو صار ومنو المسؤول وشنو الإجراء المقترح.")
	return b.String()
}

func decodeGaps(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var gaps []string
	if err := json.Unmarshal(raw, &gaps); err != nil {
		return nil
	}
	return gaps
}

func signalKindLabel(kind string) string {
	switch kind {
	case model.AiSignalWorkStopped:
		return "توقف عمل بمنتصف حجز"
	case model.AiSignalLateStart:
		return "تأخر بالخروج للحجز"
	case model.AiSignalRepeatPostpone:
		return "تأجيل متكرر لنفس الحجز"
	case model.AiSignalInvoiceAdjusted:
		return "تعديل على فاتورة"
	case model.AiSignalRepeatPartial:
		return "إنجاز جزئي متكرر"
	case model.AiSignalSelfReportMismatch:
		return "تقرير إنجاز يناقض كادر الحجز الحقيقي"
	default:
		return kind
	}
}

// toVerdict يحوّل رد النموذج لحكم — **بعد تنظيفه**.
//
// ⚠️ كل حقل ينفحص. رد النموذج مدخل خارجي مو نتيجة حساب: خطورة مخترعة
// تكسر ترتيب الصندوق، وثقة ١٥٠ تعدي عتبة الثقة، ونص بلا سقف يكسر
// العرض. والأهم: **المعرّف ينحط من الإشارة مو من النموذج**.
func (j *ModelJudge) toVerdict(sig model.AiSignal, out modelVerdictOut) *model.AiVerdict {
	name := j.Name()
	v := &model.AiVerdict{
		Source:     model.AiSourceModel,
		ModelName:  &name,
		Headline:   clampText(out.Headline, 120),
		Severity:   normalizeSeverity(out.Severity),
		Confidence: clampInt(out.Confidence, 0, 100),
	}
	if r := clampText(out.Reasoning, 1500); r != "" {
		v.Reasoning = &r
	}
	if s := clampText(out.Suggestion, 500); s != "" {
		v.Suggestion = &s
	}
	// المعرّف من الإشارة حصراً — النموذج يقرر «إي/لا» بس.
	if out.BlameOnSignalEmployee {
		v.BlameEmployeeID = sig.EmployeeID
	}
	return v
}

// normalizeSeverity ترجّع خطورة معروفة. الافتراضي `WATCH` مو `INFO`:
// رد مشوّه يعني شي غلط، فالأسلم إنه ينشاف مو إنه ينخفي بأدنى مستوى.
func normalizeSeverity(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case model.AiSeverityInfo:
		return model.AiSeverityInfo
	case model.AiSeverityWarn:
		return model.AiSeverityWarn
	case model.AiSeverityCritical:
		return model.AiSeverityCritical
	default:
		return model.AiSeverityWatch
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clampText يشيل الفراغ ويقص بحدود الأحرف مو البايتات — القص بالبايت
// يكسر حرفاً عربياً بالنص ويطلع رموزاً مشوّهة بالشاشة.
func clampText(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return strings.TrimSpace(string(r[:maxRunes])) + "…"
}
