package service

import (
	"errors"
	"fmt"
	"staffmange-api/internal/errlog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/timeutil"
)

// ═══ 📒 واجب التقييم اليومي (دازاي 11-02 — تعديلات ٩.٢) ═══
// «ماكو موظف بدون تقييم». كل مقيّم إله واجب يومي:
//   LEADER_CREW   ليدر الحجز ← فنيين الحجز (CrewRating)
//   COORD_LEADER  مثبّت/مسجّل الحجز ← الليدر
//   OPS_DAILY     مدير/نائب إدارة العمليات + إداري الكوادر (بالهيكلين) ← كل ليدر وفني اشتغل
//   MONITOR_DAILY المراقب ← كل موظف مكتب كتب تقرير إنجاز
// OPS وMONITOR: يكفي واحد من المقيّمين يقيّم الشخص؛ إذا محد قيّم، الواجب على كلهم.
// الي فاته: باچر تنقفل عليه الشاشات لحد ما يكتب السبب ويعوّض، وينكتب بسجل
// الإهمال، وينزل تقييمه (RATING_DUTY). بلا خصم فلوس ولا نقاط انضباط.

type DutyItem struct {
	Stage     string `db:"stage" json:"stage"`
	BookingID string `db:"bookingId" json:"bookingId"`
	Code      string `db:"code" json:"code"`
	RateeID   string `db:"rateeId" json:"rateeId"`
	RateeName string `db:"rateeName" json:"rateeName"`
	Hint      string `db:"hint" json:"hint"` // سطر من تقرير الإنجاز للمراقب
}

var DutyStageLabels = map[string]string{
	"LEADER_CREW": "تقييم كادرك بعد الحجز", "COORD_LEADER": "تقييم الليدر بعد الحجز",
	"OPS_DAILY": "تقييم الكوادر اليومي (إدارة العمليات)", "MONITOR_DAILY": "تقييم موظفي المكاتب اليومي (المراقبة)",
}

type RatingDuty struct{ db *sqlx.DB }

func NewRatingDuty(db *sqlx.DB) *RatingDuty { return &RatingDuty{db: db} }

// اليوم بتوقيت بغداد لعمود TIMESTAMP المخزون UTC
const bkDay = `(b."completedAt" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad')::date`

func (d *RatingDuty) since() string {
	var s string
	errlog.Err(d.db.Get(&s, `SELECT value FROM "OrgGovernorState" WHERE key = 'rating_duty_since'`))
	return s
}

// opsRaterSQL منو مقيّم إدارة العمليات: إداري الكوادر (دور أو منصب) أو مدير/نائب العمليات.
var opsRaterSQL = `(e.role::text = 'HR_COORDINATOR' OR e.id IN ` + model.OrgRoleEmployeesSQL(`ARRAY['HR_COORDINATOR']`) + `
	OR EXISTS (SELECT 1 FROM "OrgAssignment" oa WHERE oa."employeeId" = e.id AND oa."endedAt" IS NULL
	           AND oa."divisionId" = 'ops' AND oa."unitId" IS NULL AND oa.position IN ('DIRECTOR', 'DEPUTY')))`

var monitorRaterSQL = `(e.role::text = 'MONITOR' OR e.id IN ` + model.OrgRoleEmployeesSQL(`ARRAY['MONITOR']`) + `
	OR EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = e.id AND p.name = 'monitoring'))`

func (d *RatingDuty) is(sqlCond, emp string) bool {
	var ok bool
	errlog.Err(d.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Employee" e WHERE e.id = $1 AND e.status = 'ACTIVE' AND `+sqlCond+`)`, emp))
	return ok
}

// Owed الي على المقيّم ليوم معيّن (YYYY-MM-DD) وبعده ما انقضى.
func (d *RatingDuty) Owed(rater, day string) ([]DutyItem, error) {
	if s := d.since(); s != "" && day < s {
		return []DutyItem{}, nil
	}
	var exempt bool
	errlog.Err(d.db.Get(&exempt, `SELECT EXISTS (SELECT 1 FROM "Employee" e WHERE e.id = $1 AND e.role::text IN ('ADMIN','OWNER'))
		OR EXISTS (SELECT 1 FROM "LeaveRequest" l WHERE l."employeeId" = $1 AND l.status = 'APPROVED' AND $2::date BETWEEN l."startDate" AND l."endDate")`, rater, day))
	if exempt {
		return []DutyItem{}, nil
	}
	out := []DutyItem{}
	q := func(sql string, args ...any) error {
		rows := []DutyItem{}
		if err := d.db.Select(&rows, sql, args...); err != nil {
			return err
		}
		out = append(out, rows...)
		return nil
	}
	// ١. الليدر ← فنيينه (نفس منطق pendingRatingSQL، بس لذاك اليوم)
	if err := q(`SELECT 'LEADER_CREW' AS stage, b.id AS "bookingId", b.code, t.id AS "rateeId", t.name AS "rateeName", '' AS hint
		FROM "Booking" b JOIN "BookingAssignment" a ON a."bookingId" = b.id
		JOIN "Employee" t ON t.id = a."employeeId" AND NOT t."isLeader" AND t.id <> $1
		WHERE b.status::text IN ('COMPLETED', 'PARTIAL') AND b."completedAt" IS NOT NULL AND `+bkDay+` = $2::date
		  AND upper(b.code) NOT LIKE 'OLD%'
		  AND ((b."projectSupervisorId" = $1 AND EXISTS (SELECT 1 FROM "Employee" se WHERE se.id = $1 AND se."isLeader"))
		       OR EXISTS (SELECT 1 FROM "BookingAssignment" la JOIN "Employee" le ON le.id = la."employeeId" WHERE la."bookingId" = b.id AND la."employeeId" = $1 AND le."isLeader")
		       OR EXISTS (SELECT 1 FROM "Mission" m WHERE m."bookingId" = b.id AND m."leaderId" = $1))
		  AND NOT EXISTS (SELECT 1 FROM "CrewRating" r WHERE r."bookingId" = b.id AND r."technicianId" = t.id)`, rater, day); err != nil {
		return nil, err
	}
	// ٢. المثبّت/المسجّل ← الليدر (نفس منطق PendingCoordLeader)
	if err := q(`SELECT 'COORD_LEADER' AS stage, b.id AS "bookingId", b.code, l.id AS "rateeId", l.name AS "rateeName", '' AS hint
		FROM "Booking" b JOIN "Employee" l ON l.id = booking_leader_id(b.id)
		WHERE b.status::text IN ('COMPLETED', 'PARTIAL') AND b."completedAt" IS NOT NULL AND `+bkDay+` = $2::date
		  AND l.role::text NOT IN ('ADMIN', 'OWNER') AND upper(b.code) NOT LIKE 'OLD%'
		  AND (b."confirmedByEmployeeId" = $1 OR b."createdById" = $1) AND l.id <> $1
		  AND NOT EXISTS (SELECT 1 FROM "StaffRating" s WHERE s."raterId" = $1 AND s."bookingId" = b.id AND s.stage = 'COORD_LEADER')`, rater, day); err != nil {
		return nil, err
	}
	// ٣. إدارة العمليات ← كل ليدر وفني اشتغل ذاك اليوم (يكفي واحد من المقيّمين)
	if d.is(opsRaterSQL, rater) {
		if err := q(`SELECT 'OPS_DAILY' AS stage, '' AS "bookingId", '' AS code, p.id AS "rateeId", p.name AS "rateeName",
				string_agg(DISTINCT p.code, '، ') AS hint
			FROM (SELECT e.id, e.name, b.code FROM "Booking" b
			      JOIN LATERAL (SELECT a."employeeId" AS id FROM "BookingAssignment" a WHERE a."bookingId" = b.id
			                    UNION SELECT booking_leader_id(b.id)) w ON w.id IS NOT NULL
			      JOIN "Employee" e ON e.id = w.id
			      WHERE b.status::text IN ('COMPLETED', 'PARTIAL') AND b."completedAt" IS NOT NULL AND `+bkDay+` = $2::date
			        AND upper(b.code) NOT LIKE 'OLD%' AND e.role::text NOT IN ('ADMIN', 'OWNER') AND e.id <> $1) p
			WHERE NOT EXISTS (SELECT 1 FROM "StaffRating" s WHERE s."rateeId" = p.id AND s.stage = 'OPS_DAILY' AND s.period = $2::text)
			GROUP BY p.id, p.name ORDER BY p.name`, rater, day); err != nil {
			return nil, err
		}
	}
	// ٤. المراقبة ← موظفي المكاتب الي كتبوا تقرير ذاك اليوم (يكفي واحد)
	if d.is(monitorRaterSQL, rater) {
		if err := q(`SELECT 'MONITOR_DAILY' AS stage, '' AS "bookingId", '' AS code, e.id AS "rateeId", e.name AS "rateeName", left(r.body, 160) AS hint
			FROM "DailyWorkReport" r JOIN "Employee" e ON e.id = r."employeeId"
			WHERE r.day = $2::date AND e.id <> $1 AND NOT COALESCE(e."isLeader", false)
			  AND `+model.EffectiveRoleSQL(`e.id`, `e.role`)+` NOT IN ('TECHNICIAN', 'MONITOR', 'ADMIN', 'OWNER')
			  AND NOT EXISTS (SELECT 1 FROM "StaffRating" s WHERE s."rateeId" = e.id AND s.stage = 'MONITOR_DAILY' AND s.period = $2::text)
			ORDER BY e.name`, rater, day); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func companyDay(offset int) string {
	n := time.Now().In(timeutil.CompanyLocation())
	return time.Date(n.Year(), n.Month(), n.Day()+offset, 0, 0, 0, 0, timeutil.CompanyLocation()).Format("2006-01-02")
}

// Pending أقدم يوم (آخر ٣ أيام قبل اليوم) بيه واجب ما انقضى.
func (d *RatingDuty) Pending(rater string) (string, []DutyItem, error) {
	for off := -3; off <= -1; off++ {
		day := companyDay(off)
		items, err := d.Owed(rater, day)
		if err != nil {
			return "", nil, err
		}
		if len(items) > 0 {
			return day, items, nil
		}
	}
	return "", nil, nil
}

type DutyRating struct {
	Stage     string `json:"stage"`
	BookingID string `json:"bookingId"`
	RateeID   string `json:"rateeId"`
	Score     int    `json:"score"`
	Note      string `json:"note"`
}

// Compensate يحفظ التقييمات (بنفس اليوم: عادي؛ ليوم فات: لازم الكل + السبب + سجل إهمال)
// يحفظ التقييمات الفايتة ويا السبب، ويكتب/يحدّث سجل الإهمال — معاملة وحدة.
func (d *RatingDuty) Compensate(rater, day, reason string, ratings []DutyRating) error {
	reason = strings.TrimSpace(reason)
	late := day < companyDay(0)
	if late && len([]rune(reason)) < 10 {
		return errors.New("اكتب السبب بوضوح: ليش ما قيّمت بوقتها؟")
	}
	owed, err := d.Owed(rater, day)
	if err != nil {
		return err
	}
	if len(owed) == 0 {
		return errors.New("ماكو تقييمات فايتة لهذا اليوم")
	}
	key := func(s, b, r string) string { return s + "|" + b + "|" + r }
	want := map[string]DutyItem{}
	for _, it := range owed {
		want[key(it.Stage, it.BookingID, it.RateeID)] = it
	}
	got := map[string]DutyRating{}
	for _, r := range ratings {
		k := key(r.Stage, r.BookingID, r.RateeID)
		if _, ok := want[k]; !ok {
			return errors.New("تقييم مو من ضمن الي فاتك")
		}
		if r.Score < 1 || r.Score > 5 {
			return errors.New("كل تقييم من ١ لـ ٥")
		}
		got[k] = r
	}
	if late && len(got) != len(want) {
		return fmt.Errorf("لازم تعوّض الكل: باقي %d", len(want)-len(got))
	}
	if len(got) == 0 {
		return errors.New("ماكو تقييم")
	}
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	names := []string{}
	for k, r := range got {
		it := want[k]
		names = append(names, it.RateeName)
		note := strings.TrimSpace(r.Note)
		switch r.Stage {
		case "LEADER_CREW":
			_, err = tx.Exec(`INSERT INTO "CrewRating" (id, "bookingId", "leaderId", "technicianId", score, note)
				VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')) ON CONFLICT ("bookingId", "technicianId") DO NOTHING`,
				uuid.NewString(), r.BookingID, rater, r.RateeID, r.Score, note)
		case "COORD_LEADER":
			_, err = tx.Exec(`INSERT INTO "StaffRating" ("raterId", "rateeId", stage, "bookingId", period, score, note)
				VALUES ($1, $2, 'COORD_LEADER', $3, '', $4, NULLIF($5, '')) ON CONFLICT DO NOTHING`, rater, r.RateeID, r.BookingID, r.Score, note)
		default: // OPS_DAILY | MONITOR_DAILY
			_, err = tx.Exec(`INSERT INTO "StaffRating" ("raterId", "rateeId", stage, "bookingId", period, score, note)
				VALUES ($1, $2, $3, '', $4, $5, NULLIF($6, '')) ON CONFLICT DO NOTHING`, rater, r.RateeID, r.Stage, day, r.Score, note)
		}
		if err != nil {
			return err
		}
	}
	if !late { // بنفس اليوم: تقييم عادي بلا سجل إهمال
		return tx.Commit()
	}
	detail := fmt.Sprintf("فاته %d تقييم: %s", len(got), strings.Join(namesUpTo(names, 8), "، "))
	if _, err := tx.Exec(`INSERT INTO "EmployeeNegligence" ("employeeId", day, kind, detail, reason, "compensatedAt")
		VALUES ($1, $2::date, 'MISSED_RATINGS', $3, $4, now())
		ON CONFLICT ("employeeId", day, kind) DO UPDATE SET reason = EXCLUDED.reason, "compensatedAt" = now(),
			detail = CASE WHEN "EmployeeNegligence".detail = '' THEN EXCLUDED.detail ELSE "EmployeeNegligence".detail END`,
		rater, day, detail, reason); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	d.settlePeers(day)
	return nil
}

// settlePeers: OPS/MONITOR «يكفي واحد» — إذا زميل قيّم بدالهم، سجلاتهم المفتوحة
// لنفس اليوم تنغلق (عوّض متأخر) لأن ما بقى شي يعوّضونه.
func (d *RatingDuty) settlePeers(day string) {
	open := []string{}
	errlog.Err(d.db.Select(&open, `SELECT "employeeId" FROM "EmployeeNegligence" WHERE day = $1::date AND kind = 'MISSED_RATINGS' AND "compensatedAt" IS NULL`, day))
	for _, emp := range open {
		if items, err := d.Owed(emp, day); err == nil && len(items) == 0 {
			errlog.Exec(d.db.Exec(`UPDATE "EmployeeNegligence" SET "compensatedAt" = now(), reason = 'قيّمها زميله بدله (متأخر)'
				WHERE "employeeId" = $1 AND day = $2::date AND kind = 'MISSED_RATINGS' AND "compensatedAt" IS NULL`, emp, day))
		}
	}
}

func namesUpTo(xs []string, n int) []string {
	if len(xs) > n {
		return append(xs[:n:n], fmt.Sprintf("و%d غيرهم", len(xs)-n))
	}
	return xs
}

// Sweep كل ساعة: يسجّل الإهمال لأيام فاتت (أمس لحد ٣ أيام) حتى لو المقيّم ما فتح
// النظام — السجل يبقى «ما عوّض» لحد ما يعوّض.
func (d *RatingDuty) Sweep() (int, error) {
	n := 0
	for off := -3; off <= -1; off++ {
		day := companyDay(off)
		d.settlePeers(day)
		cands := []string{}
		if err := d.db.Select(&cands, `
			SELECT DISTINCT x FROM (
				SELECT booking_leader_id(b.id) AS x FROM "Booking" b WHERE b."completedAt" IS NOT NULL AND `+bkDay+` = $1::date
				UNION SELECT b."confirmedByEmployeeId" FROM "Booking" b WHERE b."completedAt" IS NOT NULL AND `+bkDay+` = $1::date
				UNION SELECT b."createdById" FROM "Booking" b WHERE b."completedAt" IS NOT NULL AND `+bkDay+` = $1::date
				UNION SELECT e.id FROM "Employee" e WHERE e.status = 'ACTIVE' AND (`+opsRaterSQL+` OR `+monitorRaterSQL+`)
			) t WHERE x IS NOT NULL`, day); err != nil {
			return n, err
		}
		for _, c := range cands {
			items, err := d.Owed(c, day)
			if err != nil || len(items) == 0 {
				continue
			}
			names := []string{}
			for _, it := range items {
				names = append(names, it.RateeName)
			}
			res, _ := d.db.Exec(`INSERT INTO "EmployeeNegligence" ("employeeId", day, kind, detail, reason)
				VALUES ($1, $2::date, 'MISSED_RATINGS', $3, '') ON CONFLICT ("employeeId", day, kind) DO NOTHING`,
				c, day, fmt.Sprintf("فاته %d تقييم: %s", len(items), strings.Join(namesUpTo(names, 8), "، ")))
			if res != nil {
				if k, _ := res.RowsAffected(); k > 0 {
					n++
				}
			}
		}
	}
	return n, nil
}
