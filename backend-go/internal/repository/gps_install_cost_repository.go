package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// GpsInstallCostRepository حساب تكاليف الشد — شهرياً ولكل موظف.
//
// نحسبه حي من بيانات الشد نفسها (اسم المنفّذ + التكلفة + تاريخ الشد) بدل
// جدول ثابت منسوخ من الإكسل، حتى يتحدّث لحاله مع كل عملية شد جديدة.
type GpsInstallCostRepository struct {
	db *sqlx.DB
}

func NewGpsInstallCostRepository(db *sqlx.DB) *GpsInstallCostRepository {
	return &GpsInstallCostRepository{db: db}
}

// Summary تكاليف الشد مجمّعة بالشهر وبالموظف.
//
// التكلفة واسم المنفّذ مخزونين بملاحظات الطلب («منفّذ الشد: فلان · تكلفة
// الشد: 5000») لأن بيانات الإكسل ما كان بيها معرّفات موظفين ولا عمود تكلفة
// بجدول الطلبات — فنستخرج الاثنين من الملاحظة.
func (r *GpsInstallCostRepository) Summary() (*model.GpsInstallCostSummary, error) {
	rows := []model.GpsInstallCostRow{}
	err := r.db.Select(&rows, `
		WITH installs AS (
			SELECT
				to_char(COALESCE(d."subscriptionStart", d."createdAt"), 'YYYY-MM') AS month,
				NULLIF(btrim(substring(d.notes from 'منفّذ الشد: ([^·]+)')), '')    AS "employeeName",
				COALESCE(NULLIF(btrim(substring(d.notes from 'تكلفة الشد: ([0-9.]+)')), '')::numeric, 0) AS cost
			FROM "GpsDeviceRequest" d
			WHERE d.notes LIKE '%منفّذ الشد:%'
		)
		SELECT month, COALESCE("employeeName", 'غير محدد') AS "employeeName",
			COUNT(*) AS installs, SUM(cost) AS total
		FROM installs
		GROUP BY month, "employeeName"
		ORDER BY month DESC, total DESC`)
	if err != nil {
		return nil, err
	}

	sum := &model.GpsInstallCostSummary{Rows: rows}
	months := map[string]bool{}
	people := map[string]float64{}
	for _, row := range rows {
		sum.GrandTotal += row.Total
		sum.TotalInstalls += row.Installs
		months[row.Month] = true
		people[row.EmployeeName] += row.Total
	}
	for name, total := range people {
		sum.ByEmployee = append(sum.ByEmployee, model.GpsInstallCostByEmployee{
			EmployeeName: name, Total: total,
		})
	}
	sum.MonthCount = len(months)
	return sum, nil
}
