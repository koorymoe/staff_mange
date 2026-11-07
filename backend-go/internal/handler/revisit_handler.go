package handler

import (
	"log"
	"net/http"
	"strconv"

	"staffmange-api/internal/repository"
)

// GET /api/revisits?days=90 — 🔁 الرجعات (ع) 11-07: المدير والمالك والمراقب.
type RevisitHandler struct{ repo *repository.RevisitRepository }

func NewRevisitHandler(repo *repository.RevisitRepository) *RevisitHandler {
	return &RevisitHandler{repo: repo}
}

func (h *RevisitHandler) Report(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	rep, err := h.repo.Report(days)
	if err != nil {
		log.Printf("revisits: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حساب الرجعات")
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}
