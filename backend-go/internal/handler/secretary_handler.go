package handler

import (
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

const secretaryOff = "السكرتير ما يكدر يفكر هسه (الذكاء مطفي أو خلص رصيد اليوم). ملخص اليوم و«ينتظر قرارك» يشتغلون عادي."

// POST /api/secretary/message — JSON {text} أو multipart بحقل file.
func (h *SecretaryHandler) Message(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	var text string
	var transcript *string
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
		if h.voice == nil || !h.voice.Enabled() {
			WriteError(w, http.StatusServiceUnavailable, "تحويل الصوت مو مشغّل على السيرفر — اكتب بدلها")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxVoiceBytes+64<<10)
		if err := r.ParseMultipartForm(maxVoiceBytes); err != nil {
			WriteError(w, http.StatusBadRequest, "البصمة أطول من المسموح (٣ دقايق)")
			return
		}
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
		t, err := h.voice.Transcribe(r.Context(), data, ct)
		if err != nil {
			WriteError(w, http.StatusBadGateway, "ما گدرت أفهم البصمة — عيدها أو اكتب")
			return
		}
		text = strings.TrimSpace(t)
		transcript = &text
	} else {
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&body); err != nil {
			WriteError(w, http.StatusBadRequest, "طلب غير صالح")
			return
		}
		text = strings.TrimSpace(body.Text)
	}
	if text == "" {
		WriteError(w, http.StatusBadRequest, "الرسالة فاضية")
		return
	}
	if r := []rune(text); len(r) > 2000 {
		text = string(r[:2000])
	}

	// السجل قبل الرسالة الجديدة — العقل ياخذ آخر ١٢ رسالة
	past, _ := h.repo.History(me, 12)
	if _, err := h.repo.SaveMessage(me, "user", text, transcript, nil, nil); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الرسالة")
		return
	}

	res := &service.SecretaryResult{Reply: secretaryOff}
	if h.brain != nil {
		turn := service.SecretaryTurn{UserID: me, Text: text, Prefs: h.repo.Prefs(me)}
		if e, err := h.employees.FindByID(me); err == nil && e != nil {
			turn.UserName = e.Name
		}
		for _, m := range past {
			turn.History = append(turn.History, service.SecretaryHistory{Role: m.Role, Text: m.Text})
		}
		out, err := h.brain.Reply(r.Context(), turn)
		switch {
		case err == nil && out != nil:
			res = out
		case errors.Is(err, service.ErrAgentOff):
		default:
			log.Printf("secretary reply: %v", err)
			res = &service.SecretaryResult{Reply: "صار خلل بالتفكير — جرّب مرة ثانية بعد شوية."}
		}
	}
	if res.Cards == nil {
		res.Cards = []service.SecretaryCard{}
	}
	var actionID *string
	if res.Proposal != nil {
		actionID = &res.Proposal.ActionID
		if res.Proposal.Candidates == nil {
			res.Proposal.Candidates = []service.SecretaryCandidate{}
		}
	}
	id, err := h.repo.SaveMessage(me, "assistant", res.Reply, nil, res.Cards, actionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الجواب")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"messageId": id, "transcript": transcript, "reply": res.Reply,
		"cards": res.Cards, "proposal": res.Proposal,
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
	f.Title = strings.TrimSpace(f.Title)
	if f.Title == "" {
		WriteError(w, http.StatusBadRequest, "عنوان المهمة فاضي")
		return
	}
	if r := []rune(f.Title); len(r) > 120 {
		f.Title = string(r[:120])
	}
	if f.Priority != "URGENT" {
		f.Priority = "NORMAL"
	}
	emp, err := h.employees.FindByID(f.EmployeeID)
	if f.EmployeeID == "" || err != nil || emp == nil || emp.Status != "ACTIVE" {
		WriteError(w, http.StatusBadRequest, "اختار الموظف أول")
		return
	}
	if emp.ID == me {
		WriteError(w, http.StatusBadRequest, "ما تكدر توجّه مهمة لنفسك")
		return
	}
	f.EmployeeName = emp.Name
	req := model.CreateExtraTaskRequest{Title: f.Title, AssignedToID: emp.ID, Priority: f.Priority}
	if d := strings.TrimSpace(f.Description); d != "" {
		req.Description = &d
	}
	if f.DueAt != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", f.DueAt, secLoc)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "صيغة الموعد غلط")
			return
		}
		s := t.Format(time.RFC3339)
		req.DueAt = &s
	}
	// القفل أول: لو ضغط «إرسال» مرتين بنفس اللحظة، وحدة بس تمر
	if err := h.repo.Decide(a.ID, me, "SENT", "", f); err != nil {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	task, err := h.tasks.Create(req, me)
	if err != nil {
		_ = h.repo.MarkFailed(a.ID, me, "تعذر إنشاء المهمة")
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء المهمة")
		return
	}
	_ = h.repo.SetResult(a.ID, me, "انرسلت لـ"+emp.Name+" · "+task.ID)
	_ = h.notifications.Create(task.AssignedToID, "extra_task", "📋 مهمة جديدة موجّهة إلك: "+task.Title)

	lines := []string{"انرسلت لـ" + emp.Name}
	if f.Priority == "URGENT" {
		lines = append(lines, "عاجلة")
	}
	if f.DueAt != "" {
		lines = append(lines, "موعدها "+strings.Replace(f.DueAt, "T", " ", 1))
	}
	now := time.Now().UTC().Format(time.RFC3339)
	link := "/extra-tasks"
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "card": service.SecretaryCard{
		Kind: "task", Title: task.Title, Status: "انرسلت", Owner: &emp.Name,
		UpdatedAt: &now, Lines: lines, Link: &link,
	}})
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
	rows, err := h.repo.History(middleware.EmployeeIDFromContext(r), limit)
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
