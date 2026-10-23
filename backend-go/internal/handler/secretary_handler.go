package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
	"staffmange-api/internal/storage"
)

// SecretaryHandler مسارات «السكرتير» (العقد: docs/secretary.md) — المالك والمدير
// بس (requireAdmin)، وكل واحد بذاكرته: كل استعلام مقيّد بـemployeeId مالته.
type SecretaryHandler struct {
	repo          *repository.SecretaryRepository
	employees     *repository.EmployeeRepository
	tasks         *repository.ExtraTaskRepository
	notifications *repository.NotificationRepository
	voice         *service.VoiceAnalysisService
	brain         service.SecretaryReplier
	meetings      *repository.MeetingRequestRepository
	stt           interface {
		Transcribe(ctx context.Context, audio []byte, contentType string) (string, error)
	}
}

func NewSecretaryHandler(repo *repository.SecretaryRepository, e *repository.EmployeeRepository,
	t *repository.ExtraTaskRepository, n *repository.NotificationRepository, v *service.VoiceAnalysisService) *SecretaryHandler {
	return &SecretaryHandler{repo: repo, employees: e, tasks: t, notifications: n, voice: v}
}

// SetBrain العقل (لوفي). بدونه message يرجّع رسالة واضحة، والباقي يشتغل.
func (h *SecretaryHandler) SetBrain(b service.SecretaryReplier) { h.brain = b }

var secLoc = func() *time.Location {
	if l, err := time.LoadLocation("Asia/Baghdad"); err == nil {
		return l
	}
	return time.FixedZone("Asia/Baghdad", 3*3600)
}()

const secretaryOff = "ايريس ما تكدر تفكر هسه (الذكاء مطفي أو خلص رصيد اليوم). ملخص اليوم و«ينتظر قرارك» يشتغلون عادي."

// POST /api/secretary/message — JSON {text} أو multipart بحقل file.
func (h *SecretaryHandler) Message(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	var text, convID string
	var transcript *string
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
		if h.stt == nil && (h.voice == nil || !h.voice.Enabled()) {
			WriteError(w, http.StatusServiceUnavailable, "تحويل الصوت مو مشغّل على السيرفر — اكتب بدلها")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxVoiceBytes+64<<10)
		if err := r.ParseMultipartForm(maxVoiceBytes); err != nil {
			WriteError(w, http.StatusBadRequest, "البصمة أطول من المسموح (٣ دقايق)")
			return
		}
		convID = strings.TrimSpace(r.FormValue("conversationId"))
		file, _, err := r.FormFile("file")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "ماكو بصمة بالطلب")
			return
		}
		defer file.Close()
		data, err := storage.ReadLimited(file, maxVoiceBytes)
		if err != nil {
			WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
			return
		}
		ct := sniffAudio(data)
		if ct == "" {
			WriteError(w, http.StatusBadRequest, "صيغة الصوت مو مدعومة")
			return
		}
		t, err := h.transcribe(r.Context(), data, ct)
		if err != nil {
			WriteError(w, http.StatusBadGateway, "ما گدرت أفهم البصمة — عيدها أو اكتب")
			return
		}
		text = strings.TrimSpace(t)
		transcript = &text
	} else {
		var body struct {
			Text           string `json:"text"`
			ConversationID string `json:"conversationId"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&body); err != nil {
			WriteError(w, http.StatusBadRequest, "طلب غير صالح")
			return
		}
		text = strings.TrimSpace(body.Text)
		convID = strings.TrimSpace(body.ConversationID)
	}
	if text == "" {
		WriteError(w, http.StatusBadRequest, "الرسالة فاضية")
		return
	}
	if r := []rune(text); len(r) > 2000 {
		text = string(r[:2000])
	}

	// 💬 المحادثة: المحددة (لازم مالته) أو آخر وحدة (دازاي 10-23: محادثات متعددة)
	var conv *repository.SecretaryConversation
	var err error
	if convID != "" {
		conv, err = h.repo.Conversation(me, convID)
	} else {
		conv, err = h.repo.LatestConversation(me)
	}
	if err != nil {
		WriteError(w, http.StatusNotFound, "المحادثة مو موجودة")
		return
	}
	// 🧠 الذاكرة: ملخص المحادثة + آخر رسائل بس — الكلفة ثابتة مهما طالت
	past, _ := h.repo.ConversationHistory(conv.ID, service.MemoryKeep)
	if _, err := h.repo.SaveMessageIn(&conv.ID, me, "user", text, transcript, nil, nil); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الرسالة")
		return
	}

	res := &service.SecretaryResult{Reply: secretaryOff}
	if h.brain != nil {
		turn := service.SecretaryTurn{UserID: me, Text: text, Prefs: h.repo.Prefs(me), Summary: conv.Summary}
		if e, err := h.employees.FindByID(me); err == nil && e != nil {
			turn.UserName = e.Name
			turn.UserTitle = service.SecretaryHonorific(e.Certificate)
			turn.UserRole = "مدير"
			if middleware.IsOwner(r) {
				turn.UserRole = "مالك"
			}
		}
		for _, m := range past {
			turn.History = append(turn.History, service.SecretaryHistory{Role: m.Role, Text: m.Text})
		}
		out, err := h.brain.Reply(r.Context(), turn)
		switch {
		case err == nil && out != nil:
			res = out
		case errors.Is(err, service.ErrSecretaryOff), errors.Is(err, service.ErrAgentOff):
		case errors.Is(err, service.ErrSecretaryCap):
			res = &service.SecretaryResult{Reply: service.ErrSecretaryCap.Error()}
		default:
			log.Printf("secretary reply: %v", err)
			msg := service.AIErrorText(err)
			if msg == "" {
				msg = "صار خلل بالتفكير — جرّب مرة ثانية بعد شوية."
			}
			res = &service.SecretaryResult{Reply: msg}
		}
	}
	if res.Cards == nil {
		res.Cards = []service.SecretaryCard{}
	}
	// (دازاي 10-23) المهمة الكاملة (موظف واحد + موعد) تنرسل فوراً بلا «إرسال».
	if p := res.Proposal; p != nil && p.Kind == "task" && p.Fields.EmployeeID != "" && p.Fields.DueAt != "" {
		if card, _, err := h.sendTask(me, p.ActionID, p.Fields); err == nil {
			res.Cards = append(res.Cards, *card)
			res.Reply = strings.TrimSpace(res.Reply + "\n✅ انرسلت المهمة لـ" + card.Lines[0][len("انرسلت لـ"):])
			res.Proposal = nil
		} else {
			res.Reply = strings.TrimSpace(res.Reply + "\n⚠️ ما گدرت أرسلها: " + err.Error())
		}
	}
	var actionID *string
	if res.Proposal != nil {
		actionID = &res.Proposal.ActionID
		if res.Proposal.Candidates == nil {
			res.Proposal.Candidates = []service.SecretaryCandidate{}
		}
	}
	id, err := h.repo.SaveMessageIn(&conv.ID, me, "assistant", res.Reply, nil, res.Cards, actionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الجواب")
		return
	}
	h.repo.TitleIfNew(conv.ID, text)
	go h.updateMemory(*conv)
	WriteJSON(w, http.StatusOK, map[string]any{
		"messageId": id, "transcript": transcript, "reply": res.Reply,
		"cards": res.Cards, "proposal": res.Proposal, "conversationId": conv.ID,
	})
}

// POST /api/secretary/actions/{id}/confirm — «إرسال»: هنا بس تنرسل المهمة.
func (h *SecretaryHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	a, err := h.repo.Action(r.PathValue("id"), me)
	if err != nil {
		WriteError(w, http.StatusNotFound, "المقترح مو موجود")
		return
	}
	if a.Status != "PROPOSED" {
		WriteError(w, http.StatusConflict, repository.ErrSecretaryActionDone.Error())
		return
	}
	if a.Kind != "task" {
		WriteError(w, http.StatusBadRequest, "نوع مقترح مو مدعوم")
		return
	}
	var f service.SecretaryTaskFields
	_ = json.Unmarshal(a.Fields, &f)
	var body struct {
		Fields *service.SecretaryTaskFields `json:"fields"`
	}
	if r.ContentLength != 0 {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&body); err != nil {
			WriteError(w, http.StatusBadRequest, "طلب غير صالح")
			return
		}
	}
	if body.Fields != nil {
		f = *body.Fields
	}
	card, code, err := h.sendTask(me, a.ID, f)
	if err != nil {
		WriteError(w, code, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "card": card})
}

// POST /api/secretary/actions/{id}/cancel
func (h *SecretaryHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	if _, err := h.repo.Action(r.PathValue("id"), me); err != nil {
		WriteError(w, http.StatusNotFound, "المقترح مو موجود")
		return
	}
	if err := h.repo.Decide(r.PathValue("id"), me, "CANCELLED", "انلغى", nil); err != nil {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GET /api/secretary/today — من النظام مباشرة، يشتغل حتى بلا ذكاء.
func (h *SecretaryHandler) Today(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	greet := "مساء الخير"
	if hr := time.Now().In(secLoc).Hour(); hr >= 4 && hr < 12 {
		greet = "صباح الخير"
	}
	if e, err := h.employees.FindByID(me); err == nil && e != nil {
		greet += " " + strings.Fields(e.Name + " ")[0]
	}
	t, err := h.repo.Today()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أرقام اليوم")
		return
	}
	lines := []string{
		fmt.Sprintf("%d حجز اليوم · %d خلص · %d بالميدان", t.BookingsToday, t.CompletedToday, t.InField),
		fmt.Sprintf("%d من %d موظف مداوم هسه", t.PresentNow, t.ActiveStaff),
	}
	if t.Overdue > 0 {
		lines = append(lines, fmt.Sprintf("%d حجز متأخر", t.Overdue))
	}
	if t.NeedsCoord > 0 {
		lines = append(lines, fmt.Sprintf("%d حجز ينتظر تنسيق", t.NeedsCoord))
	}
	if t.OpenComplaints > 0 {
		lines = append(lines, fmt.Sprintf("%d شكوى مفتوحة", t.OpenComplaints))
	}
	if t.OpenTasks > 0 {
		lines = append(lines, fmt.Sprintf("%d مهمة مفتوحة", t.OpenTasks))
	}
	pending, err := h.repo.Pending()
	if err != nil {
		pending = []repository.SecretaryPending{}
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"greeting": greet,
		"summary":  map[string]any{"lines": lines, "cards": []service.SecretaryCard{}},
		"pending":  pending,
	})
}

// GET /api/secretary/history?limit=50
func (h *SecretaryHandler) History(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	me := middleware.EmployeeIDFromContext(r)
	// ?conversationId= محادثة محددة؛ بلاه آخر محادثة (التوافق ويا الواجهة القديمة)
	var conv *repository.SecretaryConversation
	var err error
	if id := r.URL.Query().Get("conversationId"); id != "" {
		conv, err = h.repo.Conversation(me, id)
	} else {
		conv, err = h.repo.LatestConversation(me)
	}
	if err != nil {
		WriteError(w, http.StatusNotFound, "المحادثة مو موجودة")
		return
	}
	rows, err := h.repo.ConversationHistory(conv.ID, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب السجل")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, m := range rows {
		cards := m.Cards
		if len(cards) == 0 || string(cards) == "null" {
			cards = json.RawMessage("[]")
		}
		item := map[string]any{"id": m.ID, "at": m.At, "role": m.Role, "text": m.Text,
			"transcript": m.Transcript, "cards": cards, "action": nil}
		if m.ActionID != nil {
			item["action"] = map[string]any{"id": *m.ActionID, "status": deref(m.ActionStatus), "result": m.ActionResult}
		}
		out = append(out, item)
	}
	WriteJSON(w, http.StatusOK, out)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ═══ «تذكيراتي» — الي تنتظر قرار أو موعد من المالك أو المدير ═══

// SetMeetings طلبات المقابلة والضيوف.
func (h *SecretaryHandler) SetMeetings(m *repository.MeetingRequestRepository) { h.meetings = m }

type secretaryReminder struct {
	ID          string     `json:"id"`
	Kind        string     `json:"kind"` // meeting | guest | delete_request | leave
	Title       string     `json:"title"`
	Who         string     `json:"who"`
	Reason      string     `json:"reason"`
	GuestName   *string    `json:"guestName"`
	PreferredAt *time.Time `json:"preferredAt"`
	Link        *string    `json:"link"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// GET /api/secretary/reminders
func (h *SecretaryHandler) Reminders(w http.ResponseWriter, r *http.Request) {
	counts := map[string]int{"meeting": 0, "guest": 0, "delete_request": 0, "leave": 0}
	items := []secretaryReminder{}
	if h.meetings != nil {
		rows, err := h.meetings.Pending()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات المقابلة")
			return
		}
		for _, m := range rows {
			it := secretaryReminder{ID: m.ID, Who: m.Who, Reason: m.Reason, GuestName: m.GuestName,
				PreferredAt: m.PreferredAt, CreatedAt: m.CreatedAt}
			if m.Kind == "GUEST" {
				it.Kind, it.Title = "guest", m.Who+" يريد يجيب ضيف: "+deref(m.GuestName)
			} else {
				it.Kind, it.Title = "meeting", "طلب مقابلة — "+m.Who
			}
			counts[it.Kind]++
			items = append(items, it)
		}
	}
	pending, err := h.repo.Pending()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الطلبات")
		return
	}
	for _, p := range pending {
		link := p.Link
		counts[p.Kind]++
		items = append(items, secretaryReminder{ID: p.ID, Kind: p.Kind, Title: p.Question, Who: p.Affects,
			Reason: p.Why, Link: &link, CreatedAt: p.CreatedAt})
	}
	WriteJSON(w, http.StatusOK, map[string]any{"counts": counts, "items": items})
}

// POST /api/secretary/reminders/{id}/schedule — {at, note}: موعد للمقابلة أو الضيف.
func (h *SecretaryHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	var in struct {
		At   string `json:"at"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	at, err := time.ParseInLocation("2006-01-02T15:04", in.At, secLoc)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "حدد التاريخ والوقت")
		return
	}
	if at.Before(time.Now().Add(-5 * time.Minute)) {
		WriteError(w, http.StatusBadRequest, "الموعد لازم يكون بالجاي")
		return
	}
	note := strings.TrimSpace(in.Note)
	if r := []rune(note); len(r) > 300 {
		note = string(r[:300])
	}
	m := h.decideMeeting(w, r, "SCHEDULED", &at, note)
	if m == nil {
		return
	}
	msg := "📅 انحدد موعد " + map[bool]string{true: "ضيفك", false: "مقابلتك"}[m.Kind == "GUEST"] +
		" ويا الإدارة: " + at.In(secLoc).Format("2006-01-02 15:04")
	if note != "" {
		msg += " — " + note
	}
	_ = h.notifications.Create(m.EmployeeID, "meeting_request", msg)
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /api/secretary/reminders/{id}/decline — {reason}
func (h *SecretaryHandler) Decline(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	reason := strings.TrimSpace(in.Reason)
	if n := len([]rune(reason)); n < 2 || n > 300 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الرفض")
		return
	}
	m := h.decideMeeting(w, r, "DECLINED", nil, reason)
	if m == nil {
		return
	}
	_ = h.notifications.Create(m.EmployeeID, "meeting_request", "🗓️ طلبك ما انقبل: "+reason)
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *SecretaryHandler) decideMeeting(w http.ResponseWriter, r *http.Request, status string, at *time.Time, note string) *repository.MeetingRequest {
	if h.meetings == nil {
		WriteError(w, http.StatusNotFound, "الطلب مو موجود")
		return nil
	}
	m, err := h.meetings.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "هذا مو طلب مقابلة — الحذف والإجازات تنحسم من شاشتها")
		return nil
	}
	if err := h.meetings.Decide(m.ID, middleware.EmployeeIDFromContext(r), status, at, note); err != nil {
		WriteError(w, http.StatusConflict, err.Error())
		return nil
	}
	return m
}

// ═══ «مذكراتي» و«جدول أعمالي» (دازاي 10-23) — كل مستخدم ملاحظاته بس ═══

// GET /api/secretary/notes?kind=DIARY|AGENDA&q=
func (h *SecretaryHandler) Notes(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	if kind != "" && kind != "DIARY" && kind != "AGENDA" {
		WriteError(w, http.StatusBadRequest, "نوع غير معروف")
		return
	}
	rows, err := h.repo.Notes(middleware.EmployeeIDFromContext(r), kind, r.URL.Query().Get("q"), 200)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الملاحظات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/secretary/notes — {kind, body, at?}: at بصيغة 2006-01-02T15:04 بتوقيت بغداد.
func (h *SecretaryHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Kind string `json:"kind"`
		Body string `json:"body"`
		At   string `json:"at"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		WriteError(w, http.StatusBadRequest, "اكتب شي أول")
		return
	}
	if rs := []rune(body); len(rs) > 5000 {
		body = string(rs[:5000])
	}
	if in.Kind != "DIARY" && in.Kind != "AGENDA" {
		WriteError(w, http.StatusBadRequest, "نوع غير معروف")
		return
	}
	var at *time.Time
	if s := strings.TrimSpace(in.At); s != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", s, secLoc)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "صيغة الوقت غلط")
			return
		}
		at = &t
	}
	n, err := h.repo.AddNote(middleware.EmployeeIDFromContext(r), in.Kind, body, at)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusCreated, n)
}

// PUT /api/secretary/notes/{id}/done — {done}
func (h *SecretaryHandler) NoteDone(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Done bool `json:"done"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	if err := h.repo.SetNoteDone(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.Done); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /api/secretary/notes/{id} — ملاحظته هو بس.
func (h *SecretaryHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteNote(middleware.EmployeeIDFromContext(r), r.PathValue("id")); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GET /api/secretary/agenda — المقابلات الي انحدد موعدها + بنوده، من اليوم وبعد.
func (h *SecretaryHandler) Agenda(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Agenda(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب جدول الأعمال")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// sendTask يرسل مهمة ايريس للموظف — من «إرسال» أو تلقائياً (دازاي 10-23: «المهمة توصل
// من السكرتير مباشرة للنظام»). يرجّع البطاقة، أو كود HTTP ويا الخطأ.
func (h *SecretaryHandler) sendTask(me, actionID string, f service.SecretaryTaskFields) (*service.SecretaryCard, int, error) {
	f.Title = strings.TrimSpace(f.Title)
	if f.Title == "" {
		return nil, http.StatusBadRequest, errors.New("عنوان المهمة فاضي")
	}
	if r := []rune(f.Title); len(r) > 120 {
		f.Title = string(r[:120])
	}
	if f.Priority != "URGENT" {
		f.Priority = "NORMAL"
	}
	emp, err := h.employees.FindByID(f.EmployeeID)
	if f.EmployeeID == "" || err != nil || emp == nil || emp.Status != "ACTIVE" {
		return nil, http.StatusBadRequest, errors.New("اختار الموظف أول")
	}
	if emp.ID == me {
		return nil, http.StatusBadRequest, errors.New("ما تكدر توجّه مهمة لنفسك")
	}
	f.EmployeeName = emp.Name
	req := model.CreateExtraTaskRequest{Title: f.Title, AssignedToID: emp.ID, Priority: f.Priority}
	if d := strings.TrimSpace(f.Description); d != "" {
		req.Description = &d
	}
	// (ع) 10-23: الموعد إجباري لكل مهمة.
	if f.DueAt == "" {
		return nil, http.StatusBadRequest, errors.New("حدد آخر موعد للإنجاز قبل الإرسال")
	}
	t, err := time.ParseInLocation("2006-01-02T15:04", f.DueAt, secLoc)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("صيغة الموعد غلط")
	}
	due := t.Format(time.RFC3339)
	req.DueAt = &due
	// القفل أول: لو انضغط «إرسال» مرتين بنفس اللحظة، وحدة بس تمر
	if err := h.repo.Decide(actionID, me, "SENT", "", f); err != nil {
		return nil, http.StatusConflict, err
	}
	task, err := h.tasks.Create(req, me)
	if err != nil {
		_ = h.repo.MarkFailed(actionID, me, "تعذر إنشاء المهمة")
		return nil, http.StatusBadRequest, errors.New("تعذر إنشاء المهمة: " + err.Error())
	}
	_ = h.repo.SetResult(actionID, me, "انرسلت لـ"+emp.Name+" · "+task.ID)
	_ = h.notifications.Create(task.AssignedToID, "extra_task", "📋 مهمة جديدة موجّهة إلك: "+task.Title)
	lines := []string{"انرسلت لـ" + emp.Name}
	if f.Priority == "URGENT" {
		lines = append(lines, "عاجلة")
	}
	lines = append(lines, "موعدها "+strings.Replace(f.DueAt, "T", " ", 1))
	now := time.Now().UTC().Format(time.RFC3339)
	link := "/extra-tasks"
	return &service.SecretaryCard{Kind: "task", Title: task.Title, Status: "انرسلت", Owner: &emp.Name,
		UpdatedAt: &now, Lines: lines, Link: &link}, http.StatusOK, nil
}

// ═══ 🎙️ فويس ايريس: Gemini أول (أدق وأسرع بالعراقي)، وإذا فشل Whisper الداخلي ═══

// SetSTT محوّل الصوت الخارجي (Gemini) — nil = Whisper بس.
func (h *SecretaryHandler) SetSTT(t interface {
	Transcribe(ctx context.Context, audio []byte, contentType string) (string, error)
}) {
	h.stt = t
}

func (h *SecretaryHandler) transcribe(ctx context.Context, data []byte, ct string) (string, error) {
	if h.stt != nil {
		if t, err := h.stt.Transcribe(ctx, data, ct); err == nil {
			return t, nil
		} else {
			log.Printf("[iris-voice] gemini: %v — نرجع لـWhisper", err)
		}
	}
	if h.voice == nil || !h.voice.Enabled() {
		return "", errors.New("تحويل الصوت مو متوفر")
	}
	return h.voice.Transcribe(ctx, data, ct)
}

// updateMemory يدمج الرسائل القديمة بملخص المحادثة إذا تجمّعن (بالخلفية، ما يأخّر الرد).
func (h *SecretaryHandler) updateMemory(c repository.SecretaryConversation) {
	rem, ok := h.brain.(interface {
		Remember(ctx context.Context, prev string, turns []service.AgentTurn) (string, error)
	})
	if !ok {
		return
	}
	msgs, err := h.repo.ToSummarize(&c, service.MemoryKeep, service.MemoryThreshold)
	if err != nil || len(msgs) == 0 {
		return
	}
	turns := make([]service.AgentTurn, 0, len(msgs))
	for _, m := range msgs {
		turns = append(turns, service.AgentTurn{Role: m.Role, Text: m.Text})
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	sum, err := rem.Remember(ctx, c.Summary, turns)
	if err != nil {
		log.Printf("[iris-memory] %v", err)
		return
	}
	_ = h.repo.SetMemory(c.ID, sum, msgs[len(msgs)-1].At)
}

// ── 💬 مسارات المحادثات ──

// GET /api/secretary/conversations
func (h *SecretaryHandler) Conversations(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Conversations(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المحادثات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/secretary/conversations {title?}
func (h *SecretaryHandler) NewConversation(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title string `json:"title"`
	}
	if r.ContentLength != 0 {
		_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<10)).Decode(&in)
	}
	c, err := h.repo.CreateConversation(middleware.EmployeeIDFromContext(r), strings.TrimSpace(in.Title))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر فتح محادثة")
		return
	}
	WriteJSON(w, http.StatusCreated, c)
}

// PATCH /api/secretary/conversations/{id} {title}
func (h *SecretaryHandler) RenameConversation(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	t := strings.TrimSpace(in.Title)
	if n := len([]rune(t)); n < 1 || n > 80 {
		WriteError(w, http.StatusBadRequest, "العنوان من ١ إلى ٨٠ حرف")
		return
	}
	if err := h.repo.RenameConversation(middleware.EmployeeIDFromContext(r), r.PathValue("id"), t); err != nil {
		WriteError(w, http.StatusNotFound, "المحادثة مو موجودة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /api/secretary/conversations/{id} — محادثته هو بس.
func (h *SecretaryHandler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteConversation(middleware.EmployeeIDFromContext(r), r.PathValue("id")); err != nil {
		WriteError(w, http.StatusNotFound, "المحادثة مو موجودة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}
