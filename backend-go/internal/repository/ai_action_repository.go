package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// AiActionRepository سجل أفعال ماتركس + استعلامات صندوق القرارات.
type AiActionRepository struct {
	db *sqlx.DB
}

func NewAiActionRepository(db *sqlx.DB) *AiActionRepository {
	return &AiActionRepository{db: db}
}

// aiActionSuppressDays المدير رفض فعل؟ نفس النوع على نفس الشي ما
// يرجع يتكرر هالمدة — «لا تذكّر بهذا» قرار، مو إلغاء لمرة وحدة.
const aiActionSuppressDays = 30

// Claim يحجز فعلاً قبل تنفيذه. يرجّع false إذا نفس الفعل انسوّى بنفس
// الفترة، أو المدير رفضه خلال آخر ٣٠ يوم — وبالحالتين ما ننفّذ.
func (r *AiActionRepository) Claim(a model.AiAction) (bool, error) {
	var suppressed bool
	if err := r.db.Get(&suppressed, `
		SELECT EXISTS (SELECT 1 FROM "AiAction"
		  WHERE kind = $1 AND "entityId" = $2 AND status = 'UNDONE'
		    AND "undoneAt" > now() - make_interval(days => $3))`,
		a.Kind, a.EntityID, aiActionSuppressDays); err != nil {
		return false, err
	}
	if suppressed {
		return false, nil
	}
	res, err := r.db.Exec(`
		INSERT INTO "AiAction" (id, kind, "entityType", "entityId", period, "targetEmployeeId", "targetLabel", summary)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (kind, "entityId", period) DO NOTHING`,
		uuid.NewString(), a.Kind, a.EntityType, a.EntityID, a.Period, a.TargetEmployeeID, a.TargetLabel, a.Summary)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

// CountSince عدد أفعال نوع معيّن من وقت — للحد اليومي.
func (r *AiActionRepository) CountSince(kind string, since time.Time) (int, error) {
	var n int
	err := r.db.Get(&n, `SELECT COUNT(*) FROM "AiAction" WHERE kind = $1 AND "createdAt" >= $2`, kind, since)
	return n, err
}

// ListBetween أفعال فترة، الأحدث أول.
func (r *AiActionRepository) ListBetween(from, to time.Time) ([]model.AiAction, error) {
	rows := []model.AiAction{}
	err := r.db.Select(&rows, `
		SELECT * FROM "AiAction" WHERE "createdAt" >= $1 AND "createdAt" < $2
		ORDER BY "createdAt" DESC LIMIT 300`, from, to)
	return rows, err
}

// Undo المدير يرفض فعلاً.
func (r *AiActionRepository) Undo(id, byEmployeeID string) error {
	res, err := r.db.Exec(`
		UPDATE "AiAction" SET status = 'UNDONE', "undoneById" = $2, "undoneAt" = now()
		WHERE id = $1 AND status = 'DONE'`, id, byEmployeeID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("الفعل مرفوض من قبل أو مو موجود")
	}
	return nil
}

// PendingAiVerdicts أحكام ماتركس بصندوق المراقب الي بعدها ما انراجعت.
func (r *AiActionRepository) PendingAiVerdicts(limit int) ([]model.PendingAiDecision, error) {
	rows := []model.PendingAiDecision{}
	err := r.db.Select(&rows, `
		SELECT id, "entityType", "entityId", title, summary, "createdAt"
		FROM "MonitorReview"
		WHERE stage = 'AI_VERDICT' AND status = 'PENDING'
		ORDER BY "createdAt" DESC LIMIT $1`, limit)
	return rows, err
}

// UnstaffedOn حجوزات مثبّتة بيوم معيّن (بغداد) وما عليها أي كادر.
func (r *AiActionRepository) UnstaffedOn(day string) ([]model.UnstaffedBooking, error) {
	rows := []model.UnstaffedBooking{}
	err := r.db.Select(&rows, `
		SELECT b.id, b.code, b."scheduledAt"
		FROM "Booking" b
		WHERE b.status = 'CONFIRMED' AND b."scheduledAt" IS NOT NULL
		  AND baghdad_date(b."scheduledAt") = $1::date
		  AND NOT EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id)`+
		BookingCountableAndSQL(`b`)+`
		ORDER BY b."scheduledAt"`, day)
	return rows, err
}

// EmployeeName اسم موظف للعرض — فاضي إذا ماكو.
func (r *AiActionRepository) EmployeeName(id string) string {
	var name string
	_ = r.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, id)
	return name
}
