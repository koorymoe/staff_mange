package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
	"staffmange-api/internal/whatsapp"
)

// ═══ 💬 مسارات وكيلي الواتساب (دازاي 10-23) — العقد: docs/wa-api-contract.md ═══

type WaHandler struct {
	svc   *service.WaAgentService
	wa    *whatsapp.Client
	perms *repository.PermissionRepository
}

func NewWaHandler(svc *service.WaAgentService, wa *whatsapp.Client, perms *repository.PermissionRepository) *WaHandler {
	return &WaHandler{svc: svc, wa: wa, perms: perms}
}

// canAgent المدير والمالك كلشي؛ quality_control للجودة، coordinator للحجوزات.
func (h *WaHandler) canAgent(r *http.Request, agent string) bool {
	if middleware.IsOwner(r) || middleware.RoleFromContext(r) == "ADMIN" {
		return true
	}
	perm := "coordinator"
	if agent == service.WaAgentQuality {
		perm = "quality_control"
	}
	ok, err := h.perms.HasPermission(middleware.EmployeeIDFromContext(r), perm)
	return err == nil && ok
}

func (h *WaHandler) conv(w http.ResponseWriter, r *http.Request) *repository.WaConversation {
	c, err := h.svc.Repo().Conversation(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return nil
	}
	if !h.canAgent(r, c.Agent) || (c.Sandbox && !middleware.IsOwner(r)) {
		WriteError(w, http.StatusNotFound, "المحادثة مو موجودة")
		return nil
	}
	return c
}

// GET /api/wa/conversations?agent=&status=
func (h *WaHandler) List(w http.ResponseWriter, r *http.Request) {
	agent, status := r.URL.Query().Get("agent"), r.URL.Query().Get("status")
	rows, err := h.svc.Repo().Conversations(agent, status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المحادثات")
		return
	}
	owner := middleware.IsOwner(r)
	out := rows[:0]
	for _, c := range rows {
		if h.canAgent(r, c.Agent) && (!c.Sandbox || owner) {
			out = append(out, c)
		}
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/wa/conversations/{id}
func (h *WaHandler) Get(w http.ResponseWriter, r *http.Request) {
	c := h.conv(w, r)
	if c == nil {
		return
	}
	msgs, err := h.svc.Repo().Messages(c.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الرسائل")
		return
	}
	actions, _ := h.svc.Actions("", c.ID)
	h.svc.Repo().MarkRead(c.ID)
	WriteJSON(w, http.StatusOK, map[string]any{"conversation": c, "messages": msgs, "actions": actions})
}

func (h *WaHandler) setState(w http.ResponseWriter, r *http.Request, f func(c *repository.WaConversation) error) {
	c := h.conv(w, r)
	if c == nil {
		return
	}
	if err := f(c); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *WaHandler) Takeover(w http.ResponseWriter, r *http.Request) {
	h.setState(w, r, func(c *repository.WaConversation) error {
		return h.svc.Takeover(c.ID, middleware.EmployeeIDFromContext(r))
	})
}

func (h *WaHandler) Release(w http.ResponseWriter, r *http.Request) {
	h.setState(w, r, func(c *repository.WaConversation) error { return h.svc.Release(c.ID) })
}

func (h *WaHandler) Close(w http.ResponseWriter, r *http.Request) {
	h.setState(w, r, func(c *repository.WaConversation) error { return h.svc.Close(c.ID) })
}

func readText(w http.ResponseWriter, r *http.Request, field string) (string, bool) {
	var in map[string]any
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return "", false
	}
	s, _ := in[field].(string)
	s = strings.TrimSpace(s)
	if s == "" || len([]rune(s)) > 2000 {
		WriteError(w, http.StatusBadRequest, "اكتب نص (لحد ٢٠٠٠ حرف)")
		return "", false
	}
	return s, true
}

// POST /api/wa/conversations/{id}/send {text}
func (h *WaHandler) Send(w http.ResponseWriter, r *http.Request) {
	c := h.conv(w, r)
	if c == nil {
		return
	}
	text, ok := readText(w, r, "text")
	if !ok {
		return
	}
	m, err := h.svc.HumanSend(r.Context(), c.ID, middleware.EmployeeIDFromContext(r), text)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, m)
}

// ── الإجراءات ──

// GET /api/wa/actions?targetId=&conversationId=
func (h *WaHandler) Actions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Actions(r.URL.Query().Get("targetId"), r.URL.Query().Get("conversationId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإجراءات")
		return
	}
	owner := middleware.IsOwner(r)
	out := rows[:0]
	for _, a := range rows {
		if h.canAgent(r, a.Agent) && (!a.Sandbox || owner) {
			out = append(out, a)
		}
	}
	WriteJSON(w, http.StatusOK, out)
}

// POST /api/wa/actions/{id}/undo {reason}
func (h *WaHandler) Undo(w http.ResponseWriter, r *http.Request) {
	a, err := h.svc.Repo().Action(r.PathValue("id"))
	if err != nil || !h.canAgent(r, a.Agent) {
		WriteError(w, http.StatusNotFound, "الإجراء مو موجود")
		return
	}
	reason, ok := readText(w, r, "reason")
	if !ok {
		return
	}
	out, err := h.svc.Undo(a.ID, middleware.EmployeeIDFromContext(r), reason)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// ── مفتاح التشغيل (المالك) ──

// PUT /api/ai/agents/{agent}/enabled {enabled}
func (h *WaHandler) SetEnabled(w http.ResponseWriter, r *http.Request) {
	agent := r.PathValue("agent")
	if agent != service.WaAgentQuality && agent != service.WaAgentBooking {
		WriteError(w, http.StatusBadRequest, "وكيل غير معروف")
		return
	}
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	if err := h.svc.SetEnabled(agent, in.Enabled, middleware.EmployeeIDFromContext(r)); err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, whatsapp.ErrOff) {
			code = http.StatusConflict
		}
		WriteError(w, code, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"agent": agent, "enabled": in.Enabled})
}

// ── وضع التجربة (المالك) ──

// POST /api/wa/sandbox/start {agent, bookingId}
func (h *WaHandler) SandboxStart(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Agent     string `json:"agent"`
		BookingID string `json:"bookingId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	c, msgs, err := h.svc.StartSandbox(in.Agent, in.BookingID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]any{"conversation": c, "messages": msgs})
}

// POST /api/wa/sandbox/{id}/say {text}
func (h *WaHandler) SandboxSay(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.Repo().Conversation(r.PathValue("id"))
	if err != nil || !c.Sandbox {
		WriteError(w, http.StatusNotFound, "محادثة التجربة مو موجودة")
		return
	}
	text, ok := readText(w, r, "text")
	if !ok {
		return
	}
	msgs, actions, err := h.svc.Say(r.Context(), c.ID, text, nil)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if actions == nil {
		actions = []repository.AiAgentAction{}
	}
	WriteJSON(w, http.StatusOK, map[string]any{"messages": msgs, "actions": actions})
}

// ── Meta webhook (بلا دخول — التوقيع إجباري) ──

// GET /api/wa/webhook — تحقق Meta بأول ربط.
func (h *WaHandler) Verify(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("hub.mode") == "subscribe" && h.wa.VerifyToken(q.Get("hub.verify_token")) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(q.Get("hub.challenge")))
		return
	}
	http.Error(w, "forbidden", http.StatusForbidden)
}

// POST /api/wa/webhook
func (h *WaHandler) Receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil || !h.wa.ValidSignature(body, r.Header.Get("X-Hub-Signature-256")) {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	msgs, err := whatsapp.ParseWebhook(body)
	if err != nil {
		w.WriteHeader(http.StatusOK) // Meta تعيد الإرسال إذا رجّعنا خطأ — نتجاهل الجسم الغريب
		return
	}
	w.WriteHeader(http.StatusOK)
	// الرد يطلع بالخلفية — Meta تريد 200 بسرعة.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		for _, m := range msgs {
			h.svc.Inbound(ctx, m)
		}
	}()
	log.Printf("[wa] webhook: %d رسالة", len(msgs))
}
