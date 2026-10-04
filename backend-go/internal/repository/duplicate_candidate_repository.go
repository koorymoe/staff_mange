package repository

import (
	"fmt"
	"time"

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
		if c.BookingA != nil && c.BookingB != nil {
			c.Analysis, c.Suggested = analyzeBookings(c.BookingA, c.BookingB)
		}
	case model.DuplicateKindCustomer:
		c.CustomerA = r.customerBrief(c.EntityAID)
		c.CustomerB = r.customerBrief(c.EntityBID)
		if c.CustomerA != nil && c.CustomerB != nil {
			c.Analysis, c.Suggested = analyzeCustomers(c.CustomerA, c.CustomerB)
		}
	}
}

func (r *DuplicateCandidateRepository) bookingBrief(id string) *model.DuplicateBookingBrief {
	var b model.DuplicateBookingBrief
	err := r.db.Get(&b, `
		SELECT b.id, b.code, c.name AS "customerName", c.phone AS "customerPhone",
		       c."customerCode", b.address, s.name AS "serviceName", b."scheduledAt",
		       b.status::text AS status, b."createdAt", b."createdById", e.name AS "createdByName",
		       (b."startedAt" IS NOT NULL OR b.status::text IN ('IN_PROGRESS','COMPLETED','PARTIAL')) AS started,
		       EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id) AS "hasInvoice"
		FROM "Booking" b
		LEFT JOIN "Employee" e ON e.id = b."createdById"
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
	if err := r.db.Get(&c, `SELECT id, name, phone, "customerCode", "createdAt",
		(SELECT COUNT(*) FROM "Booking" b WHERE b."customerId" = "Customer".id) AS bookings
		FROM "Customer" WHERE id = $1`, id); err != nil {
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

// ═══ ماتركس يحلل كل زوج — ليش تكرار، وشنو الأرجح صار، ومنو المرشّح ═══

func durAr(d time.Duration) string {
	m := int(d.Abs().Minutes())
	switch {
	case m < 1:
		return "أقل من دقيقة"
	case m < 60:
		return fmt.Sprintf("%d دقيقة", m)
	case m < 48*60:
		return fmt.Sprintf("%d ساعة", m/60)
	}
	return fmt.Sprintf("%d يوم", m/1440)
}

func nameOr(p *string, def string) string {
	if p == nil || *p == "" {
		return def
	}
	return *p
}

func analyzeBookings(a, b *model.DuplicateBookingBrief) (string, string) {
	first, second := a, b
	if b.CreatedAt.Before(a.CreatedAt) {
		first, second = b, a
	}
	gap := second.CreatedAt.Sub(first.CreatedAt)
	n1, n2 := nameOr(first.CreatedByName, "النظام/بلا اسم"), nameOr(second.CreatedByName, "النظام/بلا اسم")
	var why string
	sameBy := first.CreatedByID != nil && second.CreatedByID != nil && *first.CreatedByID == *second.CreatedByID
	switch {
	case sameBy && gap <= 10*time.Minute:
		why = fmt.Sprintf("نفس الموظف (%s) سجّلهم بفارق %s — غالباً ضغط «احفظ» مرتين.", n1, durAr(gap))
	case sameBy:
		why = fmt.Sprintf("نفس الموظف (%s) رجع سجّل نفس الطلب بعد %s — ممكن نسى إنه مسجّل، أو الزبون اتصل مرة ثانية.", n1, durAr(gap))
	default:
		why = fmt.Sprintf("%s سجّل %s، وبعد %s %s سجّل %s لنفس الزبون ونفس العنوان — غالباً الزبون تواصل مرتين وكل واحد سجّله.", n1, first.Code, durAr(gap), n2, second.Code)
	}
	// المرشّح للحذف: الي ما بدا بي أحد وبلا فاتورة؛ وإذا الاثنين سوة، الأحدث.
	drop := second
	switch {
	case second.Started || second.HasInvoice:
		if !first.Started && !first.HasInvoice {
			drop = first
		} else {
			return why + " ⚠️ الاثنين عليهم شغل أو فاتورة — راجعهم قبل أي حذف (ممكن شغلتين صدك).", ""
		}
	}
	return why + fmt.Sprintf(" المرشّح للحذف: %s (%s).", drop.Code, map[bool]string{true: "ما بدا بي أحد وبلا فاتورة", false: "الأحدث"}[!drop.Started && !drop.HasInvoice]), drop.ID
}

func analyzeCustomers(a, b *model.DuplicateCustomerBrief) (string, string) {
	keep, drop := a, b
	if b.Bookings > a.Bookings || (b.Bookings == a.Bookings && b.CreatedAt.Before(a.CreatedAt)) {
		keep, drop = b, a
	}
	return fmt.Sprintf("نفس رقم الهاتف انسجّل مرتين باسمين: «%s» (%d حجز، من %s) و«%s» (%d حجز، من %s). الاقتراح: خلّي «%s» الأصلي لأن عليه حجوزات أكثر أو أقدم، وادمج الثاني بيه.",
		a.Name, a.Bookings, a.CreatedAt.Format("2006-01-02"), b.Name, b.Bookings, b.CreatedAt.Format("2006-01-02"), keep.Name), keep.ID + "|" + drop.ID
}

// DuplicateReport ملخص ماتركس للتكرار — للمالك وصاحب الصلاحية.
type DuplicateReport struct {
	PendingBookings  int                  `json:"pendingBookings"`
	PendingCustomers int                  `json:"pendingCustomers"`
	Resolved30       int                  `json:"resolved30"`
	Detected30       int                  `json:"detected30"`
	ByCreator        []DuplicateByCreator `json:"byCreator"`
}

type DuplicateByCreator struct {
	Name  string `db:"name" json:"name"`
	Count int    `db:"count" json:"count"`
	Quick int    `db:"quick" json:"quick"` // منها بفارق ١٠ دقايق (ضغط مرتين)
}

// Report منو سجّل الحجز الثاني (الأحدث) بأزواج آخر ٣٠ يوم — حتى نعرف وين ينتج التكرار.
func (r *DuplicateCandidateRepository) Report() (*DuplicateReport, error) {
	rep := &DuplicateReport{ByCreator: []DuplicateByCreator{}}
	_ = r.db.Get(&rep.PendingBookings, `SELECT COUNT(*) FROM "DuplicateCandidate" WHERE kind = 'BOOKING' AND status = 'PENDING'`)
	_ = r.db.Get(&rep.PendingCustomers, `SELECT COUNT(*) FROM "DuplicateCandidate" WHERE kind = 'CUSTOMER' AND status = 'PENDING'`)
	_ = r.db.Get(&rep.Resolved30, `SELECT COUNT(*) FROM "DuplicateCandidate" WHERE status = 'RESOLVED' AND "reviewedAt" > now() - interval '30 days'`)
	_ = r.db.Get(&rep.Detected30, `SELECT COUNT(*) FROM "DuplicateCandidate" WHERE "detectedAt" > now() - interval '30 days'`)
	err := r.db.Select(&rep.ByCreator, `
		SELECT COALESCE(e.name, 'بلا اسم') AS name, COUNT(*) AS count,
		       COUNT(*) FILTER (WHERE x.gap <= interval '10 minutes') AS quick
		FROM (
		  SELECT CASE WHEN b1."createdAt" >= b2."createdAt" THEN b1."createdById" ELSE b2."createdById" END AS by,
		         abs(extract(epoch FROM b1."createdAt" - b2."createdAt")) * interval '1 second' AS gap
		  FROM "DuplicateCandidate" d
		  JOIN "Booking" b1 ON b1.id = d."entityAId"
		  JOIN "Booking" b2 ON b2.id = d."entityBId"
		  WHERE d.kind = 'BOOKING' AND d.status <> 'DISMISSED' AND d."detectedAt" > now() - interval '30 days'
		) x LEFT JOIN "Employee" e ON e.id = x.by
		GROUP BY e.name ORDER BY count DESC LIMIT 10`)
	return rep, err
}
