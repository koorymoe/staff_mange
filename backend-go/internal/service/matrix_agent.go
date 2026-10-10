package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jmoiron/sqlx"
)

// ═══ ماتركس الوكيل — هايكو هو العقل (قرار (ع) 10-10) ═══
//
// «أريد هايكو هو الي يشتغل كلشي: يفحص بعينه ويحكم». قبل، الكود يجمع ملخص
// ثابت وهايكو يصيغه بس. هسه هايكو ياخذ «أدوات» (عيون على النظام) ويقرر
// بنفسه شنو يفحص وبأي ترتيب، وبعدين يحكم.
//
// السياج يبقى بالكود ولا ينكسر:
//   - الأسماء ما تطلع: كل اسم موظف يتبدّل بـ«موظف#n» قبل ما يوصل النموذج،
//     ويرجع بالخادم. أسماء الزبائن ما تدخل الأدوات أصلاً.
//   - المرحلة الأولى: أدوات قراءة بس — ما يكتب ولا يغيّر شي.
//   - حد خطوات وحد توكنات وعدّاد كلفة؛ وبلا مفتاح يرجع المحرّك القديم.

const (
	agentMaxSteps  = 12
	agentMaxTokens = 1200
	agentTimeout   = 60 * time.Second
)

// agentLLM واجهة النموذج — حتى الاختبار يركّب نموذجاً وهمياً.
type agentLLM interface {
	New(ctx context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error)
}

type anthropicLLM struct{ c *anthropic.Client }

func (a anthropicLLM) New(ctx context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	return a.c.Messages.New(ctx, p)
}

// agentTool أداة: وصف + مخطط مدخلات + تنفيذ يرجّع نصاً (بالأسماء الحقيقية —
// الإخفاء يصير مركزياً بالحلقة، حتى ما تنسى أداة تخفي).
type agentTool struct {
	Name  string
	Desc  string
	Props map[string]any
	Req   []string
	Run   func(in map[string]any, names *nameMask) (string, error)
}

type MatrixAgent struct {
	db    *sqlx.DB
	llm   agentLLM
	model string
	tools []agentTool
	// testNames للاختبار بلا قاعدة (اسم ← معرّف)
	testNames map[string]string
}

func NewMatrixAgent(db *sqlx.DB, client *anthropic.Client, model string) *MatrixAgent {
	if model == "" {
		model = "claude-haiku-4-5"
	}
	a := &MatrixAgent{db: db, model: model}
	if client != nil {
		a.llm = anthropicLLM{c: client}
	}
	return a
}

// Enabled الوكيل يشتغل؟ (مفتاح موجود + مفتاح الكلفة مفتوح)
func (a *MatrixAgent) Enabled(feature string) bool {
	return a != nil && a.llm != nil && AIFeatureOn(feature)
}

func (a *MatrixAgent) AddTool(t agentTool) { a.tools = append(a.tools, t) }

// ── إخفاء الأسماء ──

// nameMask يبدّل أسماء الموظفين برموز ثابتة طول التشغيلة.
type nameMask struct {
	names []string // مرتّبة الأطول أولاً
	ids   map[string]string
	tag   map[string]int // اسم ← رقم
	order []string       // رقم-1 ← اسم
}

func (a *MatrixAgent) newMask() *nameMask {
	m := &nameMask{ids: map[string]string{}, tag: map[string]int{}}
	rows := []struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}{}
	if a.db != nil {
		_ = a.db.Select(&rows, `SELECT id, name FROM "Employee" WHERE COALESCE(name, '') <> ''`)
	}
	for n, id := range a.testNames {
		rows = append(rows, struct {
			ID   string `db:"id"`
			Name string `db:"name"`
		}{id, n})
	}
	for _, r := range rows {
		n := strings.TrimSpace(r.Name)
		if len([]rune(n)) < 2 {
			continue
		}
		if _, dup := m.ids[n]; !dup {
			m.names = append(m.names, n)
		}
		m.ids[n] = r.ID
	}
	sort.SliceStable(m.names, func(i, j int) bool { return len([]rune(m.names[i])) > len([]rune(m.names[j])) })
	return m
}

func (m *nameMask) tagFor(name string) string {
	n, ok := m.tag[name]
	if !ok {
		m.order = append(m.order, name)
		n = len(m.order)
		m.tag[name] = n
	}
	return fmt.Sprintf("موظف#%d", n)
}

// Hide يبدّل كل اسم معروف برمزه.
func (m *nameMask) Hide(text string) string {
	for _, n := range m.names {
		if !strings.Contains(text, n) {
			continue
		}
		if len([]rune(n)) >= 4 {
			text = strings.ReplaceAll(text, n, m.tagFor(n))
			continue
		}
		// الاسم القصير (Hr، علي) ينبدّل بس إذا كلمة كاملة — مو جزء من كلمة ثانية
		re := regexp.MustCompile(`(^|[^\p{L}])` + regexp.QuoteMeta(n) + `($|[^\p{L}])`)
		text = re.ReplaceAllStringFunc(text, func(s string) string {
			return strings.Replace(s, n, m.tagFor(n), 1)
		})
	}
	return text
}

var agentTagRe = regexp.MustCompile(`موظف#(\d+)`)

// Show يرجّع الأسماء الحقيقية بجواب النموذج.
func (m *nameMask) Show(text string) string {
	return agentTagRe.ReplaceAllStringFunc(text, func(s string) string {
		var n int
		fmt.Sscanf(strings.TrimPrefix(s, "موظف#"), "%d", &n)
		if n >= 1 && n <= len(m.order) {
			return m.order[n-1]
		}
		return s
	})
}

// Resolve رمز أو اسم ← معرّف الموظف (للأدوات الي تاخذ «منو»).
func (m *nameMask) Resolve(ref string) (string, string) {
	ref = strings.TrimSpace(ref)
	if sm := agentTagRe.FindStringSubmatch(ref); sm != nil {
		var n int
		fmt.Sscanf(sm[1], "%d", &n)
		if n >= 1 && n <= len(m.order) {
			name := m.order[n-1]
			return m.ids[name], name
		}
	}
	if id, ok := m.ids[ref]; ok {
		return id, ref
	}
	for _, n := range m.names {
		if strings.Contains(n, ref) || strings.Contains(ref, n) {
			return m.ids[n], n
		}
	}
	return "", ""
}

// ── الحلقة ──

// AgentRun نتيجة تشغيلة: الجواب (بالأسماء الحقيقية) + الأدوات الي استعملها.
type AgentRun struct {
	Answer string   `json:"answer"`
	Steps  []string `json:"steps"`
}

var ErrAgentOff = errors.New("الوكيل مطفي")

// Run يشغّل الوكيل على طلب: هايكو يقرر الأدوات، والكود ينفذ ويخفي ويرجّع.
func (a *MatrixAgent) Run(feature, system, task string) (*AgentRun, error) {
	if !a.Enabled(feature) {
		return nil, ErrAgentOff
	}
	mask := a.newMask()
	tools := make([]anthropic.ToolUnionParam, 0, len(a.tools))
	byName := map[string]agentTool{}
	for _, t := range a.tools {
		byName[t.Name] = t
		props := t.Props
		if props == nil {
			props = map[string]any{}
		}
		tools = append(tools, anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{
			Name: t.Name, Description: anthropic.String(t.Desc),
			InputSchema: anthropic.ToolInputSchemaParam{Properties: props, Required: t.Req},
		}})
	}
	if len(tools) > 0 {
		tools[len(tools)-1].OfTool.CacheControl = anthropic.NewCacheControlEphemeralParam()
	}
	msgs := []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(mask.Hide(task)))}
	ctx, cancel := context.WithTimeout(context.Background(), agentTimeout)
	defer cancel()
	run := &AgentRun{}
	for step := 0; step < agentMaxSteps; step++ {
		resp, err := a.llm.New(ctx, anthropic.MessageNewParams{
			Model: anthropic.Model(a.model), MaxTokens: agentMaxTokens,
			System:   []anthropic.TextBlockParam{{Text: system, CacheControl: anthropic.NewCacheControlEphemeralParam()}},
			Messages: msgs, Tools: tools,
		})
		if err != nil {
			return nil, err
		}
		RecordAIUsage(feature, a.model, resp.Usage)
		msgs = append(msgs, resp.ToParam())
		var results []anthropic.ContentBlockParamUnion
		var text strings.Builder
		for _, block := range resp.Content {
			switch b := block.AsAny().(type) {
			case anthropic.TextBlock:
				text.WriteString(b.Text)
			case anthropic.ToolUseBlock:
				out, isErr := a.callTool(byName, b, mask)
				run.Steps = append(run.Steps, b.Name)
				results = append(results, anthropic.NewToolResultBlock(b.ID, out, isErr))
			}
		}
		if len(results) == 0 || resp.StopReason != anthropic.StopReasonToolUse {
			run.Answer = mask.Show(strings.TrimSpace(text.String()))
			return run, nil
		}
		msgs = append(msgs, anthropic.NewUserMessage(results...))
	}
	return nil, fmt.Errorf("الوكيل وصل حد الخطوات (%d) بلا جواب", agentMaxSteps)
}

// callTool ينفذ أداة ويرجّع مخرجاتها مخفية الأسماء ومقصوصة.
func (a *MatrixAgent) callTool(byName map[string]agentTool, b anthropic.ToolUseBlock, mask *nameMask) (string, bool) {
	t, ok := byName[b.Name]
	if !ok {
		return "أداة غير معروفة: " + b.Name, true
	}
	in := map[string]any{}
	if len(b.Input) > 0 {
		_ = json.Unmarshal(b.Input, &in)
	}
	out, err := t.Run(in, mask)
	if err != nil {
		log.Printf("matrix agent tool %s: %v", b.Name, err)
		return "تعذر: " + err.Error(), true
	}
	out = mask.Hide(out)
	if r := []rune(out); len(r) > 6000 {
		out = string(r[:6000]) + "\n…(انقص)"
	}
	return out, false
}

func strArg(in map[string]any, k string) string {
	if v, ok := in[k].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}
