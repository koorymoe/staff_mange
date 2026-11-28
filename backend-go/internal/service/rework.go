package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/timeutil"
)

// ═══ 🔧 الخطأ الفني والتصحيح (ل) 11-24 دازاي ═══
// «كادر الشفت الصباحي كمل حجز واستلم فلوس، وطلع غلط بشغلهم ورجع كادر المسائي عدّله — الوقت الي استغرقه
//  المسائي فلوسه تنخصم من الصباحي وتنضاف للمسائي، لأن المسائي طلعوا يشتغلون مجاني تصحيح خطأ».
// قرارات دازاي: الليدر الي صلّح يسجّل، المراقب يوافق، الساعات من الحجز، والسعر = سعر ساعة الي صلّح
// (من «تكاليف الشركة والأهداف»). الموافقة تكتب المبالغ بـReworkAdjustment وتنحسب بـ«الفلوس الي دخّلها».

const reworkMaxHours = 12

type ReworkClaim struct {
	ID           string     `db:"id" json:"id"`
	FixCode      *string    `db:"fixCode" json:"fixCode"` // (11-28) فاضي = لسه ما انربط حجز التصحيح
	OrigCode     string     `db:"origCode" json:"origCode"`
	Note         string     `db:"note" json:"note"`
	Status       string     `db:"status" json:"status"`
	Hours        *float64   `db:"hours" json:"hours"`
	Amount       *float64   `db:"amount" json:"amount"`
	By           string     `db:"by" json:"by"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
	DecidedBy    *string    `db:"decidedBy" json:"decidedBy"`
	DecisionNote *string    `db:"decisionNote" json:"decisionNote"`
	FixCrew      string     `db:"fixCrew" json:"fixCrew"`
	OrigCrew     string     `db:"origCrew" json:"origCrew"`
	DecidedAt    *time.Time `db:"decidedAt" json:"decidedAt"`
}

type Rework struct {
	db    *sqlx.DB
	costs *CompanyCosts
}

func NewRework(db *sqlx.DB) *Rework { return &Rework{db: db, costs: NewCompanyCosts(db)} }

// CanDecide المراقب (أو صلاحية المراقبة) ومدير النظام والمالك.
func (r *Rework) CanDecide(emp string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Employee" e WHERE e.id = $1 AND e.status = 'ACTIVE' AND (
		e.role::text IN ('MONITOR', 'ADMIN', 'OWNER') OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
			WHERE ep."employeeId" = e.id AND p.name = 'monitoring')))`, emp)
	return ok
}

func (r *Rework) bookingID(code string) (string, error) {
	var id string
	if err := r.db.Get(&id, `SELECT id FROM "Booking" WHERE upper(code) = upper(btrim($1))`, code); err != nil {
		return "", fmt.Errorf("الحجز %s مو موجود", strings.TrimSpace(code))
	}
	return id, nil
}

// CanReport (دازاي 11-28) أبو الجودة (ويا الشكاوى) والمراقب يسجّلون الخطأ — هم الي يكتشفوه بالاتصال أو الشكوى.
func (r *Rework) CanReport(emp string) bool {
	if r.CanDecide(emp) {
		return true
	}
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Employee" e WHERE e.id = $1 AND e.status = 'ACTIVE' AND (
		e.role::text = 'QUALITY_ENGINEER' OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
			WHERE ep."employeeId" = e.id AND p.name IN ('quality_control', 'complaints'))))`, emp)
	return ok
}

// Submit يسجّل خطأ فني على حجز أصلي. أبو الجودة/المراقب: حجز التصحيح اختياري (ينربط بعدين).
// الكادر: لازم يكون بكادر حجز التصحيح.
func (r *Rework) Submit(emp, fixCode, origCode, note string) error {
	note = strings.TrimSpace(note)
	if len([]rune(note)) < 5 {
		return errors.New("اكتب شنو چان الخطأ الفني")
	}
	orig, err := r.bookingID(origCode)
	if err != nil {
		return err
	}
	reporter := r.CanReport(emp)
	var fix *string
	if strings.TrimSpace(fixCode) != "" {
		f, err := r.bookingID(fixCode)
		if err != nil {
			return err
		}
		if f == orig {
			return errors.New("حجز التصحيح لازم غير الحجز الأصلي")
		}
		fix = &f
	}
	if !reporter {
		if fix == nil {
			return errors.New("اختار حجز التصحيح الي اشتغلتوه")
		}
		var inCrew bool
		_ = r.db.Get(&inCrew, `SELECT EXISTS (SELECT 1 FROM "BookingAssignment" WHERE "bookingId" = $1 AND "employeeId" = $2) OR booking_leader_id($1) = $2`, *fix, emp)
		if !inCrew {
			return errors.New("بس الكادر الي اشتغل حجز التصحيح يسجّله")
		}
	}
	if _, err := r.db.Exec(`INSERT INTO "ReworkClaim" ("fixBookingId", "originalBookingId", note, "createdById") VALUES ($1, $2, $3, $4)`, fix, orig, note, emp); err != nil {
		if strings.Contains(err.Error(), "ReworkClaim_fixBookingId_key") {
			return errors.New("هذا الحجز متسجّل تصحيح من قبل")
		}
		return err
	}
	return nil
}

// LinkFix (11-28) يربط حجز التصحيح بخطأ مسجّل — المراقب/أبو الجودة أو كادر حجز التصحيح.
func (r *Rework) LinkFix(emp, id, fixCode string) error {
	var c struct {
		Orig   string  `db:"originalBookingId"`
		Fix    *string `db:"fixBookingId"`
		Status string  `db:"status"`
	}
	if err := r.db.Get(&c, `SELECT "originalBookingId", "fixBookingId", status FROM "ReworkClaim" WHERE id = $1`, id); err != nil {
		return ErrDesignHidden
	}
	if c.Status != "PENDING" {
		return errors.New("انقرر عليه من قبل")
	}
	f, err := r.bookingID(fixCode)
	if err != nil {
		return err
	}
	if f == c.Orig {
		return errors.New("حجز التصحيح لازم غير الحجز الأصلي")
	}
	if !r.CanReport(emp) {
		var inCrew bool
		_ = r.db.Get(&inCrew, `SELECT EXISTS (SELECT 1 FROM "BookingAssignment" WHERE "bookingId" = $1 AND "employeeId" = $2) OR booking_leader_id($1) = $2`, f, emp)
		if !inCrew {
			return ErrDesignHidden
		}
	}
	if _, err := r.db.Exec(`UPDATE "ReworkClaim" SET "fixBookingId" = $2 WHERE id = $1`, id, f); err != nil {
		if strings.Contains(err.Error(), "ReworkClaim_fixBookingId_key") {
			return errors.New("هذا الحجز متسجّل تصحيح من قبل")
		}
		return err
	}
	return nil
}

// ReworkBooking (11-28) نتيجة البحث: معلومات الحجز تطلع لحالها بدل الكتابة باليد.
type ReworkBooking struct {
	Code        string     `db:"code" json:"code"`
	Customer    string     `db:"customer" json:"customer"`
	Phone       string     `db:"phone" json:"phone"`
	Service     string     `db:"service" json:"service"`
	CompletedAt *time.Time `db:"completedAt" json:"completedAt"`
	Status      string     `db:"status" json:"status"`
	Crew        string     `db:"crew" json:"crew"`
	Complaints  int        `db:"complaints" json:"complaints"`
}

// SearchBookings بالكود أو رقم/اسم الزبون — آخر ١٨٠ يوم، المكتمل والي بيه شكوى أول.
func (r *Rework) SearchBookings(emp, q string) ([]ReworkBooking, error) {
	q = strings.TrimSpace(q)
	out := []ReworkBooking{}
	if len([]rune(q)) < 2 {
		return out, nil
	}
	if !r.CanReport(emp) {
		var crew bool
		_ = r.db.Get(&crew, `SELECT EXISTS (SELECT 1 FROM "BookingAssignment" WHERE "employeeId" = $1)`, emp)
		if !crew {
			return nil, ErrDesignHidden
		}
	}
	digits := strings.Map(func(c rune) rune {
		if c >= '0' && c <= '9' {
			return c
		}
		return -1
	}, q)
	err := r.db.Select(&out, `SELECT b.code, COALESCE(c.name, '') AS customer, COALESCE(c.phone, '') AS phone, COALESCE(s.name, '') AS service,
		b."completedAt", b.status::text AS status,
		COALESCE((SELECT string_agg(x.name, '، ') FROM "BookingAssignment" a JOIN "Employee" x ON x.id = a."employeeId" WHERE a."bookingId" = b.id), '') AS crew,
		(SELECT count(*) FROM "Complaint" k WHERE k."bookingId" = b.id)::int AS complaints
		FROM "Booking" b LEFT JOIN "Customer" c ON c.id = b."customerId" LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b."createdAt" > now() - interval '180 days'
		  AND (upper(b.code) LIKE upper($1) || '%' OR c.name ILIKE '%' || $1 || '%' OR ($2 <> '' AND regexp_replace(COALESCE(c.phone, ''), '\D', '', 'g') LIKE '%' || $2 || '%'))
		ORDER BY (SELECT count(*) FROM "Complaint" k WHERE k."bookingId" = b.id) > 0 DESC, b.status = 'COMPLETED' DESC, COALESCE(b."completedAt", b."createdAt") DESC
		LIMIT 12`, q, digits)
	return out, err
}

// List المراقب يشوف الكل؛ غيره يشوف الي سجّلها هو أو الي تخصّه.
func (r *Rework) List(emp string) ([]ReworkClaim, error) {
	all := r.CanDecide(emp)
	rows := []ReworkClaim{}
	err := r.db.Select(&rows, `SELECT c.id, fb.code AS "fixCode", ob.code AS "origCode", c.note, c.status, c.hours::float8 AS hours, c.amount::float8 AS amount,
		e.name AS "by", c."createdAt", d.name AS "decidedBy", c."decisionNote", c."decidedAt",
		COALESCE((SELECT string_agg(x.name, '، ') FROM "BookingAssignment" a JOIN "Employee" x ON x.id = a."employeeId" WHERE a."bookingId" = c."fixBookingId"), '') AS "fixCrew",
		COALESCE((SELECT string_agg(x.name, '، ') FROM "BookingAssignment" a JOIN "Employee" x ON x.id = a."employeeId" WHERE a."bookingId" = c."originalBookingId"), '') AS "origCrew"
		FROM "ReworkClaim" c LEFT JOIN "Booking" fb ON fb.id = c."fixBookingId" JOIN "Booking" ob ON ob.id = c."originalBookingId"
		JOIN "Employee" e ON e.id = c."createdById" LEFT JOIN "Employee" d ON d.id = c."decidedById"
		WHERE $2 OR c."createdById" = $1
		   OR EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."employeeId" = $1 AND a."bookingId" IN (c."fixBookingId", c."originalBookingId"))
		ORDER BY (c.status = 'PENDING') DESC, c."createdAt" DESC LIMIT 200`, emp, all)
	return rows, err
}

// Decide موافقة المراقب: تنحسب الساعات من الحجز، والمبلغ بسعر ساعة الي صلّحوا، وينتقل.
func (r *Rework) Decide(emp, id string, approve bool, note string) error {
	if !r.CanDecide(emp) {
		return ErrDesignHidden
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var c struct {
		FixP   *string `db:"fixBookingId"`
		Fix    string  `db:"-"`
		Orig   string  `db:"originalBookingId"`
		Status string  `db:"status"`
	}
	if err := tx.Get(&c, `SELECT "fixBookingId", "originalBookingId", status FROM "ReworkClaim" WHERE id = $1 FOR UPDATE`, id); err != nil {
		return ErrDesignHidden
	}
	if c.Status != "PENDING" {
		return errors.New("انقرر عليه من قبل")
	}
	if c.FixP != nil {
		c.Fix = *c.FixP
	} else if approve {
		return errors.New("اربط حجز التصحيح أول — منه تنحسب ساعات التصليح")
	}
	if !approve {
		_, err := tx.Exec(`UPDATE "ReworkClaim" SET status = 'REJECTED', "decidedById" = $2, "decidedAt" = now(), "decisionNote" = NULLIF(btrim($3), '') WHERE id = $1`, id, emp, note)
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	var hours *float64
	_ = tx.Get(&hours, `SELECT (extract(epoch FROM ("completedAt" - COALESCE("startedAt", "arrivedAt"))) / 3600)::float8
		FROM "Booking" WHERE id = $1 AND "completedAt" IS NOT NULL AND COALESCE("startedAt", "arrivedAt") IS NOT NULL`, c.Fix)
	if hours == nil || *hours <= 0 {
		return errors.New("حجز التصحيح ما بيه وقت وصول أو بداية وإنهاء — لازم ينكمل الحجز حتى تنحسب الساعات")
	}
	h := math.Min(*hours, reworkMaxHours)
	h = math.Round(h*100) / 100
	fixCrew, origCrew := []string{}, []string{}
	if err := tx.Select(&fixCrew, `SELECT DISTINCT a."employeeId" FROM "BookingAssignment" a JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."bookingId" = $1 AND e.role::text <> 'SERVICE_MANAGER'`, c.Fix); err != nil {
		return err
	}
	if err := tx.Select(&origCrew, `SELECT DISTINCT a."employeeId" FROM "BookingAssignment" a JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."bookingId" = $1 AND e.role::text <> 'SERVICE_MANAGER' AND NOT (a."employeeId" = ANY($2))`, c.Orig, fixCrew); err != nil {
		return err
	}
	if len(fixCrew) == 0 || len(origCrew) == 0 {
		return errors.New("لازم الحجزين بيهن كادر مسجّل (وكادر الأصلي مو نفس الي صلّح)")
	}
	rep, err := r.costs.Report(timeutil.ThisPeriod(), time.Now())
	if err != nil {
		return err
	}
	hourly := map[string]float64{}
	for _, t := range rep.Targets {
		hourly[t.EmployeeID] = t.Hourly
	}
	var total float64
	missing := 0
	for _, e := range fixCrew {
		a := math.Round(h * hourly[e])
		if a <= 0 {
			missing++
			continue
		}
		total += a
		if _, err := tx.Exec(`INSERT INTO "ReworkAdjustment" ("claimId", "employeeId", amount) VALUES ($1, $2, $3)`, id, e, a); err != nil {
			return err
		}
	}
	if total <= 0 {
		return errors.New("الي صلّحوا ما إلهم سعر ساعة — سجّلهم بـ«تكاليف الشركة والأهداف» أول")
	}
	share := math.Round(total / float64(len(origCrew)))
	for _, e := range origCrew {
		if _, err := tx.Exec(`INSERT INTO "ReworkAdjustment" ("claimId", "employeeId", amount) VALUES ($1, $2, $3)`, id, e, -share); err != nil {
			return err
		}
	}
	dn := strings.TrimSpace(note)
	if missing > 0 {
		dn = strings.TrimSpace(dn + fmt.Sprintf(" (%d من الي صلّحوا بلا سعر ساعة — ما انضافلهم)", missing))
	}
	if _, err := tx.Exec(`UPDATE "ReworkClaim" SET status = 'APPROVED', hours = $2, amount = $3, "decidedById" = $4, "decidedAt" = now(), "decisionNote" = NULLIF($5, '') WHERE id = $1`,
		id, h, total, emp, dn); err != nil {
		return err
	}
	return tx.Commit()
}

// PendingCount للمراقب (رقم القائمة).
func (r *Rework) PendingCount() int {
	var n int
	_ = r.db.Get(&n, `SELECT count(*) FROM "ReworkClaim" WHERE status = 'PENDING'`)
	return n
}
