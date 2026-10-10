package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ مستودع الاستكشاف الأسبوعي ═══
//
// كل استعلام هنا **حتمي** — يحسب بسطاً ومقاماً حقيقيين من كل البيانات،
// بلا أي نموذج لغوي يبني الاستعلام أو يقترح محوراً. المحاور الأربعة
// (يوم الأسبوع، نوع المنظومة، الوردية، الموظف) ثابتة بالكود.
type AiDiscoveryRepository struct {
	db *sqlx.DB
}

func NewAiDiscoveryRepository(db *sqlx.DB) *AiDiscoveryRepository {
	return &AiDiscoveryRepository{db: db}
}

// AxisCell صف وحد بالشبكة: قيمة المحور + بسط/مقام خام. النسبة والبوابتان
// (حد أدنى للعيّنة + انحراف عن المعدل) تُحسب بالخدمة، مو هنا.
type AxisCell struct {
	ScopeID     string `db:"scope_id"`
	Numerator   int    `db:"numerator"`
	Denominator int    `db:"denominator"`
}

// bookingAxisGroupExpr تعبير SQL لمحور معيّن — من قائمة ثابتة بالكود
// حصراً (٣ قيم فقط)، فماكو خطر حقن حتى لو انبنى بـ%s.
func bookingAxisGroupExpr(axis string) (string, bool) {
	switch axis {
	case model.AiAxisDayOfWeek:
		return `EXTRACT(ISODOW FROM b."scheduledAt")::text`, true
	case model.AiAxisSystemType:
		return `b."systemType"`, true
	case model.AiAxisShift:
		return `b.shift::text`, true
	}
	return "", false
}

// ⚠️ الليدر يُحدَّد بنفس المنطق المستخدم أصلاً بكيان «الورق الناقص»
// (discipline_repository.go): آخر طلعة `BookingVisitCrew.isLeader`،
// وإلا `BookingAssignment` بدور ليدر. نفس التعريف بكل مكان — موظف
// وحده لا نتوهم إسناداً غير موجود.
const leaderLateralJoin = `
	JOIN LATERAL (
		SELECT e.id
		FROM "Employee" e
		WHERE e.id = COALESCE(
			(SELECT vc."employeeId"
			 FROM "BookingVisitCrew" vc
			 JOIN "BookingVisit" v ON v.id = vc."visitId"
			 WHERE v."bookingId" = b.id AND vc."isLeader"
			 ORDER BY v."visitNumber" DESC LIMIT 1),
			(SELECT ba."employeeId" FROM "BookingAssignment" ba
			 JOIN "Employee" le ON le.id = ba."employeeId"
			 WHERE ba."bookingId" = b.id AND le."isLeader"
			 ORDER BY ba.role LIMIT 1)
		)
	) ldr ON true`

// bookingUniverseWhere الحجوزات المؤهلة — طلعت فعلياً (وصلت أو
// اكتملت) وما هي مؤرشفة. نفس المجموعة تُستخدم كمقام لكل مقاييس
// الحجز الأربعة حتى تنقارن بمنهجية وحدة.
const bookingUniverseWhere = `(b."completedAt" IS NOT NULL OR b."arrivedAt" IS NOT NULL) AND b."archivedAt" IS NULL`

// SignalRateByBookingAxis نسبة إشارة معيّنة (LATE_START أو WORK_STOPPED)
// لكل قيمة محور، مقابل كل الحجوزات المؤهلة بنفس القيمة.
func (r *AiDiscoveryRepository) SignalRateByBookingAxis(signalKind, axis string) ([]AxisCell, error) {
	groupExpr, ok := bookingAxisGroupExpr(axis)
	if !ok {
		return nil, fmt.Errorf("محور غير معروف: %s", axis)
	}
	rows := []AxisCell{}
	query := fmt.Sprintf(`
		SELECT axis_value AS scope_id,
		       COUNT(*) FILTER (WHERE has_signal) AS numerator,
		       COUNT(*) AS denominator
		FROM (
			SELECT b.id, (%s) AS axis_value,
			       EXISTS (
			           SELECT 1 FROM "AiSignal" s
			           WHERE s."entityType" = 'BOOKING' AND s."entityId" = b.id AND s.kind = $1
			       ) AS has_signal
			FROM "Booking" b
			WHERE %s
		) sub
		WHERE axis_value IS NOT NULL AND axis_value <> ''
		GROUP BY axis_value`, groupExpr, bookingUniverseWhere)
	err := r.db.Select(&rows, query, signalKind)
	return rows, err
}

// SignalRateByEmployee نفس الشي أعلاه بس محور الموظف — الليدر المسؤول
// عن الحجز، مو الموظف المذكور بالإشارة (قد يكون فني ثانٍ بالكادر).
func (r *AiDiscoveryRepository) SignalRateByEmployee(signalKind string) ([]AxisCell, error) {
	rows := []AxisCell{}
	query := fmt.Sprintf(`
		SELECT ldr.id AS scope_id,
		       COUNT(*) FILTER (WHERE EXISTS (
		           SELECT 1 FROM "AiSignal" s
		           WHERE s."entityType" = 'BOOKING' AND s."entityId" = b.id AND s.kind = $1
		       )) AS numerator,
		       COUNT(*) AS denominator
		FROM "Booking" b
		%s
		WHERE %s
		GROUP BY ldr.id`, leaderLateralJoin, bookingUniverseWhere)
	err := r.db.Select(&rows, query, signalKind)
	return rows, err
}

// PartialRateByBookingAxis نسبة الحجوزات الي احتاجت أكثر من يوم
// (`partialCount > 0`) لكل قيمة محور.
func (r *AiDiscoveryRepository) PartialRateByBookingAxis(axis string) ([]AxisCell, error) {
	groupExpr, ok := bookingAxisGroupExpr(axis)
	if !ok {
		return nil, fmt.Errorf("محور غير معروف: %s", axis)
	}
	rows := []AxisCell{}
	query := fmt.Sprintf(`
		SELECT axis_value AS scope_id,
		       COUNT(*) FILTER (WHERE "partialCount" > 0) AS numerator,
		       COUNT(*) AS denominator
		FROM (
			SELECT b.id, b."partialCount", (%s) AS axis_value
			FROM "Booking" b
			WHERE %s
		) sub
		WHERE axis_value IS NOT NULL AND axis_value <> ''
		GROUP BY axis_value`, groupExpr, bookingUniverseWhere)
	err := r.db.Select(&rows, query)
	return rows, err
}

func (r *AiDiscoveryRepository) PartialRateByEmployee() ([]AxisCell, error) {
	rows := []AxisCell{}
	query := fmt.Sprintf(`
		SELECT ldr.id AS scope_id,
		       COUNT(*) FILTER (WHERE b."partialCount" > 0) AS numerator,
		       COUNT(*) AS denominator
		FROM "Booking" b
		%s
		WHERE %s
		GROUP BY ldr.id`, leaderLateralJoin, bookingUniverseWhere)
	err := r.db.Select(&rows, query)
	return rows, err
}

// InvoiceAdjustRateByBookingAxis نسبة فواتير الليدر الي احتاجت تعديل
// مبالغ بعد التسجيل، لكل قيمة محور (عبر ربط الفاتورة بحجزها).
func (r *AiDiscoveryRepository) InvoiceAdjustRateByBookingAxis(axis string) ([]AxisCell, error) {
	groupExpr, ok := bookingAxisGroupExpr(axis)
	if !ok {
		return nil, fmt.Errorf("محور غير معروف: %s", axis)
	}
	rows := []AxisCell{}
	query := fmt.Sprintf(`
		SELECT axis_value AS scope_id,
		       COUNT(*) FILTER (WHERE has_signal) AS numerator,
		       COUNT(*) AS denominator
		FROM (
			SELECT li.id, (%s) AS axis_value,
			       EXISTS (
			           SELECT 1 FROM "AiSignal" s
			           WHERE s."entityType" = 'LEADER_INVOICE' AND s."entityId" = li.id AND s.kind = 'INVOICE_ADJUSTED'
			       ) AS has_signal
			FROM "LeaderInvoice" li
			JOIN "Booking" b ON b.id = li."bookingId"
		) sub
		WHERE axis_value IS NOT NULL AND axis_value <> ''
		GROUP BY axis_value`, groupExpr)
	err := r.db.Select(&rows, query)
	return rows, err
}

// InvoiceAdjustRateByEmployee — دقيقة هنا بلا أي جدل: `LeaderInvoice.
// employeeId` هو الليدر صاحب الفاتورة فعلياً، بلا حاجة لأي استنتاج.
func (r *AiDiscoveryRepository) InvoiceAdjustRateByEmployee() ([]AxisCell, error) {
	rows := []AxisCell{}
	err := r.db.Select(&rows, `
		SELECT li."employeeId" AS scope_id,
		       COUNT(*) FILTER (WHERE EXISTS (
		           SELECT 1 FROM "AiSignal" s
		           WHERE s."entityType" = 'LEADER_INVOICE' AND s."entityId" = li.id AND s.kind = 'INVOICE_ADJUSTED'
		       )) AS numerator,
		       COUNT(*) AS denominator
		FROM "LeaderInvoice" li
		GROUP BY li."employeeId"`)
	return rows, err
}

// DisciplineRateByBookingAxis نسبة الحجوزات الي جرّت مخالفة انضباطية
// حقيقية (خصم نقاط، `delta < 0`) مربوطة بحجز، لكل قيمة محور.
func (r *AiDiscoveryRepository) DisciplineRateByBookingAxis(axis string) ([]AxisCell, error) {
	groupExpr, ok := bookingAxisGroupExpr(axis)
	if !ok {
		return nil, fmt.Errorf("محور غير معروف: %s", axis)
	}
	rows := []AxisCell{}
	query := fmt.Sprintf(`
		SELECT axis_value AS scope_id,
		       COUNT(*) FILTER (WHERE has_violation) AS numerator,
		       COUNT(*) AS denominator
		FROM (
			SELECT b.id, (%s) AS axis_value,
			       EXISTS (
			           SELECT 1 FROM "DisciplineEvent" de
			           WHERE de."bookingId" = b.id AND de.delta < 0
			       ) AS has_violation
			FROM "Booking" b
			WHERE %s
		) sub
		WHERE axis_value IS NOT NULL AND axis_value <> ''
		GROUP BY axis_value`, groupExpr, bookingUniverseWhere)
	err := r.db.Select(&rows, query)
	return rows, err
}

// DisciplineRateByEmployee — ⚠️ تبسيط مقصود: نعتمد `DisciplineEvent.
// employeeId` (المخالِف الحقيقي) كبسط، ونقارنه بمجموع الحجوزات الي
// قادها هذا الموظف كمقام (`leaderLateralJoin`). المخالفات المربوطة
// بحجز غالباً مخالفة الليدر أصلاً (تأخر، توقف)، فالتقريب معقول —
// وأي انحراف يطلع هنا يستاهل مراجعة إنسان بغض النظر.
func (r *AiDiscoveryRepository) DisciplineRateByEmployee() ([]AxisCell, error) {
	denomRows := []struct {
		ID    string `db:"id"`
		Total int    `db:"total"`
	}{}
	query := fmt.Sprintf(`
		SELECT ldr.id AS id, COUNT(*) AS total
		FROM "Booking" b
		%s
		WHERE %s
		GROUP BY ldr.id`, leaderLateralJoin, bookingUniverseWhere)
	if err := r.db.Select(&denomRows, query); err != nil {
		return nil, err
	}
	numRows := []struct {
		EmployeeID string `db:"employeeId"`
		N          int    `db:"n"`
	}{}
	if err := r.db.Select(&numRows, `
		SELECT "employeeId", COUNT(*) AS n
		FROM "DisciplineEvent"
		WHERE delta < 0 AND "bookingId" IS NOT NULL
		GROUP BY "employeeId"`); err != nil {
		return nil, err
	}
	numByEmployee := map[string]int{}
	for _, n := range numRows {
		numByEmployee[n.EmployeeID] = n.N
	}
	cells := make([]AxisCell, 0, len(denomRows))
	for _, d := range denomRows {
		cells = append(cells, AxisCell{ScopeID: d.ID, Numerator: numByEmployee[d.ID], Denominator: d.Total})
	}
	return cells, nil
}
