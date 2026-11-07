package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ ✍️ توقيع الزبون عند التسليم — (ع) 11-07 ═══
// الزبون يوقّع على شاشة الليدر: «استلمت الشغل ومنظومتي تشتغل». يحسم أي خلاف بعدين
// ويحمي الليدر. ⚠️ جدول مستقل (مو داخل sheet التقرير) واسم العمود فيه «signature»
// حتى صلاحيات الذكاء (ai_readers: aiDenyColumn) تحجبه — صورة شخصية للزبون.

type CustomerSignature struct {
	BookingID       string    `db:"bookingId" json:"bookingId"`
	SignerName      string    `db:"signerName" json:"signerName"`
	SignatureImage  string    `db:"signatureImage" json:"image"`
	SignedAt        time.Time `db:"signedAt" json:"signedAt"`
	CollectedByName *string   `db:"collectedByName" json:"collectedByName"`
}

type CustomerSignatureRepository struct{ db *sqlx.DB }

func NewCustomerSignatureRepository(db *sqlx.DB) *CustomerSignatureRepository {
	return &CustomerSignatureRepository{db: db}
}

const maxSignatureBytes = 300 << 10

// CanCollect الموظف من كادر الحجز (أو ليدره)؟
func (r *CustomerSignatureRepository) CanCollect(bookingID, employeeID string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Booking" b WHERE b.id = $1 AND (`+BookingWorkedBySQL("b", "$2")+` OR booking_leader_id(b.id) = $2))`, bookingID, employeeID)
	return ok
}

// Save يحفظ (أو يبدّل) توقيع الحجز.
func (r *CustomerSignatureRepository) Save(bookingID, employeeID, signerName, image string) error {
	signerName = strings.TrimSpace(signerName)
	if len([]rune(signerName)) < 2 || len([]rune(signerName)) > 120 {
		return errors.New("اكتب اسم الي وقّع")
	}
	if !strings.HasPrefix(image, "data:image/png;base64,") || len(image) < 200 {
		return errors.New("التوقيع فاضي")
	}
	if len(image) > maxSignatureBytes {
		return errors.New("التوقيع كبير هواية")
	}
	_, err := r.db.Exec(`INSERT INTO "CustomerSignature" ("bookingId", "signerName", "signatureImage", "collectedById")
		VALUES ($1, $2, $3, $4)
		ON CONFLICT ("bookingId") DO UPDATE SET "signerName" = EXCLUDED."signerName", "signatureImage" = EXCLUDED."signatureImage",
		  "collectedById" = EXCLUDED."collectedById", "signedAt" = now()`, bookingID, signerName, image, employeeID)
	return err
}

// Get توقيع الحجز، أو nil.
func (r *CustomerSignatureRepository) Get(bookingID string) (*CustomerSignature, error) {
	var s CustomerSignature
	err := r.db.Get(&s, `SELECT s."bookingId", s."signerName", s."signatureImage", s."signedAt", e.name AS "collectedByName"
		FROM "CustomerSignature" s LEFT JOIN "Employee" e ON e.id = s."collectedById" WHERE s."bookingId" = $1`, bookingID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}
