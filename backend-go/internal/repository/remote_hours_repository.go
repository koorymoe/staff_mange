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
	Source       string     `db:"source" json:"source"` // MANUAL | TIMER
	ClosedAt     *time.Time `db:"closedAt" json:"closedAt"`
}

type RemoteSummary struct {
	EmployeeID string `db:"employeeId" json:"employeeId"`
	Name       string `db:"name" json:"name"`
	Days       int    `db:"days" json:"days"`
	Seconds    int    `db:"seconds" json:"seconds"`
}

const remoteSelect = `SELECT r.id, r."employeeId", e.name AS "employeeName", r."workDate", r.seconds, r.note,
	r."addedById", COALESCE(a.name, '') AS "addedBy", r."createdAt", r.source, r."closedAt"
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
	return r.CreateFrom(employeeID, date, seconds, note, byID, "MANUAL")
}

func (r *RemoteHoursRepository) CreateFrom(employeeID, date string, seconds int, note *string, byID, source string) (string, error) {
	id := uuid.NewString()
	_, err := r.db.Exec(`INSERT INTO "RemoteWorkEntry" (id, "employeeId", "workDate", seconds, note, "addedById", source) VALUES ($1,$2,$3::date,$4,$5,NULLIF($6,''),$7)`,
		id, employeeID, date, seconds, note, byID, source)
	return id, err
}

// Open سجلات الفترة المفتوحة (من آخر تصفير) — employeeID فارغ = الكل.
func (r *RemoteHoursRepository) Open(employeeID string) ([]RemoteEntry, error) {
	rows := []RemoteEntry{}
	err := r.db.Select(&rows, remoteSelect+`
		WHERE r."closedAt" IS NULL AND ($1 = '' OR r."employeeId" = $1)
		ORDER BY e.name, r."workDate", r."createdAt"`, employeeID)
	return rows, err
}

func (r *RemoteHoursRepository) OpenSummary() ([]RemoteSummary, error) {
	rows := []RemoteSummary{}
	err := r.db.Select(&rows, `SELECT r."employeeId", e.name, COUNT(DISTINCT r."workDate") AS days, SUM(r.seconds)::int AS seconds
		FROM "RemoteWorkEntry" r JOIN "Employee" e ON e.id = r."employeeId"
		WHERE r."closedAt" IS NULL GROUP BY r."employeeId", e.name ORDER BY e.name`)
	return rows, err
}

// CloseAll التصفير: يسكّر كل السجلات المفتوحة (ما يحذفها).
func (r *RemoteHoursRepository) CloseAll(byID string) (int64, error) {
	res, err := r.db.Exec(`UPDATE "RemoteWorkEntry" SET "closedAt" = now(), "closedById" = NULLIF($1,'') WHERE "closedAt" IS NULL`, byID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── العدّاد الحي ──

type RemoteTimer struct {
	StartedAt   *time.Time `db:"startedAt" json:"startedAt"`
	Accumulated int        `db:"accumulated" json:"accumulated"`
}

func (r *RemoteHoursRepository) Timer(employeeID string) RemoteTimer {
	var t RemoteTimer
	// ⚠️ بلا صف، sqlx يترك StartedAt مؤشر لوقت صفري (مو nil) — فالعدّاد
	// ينحسب «شغّال من سنة ١» ويطلع مليارات الثواني. أي خطأ = عدّاد فاضي.
	if err := r.db.Get(&t, `SELECT "startedAt", accumulated FROM "RemoteTimer" WHERE "employeeId" = $1`, employeeID); err != nil {
		return RemoteTimer{}
	}
	if t.StartedAt != nil && t.StartedAt.IsZero() {
		t.StartedAt = nil
	}
	return t
}

func (r *RemoteHoursRepository) SaveTimer(employeeID string, startedAt *time.Time, acc int) error {
	_, err := r.db.Exec(`INSERT INTO "RemoteTimer" ("employeeId", "startedAt", accumulated, "updatedAt") VALUES ($1,$2,$3,now())
		ON CONFLICT ("employeeId") DO UPDATE SET "startedAt" = $2, accumulated = $3, "updatedAt" = now()`, employeeID, startedAt, acc)
	return err
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
