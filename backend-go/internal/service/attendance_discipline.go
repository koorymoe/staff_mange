package service

import (
	"math"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ═══ الانضباط بالدوام (دازاي 11-11: «التأخر بالعمل بدون تبرير بدون عذر هاي مشكلة») ═══
//
// قبل: النظام يحسب التأخر للعرض بتقرير الحضور بس، والتقييم يشوف «سجّل لو لا» —
// وحتى هذا بالاعتمادية بس (قرار 10-06). فالمتأخر كل يوم ياخذ نفس رقم الملتزم.
//
// هسه (يبدي من DisciplineStart — أول الشهر الجاي حتى الكادر يتعوّد على طلب العذر):
//   - كل يوم دوام إجباري (نفس أيام قاعدة ATTENDANCE بماتركس) ينحكم:
//     بوقته · متأخر بعذر مقبول · متأخر بلا عذر · غياب بلا إجازة.
//   - المتأخر = أول حضور بعد بداية الشفت + نص ساعة سماح (lateGrace).
//   - الإجازة المعتمدة والعذر المقبول ما يضرّون.
//   - الدرجة من ١٠٠: −٢ لكل يوم تأخر بلا عذر، −٥ لكل غياب بلا إجازة.
//   - تدخل ١٥٪ من التقييم الشهري للموظف، ومعيار «انضباط الكادر» بتصنيف المدراء.

var DisciplineStart = time.Date(2026, 12, 1, 0, 0, 0, 0, debriefLoc)

const (
	disciplineLatePts   = 2
	disciplineAbsentPts = 5
	// DisciplineWeight حصة الانضباط من التقييم الشهري النهائي.
	DisciplineWeight = 0.15
)

// DisciplineDay يوم مو «بوقته» — هذا الدليل الي يشوفه الموظف والمالك.
type DisciplineDay struct {
	Date    string `json:"date"`
	Status  string `json:"status"` // LATE · LATE_EXCUSED · ABSENT
	LateMin int    `json:"lateMinutes,omitempty"`
	CheckIn string `json:"checkIn,omitempty"`
	Shift   string `json:"shiftStart,omitempty"`
}

// Discipline ملخص انضباط موظف بفترة.
type Discipline struct {
	Score       *float64        `json:"score"`
	Days        int             `json:"days"`
	OnTime      int             `json:"onTime"`
	Late        int             `json:"late"`
	LateExcused int             `json:"lateExcused"`
	Absent      int             `json:"absent"`
	LateMinutes int             `json:"lateMinutes"`
	Starts      string          `json:"startsOn"`
	Items       []DisciplineDay `json:"items"`
}

// DisciplineScore الدرجة من الأعداد — منفصلة حتى تنفحص.
func DisciplineScore(late, absent int) float64 {
	return math.Max(0, float64(100-late*disciplineLatePts-absent*disciplineAbsentPts))
}

// Disciplines انضباط مجموعة موظفين بفترة (ids فاضية = الكل). اليوم الحالي ما ينحسب
// لأن الحكم عليه بعده ما خلص. الأيام قبل DisciplineStart ما تنحسب.
func Disciplines(db *sqlx.DB, ids []string, from, to, now time.Time) (map[string]*Discipline, error) {
	out := map[string]*Discipline{}
	start := from
	if start.Before(DisciplineStart) {
		start = DisciplineStart
	}
	n := now.In(debriefLoc)
	today := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, debriefLoc)
	end := to
	if end.After(today) {
		end = today
	}
	if !start.Before(end) {
		return out, nil
	}
	if ids == nil {
		ids = []string{} // pq.Array(nil) = NULL، و cardinality(NULL) ما تساوي صفر
	}
	// أيام الدوام الإجباري من قاعدة ATTENDANCE (نفس الي يقرر الغياب بماتركس):
	// points = 0 → ما سجّل وما عنده إجازة = غياب. points > 0 → حضر، نشوف التأخر.
	days := []struct {
		EmployeeID string     `db:"employeeId"`
		Day        time.Time  `db:"day"`
		Points     int        `db:"points"`
		Shift      *string    `db:"shift"`
		ShiftStart *string    `db:"shiftStart"`
		ShiftEnd   *string    `db:"shiftEnd"`
		CheckIn    *time.Time `db:"checkIn"`
		Excused    bool       `db:"excused"`
		Flexible   bool       `db:"flexible"`
	}{}
	err := db.Select(&days, `
		SELECT s."employeeId", baghdad_date(s."at")::timestamp AS day, s.points,
		       e.shift::text AS shift, e."shiftStart", e."shiftEnd", e."flexibleHours" AS flexible,
		       (SELECT min(a."checkIn") FROM "Attendance" a WHERE a."employeeId" = s."employeeId"
		         AND baghdad_date(a."checkIn") = baghdad_date(s."at")) AS "checkIn",
		       EXISTS (SELECT 1 FROM "LateExcuse" x WHERE x."employeeId" = s."employeeId"
		         AND x.date = baghdad_date(s."at") AND x.status = 'APPROVED') AS excused
		FROM "MatrixScore" s JOIN "Employee" e ON e.id = s."employeeId"
		WHERE s.rule = 'ATTENDANCE' AND s."cancelledAt" IS NULL
		  AND s."at" >= $1 AND s."at" < $2 AND e.role::text NOT IN ('ADMIN','OWNER')
		  AND (cardinality($3::text[]) = 0 OR s."employeeId" = ANY($3))
		ORDER BY 2`, start.UTC(), end.UTC(), pq.Array(ids))
	if err != nil {
		return nil, err
	}
	for _, d := range days {
		x := out[d.EmployeeID]
		if x == nil {
			x = &Discipline{Items: []DisciplineDay{}, Starts: DisciplineStart.Format("2006-01-02")}
			out[d.EmployeeID] = x
		}
		x.Days++
		date := time.Date(d.Day.Year(), d.Day.Month(), d.Day.Day(), 0, 0, 0, 0, debriefLoc)
		item := DisciplineDay{Date: date.Format("2006-01-02")}
		switch {
		case d.Points == 0:
			x.Absent++
			item.Status = "ABSENT"
		case d.CheckIn == nil || d.Flexible:
			// انحسب حاضر بماتركس (تصحيح يدوي/إجازة) — ما نحكم بلا وقت.
			// وبدون جدول محدد (دازاي 11-13): ماكو تأخير أصلاً.
			x.OnTime++
			continue
		default:
			from, _ := ShiftWindow(d.Shift, d.ShiftStart, d.ShiftEnd, date)
			if !d.CheckIn.After(from.Add(lateGrace)) {
				x.OnTime++
				continue
			}
			item.LateMin = int(d.CheckIn.Sub(from).Minutes())
			item.CheckIn = d.CheckIn.In(debriefLoc).Format("15:04")
			item.Shift = from.Format("15:04")
			if d.Excused {
				x.LateExcused++
				item.Status = "LATE_EXCUSED"
			} else {
				x.Late++
				x.LateMinutes += item.LateMin
				item.Status = "LATE"
			}
		}
		x.Items = append(x.Items, item)
	}
	for _, x := range out {
		s := DisciplineScore(x.Late, x.Absent)
		x.Score = &s
		sort.SliceStable(x.Items, func(i, j int) bool { return x.Items[i].Date < x.Items[j].Date })
	}
	return out, nil
}

// applyDiscipline يدخّل الانضباط بالتقييم النهائي: ٨٥٪ الرقم القديم + ١٥٪ الانضباط.
func applyDiscipline(rows []StaffScore, disc map[string]*Discipline) {
	for i := range rows {
		d := disc[rows[i].ID]
		if d == nil || d.Score == nil {
			continue
		}
		rows[i].Discipline = d
		if rows[i].Final != nil {
			f := *rows[i].Final*(1-DisciplineWeight) + *d.Score*DisciplineWeight
			rows[i].Final = &f
		}
	}
}
