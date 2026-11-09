package repository

import (
	"staffmange-api/internal/errlog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RemoteHoursRepository struct{ db *sqlx.DB }

func NewRemoteHoursRepository(db *sqlx.DB) *RemoteHoursRepository {
	return &RemoteHoursRepository{db: db}
}

type RemoteEntry struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	EmployeeName string     `db:"employeeName" json:"employeeName"`
	WorkDate     time.Time  `db:"workDate" json:"workDate"`
	Seconds      int        `db:"seconds" json:"seconds"`
	Note         *string    `db:"note" json:"note"`
	AddedByID    *string    `db:"addedById" json:"addedById"`
	AddedBy      string     `db:"addedBy" json:"addedBy"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
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
	errlog.Err(r.db.Get(&n, `SELECT COALESCE(SUM(seconds), 0) FROM "RemoteWorkEntry" WHERE "employeeId" = $1 AND "workDate" = $2::date`, employeeID, date))
	return n
}

func (r *RemoteHoursRepository) EmployeeActive(id string) bool {
	var ok bool
	errlog.Err(r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Employee" WHERE id = $1 AND status = 'ACTIVE')`, id))
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

// Entry سجل واحد — للتعديل (الموظف، اليوم، المدة).
func (r *RemoteHoursRepository) Entry(id string) (*RemoteEntry, error) {
	var e RemoteEntry
	err := r.db.Get(&e, remoteSelect+` WHERE r.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Update يعدّل المدة والملاحظة (طلب (ع) 10-23: «اريد اعدلها لان بيها خطأ»).
func (r *RemoteHoursRepository) Update(id string, seconds int, note *string) error {
	_, err := r.db.Exec(`UPDATE "RemoteWorkEntry" SET seconds = $2, note = $3 WHERE id = $1`, id, seconds, note)
	return err
}

// RemoteHoursPlan المطلوب بالشهر وسعر الساعة (دازاي 11-01).
type RemoteHoursPlan struct {
	TargetHours   int     `db:"targetHours" json:"targetHours"`
	HourlyRateIQD float64 `db:"hourlyRateIQD" json:"hourlyRateIQD"`
}

// Plan خطة الموظف (nil إذا ما محددة).
func (r *RemoteHoursRepository) Plan(emp string) *RemoteHoursPlan {
	var p RemoteHoursPlan
	if err := r.db.Get(&p, `SELECT "targetHours", "hourlyRateIQD"::float8 AS "hourlyRateIQD" FROM "RemoteHoursPlan" WHERE "employeeId" = $1`, emp); err != nil {
		return nil
	}
	return &p
}

// SetPlan يحفظ الخطة. onlyIfNew: الموظف لنفسه يحددها مرة وحدة بس.
func (r *RemoteHoursRepository) SetPlan(emp, by string, p RemoteHoursPlan, onlyIfNew bool) (bool, error) {
	q := `INSERT INTO "RemoteHoursPlan" ("employeeId", "targetHours", "hourlyRateIQD", "setById") VALUES ($1, $2, $3, $4)`
	if onlyIfNew {
		q += ` ON CONFLICT ("employeeId") DO NOTHING`
	} else {
		q += ` ON CONFLICT ("employeeId") DO UPDATE SET "targetHours" = EXCLUDED."targetHours", "hourlyRateIQD" = EXCLUDED."hourlyRateIQD", "setById" = EXCLUDED."setById", "setAt" = now()`
	}
	res, err := r.db.Exec(q, emp, p.TargetHours, p.HourlyRateIQD, by)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ═══ (دازاي 11-02) سجل الأشهر السابقة — للمالك بس ═══

type RemoteMonthTotal struct {
	Month     string `db:"month" json:"month"`
	Employees int    `db:"employees" json:"employees"`
	Days      int    `db:"days" json:"days"`
	Seconds   int    `db:"seconds" json:"seconds"`
}

// Months كل شهر بيه دوام من البيت (الأحدث أول).
func (r *RemoteHoursRepository) Months() ([]RemoteMonthTotal, error) {
	rows := []RemoteMonthTotal{}
	err := r.db.Select(&rows, `SELECT to_char("workDate", 'YYYY-MM') AS month, COUNT(DISTINCT "employeeId")::int AS employees,
		COUNT(DISTINCT ("employeeId", "workDate"))::int AS days, SUM(seconds)::int AS seconds
		FROM "RemoteWorkEntry" GROUP BY 1 ORDER BY 1 DESC LIMIT 36`)
	return rows, err
}

type RemoteMonthRow struct {
	EmployeeID    string   `db:"employeeId" json:"employeeId"`
	Name          string   `db:"name" json:"name"`
	Days          int      `db:"days" json:"days"`
	Seconds       int      `db:"seconds" json:"seconds"`
	TargetHours   *int     `db:"targetHours" json:"targetHours"`
	HourlyRateIQD *float64 `db:"hourlyRateIQD" json:"hourlyRateIQD"`
}

// MonthDetail كل موظف بالشهر ويا المطلوب وسعر الساعة (الحالي).
func (r *RemoteHoursRepository) MonthDetail(month string) ([]RemoteMonthRow, error) {
	rows := []RemoteMonthRow{}
	err := r.db.Select(&rows, `SELECT r."employeeId", e.name, COUNT(DISTINCT r."workDate")::int AS days, SUM(r.seconds)::int AS seconds,
		p."targetHours", p."hourlyRateIQD"::float8 AS "hourlyRateIQD"
		FROM "RemoteWorkEntry" r JOIN "Employee" e ON e.id = r."employeeId" LEFT JOIN "RemoteHoursPlan" p ON p."employeeId" = r."employeeId"
		WHERE to_char(r."workDate", 'YYYY-MM') = $1
		GROUP BY r."employeeId", e.name, p."targetHours", p."hourlyRateIQD" ORDER BY seconds DESC`, month)
	return rows, err
}

// AgentDaySeconds مجموع ما سجّله الوكيل (Claude) لموظف بيوم (دازاي 11-06).
func (r *RemoteHoursRepository) AgentDaySeconds(employeeID, date string) int {
	var n int
	errlog.Err(r.db.Get(&n, `SELECT COALESCE(SUM(seconds), 0) FROM "RemoteWorkEntry" WHERE "employeeId" = $1 AND "workDate" = $2::date AND source = 'AGENT'`, employeeID, date))
	return n
}

// AgentKeyUsed هل انسجّل سطر بهالمفتاح قبل؟ (المفتاح محفوظ بأول الملاحظة «🤖 Claude [key]»)
func (r *RemoteHoursRepository) AgentKeyUsed(employeeID, key string) bool {
	var ok bool
	errlog.Err(r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "RemoteWorkEntry" WHERE "employeeId" = $1 AND source = 'AGENT' AND starts_with(note, $2))`,
		employeeID, "🤖 Claude ["+key+"]"))
	return ok
}
