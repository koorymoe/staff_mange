package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/timeutil"
)

// ═══ عذر التأخير (دازاي 11-11) ═══
// الموظف يقدّم عذر ليوم (اليوم، أو آخر ٣ أيام، أو لحد أسبوع قدّام)، ومسؤوله المباشر
// يوافق أو يرفض. المقبول ما ينحسب عليه بالانضباط. الغياب الكامل يمشي بطلب إجازة.
//
// منو يوافق (الأعلى مباشرة بالهيكل التنظيمي):
//   - مدير/نائب الإدارة: كل منتسبي الوحدات بإدارته، وأعضاء الإدارة؛ والمدير يوافق للنائب.
//   - مدير/نائب الوحدة: أعضاء وحدته؛ ومدير الوحدة يوافق لنائبها.
//   - مدير الإدارة نفسه، والي ماله مكان بالهيكل: المالك أو مدير النظام.
//   - ماكو أحد يوافق على عذره هو.

var ErrExcuseHidden = errors.New("مو موجود")

type LateExcuse struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	EmployeeName string     `db:"employeeName" json:"employeeName"`
	Date         time.Time  `db:"date" json:"date"`
	Minutes      *int       `db:"minutes" json:"minutes"`
	Reason       string     `db:"reason" json:"reason"`
	Status       string     `db:"status" json:"status"`
	DecidedBy    *string    `db:"decidedBy" json:"decidedBy"`
	DecidedAt    *time.Time `db:"decidedAt" json:"decidedAt"`
	DecisionNote *string    `db:"decisionNote" json:"decisionNote"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
}

type LateExcuseService struct {
	db    *sqlx.DB
	notif *repository.NotificationRepository
}

func NewLateExcuseService(db *sqlx.DB, notif *repository.NotificationRepository) *LateExcuseService {
	return &LateExcuseService{db: db, notif: notif}
}

// excuseApproverSQL: $1 = الموافِق، x."employeeId" = صاحب العذر.
const excuseApproverSQL = `EXISTS (
	SELECT 1 FROM "OrgAssignment" ap
	JOIN "OrgAssignment" em ON em."divisionId" = ap."divisionId" AND em."endedAt" IS NULL AND em."employeeId" = x."employeeId"
	WHERE ap."employeeId" = $1 AND ap."endedAt" IS NULL AND ap."employeeId" <> em."employeeId"
	  AND ap.position IN ('DIRECTOR','DEPUTY') AND (
		(ap."unitId" IS NULL AND (em."unitId" IS NOT NULL OR em.position = 'MEMBER' OR (ap.position = 'DIRECTOR' AND em.position = 'DEPUTY')))
		OR (ap."unitId" IS NOT NULL AND em."unitId" = ap."unitId" AND (em.position = 'MEMBER' OR (ap.position = 'DIRECTOR' AND em.position = 'DEPUTY')))))`

const excuseSelect = `SELECT x.id, x."employeeId", e.name AS "employeeName", x.date, x.minutes, x.reason, x.status,
	d.name AS "decidedBy", x."decidedAt", x."decisionNote", x."createdAt"
	FROM "LateExcuse" x JOIN "Employee" e ON e.id = x."employeeId" LEFT JOIN "Employee" d ON d.id = x."decidedById"`

func isTop(role string) bool { return model.IsAdminRole(role) }

// Submit يقدّم (أو يعدّل وهو معلّق) عذر ليوم.
func (s *LateExcuseService) Submit(empID, day, reason string, minutes *int) (*LateExcuse, error) {
	reason = strings.TrimSpace(reason)
	if len([]rune(reason)) < 3 {
		return nil, errors.New("اكتب سبب التأخير")
	}
	d, err := time.ParseInLocation("2006-01-02", day, debriefLoc)
	if err != nil {
		return nil, errors.New("التاريخ لازم بشكل YYYY-MM-DD")
	}
	n := time.Now().In(debriefLoc)
	today := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, debriefLoc)
	if d.Before(today.AddDate(0, 0, -3)) || d.After(today.AddDate(0, 0, 7)) {
		return nil, errors.New("العذر يكون لليوم، أو لآخر ٣ أيام، أو لحد أسبوع قدّام")
	}
	if minutes != nil && (*minutes <= 0 || *minutes > 600) {
		minutes = nil
	}
	var id string
	err = s.db.Get(&id, `INSERT INTO "LateExcuse" ("employeeId", date, minutes, reason) VALUES ($1, $2, $3, $4)
		ON CONFLICT ("employeeId", date) DO UPDATE SET reason = EXCLUDED.reason, minutes = EXCLUDED.minutes, "createdAt" = now()
		  WHERE "LateExcuse".status = 'PENDING'
		RETURNING id`, empID, day, minutes, reason)
	if err != nil {
		return nil, errors.New("عندك عذر محسوم لهذا اليوم — ما يتعدّل")
	}
	s.notifyApprovers(empID, day)
	return s.get(id)
}

// notifyApprovers يبلّغ الموافقين المباشرين؛ وإذا ماكو، المالك ومدير النظام.
func (s *LateExcuseService) notifyApprovers(empID, day string) {
	if s.notif == nil {
		return
	}
	var name string
	_ = s.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, empID)
	ids := []string{}
	_ = s.db.Select(&ids, `SELECT DISTINCT who.id FROM "Employee" who, (SELECT $1::text AS "employeeId") x
		WHERE who.role::text NOT IN ('ADMIN','OWNER') AND `+strings.ReplaceAll(excuseApproverSQL, "$1", "who.id"), empID)
	msg := "⏰ عذر تأخير من " + name + " ليوم " + day + " — بانتظار موافقتك (الحضور ← أعذار التأخير)"
	if len(ids) == 0 {
		_ = s.notif.CreateForRole("ADMIN", "LATE_EXCUSE", msg)
		return
	}
	for _, id := range ids {
		_ = s.notif.Create(id, "LATE_EXCUSE", msg)
	}
}

func (s *LateExcuseService) get(id string) (*LateExcuse, error) {
	var x LateExcuse
	if err := s.db.Get(&x, excuseSelect+` WHERE x.id = $1`, id); err != nil {
		return nil, ErrExcuseHidden
	}
	return &x, nil
}

// Mine أعذاري (آخر ٩٠ يوم).
func (s *LateExcuseService) Mine(empID string) ([]LateExcuse, error) {
	out := []LateExcuse{}
	err := s.db.Select(&out, excuseSelect+` WHERE x."employeeId" = $1 AND x.date >= baghdad_today() - 90 ORDER BY x.date DESC`, empID)
	return out, err
}

// Inbox الأعذار الي أگدر أبت بيها: المعلّقة، وآخر المحسومة (٣٠ يوم).
func (s *LateExcuseService) Inbox(empID, role string) ([]LateExcuse, error) {
	out := []LateExcuse{}
	cond := excuseApproverSQL
	if isTop(role) {
		// المالك ومدير النظام: الكل، عدا عذرهم هم
		cond = `x."employeeId" <> $1`
	}
	err := s.db.Select(&out, excuseSelect+` WHERE `+cond+`
		AND (x.status = 'PENDING' OR x."decidedAt" > now() - interval '30 days')
		ORDER BY (x.status = 'PENDING') DESC, x.date DESC LIMIT 300`, empID)
	return out, err
}

// CanDecide هل هذا الشخص يگدر يبت بعذر هذا الموظف.
func (s *LateExcuseService) CanDecide(approverID, role, excuseID string) bool {
	var ok bool
	cond := excuseApproverSQL
	if isTop(role) {
		cond = `x."employeeId" <> $1`
	}
	_ = s.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "LateExcuse" x WHERE x.id = $2 AND `+cond+`)`, approverID, excuseID)
	return ok
}

// Decide موافقة أو رفض. الي مو مسؤول عنه = مو موجود (404).
func (s *LateExcuseService) Decide(approverID, role, id string, approve bool, note string) (*LateExcuse, error) {
	if !s.CanDecide(approverID, role, id) {
		return nil, ErrExcuseHidden
	}
	status := "REJECTED"
	if approve {
		status = "APPROVED"
	}
	res, err := s.db.Exec(`UPDATE "LateExcuse" SET status = $2, "decidedById" = $3, "decidedAt" = now(), "decisionNote" = NULLIF($4,'')
		WHERE id = $1 AND status = 'PENDING'`, id, status, approverID, strings.TrimSpace(note))
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("العذر محسوم من قبل")
	}
	x, err := s.get(id)
	if err == nil && s.notif != nil {
		word := "انرفض"
		if approve {
			word = "انقبل ✅"
		}
		_ = s.notif.Create(x.EmployeeID, "LATE_EXCUSE", "⏰ عذر تأخيرك ليوم "+x.Date.Format("2006-01-02")+" "+word)
	}
	return x, err
}

// DisciplineFor انضباط موظف بشهر (للموظف نفسه ولشاشة المراقبة).
func (s *LateExcuseService) DisciplineFor(empID string, from, to time.Time) *Discipline {
	m, err := Disciplines(s.db, []string{empID}, from, to, time.Now())
	if err != nil {
		return nil
	}
	return m[empID]
}

// ExcuseApprover كم عذر قبل/رفض كل مسؤول — حتى تبين المحاباة (مراقبة التقييم).
type ExcuseApprover struct {
	Name     string `db:"name" json:"name"`
	Approved int    `db:"approved" json:"approved"`
	Rejected int    `db:"rejected" json:"rejected"`
}

// Audit أعذار الشهر كلها، ومين وافق على شنو (للمالك).
func (s *LateExcuseService) Audit(from, to time.Time) ([]ExcuseApprover, []LateExcuse, error) {
	appr := []ExcuseApprover{}
	if err := s.db.Select(&appr, `SELECT d.name, count(*) FILTER (WHERE x.status = 'APPROVED') AS approved,
			count(*) FILTER (WHERE x.status = 'REJECTED') AS rejected
		FROM "LateExcuse" x JOIN "Employee" d ON d.id = x."decidedById"
		WHERE x.date >= $1::date AND x.date < $2::date GROUP BY d.name ORDER BY approved DESC`, from, to); err != nil {
		return nil, nil, err
	}
	rows := []LateExcuse{}
	err := s.db.Select(&rows, excuseSelect+` WHERE x.date >= $1::date AND x.date < $2::date ORDER BY x.date DESC LIMIT 500`, from, to)
	return appr, rows, err
}

// DisciplineAlerts (ماتركس) — كل ما يوصل التأخير بلا عذر لـ٣ و٦ و٩… بالشهر، ينبّه
// الموظف ومسؤوله المباشر مرة وحدة لكل عتبة. يشتغل بعد ٢ الظهر بغداد.
func (s *LateExcuseService) DisciplineAlerts(now time.Time, claim func(key, day string) (bool, error)) (int, error) {
	n := now.In(debriefLoc)
	if n.Hour() < 14 || s.notif == nil {
		return 0, nil
	}
	month := n.Format("2006-01")
	from, to, _ := timeutil.MonthStart(month)
	all, err := Disciplines(s.db, nil, from, to, now)
	if err != nil {
		return 0, err
	}
	sent := 0
	for id, d := range all {
		if d.Late < 3 {
			continue
		}
		step := d.Late / 3 * 3
		ok, err := claim("DISC_LATE_"+id+"_"+strconv.Itoa(step), from.Format("2006-01-02"))
		if err != nil || !ok {
			continue
		}
		var name string
		_ = s.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, id)
		mins := strconv.Itoa(d.LateMinutes)
		_ = s.notif.Create(id, "DISCIPLINE", "⏰ تأخرت "+strconv.Itoa(d.Late)+" مرات بلا عذر هالشهر ("+mins+" دقيقة) — كل مرة تنزّل تقييمك. إذا عندك سبب قدّم عذر من صفحة الحضور.")
		mgr := []string{}
		_ = s.db.Select(&mgr, `SELECT DISTINCT who.id FROM "Employee" who, (SELECT $1::text AS "employeeId") x
		WHERE who.role::text NOT IN ('ADMIN','OWNER') AND `+strings.ReplaceAll(excuseApproverSQL, "$1", "who.id"), id)
		msg := "⏰ " + name + " تأخر " + strconv.Itoa(d.Late) + " مرات بلا عذر هالشهر (" + mins + " دقيقة) — تابعه"
		if len(mgr) == 0 {
			_ = s.notif.CreateForRole("ADMIN", "DISCIPLINE", msg)
		}
		for _, m := range mgr {
			_ = s.notif.Create(m, "DISCIPLINE", msg)
		}
		sent++
	}
	return sent, nil
}
