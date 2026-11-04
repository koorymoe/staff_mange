package repository

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"

	"staffmange-api/internal/model"
)

type WorkReportRepository struct {
	db *sqlx.DB
}

func NewWorkReportRepository(db *sqlx.DB) *WorkReportRepository {
	return &WorkReportRepository{db: db}
}

func (r *WorkReportRepository) hydrate(wr *model.WorkReport) {
	var emp model.EmployeeBrief
	if err := r.db.Get(&emp, `SELECT id, name FROM "Employee" WHERE id = $1`, wr.EmployeeID); err == nil {
		wr.Employee = &emp
	}
	var booking model.WorkReportBookingBrief
	err := r.db.Get(&booking, `
		SELECT b.id, b.code, c.name AS "customerName", c.phone AS "customerPhone",
		       c."customerCode", b.address,
		       -- الليدر: المشرف المعيّن، وإلا أول مكلّف مؤشّر «تيم ليدر».
		       -- COALESCE ما تنفع هنا لأن الثاني يحتاج استعلام فرعي.
		       (SELECT e2.name FROM "Employee" e2 WHERE e2.id = booking_leader_id(b.id)) AS "leaderName" -- (ع 10-16) التعريف الموحّد
		FROM "Booking" b
		JOIN "Customer" c ON c.id = b."customerId"
		WHERE b.id = $1
	`, wr.BookingID)
	if err == nil {
		wr.Booking = &booking
	}
	if wr.ReviewedByID != nil {
		var n string
		if err := r.db.Get(&n, `SELECT name FROM "Employee" WHERE id = $1`, *wr.ReviewedByID); err == nil {
			wr.ReviewedByName = &n
		}
	}
}

// Review «تم التدقيق» على تقرير حجز — مرة وحدة، والملاحظة اختيارية.
func (r *WorkReportRepository) Review(id, byID, note string) error {
	res, err := r.db.Exec(`UPDATE "WorkReport" SET "reviewedAt" = now(), "reviewedById" = $2, "reviewNote" = NULLIF($3,'')
		WHERE id = $1 AND "reviewedAt" IS NULL`, id, byID, note)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("التقرير متدقّق من قبل")
	}
	return nil
}

func (r *WorkReportRepository) Create(employeeID string, req model.CreateWorkReportRequest) (*model.WorkReport, error) {
	var wr model.WorkReport
	err := r.db.Get(&wr, `
		INSERT INTO "WorkReport" (id, "bookingId", "employeeId", "workStatus", events, "extraRequests", "cleanedSite", "gaveInfo", "tookPhotos", "stopReason", notes, sheet)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING *
	`, req.BookingID, employeeID, req.WorkStatus, req.Events, req.ExtraRequests, req.CleanedSite, req.GaveInfo, req.TookPhotos, req.StopReason, req.Notes, req.Sheet)
	if err != nil {
		return nil, err
	}
	r.hydrate(&wr)
	return &wr, nil
}

func (r *WorkReportRepository) List(employeeID string) ([]model.WorkReport, error) {
	return r.ListByReview(employeeID, "ALL")
}

// ListByReview (دازاي 10-28: «دققت ٢ اختفن، حدثت رجعن ٢٠٠»): العرض لحد ٢٠٠،
// فالفلترة بالخادم مو بالواجهة — حتى تبويب «تنتظر» يجيب الي تنتظر فعلاً.
// review: PENDING (الأقدم أول حتى تخلص بالترتيب) · DONE (الأحدث أول) · ALL.
func (r *WorkReportRepository) ListByReview(employeeID, review string) ([]model.WorkReport, error) {
	reports := []model.WorkReport{}
	query := `SELECT * FROM "WorkReport" WHERE 1=1`
	args := []any{}
	if employeeID != "" {
		args = append(args, employeeID)
		query += ` AND "employeeId" = $1`
	}
	switch review {
	case "PENDING":
		query += ` AND "reviewedAt" IS NULL ORDER BY "createdAt" ASC LIMIT 200`
	case "DONE":
		query += ` AND "reviewedAt" IS NOT NULL ORDER BY "reviewedAt" DESC LIMIT 200`
	default:
		// الي ينتظر التدقيق أول — حتى ما يضيع ورا الحد (٢٠٠)
		query += ` ORDER BY ("reviewedAt" IS NULL) DESC, "createdAt" DESC LIMIT 200`
	}
	if err := r.db.Select(&reports, query, args...); err != nil {
		return nil, err
	}
	for i := range reports {
		r.hydrate(&reports[i])
	}
	return reports, nil
}

// ReviewCounts العدد الحقيقي (مو المعروض): ينتظر التدقيق ومتدقق.
func (r *WorkReportRepository) ReviewCounts(employeeID string) (pending, done int, err error) {
	var c struct {
		Pending int `db:"pending"`
		Done    int `db:"done"`
	}
	err = r.db.Get(&c, `SELECT count(*) FILTER (WHERE "reviewedAt" IS NULL) AS pending, count(*) FILTER (WHERE "reviewedAt" IS NOT NULL) AS done
		FROM "WorkReport" WHERE $1 = '' OR "employeeId" = $1`, employeeID)
	return c.Pending, c.Done, err
}

// WorkReportBooking الحجز بنتيجة البحث بكوده.
type WorkReportBooking struct {
	ID          string     `db:"id" json:"id"`
	Code        string     `db:"code" json:"code"`
	Status      string     `db:"status" json:"status"`
	CompletedAt *time.Time `db:"completedAt" json:"completedAt"`
	LeaderName  *string    `db:"leaderName" json:"leaderName"`
}

// WorkReportBookingLookup نتيجة البحث بكود الحجز (دازاي 10-31: «تقرير B964 ماموجود —
// شلون أعرف الموظف دز تقرير لو لا؟»): الحجز نفسه، ومسؤول ورقه، وكل تقاريره بلا حد الـ٢٠٠.
type WorkReportBookingLookup struct {
	Booking *WorkReportBooking `json:"booking"`
	Reports []model.WorkReport `json:"reports"`
}

// ByBookingCode تقارير حجز بكوده (B964 أو 964). employeeID فاضي = الكل، وإلا تقاريره هو بس.
func (r *WorkReportRepository) ByBookingCode(code, employeeID string) (*WorkReportBookingLookup, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code != "" && code[0] >= '0' && code[0] <= '9' {
		code = "B" + code
	}
	out := &WorkReportBookingLookup{Reports: []model.WorkReport{}}
	var b WorkReportBooking
	err := r.db.Get(&b, `SELECT b.id, b.code, b.status::text AS status, b."completedAt", e.name AS "leaderName"
		FROM "Booking" b LEFT JOIN "Employee" e ON e.id = booking_leader_id(b.id)
		WHERE upper(b.code) = $1 LIMIT 1`, code)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	out.Booking = &b
	q := `SELECT * FROM "WorkReport" WHERE "bookingId" = $1`
	args := []any{b.ID}
	if employeeID != "" {
		q += ` AND "employeeId" = $2`
		args = append(args, employeeID)
	}
	if err := r.db.Select(&out.Reports, q+` ORDER BY "createdAt" DESC`, args...); err != nil {
		return nil, err
	}
	for i := range out.Reports {
		r.hydrate(&out.Reports[i])
	}
	return out, nil
}
