package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// EmployeeCharacterRepository شخصية الكيان لكل موظف.
type EmployeeCharacterRepository struct {
	db *sqlx.DB
}

func NewEmployeeCharacterRepository(db *sqlx.DB) *EmployeeCharacterRepository {
	return &EmployeeCharacterRepository{db: db}
}

// FindByEmployee يرجّع شخصية الموظف، وnil لو ماكو (مو خطأ — أغلب
// الموظفين ما انولدت شخصيتهم بعد، والكيان يشتغل بإيموجيهم لحد ما تنولد).
func (r *EmployeeCharacterRepository) FindByEmployee(employeeID string) (*model.EmployeeCharacter, error) {
	var row model.EmployeeCharacter
	err := r.db.Get(&row, `SELECT * FROM "EmployeeCharacter" WHERE "employeeId" = $1`, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// MarkPending يحجز صف الشخصية بحالة «قيد التوليد».
//
// ⚠️ UPSERT مو INSERT: إعادة التوليد لنفس الموظف لازم تشتغل، والمفتاح
// الفريد على employeeId يمنع صفّين. والصور القديمة تبقى بمكانها لحد
// ما تنجح الجديدة — الموظف ما يفقد كيانه أثناء التوليد.
func (r *EmployeeCharacterRepository) MarkPending(employeeID, byEmployeeID string) error {
	_, err := r.db.Exec(`
		INSERT INTO "EmployeeCharacter" (id, "employeeId", status, "generatedById", "updatedAt")
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT ("employeeId") DO UPDATE
		SET status = EXCLUDED.status,
		    "generatedById" = EXCLUDED."generatedById",
		    error = NULL,
		    "updatedAt" = now()
	`, uuid.NewString(), employeeID, model.CharacterPending, byEmployeeID)
	return err
}

// SaveReady يثبّت نتيجة توليد ناجحة.
func (r *EmployeeCharacterRepository) SaveReady(employeeID, persona, prompt, calmKey, happyKey, angryKey string) error {
	now := time.Now()
	_, err := r.db.Exec(`
		UPDATE "EmployeeCharacter"
		SET persona = $2, prompt = $3,
		    "calmKey" = $4, "happyKey" = $5, "angryKey" = $6,
		    status = $7, error = NULL,
		    "generatedAt" = $8, "updatedAt" = now()
		WHERE "employeeId" = $1
	`, employeeID, persona, prompt, calmKey, happyKey, angryKey, model.CharacterReady, now)
	return err
}

// MarkFailed يسجّل سبب الفشل — الكيان يبقى شغّال بالإيموجي، بس
// المدير لازم يعرف ليش ما انولدت حتى يعيد المحاولة.
func (r *EmployeeCharacterRepository) MarkFailed(employeeID, reason string) error {
	_, err := r.db.Exec(`
		UPDATE "EmployeeCharacter"
		SET status = $2, error = $3, "updatedAt" = now()
		WHERE "employeeId" = $1
	`, employeeID, model.CharacterFailed, reason)
	return err
}
