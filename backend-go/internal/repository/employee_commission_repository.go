package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// EmployeeCommissionRepository يخزّن عمولات الليدر/الفنيين المحسوبة تلقائياً
// عند إنشاء كل فاتورة ليدر.
type EmployeeCommissionRepository struct {
	db *sqlx.DB
}

func NewEmployeeCommissionRepository(db *sqlx.DB) *EmployeeCommissionRepository {
	return &EmployeeCommissionRepository{db: db}
}

// Create يحفظ صف عمولة واحد (ليدر أو فني) لفاتورة ليدر معيّنة.
func (r *EmployeeCommissionRepository) Create(employeeID, leaderInvoiceID, role string, executionCommission, salesCommission float64) (*model.EmployeeCommission, error) {
	var c model.EmployeeCommission
	err := r.db.Get(&c, `
		INSERT INTO "EmployeeCommission" (
			id, "employeeId", "leaderInvoiceId", role, "executionCommission", "salesCommission", "totalCommission"
		) VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, employeeID, leaderInvoiceID, role, executionCommission, salesCommission, executionCommission+salesCommission)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListByLeaderInvoice يرجّع كل عمولات فاتورة ليدر واحدة.
func (r *EmployeeCommissionRepository) ListByLeaderInvoice(leaderInvoiceID string) ([]model.EmployeeCommission, error) {
	rows := []model.EmployeeCommission{}
	err := r.db.Select(&rows, `SELECT * FROM "EmployeeCommission" WHERE "leaderInvoiceId" = $1 ORDER BY "createdAt"`, leaderInvoiceID)
	return rows, err
}

// SumForEmployeeMonth يرجّع مجموع "totalCommission" لموظف معيّن خلال شهر معيّن
// (monthPrefix بصيغة "YYYY-MM").
func (r *EmployeeCommissionRepository) SumForEmployeeMonth(employeeID, monthPrefix string) (float64, error) {
	var total sql64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM("totalCommission"), 0) FROM "EmployeeCommission"
		WHERE "employeeId" = $1 AND to_char("createdAt", 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return float64(total), err
}

// SumForEmployeeRange نفس SumForEmployeeMonth لكن لمدى تاريخ حر (from/to بصيغة
// "YYYY-MM-DD").
func (r *EmployeeCommissionRepository) SumForEmployeeRange(employeeID, from, to string) (float64, error) {
	var total sql64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM("totalCommission"), 0) FROM "EmployeeCommission"
		WHERE "employeeId" = $1 AND baghdad_date("createdAt") BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return float64(total), err
}

// MonthlyCommissionSeriesForEmployee يرجّع مجموع "totalCommission" لكل شهر من
// monthsCount شهر تنتهي بشهر endMonth (بصيغة "YYYY-MM-01"، بالترتيب من الأقدم
// للأحدث) — يُستخدم بمنحنى الأداء المتحرك بصفحة إحصائيات الموظفين.
func (r *EmployeeCommissionRepository) MonthlyCommissionSeriesForEmployee(employeeID string, monthsCount int, endMonth string) ([]model.MonthlyCommissionBucket, error) {
	buckets := []model.MonthlyCommissionBucket{}
	err := r.db.Select(&buckets, `
		SELECT to_char(m, 'YYYY-MM') AS month, COALESCE(SUM(c."totalCommission"), 0) AS amount
		FROM generate_series(
			$3::date - ($2 - 1) * interval '1 month',
			$3::date,
			interval '1 month'
		) m
		LEFT JOIN "EmployeeCommission" c ON c."employeeId" = $1
			AND to_char(c."createdAt", 'YYYY-MM') = to_char(m, 'YYYY-MM')
		GROUP BY m
		ORDER BY m
	`, employeeID, monthsCount, endMonth)
	return buckets, err
}

// SumForEmployeeLast7Days يرجّع مجموع "totalCommission" لموظف معيّن خلال آخر 7
// أيام — يُستخدم لحساب "حجم المبيعات" الأسبوعي.
func (r *EmployeeCommissionRepository) SumForEmployeeLast7Days(employeeID string) (float64, error) {
	var total sql64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM("totalCommission"), 0) FROM "EmployeeCommission"
		WHERE "employeeId" = $1 AND "createdAt" >= now() - interval '7 days'
	`, employeeID)
	return float64(total), err
}

// SumForDate يرجّع مجموع "totalCommission" لكل الموظفين بتاريخ معيّن —
// يُستخدم كـ"إجمالي الأرباح" اليومي.
func (r *EmployeeCommissionRepository) SumForDate(date string) (float64, error) {
	var total sql64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM("totalCommission"), 0) FROM "EmployeeCommission" WHERE baghdad_date("createdAt") = $1::date
	`, date)
	return float64(total), err
}

type sql64 float64
