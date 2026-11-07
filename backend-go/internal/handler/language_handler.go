package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// PUT /api/me/language — {language: "ar"|"ur"} — الموظف يغيّر لغة واجهته هو بس (بلاغ #11).
func SetMyLanguage(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Language string `json:"language"`
		}
		if err := DecodeJSON(r, &in); err != nil || (in.Language != "ar" && in.Language != "ur") {
			WriteError(w, http.StatusBadRequest, "اللغة لازم ar أو ur")
			return
		}
		me := middleware.EmployeeIDFromContext(r)
		if _, err := db.Exec(`UPDATE "Employee" SET language = $1 WHERE id = $2`, in.Language, me); err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"language": in.Language})
	}
}
