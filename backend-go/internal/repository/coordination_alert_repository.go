package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type CoordinationAlertRepository struct {
	db *sqlx.DB
}

func NewCoordinationAlertRepository(db *sqlx.DB) *CoordinationAlertRepository {
	return &CoordinationAlertRepository{db: db}
}

// Add يسجّل تنبيهاً ويرجّع عدد المفتوحة بعده.
//
// ⚠️ العدّ يرجع من نفس المعاملة: المستدعي يقرر النشر على أساسه،
// ولو عدّ بطلب منفصل ممكن ضغطتان متزامنتان تشوفان نفس الرقم
// وتنشران الإعلان مرتين.
func (r *CoordinationAlertRepository) Add(
	bookingID string, coordinatorID, coordinatorName *string,
	reason, byID, byName string,
) (openCount int, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err = tx.Exec(`
		INSERT INTO "CoordinationAlert"
			(id, "bookingId", "coordinatorId", "coordinatorName", reason, "byEmployeeId", "byName")
		VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), $6, NULLIF($7,''))`,
		uuid.NewString(), bookingID, coordinatorID, deref(coordinatorName),
		reason, byID, byName); err != nil {
		return 0, err
	}
	if err = tx.Get(&openCount, `
		SELECT COUNT(*) FROM "CoordinationAlert"
		WHERE "bookingId" = $1 AND "resolvedAt" IS NULL`, bookingID); err != nil {
		return 0, err
	}
	return openCount, tx.Commit()
}

// Resolve يأشّر تنبيهات الحجز المفتوحة «انعالجت».
//
// ⚠️ ما يمحيهن: المعالجة واقعة تنضاف للسجل مو ممحاة له — وإلا
// يقدر أحد ينظّف تاريخه بضغطة.
func (r *CoordinationAlertRepository) Resolve(bookingID, byID, byName, note string) error {
	_, err := r.db.Exec(`
		UPDATE "CoordinationAlert"
		SET "resolvedAt" = now(), "resolvedById" = $2,
		    "resolvedByName" = NULLIF($3,''), "resolveNote" = NULLIF($4,'')
		WHERE "bookingId" = $1 AND "resolvedAt" IS NULL`, bookingID, byID, byName, note)
	return err
}

// ListForBooking سجل الحجز — الأحدث أولاً.
func (r *CoordinationAlertRepository) ListForBooking(bookingID string) ([]model.CoordinationAlert, error) {
	rows := []model.CoordinationAlert{}
	err := r.db.Select(&rows, `
		SELECT * FROM "CoordinationAlert"
		WHERE "bookingId" = $1 ORDER BY "createdAt" DESC`, bookingID)
	return rows, err
}

// Summaries ملخص كل الحجوزات الي عندها تنبيهات — دفعة وحدة.
//
// ⚠️ استعلام واحد مو واحد لكل حجز: الشاشة تعرض عشرات الحجوزات،
// وطلب لكل صف يخليها تزحف.
func (r *CoordinationAlertRepository) Summaries() ([]model.CoordinationAlertSummary, error) {
	rows := []model.CoordinationAlertSummary{}
	err := r.db.Select(&rows, `
		SELECT "bookingId",
		       COUNT(*) FILTER (WHERE "resolvedAt" IS NULL) AS "openCount",
		       COUNT(*)                                     AS "totalCount",
		       MAX("createdAt")                             AS "lastAlertAt"
		FROM "CoordinationAlert"
		GROUP BY "bookingId"`)
	return rows, err
}
