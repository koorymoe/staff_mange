package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/service"
)

// GET /api/sla/board — الحجوزات الي قرب أو فات موعد التزامها (٢٤ ساعة الجاية) (دازاي 11-06).
func SlaBoardHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := service.SlaBoard(db, 24)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر حساب مواعيد الالتزام")
			return
		}
		WriteJSON(w, http.StatusOK, rows)
	}
}
