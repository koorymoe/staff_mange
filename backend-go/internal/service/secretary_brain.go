package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"staffmange-api/internal/model"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jmoiron/sqlx"
)

// ═══ 🗂️ عقل «السكرتير» — قرار دازاي 10-22 (العقد: docs/secretary.md) ═══
//
// عقل + تعليمات صفحة وحدة + أدوات. هو يقرر أي أداة، مو قواعد مبرمجة.
// الأدوات والمسارات والتخزين عند ناروتو (secretary_tools.go وsecretary_handler.go)؛
// هذا الملف العقل بس.
//
// ⚠️ الأسماء مفتوحة هنا — استثناء مكتوب من دازاي **للسكرتير وحده**. ماتركس
// وغيره يبقون على nameMask. لهذا الحلقة هنا منفصلة عن MatrixAgent.ChatWith
// (الي تخفي) بدل ما نضيف علم «لا تخفي» يكدر ينفتح بالغلط هناك.
//
// اتفاق مخرجات الأدوات: الأداة ترجّع نص عادي، أو JSON:
//
//	{"facts": "نص للنموذج", "cards": [SecretaryCard…], "proposal": {SecretaryProposal}}
//
// العقل ينطي النموذج facts، ويجمع البطاقات (بلا تكرار، حد ٦) وآخر مقترح.

const (
	secretaryMaxSteps  = 8
	secretaryMaxTokens = 900
	secretaryTimeout   = 60 * time.Second
	secretaryMaxCards  = 6
)

var (
	ErrSecretaryOff = errors.New("السكرتير مطفي — ماكو مفتاح ذكاء أو المفتاح مسدود")
	ErrSecretaryCap = errors.New("ايريس وصلت حدها اليومي — ترجع باچر، والملخص وينتظر قرارك يشتغلون عادي")
)

// الأنواع المشتركة (SecretaryTurn · SecretaryCard · SecretaryProposal · SecretaryResult) بـsecretary_types.go.

// secretaryToolOut الشكل المنظم الاختياري لمخرجات الأداة.
type secretaryToolOut struct {
	Facts    string             `json:"facts"`
	Cards    []SecretaryCard    `json:"cards"`
	Proposal *SecretaryProposal `json:"proposal"`
}

type SecretaryBrain struct {
	db    *sqlx.DB
	llm   agentLLM
	model string
	tools []agentTool
	// toolsFor أدوات ناروتو لكل دور: تكتب بطاقاتها ومقترحها على tc (secretary_tools.go).
	toolsFor func(tc *SecretaryTurnCtx) []agentTool

	capMu    sync.Mutex
	dailyCap int
	capDay   string
	used     int

	webSearch bool // بحث الويب (SetWebSearch)
}

// NewSecretaryBrain client = nil يعني بلا ذكاء (Reply يرجّع ErrSecretaryOff).
// dailyCap <= 0 = بلا سقف.
func NewSecretaryBrain(db *sqlx.DB, client *anthropic.Client, model string, tools []agentTool, dailyCap int) *SecretaryBrain {
	if model == "" {
		model = "claude-haiku-4-5"
	}
	b := &SecretaryBrain{db: db, model: model, tools: tools, dailyCap: dailyCap}
	if client != nil {
		b.llm = anthropicLLM{c: client}
	}
	return b
}

// NewSecretaryBrainFor العقل فوق أدوات السكرتير الحقيقية (main.go).
func NewSecretaryBrainFor(d SecretaryDeps, dailyCap int) *SecretaryBrain {
	b := NewSecretaryBrain(d.DB, d.Client, d.Model, nil, dailyCap)
	b.toolsFor = func(tc *SecretaryTurnCtx) []agentTool { return NewSecretaryAgent(d, tc).tools }
	return b
}

// SetWebSearch يفعّل بحث الويب لايريس.
func (b *SecretaryBrain) SetWebSearch(on bool) { b.webSearch = on }

// Enabled السكرتير يكدر يفكر هسه؟
func (b *SecretaryBrain) Enabled() bool {
	return b != nil && b.llm != nil && AIFeatureOn("SECRETARY")
}

// takeQuota كل نداء للنموذج ينحسب — حارس فاتورة باليوم (بغداد).
func (b *SecretaryBrain) takeQuota() bool {
	if b.dailyCap <= 0 {
		return true
	}
	b.capMu.Lock()
	defer b.capMu.Unlock()
	day := time.Now().In(debriefLoc).Format("2006-01-02")
	if day != b.capDay {
		b.capDay, b.used = day, 0
	}
	if b.used >= b.dailyCap {
		return false
	}
	b.used++
	return true
}

const secretaryInstructions = `انت «ايريس» — سكرتيرة مدير شركة الأماني (طاقة شمسية، كهرباء، شبكات، كاميرات). تشتغل لـ%s.
- تفهم العراقي وتجاوب بالعراقي، بجملة أو جملتين. التفصيل بس إذا طلبه.
- إنتِ نفس عقل ماتركس (تشوفين كل النظام بأدواته) بس شغلچ غير: سكرتيرة تسولف وياه.
- الترتيب: ممنوع ** و# وعلامات ماركداون (الشاشة ما تعرضها). استعمل أسطر قصيرة تبدي بـ«- » وأرقام واضحة. وفلوس الموظفين من revenue_by_responsible بس: كل حجز إله مسؤول واحد، لا تكرر المبلغ على كل الكادر.
- أذنچ تتعلم: إذا صحّحلچ كلمة («قصدي…»، «مو هيچ، هي…») أو حچى كلمة غريبة ما تعرفيها، اسأليه شنو معناها بسطر، ومن يجاوب احفظيها بـlearn_term (ويا الشكل الي انكتب غلط إذا اكو) — حتى المرة الجاية تنكتب صح.
- رأيك التنفيذي بأي شخص («شنو رأيك بفلان؟»، «منو الأحسن بوحدة كذا؟»): اسحبي حقائقه بـexecutive_review أول، وابني رأيك على الأرقام بس (تقدير ١-٥، قوة، ضعف مربوط بفقرة اللائحة، توصية عملية). لا تحكمين على شخصيته أو نيته. احفظيه بـsave_executive_opinion إذا طلب.
- «عمّم/وجّه الكل/خلي وحدة كذا تلتزم بـ…» = توجيه رسمي: صدّريه بـissue_directive (النطاق المناسب، ومربوط بفقرة اللائحة) — يصير سياسة عمل ملزمة ويوصل إشعار للكل بنطاقه.
- لا تگولين «ما أگدر» وتوقفين: إذا ماكو أداة للطلب، سوّي أقرب شي يحققه بأدواتچ (تذكير، مهمة، ملف، مسودة)، وسجّلي الناقص بـsuggest_improvement حتى ينبنى.
- جاوبي على قد رسالته بالضبط. تحية أو سلام («هلو»، «شلونك») = رد تحية قصير بس، مثل «هلا سيدنا، شلون أگدر أساعدك؟» — بلا معلومات ولا اقتراحات ولا تصحيحات.
- الاقتراحات من نفسچ بس إذا سأل سؤال مفتوح («شنو عندي؟»، «شلون اليوم؟»، «شنو المهم؟») أو اكو شي عاجل فعلاً يخص نفس الي سأل عنه.
- إذا جواب قديم طلع غلط، صححيه بس لما يرجع لنفس الموضوع — مو بنص تحية أو موضوع ثاني.
- فكّري بالإنكليزي (التفكير يطلع إله)، والجواب بالعراقي.
- إذا لاحظتي نقص حقيقي (سؤال ما تگدرين تجاوبينه بأدواتچ، طلب يتكرر)، سجّليه بـsuggest_improvement وكمّلي.
- كل رقم أو حالة أو اسم يجي من الأدوات. لا تخمّن ولا تكمّل من عندك. إذا ما لكيت، كول ما لكيت وشنو تحتاج منه.
- دوّر بنفسك: اختار الأداة المناسبة، وإذا الجواب يحتاج أكثر من أداة استعملها.
- الموظفين والزبائن شي، والحجوزات شي: لكل وحدة أداة. موظف/مهندس/فني/ليدر ← find_employee. زبون ← find_customer. حجز أو «حجز فلان» ← find_booking. مشروع ← find_project. find_any بس إذا ما واضح أبداً، وإذا طلع الاسم موظف وزبون سوه اسأله «تقصد الموظف لو الزبون؟». اسم الزبون ما تنسبه لموظف أبداً، ولا العكس.
- عندك وصول قراءة لكل النظام: إذا ماكو أداة جاهزة للسؤال، استعمل describe_data ثم query_data. لا تكول «ما أكدر أوصل للنظام» قبل ما تجرب. وإذا سأل عن نفسه استعمل whoami.
- إذا طلب يوجّه شغلة لموظف: نفّذها بأداة propose_task — تنرسل فوراً. حوّل كلامه لموعد (باچر الصبح ← باچر 10:00). إذا الأداة كالت ناقص شي، اسأله بجملة وحدة بالكلام — بلا قوائم.
- اشتغل عنه بالكامل مثل سكرتير حقيقي: لا تكله «روح للشاشة الفلانية» إذا تكدر تجيب الجواب بأداة. دوّر بأكثر من أداة إذا يحتاج.
- إذا سأل عن شي عام برّا الشركة (سعر، خبر، معلومة، مكان، شخص مشهور): ابحث بالويب. وهنا انتبه زين:
  · جاوب **بس** بالي مكتوب صراحة بالمصادر الي لكيتها — لا تكمّل من ذاكرتك ولا تخمّن أعداد أو أسماء أو تواريخ.
  · إذا المصادر تختلف أو ناقصة، كول هالشي بصراحة («المصادر تكول كذا، بس ما متأكدة من كذا»).
  · الأرقام والأسماء والتواريخ: إذا ما لكيتها بمصدر واضح، كول ما لكيتها — أحسن من معلومة غلط.
  · اذكر اسم المصدر بالآخر.
- ممنوع: أي تعديل فلوس، حذف، تغيير منصب أو صلاحية، غرامة. إذا طلبها، كله هاي يسويها من شاشتها بالنظام.
- البطاقات تطلع تحت جوابك تلقائياً، فلا تعيد كل تفاصيلها بالنص — قول الخلاصة.
- لا مقدمات ولا مجاملات طويلة.`

func (b *SecretaryBrain) system(t SecretaryTurn) string {
	name := strings.TrimSpace(t.UserName)
	if name == "" {
		name = "المدير"
	}
	var s strings.Builder
	fmt.Fprintf(&s, secretaryInstructions, name)
	// (ع) 10-23: «من يخاطبه يكله أهلاً دكتور… ومن يحاچي يكله سيدنا» — اللقب من ملفه.
	title := strings.TrimSpace(t.UserTitle)
	if title == "" {
		title = "سيدنا"
	}
	if role := strings.TrimSpace(t.UserRole); role != "" {
		fmt.Fprintf(&s, "\n- انت تحچي هسه ويا %s، %s شركة الأماني. إذا سألك «منو آني؟» جاوبه باسمه ومنصبه.", name, role)
	}
	fmt.Fprintf(&s, "\n- الاحترام: ناديه «%s» — الترحيب مثل «أهلاً %s، بشنو أكدر أساعدك اليوم؟»، وبالحچي خاطبه «%s» أو «سيدنا». لا تناديه باسمه المجرّد أبداً.", title, title, title)
	s.WriteString("\nالوقت هسه (بغداد): " + time.Now().In(debriefLoc).Format("2006-01-02 15:04 Monday"))
	// (10-30) الهيكل الجديد ولائحة العمل — النقطة ٢ «الذكاء يفهم النظام»
	s.WriteString(OrgBrief(b.db, t.UserID))
	if sm := strings.TrimSpace(t.Summary); sm != "" {
		s.WriteString("\n\nذاكرة هالمحادثة (ملخص الي صار قبل آخر الرسائل):\n" + sm + "\n")
	}
	if len(t.Prefs) > 0 {
		s.WriteString("\n\nتفضيلات المدير (معتمدة منه — امشي عليها):\n")
		for _, p := range t.Prefs {
			if p = strings.TrimSpace(p); p != "" {
				s.WriteString("- " + p + "\n")
			}
		}
	}
	return s.String()
}

func (b *SecretaryBrain) mask() *nameMask {
	if b.db == nil {
		return &nameMask{ids: map[string]string{}, tag: map[string]int{}}
	}
	// للـResolve بس (الأدوات الي تاخذ «منو») — ما ننادي Hide هنا.
	return (&MatrixAgent{db: b.db}).newMask()
}

// Reply جواب رسالة وحدة: النموذج يقرر الأدوات، والعقل يجمع البطاقات والمقترح.
func (b *SecretaryBrain) Reply(ctx context.Context, t SecretaryTurn) (*SecretaryResult, error) {
	if !b.Enabled() {
		return nil, ErrSecretaryOff
	}
	if strings.TrimSpace(t.Text) == "" {
		return &SecretaryResult{Reply: "ما وصلني شي — اكتب أو احچي مرة ثانية."}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, secretaryTimeout)
	defer cancel()

	turnTools := b.tools
	var tc *SecretaryTurnCtx
	if b.toolsFor != nil {
		tc = &SecretaryTurnCtx{UserID: t.UserID, Request: t.Text}
		turnTools = b.toolsFor(tc)
	}
	byName := map[string]agentTool{}
	tools := make([]anthropic.ToolUnionParam, 0, len(turnTools))
	for _, tl := range turnTools {
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
	// 🔎 بحث الويب (دازاي 10-23: «اكله ابحثلي ع كوكل يكلي ما اكدر») — أداة Anthropic نفسها،
	// لحد ٣ بحثات بالسؤال. للمعلومات العامة بس؛ بيانات الشركة من أدوات النظام.
	if b.webSearch {
		tools = append(tools, anthropic.ToolUnionParam{OfWebSearchTool20250305: &anthropic.WebSearchTool20250305Param{
			MaxUses: anthropic.Int(3)}})
	}

	msgs := []anthropic.MessageParam{}
	for _, h := range t.History {
		if strings.TrimSpace(h.Text) == "" {
			continue
		}
		if h.Role == "assistant" {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Text)))
		} else {
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Text)))
		}
	}
	// المحادثة لازم تبدي بالمستخدم وتتناوب — نشيل أي assistant بالبداية.
	for len(msgs) > 0 && msgs[0].Role == anthropic.MessageParamRoleAssistant {
		msgs = msgs[1:]
	}
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(t.Text)))

	res := &SecretaryResult{}
	seen := map[string]bool{}
	mask := b.mask()
	system := b.system(t)
	// (دازاي 10-29) «هلو حبي شونك» ما يحتاج تفكير ٤٠ ثانية — التفكير للطلبات الطويلة بس.
	think := t.Emit != nil && len([]rune(t.Text)) > 80
	emit := func(e AgentEvent) {
		if t.Emit != nil {
			t.Emit(e)
		}
	}

	for step := 0; step < secretaryMaxSteps; step++ {
		if !b.takeQuota() {
			return nil, ErrSecretaryCap
		}
		p := anthropic.MessageNewParams{
			Model: anthropic.Model(b.model), MaxTokens: secretaryMaxTokens,
			System:   []anthropic.TextBlockParam{{Text: model.Rebrand(system), CacheControl: anthropic.NewCacheControlEphemeralParam()}},
			Messages: msgs, Tools: tools,
		}
		resp, err := llmStep(ctx, b.llm, &p, t.Emit != nil, &think, emit, mask.Show)
		if err != nil {
			return nil, err
		}
		RecordAIUsage("SECRETARY", b.model, resp.Usage)
		msgs = append(msgs, resp.ToParam())

		var text strings.Builder
		var results []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			switch blk := block.AsAny().(type) {
			case anthropic.TextBlock:
				text.WriteString(blk.Text)
			case anthropic.ToolUseBlock:
				emit(AgentEvent{Type: "tool", Name: blk.Name})
				out, isErr := b.callTool(byName, blk, mask, res, seen)
				results = append(results, anthropic.NewToolResultBlock(blk.ID, out, isErr))
			}
		}
		// بحث الويب الطويل يوكّف الدور (pause_turn) — نكمل من وين وكف.
		if len(results) == 0 && resp.StopReason == anthropic.StopReasonPauseTurn {
			continue
		}
		if len(results) == 0 || resp.StopReason != anthropic.StopReasonToolUse {
			if tc != nil {
				for _, c := range tc.Cards {
					if key := c.Kind + "|" + c.Title; !seen[key] && len(res.Cards) < secretaryMaxCards {
						seen[key] = true
						res.Cards = append(res.Cards, c)
					}
				}
				if tc.Proposal != nil {
					res.Proposal = tc.Proposal
				}
			}
			res.Reply = stripMarkdown(strings.TrimSpace(text.String()))
			if res.Reply == "" && res.Proposal != nil {
				res.Reply = res.Proposal.Summary
			}
			return res, nil
		}
		msgs = append(msgs, anthropic.NewUserMessage(results...))
	}
	return nil, fmt.Errorf("السكرتير وصل حد الخطوات (%d) بلا جواب", secretaryMaxSteps)
}

// callTool ينفذ الأداة ويفك مخرجاتها المنظمة إذا موجودة.
func (b *SecretaryBrain) callTool(byName map[string]agentTool, blk anthropic.ToolUseBlock, mask *nameMask,
	res *SecretaryResult, seen map[string]bool) (string, bool) {
	tl, ok := byName[blk.Name]
	if !ok {
		return "أداة غير معروفة: " + blk.Name, true
	}
	in := map[string]any{}
	if len(blk.Input) > 0 {
		_ = json.Unmarshal(blk.Input, &in)
	}
	out, err := tl.Run(in, mask)
	if err != nil {
		log.Printf("secretary tool %s: %v", blk.Name, err)
		return "تعذر: " + err.Error(), true
	}
	if s := strings.TrimSpace(out); strings.HasPrefix(s, "{") {
		var o secretaryToolOut
		if json.Unmarshal([]byte(s), &o) == nil && (o.Facts != "" || len(o.Cards) > 0 || o.Proposal != nil) {
			for _, c := range o.Cards {
				key := c.Kind + "|" + c.Title
				if seen[key] || len(res.Cards) >= secretaryMaxCards {
					continue
				}
				seen[key] = true
				res.Cards = append(res.Cards, c)
			}
			if o.Proposal != nil {
				res.Proposal = o.Proposal
			}
			out = o.Facts
			if out == "" && o.Proposal != nil {
				out = "المسودة جاهزة وتنتظر «إرسال» منه: " + o.Proposal.Summary
			}
			if out == "" {
				out = fmt.Sprintf("%d بطاقة جاهزة للعرض", len(o.Cards))
			}
		}
	}
	if r := []rune(out); len(r) > 6000 {
		out = string(r[:6000]) + "\n…(انقص)"
	}
	return out, false
}

var secretaryBulletRe = regexp.MustCompile(`^(?:[-•*]+|[0-9٠-٩]+[.)])\s*`)

const secretarySummaryInstructions = `انت «ايريس»، سكرتيرة %s. هاي أرقام اليوم الخام من النظام. اكتب ملخص اليوم من ٣ إلى ٦ أسطر قصيرة بالعراقي — الأهم أول.
كل سطر حقيقة وحدة برقمها من الأرقام بس. لا تضيف شي مو موجود. سطر بلا رقم أو حالة ما يفيد.
رجّع الأسطر بس، كل سطر بسطر جديد، بلا ترقيم ولا مقدمة.`

// Summarize يصيغ «ملخص اليوم» فوق أرقام today_facts. إذا فشل، ناروتو يعرض الخام.
func (b *SecretaryBrain) Summarize(ctx context.Context, userName, facts string) ([]string, error) {
	if !b.Enabled() {
		return nil, ErrSecretaryOff
	}
	if strings.TrimSpace(facts) == "" {
		return nil, nil
	}
	if !b.takeQuota() {
		return nil, ErrSecretaryCap
	}
	if strings.TrimSpace(userName) == "" {
		userName = "المدير"
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	resp, err := b.llm.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.Model(b.model), MaxTokens: 500,
		System:   []anthropic.TextBlockParam{{Text: model.Rebrand(fmt.Sprintf(secretarySummaryInstructions, userName))}},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(facts))},
	})
	if err != nil {
		return nil, err
	}
	RecordAIUsage("SECRETARY", b.model, resp.Usage)
	var text strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			text.WriteString(tb.Text)
		}
	}
	var lines []string
	for _, l := range strings.Split(text.String(), "\n") {
		// نشيل علامة القائمة بس («- » «• » «١. ») — مو الأرقام نفسها
		l = strings.TrimSpace(secretaryBulletRe.ReplaceAllString(strings.TrimSpace(l), ""))
		if l != "" {
			lines = append(lines, l)
		}
		if len(lines) == 6 {
			break
		}
	}
	return lines, nil
}

// AIErrorText يترجم رفض مزوّد الذكاء لرسالة يفهمها المالك (دازاي 10-23: «نضيف المفاتيح»)
// — حتى مفتاح غلط أو رصيد خالص ما يطلع «خلل» عام وهو ما يدري شنو يصلّح.
func AIErrorText(err error) string {
	if errors.Is(err, ErrAIBudget) || strings.Contains(err.Error(), ErrAIBudget.Error()) {
		return "💵 ايريس وصلت حد صرفها اليومي — ترجع باچر. (الحد: AI_IRIS_DAILY_USD بملف .env)"
	}
	var apiErr *anthropic.Error
	if !errors.As(err, &apiErr) {
		return ""
	}
	msg := strings.ToLower(apiErr.Error())
	switch {
	case apiErr.StatusCode == 401 || apiErr.StatusCode == 403:
		return "🔑 مفتاح ايريس مرفوض — تأكد من AI_KEY_SECRETARY (أو ANTHROPIC_API_KEY) بملف .env بالسيرفر."
	case strings.Contains(msg, "credit") || strings.Contains(msg, "billing") || apiErr.StatusCode == 402:
		return "💳 رصيد الذكاء خلص — اشحن من console.anthropic.com ← Billing."
	case apiErr.StatusCode == 404:
		return "🤖 اسم الموديل غلط — تأكد من AI_MODEL بملف .env."
	case apiErr.StatusCode == 429 || apiErr.StatusCode == 529:
		return "⏳ الذكاء مضغوط هسه — جرّب بعد دقيقة."
	}
	return ""
}

// NewAnthropicClient عميل بمفتاح وكيل معيّن — nil إذا المفتاح فاضي (الميزة تنطفي، النظام يبقى).
func NewAnthropicClient(key string) *anthropic.Client { return NewAnthropicClientFor(AIAgentIris, key) }

// NewAnthropicClientFor عميل بمفتاح وكيل وحده اليومي بالدولار.
func NewAnthropicClientFor(agent, key string) *anthropic.Client {
	if key == "" {
		return nil
	}
	c := NewAIClient(agent, key)
	return &c
}

// SecretaryHonorific لقب المستخدم من «الشهادة» بملفه (دازاي 10-23).
func SecretaryHonorific(certificate *string) string {
	if certificate == nil {
		return "سيدنا"
	}
	c := strings.ToLower(*certificate)
	has := func(words ...string) bool {
		for _, w := range words {
			if strings.Contains(c, w) {
				return true
			}
		}
		return false
	}
	switch {
	case has("دكتوراه", "دكتوراة", "دكتوراء", "دكتور", "phd", "ph.d", "doctor"):
		return "دكتور"
	case has("ماجستير", "master", "msc", "m.sc"):
		return "أستاذ"
	case has("هندسة", "مهندس", "engineer", "b.eng"):
		return "مهندس"
	}
	return "سيدنا"
}

// Remember يحدّث ذاكرة محادثة ايريس (ملخص + رسائل جديدة) — نداء رخيص بلا أدوات.
func (b *SecretaryBrain) Remember(ctx context.Context, prev string, turns []AgentTurn) (string, error) {
	if !b.Enabled() {
		return "", ErrSecretaryOff
	}
	return Remember(ctx, b.llm, b.model, "SECRETARY", prev, turns)
}
