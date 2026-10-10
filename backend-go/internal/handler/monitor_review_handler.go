package handler

import (
	"net/http"
	"strconv"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// MonitorReviewHandler صندوق المراقب.
type MonitorReviewHandler struct {
	svc *service.MonitorReviewService
}

func NewMonitorReviewHandler(svc *service.MonitorReviewService) *MonitorReviewHandler {
	return &MonitorReviewHandler{svc: svc}
}

// GET /api/monitor-reviews?stage=&status=&ownerRole=&limit=
//
// بدون status نرجّع الكل — المراقب يحتاج يشوف تاريخ قراراته مو
// المعلّق بس، خاصة لو انفتح خلاف على ملاحظة قديمة.
func (h *MonitorReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	rows, err := h.svc.List(q.Get("stage"), q.Get("status"), q.Get("ownerRole"), limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/monitor-reviews/counts — عدّاد المعلّق لكل محطة.
func (h *MonitorReviewHandler) Counts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Counts()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/monitor-reviews/{id}/decide — «سليم» أو «عندي ملاحظة».
func (h *MonitorReviewHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var req model.DecideMonitorReviewRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	monitorID, _ := r.Context().Value(middleware.ContextEmployeeID).(string)
	row, err := h.svc.Decide(r.PathValue("id"), monitorID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, row)
}
