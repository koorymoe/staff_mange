package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// ═══ عقل وكيلَي الواتساب: «ابو الجودة» و«ابو الحجوزات» — دازاي 10-23 ═══
// الخطة: docs/ai-quality-booking-plan.md. البوابة والأدوات والتخزين عند ناروتو؛
// هنا الحلقة المشتركة + الحراسات الي ما نعتمد بيها على التعليمات بس:
//   - طلب إنسان أو عصبية واضحة ← تحويل فوري بلا نداء للنموذج.
//   - create_booking / confirm_booking ما تنفّذ إلا بعد موافقة صريحة من الزبون.
//   - record_contact تقييماته ١–٥ بس.
//   - أي فشل (مطفي، سقف، خطأ نموذج) ← تحويل لإنسان؛ الزبون ما يبقى بلا جواب.
// الأسماء مفتوحة هنا مثل السكرتير — الوكيل يحچي الزبون نفسه (قرار الخصوصية المكتوب
// شرط قبل التشغيل الحقيقي، والمفاتيح مطفية لحد ذاك).

const (
	waMaxSteps    = 6
	waMaxTokens   = 500
	waTimeout     = 45 * time.Second
	waMaxReplyLen = 600 // حرف — رسالة واتساب وحدة قصيرة
)

// الرد الثابت وكت التحويل — ما يعتمد على النموذج.
const waHandoffReply = "تدلل، هسه أحوّلك لزميلنا بالشركة ويتواصل وياك بأقرب وقت 🌹"

// WaHistory رسالة سابقة بالمحادثة: Role = "customer" | "agent".
type WaHistory struct {
	Role string
	Text string
}

// WaAgentTurn رسالة زبون وحدة ويا سياقها (ناروتو يجهّزها من WaConversation).
type WaAgentTurn struct {
	ConversationID string
	CustomerName   string
	AgentName      string // اسم الوكيل من إعدادات المالك
	Facts          string // تفاصيل الحجز أو المتابعة — حتى ما نصرف خطوة على شي معروف
	History        []WaHistory
	Text           string
}

// WaTurnCtx أدوات ناروتو تكتب عليه (مثل SecretaryTurnCtx).
type WaTurnCtx struct {
	ConversationID string
	Handoff        bool
	HandoffReason  string
	Actions        []string // أسماء الأدوات الي تنفّذت فعلاً
}

// WaAgentResult الي يرجع لناروتو: يرسل Reply، وإذا Handoff يحوّل المحادثة لإنسان.
type WaAgentResult struct {
	Reply         string
	Handoff       bool
	HandoffReason string
	Actions       []string
}

// waAgent إعدادات كل وكيل.
type waAgent struct {
	feature  string // مفتاح AIFeatureOn وعدّاد الكلفة
	llm      agentLLM
	model    string
	toolsFor func(tc *WaTurnCtx) []agentTool
	system   func(t WaAgentTurn) string

	capMu    sync.Mutex
	dailyCap int
	capDay   string
	used     int
}

func newWaAgent(feature string, client *anthropic.Client, model string, toolsFor func(*WaTurnCtx) []agentTool, dailyCap int) *waAgent {
	if model == "" {
		model = "claude-haiku-4-5"
	}
	a := &waAgent{feature: feature, model: model, toolsFor: toolsFor, dailyCap: dailyCap}
	if client != nil {
		a.llm = anthropicLLM{c: client}
	}
	return a
}

func (a *waAgent) Enabled() bool { return a != nil && a.llm != nil && AIFeatureOn(a.feature) }

func (a *waAgent) takeQuota() bool {
	if a.dailyCap <= 0 {
		return true
	}
	a.capMu.Lock()
	defer a.capMu.Unlock()
	day := time.Now().In(debriefLoc).Format("2006-01-02")
	if day != a.capDay {
		a.capDay, a.used = day, 0
	}
	if a.used >= a.dailyCap {
		return false
	}
	a.used++
	return true
}

var (
	// الزبون يطلب إنسان صراحة — طلب، مو مجرد ذكر كلمة «موظف» («الموظف چان خوش» ما يحوّل).
	waHumanRe = regexp.MustCompile(`(?i)((اريد|أريد|ابي|أبي|ابغى|بدي|خلي|خلّي|حولني|حوّلني|حولوني|اكدر|أكدر|ممكن)[^.؟?!\n]{0,25}(موظف|انسان|إنسان|شخص|بشر|مدير|مسؤول|احد)|(احچي|احجي|احكي|اتكلم|اتحاچى)[^.؟?!\n]{0,10}(ويا|مع|وي)[^.؟?!\n]{0,10}(موظف|انسان|إنسان|شخص|بشر|مدير|مسؤول|احد)|(اتصلوا|اتصل|خابروني|خابرني|كلموني|كلمني)[^.؟?!\n]{0,8}(بي|بيه|علي|عليه|الي|إلي)?\s*$|talk to (a )?(human|person|agent))`)
	// عصبية أو شتيمة واضحة.
	waAngryRe = regexp.MustCompile(`(نصابين|حرامية|حراميه|كذابين|غشاشين|زبالة|زباله|تفو|لعنة|لعنه|خرب بيتكم|راح اشتكي|محكمة|محكمه|شرطة|شرطه)`)
	// موافقة صريحة من الزبون.
	waYesRe = regexp.MustCompile(`(?i)(^|[^\p{L}])(اي|ايي|ايه|إي|إيه|نعم|اكيد|أكيد|تمام|موافق|ثبت|ثبّت|ثبتها|ثبّتها|اوكي|أوكي|ok|okay|yes|صح|ماشي|زين|خوش|تمت|يلا)($|[^\p{L}])`)
	// نفي يلغي الموافقة («لا مو تمام»).
	waNoRe = regexp.MustCompile(`(^|[^\p{L}])(لا|لاء|مو|ما|كلا|بعدين|انتظر|لحظة|لحظه|اصبر|غير|غيّر)($|[^\p{L}])`)
)

// waNeedsHuman تحويل فوري بلا نموذج: طلب إنسان أو عصبية.
func waNeedsHuman(text string) (bool, string) {
	switch {
	case waHumanRe.MatchString(text):
		return true, "الزبون طلب يحچي ويا موظف"
	case waAngryRe.MatchString(text):
		return true, "الزبون معصّب — يحتاج إنسان"
	}
	return false, ""
}

// waConfirmed آخر رسالة من الزبون موافقة صريحة بلا نفي.
func waConfirmed(text string) bool {
	return waYesRe.MatchString(text) && !waNoRe.MatchString(text)
}

// أدوات ما تتنفذ إلا بعد موافقة الزبون.
var waConfirmTools = map[string]bool{"create_booking": true, "confirm_booking": true}

func handoffResult(reason string, actions []string) *WaAgentResult {
	return &WaAgentResult{Reply: waHandoffReply, Handoff: true, HandoffReason: reason, Actions: actions}
}

// Reply رد الوكيل على رسالة الزبون. ما يرجّع خطأ: أي فشل يصير تحويل لإنسان
// ويا السبب، حتى ناروتو دائماً عنده رسالة يرسلها.
func (a *waAgent) Reply(ctx context.Context, t WaAgentTurn) *WaAgentResult {
	if need, why := waNeedsHuman(t.Text); need {
		return handoffResult(why, nil)
	}
	if !a.Enabled() {
		return handoffResult("الوكيل مطفي", nil)
	}
	if strings.TrimSpace(t.Text) == "" {
		return &WaAgentResult{Reply: "ما وصلتني رسالتك، تكدر تعيدها؟"}
	}
	ctx, cancel := context.WithTimeout(ctx, waTimeout)
	defer cancel()

	tc := &WaTurnCtx{ConversationID: t.ConversationID}
	var list []agentTool
	if a.toolsFor != nil {
		list = a.toolsFor(tc)
	}
	byName := map[string]agentTool{}
	tools := make([]anthropic.ToolUnionParam, 0, len(list))
	for _, tl := range list {
		byName[tl.Name] = tl
		props := tl.Props
		if props == nil {
			props = map[string]any{}
		}
		tools = append(tools, anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{
			Name: tl.Name, Description: anthropic.String(tl.Desc),
			InputSchema: anthropic.ToolInputSchemaParam{Properties: props, Required: tl.Req},
		}})
	}
	if len(tools) > 0 {
		tools[len(tools)-1].OfTool.CacheControl = anthropic.NewCacheControlEphemeralParam()
	}

	msgs := []anthropic.MessageParam{}
	for _, h := range t.History {
		if strings.TrimSpace(h.Text) == "" {
			continue
		}
		if h.Role == "agent" {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Text)))
		} else {
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Text)))
		}
	}
	// أول رسالة للوكيل هي القالب — النموذج يحتاج المحادثة تبدي بالزبون.
	for len(msgs) > 0 && msgs[0].Role == anthropic.MessageParamRoleAssistant {
		msgs = msgs[1:]
	}
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(t.Text)))

	system := a.system(t)
	confirmed := waConfirmed(t.Text)
	mask := &nameMask{ids: map[string]string{}, tag: map[string]int{}}

	for step := 0; step < waMaxSteps; step++ {
		if !a.takeQuota() {
			return handoffResult("الوكيل وصل حده اليومي", tc.Actions)
		}
		resp, err := a.llm.New(ctx, anthropic.MessageNewParams{
			Model: anthropic.Model(a.model), MaxTokens: waMaxTokens,
			System:   []anthropic.TextBlockParam{{Text: system, CacheControl: anthropic.NewCacheControlEphemeralParam()}},
			Messages: msgs, Tools: tools,
		})
		if err != nil {
			log.Printf("wa agent %s: %v", a.feature, err)
			return handoffResult("تعذر الذكاء: "+err.Error(), tc.Actions)
		}
		RecordAIUsage(a.feature, a.model, resp.Usage)
		msgs = append(msgs, resp.ToParam())

		var text strings.Builder
		var results []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			switch blk := block.AsAny().(type) {
			case anthropic.TextBlock:
				text.WriteString(blk.Text)
			case anthropic.ToolUseBlock:
				out, isErr := a.callTool(byName, blk, mask, tc, confirmed)
				results = append(results, anthropic.NewToolResultBlock(blk.ID, out, isErr))
			}
		}
		if tc.Handoff {
			return handoffResult(tc.HandoffReason, tc.Actions)
		}
		if len(results) == 0 || resp.StopReason != anthropic.StopReasonToolUse {
			reply := strings.TrimSpace(text.String())
			if reply == "" {
				return handoffResult("الوكيل ما عرف شنو يرد", tc.Actions)
			}
			if r := []rune(reply); len(r) > waMaxReplyLen {
				reply = strings.TrimSpace(string(r[:waMaxReplyLen])) + "…"
			}
			return &WaAgentResult{Reply: reply, Actions: tc.Actions}
		}
		msgs = append(msgs, anthropic.NewUserMessage(results...))
	}
	return handoffResult(fmt.Sprintf("الوكيل وصل حد الخطوات (%d)", waMaxSteps), tc.Actions)
}

// callTool الحراسات قبل أداة ناروتو.
func (a *waAgent) callTool(byName map[string]agentTool, blk anthropic.ToolUseBlock, mask *nameMask, tc *WaTurnCtx, confirmed bool) (string, bool) {
	tl, ok := byName[blk.Name]
	if !ok {
		return "أداة غير معروفة: " + blk.Name, true
	}
	in := map[string]any{}
	if len(blk.Input) > 0 {
		_ = json.Unmarshal(blk.Input, &in)
	}
	if waConfirmTools[blk.Name] && !confirmed {
		return "ممنوع: الزبون بعده ما أكد. اسأله سؤال تأكيد واضح (منو، شنو، شوكت) واستنى جوابه.", true
	}
	if blk.Name == "record_contact" {
		for _, k := range []string{"customerRating", "crewRating"} {
			v, ok := in[k].(float64)
			if !ok || v < 1 || v > 5 || v != float64(int(v)) {
				return "ممنوع: " + k + " لازم رقم صحيح من ١ إلى ٥ قاله الزبون نفسه — اسأله إذا ما قال.", true
			}
		}
	}
	out, err := tl.Run(in, mask)
	if err != nil {
		log.Printf("wa agent tool %s: %v", blk.Name, err)
		return "تعذر: " + err.Error(), true
	}
	tc.Actions = append(tc.Actions, blk.Name)
	if blk.Name == "handoff" && !tc.Handoff {
		// أداة ناروتو المفروض تأشر؛ احتياط إذا ما أشّرت.
		tc.Handoff = true
		if r := strArg(in, "reason"); r != "" {
			tc.HandoffReason = r
		} else {
			tc.HandoffReason = "الوكيل قرر يحوّل"
		}
	}
	if r := []rune(out); len(r) > 4000 {
		out = string(r[:4000]) + "\n…(انقص)"
	}
	return out, false
}

// waCommonRules قواعد مشتركة للوكيلين.
const waCommonRules = `
قواعد ثابتة:
- تحچي عراقي بأدب وود. رسالة وحدة قصيرة (سطرين أو ثلاثة) و**سؤال واحد بالمرة**.
- انت مساعد ذكي، ولا تدّعي إنك إنسان. إذا سألك «انت روبوت؟» جاوب بصدق.
- ممنوع: تذكر أسعار أو مبالغ، تعد بموعد مو مثبّت، تحذف شي، أو تتكلم عن موظف بسوء.
- كل حجز أو تثبيت **بعد ما الزبون يكول «اي/تمام» صريحة** على ملخص واضح انت كتبته (شنو، وين، شوكت).
- إذا الزبون طلب إنسان، أو معصّب، أو شي ما تعرفه أو خارج شغلك، أو انت مو متأكد: استعمل أداة handoff ويا السبب.
- لا تخترع معلومة عن الحجز — اعتمد على «معلومات الحجز» تحت وعلى الأدوات.`

func waContext(t WaAgentTurn) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nالوقت هسه (بغداد): %s", time.Now().In(debriefLoc).Format("2006-01-02 15:04 Monday"))
	if n := strings.TrimSpace(t.CustomerName); n != "" {
		b.WriteString("\nاسم الزبون: " + n)
	}
	if f := strings.TrimSpace(t.Facts); f != "" {
		b.WriteString("\n\nمعلومات الحجز (من النظام):\n" + f)
	}
	return b.String()
}

func waAgentName(n string) string {
	if n = strings.TrimSpace(n); n == "" {
		return "مساعد الأماني"
	}
	return n
}
