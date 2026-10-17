package repository

import "time"

// ═══ شغل أبو الكميات بعين ماتركس (قرار (ع) 10-10) ═══
// «كل شغلة يشتغلها لازم تكون موجودة بالنظام». البنود من تاريخ التفعيل وطالع —
// الماضي ما ينحسب على أحد لأن القرارات وقتها ما چانت تنسجّل.
const ProcurementWatchStart = "2026-10-10"

// ProcDecision طلب (مواد أو أداة) وقرار أبو الكميات عليه.
type ProcDecision struct {
	ID        string     `db:"id" json:"id"`
	Label     string     `db:"label" json:"label"`
	CreatedAt time.Time  `db:"createdAt" json:"createdAt"`
	DecidedAt *time.Time `db:"decidedAt" json:"decidedAt"`
	DecidedBy *string    `db:"decidedBy" json:"decidedBy"`
	ByName    *string    `db:"byName" json:"byName"`
}

// MaterialDecisions طلبات المواد الي انحسمت، أو بعدها معلّقة. pendingOnly للتذكير.
func (r *MatrixScoreRepository) MaterialDecisions(pendingOnly bool) ([]ProcDecision, error) {
	rows := []ProcDecision{}
	q := `SELECT p.id, COALESCE(e.name, '—') || ' — طلب مواد' AS label, p."createdAt", p."decidedAt", p."decidedById" AS "decidedBy", d.name AS "byName"
		FROM "ProcurementRequest" p LEFT JOIN "Employee" e ON e.id = p."requestedById" LEFT JOIN "Employee" d ON d.id = p."decidedById"
		WHERE p."createdAt" >= $1::date`
	if pendingOnly {
		q += ` AND p.status = 'PENDING'`
	}
	err := r.db.Select(&rows, q+` ORDER BY p."createdAt"`, ProcurementWatchStart)
	return rows, err
}

// ToolDecisions طلبات الأدوات.
func (r *MatrixScoreRepository) ToolDecisions(pendingOnly bool) ([]ProcDecision, error) {
	rows := []ProcDecision{}
	q := `SELECT t.id, COALESCE(e.name, '—') || ' — طلب أداة' AS label, t."requestedAt" AS "createdAt",
		COALESCE(t."approvedAt", t."rejectedAt") AS "decidedAt", COALESCE(t."approvedById", t."rejectedById") AS "decidedBy", d.name AS "byName"
		FROM "ToolRequest" t LEFT JOIN "Employee" e ON e.id = t."employeeId"
		LEFT JOIN "Employee" d ON d.id = COALESCE(t."approvedById", t."rejectedById")
		WHERE t."requestedAt" >= $1::date`
	if pendingOnly {
		q += ` AND t.status = 'PENDING'`
	}
	err := r.db.Select(&rows, q+` ORDER BY t."requestedAt"`, ProcurementWatchStart)
	return rows, err
}

// OpenShortages نواقص الجرد (الي ما انحلت والي انحلت) من تاريخ التفعيل.
func (r *MatrixScoreRepository) Shortages(openOnly bool) ([]ProcDecision, error) {
	rows := []ProcDecision{}
	q := `SELECT c.id, COALESCE(e.name, '—') || ' — نقص: ' || COALESCE(c."missingItems", '') AS label, c."checkedAt" AS "createdAt",
		c."resolvedAt" AS "decidedAt", c."resolvedById" AS "decidedBy", d.name AS "byName"
		FROM "InventoryCheck" c LEFT JOIN "Employee" e ON e.id = c."employeeId" LEFT JOIN "Employee" d ON d.id = c."resolvedById"
		WHERE NOT c.complete AND c."checkedAt" >= $1::date`
	if openOnly {
		q += ` AND NOT c.resolved`
	}
	err := r.db.Select(&rows, q+` ORDER BY c."checkedAt"`, ProcurementWatchStart)
	return rows, err
}

// VehicleDay كم سيارة فعّالة وكم انقيّمت بيوم.
type VehicleDay struct {
	Day   string `db:"day" json:"day"`
	Total int    `db:"total" json:"total"`
	Rated int    `db:"rated" json:"rated"`
}

// VehicleDays أيام متابعة السيارات من تاريخ التفعيل لحد أمس (اليوم بعده ما خلص).
func (r *MatrixScoreRepository) VehicleDays() ([]VehicleDay, error) {
	rows := []VehicleDay{}
	err := r.db.Select(&rows, `
		SELECT to_char(d, 'YYYY-MM-DD') AS day,
		       (SELECT count(*) FROM "Vehicle" v WHERE v."isActive")::int AS total,
		       (SELECT count(DISTINCT x."vehicleId") FROM "VehicleDailyRating" x JOIN "Vehicle" v ON v.id = x."vehicleId" AND v."isActive"
		         WHERE x."ratedDate" = d::date)::int AS rated
		FROM generate_series($1::date, baghdad_today() - 1, interval '1 day') d`, ProcurementWatchStart)
	return rows, err
}

// VehiclesUnratedToday سيارات فعّالة ما انقيّمت اليوم.
func (r *MatrixScoreRepository) VehiclesUnratedToday() ([]string, error) {
	names := []string{}
	err := r.db.Select(&names, `SELECT v.name FROM "Vehicle" v WHERE v."isActive"
		AND NOT EXISTS (SELECT 1 FROM "VehicleDailyRating" x WHERE x."vehicleId" = v.id AND x."ratedDate" = baghdad_today())
		ORDER BY v.name`)
	return names, err
}

// FleetOverdue صيانات فات موعدها (آخر سجل لكل نوع) + وثائق منتهية.
func (r *MatrixScoreRepository) FleetOverdue() ([]string, error) {
	rows := []string{}
	err := r.db.Select(&rows, `
		SELECT v.name || ' — ' || x.type FROM (
			SELECT DISTINCT ON ("vehicleId", type) "vehicleId", type, "nextDueAt"
			FROM "VehicleLog" WHERE "nextDueAt" IS NOT NULL ORDER BY "vehicleId", type, "performedAt" DESC) x
		JOIN "Vehicle" v ON v.id = x."vehicleId" AND v."isActive"
		WHERE x."nextDueAt" < now()
		UNION ALL
		SELECT v.name || ' — وثيقة ' || COALESCE(d."documentType", '') || ' منتهية' FROM "VehicleDocument" d
		JOIN "Vehicle" v ON v.id = d."vehicleId" AND v."isActive"
		WHERE d."expiryDate" IS NOT NULL AND d."expiryDate" < now()`)
	return rows, err
}
