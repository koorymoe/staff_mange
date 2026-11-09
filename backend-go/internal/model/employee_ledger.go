package model

import "time"

// ═══ دفتر ذمة الموظف ═══
//
// قيد واحد = حركة فلوس وحدة بين الموظف والشركة. الرصيد ما ينخزن —
// ينحسب من القيود، فما يگدر يكذب.

type EmployeeLedgerEntry struct {
	ID         string  `db:"id" json:"id"`
	EmployeeID string  `db:"employeeId" json:"employeeId"`
	Kind       string  `db:"kind" json:"kind"`
	Amount     float64 `db:"amount" json:"amount"`
	BookingID  *string `db:"bookingId" json:"bookingId,omitempty"`
	InvoiceID  *string `db:"invoiceId" json:"invoiceId,omitempty"`
	Method     *string `db:"method" json:"method,omitempty"`
	Reference  *string `db:"reference" json:"reference,omitempty"`

	OccurredAt   time.Time `db:"occurredAt" json:"occurredAt"`
	RecordedByID *string   `db:"recordedById" json:"recordedById,omitempty"`
	Note         *string   `db:"note" json:"note,omitempty"`
	// ReversesEntryID القيد الي هذا يصحّحه. التصحيح قيد جديد مو تعديل.
	ReversesEntryID *string   `db:"reversesEntryId" json:"reversesEntryId,omitempty"`
	CreatedAt       time.Time `db:"createdAt" json:"createdAt"`

	// مفكوكة للواجهة
	EmployeeName   *string `db:"-" json:"employeeName,omitempty"`
	BookingCode    *string `db:"-" json:"bookingCode,omitempty"`
	RecordedByName *string `db:"-" json:"recordedByName,omitempty"`
}

const (
	// LedgerCollected استلم فلوساً من زبون — تزيد ذمته.
	LedgerCollected = "COLLECTED"
	// LedgerHandedOver سلّم للشركة (كاش أو بطاقة أو حوالة) — تنقص ذمته.
	LedgerHandedOver = "HANDED_OVER"
	// LedgerAdjustment تصحيح بقرار إنسان. اتجاهه يجي من القيد الي
	// يعكسه: يعكس تحصيلاً فينقص، ويعكس تسليماً فيزيد.
	LedgerAdjustment = "ADJUSTMENT"
	// LedgerWriteOff إعفاء بقرار المالك — تنقص ذمته بلا ما يدفع.
	LedgerWriteOff = "WRITE_OFF"

	LedgerMethodCash     = "CASH"
	LedgerMethodCard     = "CARD"
	LedgerMethodTransfer = "TRANSFER"
)

// EmployeeOutstanding الرصيد المعلّق — «شكد ماسك بجيبه هسه».
//
// ⚠️ هذا الرقم هو كل شي: منه تطلع أسئلة ماتركس («ماسك ٣٤٠ ألف من ٦
// أيام — ليش؟») ومنه تطلع مسؤولية الموظف مؤرّخة مو كلاماً.
type EmployeeOutstanding struct {
	EmployeeID   string  `db:"employeeId" json:"employeeId"`
	EmployeeName string  `db:"employeeName" json:"employeeName"`
	Collected    float64 `db:"collected" json:"collected"`
	HandedOver   float64 `db:"handedOver" json:"handedOver"`
	Adjusted     float64 `db:"adjusted" json:"adjusted"`
	WrittenOff   float64 `db:"writtenOff" json:"writtenOff"`
	// Outstanding = collected - handedOver - adjusted - writtenOff
	Outstanding float64 `db:"outstanding" json:"outstanding"`
	// OldestOpenAt أقدم تحصيل ما انسدّد — العمر يفرّق بين «ماسك من
	// الصبح» و«ماسك من عشرة أيام».
	OldestOpenAt *time.Time `db:"oldestOpenAt" json:"oldestOpenAt,omitempty"`
	EntryCount   int        `db:"entryCount" json:"entryCount"`
}
