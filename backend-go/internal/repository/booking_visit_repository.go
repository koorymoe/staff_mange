package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// BookingVisitRepository الطلعات — كل مرة طلع بيها كادر على حجز.
//
// «أريد حتى لو الحجز نفسه طلعناله أربع أيام، كل مرة طلعناله تنحسب
// حجز للموظف، وكل مرة ينكتب بيها تاريخ وكادر طلع — لأن يجوز الكادر
// يتغيّر».
type BookingVisitRepository struct {
	db *sqlx.DB
}

func NewBookingVisitRepository(db *sqlx.DB) *BookingVisitRepository {
	return &BookingVisitRepository{db: db}
}

// recordVisitTx يسجّل طلعة وكادرها **بنفس المعاملة** الي غيّرت الحجز.
//
// ⚠️ لازم تكون بنفس المعاملة: لو انقفل يوم الشغل وما انسجّلت طلعته،
// الموظف يشتغل يوم كامل وما ينحسبله — وهذا نفس الظلم الي نصلّحه.
// ولو انعكست، تنسجّل طلعة لشغل ما صار.
//
// الكادر ينلقط من `BookingAssignment` **بلحظته** — لأن الإداري يجوز
// يبدّله للطلعة الجاية، والطلعة لازم تحتفظ بمنو طلع بيها هي.
func recordVisitTx(tx *sqlx.Tx, bookingID, outcome string, percentDone *int, progressReportID *string) (string, error) {
	var visitNumber int
	if err := tx.Get(&visitNumber, `
		SELECT COALESCE(MAX("visitNumber"), 0) + 1 FROM "BookingVisit" WHERE "bookingId" = $1
	`, bookingID); err != nil {
		return "", err
	}

	visitID := uuid.NewString()
	if _, err := tx.Exec(`
		INSERT INTO "BookingVisit" (id, "bookingId", "visitNumber", outcome, "percentDone", "progressReportId", "scheduledAt")
		VALUES ($1, $2, $3, $4, $5, $6, (SELECT "scheduledAt" FROM "Booking" WHERE id = $2))
	`, visitID, bookingID, visitNumber, outcome, percentDone, progressReportID); err != nil {
		return "", err
	}

	// ⚠️ الطلعة بلا كادر مسموحة (حجز خلّصه الليدر بلا تكليف مسجّل):
	// نسجّلها بلا صفوف كادر بدل ما نفشل العملية كلها ونمنع قفل اليوم.
	if _, err := tx.Exec(`
		INSERT INTO "BookingVisitCrew" (id, "visitId", "employeeId", role, "isLeader")
		SELECT gen_random_uuid()::text, $1, a."employeeId", a.role, COALESCE(e."isLeader", false)
		FROM "BookingAssignment" a
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."bookingId" = $2
		ON CONFLICT ("visitId", "employeeId") DO NOTHING
	`, visitID, bookingID); err != nil {
		return "", err
	}
	return visitID, nil
}

// ByBooking كل طلعات حجز — الأقدم أول، لأنها قصة بالترتيب.
func (r *BookingVisitRepository) ByBooking(bookingID string) ([]model.BookingVisit, error) {
	visits := []model.BookingVisit{}
	if err := r.db.Select(&visits, `
		SELECT * FROM "BookingVisit" WHERE "bookingId" = $1 ORDER BY "visitNumber" ASC
	`, bookingID); err != nil {
		return nil, err
	}
	for i := range visits {
		crew := []model.BookingVisitCrewMember{}
		if err := r.db.Select(&crew, `
			SELECT vc."employeeId", vc.role, vc."isLeader", e.name
			FROM "BookingVisitCrew" vc
			JOIN "Employee" e ON e.id = vc."employeeId"
			WHERE vc."visitId" = $1
			ORDER BY vc."isLeader" DESC, vc.role ASC
		`, visits[i].ID); err != nil {
			return nil, err
		}
		visits[i].Crew = crew
	}
	return visits, nil
}

// CountVisitsForEmployee عدد طلعات موظف بمدة — **هذا** هو مقياس
// الإنتاجية الصح: يعدّ الطلعات مو الحجوزات، ويعتمد على كادر كل طلعة
// وقتها مو على التكليف الحالي.
func (r *BookingVisitRepository) CountVisitsForEmployee(employeeID string, sinceDays int) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "BookingVisitCrew" vc
		JOIN "BookingVisit" v ON v.id = vc."visitId"
		WHERE vc."employeeId" = $1
		  AND ($2 <= 0 OR v."occurredAt" >= now() - ($2::text || ' days')::interval)
	`, employeeID, sinceDays)
	return n, err
}
