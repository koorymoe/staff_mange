package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ ماتركس يقيّم شغل أبو الكميات (قرار (ع) 10-10) ═══
// سرعة الرد على طلبات المواد والأدوات، ومتابعة السيارات كل يوم، وحل نواقص
// الجرد. البند المعلّق ينحسب على إداري الكميات إذا هو الوحيد بدوره.

// respPoints ٢ إذا خلال ٢٤ ساعة، ١ خلال ٧٢، ٠ أكثر. ok=false = بعده ما ينحكم.
func respPoints(created time.Time, decided *time.Time, now time.Time, fast, slow time.Duration) (int, bool) {
	if decided == nil {
		if now.Sub(created) > slow {
			return 0, true
		}
		return 0, false
	}
	d := decided.Sub(created)
	switch {
	case d <= fast:
		return 2, true
	case d <= slow:
		return 1, true
	}
	return 0, true
}

func (s *MatrixScoreService) scoreProcurement(now time.Time, add func(repository.MatrixScoreRow)) {
	sole := s.repo.SoleHolder("PROCUREMENT_ADMIN", "__role_only__")
	who := func(d repository.ProcDecision) string {
		if d.DecidedBy != nil && *d.DecidedBy != "" {
			return *d.DecidedBy
		}
		return sole
	}
	score := func(rows []repository.ProcDecision, source, rule string, fast, slow time.Duration, what string) {
		for _, d := range rows {
			pts, ok := respPoints(d.CreatedAt, d.DecidedAt, now, fast, slow)
			emp := who(d)
			if !ok || emp == "" {
				continue
			}
			label := d.Label
			reason := fmt.Sprintf("%s: انحسم خلال %s", label, humanDur(d.DecidedAt, d.CreatedAt, now))
			if d.DecidedAt == nil {
				reason = fmt.Sprintf("%s: بعده %s من %s", label, what, d.CreatedAt.In(debriefLoc).Format("2006-01-02"))
			}
			add(repository.MatrixScoreRow{EmployeeID: emp, Source: source, SourceID: d.ID, SourceLabel: &label,
				Rule: rule, Points: pts, MaxPoints: 2, At: now, Reason: reason})
		}
	}
	if rows, err := s.repo.MaterialDecisions(false); err == nil {
		score(rows, "PROCUREMENT", "PROC_RESPONSE", 24*time.Hour, 72*time.Hour, "ما انرد عليه")
	}
	if rows, err := s.repo.ToolDecisions(false); err == nil {
		score(rows, "PROCUREMENT", "TOOL_RESPONSE", 24*time.Hour, 72*time.Hour, "ما انرد عليه")
	}
	if rows, err := s.repo.Shortages(false); err == nil {
		score(rows, "PROCUREMENT", "SHORTAGE_RESOLVE", 48*time.Hour, 120*time.Hour, "ما انحل")
	}
	if sole == "" {
		return
	}
	if days, err := s.repo.VehicleDays(); err == nil {
		for _, d := range days {
			if d.Total == 0 {
				continue
			}
			pts := 0
			if d.Rated >= d.Total {
				pts = 2
			} else if d.Rated > 0 {
				pts = 1
			}
			label := "متابعة السيارات " + d.Day
			add(repository.MatrixScoreRow{EmployeeID: sole, Source: "PROCUREMENT", SourceID: "vehicles@" + d.Day, SourceLabel: &label,
				Rule: "VEHICLE_DAILY", Points: pts, MaxPoints: 2, At: now,
				Reason: fmt.Sprintf("%s: انقيّمت %d من %d سيارة", label, d.Rated, d.Total)})
		}
	}
}

func humanDur(decided *time.Time, created, now time.Time) string {
	end := now
	if decided != nil {
		end = *decided
	}
	h := end.Sub(created).Hours()
	if h < 24 {
		return fmt.Sprintf("%.0f ساعة", h)
	}
	return fmt.Sprintf("%.0f يوم", h/24)
}
