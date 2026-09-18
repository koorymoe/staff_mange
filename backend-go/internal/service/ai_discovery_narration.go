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

// ═══ تفسير خلية الشبكة ═══
//
// نفس فصل `ai_model_judge.go` بالضبط: النموذج يفسّر **أرقاماً محسوبة
// حتمياً**، ما يقترح محوراً ولا يبني استعلاماً. الخلية وصلت هنا بعد
// ما عبرت بوابتين (حد أدنى للعيّنة + انحراف عن المعدل) — النموذج بس
// يصوغ الجملة.
//
// ⚠️ محور الموظف لا يُرسل للنموذج أبداً — نفس قيد `buildEvidencePrompt`
// («ولا اسم موظف يطلع من هنا»). الجملة الحتمية (تشتغل محلياً بلا
// مزوّد خارجي) هي الوحيدة الي تستخدم الاسم الحقيقي، وتبقى داخل
// النظام — ما توصل لأي مزوّد أبداً.
type DiscoveryCell struct {
	MetricKey  string
	Axis       string
	ScopeID    string
	ValueLabel string
	Rate       float64
	OverallAvg float64
	Sample     int
}

type DiscoveryNarrator interface {
	Narrate(c DiscoveryCell) string
}

// DeterministicDiscoveryNarrator جملة حتمية — تشتغل دائماً، بلا شبكة.
type DeterministicDiscoveryNarrator struct{}

func (DeterministicDiscoveryNarrator) Narrate(c DiscoveryCell) string {
	dir := "أعلى من"
	if c.Rate < c.OverallAvg {
		dir = "أوطأ من"
	}
	return fmt.Sprintf(
		"%s عند %s: %.0f%% مقابل المعدل العام %.0f%% (%s المعدل بوضوح، من عيّنة %d).",
		model.AiDiscoveryMetricLabel(c.MetricKey), c.ValueLabel, c.Rate, c.OverallAvg, dir, c.Sample,
	)
}

// ModelDiscoveryNarrator يستعين بنموذج خارجي لصياغة أوضح، ويرجع
// للجملة الحتمية عند أي عثرة — نفس ضمان `ModelJudge` بالضبط.
type ModelDiscoveryNarrator struct {
	client   anthropic.Client
	model    string
	fallback DiscoveryNarrator

	mu        sync.Mutex
	dailyCap  int
	usedToday int
	resetDate string
}

func NewModelDiscoveryNarrator(apiKey, modelName string, dailyCap int) *ModelDiscoveryNarrator {
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}
	return &ModelDiscoveryNarrator{
		client:    anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:     modelName,
		fallback:  DeterministicDiscoveryNarrator{},
		dailyCap:  dailyCap,
		resetDate: time.Now().Format("2006-01-02"),
	}
}

func (n *ModelDiscoveryNarrator) takeQuota() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != n.resetDate {
		n.resetDate = today
		n.usedToday = 0
	}
	if n.usedToday >= n.dailyCap {
		return false
	}
	n.usedToday++
	return true
}

type discoveryNarrationOut struct {
	Narration string `json:"narration"`
}

var discoveryNarrationSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"narration": map[string]any{
			"type":        "string",
			"description": "جملة أو جملتين بالعربية المبسّطة تشرح الانحراف من الأرقام المعطاة حصراً، بلا اقتراح غرامة",
		},
	},
	"required":             []string{"narration"},
	"additionalProperties": false,
}

const discoveryNarrationSystemPrompt = `أنت محلّل عمليات بشركة تركيب وصيانة أنظمة (كاميرات، بصمة، ستلايت) بمدينة كربلاء.

تستلم رقماً محسوباً حتمياً: نسبة حدث معيّن عند قيمة محور معيّنة، مقابل المعدل العام لبقية القيم. مهمتك جملة أو جملتين تشرح الانحراف بوضوح لمدير يقرأها بسرعة.

قواعد:
١. اعتمد على الأرقام المعطاة حصراً — لا تخترع سبباً غير مذكور.
٢. لا تقترح غرامة ولا عقوبة — هذا وصف نمط، مو حكم على شخص.
٣. اكتب بالعربية المبسّطة، بلا مصطلحات تقنية.`

// Narrate يصوغ الجملة، ويرجع للحتمية عند أي عثرة أو تجاوز الحد اليومي.
func (n *ModelDiscoveryNarrator) Narrate(c DiscoveryCell) string {
	if c.Axis == model.AiAxisEmployee {
		// ⚠️ صفر إرسال لأي بيانات موظف لمزوّد خارجي — الجملة الحتمية بس.
		return n.fallback.Narrate(c)
	}
	if !n.takeQuota() {
		return n.fallback.Narrate(c)
	}
	out, err := n.ask(c)
	if err != nil || out == "" {
		if err != nil {
			log.Printf("[ai-discovery] النموذج ما جاوب: %v", err)
		}
		return n.fallback.Narrate(c)
	}
	return out
}

func (n *ModelDiscoveryNarrator) ask(c DiscoveryCell) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(
		"المقياس: %s\nالمحور: %s\nالقيمة: %s\nالنسبة عند هذي القيمة: %.1f%%\nالمعدل العام لبقية القيم: %.1f%%\nحجم العيّنة: %d",
		model.AiDiscoveryMetricLabel(c.MetricKey), model.AiAxisLabel(c.Axis), c.ValueLabel, c.Rate, c.OverallAvg, c.Sample,
	)

	resp, err := n.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(n.model),
		MaxTokens: 512,
		System: []anthropic.TextBlockParam{{
			Text:         discoveryNarrationSystemPrompt,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}},
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{Schema: discoveryNarrationSchema},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", err
	}
	var text strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			text.WriteString(tb.Text)
		}
	}
	if text.Len() == 0 {
		return "", fmt.Errorf("رد فاضي")
	}
	var out discoveryNarrationOut
	if err := json.Unmarshal([]byte(text.String()), &out); err != nil {
		return "", fmt.Errorf("رد مو مقروء: %w", err)
	}
	return out.Narration, nil
}
