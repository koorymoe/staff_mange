package repository

import (
	"encoding/json"
	"time"
)

// ═══ لقطة الواجبات اليومية — قرار (ع) 10-22 ═══
// الواجبات تنحسب لحظياً، فبلا لقطة ماكو تاريخ: «فلان ثلاث أيام من سبعة ما
// كمّل واجباته» ما ينعرف. كل ليلة قبل نص الليل (بغداد) ينحفظ تقرير اليوم لكل
// موظف فعّال — نفس Today() بالضبط، فالي شافه الموظف هو الي ينحفظ.

// dutySnapshotFrom ساعة اللقطة بتوقيت بغداد (١١:٥٠ بالليل).
const dutySnapshotFrom = 23*60 + 50

// SnapshotIfDue يلقط اليوم إذا وصل وقته وبعده ما انلقط. يرجّع عدد الموظفين.
func (r *DutiesRepository) SnapshotIfDue(now time.Time) (int, error) {
	b := now.In(baghdadLoc())
	if b.Hour()*60+b.Minute() < dutySnapshotFrom {
		return 0, nil
	}
	return r.SnapshotToday()
}

// SnapshotToday يلقط كل موظف فعّال ما انلقط اليوم (المالك والمدير ما ينقيّمون).
func (r *DutiesRepository) SnapshotToday() (int, error) {
	ids := []string{}
	if err := r.db.Select(&ids, `SELECT e.id FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER')
		  AND NOT EXISTS (SELECT 1 FROM "DutySnapshot" s WHERE s."employeeId" = e.id AND s.day = baghdad_today())`); err != nil {
		return 0, err
	}
	n := 0
	for _, id := range ids {
		rep := r.Today(id)
		if rep.Position == "ADMIN" {
			continue
		}
		raw, err := json.Marshal(rep.Duties)
		if err != nil {
			continue
		}
		res, err := r.db.Exec(`INSERT INTO "DutySnapshot" ("employeeId", day, position, "doneCount", total, verdict, off, duties)
			VALUES ($1, baghdad_today(), $2, $3, $4, $5, NULLIF($6, ''), $7) ON CONFLICT ("employeeId", day) DO NOTHING`,
			id, rep.Position, rep.DoneCount, rep.Total, rep.Verdict, rep.Off, raw)
		if err != nil {
			return n, err
		}
		if k, _ := res.RowsAffected(); k > 0 {
			n++
		}
	}
	return n, nil
}

// DutySnapshotRow يوم من سجل الموظف.
type DutySnapshotRow struct {
	Day       string          `db:"day" json:"day"`
	Position  string          `db:"position" json:"position"`
	DoneCount int             `db:"doneCount" json:"doneCount"`
	Total     int             `db:"total" json:"total"`
	Verdict   string          `db:"verdict" json:"verdict"`
	Off       *string         `db:"off" json:"off"`
	Duties    json.RawMessage `db:"duties" json:"duties"`
}

// History آخر الأيام الملقوطة للموظف (الأحدث أول).
func (r *DutiesRepository) History(employeeID string, days int) ([]DutySnapshotRow, error) {
	rows := []DutySnapshotRow{}
	err := r.db.Select(&rows, `SELECT to_char(day, 'YYYY-MM-DD') AS day, position, "doneCount", total, verdict, off, duties
		FROM "DutySnapshot" WHERE "employeeId" = $1 AND day > baghdad_today() - $2::int
		ORDER BY day DESC`, employeeID, days)
	return rows, err
}

func baghdadLoc() *time.Location {
	if l, err := time.LoadLocation("Asia/Baghdad"); err == nil {
		return l
	}
	return time.FixedZone("Asia/Baghdad", 3*3600)
}
