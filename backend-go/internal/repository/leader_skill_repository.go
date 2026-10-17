package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type LeaderSkillRepository struct {
	db *sqlx.DB
}

func NewLeaderSkillRepository(db *sqlx.DB) *LeaderSkillRepository {
	return &LeaderSkillRepository{db: db}
}

// ListForEmployee درجات موظف واحد.
func (r *LeaderSkillRepository) ListForEmployee(employeeID string) ([]model.LeaderSkillRating, error) {
	rows := []model.LeaderSkillRating{}
	err := r.db.Select(&rows, `
		SELECT r.*, e.name AS "ratedByName"
		FROM "LeaderSkillRating" r
		LEFT JOIN "Employee" e ON e.id = r."ratedById"
		WHERE r."employeeId" = $1
		ORDER BY r.skill`, employeeID)
	return rows, err
}

// SetRatings يحفظ الدرجات دفعة وحدة.
//
// ⚠️ UPSERT على (employeeId, skill): التقييم **حالة** مو سجل
// أحداث — «كم درجته بالقيادة الآن؟» جوابها واحد. بلا هذا كل سحبة
// شريط تضيف صفاً وما نعرف أيّهن الحالي.
func (r *LeaderSkillRepository) SetRatings(
	employeeID string, scores map[string]int, ratedByID string,
) error {
	if len(scores) == 0 {
		return nil
	}
	// ⚠️ الفحص هنا **زيادة** على قيد قاعدة البيانات مو بديلاً عنه:
	// القيد هو الي يحمي من نداء مباشر أو مسار ينكتب بكرة.
	for skill, sc := range scores {
		if sc < 1 || sc > 10 {
			return errors.New("الدرجة لازم تكون من ١ إلى ١٠")
		}
		if !model.IsLeaderSkill(skill) {
			return errors.New("مهارة قيادية غير معروفة: " + skill)
		}
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for skill, sc := range scores {
		if _, err := tx.Exec(`
			INSERT INTO "LeaderSkillRating" (id, "employeeId", skill, score, "ratedById", "ratedAt")
			VALUES ($1, $2, $3, $4, $5, now())
			ON CONFLICT ("employeeId", skill) DO UPDATE
			SET score = EXCLUDED.score,
			    "ratedById" = EXCLUDED."ratedById",
			    "ratedAt" = now()`,
			uuid.NewString(), employeeID, skill, sc, ratedByID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
