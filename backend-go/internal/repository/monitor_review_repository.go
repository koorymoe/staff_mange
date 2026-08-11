package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// MonitorReviewRepository صندوق المراقب.
type MonitorReviewRepository struct {
	db *sqlx.DB
}

func NewMonitorReviewRepository(db *sqlx.DB) *MonitorReviewRepository {
	return &MonitorReviewRepository{db: db}
}

// Enqueue يضيف صف للصندوق.
//
// ON CONFLICT DO NOTHING مقصود: المحطة الوحدة ما تتكرر لنفس الشي.
// لو الحجز انثبّت وانلغى وانثبّت مرة ثانية، ما نريد صفين — الصف
// الأول لسه معلق ونفس الشغلة.
//
// ⚠️ ما يرجّع خطأ يوقف العملية الأصلية: فشل إضافة صف مراقبة ما يصير
// يمنع تثبيت حجز أو إصدار فاتورة. نرجّع الخطأ والمنادي يسجّله بس.
func (r *MonitorReviewRepository) Enqueue(in model.EnqueueMonitorReview) error {
	_, err := r.db.Exec(`
		INSERT INTO "MonitorReview"
			(id, stage, "entityType", "entityId", title, summary, "ownerRole", "ownerEmployeeId")
		VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,''),$8)
		ON CONFLICT ("entityType", "entityId", stage) DO NOTHING`,
		uuid.NewString(), in.Stage, in.EntityType, in.EntityID, in.Title,
		in.Summary, in.OwnerRole, in.OwnerEmployeeID)
	return err
}

// List الصندوق مع فلاتر: المحطة، الحالة، ودور صاحب الشغل.
func (r *MonitorReviewRepository) List(stage, status, ownerRole string, limit int) ([]model.MonitorReview, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := []string{}
	args := []any{}
	add := func(clause string, v any) {
		args = append(args, v)
		where = append(where, clause)
	}
	if stage != "" {
		add(`stage = ?`, stage)
	}
	if status != "" {
		add(`status = ?`, status)
	}
	if ownerRole != "" {
		add(`"ownerRole" = ?`, ownerRole)
	}
	q := `SELECT * FROM "MonitorReview"`
	if len(where) > 0 {
		q += ` WHERE ` + strings.Join(where, " AND ")
	}
	// المعلّق أول: هو الي ينتظر قرار، والباقي تاريخ.
	q += ` ORDER BY (status = 'PENDING') DESC, "createdAt" DESC LIMIT ?`
	args = append(args, limit)

	rows := []model.MonitorReview{}
	if err := r.db.Select(&rows, r.db.Rebind(q), args...); err != nil {
		return nil, err
	}
	r.hydrate(rows)
	r.hydrateIdentity(rows)
	return rows, nil
}

// Counts عدّاد المعلّق لكل محطة — للأرقام فوق التبويبات.
func (r *MonitorReviewRepository) Counts() ([]model.MonitorInboxCount, error) {
	rows := []model.MonitorInboxCount{}
	err := r.db.Select(&rows, `
		SELECT stage, COUNT(*) AS count FROM "MonitorReview"
		WHERE status = 'PENDING' GROUP BY stage`)
	return rows, err
}

// Decide قرار المراقب — مرة وحدة بس.
//
// شرط status = 'PENDING' يمنع تغيير القرار بعد ما ينصدر: الموظف
// انبلغ بالملاحظة، وتغييرها بالخفية بعدها يكسر الثقة بالسجل.
func (r *MonitorReviewRepository) Decide(id, monitorID string, req model.DecideMonitorReviewRequest) (*model.MonitorReview, error) {
	status := model.MonitorStatusOK
	if req.Flag {
		status = model.MonitorStatusFlagged
	}
	var row model.MonitorReview
	err := r.db.Get(&row, `
		UPDATE "MonitorReview"
		SET status = $2, note = NULLIF($3,''), "reviewedById" = $4, "reviewedAt" = now()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING *`, id, status, req.Note, monitorID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("الصف مو موجود أو انبتّ بيه من قبل")
	}
	if err != nil {
		return nil, err
	}
	rows := []model.MonitorReview{row}
	r.hydrate(rows)
	r.hydrateIdentity(rows)
	return &rows[0], nil
}

// hydrate أسماء الموظفين بدفعة وحدة مو استعلام لكل صف.
func (r *MonitorReviewRepository) hydrate(rows []model.MonitorReview) {
	if len(rows) == 0 {
		return
	}
	ids := map[string]bool{}
	for i := range rows {
		if rows[i].OwnerEmployeeID != nil {
			ids[*rows[i].OwnerEmployeeID] = true
		}
		if rows[i].ReviewedByID != nil {
			ids[*rows[i].ReviewedByID] = true
		}
	}
	if len(ids) == 0 {
		return
	}
	list := make([]string, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}
	briefs := []model.EmployeeBrief{}
	q, args, err := sqlx.In(`SELECT id, name FROM "Employee" WHERE id IN (?)`, list)
	if err != nil {
		return
	}
	if err := r.db.Select(&briefs, r.db.Rebind(q), args...); err != nil {
		return
	}
	byID := map[string]model.EmployeeBrief{}
	for _, b := range briefs {
		byID[b.ID] = b
	}
	for i := range rows {
		if rows[i].OwnerEmployeeID != nil {
			if b, ok := byID[*rows[i].OwnerEmployeeID]; ok {
				c := b
				rows[i].OwnerEmployee = &c
			}
		}
		if rows[i].ReviewedByID != nil {
			if b, ok := byID[*rows[i].ReviewedByID]; ok {
				c := b
				rows[i].ReviewedBy = &c
			}
		}
	}
}

// hydrateIdentity يجيب هوية الحجز وراء كل صف — بدفعة وحدة.
//
// كل نوع كيان يوصل للحجز بطريقه: الحجز نفسه مباشرة، والفاتورة
// وتعديلها والمشتريات عبر "bookingId"، ومتابعة الجودة عبر قيدها
// الفريد. الأنواع الي ما إلها حجز (جهاز جي بي اس) ما تطلع بالنتيجة
// وتبقى بلا هوية — وهذا الصحيح.
//
// ⚠️ الفشل ما يوقف العرض: الهوية زينة مو شرط. لو الاستعلام طاح
// يرجع الصندوق بلا هوية بدل ما يطيح كله.
func (r *MonitorReviewRepository) hydrateIdentity(rows []model.MonitorReview) {
	if len(rows) == 0 {
		return
	}
	// نجمع المعرّفات حسب النوع حتى ما ندير استعلام لكل صف
	byType := map[string][]string{}
	for i := range rows {
		byType[rows[i].EntityType] = append(byType[rows[i].EntityType], rows[i].EntityID)
	}

	// كل نوع → استعلام يرجّع (key, bookingId)
	sources := map[string]string{
		"BOOKING":            `SELECT 'BOOKING|' || id AS key, id AS "bookingId", NULL::text AS "externalNo" FROM "Booking" WHERE id IN (?)`,
		"LEADER_INVOICE":     `SELECT 'LEADER_INVOICE|' || id AS key, "bookingId", "externalInvoiceNumber" AS "externalNo" FROM "LeaderInvoice" WHERE id IN (?) AND "bookingId" IS NOT NULL`,
		"INVOICE_ADJUSTMENT": `SELECT 'INVOICE_ADJUSTMENT|' || a.id AS key, i."bookingId", i."externalInvoiceNumber" AS "externalNo" FROM "LeaderInvoiceAdjustment" a JOIN "LeaderInvoice" i ON i.id = a."invoiceId" WHERE a.id IN (?) AND i."bookingId" IS NOT NULL`,
		"PROCUREMENT":        `SELECT 'PROCUREMENT|' || id AS key, "bookingId", NULL::text AS "externalNo" FROM "ProcurementRequest" WHERE id IN (?) AND "bookingId" IS NOT NULL`,
		"QUALITY_FOLLOW_UP":  `SELECT 'QUALITY_FOLLOW_UP|' || id AS key, "bookingId", NULL::text AS "externalNo" FROM "QualityFollowUp" WHERE id IN (?)`,
	}

	type link struct {
		Key       string `db:"key"`
		BookingID string `db:"bookingId"`
		// رقم الفاتورة المحاسبية — ينجي وية الرابط لأنه على الفاتورة
		// مو على الحجز. الحجز الواحد ممكن إله أكثر من فاتورة، فربطه
		// بالحجز چان يخلط أرقام فواتير مختلفة بنفس الصف.
		ExternalNo *string `db:"externalNo"`
	}
	links := []link{}
	for typ, ids := range byType {
		tmpl, ok := sources[typ]
		if !ok {
			continue
		}
		q, args, err := sqlx.In(tmpl, ids)
		if err != nil {
			continue
		}
		part := []link{}
		if err := r.db.Select(&part, r.db.Rebind(q), args...); err != nil {
			continue
		}
		links = append(links, part...)
	}
	if len(links) == 0 {
		return
	}

	bookingIDs := map[string]bool{}
	for _, l := range links {
		bookingIDs[l.BookingID] = true
	}
	list := make([]string, 0, len(bookingIDs))
	for id := range bookingIDs {
		list = append(list, id)
	}

	// الليدر: المشرف المعيّن، وإلا أول مكلّف مؤشّر «تيم ليدر».
	// ⚠️ ما نرجّع أول فني لما ماكو ليدر — تنسب مسؤولية لواحد ما تحمّلها.
	q, args, err := sqlx.In(`
		SELECT b.id AS "bookingId", b.code AS "bookingCode",
		       c."customerCode", c.name AS "customerName", c.phone AS "customerPhone",
		       b.address,
		       COALESCE(sup.name, (
		         SELECT e2.name FROM "BookingAssignment" ba
		         JOIN "Employee" e2 ON e2.id = ba."employeeId"
		         WHERE ba."bookingId" = b.id AND e2."isLeader" = true
		         ORDER BY ba."createdAt" LIMIT 1
		       )) AS "leaderName"
		FROM "Booking" b
		JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Employee" sup ON sup.id = b."projectSupervisorId"
		WHERE b.id IN (?)`, list)
	if err != nil {
		return
	}
	idents := []model.MonitorIdentity{}
	if err := r.db.Select(&idents, r.db.Rebind(q), args...); err != nil {
		return
	}
	byBooking := map[string]model.MonitorIdentity{}
	for _, it := range idents {
		byBooking[it.BookingID] = it
	}
	byKey := map[string]model.MonitorIdentity{}
	for _, l := range links {
		if it, ok := byBooking[l.BookingID]; ok {
			// نسخة لكل مفتاح: رقم الفاتورة يخص الصف مو الحجز
			row := it
			row.ExternalInvoiceNumber = l.ExternalNo
			byKey[l.Key] = row
		}
	}
	for i := range rows {
		if it, ok := byKey[rows[i].EntityType+"|"+rows[i].EntityID]; ok {
			c := it
			rows[i].Identity = &c
		}
	}
}
