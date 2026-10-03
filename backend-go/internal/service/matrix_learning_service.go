package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ ماتركس يقترح ويتعلّم ═══
//
// قرار (ع): ماتركس يطور نفسه، بس القرار بإيد المدير بالفترة الأولى.
// الحلقة: ماتركس يشوف (نتائج تذكيراته، وتيرة الموظفين، تعليماته) ←
// يقترح (تعليمة، توقع، تعديل) ← المدير يوافق/يرفض بسبب ← الموافَق
// يصير جزء منه، والمرفوض ما يتكرر ويندز للنموذج كدرس.
//
// ⚠️ ولا اقتراح ينفّذ بلا موافقة. وماكو غرامة ولا نقاط. وللنموذج أرقام
// ومجموعات بس — بلا أسماء.

const (
	predictHour     = 13
	workdayStart    = 8.0
	workdayEnd      = 17.0
	predictMinLeft  = 3
	escalateRatio   = 0.5
	escalateMinSeen = 4
)

// شاشة الشغل لكل نوع تذكير — وين تنفع التعليمة المقترحة.
var kindRoute = map[string]struct{ Route, Group string }{
	model.AiActionPaperworkReminder: {"/leader-invoices/new", "LEADERS"},
	model.AiActionUnstaffedAlert:    {"/coordinator", "COORDINATORS"},
	model.AiActionExtraTaskOverdue:  {"/my-extra-tasks", ""},
	model.AiActionInvoiceApproval:   {"/leader-invoices", "FINANCE"},
	model.AiActionGpsExpiry:         {"/gps", ""},
	model.AiActionVehicleDocExpiry:  {"/vehicles", ""},
	model.AiActionLowStock:          {"/inventory", ""},
}

type MatrixLearningService struct {
	props   *repository.MatrixProposalRepository
	rules   *repository.MatrixGuideRepository
	actions *repository.AiActionRepository
	aiRepo  *repository.AiRepository
	notif   *repository.NotificationRepository
	watch   *MatrixAutopilotService
	client  *anthropic.Client
	model   string
	reports *MatrixEmployeeReportService
	// أسماء آخر توقع — تنذكر بالإشعار.
	lastPredicted []string
}

func NewMatrixLearningService(props *repository.MatrixProposalRepository, rules *repository.MatrixGuideRepository,
	actions *repository.AiActionRepository, aiRepo *repository.AiRepository, notif *repository.NotificationRepository,
	watch *MatrixAutopilotService) *MatrixLearningService {
	return &MatrixLearningService{props: props, rules: rules, actions: actions, aiRepo: aiRepo, notif: notif, watch: watch}
}

func (s *MatrixLearningService) EnableModel(apiKey, modelName string) {
	if apiKey == "" {
		return
	}
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	s.client, s.model = &c, modelName
}

func toJSON(v any) model.NullJSON {
	b, _ := json.Marshal(v)
	return b
}

// RunIfDue: التوقعات يومياً بعد ١ الظهر، والتعلّم من النتائج أسبوعياً.
func (s *MatrixLearningService) RunIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Hour() >= predictHour && now.Weekday() != time.Friday {
		if ok, err := s.aiRepo.ClaimDailyMarker("DAILY_MATRIX_PREDICT", now.Format("2006-01-02")); err == nil && ok {
			n := s.predict(now)
			if n > 0 {
				_ = s.notif.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "", "AI_DECISIONS",
					fmt.Sprintf("🤖 ماتركس — توقّعت %d من الليدرية ما راح يخلّصون شغل اليوم بوتيرتهم: %s. اضغط حتى تشوف منو وليش، والقرار إلك.",
						n, strings.Join(firstN(uniqueStrings(s.lastPredicted), 4), "، ")))
			}
		}
	}
	if now.Hour() >= declineHour && now.Weekday() != time.Friday {
		if ok, err := s.aiRepo.ClaimDailyMarker("DAILY_MATRIX_DECLINE", now.Format("2006-01-02")); err == nil && ok {
			if n := s.detectDeclines(); n > 0 {
				_ = s.notif.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "", "AI_DECISIONS",
					fmt.Sprintf("🤖 ماتركس — %d موظف نزلت سرعتهم بالحجوزات هالشهر. التفاصيل بصندوق القرارات.", n))
			}
		}
	}
	if now.Hour() >= 11 {
		if ok, err := s.aiRepo.ClaimDailyMarker("WEEKLY_MATRIX_LEARN", isoWeekMonday(now)); err == nil && ok {
			s.learnFromOutcomes()
			s.askModel()
		}
	}
	return nil
}

// ── (أ) التوقع: ما راح يخلّص اليوم ──

func (s *MatrixLearningService) predict(now time.Time) int {
	subs, err := s.actions.ActiveSubjects()
	if err != nil {
		return 0
	}
	hour := float64(now.Hour()) + float64(now.Minute())/60
	elapsed := math.Max(hour-workdayStart, 1)
	remaining := math.Max(workdayEnd-hour, 0)
	n := 0
	names := []string{}
	for _, sub := range subs {
		// قرار (ع): نتوقع لليدرية بس — الفنيين يتبعون الليدر، فالتوقع عليه.
		if WatchGroup(sub) != "LEADERS" {
			continue
		}
		for _, w := range s.watch.workload(sub) {
			if w.Done == 0 && w.Left < predictMinLeft*2 { // بلا منجز ما نعرف وتيرة — بس لو الكومة كبيرة
				continue
			}
			pace := float64(w.Done) / elapsed
			expected := int(math.Floor(pace * remaining))
			if w.Left < predictMinLeft || expected >= w.Left {
				continue
			}
			gap := w.Left - expected
			title := fmt.Sprintf("%s: %s %d بـ%.0f ساعة، وباقي %d. بهالوتيرة يخلّص %d بس.", sub.Name, w.Verb, w.Done, elapsed, w.Left, expected)
			ok, _ := s.props.Create(model.MatrixProposal{
				Kind: model.ProposalPrediction, Source: "RULES",
				Title:     title,
				Rationale: fmt.Sprintf("الوتيرة %.1f بالساعة، والدوام باقيله %.0f ساعة — راح يبقى %d من «%s».", pace, remaining, gap, w.Label),
				Evidence:  toJSON(map[string]any{"done": w.Done, "left": w.Left, "hoursElapsed": elapsed, "hoursRemaining": remaining, "pace": math.Round(pace*10) / 10, "expected": expected}),
				Payload: toJSON(map[string]any{"employeeId": sub.ID, "route": w.Route,
					"message": fmt.Sprintf("🤖 ماتركس — باقي عليك %d من «%s»، وبوتيرتك الحالية ما راح تخلّصها اليوم. ركّز عليها هسه.", w.Left, w.Label)}),
				Signature: fmt.Sprintf("PRED|%s|%s|%s", sub.ID, w.Key, now.Format("2006-01-02")),
			})
			if ok {
				n++
				names = append(names, sub.Name)
			}
		}
	}
	s.lastPredicted = names
	return n
}

// ── (ب) التعلّم من النتائج ──

func (s *MatrixLearningService) learnFromOutcomes() {
	outs, err := s.props.KindOutcomes()
	if err == nil {
		for _, o := range outs {
			kr, ok := kindRoute[o.Kind]
			if !ok || o.Total < escalateMinSeen || float64(o.Escalated)/float64(o.Total) < escalateRatio {
				continue
			}
			label := model.AiActionLabels[o.Kind]
			_, _ = s.props.Create(model.MatrixProposal{
				Kind: model.ProposalGuideRule, Source: "RULES",
				Title:     fmt.Sprintf("«%s» يصعد كثير: %d من %d ما انحلت بعد التذكير", label, o.Escalated, o.Total),
				Rationale: "التذكير بالإشعار ما يكفي. أقترح ماتركس ينبّه الموظف لحظة يفتح شاشة هالشغل، قبل ما يتأخر.",
				Evidence:  toJSON(o),
				Payload: toJSON(model.MatrixGuideRule{Route: kr.Route, Groups: kr.Group, OnlyIfPending: true, Priority: 7,
					Text: "عندك {work}. هذا الي يتأخر عادةً ويوصل للمدير — خلّصه أول."}),
				Signature: "ESC|" + o.Kind,
			})
		}
	}
	if offs, err := s.props.RepeatEscalations(); err == nil {
		for _, off := range offs {
			sub, err := s.actions.Subject(off.EmployeeID)
			if err != nil {
				continue
			}
			g := WatchGroup(*sub)
			_, _ = s.props.Create(model.MatrixProposal{
				Kind: model.ProposalGuideRule, Source: "RULES",
				Title:     fmt.Sprintf("%s صعد عليه %d مرات بشهر", sub.Name, off.Escalated),
				Rationale: "نمط متكرر بنفس المجموعة. أقترح تعليمة أصرم لمجموعته تنطلع بأي شاشة لو عنده شغل باقي.",
				Evidence:  toJSON(map[string]any{"escalations30d": off.Escalated, "group": g}),
				Payload: toJSON(model.MatrixGuideRule{Route: "/", Groups: g, OnlyIfPending: true, Priority: 3,
					Text: "عندك {work} وما انحلت. ماتركس يتابعك — خلّصها قبل أي شي ثاني."}),
				Signature: "REPEAT|" + g,
			})
		}
	}
	if unused, err := s.props.UnusedRules(); err == nil {
		for _, r := range unused {
			_, _ = s.props.Create(model.MatrixProposal{
				Kind: model.ProposalRuleTune, Source: "RULES",
				Title:     fmt.Sprintf("تعليمة ما انطبقت ولا مرة من ٣٠ يوم: «%s»", r.Text),
				Rationale: fmt.Sprintf("على %s. يا الزر تغيّر اسمه، يا الموقف ما يصير. أقترح أوقفها.", r.Route),
				Evidence:  toJSON(map[string]any{"hits": r.Hits, "lastHitAt": r.LastHitAt}),
				Payload:   toJSON(map[string]any{"ruleId": r.ID, "enabled": false}),
				Signature: "UNUSED|" + r.ID,
			})
		}
	}
}

// ── (ج) هايكو يقترح ──

const learnSystemPrompt = `أنت «ماتركس»، عقل المتابعة بنظام شركة الأماني (العراق). تشوف أرقام تذكيراتك ونتائجها، وتعليماتك الحالية، والاقتراحات الي رفضها المدير وأسبابها.
اقترح لحد ٣ تعليمات توجيه جديدة تقلّل التأخير والتصعيد. كل تعليمة: مسار شاشة يبدي بـ/، كلمات زر اختيارية مفصولة بـ|، مجموعة أدوار اختيارية، ونص جملة أو جملتين بالعراقي، صارم بلا تهديد بعقوبة، ويقبل {left} و{work}.
لا تكرر تعليمة موجودة ولا شي رفضه المدير. إذا ماكو شي مفيد، رجّع قائمة فاضية.`

var learnSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"proposals": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":         map[string]any{"type": "string"},
					"rationale":     map[string]any{"type": "string"},
					"route":         map[string]any{"type": "string"},
					"match":         map[string]any{"type": "string"},
					"groups":        map[string]any{"type": "string"},
					"text":          map[string]any{"type": "string"},
					"onlyIfPending": map[string]any{"type": "boolean"},
				},
				"required":             []string{"title", "rationale", "route", "match", "groups", "text", "onlyIfPending"},
				"additionalProperties": false,
			},
		},
	},
	"required":             []string{"proposals"},
	"additionalProperties": false,
}

func (s *MatrixLearningService) askModel() {
	if s.client == nil {
		return
	}
	outs, _ := s.props.KindOutcomes()
	rules, _ := s.rules.List()
	rej, _ := s.props.RecentRejections(15)
	type ruleLite struct{ Route, Match, Groups, Text string }
	rl := []ruleLite{}
	for _, r := range rules {
		if r.Enabled {
			rl = append(rl, ruleLite{r.Route, r.Match, r.Groups, r.Text})
		}
	}
	type rejLite struct{ Title, Note string }
	rj := []rejLite{}
	for _, p := range rej {
		note := ""
		if p.Note != nil {
			note = *p.Note
		}
		rj = append(rj, rejLite{p.Title, note})
	}
	user := fmt.Sprintf("نتائج التذكيرات (٣٠ يوم):\n%s\n\nالتعليمات الحالية:\n%s\n\nالمرفوض سابقاً وسببه:\n%s",
		toJSON(outs), toJSON(rl), toJSON(rj))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:        anthropic.Model(s.model),
		MaxTokens:    1500,
		System:       []anthropic.TextBlockParam{{Text: learnSystemPrompt}},
		OutputConfig: anthropic.OutputConfigParam{Format: anthropic.JSONOutputFormatParam{Schema: learnSchema}},
		Messages:     []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(user))},
	})
	if err != nil {
		log.Printf("[ai] اقتراحات النموذج: %v", err)
		return
	}
	var b strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	var out struct {
		Proposals []struct {
			Title, Rationale, Route, Match, Groups, Text string
			OnlyIfPending                                bool
		} `json:"proposals"`
	}
	if err := json.Unmarshal([]byte(b.String()), &out); err != nil {
		log.Printf("[ai] اقتراحات النموذج مو مقروءة: %v", err)
		return
	}
	for i, p := range out.Proposals {
		if i >= 3 || !strings.HasPrefix(p.Route, "/") || strings.TrimSpace(p.Text) == "" {
			continue
		}
		_, _ = s.props.Create(model.MatrixProposal{
			Kind: model.ProposalGuideRule, Source: "MODEL", Title: p.Title, Rationale: p.Rationale,
			Payload: toJSON(model.MatrixGuideRule{Route: p.Route, Match: p.Match, Groups: p.Groups, Text: p.Text,
				OnlyIfPending: p.OnlyIfPending, Priority: 6}),
			Signature: "MODEL|" + p.Route + "|" + p.Match + "|" + p.Text,
		})
	}
}

// ── القرار ──

func (s *MatrixLearningService) List(status string) ([]model.MatrixProposal, error) {
	return s.props.List(status, 200)
}

// Approve ينفّذ الاقتراح. override: تعديل المدير على التعليمة قبل الموافقة.
func (s *MatrixLearningService) Approve(id, by string, override json.RawMessage) error {
	p, err := s.props.Find(id)
	if err != nil {
		return err
	}
	if p.Status != "PENDING" {
		return fmt.Errorf("الاقتراح انقرر عليه من قبل")
	}
	payload := []byte(p.Payload)
	if len(override) > 0 && string(override) != "null" {
		payload = override
	}
	switch p.Kind {
	case model.ProposalGuideRule:
		var r model.MatrixGuideRule
		if err := json.Unmarshal(payload, &r); err != nil {
			return fmt.Errorf("التعليمة المقترحة مو مقروءة")
		}
		if r.Priority == 0 {
			r.Priority = 5
		}
		if _, err := s.rules.CreateFrom(r, "MATRIX"); err != nil {
			return err
		}
	case model.ProposalRuleTune:
		var t struct {
			RuleID  string `json:"ruleId"`
			Enabled bool   `json:"enabled"`
		}
		if err := json.Unmarshal(payload, &t); err != nil || t.RuleID == "" {
			return fmt.Errorf("التعديل المقترح مو مقروء")
		}
		if err := s.rules.SetEnabled(t.RuleID, t.Enabled); err != nil {
			return err
		}
	case model.ProposalPrediction:
		var pr struct {
			EmployeeID string `json:"employeeId"`
			Message    string `json:"message"`
		}
		if err := json.Unmarshal(payload, &pr); err != nil || pr.EmployeeID == "" {
			return fmt.Errorf("التوقع مو مقروء")
		}
		emp := pr.EmployeeID
		ok, _ := s.actions.Claim(model.AiAction{Kind: model.AiActionPredictionNudge, EntityType: "EMPLOYEE", EntityID: emp,
			Period: time.Now().In(debriefLoc).Format("2006-01-02"), TargetEmployeeID: &emp, TargetLabel: s.actions.EmployeeName(emp),
			Summary: "تذكير بعد توقّع ماتركس: " + p.Title, Details: p.Evidence})
		if ok {
			_ = s.notif.Create(emp, "AI_AUTOPILOT", pr.Message)
		}
	}
	return s.props.Decide(id, "APPROVED", by, "")
}

func (s *MatrixLearningService) Reject(id, by, note string) error {
	return s.props.Decide(id, "REJECTED", by, strings.TrimSpace(note))
}

func (s *MatrixLearningService) PendingCount() int { return s.props.PendingCount() }

func uniqueStrings(xs []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
