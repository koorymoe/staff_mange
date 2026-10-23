package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ 🤖 أسماء الوكلاء (دازاي 10-23) ═══
// ماتركس وايريس ثابتين. وكيل الجودة ووكيل الحجوزات اسمهم يحدده المالك —
// هذا الاسم يطلع للزبون، فلازم يعرّف نفسه بأول رسالة إنه مساعد ذكي.

type AiAgentNameHandler struct{ db *sqlx.DB }

func NewAiAgentNameHandler(db *sqlx.DB) *AiAgentNameHandler { return &AiAgentNameHandler{db: db} }

var aiAgentDefaults = map[string]string{"QUALITY": "مساعد الجودة", "BOOKING": "مساعد الحجوزات"}

// AgentName اسم الوكيل الحالي — الافتراضي إذا المالك بعده ما سمّاه.
func (h *AiAgentNameHandler) AgentName(agent string) string {
	var name string
	if err := h.db.Get(&name, `SELECT name FROM "AiAgentName" WHERE agent = $1`, agent); err == nil && name != "" {
		return name
	}
	return aiAgentDefaults[agent]
}

// GET /api/ai/agents
func (h *AiAgentNameHandler) List(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, http.StatusOK, []map[string]any{
		{"agent": "MATRIX", "name": "ماتركس", "editable": false},
		{"agent": "IRIS", "name": "ايريس", "editable": false},
		{"agent": "QUALITY", "name": h.AgentName("QUALITY"), "editable": true},
		{"agent": "BOOKING", "name": h.AgentName("BOOKING"), "editable": true},
	})
}

// PUT /api/ai/agents/{agent} — {name}
func (h *AiAgentNameHandler) Rename(w http.ResponseWriter, r *http.Request) {
	agent := r.PathValue("agent")
	if _, ok := aiAgentDefaults[agent]; !ok {
		WriteError(w, http.StatusBadRequest, "هذا الوكيل اسمه ثابت")
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	name := strings.TrimSpace(in.Name)
	if n := len([]rune(name)); n < 2 || n > 30 {
		WriteError(w, http.StatusBadRequest, "الاسم من ٢ إلى ٣٠ حرف")
		return
	}
	if _, err := h.db.Exec(`INSERT INTO "AiAgentName" (agent, name, "updatedById") VALUES ($1, $2, $3)
		ON CONFLICT (agent) DO UPDATE SET name = EXCLUDED.name, "updatedById" = EXCLUDED."updatedById", "updatedAt" = now()`,
		agent, name, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"agent": agent, "name": name})
}
