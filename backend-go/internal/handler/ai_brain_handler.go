package handler

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ 🧠 الدماغين (دازاي 10-28): السقف من الصفحة + اقتراحات الذكاء ═══
// العقد: docs/ai-brains-contract.md

type AiBrainHandler struct{ db *sqlx.DB }

func NewAiBrainHandler(db *sqlx.DB) *AiBrainHandler { return &AiBrainHandler{db: db} }

// LoadBudgets يحمّل سقوف الدماغين من القاعدة (تغلب على .env) — بتشغيل الخادم.
func (h *AiBrainHandler) LoadBudgets(defaults map[string]float64) {
	rows := []struct {
		Brain string  `db:"brain"`
		USD   float64 `db:"dailyUSD"`
	}{}
	_ = h.db.Select(&rows, `SELECT brain, "dailyUSD" FROM "AiBrainBudget"`)
	saved := map[string]float64{}
	for _, r := range rows {
		saved[r.Brain] = r.USD
	}
	for brain, def := range defaults {
		if v, ok := saved[brain]; ok {
			def = v
		}
		service.SetAIBrainBudget(brain, def)
	}
}

// PUT /api/ai/budgets/{brain} — {dailyUSD} — المالك بس
func (h *AiBrainHandler) SetBudget(w http.ResponseWriter, r *http.Request) {
	brain := r.PathValue("brain")
	if brain != service.AIBrainInternal && brain != service.AIBrainExternal {
		WriteError(w, http.StatusBadRequest, "دماغ غير معروف")
		return
	}
	var in struct {
		DailyUSD float64 `json:"dailyUSD"`
	}
	if err := DecodeJSON(r, &in); err != nil || in.DailyUSD <= 0 || in.DailyUSD > 50 {
		WriteError(w, http.StatusBadRequest, "السقف بين ٠.١ و٥٠ دولار باليوم")
		return
	}
	if _, err := h.db.Exec(`INSERT INTO "AiBrainBudget" (brain, "dailyUSD", "updatedById") VALUES ($1, $2, $3)
		ON CONFLICT (brain) DO UPDATE SET "dailyUSD" = EXCLUDED."dailyUSD", "updatedById" = EXCLUDED."updatedById", "updatedAt" = now()`,
		brain, in.DailyUSD, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	service.SetAIBrainBudget(brain, in.DailyUSD)
	spent, limit := service.AIBrainState(brain)
	WriteJSON(w, http.StatusOK, map[string]any{"brain": brain, "spentTodayUSD": spent, "dailyLimitUSD": limit})
}

type aiSuggestion struct {
	ID        string `db:"id" json:"id"`
	Agent     string `db:"agent" json:"agent"`
	Title     string `db:"title" json:"title"`
	Details   string `db:"details" json:"details"`
	NeedsCode bool   `db:"needsCode" json:"needsCode"`
	Status    string `db:"status" json:"status"`
	CreatedAt string `db:"createdAt" json:"createdAt"`
}

// GET /api/ai/suggestions?status=NEW|ACCEPTED|REJECTED|DONE&needsCode=1
func (h *AiBrainHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	q := `SELECT id, agent, title, details, "needsCode", status, to_char("createdAt" AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS "createdAt"
		FROM "AiSuggestion" WHERE ($1 = '' OR status = $1) AND (NOT $2 OR "needsCode") ORDER BY "createdAt" DESC LIMIT 200`
	rows := []aiSuggestion{}
	if err := h.db.Select(&rows, q, strings.ToUpper(r.URL.Query().Get("status")), r.URL.Query().Get("needsCode") == "1"); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/ai/suggestions/{id}/decide — {status}
func (h *AiBrainHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
	}
	if err := DecodeJSON(r, &in); err != nil || (in.Status != "ACCEPTED" && in.Status != "REJECTED" && in.Status != "DONE" && in.Status != "NEW") {
		WriteError(w, http.StatusBadRequest, "الحالة لازم ACCEPTED أو REJECTED أو DONE")
		return
	}
	res, err := h.db.Exec(`UPDATE "AiSuggestion" SET status = $2, "decidedById" = $3, "decidedAt" = now() WHERE id = $1`,
		r.PathValue("id"), in.Status, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		WriteError(w, http.StatusNotFound, "الاقتراح مو موجود")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
