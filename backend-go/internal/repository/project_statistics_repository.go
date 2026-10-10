package repository

import (
	"database/sql"

	"staffmange-api/internal/model"
)

// priceValueExpr يحوّل عمود السعر (نص حر يكتبه المستخدم مثل "1,500,000 د.ع")
// لرقم: نشيل كل شي مو رقم، ولو ما ضل شي نرجّع NULL بدل ما نطيح الاستعلام.
const priceValueExpr = `NULLIF(regexp_replace(COALESCE(p.price, ''), '\D', '', 'g'), '')::numeric`

// ProjectValueRows سطر لكل مشروع بقيمته المالية ومنو مرتبط بيه.
func (r *ProjectRepository) ProjectValueRows() ([]model.ProjectValueRow, error) {
	rows := []model.ProjectValueRow{}
	err := r.db.Select(&rows, `
		SELECT p.id, p.code, p.name, p.stage, p."workType", p.priority, p.price,
			`+priceValueExpr+` AS "priceValue",
			(SELECT q."netTotal" FROM "Quotation" q WHERE q."projectId" = p.id ORDER BY q."createdAt" DESC LIMIT 1) AS "quotedAmount",
			COALESCE(pp.paid, 0) AS "paidAmount", COALESCE(pp.unverified, 0) AS "unverifiedAmount",
			c.name AS "createdByName",
			resp.name AS "responsibleName", p."responsibleEmployeeId" AS "responsibleId",
			surv.name AS "surveyorName",
			deleg.name AS "delegatedToName",
			(p.survey IS NOT NULL) AS "hasSurvey",
			to_char(p."createdAt", 'YYYY-MM-DD') AS "createdAt"
		FROM "Project" p
		LEFT JOIN "Employee" c     ON c.id = p."createdByEmployeeId"
		LEFT JOIN "Employee" resp  ON resp.id = p."responsibleEmployeeId"
		LEFT JOIN "Employee" surv  ON surv.id = p."surveyorEmployeeId"
		LEFT JOIN "Employee" deleg ON deleg.id = p."delegatedToEmployeeId"
		LEFT JOIN LATERAL (SELECT SUM(amount) AS paid, SUM(amount) FILTER (WHERE "verifiedAt" IS NULL) AS unverified
			FROM "ProjectPayment" WHERE "projectId" = p.id AND "cancelledAt" IS NULL) pp ON true
		ORDER BY `+priceValueExpr+` DESC NULLS LAST, COALESCE(pp.paid, 0) DESC, p."createdAt" DESC`)
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

// FillInfo يكمّل معلومات المشروع الناقصة من الإحصائيات — طلب (ع) 10-10:
// «كل المعلومات المبهمة تنكتب لمرة وحدة». الفارغ ما يمسح الموجود.
func (r *ProjectRepository) FillInfo(id string, price, responsibleID *string) error {
	res, err := r.db.Exec(`UPDATE "Project" SET
		price = COALESCE(NULLIF(btrim($2), ''), price),
		"responsibleEmployeeId" = COALESCE(NULLIF($3, ''), "responsibleEmployeeId"),
		"updatedAt" = now()
		WHERE id = $1`, id, price, responsibleID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
