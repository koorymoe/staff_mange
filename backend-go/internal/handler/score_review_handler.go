package handler

import (
	"errors"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ تأكيد السلبي بالتقييم (دازاي 11-18) — service/score_review.go ═══

type ScoreReviewHandler struct{ svc *service.ScoreReviewService }

func NewScoreReviewHandler(svc *service.ScoreReviewService) *ScoreReviewHandler {
	return &ScoreReviewHandler{svc: svc}
}

func reviewErr(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, service.ErrReviewHidden) {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return true
	}
	WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
	return true
}

// GET /api/score-reviews — السلبي الي ينتظر تأكيدي (الخدمة ترجع بس إدارتي)
func (h *ScoreReviewHandler) Queue(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Queue(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

// POST /api/score-reviews/ratings/{id} {confirm}
func (h *ScoreReviewHandler) DecideRating(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Confirm bool `json:"confirm"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if reviewErr(w, h.svc.DecideRating(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"), in.Confirm)) {
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/score-reviews/issues/{id} {faultEmployeeId} — فارغ = ماكو غلطان
func (h *ScoreReviewHandler) DecideIssue(w http.ResponseWriter, r *http.Request) {
	var in struct {
		FaultEmployeeID string `json:"faultEmployeeId"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if reviewErr(w, h.svc.DecideIssue(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"), in.FaultEmployeeID)) {
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/inventory/after-shortages — نواقص العدّة بعد الحجوزات
func (h *ScoreReviewHandler) AfterShortages(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.AfterShortages()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

// POST /api/inventory/after-shortages/justify {bookingId, employeeId}
func (h *ScoreReviewHandler) Justify(w http.ResponseWriter, r *http.Request) {
	var in struct {
		BookingID  string `json:"bookingId"`
		EmployeeID string `json:"employeeId"`
	}
	if err := DecodeJSON(r, &in); err != nil || in.BookingID == "" || in.EmployeeID == "" {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if reviewErr(w, h.svc.Justify(middleware.EmployeeIDFromContext(r), in.BookingID, in.EmployeeID)) {
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/me/station-excuses {bookingId, stationKey, reason} — الموظف يعتذر عن محطة خسر بيها
func (h *ScoreReviewHandler) SubmitExcuse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		BookingID  string `json:"bookingId"`
		StationKey string `json:"stationKey"`
		Reason     string `json:"reason"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.svc.SubmitExcuse(middleware.EmployeeIDFromContext(r), in.BookingID, in.StationKey, in.Reason); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/score-reviews/excuses/{id} {approve}
func (h *ScoreReviewHandler) DecideExcuse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Approve bool `json:"approve"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if reviewErr(w, h.svc.DecideExcuse(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), r.PathValue("id"), in.Approve)) {
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
