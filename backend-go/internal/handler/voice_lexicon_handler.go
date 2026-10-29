package handler

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/service"
)

// ═══ 👂 قاموس الأذن (دازاي 10-29) — المالك والمدير يشوفون ويضيفون ويشيلون ═══

type VoiceLexiconHandler struct{ db *sqlx.DB }

func NewVoiceLexiconHandler(db *sqlx.DB) *VoiceLexiconHandler { return &VoiceLexiconHandler{db: db} }

// GET /api/voice/terms?q=&source= — الكلمات + التصحيحات
func (h *VoiceLexiconHandler) List(w http.ResponseWriter, r *http.Request) {
	q := "%" + strings.TrimSpace(r.URL.Query().Get("q")) + "%"
	src := r.URL.Query().Get("source")
	terms := []struct {
		Term     string `db:"term" json:"term"`
		Meaning  string `db:"meaning" json:"meaning"`
		Category string `db:"category" json:"category"`
		Source   string `db:"source" json:"source"`
	}{}
	_ = h.db.Select(&terms, `SELECT term, meaning, category, source FROM "VoiceTerm"
		WHERE (term ILIKE $1 OR meaning ILIKE $1) AND ($2 = '' OR source = $2)
		ORDER BY "updatedAt" DESC LIMIT 500`, q, src)
	fixes := []struct {
		Wrong  string `db:"wrong" json:"wrong"`
		Right  string `db:"right" json:"right"`
		Source string `db:"source" json:"source"`
	}{}
	_ = h.db.Select(&fixes, `SELECT wrong, "right", source FROM "VoiceFix" ORDER BY "createdAt" DESC`)
	var total int
	_ = h.db.Get(&total, `SELECT count(*) FROM "VoiceTerm"`)
	WriteJSON(w, http.StatusOK, map[string]any{"terms": terms, "fixes": fixes, "total": total})
}

// POST /api/voice/terms — {term, meaning, heardAs}
func (h *VoiceLexiconHandler) Add(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Term    string `json:"term"`
		Meaning string `json:"meaning"`
		HeardAs string `json:"heardAs"`
	}
	if err := DecodeJSON(r, &in); err != nil || len([]rune(strings.TrimSpace(in.Term))) < 2 {
		WriteError(w, http.StatusBadRequest, "اكتب الكلمة")
		return
	}
	if service.DefaultVoiceLexicon == nil {
		WriteError(w, http.StatusServiceUnavailable, "الأذن مطفية")
		return
	}
	if err := service.DefaultVoiceLexicon.Learn(in.Term, in.Meaning, in.HeardAs, "OWNER"); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /api/voice/terms?term= أو ?wrong= — يشيل كلمة أو تصحيح
func (h *VoiceLexiconHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if t := r.URL.Query().Get("term"); t != "" {
		_, _ = h.db.Exec(`DELETE FROM "VoiceTerm" WHERE term = $1`, t)
	}
	if t := r.URL.Query().Get("wrong"); t != "" {
		_, _ = h.db.Exec(`DELETE FROM "VoiceFix" WHERE wrong = $1`, t)
	}
	if service.DefaultVoiceLexicon != nil {
		service.DefaultVoiceLexicon.Invalidate()
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}
