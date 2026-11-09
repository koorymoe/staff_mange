package handler

import (
	"log"
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

// GET /api/drill/{key} — «الرقم = القائمة» (ع 10-16): البطاقة تفتح نفس
// الحجوزات الي عدّتها، ويا شرح عراقي شنو هذا الرقم.
func (h *TodayHandler) Drill(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if repository.DrillByKey(key) == nil {
		WriteError(w, http.StatusNotFound, "هذا الرقم ما إله قائمة")
		return
	}
	res, err := h.repo.Drill(key)
	if err != nil {
		log.Printf("drill %s: %v", key, err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب القائمة")
		return
	}
	WriteJSON(w, http.StatusOK, res)
}
