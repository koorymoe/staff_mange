package handler

import (
	"log"
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type MonitorDeskHandler struct {
	desk    *repository.MonitorDeskRepository
	reviews *repository.MonitorReviewRepository
}

func NewMonitorDeskHandler(desk *repository.MonitorDeskRepository, reviews *repository.MonitorReviewRepository) *MonitorDeskHandler {
	return &MonitorDeskHandler{desk: desk, reviews: reviews}
}

// GET /api/monitor-desk/counts — عدّاد حي للطوابير الخمسة كلها.
func (h *MonitorDeskHandler) Counts(w http.ResponseWriter, r *http.Request) {
	issues, invoices, quality, crew, err := h.desk.Counts()
	if err != nil {
		log.Printf("monitor desk counts: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب عدّادات المكتب")
		return
	}
	inboxRows, err := h.reviews.Counts()
	if err != nil {
		log.Printf("monitor desk inbox counts: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب عدّادات المكتب")
		return
	}
	inbox := 0
	for _, row := range inboxRows {
		inbox += row.Count
	}
	WriteJSON(w, http.StatusOK, model.MonitorDeskCounts{
		Inbox: inbox, Issues: issues, Invoices: invoices, Quality: quality, Crew: crew,
	})
}
