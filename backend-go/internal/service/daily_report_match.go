package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"staffmange-api/internal/errlog"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// ═══ 🔎 مطابقة تقرير الإنجاز ويا الشغل المسجّل (دازاي 11-02 — تعديلات ٩.٤) ═══
// قواعد ثابتة بلا AI: شنو كتب بالتقرير ↔ شنو سجّل النظام (نشاطه + حجوزاته).
//   MATCH    عنده أثر وكل حجز ذكره لمسه فعلاً
//   PARTIAL  ذكر حجز ما لمسه، أو اشتغل على حجز وما ذكره
//   NO_TRACE كتب تقرير بس ماكو ولا أثر بالنظام لذاك اليوم
// الفرق يروح للمراقب ويدخل بماتركس (REPORT_MATCH). المراقب يكدر يگول «مبرّر».

type ReportMatch struct {
	ReportID       string         `json:"reportId"`
	EmployeeID     string         `json:"employeeId"`
	Day            string         `json:"day"`
	Status         string         `json:"status"`
	Note           string         `json:"note"`
	ClaimedCodes   []string       `json:"claimedCodes"`
	MissingCodes   []string       `json:"missingCodes"`
	UnclaimedCodes []string       `json:"unclaimedCodes"`
	ActivityCount  int            `json:"activityCount"`
	Activity       []ActivityLine `json:"activity"`
}

type ReportMatcher struct {
	db  *sqlx.DB
	act *EmployeeActivityService
}

func NewReportMatcher(db *sqlx.DB) *ReportMatcher {
	return &ReportMatcher{db: db, act: NewEmployeeActivityService(repository.NewEmployeeActivityRepository(db))}
}

// الكود ممكن ينتهي بحروف (SIM1A) — نطابق الكلمة كاملة ويا كود حقيقي بالنظام (٩.٦ لگتها).
var codeToken = regexp.MustCompile(`[A-Za-z]{1,5}-?[0-9]{1,7}[A-Za-z]{0,3}`)

func dayOf(col string) string {
	return `((b."` + col + `" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad')::date = $2::date)`
}

// touched الحجوزات الي إله يد بيها ذاك اليوم: معيّن عليها وانجزت/انجدولت، أو سجّلها/ثبّتها/جهّزها/سلّمها.
func (m *ReportMatcher) touched(emp, day string) (assigned, acted []string, err error) {
	err = m.db.Select(&assigned, `SELECT DISTINCT b.code FROM "Booking" b JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		WHERE ba."employeeId" = $1 AND (`+dayOf("completedAt")+` OR `+dayOf("scheduledAt")+`) AND b.code IS NOT NULL`, emp, day)
	if err != nil {
		return
	}
	err = m.db.Select(&acted, `SELECT DISTINCT b.code FROM "Booking" b WHERE b.code IS NOT NULL AND (
		(b."createdById" = $1 AND `+dayOf("createdAt")+`) OR (b."confirmedByEmployeeId" = $1 AND `+dayOf("confirmedAt")+`)
		OR (b."materialsReadyById" = $1 AND `+dayOf("materialsReadyAt")+`) OR (b."handoverById" = $1 AND `+dayOf("handoverAt")+`)
		OR (b."lastEditedById" = $1 AND `+dayOf("lastEditedAt")+`))`, emp, day)
	return
}

func upperSet(xs []string) map[string]string {
	s := map[string]string{}
	for _, x := range xs {
		s[strings.ToUpper(strings.TrimSpace(x))] = x
	}
	return s
}

// Match يطابق تقرير موظف ليوم. يرجّع nil إذا ماكو تقرير.
func (m *ReportMatcher) Match(emp, day string) (*ReportMatch, error) {
	var rep struct {
		ID   string `db:"id"`
		Body string `db:"body"`
	}
	if err := m.db.Get(&rep, `SELECT id, body FROM "DailyWorkReport" WHERE "employeeId" = $1 AND day = $2::date`, emp, day); err != nil {
		return nil, nil
	}
	assigned, acted, err := m.touched(emp, day)
	if err != nil {
		return nil, fmt.Errorf("touched: %w", err)
	}
	out := &ReportMatch{ReportID: rep.ID, EmployeeID: emp, Day: day, ClaimedCodes: []string{}, MissingCodes: []string{}, UnclaimedCodes: []string{}}
	out.Activity = m.act.Day(emp, day)
	for _, a := range out.Activity {
		out.ActivityCount += a.Count
	}
	touchedAll := upperSet(append(append([]string{}, assigned...), acted...))
	// أكواد بالنص تطابق حجز حقيقي بالنظام (أي كلمة ثانية نتركها)
	tokens := map[string]bool{}
	for _, t := range codeToken.FindAllString(rep.Body, -1) {
		tokens[strings.ToUpper(t)] = true
	}
	if len(tokens) > 0 {
		list := []string{}
		for t := range tokens {
			list = append(list, t)
		}
		real := []string{}
		if err := m.db.Select(&real, `SELECT DISTINCT upper(code) FROM "Booking" WHERE upper(code) = ANY($1)`, pq.Array(list)); err != nil {
			return nil, fmt.Errorf("codes: %w", err)
		}
		for _, c := range real {
			out.ClaimedCodes = append(out.ClaimedCodes, c)
			if _, ok := touchedAll[c]; !ok {
				out.MissingCodes = append(out.MissingCodes, c)
			}
		}
	}
	for _, c := range assigned {
		if !tokens[strings.ToUpper(c)] {
			out.UnclaimedCodes = append(out.UnclaimedCodes, c)
		}
	}
	hasTrace := out.ActivityCount > 0 || len(touchedAll) > 0
	switch {
	case !hasTrace:
		out.Status, out.Note = "NO_TRACE", "كتب تقرير بس ماكو ولا شغل مسجّل باسمه بالنظام لهذا اليوم."
	case len(out.MissingCodes) > 0 || len(out.UnclaimedCodes) > 0:
		out.Status = "PARTIAL"
		parts := []string{}
		if len(out.MissingCodes) > 0 {
			parts = append(parts, "ذكر حجوزات ما إله أثر بيها: "+strings.Join(out.MissingCodes, "، "))
		}
		if len(out.UnclaimedCodes) > 0 {
			parts = append(parts, "اشتغل على حجوزات وما ذكرها: "+strings.Join(out.UnclaimedCodes, "، "))
		}
		out.Note = strings.Join(parts, ". ") + "."
	default:
		out.Status, out.Note = "MATCH", fmt.Sprintf("التقرير يطابق المسجّل (%d فعل بالنظام).", out.ActivityCount+len(touchedAll))
	}
	return out, nil
}

// Save يخزّن النتيجة (ويحافظ على مراجعة المراقب). يرجّع true إذا الحالة جديدة وبيها فرق.
func (m *ReportMatcher) Save(r *ReportMatch) (bool, error) {
	detail, _ := json.Marshal(map[string]any{"claimed": r.ClaimedCodes, "missing": r.MissingCodes, "unclaimed": r.UnclaimedCodes, "activityCount": r.ActivityCount})
	var prev *string
	errlog.Err(m.db.Get(&prev, `SELECT status FROM "DailyReportMatch" WHERE "employeeId" = $1 AND day = $2::date`, r.EmployeeID, r.Day))
	_, err := m.db.Exec(`INSERT INTO "DailyReportMatch" ("employeeId", day, "reportId", status, note, detail) VALUES ($1, $2::date, $3, $4, $5, $6)
		ON CONFLICT ("employeeId", day) DO UPDATE SET "reportId" = EXCLUDED."reportId", status = EXCLUDED.status, note = EXCLUDED.note,
		detail = EXCLUDED.detail, "updatedAt" = now()`, r.EmployeeID, r.Day, r.ReportID, r.Status, r.Note, string(detail))
	isNew := r.Status != "MATCH" && (prev == nil || *prev != r.Status)
	return isNew, err
}

// SweepDay يطابق كل تقارير يوم. يرجّع الي بيها فرق جديد (للإشعار).
func (m *ReportMatcher) SweepDay(day string) (flagged []*ReportMatch, err error) {
	emps := []string{}
	if err = m.db.Select(&emps, `SELECT d."employeeId" FROM "DailyWorkReport" d JOIN "Employee" e ON e.id = d."employeeId"
		WHERE d.day = $1::date AND e.role::text NOT IN ('ADMIN','OWNER')`, day); err != nil {
		return
	}
	for _, e := range emps {
		r, e2 := m.Match(e, day)
		if e2 != nil {
			return flagged, e2
		}
		if r == nil {
			continue
		}
		isNew, e2 := m.Save(r)
		if e2 != nil {
			return flagged, e2
		}
		if isNew {
			flagged = append(flagged, r)
		}
	}
	return
}

// MatchRate نسبة المطابقة (MATCH أو مبرّر) آخر n يوم — للملف الشامل.
func MatchRate(db *sqlx.DB, emp string, days int) (pct int, n int) {
	var row struct {
		N  int `db:"n"`
		OK int `db:"ok"`
	}
	errlog.Err(db.Get(&row, `SELECT count(*) AS n, count(*) FILTER (WHERE status = 'MATCH' OR "reviewVerdict" = 'JUSTIFIED') AS ok
		FROM "DailyReportMatch" WHERE "employeeId" = $1 AND day >= current_date - $2::int`, emp, days))
	if row.N == 0 {
		return 0, 0
	}
	return row.OK * 100 / row.N, row.N
}
