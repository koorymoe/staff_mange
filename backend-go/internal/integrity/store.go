package integrity

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// Run دورة محفوظة.
type Run struct {
	ID         string    `db:"id" json:"id"`
	StartedAt  time.Time `db:"startedAt" json:"startedAt"`
	FinishedAt time.Time `db:"finishedAt" json:"finishedAt"`
	Results    []byte    `db:"results" json:"-"`
	Problems   int64     `db:"problems" json:"problems"`
	Critical   int64     `db:"critical" json:"critical"`
	Failed     int       `db:"failed" json:"failed"`
	Trigger    string    `db:"trigger" json:"trigger"`
}

type Notifier interface {
	CreateForRole(role, notifType, message string) error
}

// RunAndStore يشغّل الفحوصات، يحفظها، ويبلّغ المالك إذا خلل خطير أو عالي كبر
// عن آخر دورة (أو أول ما يظهر).
func RunAndStore(db *sqlx.DB, trigger string, n Notifier) (*Run, []Result, error) {
	prev := map[string]int64{}
	var last Run
	if err := db.Get(&last, `SELECT * FROM "IntegrityRun" ORDER BY "startedAt" DESC LIMIT 1`); err == nil {
		var pr []Result
		_ = json.Unmarshal(last.Results, &pr)
		for _, r := range pr {
			prev[r.Key] = r.Count
		}
	}
	start := time.Now()
	res := RunAll(db)
	var problems, critical int64
	failed := 0
	grew := []string{}
	for _, r := range res {
		if r.Error != "" {
			failed++
			continue
		}
		problems += r.Count
		if r.Severity == "critical" {
			critical += r.Count
		}
		if (r.Severity == "critical" || r.Severity == "high") && r.Count > prev[r.Key] {
			grew = append(grew, fmt.Sprintf("%s: %d (كان %d)", r.Title, r.Count, prev[r.Key]))
		}
	}
	js, _ := json.Marshal(res)
	var run Run
	err := db.Get(&run, `INSERT INTO "IntegrityRun" ("startedAt", "finishedAt", results, problems, critical, failed, trigger)
		VALUES ($1, now(), $2, $3, $4, $5, $6) RETURNING *`, start, js, problems, critical, failed, trigger)
	if err != nil {
		return nil, res, err
	}
	if n != nil && len(grew) > 0 {
		msg := "🩺 فحص سلامة البيانات: خلل جديد أو زاد — "
		for i, g := range grew {
			if i == 5 {
				msg += fmt.Sprintf(" و%d غيرها", len(grew)-5)
				break
			}
			if i > 0 {
				msg += " · "
			}
			msg += g
		}
		if err := n.CreateForRole("OWNER", "integrity_alert", msg+". التفاصيل بشاشة «سلامة البيانات»."); err != nil {
			log.Printf("[integrity] notify: %v", err)
		}
	}
	return &run, res, nil
}
