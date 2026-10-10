package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ 💬 دردشة ماتركس — المالك والمدير (قرار (ع) 10-10) ═══
// «نسولف ويا ويرد مثل أي ذكاء اصطناعي». ماتركس يتذكر المحادثة، ويفتح أدواته
// (قراءة بس) لما يحتاج أرقام الشركة.
type MatrixChatHandler struct {
	db    *sqlx.DB
	agent *service.MatrixAgent
}

func NewMatrixChatHandler(db *sqlx.DB, agent *service.MatrixAgent) *MatrixChatHandler {
	return &MatrixChatHandler{db: db, agent: agent}
}

const chatDailyLimit = 200

const chatSystemPrompt = `أنت «ماتركس»، الذكاء الاصطناعي لشركة الأماني بالعراق (كاميرات، جي بي اس، طاقة شمسية، أنظمة أمنية، مشاريع). تسولف ويا المالك أو المدير.
- سولف طبيعي باللهجة العراقية مثل صاحب ذكي: ردود حيّة ومتنوعة، مو قالب. جاوب على أي سؤال، حتى العام.
- لما الكلام عن الشركة (فلوس، فواتير، موظفين، حجوزات، مشاريع، مخازن، دوام، تقييم) افتح أدواتك وشوف بعينك قبل ما تحكم، واربط بين المصادر.
- لا تخترع رقم ما شفته بأداة. إذا ما تگدر تتأكد گول بصراحة.
- الموظفين يجوك كرموز «موظف#n» — استعمل نفس الرمز بالضبط.
- لا تقترح غرامة أو خصم فلوس، ولا تخمّن أسباب شخصية. اقتراحاتك عملية.
- أدواتك قراءة بس: إذا طلبوا منك تغيّر شي، گلهم شنو يسوون هم ومن وين.`

type chatRow struct {
	ID        string    `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	UpdatedAt time.Time `db:"updatedAt" json:"updatedAt"`
}

type chatMsg struct {
	ID        string    `db:"id" json:"id"`
	Role      string    `db:"role" json:"role"`
	Text      string    `db:"text" json:"text"`
	Steps     *string   `db:"steps" json:"steps"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

func (h *MatrixChatHandler) own(r *http.Request, id string) bool {
	var ok bool
	_ = h.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "MatrixTalk" WHERE id = $1 AND "employeeId" = $2)`, id, middleware.EmployeeIDFromContext(r))
	return ok
}

// GET /api/matrix/chats
func (h *MatrixChatHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []chatRow{}
	if err := h.db.Select(&rows, `SELECT id, title, "updatedAt" FROM "MatrixTalk" WHERE "employeeId" = $1 ORDER BY "updatedAt" DESC LIMIT 50`,
		middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المحادثات")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"chats": rows, "enabled": h.agent.Enabled("CHAT")})
}

// POST /api/matrix/chats
func (h *MatrixChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	var row chatRow
	if err := h.db.Get(&row, `INSERT INTO "MatrixTalk" (id, "employeeId") VALUES (gen_random_uuid()::text, $1) RETURNING id, title, "updatedAt"`,
		middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء المحادثة")
		return
	}
	WriteJSON(w, http.StatusOK, row)
}

// GET /api/matrix/chats/{id}
func (h *MatrixChatHandler) Messages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.own(r, id) {
		WriteError(w, http.StatusNotFound, "المحادثة غير موجودة")
		return
	}
	rows := []chatMsg{}
	_ = h.db.Select(&rows, `SELECT id, role, text, steps, "createdAt" FROM "MatrixTalkMessage" WHERE "chatId" = $1 ORDER BY "createdAt"`, id)
	WriteJSON(w, http.StatusOK, rows)
}

// DELETE /api/matrix/chats/{id}
func (h *MatrixChatHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`DELETE FROM "MatrixTalk" WHERE id = $1 AND "employeeId" = $2`, r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/matrix/chats/{id}/messages {text}
func (h *MatrixChatHandler) Send(w http.ResponseWriter, r *http.Request) {
	id, me := r.PathValue("id"), middleware.EmployeeIDFromContext(r)
	if !h.own(r, id) {
		WriteError(w, http.StatusNotFound, "المحادثة غير موجودة")
		return
	}
	var in struct {
		Text string `json:"text"`
	}
	if err := DecodeJSON(r, &in); err != nil || strings.TrimSpace(in.Text) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب رسالتك")
		return
	}
	text := strings.TrimSpace(in.Text)
	if len([]rune(text)) > 2000 {
		WriteError(w, http.StatusBadRequest, "الرسالة طويلة — خليها أقصر")
		return
	}
	if !h.agent.Enabled("CHAT") {
		WriteError(w, http.StatusServiceUnavailable, "ماتركس يحتاج مفتاح هايكو (ANTHROPIC_API_KEY) حتى يسولف — أو مفتاح الدردشة مطفي")
		return
	}
	var today int
	_ = h.db.Get(&today, `SELECT count(*) FROM "MatrixTalkMessage" m JOIN "MatrixTalk" c ON c.id = m."chatId"
		WHERE c."employeeId" = $1 AND m.role = 'USER' AND m."createdAt" > now() - interval '24 hours'`, me)
	if today >= chatDailyLimit {
		WriteError(w, http.StatusTooManyRequests, "وصلت حد الرسائل اليوم — باچر يرجع")
		return
	}
	hist := []chatMsg{}
	_ = h.db.Select(&hist, `SELECT * FROM (SELECT id, role, text, steps, "createdAt" FROM "MatrixTalkMessage" WHERE "chatId" = $1
		ORDER BY "createdAt" DESC LIMIT 20) x ORDER BY "createdAt"`, id)
	turns := make([]service.AgentTurn, 0, len(hist))
	for _, m := range hist {
		turns = append(turns, service.AgentTurn{Role: m.Role, Text: m.Text})
	}
	// التاريخ لازم يبدي برسالة مستخدم (النموذج يرفض غيرها)
	for len(turns) > 0 && turns[0].Role != "USER" {
		turns = turns[1:]
	}
	run, err := h.agent.Chat("CHAT", chatSystemPrompt, turns, text)
	if err != nil || strings.TrimSpace(run.Answer) == "" {
		msg := "ماتركس ما گدر يرد هسه — جرّب بعد شوية"
		if errors.Is(err, service.ErrAgentOff) {
			msg = "ماتركس مطفي هسه"
		}
		WriteError(w, http.StatusBadGateway, msg)
		return
	}
	steps := strings.Join(run.Steps, ",")
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	defer tx.Rollback()
	_, _ = tx.Exec(`INSERT INTO "MatrixTalkMessage" (id, "chatId", role, text) VALUES (gen_random_uuid()::text, $1, 'USER', $2)`, id, text)
	var reply chatMsg
	err = tx.Get(&reply, `INSERT INTO "MatrixTalkMessage" (id, "chatId", role, text, steps, "createdAt")
		VALUES (gen_random_uuid()::text, $1, 'ASSISTANT', $2, NULLIF($3, ''), now() + interval '1 millisecond')
		RETURNING id, role, text, steps, "createdAt"`, id, run.Answer, steps)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	title := text
	if r := []rune(title); len(r) > 40 {
		title = string(r[:40]) + "…"
	}
	_, _ = tx.Exec(`UPDATE "MatrixTalk" SET "updatedAt" = now(), title = CASE WHEN title = 'محادثة جديدة' THEN $2 ELSE title END WHERE id = $1`, id, title)
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, reply)
}
