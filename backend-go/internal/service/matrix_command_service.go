package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"staffmange-api/internal/repository"
)

// ═══ مركز قيادة ماتركس — شاشة المدير الرئيسية ═══
//
// كل رقم بالشاشة من بيانات حقيقية؛ لو البيانات ما تكفي نگول «ما تكفي» بدل
// رقم وهمي. و«اسأل ماتركس»: قراءة بس — ما يغيّر شي، والأسماء تتبدّل بأرقام
// قبل ما توصل النموذج وترجع بالخادم.

const askDailyLimit = 30

var ErrAskLimit = errors.New("وصلت حد الأسئلة اليوم (٣٠) — باچر يرجع")

type MatrixCommandService struct {
	repo     *repository.MatrixCommandRepository
	watch    *MatrixAutopilotService
	reports  *MatrixEmployeeReportService
	business *MatrixBusinessService
	props    *repository.MatrixProposalRepository
	client   *anthropic.Client
	model    string
	agent    *MatrixAgent
}

// SetAgent يربط ماتركس الوكيل — «اسأل ماتركس» يصير يفتش بنفسه بالأدوات.
func (s *MatrixCommandService) SetAgent(a *MatrixAgent) { s.agent = a }

// Client للوكيل — نفس عميل هايكو ونفس الموديل.
func (s *MatrixCommandService) Client() (*anthropic.Client, string) { return s.client, s.model }

const agentAskPrompt = `أنت «ماتركس»، عقل المتابعة والرقابة بشركة الأماني (العراق). المالك أو المدير يسألك أو يطلب فحص.
عندك أدوات تفتح بيها النظام وتشوف بعينك: الفلوس، الفواتير، الحجوزات المتأخرة، أداء الفرق، ملف أي موظف، الحضور، تأخير المشاريع، شغل المخازن، التقييم.
- قرر بنفسك شنو تفحص وبأي ترتيب. افحص أكثر من مصدر إذا السؤال يحتاج، واربط بين الأرقام.
- جاوب باللهجة العراقية، مرتّب: الخلاصة أول سطر، بعدها نقاط قصيرة بالأرقام، وآخر شي اقتراح عملي واحد أو اثنين.
- الموظفين يجوك كرموز «موظف#n» — استعمل نفس الرمز بالضبط، ولا تخمّن الأسماء.
- لا تخترع رقم ما شفته بأداة. إذا ما تكدر تتأكد گول بصراحة.
- لا تقترح غرامة أو خصم فلوس، ولا تخمّن أسباب شخصية (مرض، ظروف).`

func NewMatrixCommandService(repo *repository.MatrixCommandRepository, watch *MatrixAutopilotService,
	reports *MatrixEmployeeReportService, business *MatrixBusinessService, props *repository.MatrixProposalRepository) *MatrixCommandService {
	return &MatrixCommandService{repo: repo, watch: watch, reports: reports, business: business, props: props}
}

func (s *MatrixCommandService) EnableModel(apiKey, modelName string) {
	if apiKey == "" {
		return
	}
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	s.client, s.model = &c, modelName
}

func (s *MatrixCommandService) Feed() ([]repository.FeedRow, error) { return s.repo.Feed(10) }

func (s *MatrixCommandService) Trend() ([]repository.TrendDay, error) { return s.repo.DailyTrend(30) }

// ── تركيز التأخير: آخر ٣ أيام مقابل الـ٣ الي قبلها ──

type LateFocus struct {
	Recent     *repository.LateWindow `json:"recent"`
	Previous   *repository.LateWindow `json:"previous"`
	RecentPct  int                    `json:"recentPct"`
	PrevPct    int                    `json:"prevPct"`
	ChangePct  *int                   `json:"changePct"` // nil = ما يتحسب
	Cause      string                 `json:"cause"`
	Impact     string                 `json:"impact"`
	Suggestion string                 `json:"suggestion"`
	Severity   string                 `json:"severity"` // HIGH | MEDIUM | LOW
	Items      []repository.LateItem  `json:"items"`    // الحجوزات المتأخرة نفسها
}

func pct(a, b int) int {
	if b == 0 {
		return 0
	}
	return int(math.Round(float64(a) * 100 / float64(b)))
}

func (s *MatrixCommandService) LateFocus() (*LateFocus, error) {
	cur, err := s.repo.LateWindow(2, 0)
	if err != nil {
		return nil, err
	}
	prev, err := s.repo.LateWindow(5, 3)
	if err != nil {
		return nil, err
	}
	f := &LateFocus{Recent: cur, Previous: prev, RecentPct: pct(cur.Late, cur.Total), PrevPct: pct(prev.Late, prev.Total), Severity: "LOW"}
	if items, err := s.repo.LateItems(2, 0, 50); err == nil {
		f.Items = items
	} else {
		f.Items = []repository.LateItem{}
	}
	if prev.Total >= 3 && f.PrevPct > 0 {
		c := pct(f.RecentPct-f.PrevPct, f.PrevPct)
		f.ChangePct = &c
	}
	switch {
	case cur.Late == 0:
		f.Cause, f.Impact, f.Suggestion = "ماكو تأخير بآخر ٣ أيام.", "—", "كمّلوا بنفس الوتيرة."
	case cur.Unstaffed*2 >= cur.Late:
		f.Cause = fmt.Sprintf("%d من الحجوزات بلا كادر مكلّف — نقص تغطية بقسم التنسيق.", cur.Unstaffed)
		f.Suggestion = "كلّف كادر للحجوزات الفارغة، ووزّع المهام على الفترات الأزحم."
	case cur.Partial*2 >= cur.Late:
		f.Cause = fmt.Sprintf("%d حجز صار جزئي — غالباً نقص مواد أو عدّة.", cur.Partial)
		f.Suggestion = "تأكد من تجهيز المواد قبل الطلعة، وراجع أسباب الجزئي."
	default:
		f.Cause = "الحجوزات تخلص بعد يومها — بطء بالتنفيذ أو ضغط حجوزات."
		f.Suggestion = "شوف أداء المجموعات (عيون ماتركس) ووزّع الحجوزات الكبيرة."
	}
	if cur.Late > 0 {
		f.Impact = fmt.Sprintf("%d حجز بعده مفتوح وموعده فات — احتمال زعل الزبائن وتأخّر الإيراد.", cur.OpenNow)
		if f.RecentPct >= 30 || cur.OpenNow >= 5 {
			f.Severity = "HIGH"
		} else if f.RecentPct >= 15 {
			f.Severity = "MEDIUM"
		}
	}
	return f, nil
}

// ── اسأل ماتركس ──

type AskAnswer struct {
	Answer string   `json:"answer"`
	Source string   `json:"source"` // AGENT | MODEL | RULES
	Steps  []string `json:"steps,omitempty"`
}

type askContext struct {
	lines []string
	names []string // الأسماء الحقيقية — الترتيب = رقم الموظف
}

func (s *MatrixCommandService) buildContext() *askContext {
	ctx := &askContext{}
	add := func(f string, a ...any) { ctx.lines = append(ctx.lines, fmt.Sprintf(f, a...)) }
	if lf, err := s.LateFocus(); err == nil {
		add("الحجوزات آخر ٣ أيام: %d، المتأخر %d (%d٪)، بعده مفتوح وموعده فات %d، بلا كادر %d، جزئي %d. الـ٣ أيام الي قبلها: متأخر %d٪.",
			lf.Recent.Total, lf.Recent.Late, lf.RecentPct, lf.Recent.OpenNow, lf.Recent.Unstaffed, lf.Recent.Partial, lf.PrevPct)
	}
	if b, err := s.business.View(); err == nil && b.MTD != nil {
		add("هالشهر لحد اليوم: %d حجز منجز، إيراد %.0f د.ع (نفس الفترة الشهر الماضي: %d حجز، %.0f د.ع). توقع آخر الشهر: %.0f د.ع (%s).",
			b.MTD.Bookings, b.MTD.Revenue, b.MTD.LastBookings, b.MTD.LastRevenue, b.Forecast.Expected, b.Forecast.Basis)
	}
	if groups, err := s.watch.RoleWatch(); err == nil {
		for _, g := range groups {
			if len(g.Employees) == 0 {
				continue
			}
			add("مجموعة %s: %d موظف، %d حمرة، %d منتبهة.", g.Group, len(g.Employees), g.Red, g.Alert)
			gp, err := s.reports.GroupPerformance(g.Group)
			if err != nil {
				continue
			}
			for _, m := range gp.Members {
				if m.Score == 0 && !m.Absent && m.Late == 0 {
					continue
				}
				ctx.names = append(ctx.names, m.Name)
				tag := fmt.Sprintf("موظف#%d", len(ctx.names))
				// كل موظف بشغل دوره: الحجوزات للميدانيين، والباقين بجداول دورهم.
				var line string
				if gp.Field {
					line = fmt.Sprintf("  %s (%s): حجوزات ٣٠ يوم %d، منجز %d، جزئي %d، مفتوح %d", tag, g.Group, m.Jobs, m.Completed, m.Partial, m.Open)
					if m.Speed != nil {
						line += fmt.Sprintf("، السرعة ×%.2f", *m.Speed)
					}
				} else {
					line = fmt.Sprintf("  %s (%s): شغل دوره ٣٠ يوم:", tag, g.Group)
					for i, x := range m.Metrics {
						if i > 0 {
							line += "،"
						}
						line += fmt.Sprintf(" %s %d", x.Label, x.Value)
					}
					if m.Pending > 0 {
						line += fmt.Sprintf("، طابور دوره هسه %d", m.Pending)
					}
				}
				if m.Absent {
					line += "، ما حضر اليوم"
				} else if m.Late > 0 {
					line += fmt.Sprintf("، تأخّر اليوم %d دقيقة", m.Late)
				}
				ctx.lines = append(ctx.lines, line)
			}
		}
	}
	if ps, err := s.props.List("PENDING", 5); err == nil {
		for _, p := range ps {
			add("اقتراح معلّق: %s", p.Title)
		}
	}
	return ctx
}

// anonymize يبدّل كل اسم بـ«موظف#n» — الأطول أولاً حتى الأسماء المتداخلة ما تنكسر.
func (c *askContext) anonymize(text string) string {
	idx := make([]int, len(c.names))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return len([]rune(c.names[idx[a]])) > len([]rune(c.names[idx[b]])) })
	for _, i := range idx {
		if c.names[i] != "" {
			text = strings.ReplaceAll(text, c.names[i], fmt.Sprintf("موظف#%d", i+1))
		}
	}
	return text
}

var tagRe = regexp.MustCompile(`موظف#(\d+)`)

func (c *askContext) restore(text string) string {
	return tagRe.ReplaceAllStringFunc(text, func(m string) string {
		var n int
		fmt.Sscanf(strings.TrimPrefix(m, "موظف#"), "%d", &n)
		if n >= 1 && n <= len(c.names) {
			return c.names[n-1]
		}
		return m
	})
}

const askSystemPrompt = `أنت «ماتركس»، عقل المتابعة بشركة الأماني (العراق). تجاوب المدير على سؤاله من الأرقام المعطاة بس.
جاوب باللهجة العراقية بـ٢-٥ جمل، بالأرقام. الموظفين مذكورين كـ«موظف#n» — استعمل نفس الرمز.
لا تخترع رقم ما موجود. إذا الأرقام ما تجاوب السؤال گول هذا بصراحة.
لا تخمّن أسباب شخصية (مرض، سفر، ظروف)، ولا تقترح عقوبة أو غرامة. اقتراحاتك عملية (توزيع، تذكير، تجهيز مواد).`

func (s *MatrixCommandService) Ask(employeeID, question string) (*AskAnswer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, errors.New("اكتب سؤالك")
	}
	if len([]rune(question)) > 500 {
		return nil, errors.New("السؤال طويل — خليه أقصر")
	}
	if s.repo.AsksToday(employeeID) >= askDailyLimit {
		return nil, ErrAskLimit
	}
	// قرار (ع) 10-10: ماتركس الوكيل يفتش بنفسه بالأدوات، وإذا تعذّر يرجع للطريقة القديمة.
	if s.agent.Enabled("ASK") {
		if run, err := s.agent.Run("ASK", agentAskPrompt, question); err == nil && run.Answer != "" {
			out := &AskAnswer{Answer: run.Answer, Source: "AGENT", Steps: run.Steps}
			s.repo.LogAsk(employeeID, question, out.Answer, out.Source)
			return out, nil
		} else if err != nil {
			log.Printf("matrix agent ask: %v", err)
		}
	}
	ctx := s.buildContext()
	out := &AskAnswer{Source: "RULES"}
	if s.client != nil {
		if a, err := s.askModel(ctx, question); err == nil && a != "" {
			out.Answer, out.Source = ctx.restore(a), "MODEL"
		}
	}
	if out.Answer == "" {
		out.Answer = ctx.restore(rulesAnswer(ctx, question))
	}
	s.repo.LogAsk(employeeID, question, out.Answer, out.Source)
	return out, nil
}

func (s *MatrixCommandService) askModel(c *askContext, question string) (string, error) {
	if !AIFeatureOn("ASK") {
		return "", ErrAIFeatureOff
	}
	body := "الأرقام:\n" + c.anonymize(strings.Join(c.lines, "\n")) + "\n\nسؤال المدير: " + c.anonymize(question)
	cx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := s.client.Messages.New(cx, anthropic.MessageNewParams{
		Model: anthropic.Model(s.model), MaxTokens: 400,
		System:   []anthropic.TextBlockParam{{Text: askSystemPrompt, CacheControl: anthropic.NewCacheControlEphemeralParam()}},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(body))},
	})
	if err == nil {
		RecordAIUsage("ASK", s.model, resp.Usage)
	}
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	return strings.TrimSpace(b.String()), nil
}

// rulesAnswer بلا نموذج: نختار سطور السياق الي تخص كلمات السؤال.
func rulesAnswer(c *askContext, q string) string {
	pick := func(keys ...string) []string {
		var out []string
		for _, l := range c.lines {
			for _, k := range keys {
				if strings.Contains(l, k) {
					out = append(out, strings.TrimSpace(l))
					break
				}
			}
		}
		return out
	}
	var lines []string
	switch {
	case containsAny(q, "متأخر", "تأخير", "تأخر", "التأخر"):
		lines = pick("الحجوزات آخر", "تأخّر اليوم", "ما حضر")
	case containsAny(q, "ربح", "أرباح", "ارباح", "إيراد", "ايراد", "فلوس", "مبيعات"):
		lines = pick("هالشهر لحد اليوم")
	// سؤال عن دور معيّن: سطور مجموعته بس، بشغل دورها.
	case containsAny(q, "محاسب", "المحاسبة", "حسابات"):
		lines = pick("(FINANCE)", "مجموعة FINANCE")
	case containsAny(q, "مراقب", "المراقبة", "تدقيق"):
		lines = pick("(MONITORS)", "مجموعة MONITORS")
	case containsAny(q, "منسق", "تنسيق", "منسّق"):
		lines = pick("(COORDINATORS)", "مجموعة COORDINATORS")
	case containsAny(q, "ليدر", "فني", "فنيين", "كادر"):
		lines = pick("(LEADERS)", "(TECHS)", "مجموعة LEADERS", "مجموعة TECHS")
	case containsAny(q, "أداء", "اداء", "موظف", "منو", "مين", "سرعة", "جزئي"):
		lines = pick("موظف#", "مجموعة")
	case containsAny(q, "اقتراح", "مقترح", "تحسين", "شسوي", "حل"):
		lines = pick("اقتراح معلّق", "الحجوزات آخر")
	default:
		lines = pick("الحجوزات آخر", "هالشهر لحد اليوم")
	}
	if len(lines) == 0 {
		return "ما عندي أرقام تخص هالسؤال هسه. جرّب تسأل عن التأخير، أو الأرباح، أو أداء الموظفين، أو المقترحات."
	}
	if len(lines) > 8 {
		lines = lines[:8]
	}
	return "حسب الأرقام الحالية:\n• " + strings.Join(lines, "\n• ")
}

func containsAny(s string, keys ...string) bool {
	for _, k := range keys {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}
