package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ تقييم الموظفين: ماتركس ٦٠٪ + البشر ٤٠٪ (قرار (ع) 10-05) ═══
// المدير/المالك: الكل. المراقب: الكل عدا نفسه. الموظف: تقييمه هو بس.
type StaffScoreHandler struct{ svc *service.MatrixScoreService }

func NewStaffScoreHandler(svc *service.MatrixScoreService) *StaffScoreHandler {
	return &StaffScoreHandler{svc: svc}
}

// GET /api/staff-score?month=
func (h *StaffScoreHandler) Board(w http.ResponseWriter, r *http.Request) {
	b, err := h.svc.Board(r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب التقييم")
		return
	}
	if !isAdminRole(r) {
		me := middleware.EmployeeIDFromContext(r)
		kept := b.Staff[:0]
		for _, s := range b.Staff {
			if s.ID != me {
				kept = append(kept, s)
			}
		}
		b.Staff = kept
	}
	WriteJSON(w, http.StatusOK, b)
}

// GET /api/staff-score/{id}
func (h *StaffScoreHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isAdminRole(r) && id == middleware.EmployeeIDFromContext(r) {
		WriteError(w, http.StatusForbidden, "تقييمك تشوفه من «تقييمي»")
		return
	}
	d, err := h.svc.Detail(id, r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, d)
}

// GET /api/staff-score/me
func (h *StaffScoreHandler) Mine(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.Detail(middleware.EmployeeIDFromContext(r), r.URL.Query().Get("month"))
	if err != nil {
		WriteJSON(w, http.StatusOK, nil)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"score": d, "on": h.svc.On()})
}

// POST /api/matrix-score/{id}/cancel — المدير يلغي نقطة غلط.
func (h *StaffScoreHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Note string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if len([]rune(in.Note)) < 3 {
		WriteError(w, http.StatusBadRequest, "اكتب ليش تلغيها")
		return
	}
	if err := h.svc.Repo().Cancel(r.PathValue("id"), middleware.EmployeeIDFromContext(r), in.Note); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/staff-score/run — احسب هسه (المدير).
func (h *StaffScoreHandler) Run(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.Run(time.Now())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"added": n})
}

type rateIn struct {
	BookingID string  `json:"bookingId"`
	RateeID   string  `json:"rateeId"`
	Score     int     `json:"score"`
	Note      *string `json:"note"`
}

// POST /api/staff-ratings/{stage} — COORD_LEADER | AUDIT | QUALITY_CALL | MONITOR_PERIODIC
// (كل محطة بحارسها بالمسار).
func (h *StaffScoreHandler) rate(stage string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in rateIn
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
			return
		}
		me := middleware.EmployeeIDFromContext(r)
		var err error
		if stage == "MONITOR_PERIODIC" {
			err = h.svc.RatePeriodic(me, in.RateeID, in.Score, in.Note)
		} else {
			err = h.svc.RateOnBooking(stage, me, in.BookingID, in.RateeID, in.Score, in.Note)
		}
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func (h *StaffScoreHandler) RateCoord() http.HandlerFunc    { return h.rate("COORD_LEADER") }
func (h *StaffScoreHandler) RateAudit() http.HandlerFunc    { return h.rate("AUDIT") }
func (h *StaffScoreHandler) RateQuality() http.HandlerFunc  { return h.rate("QUALITY_CALL") }
func (h *StaffScoreHandler) RatePeriodic() http.HandlerFunc { return h.rate("MONITOR_PERIODIC") }

// GET /api/staff-ratings/booking/{id}?stage=
func (h *StaffScoreHandler) BookingState(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.BookingRatings(r.URL.Query().Get("stage"), middleware.EmployeeIDFromContext(r), r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, st)
}

// GET /api/staff-ratings/coord-pending — حجوزات رجعت وتنتظر تقييم الليدر.
func (h *StaffScoreHandler) CoordPending(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Repo().PendingCoordLeader(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/staff-ratings/periodic
func (h *StaffScoreHandler) PeriodicState(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.Periodic(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, st)
}

// GET /api/staff-ratings/audit-pending — حجوزات تنتظر تقييم المراقب.
func (h *StaffScoreHandler) AuditPending(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Repo().PendingAudit(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
