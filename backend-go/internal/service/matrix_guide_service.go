package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ عقل التوجيه — ماتركس يحچي ويا كل إجراء ═══
//
// الطبقتين:
//  1. التعليمات (MatrixGuideRule): يكتبها المدير — «الخطوات التعليمية».
//  2. النموذج (هايكو، لو المفتاح موجود): ياخذ التعليمة + موقف الموظف
//     (دوره، الشاشة، الزر، شغله الباقي) ويكتب التوجيه بنفسه.
//     بلا مفتاح، أو لو خلص الحد/فشل: التعليمة نفسها بأرقام الموظف.
//
// ⚠️ ما يطلع للنموذج اسم ولا هاتف ولا صورة — دور ومسار وأرقام بس.
// ⚠️ التوجيه كلام بس: ما يغرّم ولا ينقّص نقاط ولا يغيّر شي بالنظام.

const (
	guideCacheTTL  = 10 * time.Minute
	guideModelCap  = 400
	guideMaxTokens = 160
)

type GuideResult struct {
	Text   string `json:"text"`
	Source string `json:"source"` // RULES | MODEL | NONE
	Rule   string `json:"rule,omitempty"`
}

type MatrixGuideService struct {
	rules  *repository.MatrixGuideRepository
	watch  *MatrixAutopilotService
	client *anthropic.Client
	model  string
	mu     sync.Mutex
	used   int
	day    string
	cache  map[string]guideCached
}

type guideCached struct {
	res GuideResult
	at  time.Time
}

func NewMatrixGuideService(rules *repository.MatrixGuideRepository, watch *MatrixAutopilotService) *MatrixGuideService {
	return &MatrixGuideService{rules: rules, watch: watch, cache: map[string]guideCached{}}
}

// EnableModel يشغّل هايكو لكتابة التوجيه — يُنادى بس لو المفتاح موجود.
func (s *MatrixGuideService) EnableModel(apiKey, modelName string) {
	if apiKey == "" {
		return
	}
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	s.client, s.model = &c, modelName
}

func (s *MatrixGuideService) ModelEnabled() bool { return s.client != nil }

func routeMatches(path, route string) bool {
	if route == "/" {
		return true
	}
	return path == route || strings.HasPrefix(path, route+"/")
}

func labelMatches(label, match string) bool {
	if strings.TrimSpace(match) == "" {
		return true
	}
	for _, m := range strings.Split(match, "|") {
		if m = strings.TrimSpace(m); m != "" && strings.Contains(label, m) {
			return true
		}
	}
	return false
}

func groupMatches(group, groups string) bool {
	if strings.TrimSpace(groups) == "" {
		return true
	}
	for _, g := range strings.Split(groups, ",") {
		if strings.TrimSpace(g) == group {
			return true
		}
	}
	return false
}

// workSummary «5 بصندوق المراقب، 2 حجوزات ناقصها ورق».
func workSummary(w *WatchState) (int, string) {
	left, parts := 0, []string{}
	for _, x := range w.Workload {
		if x.Left > 0 {
			left += x.Left
			parts = append(parts, fmt.Sprintf("%d %s", x.Left, x.Label))
		}
	}
	if w.Open > 0 {
		parts = append(parts, fmt.Sprintf("%d تذكير ما انحل", w.Open))
		left += w.Open
	}
	return left, strings.Join(parts, "، ")
}

// Guide التوجيه لإجراء: الشاشة + نص الزر.
func (s *MatrixGuideService) Guide(employeeID, path, label string) (*GuideResult, error) {
	label = strings.TrimSpace(label)
	if len([]rune(label)) > 60 {
		label = string([]rune(label)[:60])
	}
	w, err := s.watch.WatchState(employeeID)
	if err != nil {
		return nil, err
	}
	left, work := workSummary(w)
	key := fmt.Sprintf("%s|%s|%s|%d|%s", employeeID, path, label, left, w.Mood)
	s.mu.Lock()
	if c, ok := s.cache[key]; ok && time.Since(c.at) < guideCacheTTL {
		s.mu.Unlock()
		r := c.res
		return &r, nil
	}
	s.mu.Unlock()

	rules, err := s.rules.Enabled()
	if err != nil {
		return nil, err
	}
	var rule *model.MatrixGuideRule
	for i := range rules {
		r := rules[i]
		if !routeMatches(path, r.Route) || !labelMatches(label, r.Match) || !groupMatches(w.Group, r.Groups) {
			continue
		}
		if r.OnlyIfPending && left == 0 {
			continue
		}
		rule = &r
		break
	}
	res := GuideResult{Source: "NONE"}
	if rule != nil {
		txt := strings.ReplaceAll(rule.Text, "{left}", fmt.Sprint(left))
		txt = strings.ReplaceAll(txt, "{work}", work)
		res = GuideResult{Text: txt, Source: "RULES", Rule: rule.ID}
		if s.client != nil && s.takeQuota() {
			if t, err := s.ask(w, path, label, rule.Text, work, left); err == nil && t != "" {
				res.Text, res.Source = t, "MODEL"
			} else if err != nil {
				log.Printf("[ai] توجيه النموذج فشل، رجعنا للتعليمة: %v", err)
			}
		}
	}
	s.mu.Lock()
	if len(s.cache) > 2000 {
		s.cache = map[string]guideCached{}
	}
	s.cache[key] = guideCached{res: res, at: time.Now()}
	s.mu.Unlock()
	return &res, nil
}

func (s *MatrixGuideService) takeQuota() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if s.day != today {
		s.day, s.used = today, 0
	}
	if s.used >= guideModelCap {
		return false
	}
	s.used++
	return true
}

const guideSystemPrompt = `أنت «ماتركس»، عين المتابعة بنظام شركة الأماني (العراق).
تحچي ويا الموظف لحظة ما يسوي إجراء بشاشة، بجملة أو جملتين بس، باللهجة العراقية البسيطة.
قواعد ما تنكسر:
- صارم ودقيق ورسمي، بلا مدح فاضي وبلا مزاح.
- استعمل الأرقام الي تنطيك إياها بالضبط، ولا تخترع رقم.
- اتبع «تعليمة المدير» — هي الأساس، وإنت تصيغها حسب الموقف.
- لا تهدد بعقوبة ولا غرامة ولا نقاط — إنت توجّه بس.
- لا تذكر أسماء أشخاص. أقصى حد ٣٠ كلمة. بلا مقدمات ولا علامات تنصيص.`

func (s *MatrixGuideService) ask(w *WatchState, path, label, rule, work string, left int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	if work == "" {
		work = "ماكو"
	}
	user := fmt.Sprintf("مجموعة الموظف: %s\nالشاشة: %s\nالزر الي ضغطه: %s\nحالة ماتركس عنه: %s\nشغله الباقي اليوم (%d): %s\nتعليمة المدير لهالموقف: %s\nاكتب التوجيه.",
		w.Group, path, label, w.Mood, left, work, rule)
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: guideMaxTokens,
		System:    []anthropic.TextBlockParam{{Text: guideSystemPrompt, CacheControl: anthropic.NewCacheControlEphemeralParam()}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(user))},
	})
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	out := strings.Trim(strings.TrimSpace(b.String()), "\"«»")
	if len([]rune(out)) > 300 {
		out = string([]rune(out)[:300])
	}
	return out, nil
}

// ── إدارة التعليمات (المدير) ──

func (s *MatrixGuideService) Rules() ([]model.MatrixGuideRule, error) { return s.rules.List() }
func (s *MatrixGuideService) CreateRule(in model.MatrixGuideRule) (*model.MatrixGuideRule, error) {
	s.flush()
	return s.rules.Create(in)
}
func (s *MatrixGuideService) UpdateRule(id string, in model.MatrixGuideRule) (*model.MatrixGuideRule, error) {
	s.flush()
	return s.rules.Update(id, in)
}
func (s *MatrixGuideService) DeleteRule(id string) error {
	s.flush()
	return s.rules.Delete(id)
}

// flush التعليمات تبدّلت — الذاكرة المؤقتة تنمسح حتى يبين التعديل فوراً.
func (s *MatrixGuideService) flush() {
	s.mu.Lock()
	s.cache = map[string]guideCached{}
	s.mu.Unlock()
}
