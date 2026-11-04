package handler

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ لوحة الدوام — قرار (ع) 10-13 ═══
// «يطلعلي كل الموظفين منين بدا دوامهم ووين انتهى» — بأي يوم، ويا الإجازة
// المعتمدة. وأيام إجازة الموظف بالشهر حتى تنكتب بإكسله.

type AttendanceBoardHandler struct{ db *sqlx.DB }

func NewAttendanceBoardHandler(db *sqlx.DB) *AttendanceBoardHandler { return &AttendanceBoardHandler{db: db} }

type boardRow struct {
	ID        string     `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Role      string     `db:"role" json:"role"`
	Shift     *string    `db:"shift" json:"shift"`
	Start     *string    `db:"shiftStart" json:"shiftStart"`
	End       *string    `db:"shiftEnd" json:"shiftEnd"`
	FirstIn   *time.Time `db:"firstIn" json:"firstIn"`
	LastOut   *time.Time `db:"lastOut" json:"lastOut"`
	Open      bool       `db:"open" json:"open"`
	Minutes   int        `db:"minutes" json:"minutes"`
	LeaveKind *string    `db:"leaveKind" json:"leaveKind"`
}

// GET /api/attendance/board?date=YYYY-MM-DD
func (h *AttendanceBoardHandler) Board(w http.ResponseWriter, r *http.Request) {
	day := r.URL.Query().Get("date")
	if _, err := time.Parse("2006-01-02", day); err != nil {
		day = time.Now().In(baghdadLoc()).Format("2006-01-02")
	}
	rows := []boardRow{}
	if err := h.db.Select(&rows, `
		SELECT e.id, e.name, e.role::text AS role, e.shift::text AS shift, e."shiftStart", e."shiftEnd",
		       MIN(a."checkIn") AS "firstIn", MAX(a."checkOut") AS "lastOut",
		       COALESCE(BOOL_OR(a.id IS NOT NULL AND a."checkOut" IS NULL), false) AS open,
		       COALESCE(SUM(EXTRACT(EPOCH FROM (COALESCE(a."checkOut", now()) - a."checkIn")) / 60), 0)::int AS minutes,
		       (SELECT l.kind FROM "LeaveRequest" l WHERE l."employeeId" = e.id AND l.status = 'APPROVED'
		         AND $1::date BETWEEN l."startDate"::date AND l."endDate"::date LIMIT 1) AS "leaveKind"
		FROM "Employee" e
		LEFT JOIN "Attendance" a ON a."employeeId" = e.id AND baghdad_date(a."checkIn") = $1::date
		WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('OWNER')
		GROUP BY e.id ORDER BY e.name`, day); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الدوام")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"date": day, "rows": rows})
}

// LeaveDays أيام الإجازة المعتمدة للموظف بالشهر — date → نوعها.
func (h *AttendanceBoardHandler) LeaveDays(employeeID, month string) map[string]string {
	out := map[string]string{}
	rows := []struct {
		D    string `db:"d"`
		Kind string `db:"kind"`
	}{}
	_ = h.db.Select(&rows, `SELECT to_char(g::date, 'YYYY-MM-DD') AS d, l.kind
		FROM "LeaveRequest" l, generate_series(l."startDate"::date, l."endDate"::date, interval '1 day') g
		WHERE l."employeeId" = $1 AND l.status = 'APPROVED' AND to_char(g::date, 'YYYY-MM') = $2`, employeeID, month)
	for _, x := range rows {
		label := model.LeaveKindLabels[x.Kind]
		if label == "" {
			label = "إجازة"
		}
		out[x.D] = label
	}
	return out
}
