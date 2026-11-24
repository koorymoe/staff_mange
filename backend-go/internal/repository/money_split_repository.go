package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ توزيع الفلوس (ل) 11-24 دازاي ═══
// «الإحصائيات: الليدر 50٪ والفنيين 50٪ يتقاسمونها؛ الحجز البلا ليدر الفني 30٪؛ مسؤول الخدمة ما ينحسبله شي.
//  الإيراد: الليدر 40٪، الفني 30٪، الشركة 30٪ — اريد اشوف صافي ارباح الشركة. ضيفها بالإعدادات».

type MoneySplit struct {
	StatsLeaderPct       float64    `db:"statsLeaderPct" json:"statsLeaderPct"`
	StatsTechPct         float64    `db:"statsTechPct" json:"statsTechPct"`
	StatsNoLeaderTechPct float64    `db:"statsNoLeaderTechPct" json:"statsNoLeaderTechPct"`
	RevenueLeaderPct     float64    `db:"revenueLeaderPct" json:"revenueLeaderPct"`
	RevenueTechPct       float64    `db:"revenueTechPct" json:"revenueTechPct"`
	RevenueTechEach      bool       `db:"revenueTechEach" json:"revenueTechEach"`
	UpdatedByName        *string    `db:"updatedByName" json:"updatedByName"`
	UpdatedAt            *time.Time `db:"updatedAt" json:"updatedAt"`
}

// DefaultMoneySplit نسب دازاي (11-24) — تنطبق لحد ما يغيّرها من الإعدادات.
var DefaultMoneySplit = MoneySplit{StatsLeaderPct: 50, StatsTechPct: 50, StatsNoLeaderTechPct: 30, RevenueLeaderPct: 40, RevenueTechPct: 30}

// LoadMoneySplit النسب الحالية (الافتراضي إذا بعد ما انحفظ شي).
func LoadMoneySplit(db sqlx.Queryer) MoneySplit {
	m := DefaultMoneySplit
	err := sqlx.Get(db, &m, `SELECT c."statsLeaderPct", c."statsTechPct", c."statsNoLeaderTechPct", c."revenueLeaderPct", c."revenueTechPct",
		c."revenueTechEach", e.name AS "updatedByName", c."updatedAt"
		FROM "MoneySplitConfig" c LEFT JOIN "Employee" e ON e.id = c."updatedById" WHERE c.id = 1`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return DefaultMoneySplit
	}
	return m
}

// SaveMoneySplit يحفظ النسب (صف واحد).
func SaveMoneySplit(db *sqlx.DB, m MoneySplit, by string) error {
	_, err := db.Exec(`INSERT INTO "MoneySplitConfig" (id, "statsLeaderPct", "statsTechPct", "statsNoLeaderTechPct", "revenueLeaderPct", "revenueTechPct", "revenueTechEach", "updatedById", "updatedAt")
		VALUES (1, $1, $2, $3, $4, $5, $6, NULLIF($7, ''), now())
		ON CONFLICT (id) DO UPDATE SET "statsLeaderPct" = $1, "statsTechPct" = $2, "statsNoLeaderTechPct" = $3,
			"revenueLeaderPct" = $4, "revenueTechPct" = $5, "revenueTechEach" = $6, "updatedById" = NULLIF($7, ''), "updatedAt" = now()`,
		m.StatsLeaderPct, m.StatsTechPct, m.StatsNoLeaderTechPct, m.RevenueLeaderPct, m.RevenueTechPct, m.RevenueTechEach, by)
	return err
}

// workCrewSQL كادر كل فاتورة ودور كل واحد (CTE تحتاج inv قبلها): من العمولة نفسها (LEADER/TECHNICIAN)،
// وفاتورة بلا عمولات ← كادر تكليف الحجز والليدر = booking_leader_id. مسؤول الخدمة ما ينحسب بالكادر.
const workCrewSQL = `
		crew AS (
			(SELECT DISTINCT ON (ec."leaderInvoiceId", ec."employeeId") ec."leaderInvoiceId" AS inv, ec."employeeId", (ec.role = 'LEADER') AS lead
			FROM "EmployeeCommission" ec WHERE ec."leaderInvoiceId" IN (SELECT id FROM inv)
			ORDER BY ec."leaderInvoiceId", ec."employeeId", (ec.role = 'LEADER') DESC)
			UNION ALL
			SELECT inv.id, a."employeeId", (a."employeeId" = booking_leader_id(inv."bookingId"))
			FROM inv JOIN "BookingAssignment" a ON a."bookingId" = inv."bookingId"
			WHERE NOT EXISTS (SELECT 1 FROM "EmployeeCommission" ec WHERE ec."leaderInvoiceId" = inv.id)
		), crew2 AS (
			SELECT c.* FROM crew c JOIN "Employee" e ON e.id = c."employeeId" WHERE e.role::text <> 'SERVICE_MANAGER'
		), cnt AS (
			SELECT inv, bool_or(lead) AS has_lead, count(*) FILTER (WHERE NOT lead) AS techs FROM crew2 GROUP BY 1
		)`

// CompanyProfit صافي ربح الشركة بفترة — من صافي فواتير الحجوزات، بنسب «الإيراد»:
// الليدر نسبته (إذا الفاتورة بيها ليدر)، والفنيين نسبتهم (يتقاسمونها، أو لكل فني إذا revenueTechEach)،
// والشركة الباقي. مسؤول الخدمة ما ياخذ شي (حصته للشركة).
type CompanyProfit struct {
	Invoices int     `db:"invoices" json:"invoices"`
	Total    float64 `db:"total" json:"total"`
	Leaders  float64 `db:"leaders" json:"leaders"`
	Techs    float64 `db:"techs" json:"techs"`
	Company  float64 `db:"company" json:"company"`
}

func (r *MatrixScoreRepository) CompanyProfit(from, to time.Time) (CompanyProfit, error) {
	ms := LoadMoneySplit(r.db)
	var p CompanyProfit
	err := r.db.Get(&p, `
		WITH inv AS (
			SELECT li.id, li."bookingId", li."employeeId", li."netTotal" AS n FROM "LeaderInvoice" li
			WHERE li."revokedAt" IS NULL AND li.status <> 'DRAFT' AND li."createdAt" >= $1 AND li."createdAt" < $2
		),`+workCrewSQL+`, per AS (
			SELECT inv.n,
				CASE WHEN COALESCE(cnt.has_lead, false) THEN inv.n * $3::float8 / 100.0 ELSE 0 END AS l,
				CASE WHEN COALESCE(cnt.techs, 0) = 0 THEN 0
				     WHEN $5::bool THEN inv.n * $4::float8 / 100.0 * cnt.techs ELSE inv.n * $4::float8 / 100.0 END AS t
			FROM inv LEFT JOIN cnt ON cnt.inv = inv.id
		)
		SELECT count(*)::int AS invoices, COALESCE(sum(n), 0)::float8 AS total, COALESCE(sum(l), 0)::float8 AS leaders,
			COALESCE(sum(t), 0)::float8 AS techs, COALESCE(sum(n - l - t), 0)::float8 AS company FROM per`,
		from, to, ms.RevenueLeaderPct, ms.RevenueTechPct, ms.RevenueTechEach)
	return p, err
}
