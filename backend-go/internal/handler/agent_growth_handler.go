package handler

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/errlog"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/monitor"
	"staffmange-api/internal/service"
)

// AgentGrowthHandler (دازاي 11-06): 👍/👎 على أجوبة الوكلاء، وشاشة «صحة الوكلاء».
type AgentGrowthHandler struct{ db *sqlx.DB }

func NewAgentGrowthHandler(db *sqlx.DB) *AgentGrowthHandler { return &AgentGrowthHandler{db: db} }

type feedbackIn struct {
	Good bool   `json:"good"`
	Note string `json:"note"`
}

func (h *AgentGrowthHandler) saveFeedback(w http.ResponseWriter, r *http.Request, agent, msgID string, owned bool) {
	if !owned {
		WriteError(w, http.StatusNotFound, "الرد غير موجود")
		return
	}
	var in feedbackIn
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	note := monitor.Scrub(strings.TrimSpace(in.Note)) // بلا أسماء وأرقام
	if r := []rune(note); len(r) > 500 {
		note = string(r[:500])
	}
	emp := middleware.EmployeeIDFromContext(r)
	if _, err := h.db.Exec(`INSERT INTO "AgentFeedback" (agent, "messageId", "employeeId", good, note) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ("messageId", "employeeId") DO UPDATE SET good = EXCLUDED.good, note = EXCLUDED.note, "createdAt" = now()`,
		agent, msgID, emp, in.Good, note); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/matrix/chats/{id}/messages/{msgId}/feedback {good, note} — صاحب المحادثة بس.
func (h *AgentGrowthHandler) MatrixFeedback(w http.ResponseWriter, r *http.Request) {
	var ok bool
	errlog.Err(h.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "MatrixTalkMessage" m JOIN "MatrixTalk" t ON t.id = m."chatId"
		WHERE m.id = $1 AND m."chatId" = $2 AND m.role = 'ASSISTANT' AND t."employeeId" = $3)`,
		r.PathValue("msgId"), r.PathValue("id"), middleware.EmployeeIDFromContext(r)))
	h.saveFeedback(w, r, "MATRIX", r.PathValue("msgId"), ok)
}

// POST /api/secretary/messages/{msgId}/feedback {good, note} — صاحب الرسالة بس.
func (h *AgentGrowthHandler) SecretaryFeedback(w http.ResponseWriter, r *http.Request) {
	var ok bool
	errlog.Err(h.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "SecretaryMessage" WHERE id = $1 AND role = 'assistant' AND "employeeId" = $2)`,
		r.PathValue("msgId"), middleware.EmployeeIDFromContext(r)))
	h.saveFeedback(w, r, "IRIS", r.PathValue("msgId"), ok)
}

// GET /api/ai/health — صحة الوكلاء للمالك: الأدوات والفجوات والتقييم (آخر ٧ أيام).
func (h *AgentGrowthHandler) Health(w http.ResponseWriter, r *http.Request) {
	out := map[string]any{}
	for _, agent := range []string{"MATRIX", "IRIS"} {
		stats, _ := service.AgentToolStats(h.db, agent, 7)
		gaps, _ := service.AgentGaps(h.db, agent, 14)
		var good, bad, proposed int
		errlog.Err(h.db.Get(&good, `SELECT count(*) FROM "AgentFeedback" WHERE agent = $1 AND good AND "createdAt" > now() - interval '7 days'`, agent))
		errlog.Err(h.db.Get(&bad, `SELECT count(*) FROM "AgentFeedback" WHERE agent = $1 AND NOT good AND "createdAt" > now() - interval '7 days'`, agent))
		errlog.Err(h.db.Get(&proposed, `SELECT count(*) FROM "MatrixLesson" WHERE agent = $1 AND status = 'PROPOSED'`, agent))
		notes := []string{}
		errlog.Err(h.db.Select(&notes, `SELECT note FROM "AgentFeedback" WHERE agent = $1 AND NOT good AND note <> '' ORDER BY "createdAt" DESC LIMIT 5`, agent))
		out[agent] = map[string]any{"tools": stats, "gaps": gaps, "good": good, "bad": bad, "badNotes": notes, "proposedLessons": proposed}
	}
	WriteJSON(w, http.StatusOK, out)
}
