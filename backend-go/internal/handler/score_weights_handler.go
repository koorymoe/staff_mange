package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ScoreWeightsHandler ⚖️ بنود التقييم وأوزانها (دازاي 11-27) — المالك بس.
type ScoreWeightsHandler struct{ db *sqlx.DB }

func NewScoreWeightsHandler(db *sqlx.DB) *ScoreWeightsHandler { return &ScoreWeightsHandler{db: db} }

// GET /api/owner/score-weights
func (h *ScoreWeightsHandler) Get(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{"groups": service.ScoreWeightsView(h.db)})
}

// PUT /api/owner/score-weights {weights: {group: {key: weight}}}
func (h *ScoreWeightsHandler) Put(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Weights service.ScoreWeights `json:"weights"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := service.SaveScoreWeights(h.db, middleware.EmployeeIDFromContext(r), in.Weights); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"groups": service.ScoreWeightsView(h.db)})
}
