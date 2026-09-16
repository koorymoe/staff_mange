package repository

import "staffmange-api/internal/model"

// priceValueExpr يحوّل عمود السعر (نص حر يكتبه المستخدم مثل "1,500,000 د.ع")
// لرقم: نشيل كل شي مو رقم، ولو ما ضل شي نرجّع NULL بدل ما نطيح الاستعلام.
const priceValueExpr = `NULLIF(regexp_replace(COALESCE(p.price, ''), '\D', '', 'g'), '')::numeric`

// ProjectValueRows سطر لكل مشروع بقيمته المالية ومنو مرتبط بيه.
func (r *ProjectRepository) ProjectValueRows() ([]model.ProjectValueRow, error) {
	rows := []model.ProjectValueRow{}
	err := r.db.Select(&rows, `
		SELECT p.id, p.code, p.name, p.stage, p."workType", p.priority, p.price,
			`+priceValueExpr+` AS "priceValue",
			c.name AS "createdByName",
			resp.name AS "responsibleName",
			surv.name AS "surveyorName",
			deleg.name AS "delegatedToName",
			(p.survey IS NOT NULL) AS "hasSurvey",
			to_char(p."createdAt", 'YYYY-MM-DD') AS "createdAt"
		FROM "Project" p
		LEFT JOIN "Employee" c     ON c.id = p."createdByEmployeeId"
		LEFT JOIN "Employee" resp  ON resp.id = p."responsibleEmployeeId"
		LEFT JOIN "Employee" surv  ON surv.id = p."surveyorEmployeeId"
		LEFT JOIN "Employee" deleg ON deleg.id = p."delegatedToEmployeeId"
		ORDER BY `+priceValueExpr+` DESC NULLS LAST, p."createdAt" DESC`)
	return rows, err
}

// ProjectEmployeeStats إحصائية كل موظف داخل المشاريع — كل عدد ينحسب بجملة
// فرعية مستقلة حتى ما تتضاعف الأرقام لو الموظف إله أكثر من دور بنفس المشروع
// (مثلاً هو نفسه المسؤول ومنفّذ الكشف).
func (r *ProjectRepository) ProjectEmployeeStats() ([]model.ProjectEmployeeStatRow, error) {
	rows := []model.ProjectEmployeeStatRow{}
	err := r.db.Select(&rows, `
		SELECT e.id AS "employeeId", e.name, e.role,
			(SELECT COUNT(*) FROM "Project" p WHERE p."createdByEmployeeId" = e.id) AS "addedCount",
			(SELECT COUNT(*) FROM "Project" p WHERE p."surveyorEmployeeId" = e.id) AS "surveyAssignedCount",
			(SELECT COUNT(*) FROM "Project" p WHERE p."surveyorEmployeeId" = e.id AND p.survey IS NOT NULL) AS "surveyFilledCount",
			(SELECT COUNT(*) FROM "Project" p WHERE p."responsibleEmployeeId" = e.id) AS "responsibleCount",
			(SELECT COUNT(*) FROM "ProjectDelegationLog" l WHERE l."employeeId" = e.id AND l.action = 'ASSIGN') AS "delegationsReceived",
			(SELECT COUNT(*) FROM "Project" p WHERE p."delegatedToEmployeeId" = e.id) AS "currentlyDelegated",
			COALESCE((SELECT SUM(`+priceValueExpr+`) FROM "Project" p WHERE p."responsibleEmployeeId" = e.id), 0) AS "responsibleValue",
			(SELECT COUNT(*) FROM "Project" p WHERE p."responsibleEmployeeId" = e.id AND p.stage LIKE '%مكتمل%') AS "completedCount"
		FROM "Employee" e
		ORDER BY e.name`)
	if err != nil {
		return nil, err
	}
	// نرجّع بس الموظفين الي إلهم أثر فعلي بالمشاريع — الباقي ضجيج بالجدول
	active := make([]model.ProjectEmployeeStatRow, 0, len(rows))
	for _, row := range rows {
		if row.AddedCount+row.SurveyAssignedCount+row.ResponsibleCount+
			row.DelegationsReceived+row.CurrentlyDelegated > 0 {
			active = append(active, row)
		}
	}
	return active, nil
}
