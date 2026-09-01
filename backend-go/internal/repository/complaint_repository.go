package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ComplaintRepository struct {
	db *sqlx.DB
}

func NewComplaintRepository(db *sqlx.DB) *ComplaintRepository {
	return &ComplaintRepository{db: db}
}

func (r *ComplaintRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *ComplaintRepository) hydrate(c *model.Complaint) {
	var customer model.Customer
	if err := r.db.Get(&customer, `SELECT * FROM "Customer" WHERE id = $1`, c.CustomerID); err == nil {
		c.Customer = &customer
	}
	if c.BookingID != nil {
		var booking model.Booking
		if err := r.db.Get(&booking, `SELECT * FROM "Booking" WHERE id = $1`, *c.BookingID); err == nil {
			c.Booking = &booking
		}
	}
	c.CreatedByEmployee = r.loadEmployeeBrief(c.CreatedByEmployeeID)
	if c.AssignedToEmployeeID != nil {
		c.AssignedToEmployee = r.loadEmployeeBrief(*c.AssignedToEmployeeID)
	}
	if c.RelatedEmployeeID != nil {
		c.RelatedEmployee = r.loadEmployeeBrief(*c.RelatedEmployeeID)
	}
	if c.AuditedByID != nil {
		c.AuditedBy = r.loadEmployeeBrief(*c.AuditedByID)
	}
	// منو اتصل بالزبون — لازم يبين كدام الشكوى مثل ما انطلب
	if c.ContactedByID != nil {
		if e := r.loadEmployeeBrief(*c.ContactedByID); e != nil {
			name := e.Name
			c.ContactedByName = &name
		}
	}
}

// SetContacted يأشر إن أحد اتصل بالزبون (أو يشيل التأشير)، ويخزن منو
// اتصل ومتى. أي موظف يقدر — المهم نعرف منو.
func (r *ComplaintRepository) SetContacted(
	id string, contacted bool, byID string, rating *int,
) (*model.Complaint, error) {
	// ⚠️ التقييم يتفحّص هنا **زيادة** على قيد قاعدة البيانات مو بديلاً
	// عنه: القيد هو الي يحمي من نداء مباشر أو مسار ثاني ينكتب بكرة.
	if rating != nil && (*rating < 1 || *rating > 5) {
		return nil, errors.New("التقييم لازم يكون من ١ إلى ٥")
	}
	var q string
	var args []any
	if contacted {
		// التقييم ينكتب بس لمن ينعطى: nil يعني «ما سأل» فنخلي
		// القديم مثل ما هو بدل ما نمحيه بكل تأشيرة تواصل.
		if rating != nil {
			q = `UPDATE "Complaint" SET "contactedAt" = now(), "contactedById" = $2,
			     "customerRating" = $3 WHERE id = $1`
			args = []any{id, byID, *rating}
		} else {
			q = `UPDATE "Complaint" SET "contactedAt" = now(), "contactedById" = $2 WHERE id = $1`
			args = []any{id, byID}
		}
	} else {
		// شيل التأشير = شيل التقييم وياه: تقييم بلا مكالمة ما إله مصدر.
		q = `UPDATE "Complaint" SET "contactedAt" = NULL, "contactedById" = NULL,
		     "customerRating" = NULL WHERE id = $1`
		args = []any{id}
	}
	if _, err := r.db.Exec(q, args...); err != nil {
		return nil, err
	}
	return r.Find(id)
}

// Audit حكم المدقق على شغل مهندس الجودة بهذي الشكوى.
func (r *ComplaintRepository) Audit(
	id, verdict string, note *string, byID string,
) (*model.Complaint, error) {
	if _, ok := model.AuditVerdictLabels[verdict]; !ok {
		return nil, errors.New("حكم تدقيق غير معروف")
	}
	_, err := r.db.Exec(`
		UPDATE "Complaint"
		SET "auditVerdict" = $2, "auditNote" = NULLIF($3,''),
		    "auditedAt" = now(), "auditedById" = $4
		WHERE id = $1`, id, verdict, deref(note), byID)
	if err != nil {
		return nil, err
	}
	return r.Find(id)
}

// AddEvent يكتب سطراً بسجل الشكوى.
//
// ⚠️ ماكو دالة تعديل ولا حذف بقصد — سجل ينعدّل مو سجل.
// ⚠️ وفشل الكتابة **ما يلغي** العملية الأصلية: التواصل صار فعلاً،
// وضياع سطر بالسجل أهون من رفض عملية نجحت.
func (r *ComplaintRepository) AddEvent(complaintID, kind string, detail *string, byID string) {
	// الاسم ينحلّ هنا وينحفظ نصاً: لو انحذف الموظف بعدين يبقى السطر
	// مقروءاً بدل «—».
	byName := ""
	if byID != "" {
		if e := r.loadEmployeeBrief(byID); e != nil {
			byName = e.Name
		}
	}
	_, _ = r.db.Exec(`
		INSERT INTO "ComplaintEvent" (id, "complaintId", kind, detail, "byEmployeeId", "byName")
		VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), NULLIF($6,''))`,
		uuid.NewString(), complaintID, kind, deref(detail), byID, byName)
}

// Events أحداث الشكوى — الأحدث أولاً.
func (r *ComplaintRepository) Events(complaintID string) ([]model.ComplaintEvent, error) {
	events := []model.ComplaintEvent{}
	err := r.db.Select(&events, `
		SELECT * FROM "ComplaintEvent"
		WHERE "complaintId" = $1
		ORDER BY "createdAt" DESC`, complaintID)
	return events, err
}

// SetNotes ملاحظات الزبون — نستفاد منها بالتحسين.
func (r *ComplaintRepository) SetNotes(id, notes string) (*model.Complaint, error) {
	if _, err := r.db.Exec(`UPDATE "Complaint" SET notes = NULLIF($2,'') WHERE id = $1`, id, notes); err != nil {
		return nil, err
	}
	return r.Find(id)
}

func (r *ComplaintRepository) Find(id string) (*model.Complaint, error) {
	var c model.Complaint
	if err := r.db.Get(&c, `SELECT * FROM "Complaint" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
}

func (r *ComplaintRepository) List() ([]model.Complaint, error) {
	complaints := []model.Complaint{}
	if err := r.db.Select(&complaints, `SELECT * FROM "Complaint" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range complaints {
		r.hydrate(&complaints[i])
	}
	return complaints, nil
}

func (r *ComplaintRepository) Create(customerID string, bookingID *string, complaintType, description, createdByEmployeeID string, relatedEmployeeID *string) (*model.Complaint, error) {
	var c model.Complaint
	err := r.db.Get(&c, `
		INSERT INTO "Complaint" (id, "customerId", "bookingId", type, description, "createdByEmployeeId", "relatedEmployeeId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, customerID, bookingID, complaintType, description, createdByEmployeeID, relatedEmployeeID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
}

func (r *ComplaintRepository) Update(id string, status, assignedToEmployeeID, resolution *string) (*model.Complaint, error) {
	var c model.Complaint
	err := r.db.Get(&c, `
		UPDATE "Complaint" SET
			status = COALESCE($2, status),
			"assignedToEmployeeId" = COALESCE($3, "assignedToEmployeeId"),
			resolution = COALESCE($4, resolution)
		WHERE id = $1
		RETURNING *
	`, id, status, assignedToEmployeeID, resolution)
	if err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
}

// StatsByCustomer يرجع عدد الشكاوى لكل زبون اشتكى (مرة واحدة على الأقل)، مرتبة من
// الأكثر شكاوى — تقرير منفصل تماماً عن إحصائيات الحجوزات.
// StatsByCustomer صف لكل زبون — هذا الي تعرضه شاشة الشكاوى.
//
// ⚠️ الصف **مجمّع**، والتصميم يطلب «حالة الشكوى» و«مهندس الجودة»
// مفردين — فنجيبهن من **آخر** شكوى للزبون (`DISTINCT ON`) ونسمّيهن
// latest* حتى الواجهة تكتبها صراحةً. بلا هذا تصير قراءتان مختلفتان
// لنفس السطر: «٤ شكاوى» وحالة وحدة، وما يبين إلى أي وحدة تعود.
func (r *ComplaintRepository) StatsByCustomer() ([]model.ComplaintCustomerStat, error) {
	stats := []model.ComplaintCustomerStat{}
	err := r.db.Select(&stats, `
		WITH latest AS (
			SELECT DISTINCT ON (c."customerId")
				c."customerId", c.id AS "complaintId", c.status, e.name AS "engineerName"
			FROM "Complaint" c
			LEFT JOIN "Employee" e ON e.id = c."assignedToEmployeeId"
			ORDER BY c."customerId", c."createdAt" DESC
		)
		SELECT
			cu.id    AS "customerId",
			cu.name  AS "customerName",
			cu.phone AS "customerPhone",
			COUNT(c.id) AS "complaintCount",
			COUNT(*) FILTER (WHERE c.status IN ('NEW', 'IN_PROGRESS')) AS "openCount",
			COUNT(*) FILTER (WHERE c."contactedAt" IS NULL)            AS "notContactedCount",
			COUNT(*) FILTER (WHERE c."contactedAt" >= now() - interval '30 days')
				AS "contactedLast30",
			-- ⚠️ التدقيق مطلوب بس على الشكاوى الي انتصل بيها فعلاً:
			-- شكوى ما انتصل بيها ماكو شغل مهندس جودة حتى يتدقّق.
			COUNT(*) FILTER (
				WHERE c."contactedAt" IS NOT NULL
				  AND COALESCE(c."auditVerdict", '') <> 'APPROVED'
			) AS "needsAuditCount",
			MAX(c."contactedAt") AS "lastContactAt",
			-- ⚠️ AVG بلا FILTER يحسب الفاضيات صفراً ويهبّط المتوسط.
			-- ولمن ماكو ولا تقييم يرجّع NULL — وهاي مقصودة: الواجهة
			-- تعرض «—» مو «٠».
			AVG(c."customerRating") FILTER (WHERE c."customerRating" IS NOT NULL)
				AS "avgRating",
			MAX(l.status)         AS "latestStatus",
			MAX(l."engineerName") AS "latestEngineer",
			MAX(l."complaintId")  AS "latestComplaintId"
		FROM "Complaint" c
		JOIN "Customer" cu ON cu.id = c."customerId"
		LEFT JOIN latest l ON l."customerId" = cu.id
		GROUP BY cu.id, cu.name, cu.phone
		ORDER BY "notContactedCount" DESC, "complaintCount" DESC
	`)
	return stats, err
}

// CountForEmployeeMonth يرجّع عدد الشكاوى المرتبطة بموظف معيّن (relatedEmployeeId)
// خلال شهر معيّن (monthPrefix بصيغة "YYYY-MM").
func (r *ComplaintRepository) CountForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "Complaint"
		WHERE "relatedEmployeeId" = $1 AND to_char("createdAt", 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return count, err
}

// CountForEmployeeRange نفس CountForEmployeeMonth لكن لمدى تاريخ حر (from/to
// بصيغة "YYYY-MM-DD").
func (r *ComplaintRepository) CountForEmployeeRange(employeeID, from, to string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "Complaint"
		WHERE "relatedEmployeeId" = $1 AND baghdad_date("createdAt") BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return count, err
}

func (r *ComplaintRepository) Resolve(id string, resolution *string) (*model.Complaint, error) {
	var c model.Complaint
	err := r.db.Get(&c, `
		UPDATE "Complaint" SET status = 'RESOLVED', resolution = COALESCE($2, resolution), "resolvedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, resolution)
	if err != nil {
		return nil, err
	}
	r.hydrate(&c)
	return &c, nil
}
