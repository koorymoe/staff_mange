package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/service"
)

// GET /api/maintenance-due — زبائن يستاهلون اتصال صيانة وقائية (فكرة ٧) — بصلاحية قراءة الزبائن.
func MaintenanceDueHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// (بلاغ #55) الكل بلا حد ١٠٠ — العدد الحقيقي، والواجهة تعرض ٥٠ ٥٠
		rows, err := service.MaintenanceDueList(db, 5000)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر")
			return
		}
		WriteJSON(w, http.StatusOK, rows)
	}
}

// POST /api/maintenance-due/{customerId}/decision {decision, note} — (بلاغ #55) قرار الجودة على الزبون
func MaintenanceDecisionHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Decision string `json:"decision"`
			Note     string `json:"note"`
		}
		if err := DecodeJSON(r, &in); err != nil {
			WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
			return
		}
		if err := service.DecideMaintenance(db, middleware.EmployeeIDFromContext(r), r.PathValue("customerId"), in.Decision, in.Note); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}
