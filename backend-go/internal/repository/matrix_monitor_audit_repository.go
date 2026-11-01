package repository

import "time"

// MonitorAuditDecisions بنود صندوق المراقب الي مو مربوطة بحجز (الحجز ينحسب بسلسلة الحجز):
// موظف، شكوى، مشروع، مهمة، تسوية فاتورة، حكم ذكاء… كل بند حكم عليه مراقب ينحسب إله.
func (r *MatrixScoreRepository) MonitorAuditDecisions(since time.Time) ([]ProcDecision, error) {
	rows := []ProcDecision{}
	err := r.db.Select(&rows, `SELECT m.id, 'تدقيق: ' || COALESCE(NULLIF(m.title, ''), m."entityType") AS label,
			m."createdAt", m."reviewedAt" AS "decidedAt", m."reviewedById" AS "decidedBy", e.name AS "byName"
		FROM "MonitorReview" m LEFT JOIN "Employee" e ON e.id = m."reviewedById"
		WHERE m."entityType" NOT IN ('BOOKING', 'LEADER_INVOICE')
		  AND (m."createdAt" >= $1 OR m."reviewedAt" >= $1)
		ORDER BY m."createdAt"`, since)
	return rows, err
}
