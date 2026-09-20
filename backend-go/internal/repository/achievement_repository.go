package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// AchievementRepository الإنجازات — تقرير يومي حر من أي موظف بأي دور.
type AchievementRepository struct {
	db *sqlx.DB
}

func NewAchievementRepository(db *sqlx.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func (r *AchievementRepository) hydrate(a *model.Achievement) {
	var name string
	if err := r.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, a.EmployeeID); err == nil {
		a.EmployeeName = &name
	}
	if a.BookingID != nil {
		var code string
		if err := r.db.Get(&code, `SELECT code FROM "Booking" WHERE id = $1`, *a.BookingID); err == nil {
			a.BookingCode = &code
		}
	}
	if a.ReviewedByID != nil {
		var name string
		if err := r.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, *a.ReviewedByID); err == nil {
			a.ReviewedByName = &name
		}
	}
}

func (r *AchievementRepository) Create(employeeID, role string, req model.CreateAchievementRequest) (*model.Achievement, error) {
	var a model.Achievement
	err := r.db.Get(&a, `
		INSERT INTO "Achievement" (id, "employeeId", role, "bookingId", "reportText")
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *
	`, uuid.NewString(), employeeID, role, req.BookingID, req.ReportText)
	if err != nil {
		return nil, err
	}
	r.hydrate(&a)
	return &a, nil
}

// List يرجّع الإنجازات — بفلترة اختيارية بموظف و/أو يوم (YYYY-MM-DD).
func (r *AchievementRepository) List(employeeID, day string, limit int) ([]model.Achievement, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows := []model.Achievement{}
	q := `SELECT * FROM "Achievement" WHERE 1=1`
	args := []any{}
	if employeeID != "" {
		args = append(args, employeeID)
		q += ` AND "employeeId" = $` + itoa(len(args))
	}
	if day != "" {
		args = append(args, day)
		q += ` AND "createdAt"::date = $` + itoa(len(args))
	}
	args = append(args, limit)
	q += ` ORDER BY "createdAt" DESC LIMIT $` + itoa(len(args))
	if err := r.db.Select(&rows, q, args...); err != nil {
		return nil, err
	}
	for i := range rows {
		r.hydrate(&rows[i])
	}
	return rows, nil
}

func (r *AchievementRepository) FindByID(id string) (*model.Achievement, error) {
	var a model.Achievement
	if err := r.db.Get(&a, `SELECT * FROM "Achievement" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate(&a)
	return &a, nil
}

func (r *AchievementRepository) Review(id, reviewerID, status, note string) (*model.Achievement, error) {
	var a model.Achievement
	err := r.db.Get(&a, `
		UPDATE "Achievement" SET
			"reviewStatus" = $2, "reviewNote" = NULLIF($3, ''),
			"reviewedById" = $4, "reviewedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, status, note, reviewerID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&a)
	return &a, nil
}

// RoleCount عدد تقارير دور معيّن — مادة الملخص اليومي.
type RoleCount struct {
	Role  string `db:"role"`
	Count int    `db:"count"`
}

// CountsForDay: كم تقرير انرفع اليوم، ومنو الأدوار الي رفعت، مبوّبة.
// نفس أساس ملخص "الفضفضة اليومية" — بلا جدول ثاني.
func (r *AchievementRepository) CountsForDay(day time.Time) (total int, byRole []RoleCount, err error) {
	dayStr := day.Format("2006-01-02")
	if err = r.db.Get(&total, `SELECT COUNT(*) FROM "Achievement" WHERE "createdAt"::date = $1`, dayStr); err != nil {
		return 0, nil, err
	}
	byRole = []RoleCount{}
	err = r.db.Select(&byRole, `
		SELECT role, COUNT(*) AS count FROM "Achievement"
		WHERE "createdAt"::date = $1
		GROUP BY role ORDER BY count DESC`, dayStr)
	return total, byRole, err
}
