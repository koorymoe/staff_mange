package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// InternalWorksRepository إحصائية الأعمال الي انجزت *داخل الشركة* بشهر
// معيّن — شنو انخلص جوه، وشكد انشتغل جوه، ومنو اشتغل.
type InternalWorksRepository struct {
	db *sqlx.DB
}

func NewInternalWorksRepository(db *sqlx.DB) *InternalWorksRepository {
	return &InternalWorksRepository{db: db}
}

// workLocationOrEmpty يرجع القيمة بس إذا كانت معروفة — أي شي غيرها
// يصير فاضي حتى COALESCE(NULLIF(...)) تبقي القديم بدل ما تخرب العمود.
func workLocationOrEmpty(v *string) string {
	if v != nil && model.ValidWorkLocation(*v) {
		return *v
	}
	return ""
}

// Monthly ملخص الشهر (بصيغة YYYY-MM).
//
// نحسب المنجز داخل الشركة مقابل المنجز عند الزبون بنفس الاستعلام، حتى
// النسبة تطلع من نفس المجموعة ولا تحتاج طلب ثاني.
func (r *InternalWorksRepository) Monthly(month string) (*model.InternalWorksReport, error) {
	rep := &model.InternalWorksReport{
		Month:    month,
		Services: []model.InternalWorkServiceRow{},
		Crew:     []model.InternalWorkCrewRow{},
		Works:    []model.InternalWorkRow{},
	}

	if err := r.db.Get(rep, `
		SELECT
			COUNT(*) FILTER (WHERE b."workLocation" = 'IN_HOUSE')                AS "inHouseCount",
			COUNT(*) FILTER (WHERE b."workLocation" <> 'IN_HOUSE')               AS "onSiteCount",
			COALESCE(SUM(b."amountCollected") FILTER (WHERE b."workLocation" = 'IN_HOUSE'), 0) AS "inHouseAmount"
		FROM "Booking" b
		WHERE b.status = 'COMPLETED'
		  AND to_char(b."completedAt", 'YYYY-MM') = $1
	`, month); err != nil {
		return nil, err
	}

	// شنو الأعمال الي انخلصت جوه — مصنّفة بالخدمة
	if err := r.db.Select(&rep.Services, `
		SELECT COALESCE(s.name, 'بلا خدمة') AS name,
		       COUNT(*) AS count,
		       COALESCE(SUM(b."amountCollected"), 0) AS amount
		FROM "Booking" b
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.status = 'COMPLETED'
		  AND b."workLocation" = 'IN_HOUSE'
		  AND to_char(b."completedAt", 'YYYY-MM') = $1
		GROUP BY s.name
		ORDER BY count DESC
	`, month); err != nil {
		return nil, err
	}

	// شكد اشتغل كل واحد جوه — من كادر الحجز نفسه
	if err := r.db.Select(&rep.Crew, `
		SELECT e.name AS "employeeName", COUNT(DISTINCT b.id) AS count
		FROM "Booking" b
		JOIN "BookingAssignment" a ON a."bookingId" = b.id
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE b.status = 'COMPLETED'
		  AND b."workLocation" = 'IN_HOUSE'
		  AND to_char(b."completedAt", 'YYYY-MM') = $1
		GROUP BY e.name
		ORDER BY count DESC
	`, month); err != nil {
		return nil, err
	}

	// تفاصيل الأعمال نفسها — بلا معلومات الزبون، الإحصائية ما تحتاجها
	if err := r.db.Select(&rep.Works, `
		SELECT b.code, b."completedAt", COALESCE(s.name, '—') AS "serviceName",
		       COALESCE(b."amountCollected", 0) AS amount
		FROM "Booking" b
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.status = 'COMPLETED'
		  AND b."workLocation" = 'IN_HOUSE'
		  AND to_char(b."completedAt", 'YYYY-MM') = $1
		ORDER BY b."completedAt" DESC
		LIMIT 500
	`, month); err != nil {
		return nil, err
	}

	return rep, nil
}
