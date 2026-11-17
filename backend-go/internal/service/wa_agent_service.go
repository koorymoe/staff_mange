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
	// WaAgentReception الاستقبال (دازاي 11-15): الزبون هو الي يراسل أو يدز فويس أو يتصل، والوكيل يرد.
	// مو قيمة بعمود agent — محادثته BOOKING ومؤشّرة inbound، وإلها مفتاح تشغيل وعقل خاص.
	WaAgentReception = "RECEPTION"
	// WaAgentVoice المكالمة الحية (دازاي 11-15): نفس العقل والأدوات، بس الزبون يحچي والوكيل يرد بصوت.
	// هم مو قيمة بعمود agent — مفتاح تشغيل بس. التفاصيل بـwa_voice.go.
	WaAgentVoice = "VOICE"
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
	Voice     bool // الرد راح ينگال بصوت (مكالمة حية) — بلا إيموجي ولا أرقام مكتوبة
}

// WaBrain العقل — لوفي يبدّله بسياسته. Reply يرجّع نص الرد للزبون.
type WaBrain interface {
	Reply(ctx context.Context, t WaTurn) (string, error)
}

type waNotifier interface {
	Create(employeeID, kind, message string) error
	CreateForRole(role, kind, message string) error
	CreateForRolesOrPermission(roles []string, permissionName, kind, message string) error
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
	customers *repository.CustomerRepository
	slots     waSlots
	hear      WaTranscriber

	DailyNew        int
	Templates       map[string]string // agent ← اسم قالب Meta
	TemplateLang    string
	PrivacyApproved bool
	// VoiceReady سر ربط منصة الصوت محطوط (VOICE_LLM_SECRET) — بلاه مفتاح المكالمة ما يشتغل.
	VoiceReady bool
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
	switch agent {
	case WaAgentQuality:
		return model.SwitchAIQualityAgent
	case WaAgentReception:
		return model.SwitchAIReceptionAgent
	case WaAgentVoice:
		return model.SwitchAIVoiceCalls
	}
	return model.SwitchAIBookingAgent
}

// convAgent الوكيل الي يمسك المحادثة: الي بدأها الزبون للاستقبال، والباقي لوكيلها.
func convAgent(c *repository.WaConversation) string {
	if c.Inbound {
		return WaAgentReception
	}
	return c.Agent
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
	if agent == WaAgentVoice {
		// المكالمة الحية تجي من منصة الصوت، مو من ربط Meta مالتنا — شرطها سر الربط.
		if on && !s.VoiceReady {
			return ErrVoiceOff
		}
		return s.switches.Set(waSwitch(agent), on, by)
	}
	if on && !s.wa.On() {
		return whatsapp.ErrOff
	}
	return s.switches.Set(waSwitch(agent), on, by)
}

func (s *WaAgentService) name(agent string) string {
	if agent == WaAgentReception || agent == WaAgentVoice {
		agent = WaAgentBooking // الاستقبال هو ابو الحجوزات نفسه لمّا الزبون يبدي — نفس الاسم والمفتاح والميزانية
	}
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
	if agent == WaAgentReception {
		// تجربة الاستقبال: المالك يكتب كأنه زبون جديد يراسل أول — بلا حجز.
		conv, err := s.repo.CreateInbound(WaAgentBooking, nil, "sandbox", true)
		if err != nil {
			return nil, nil, err
		}
		return conv, []repository.WaMessage{}, nil
	}
	if agent != WaAgentQuality && agent != WaAgentBooking {
		return nil, nil, errors.New("وكيل غير معروف")
	}
	// (دازاي 11-17) المالك يكتب كود الحجز (B1206) — المعرّف الداخلي ما يبين بالشاشات.
	bookingID = s.repo.ResolveBookingRef(bookingID)
	c, err := s.repo.BookingContext(bookingID)
	if err != nil {
		return nil, nil, errors.New("الحجز مو موجود — اكتب كوده مثل B1206")
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
	return s.say(ctx, convID, text, metaID, false)
}

// say قلب الدور. voice = الزبون عالمكالمة: الرد يرجع للمنصة تنطقه (ما ينرسل واتساب)، ومفتاحه مفتاح المكالمة.
func (s *WaAgentService) say(ctx context.Context, convID, text string, metaID *string, voice bool) ([]repository.WaMessage, []repository.AiAgentAction, error) {
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
	gate := convAgent(conv)
	if voice {
		gate = WaAgentVoice
	}
	if !conv.Sandbox && !s.Enabled(gate) {
		return out, nil, nil
	}
	reply, actions, err := s.replyTurn(ctx, conv, voice)
	if err != nil {
		log.Printf("[wa] %s reply: %v", convAgent(conv), err)
		reply = "شكراً إلك 🙏 راح يتواصل وياك موظف من شركة الأماني بأقرب وقت."
		s.handoff(conv, "خلل بالوكيل: "+err.Error())
	}
	if strings.TrimSpace(reply) == "" {
		return out, actions, nil
	}
	if voice {
		reply = speakable(reply)
	} else if !conv.Sandbox {
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
	return s.replyTurn(ctx, conv, false)
}

func (s *WaAgentService) replyTurn(ctx context.Context, conv *repository.WaConversation, voice bool) (string, []repository.AiAgentAction, error) {
	var c *repository.WaBookingContext
	if conv.BookingID != nil {
		bc, err := s.repo.BookingContext(*conv.BookingID)
		if err != nil {
			return "", nil, err
		}
		c = bc
	} else if !conv.Inbound {
		return "", nil, errors.New("المحادثة بلا حجز")
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
	agent := convAgent(conv)
	if agent == WaAgentReception && len(msgs) > waReceptionMaxMsgs {
		// محادثة طالت بلا نتيجة — إنسان يكمل، والكلفة ما تفلت.
		s.handoff(conv, "محادثة الاستقبال طالت — تحتاج موظف")
		return waHandoffReply, nil, nil
	}
	turn := WaTurn{Agent: agent, AgentName: s.name(agent), Context: s.turnContext(conv, c),
		History: hist, Message: msgs[len(msgs)-1].Text, Tools: tc.tools(), Voice: voice}
	brain := s.brains[agent]
	if brain == nil {
		brain = waStubBrain{}
	}
	text, err := brain.Reply(ctx, turn)
	if err == nil && !voice && agent == WaAgentReception && len(hist) == 0 && strings.TrimSpace(text) != "" && !strings.Contains(text, "مساعد ذكي") {
		// أول رد للزبون لازم يعرّف إنه مساعد ذكي — ما نعتمد بيها على النموذج.
		text = "هلا بيك 🌹 وياك " + turn.AgentName + "، مساعد ذكي من شركة الأماني.\n" + text
	}
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
	msg := "💬 " + s.name(convAgent(conv)) + " يحتاج موظف يكمل ويا الزبون" + code + ": " + reason
	// (دازاي 11-04) التنبيه للمالك والي عنده صلاحية المحادثات بس
	_ = s.notif.CreateForRolesOrPermission([]string{"OWNER"}, "whatsapp_chats", "wa_handoff", msg)
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

// ── المشغّل ──

// Scan يبدي محادثات جديدة: الجودة للمتابعات الجديدة، والحجوزات للحجوزات الي تنتظر تواصل.
// بس إذا المفتاح مشتغل والواتساب مربوط، وضمن حد WA_DAILY_NEW.
func (s *WaAgentService) Scan(ctx context.Context) {
	if !s.wa.On() {
		return
	}
	if n := s.repo.CloseIdleInbound(); n > 0 {
		log.Printf("[wa] انسدّت %d محادثة استقبال ساكتة", n)
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
	if tpl := s.Templates[agent]; tpl != "" {
		// القالب بيه متغيّر واحد {{1}} = اسم الوكيل (من إعدادات المالك).
		_, err = s.wa.SendTemplate(ctx, phone, tpl, s.TemplateLang, []string{s.name(agent)})
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
