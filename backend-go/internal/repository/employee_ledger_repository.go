package repository

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ مستودع دفتر الذمة ═══
//
// ⚠️ ماكو `Update` ولا `Delete` بهذا الملف — **مو سهو**. قاعدة البيانات
// ترفضهن بمُشغّل، والمستودع ما ينطي واجهة لشي مرفوض أصلاً. التصحيح
// يمر بـ`Reverse` حصراً.
type EmployeeLedgerRepository struct {
	db *sqlx.DB
}

func NewEmployeeLedgerRepository(db *sqlx.DB) *EmployeeLedgerRepository {
	return &EmployeeLedgerRepository{db: db}
}

// ═══ الرصيد المعلّق ═══
//
// ⚠️ ينحسب من القيود كل مرة، ما ينخزن بعمود. عمود «رصيد» يعني حقيقتين
// لنفس الشي: قيود تگول رقماً وعمود يگول ثانياً لأن أحداً نسى يحدّثه.
// الحساب أبطأ بشوي ودائماً صادق.
//
// ⚠️ والتصحيح (`ADJUSTMENT`) ينطرح: يعكس تحصيلاً فينقص الذمة. عكس
// التسليم نادر ويتعالج بقيد تحصيل جديد بملاحظته.
const outstandingSelectSQL = `
	SELECT
		e.id   AS "employeeId",
		e.name AS "employeeName",
		COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'COLLECTED'), 0)    AS collected,
		COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'HANDED_OVER'), 0)  AS "handedOver",
		COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'ADJUSTMENT'), 0)   AS adjusted,
		COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'WRITE_OFF'), 0)    AS "writtenOff",
		COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'COLLECTED'), 0)
			- COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'HANDED_OVER'), 0)
			- COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'ADJUSTMENT'), 0)
			- COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'WRITE_OFF'), 0) AS outstanding,
		MIN(l."occurredAt") FILTER (WHERE l.kind = 'COLLECTED')           AS "oldestOpenAt",
		COUNT(l.id)                                                        AS "entryCount"
	FROM "Employee" e
	LEFT JOIN "EmployeeLedgerEntry" l ON l."employeeId" = e.id
`

// Outstanding يرجّع رصيد موظف واحد.
func (r *EmployeeLedgerRepository) Outstanding(employeeID string) (*model.EmployeeOutstanding, error) {
	var row model.EmployeeOutstanding
	err := r.db.Get(&row, outstandingSelectSQL+`
		WHERE e.id = $1
		GROUP BY e.id, e.name`, employeeID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// OutstandingAll يرجّع كل الموظفين الي عليهم ذمة، الأكبر أولاً.
//
// ⚠️ `HAVING` على الفرق مو على وجود قيود: موظف سلّم كلشي رصيده صفر
// وما يستاهل يزاحم بالقائمة.
func (r *EmployeeLedgerRepository) OutstandingAll() ([]model.EmployeeOutstanding, error) {
	rows := []model.EmployeeOutstanding{}
	err := r.db.Select(&rows, outstandingSelectSQL+`
		GROUP BY e.id, e.name
		HAVING COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'COLLECTED'), 0)
			- COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'HANDED_OVER'), 0)
			- COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'ADJUSTMENT'), 0)
			- COALESCE(SUM(l.amount) FILTER (WHERE l.kind = 'WRITE_OFF'), 0) <> 0
		ORDER BY outstanding DESC`)
	return rows, err
}

// Record يكتب قيداً جديداً.
//
// ⚠️ الفحوص الحقيقية بقاعدة البيانات (تحصيل بلا حجز، مبلغ سالب، نوع
// مخترع) — وهنا نترجم رفضها لرسالة يفهمها المستخدم بدل نص Postgres.
func (r *EmployeeLedgerRepository) Record(e model.EmployeeLedgerEntry) (*model.EmployeeLedgerEntry, error) {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	var out model.EmployeeLedgerEntry
	err := r.db.Get(&out, `
		INSERT INTO "EmployeeLedgerEntry"
			(id, "employeeId", kind, amount, "bookingId", "invoiceId", method,
			 reference, "occurredAt", "recordedById", note, "reversesEntryId")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE($9, NOW()),$10,$11,$12)
		RETURNING *`,
		e.ID, e.EmployeeID, e.Kind, e.Amount, e.BookingID, e.InvoiceID, e.Method,
		e.Reference, nullTime(e.OccurredAt), e.RecordedByID, e.Note, e.ReversesEntryID)
	if err != nil {
		return nil, translateLedgerError(err)
	}
	return &out, nil
}

// Reverse يصحّح قيداً بقيد عكسي — الطريق الوحيد للتصحيح.
//
// ⚠️ ياخذ المبلغ من القيد الأصلي مو من المنادي: مبلغ يجي من برّا يعني
// عكساً جزئياً صامتاً يخلّي الدفتر ما يوازن.
func (r *EmployeeLedgerRepository) Reverse(entryID, reason, byEmployeeID string) (*model.EmployeeLedgerEntry, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("التصحيح لازم إله سبب")
	}
	var orig model.EmployeeLedgerEntry
	if err := r.db.Get(&orig, `SELECT * FROM "EmployeeLedgerEntry" WHERE id = $1`, entryID); err != nil {
		return nil, fmt.Errorf("القيد مو موجود")
	}
	// عكس التحصيل تصحيح ينقص الذمة؛ وعكس التسليم يزيدها فينكتب تحصيلاً.
	kind := model.LedgerAdjustment
	if orig.Kind == model.LedgerHandedOver {
		kind = model.LedgerCollected
	}
	return r.Record(model.EmployeeLedgerEntry{
		EmployeeID:      orig.EmployeeID,
		Kind:            kind,
		Amount:          orig.Amount,
		BookingID:       orig.BookingID,
		Note:            &reason,
		RecordedByID:    &byEmployeeID,
		ReversesEntryID: &orig.ID,
	})
}

// ListForEmployee سجل موظف — الأحدث أولاً.
func (r *EmployeeLedgerRepository) ListForEmployee(employeeID string, limit int) ([]model.EmployeeLedgerEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows := []model.EmployeeLedgerEntry{}
	err := r.db.Select(&rows, `
		SELECT * FROM "EmployeeLedgerEntry"
		WHERE "employeeId" = $1
		ORDER BY "occurredAt" DESC, "createdAt" DESC
		LIMIT $2`, employeeID, limit)
	return rows, err
}

// ListForBooking كل حركات الفلوس لحجز واحد.
func (r *EmployeeLedgerRepository) ListForBooking(bookingID string) ([]model.EmployeeLedgerEntry, error) {
	rows := []model.EmployeeLedgerEntry{}
	err := r.db.Select(&rows, `
		SELECT * FROM "EmployeeLedgerEntry"
		WHERE "bookingId" = $1
		ORDER BY "occurredAt"`, bookingID)
	return rows, err
}

// CollectedForBooking شكد انستلم من حجز — بعد طرح ما انعكس.
//
// ⚠️ القيود المعكوسة تنشال من الحساب: قيد انعكس يعني ما صار، وتركه
// بالمجموع يخلّي الحجز يبان مدفوعاً وهو لا.
func (r *EmployeeLedgerRepository) CollectedForBooking(bookingID string) (float64, error) {
	var total float64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM(amount), 0) FROM "EmployeeLedgerEntry" l
		WHERE l."bookingId" = $1 AND l.kind = 'COLLECTED'
		  AND NOT EXISTS (
		      SELECT 1 FROM "EmployeeLedgerEntry" rev
		      WHERE rev."reversesEntryId" = l.id
		  )`, bookingID)
	return total, err
}

func nullTime(t interface{ IsZero() bool }) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return t
}

// translateLedgerError يحوّل رفض قاعدة البيانات لرسالة عربية مفهومة.
func translateLedgerError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "employee_ledger_collected_needs_booking"):
		return fmt.Errorf("التحصيل لازم ينربط بحجز — المبلغ بلا حجز ما ينقبل")
	case strings.Contains(msg, "employee_ledger_reversal_needs_note"):
		return fmt.Errorf("التصحيح لازم إله سبب مكتوب")
	case strings.Contains(msg, "EmployeeLedgerEntry_amount_check"):
		return fmt.Errorf("المبلغ لازم يكون أكبر من صفر")
	case strings.Contains(msg, "EmployeeLedgerEntry_kind_check"):
		return fmt.Errorf("نوع الحركة مو معروف")
	case strings.Contains(msg, "EmployeeLedgerEntry_reverse_unique"):
		return fmt.Errorf("هذا القيد انصحّح من قبل — ما ينصحّح مرتين")
	case strings.Contains(msg, "ما ينعدّل ولا ينمسح"):
		return fmt.Errorf("دفتر الذمة ما ينعدّل — التصحيح بقيد عكسي جديد")
	}
	return err
}
