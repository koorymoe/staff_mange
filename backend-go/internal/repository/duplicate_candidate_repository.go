package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type DuplicateCandidateRepository struct {
	db *sqlx.DB
}

func NewDuplicateCandidateRepository(db *sqlx.DB) *DuplicateCandidateRepository {
	return &DuplicateCandidateRepository{db: db}
}

// ScanBookings يدوّر على حجزين لنفس الزبون بنفس العنوان بفارق يوم
// واحد أو أقل — تكرار غلط محتمل (مثلاً ضغط "احفظ" مرتين).
//
// ⚠️ يستثني الحجوزات المؤرشفة والمربوطة أصلاً (partialJobBookingId
// بالاتجاهين): هذولا شغلة مفسَّرة أصلاً (يوم إضافي بشغلة، أو كشف
// سبق حجزاً حقيقياً) — مو تكرار غلط، وتأشيرها هنا يزاحم التكرار
// الحقيقي بضجيج مألوف.
func (r *DuplicateCandidateRepository) ScanBookings() (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO "DuplicateCandidate" (id, kind, "entityAId", "entityBId", "matchReason")
		SELECT gen_random_uuid()::text, 'BOOKING', b1.id, b2.id,
		       'نفس الزبون ونفس العنوان بفارق يوم واحد أو أقل'
		FROM "Booking" b1
		JOIN "Booking" b2 ON b2."customerId" = b1."customerId" AND b2.id > b1.id
		WHERE b1."archivedAt" IS NULL AND b2."archivedAt" IS NULL
		  AND b1."partialJobBookingId" IS NULL AND b2."partialJobBookingId" IS NULL
		  AND NOT EXISTS (SELECT 1 FROM "Booking" x WHERE x."partialJobBookingId" = b1.id)
		  AND NOT EXISTS (SELECT 1 FROM "Booking" x WHERE x."partialJobBookingId" = b2.id)
		  AND b1.address IS NOT NULL AND btrim(b1.address) <> ''
		  AND ar_norm(b1.address) = ar_norm(b2.address)
		  AND ABS(EXTRACT(EPOCH FROM (
		        COALESCE(b1."scheduledAt", b1."createdAt") - COALESCE(b2."scheduledAt", b2."createdAt")
		      ))) <= 86400
		ON CONFLICT (kind, "entityAId", "entityBId") DO NOTHING
	`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// ScanCustomers يدوّر على زبونين بنفس رقم الهاتف (بعد التطبيع) بس
// بأسماء مختلفة — نفس الزبون انسجّل مرتين بغلط.
func (r *DuplicateCandidateRepository) ScanCustomers() (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO "DuplicateCandidate" (id, kind, "entityAId", "entityBId", "matchReason")
		SELECT gen_random_uuid()::text, 'CUSTOMER', c1.id, c2.id,
		       'نفس رقم الهاتف (بعد التطبيع) بأسماء مختلفة'
		FROM "Customer" c1
		JOIN "Customer" c2 ON c2.id > c1.id
		  AND phone_norm(c2.phone) = phone_norm(c1.phone)
		WHERE phone_norm(c1.phone) <> ''
		  AND ar_norm(c1.name) <> ar_norm(c2.name)
		ON CONFLICT (kind, "entityAId", "entityBId") DO NOTHING
	`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *DuplicateCandidateRepository) hydrate(c *model.DuplicateCandidate) {
	if c.ReviewedByID != nil {
		var name string
		if err := r.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, *c.ReviewedByID); err == nil {
			c.ReviewedByName = &name
		}
	}
	switch c.Kind {
	case model.DuplicateKindBooking:
		c.BookingA = r.bookingBrief(c.EntityAID)
		c.BookingB = r.bookingBrief(c.EntityBID)
	case model.DuplicateKindCustomer:
		c.CustomerA = r.customerBrief(c.EntityAID)
		c.CustomerB = r.customerBrief(c.EntityBID)
	}
}

func (r *DuplicateCandidateRepository) bookingBrief(id string) *model.DuplicateBookingBrief {
	var b model.DuplicateBookingBrief
	err := r.db.Get(&b, `
		SELECT b.id, b.code, c.name AS "customerName", c.phone AS "customerPhone",
		       c."customerCode", b.address, s.name AS "serviceName", b."scheduledAt"
		FROM "Booking" b
		JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.id = $1
	`, id)
	if err != nil {
		return nil
	}
	return &b
}

func (r *DuplicateCandidateRepository) customerBrief(id string) *model.DuplicateCustomerBrief {
	var c model.DuplicateCustomerBrief
	if err := r.db.Get(&c, `SELECT id, name, phone, "customerCode" FROM "Customer" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &c
}

// List يرجّع الأزواج المشتبه بها — kind/status فاضي يعني "الكل".
func (r *DuplicateCandidateRepository) List(kind, status string) ([]model.DuplicateCandidate, error) {
	query := `SELECT * FROM "DuplicateCandidate" WHERE 1=1`
	args := []any{}
	if kind != "" {
		args = append(args, kind)
		query += fmt.Sprintf(` AND kind = $%d`, len(args))
	}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(` AND status = $%d`, len(args))
	}
	query += ` ORDER BY "detectedAt" DESC LIMIT 300`

	rows := []model.DuplicateCandidate{}
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	for i := range rows {
		r.hydrate(&rows[i])
	}
	return rows, nil
}

// Dismiss يأشّر الزوج "مو تكرار" — قرار إنسان، ما يحذف ولا يدمج شي.
func (r *DuplicateCandidateRepository) Dismiss(id, byEmployeeID string) error {
	_, err := r.db.Exec(`
		UPDATE "DuplicateCandidate"
		SET status = 'DISMISSED', "reviewedById" = $2, "reviewedAt" = now()
		WHERE id = $1
	`, id, byEmployeeID)
	return err
}

// Get زوج واحد.
func (r *DuplicateCandidateRepository) Get(id string) (*model.DuplicateCandidate, error) {
	var c model.DuplicateCandidate
	if err := r.db.Get(&c, `SELECT * FROM "DuplicateCandidate" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
}

// Resolve يسجّل الحل على الزوج.
func (r *DuplicateCandidateRepository) Resolve(id, byEmployeeID, resolution, note string) error {
	_, err := r.db.Exec(`UPDATE "DuplicateCandidate"
		SET status = 'RESOLVED', resolution = $3, "resolutionNote" = $4, "reviewedById" = $2, "reviewedAt" = now()
		WHERE id = $1 AND status = 'PENDING'`, id, byEmployeeID, resolution, note)
	return err
}

// جداول تشير لـ"Customer" (مفتاح أجنبي) — تنتقل كلها للأصلي بالدمج.
// ⚠️ جداول الجي بي اس والشرائح عمودها "customerId" يشير لـ"GpsCustomer"
// (زبائن منظومة الجي بي اس) مو لـ"Customer" — نقلها هنا يكسر المفتاح.
var customerRefTables = []string{
	"Booking", "Complaint", "QualityFollowUp", "SolarInstallation", "DeviceMaintenanceTicket",
}

// جداول صف واحد لكل زبون (فهرس فريد) — تنتقل بس إذا الأصلي ما عنده.
var customerUniqueTables = []string{"CustomerGpsInfo", "VipCustomer"}

// MergeCounts شكد صف راح ينتقل من المكرر — للعرض قبل التأكيد.
func (r *DuplicateCandidateRepository) MergeCounts(dropID string) map[string]int {
	out := map[string]int{}
	for _, t := range append(append([]string{}, customerRefTables...), "CustomerServiceTag") {
		var n int
		if err := r.db.Get(&n, `SELECT COUNT(*) FROM "`+t+`" WHERE "customerId" = $1`, dropID); err == nil && n > 0 {
			out[t] = n
		}
	}
	return out
}

// MergeCustomers ينقل كل شي من dropID لـkeepID ويشيل dropID — بمعاملة وحدة:
// أي خطأ يرجّع كلشي مثل ما چان.
func (r *DuplicateCandidateRepository) MergeCustomers(keepID, dropID string) (map[string]int64, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	moved := map[string]int64{}
	for _, t := range customerRefTables {
		res, err := tx.Exec(`UPDATE "`+t+`" SET "customerId" = $1 WHERE "customerId" = $2`, keepID, dropID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			moved[t] = n
		}
	}
	for _, t := range customerUniqueTables {
		if _, err := tx.Exec(`UPDATE "`+t+`" SET "customerId" = $1 WHERE "customerId" = $2
			AND NOT EXISTS (SELECT 1 FROM "`+t+`" WHERE "customerId" = $1)`, keepID, dropID); err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}
		if _, err := tx.Exec(`DELETE FROM "`+t+`" WHERE "customerId" = $1`, dropID); err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}
	}
	if _, err := tx.Exec(`UPDATE "CustomerServiceTag" t SET "customerId" = $1 WHERE "customerId" = $2
		AND NOT EXISTS (SELECT 1 FROM "CustomerServiceTag" k WHERE k."customerId" = $1 AND k.service = t.service)`, keepID, dropID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM "CustomerServiceTag" WHERE "customerId" = $1`, dropID); err != nil {
		return nil, err
	}
	// الموقع والخريطة: الأصلي ياخذها بس إذا هو فاضي — ما نكتب فوگ شي موجود.
	if _, err := tx.Exec(`UPDATE "Customer" k SET
		  location = COALESCE(NULLIF(k.location, ''), d.location),
		  "mapLatitude" = COALESCE(k."mapLatitude", d."mapLatitude"),
		  "mapLongitude" = COALESCE(k."mapLongitude", d."mapLongitude"),
		  "locationUrl" = COALESCE(NULLIF(k."locationUrl", ''), d."locationUrl")
		FROM "Customer" d WHERE k.id = $1 AND d.id = $2`, keepID, dropID); err != nil {
		return nil, err
	}
	// أزواج تكرار ثانية على المكرر تنتهي (الزبون انشال).
	if _, err := tx.Exec(`UPDATE "DuplicateCandidate" SET status = 'DISMISSED', "reviewedAt" = now()
		WHERE kind = 'CUSTOMER' AND status = 'PENDING' AND ("entityAId" = $1 OR "entityBId" = $1)
		  AND NOT ("entityAId" = $2 OR "entityBId" = $2)`, dropID, keepID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM "Customer" WHERE id = $1`, dropID); err != nil {
		return nil, fmt.Errorf("حذف المكرر: %w", err)
	}
	return moved, tx.Commit()
}
