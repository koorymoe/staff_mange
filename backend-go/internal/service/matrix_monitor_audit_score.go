package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/repository"
)

// 👁️ تدقيق المراقب (دازاي 11-01): «ماتركس ما يحسب أداء التدقيق الخاص بالمراقب».
// قبل: بس بنود الحجز تنحسب للمراقب (ولآخر واحد حكم بس). بنود الموظفين والشكاوي
// والمشاريع والمهام والتسويات ما چانت تنحسب لأحد. هسه كل بند حكم عليه مراقب ينحسب
// إله بسرعة حكمه: خلال يوم ٢، خلال ٣ أيام ١، أكثر صفر. والبند الي ما انحكم وفاته
// ٣ أيام ينحسب صفر على المراقب الوحيد (إذا أكو أكثر من مراقب ما نظلم أحد).
func (s *MatrixScoreService) scoreMonitorAudit(now time.Time, add func(repository.MatrixScoreRow)) {
	rows, err := s.repo.MonitorAuditDecisions(now.AddDate(0, 0, -scoreLookbackDays))
	if err != nil {
		return
	}
	sole := s.repo.SoleHolder("MONITOR", "monitoring")
	for _, d := range rows {
		pts, ok := respPoints(d.CreatedAt, d.DecidedAt, now, 24*time.Hour, 72*time.Hour)
		emp := sole
		if d.DecidedBy != nil && *d.DecidedBy != "" {
			emp = *d.DecidedBy
		}
		if !ok || emp == "" {
			continue
		}
		label := d.Label
		reason := fmt.Sprintf("%s: انحكم خلال %s", label, humanDur(d.DecidedAt, d.CreatedAt, now))
		at := now
		if d.DecidedAt == nil {
			reason = fmt.Sprintf("%s: بعده ينتظر حكم من %s", label, d.CreatedAt.In(debriefLoc).Format("2006-01-02"))
		} else {
			at = *d.DecidedAt
		}
		add(repository.MatrixScoreRow{EmployeeID: emp, Source: "MONITOR", SourceID: d.ID, SourceLabel: &label,
			Rule: "MONITOR_AUDIT", Points: pts, MaxPoints: 2, At: at, Reason: reason})
	}
}
