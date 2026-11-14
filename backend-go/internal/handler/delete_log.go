package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/errlog"
	"staffmange-api/internal/middleware"
)

// DeleteLogged (دازاي 11-15 «لازم كلشي يتسجل») يحفظ نسخة كاملة من السطر بـ"DeleteLog" قبل ما
// المعالج يحذفه: منو، ومتى، ومنين. إذا الحذف ما صار (رفض، خطأ، السطر بعده موجود) تنشال النسخة.
// table ثابت من الكود (مو من الطلب)، وparam اسم متغيّر المسار الي بيه المعرّف.
func DeleteLogged(db *sqlx.DB, table, param string) func(http.Handler) http.Handler {
	q := `INSERT INTO "DeleteLog"(id, "tableName", "rowId", "actorId", route, data)
		SELECT gen_random_uuid()::text, $1, t.id, NULLIF($3, ''), $4, to_jsonb(t) FROM "` + table + `" t WHERE t.id = $2 RETURNING id`
	exists := `SELECT EXISTS (SELECT 1 FROM "` + table + `" WHERE id = $1)`
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue(param)
			var logID string
			if id != "" {
				_ = db.Get(&logID, q, table, id, middleware.EmployeeIDFromContext(r), r.Method+" "+r.URL.Path)
			}
			next.ServeHTTP(w, r)
			if logID == "" {
				return
			}
			var still bool
			if err := db.Get(&still, exists, id); err == nil && still {
				errlog.Exec(db.Exec(`DELETE FROM "DeleteLog" WHERE id = $1`, logID))
			}
		})
	}
}
