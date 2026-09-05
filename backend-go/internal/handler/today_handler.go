package handler

import (
	"net/http"

	"staffmange-api/internal/repository"
)

// TodayHandler لوحة اليوم — «شنو صاير اليوم» و«شغلي اليوم».
type TodayHandler struct {
	repo *repository.TodayRepository
}

func NewTodayHandler(repo *repository.TodayRepository) *TodayHandler {
	return &TodayHandler{repo: repo}
}

// GET /api/dashboard/today
func (h *TodayHandler) Board(w http.ResponseWriter, r *http.Request) {
	board, err := h.repo.Board()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تجهيز لوحة اليوم")
		return
	}
	WriteJSON(w, http.StatusOK, board)
}
