package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/whatsapp"
)

// ═══ 💬 وكيلي الواتساب: الجودة والحجوزات (دازاي 10-23 — docs/ai-quality-booking-plan.md) ═══
// ناروتو: المحادثة والأدوات والمشغّل ووضع التجربة. لوفي: العقل والسياسة (WaBrain + wa_agent_prompts.go).
//
// 🔴 حدود ما تنكسر:
//   - الوكيل ما يغرّم ولا يحذف ولا يلمس فلوس. ما يستدعي Verdict/Inspect (بيهن خصم نقاط).
//   - كل فعل ينسجل بـAiAgentAction ويا سببه، وينلغى بضغطة.
//   - بالتجربة الأفعال تنسجل بس وما تنفّذ.
//   - اسم الزبون ورقمه ما يروحن للنموذج لحد قرار الخصوصية المكتوب (PrivacyApproved).
//   - المشغّل مطفي لحد ما المالك يشغّل المفتاح، والواتساب مربوط.

const (
	WaAgentQuality = "QUALITY"
	WaAgentBooking = "BOOKING"
)

var ErrWaAgentOff = errors.New("الوكيل مطفي من الإعدادات")

// WaTurn دور واحد بالمحادثة يوصل للعقل.
type WaTurn struct {
	Agent     string
	AgentName string
	Context   string // سياق الحجز كنص (بلا اسم الزبون ورقمه)
	History   []AgentTurn
	Message   string
	Tools     []agentTool
}

// WaBrain العقل — لوفي يبدّله بسياسته. Reply يرجّع نص الرد للزبون.
type WaBrain interface {
	Reply(ctx context.Context, t WaTurn) (string, error)
}

type waNotifier interface {
	Create(employeeID, kind, message string) error
	CreateForRole(role, kind, message string) error
}

type WaAgentService struct {
	db        *sqlx.DB
	repo      *repository.WaRepository
	wa        *whatsapp.Client
	bookings  *BookingService
	followups *repository.QualityFollowUpRepository
	proposals *repository.MatrixProposalRepository
	deletes   *repository.BookingDeleteRequestRepository
	switches  *repository.SystemSwitchRepository
	notif     waNotifier
	names     func(agent string) string
	brains    map[string]WaBrain

	DailyNew        int
	Template        string
	TemplateLang    string
	PrivacyApproved bool
}

func NewWaAgentService(db *sqlx.DB, repo *repository.WaRepository, wa *whatsapp.Client, bookings *BookingService,
	followups *repository.QualityFollowUpRepository, proposals *repository.MatrixProposalRepository,
	deletes *repository.BookingDeleteRequestRepository, switches *repository.SystemSwitchRepository,
	notif waNotifier, names func(string) string) *WaAgentService {
	return &WaAgentService{db: db, repo: repo, wa: wa, bookings: bookings, followups: followups, proposals: proposals,
		deletes: deletes, switches: switches, notif: notif, names: names, brains: map[string]WaBrain{}, DailyNew: 10, TemplateLang: "ar"}
}

// SetBrain عقل وكيل (nil = عقل مؤقت يرحّب ويحوّل لإنسان).
func (s *WaAgentService) SetBrain(agent string, b WaBrain) { s.brains[agent] = b }

func (s *WaAgentService) Repo() *repository.WaRepository { return s.repo }

// WhatsAppOn الإرسال الحقيقي مربوط؟
func (s *WaAgentService) WhatsAppOn() bool { return s.wa.On() }

func waSwitch(agent string) string {
	if agent == WaAgentQuality {
		return model.SwitchAIQualityAgent
	}
	return model.SwitchAIBookingAgent
}

// Enabled مفتاح التشغيل (المالك).
func (s *WaAgentService) Enabled(agent string) bool {
	if s.switches == nil {
		return false
	}
	m, err := s.switches.All()
	return err == nil && m[waSwitch(agent)]
}

func (s *WaAgentService) SetEnabled(agent string, on bool, by string) error {
	if on && !s.wa.On() {
		return whatsapp.ErrOff
	}
	return s.switches.Set(waSwitch(agent), on, by)
}

func (s *WaAgentService) name(agent string) string {
	if s.names != nil {
		if n := s.names(agent); n != "" {
			return n
		}
	}
	if agent == WaAgentQuality {
		return "مساعد الجودة"
	}
	return "مساعد الحجوزات"
}

// Greeting أول رسالة — كلام دازاي: «مرحبا زبوننا الكريم وياك مهندس الجودة من شركة الأماني».
func (s *WaAgentService) Greeting(agent string, c *repository.WaBookingContext) string {
	n := s.name(agent)
	if agent == WaAgentQuality {
		return fmt.Sprintf("مرحبا زبوننا الكريم 🌷 وياك %s، مهندس الجودة من شركة الأماني (مساعد ذكي). حبينا نتطمّن على الشغل الي سويناه إلك (حجز %s) — كلشي تمام؟", n, c.Code)
	}
	what := c.Services
	if what == "" {
		what = "طلبك"
	}
	return fmt.Sprintf("مرحبا زبوننا الكريم 🌷 وياك %s من قسم الحجوزات بشركة الأماني (مساعد ذكي). وصلنا طلبك (%s) وحابين نثبّتلك موعد — يناسبك نكمل التفاصيل هسه؟", n, what)
}

func (s *WaAgentService) contextText(c *repository.WaBookingContext) string {
	var b strings.Builder
	fmt.Fprintf(&b, "الحجز: %s · النوع: %s · الحالة: %s", c.Code, c.Type, c.Status)
	if c.Services != "" {
		b.WriteString(" · الخدمة: " + c.Services)
	}
	if c.ScheduledAt != nil {
		b.WriteString(" · الموعد: " + secFmt(c.ScheduledAt))
	}
	if c.CompletedAt != nil {
		b.WriteString(" · خلص: " + secFmt(c.CompletedAt))
	}
	// ⚠️ اسم الليدر ما ينرسل — أسماء الموظفين ممنوعة على الذكاء الخارجي (إلا ايريس).
	// ملاحظات الموظفين ممكن بيها اسم الزبون — بس بعد قرار الخصوصية.
	if s.PrivacyApproved && c.Notes != nil && strings.TrimSpace(*c.Notes) != "" {
		b.WriteString("\nملاحظات الحجز: " + *c.Notes)
	}
	return b.String()
}

// ── وضع التجربة (المالك) ──

// StartSandbox محادثة تجربة على حجز — بلا واتساب، وبلا تنفيذ على النظام.
func (s *WaAgentService) StartSandbox(agent, bookingID string) (*repository.WaConversation, []repository.WaMessage, error) {
	if agent != WaAgentQuality && agent != WaAgentBooking {
		return nil, nil, errors.New("وكيل غير معروف")
	}
	c, err := s.repo.BookingContext(bookingID)
	if err != nil {
		return nil, nil, errors.New("الحجز مو موجود")
	}
	var fu *string
	if agent == WaAgentQuality {
		if q, err := s.followups.ByBooking(bookingID); err == nil && q != nil {
			fu = &q.ID
		}
	}
	cust := c.CustomerID
	conv, err := s.repo.CreateConversation(agent, &cust, "sandbox", &bookingID, fu, true)
	if err != nil {
		return nil, nil, err
	}
	if _, _, err := s.repo.AddMessage(conv.ID, "OUT", "AGENT", nil, s.Greeting(agent, c), nil); err != nil {
		return nil, nil, err
	}
	msgs, err := s.repo.Messages(conv.ID)
	return conv, msgs, err
}

// Say رسالة زبون (تجربة أو حقيقية واردة) والوكيل يرد. يرجّع رسائل هذا الدور وإجراءاته.
func (s *WaAgentService) Say(ctx context.Context, convID, text string, metaID *string) ([]repository.WaMessage, []repository.AiAgentAction, error) {
	conv, err := s.repo.Conversation(convID)
	if err != nil {
		return nil, nil, err
	}
	if conv.Status == "CLOSED" {
		return nil, nil, errors.New("المحادثة مسدودة")
	}
	in, inserted, err := s.repo.AddMessage(conv.ID, "IN", "CUSTOMER", nil, text, metaID)
	if err != nil || !inserted {
		return nil, nil, err // المكرر من Meta: ولا شي
	}
	out := []repository.WaMessage{*in}
	if conv.Status != "BOT" {
		return out, nil, nil // إنسان ماسكها — الوكيل ساكت
	}
	if !conv.Sandbox && !s.Enabled(conv.Agent) {
		return out, nil, nil
	}
	reply, actions, err := s.reply(ctx, conv)
	if err != nil {
		log.Printf("[wa] %s reply: %v", conv.Agent, err)
		reply = "شكراً إلك 🙏 راح يتواصل وياك موظف من شركة الأماني بأقرب وقت."
		s.handoff(conv, "خلل بالوكيل: "+err.Error())
	}
	if strings.TrimSpace(reply) == "" {
		return out, actions, nil
	}
	if !conv.Sandbox {
		if _, err := s.wa.SendText(ctx, conv.Phone, reply); err != nil {
			log.Printf("[wa] send: %v", err)
		}
	}
	m, _, err := s.repo.AddMessage(conv.ID, "OUT", "AGENT", nil, reply, nil)
	if err == nil && m != nil {
		out = append(out, *m)
	}
	return out, actions, nil
}

func (s *WaAgentService) reply(ctx context.Context, conv *repository.WaConversation) (string, []repository.AiAgentAction, error) {
	if conv.BookingID == nil {
		return "", nil, errors.New("المحادثة بلا حجز")
	}
	c, err := s.repo.BookingContext(*conv.BookingID)
	if err != nil {
		return "", nil, err
	}
	msgs, err := s.repo.Messages(conv.ID)
	if err != nil || len(msgs) == 0 {
		return "", nil, err
	}
	hist := make([]AgentTurn, 0, len(msgs))
	for _, m := range msgs[:len(msgs)-1] {
		role := "USER"
		if m.Direction == "OUT" {
			role = "ASSISTANT"
		}
		hist = append(hist, AgentTurn{Role: role, Text: m.Text})
	}
	tc := &waTurnCtx{svc: s, conv: conv, booking: c}
	turn := WaTurn{Agent: conv.Agent, AgentName: s.name(conv.Agent), Context: s.contextText(c),
		History: hist, Message: msgs[len(msgs)-1].Text, Tools: tc.tools()}
	brain := s.brains[conv.Agent]
	if brain == nil {
		brain = waStubBrain{}
	}
	text, err := brain.Reply(ctx, turn)
	return text, tc.actions, err
}

// waStubBrain بلا مفتاح ذكاء: يشكر ويحوّل لإنسان — حتى التجربة تشتغل قبل عقل لوفي.
type waStubBrain struct{}

func (waStubBrain) Reply(_ context.Context, t WaTurn) (string, error) {
	for _, tool := range t.Tools {
		if tool.Name == "handoff" {
			_, _ = tool.Run(map[string]any{"reason": "الذكاء مطفي — يحتاج موظف"}, nil)
		}
	}
	return "شكراً إلك 🙏 وصلت رسالتك، وراح يكمل وياك موظف من شركة الأماني.", nil
}

// ── تدخّل الإنسان ──

func (s *WaAgentService) Takeover(convID, by string) error {
	return s.repo.SetStatus(convID, "HUMAN", &by)
}

func (s *WaAgentService) Release(convID string) error { return s.repo.SetStatus(convID, "BOT", nil) }

func (s *WaAgentService) Close(convID string) error { return s.repo.SetStatus(convID, "CLOSED", nil) }

// HumanSend رد إنسان — بس إذا المحادثة مستلمة.
func (s *WaAgentService) HumanSend(ctx context.Context, convID, by, text string) (*repository.WaMessage, error) {
	conv, err := s.repo.Conversation(convID)
	if err != nil {
		return nil, err
	}
	if conv.Status != "HUMAN" {
		return nil, errors.New("استلم المحادثة أول")
	}
	if !conv.Sandbox {
		if _, err := s.wa.SendText(ctx, conv.Phone, text); err != nil {
			return nil, err
		}
	}
	m, _, err := s.repo.AddMessage(conv.ID, "OUT", "HUMAN", &by, text, nil)
	return m, err
}

func (s *WaAgentService) handoff(conv *repository.WaConversation, reason string) {
	_ = s.repo.SetStatus(conv.ID, "HUMAN", nil)
	if conv.Sandbox || s.notif == nil {
		return
	}
	code := ""
	if conv.BookingCode != nil {
		code = " (حجز " + *conv.BookingCode + ")"
	}
	msg := "💬 " + s.name(conv.Agent) + " يحتاج موظف يكمل ويا الزبون" + code + ": " + reason
	_ = s.notif.CreateForRole("ADMIN", "wa_handoff", msg)
	if conv.Agent == WaAgentQuality {
		_ = s.notif.CreateForRole("QUALITY_ENGINEER", "wa_handoff", msg)
	}
}

// ── الإلغاء ──

// Undo يلغي إجراء الوكيل. إلغاء حجز سواه = طلب حذف عادي (مو حذف مباشر).
func (s *WaAgentService) Undo(actionID, by, reason string) (*repository.AiAgentAction, error) {
	a, err := s.repo.Action(actionID)
	if err != nil {
		return nil, err
	}
	if a.UndoneAt != nil {
		return nil, errors.New("الإجراء انلغى من قبل")
	}
	if !a.Sandbox && a.TargetID != nil {
		switch a.Kind {
		case "CREATE_BOOKING":
			if _, err := s.deletes.Create(*a.TargetID, by, "إلغاء حجز سواه "+s.name(a.Agent)+": "+reason,
				model.BookingDeleteChannelCallCenter, model.BookingDeleteTypeDataCorrection); err != nil {
				return nil, err
			}
		case "RECORD_CONTACT":
			if _, err := s.db.Exec(`UPDATE "QualityFollowUp" SET status = 'PENDING', "contactedAt" = NULL, "contactedByAgent" = NULL,
				"customerRating" = NULL, "crewRating" = NULL WHERE id = $1 AND status IN ('CONTACTED_OK', 'CONTACTED_ISSUE')`, *a.TargetID); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("هذا الإجراء ما ينلغى من هنا — عدّله من شاشته")
		}
	}
	if err := s.repo.MarkUndone(a.ID, by, reason); err != nil {
		return nil, err
	}
	return s.decorate(s.repo.Action(a.ID))
}

func (s *WaAgentService) decorate(a *repository.AiAgentAction, err error) (*repository.AiAgentAction, error) {
	if err != nil {
		return nil, err
	}
	a.AgentName = s.name(a.Agent)
	a.Undoable = a.UndoneAt == nil && (a.Sandbox || a.Kind == "CREATE_BOOKING" || a.Kind == "RECORD_CONTACT")
	return a, nil
}

// Actions الإجراءات ويا الاسم وقابلية الإلغاء.
func (s *WaAgentService) Actions(targetID, convID string) ([]repository.AiAgentAction, error) {
	rows, err := s.repo.Actions(targetID, convID)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		r := rows[i]
		d, _ := s.decorate(&r, nil)
		rows[i] = *d
	}
	return rows, nil
}

// ── الوارد من Meta ──

// Inbound رسالة زبون حقيقية: تروح لمحادثته المفتوحة، والوكيل يرد إذا ماسكها.
func (s *WaAgentService) Inbound(ctx context.Context, m whatsapp.Inbound) {
	conv, err := s.repo.OpenByPhone(m.From)
	if err != nil {
		log.Printf("[wa] رسالة من رقم بلا محادثة مفتوحة — تنتظر موظف")
		return
	}
	id := m.MetaID
	if _, _, err := s.Say(ctx, conv.ID, m.Text, &id); err != nil {
		log.Printf("[wa] inbound: %v", err)
	}
}

// ── المشغّل ──

// Scan يبدي محادثات جديدة: الجودة للمتابعات الجديدة، والحجوزات للحجوزات الي تنتظر تواصل.
// بس إذا المفتاح مشتغل والواتساب مربوط، وضمن حد WA_DAILY_NEW.
func (s *WaAgentService) Scan(ctx context.Context) {
	if !s.wa.On() {
		return
	}
	for _, agent := range []string{WaAgentQuality, WaAgentBooking} {
		if !s.Enabled(agent) {
			continue
		}
		left := s.DailyNew - s.repo.NewToday(agent)
		if left <= 0 {
			continue
		}
		if agent == WaAgentQuality {
			rows, err := s.repo.QualityCandidates(left)
			if err != nil {
				log.Printf("[wa] quality candidates: %v", err)
				continue
			}
			for _, r := range rows {
				fu := r.FollowUpID
				s.open(ctx, agent, r.BookingID, &fu)
			}
			continue
		}
		if !waWorkHours(time.Now()) {
			continue
		}
		ids, err := s.repo.BookingCandidates(left)
		if err != nil {
			log.Printf("[wa] booking candidates: %v", err)
			continue
		}
		for _, id := range ids {
			s.open(ctx, agent, id, nil)
		}
	}
}

// waWorkHours ٩ص–٨م بغداد — ما نراسل الزبون بنص الليل.
func waWorkHours(t time.Time) bool {
	h := t.In(debriefLoc).Hour()
	return h >= 9 && h < 20
}

func (s *WaAgentService) open(ctx context.Context, agent, bookingID string, followUpID *string) {
	c, err := s.repo.BookingContext(bookingID)
	if err != nil || c.Phone == "" {
		return
	}
	phone := waPhone(c.Phone)
	cust := c.CustomerID
	conv, err := s.repo.CreateConversation(agent, &cust, phone, &bookingID, followUpID, false)
	if err != nil {
		return // محادثة مفتوحة لهذا الرقم أصلاً
	}
	greet := s.Greeting(agent, c)
	if s.Template != "" {
		_, err = s.wa.SendTemplate(ctx, phone, s.Template, s.TemplateLang, []string{s.name(agent)})
	} else {
		_, err = s.wa.SendText(ctx, phone, greet)
	}
	if err != nil {
		log.Printf("[wa] open %s: %v", agent, err)
		_ = s.repo.SetStatus(conv.ID, "CLOSED", nil)
		return
	}
	_, _, _ = s.repo.AddMessage(conv.ID, "OUT", "AGENT", nil, greet, nil)
}

// waPhone رقم عراقي بصيغة Meta (9647…) — 07xx → 9647xx.
func waPhone(p string) string {
	d := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r >= '٠' && r <= '٩' {
			return '0' + (r - '٠')
		}
		return -1
	}, p)
	switch {
	case strings.HasPrefix(d, "00"):
		d = d[2:]
	case strings.HasPrefix(d, "07"):
		d = "964" + d[1:]
	case strings.HasPrefix(d, "7") && len(d) == 10:
		d = "964" + d
	}
	return d
}
