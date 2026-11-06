package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/service"
)

// GET /api/maintenance-due — زبائن يستاهلون اتصال صيانة وقائية (فكرة ٧) — بصلاحية قراءة الزبائن.
func MaintenanceDueHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := service.MaintenanceDueList(db, 100)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر")
			return
		}
		WriteJSON(w, http.StatusOK, rows)
	}
}
