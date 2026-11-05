package handler

import (
	"encoding/json"
	"net/http"
	"staffmange-api/internal/errlog"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/integrity"
	"staffmange-api/internal/repository"
)

// 🩺 سلامة البيانات — للمالك بس. قراءة بس.
type IntegrityHandler struct {
	db    *sqlx.DB
	notif *repository.NotificationRepository
}

func NewIntegrityHandler(db *sqlx.DB, n *repository.NotificationRepository) *IntegrityHandler {
	return &IntegrityHandler{db: db, notif: n}
}

type integrityHistory struct {
	StartedAt string `db:"startedAt" json:"startedAt"`
	Problems  int64  `db:"problems" json:"problems"`
	Critical  int64  `db:"critical" json:"critical"`
}

// GET /api/integrity — آخر دورة + الدورة الي قبلها (للمقارنة) + آخر ٣٠ للمنحنى
func (h *IntegrityHandler) Latest(w http.ResponseWriter, r *http.Request) {
	runs := []integrity.Run{}
	errlog.Err(h.db.Select(&runs, `SELECT * FROM "IntegrityRun" ORDER BY "startedAt" DESC LIMIT 2`))
	hist := []integrityHistory{}
	errlog.Err(h.db.Select(&hist, `SELECT "startedAt"::text AS "startedAt", problems, critical FROM "IntegrityRun" ORDER BY "startedAt" DESC LIMIT 30`))
	out := map[string]any{"run": nil, "results": []integrity.Result{}, "previous": map[string]int64{}, "history": hist}
	if len(runs) > 0 {
		var res []integrity.Result
		_ = json.Unmarshal(runs[0].Results, &res)
		out["run"], out["results"] = runs[0], res
	}
	if len(runs) > 1 {
		var pr []integrity.Result
		_ = json.Unmarshal(runs[1].Results, &pr)
		prev := map[string]int64{}
		for _, x := range pr {
			prev[x.Key] = x.Count
		}
		out["previous"] = prev
	}
	WriteJSON(w, http.StatusOK, out)
}

// POST /api/integrity/run — فحص هسه
func (h *IntegrityHandler) RunNow(w http.ResponseWriter, r *http.Request) {
	run, res, err := integrity.RunAndStore(h.db, "manual", h.notif)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الفحص")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"run": run, "results": res})
}

// GET /api/integrity/{key}/sample — أمثلة (٢٠ صف)
func (h *IntegrityHandler) Sample(w http.ResponseWriter, r *http.Request) {
	rows, ok, err := integrity.Sample(h.db, r.PathValue("key"))
	if !ok {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر جلب الأمثلة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
