package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ 📦 شغل المخازن — مكتب المراقب (قرار (ع) 10-10) ═══
// نفس البنود الي ماتركس يقيّم بيها أبو الكميات ويذكّره عليها.
type ProcurementWatchHandler struct{ repo *repository.MatrixScoreRepository }

func NewProcurementWatchHandler(db *sqlx.DB) *ProcurementWatchHandler {
	return &ProcurementWatchHandler{repo: repository.NewMatrixScoreRepository(db)}
}

// GET /api/monitor/procurement-watch
func (h *ProcurementWatchHandler) Get(w http.ResponseWriter, r *http.Request) {
	mats, err1 := h.repo.MaterialDecisions(false)
	tools, err2 := h.repo.ToolDecisions(false)
	shorts, err3 := h.repo.Shortages(false)
	days, err4 := h.repo.VehicleDays()
	unrated, err5 := h.repo.VehiclesUnratedToday()
	fleet, err6 := h.repo.FleetOverdue()
	for _, e := range []error{err1, err2, err3, err4, err5, err6} {
		if e != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر جلب شغل المخازن")
			return
		}
	}
	if len(days) > 14 {
		days = days[len(days)-14:]
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"since": repository.ProcurementWatchStart, "materials": mats, "tools": tools, "shortages": shorts,
		"vehicleDays": days, "unratedToday": unrated, "fleetOverdue": fleet,
	})
}
