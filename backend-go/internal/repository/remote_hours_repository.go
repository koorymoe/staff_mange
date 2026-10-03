package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RemoteHoursRepository struct{ db *sqlx.DB }

func NewRemoteHoursRepository(db *sqlx.DB) *RemoteHoursRepository { return &RemoteHoursRepository{db: db} }

type RemoteEntry struct {
	ID           string    `db:"id" json:"id"`
	EmployeeID   string    `db:"employeeId" json:"employeeId"`
	EmployeeName string    `db:"employeeName" json:"employeeName"`
	WorkDate     time.Time `db:"workDate" json:"workDate"`
	Seconds      int       `db:"seconds" json:"seconds"`
	Note         *string   `db:"note" json:"note"`
	AddedByID    *string   `db:"addedById" json:"addedById"`
	AddedBy      string    `db:"addedBy" json:"addedBy"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

type RemoteSummary struct {
	EmployeeID string `db:"employeeId" json:"employeeId"`
	Name       string `db:"name" json:"name"`
	Days       int    `db:"days" json:"days"`
	Seconds    int    `db:"seconds" json:"seconds"`
}

const remoteSelect = `SELECT r.id, r."employeeId", e.name AS "employeeName", r."workDate", r.seconds, r.note,
	r."addedById", COALESCE(a.name, '') AS "addedBy", r."createdAt"
	FROM "RemoteWorkEntry" r JOIN "Employee" e ON e.id = r."employeeId" LEFT JOIN "Employee" a ON a.id = r."addedById"`

// List سجلات شهر (month = YYYY-MM)؛ employeeID فارغ = الكل.
func (r *RemoteHoursRepository) List(employeeID, month string) ([]RemoteEntry, error) {
	rows := []RemoteEntry{}
	err := r.db.Select(&rows, remoteSelect+`
		WHERE to_char(r."workDate", 'YYYY-MM') = $1 AND ($2 = '' OR r."employeeId" = $2)
		ORDER BY e.name, r."workDate", r."createdAt"`, month, employeeID)
	return rows, err
}

func (r *RemoteHoursRepository) Summary(month string) ([]RemoteSummary, error) {
	rows := []RemoteSummary{}
	err := r.db.Select(&rows, `SELECT r."employeeId", e.name, COUNT(DISTINCT r."workDate") AS days, SUM(r.seconds)::int AS seconds
		FROM "RemoteWorkEntry" r JOIN "Employee" e ON e.id = r."employeeId"
		WHERE to_char(r."workDate", 'YYYY-MM') = $1
		GROUP BY r."employeeId", e.name ORDER BY e.name`, month)
	return rows, err
}

func (r *RemoteHoursRepository) DaySeconds(employeeID, date string) int {
	var n int
	_ = r.db.Get(&n, `SELECT COALESCE(SUM(seconds), 0) FROM "RemoteWorkEntry" WHERE "employeeId" = $1 AND "workDate" = $2::date`, employeeID, date)
	return n
}

func (r *RemoteHoursRepository) EmployeeActive(id string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Employee" WHERE id = $1 AND status = 'ACTIVE')`, id)
	return ok
}

func (r *RemoteHoursRepository) Create(employeeID, date string, seconds int, note *string, byID string) (string, error) {
	id := uuid.NewString()
	_, err := r.db.Exec(`INSERT INTO "RemoteWorkEntry" (id, "employeeId", "workDate", seconds, note, "addedById") VALUES ($1,$2,$3::date,$4,$5,NULLIF($6,''))`,
		id, employeeID, date, seconds, note, byID)
	return id, err
}

func (r *RemoteHoursRepository) AddedBy(id string) (string, bool) {
	var by *string
	if err := r.db.Get(&by, `SELECT "addedById" FROM "RemoteWorkEntry" WHERE id = $1`, id); err != nil {
		return "", false
	}
	if by == nil {
		return "", true
	}
	return *by, true
}

func (r *RemoteHoursRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "RemoteWorkEntry" WHERE id = $1`, id)
	return err
}
