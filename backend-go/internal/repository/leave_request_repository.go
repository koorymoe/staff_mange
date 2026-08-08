package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// LeaveRequestRepository طلبات الإجازة وموافقاتها.
type LeaveRequestRepository struct {
	db *sqlx.DB
}

func NewLeaveRequestRepository(db *sqlx.DB) *LeaveRequestRepository {
	return &LeaveRequestRepository{db: db}
}

const leaveSelect = `SELECT l.*, e.name AS "employeeName", e.role::text AS "employeeRole",
		e.shift::text AS "employeeShift", e."jobTitle", d.name AS "decidedByName",
		pb.name AS "preliminaryByName"
	FROM "LeaveRequest" l
	JOIN "Employee" e ON e.id = l."employeeId"
	LEFT JOIN "Employee" d ON d.id = l."decidedById"
	LEFT JOIN "Employee" pb ON pb.id = l."preliminaryById"`

func decorateLeaves(rows []model.LeaveRequest) []model.LeaveRequest {
	for i := range rows {
		rows[i].RouteLabel = model.LeaveRouteLabels[rows[i].Route]
		rows[i].StatusLabel = model.LeaveStatusLabels[rows[i].Status]
		days := int(rows[i].EndDate.Sub(rows[i].StartDate).Hours()/24) + 1
		if days < 1 {
			days = 1
		}
		rows[i].Days = days
	}
	return rows
}

// Create يقدّم طلب إجازة. المسار يتحدد من بيانات الموظف بقاعدة البيانات،
// مو من الي يرسله العميل — حتى ما يوجّه طلبه لأسهل واحد يوافق.
func (r *LeaveRequestRepository) Create(employeeID string, start, end time.Time, reason *string) (*model.LeaveRequest, error) {
	var shift *string
	if err := r.db.Get(&shift, `SELECT shift::text FROM "Employee" WHERE id = $1`, employeeID); err != nil {
		return nil, fmt.Errorf("تعذر قراءة بيانات الموظف")
	}
	route := model.LeaveRouteFor(shift)

	var id string
	if err := r.db.Get(&id, `
		INSERT INTO "LeaveRequest" (id, "employeeId", "startDate", "endDate", reason, route)
		VALUES (gen_random_uuid()::text, $1, $2, $3, NULLIF($4,''), $5)
		RETURNING id`, employeeID, start, end, derefStr(reason), route); err != nil {
		return nil, err
	}
	return r.Get(id)
}

func (r *LeaveRequestRepository) Get(id string) (*model.LeaveRequest, error) {
	var l model.LeaveRequest
	if err := r.db.Get(&l, leaveSelect+` WHERE l.id = $1`, id); err != nil {
		return nil, err
	}
	out := decorateLeaves([]model.LeaveRequest{l})
	return &out[0], nil
}

// ListForEmployee طلبات موظف واحد — سجله وأرشيفه.
func (r *LeaveRequestRepository) ListForEmployee(employeeID string) ([]model.LeaveRequest, error) {
	rows := []model.LeaveRequest{}
	err := r.db.Select(&rows, leaveSelect+` WHERE l."employeeId" = $1 ORDER BY l."createdAt" DESC`, employeeID)
	return decorateLeaves(rows), err
}

// ListForApprover الطلبات الي يحق لهذا الشخص يبت بيها.
//
// routes هي المسارات الي يغطيها حسب صلاحياته. المالك يمرّر كل المسارات.
// status فاضي = الكل (للأرشيف)، وإلا نفلتر (PENDING للي ينتظر بته).
func (r *LeaveRequestRepository) ListForApprover(routes []string, status string) ([]model.LeaveRequest, error) {
	rows := []model.LeaveRequest{}
	if len(routes) == 0 {
		return rows, nil
	}
	// الملغى ما يهم المخوّل — الموظف سحبه قبل ما يوصل لقراره، فوجوده
	// بصندوق الموافقات ضجيج بس.
	q := leaveSelect + ` WHERE l.route = ANY($1) AND l.status <> 'CANCELLED'`
	args := []any{pq.Array(routes)}
	// status=OPEN اختصار للي لسّه يحتاج قرار (جديد + موافقة أولية)
	if status == "OPEN" {
		args = append(args, pq.Array(model.LeaveOpenStatuses()))
		q += fmt.Sprintf(` AND l.status = ANY($%d)`, len(args))
	} else if status != "" {
		args = append(args, status)
		q += fmt.Sprintf(` AND l.status = $%d`, len(args))
	}
	// المعلّق أولاً وبالأقرب تاريخاً — هذا الي يحتاج قرار اليوم
	// المفتوح أولاً (جديد أو موافقة أولية) وبالأقرب تاريخاً — هذا الي
	// يحتاج قرار اليوم. الموافقة الأولية تبقى فوق حتى ما تنتسى.
	q += ` ORDER BY (l.status IN ('PENDING','PRELIMINARY')) DESC,
		(l.status = 'PRELIMINARY') DESC, l."startDate" ASC, l."createdAt" DESC LIMIT 500`
	err := r.db.Select(&rows, q, args...)
	return decorateLeaves(rows), err
}

// Decide الموافقة أو الرفض.
//
// شرط status='PENDING' داخل التحديث يمنع البت بنفس الطلب مرتين، ولا
// يخلي طلب ملغى يترجع للحياة.
// Decide القرار النهائي. ينقبل من PENDING أو من PRELIMINARY — يعني
// المدير يكدر يبت مباشرة، أو ينطي موافقة أولية أول وبعدين يأكد.
func (r *LeaveRequestRepository) Decide(id string, approve bool, note *string, byID string) (*model.LeaveRequest, error) {
	status := model.LeaveStatusRejected
	if approve {
		status = model.LeaveStatusApproved
	}
	var updated string
	if err := r.db.Get(&updated, `
		UPDATE "LeaveRequest"
		SET status = $2, "decidedById" = $3, "decidedAt" = now(), "decisionNote" = NULLIF($4,'')
		WHERE id = $1 AND status IN ('PENDING', 'PRELIMINARY')
		RETURNING id`, id, status, byID, derefStr(note)); err != nil {
		return nil, fmt.Errorf("الطلب مو موجود أو انبتّ بيه من قبل")
	}
	return r.Get(updated)
}

// Preliminary موافقة أولية — مو قرار نهائي.
//
// الطلب يبقى مفتوح بصندوق المدير بعدها (LeaveOpenStatuses تشمل
// PRELIMINARY)، لأنه لو انشال المدير ينساه والموظف ينظلم.
func (r *LeaveRequestRepository) Preliminary(id string, note *string, byID string) (*model.LeaveRequest, error) {
	var updated string
	if err := r.db.Get(&updated, `
		UPDATE "LeaveRequest"
		SET status = 'PRELIMINARY', "preliminaryById" = $2, "preliminaryAt" = now(),
			"preliminaryNote" = NULLIF($3,'')
		WHERE id = $1 AND status = 'PENDING'
		RETURNING id`, id, byID, derefStr(note)); err != nil {
		return nil, fmt.Errorf("الطلب مو بانتظار الموافقة — يمكن انبتّ بيه من قبل")
	}
	return r.Get(updated)
}

// Cancel الموظف يسحب طلبه — قبل ما ينبت بيه بس.
func (r *LeaveRequestRepository) Cancel(id, employeeID string) error {
	var got string
	err := r.db.Get(&got, `
		UPDATE "LeaveRequest" SET status = 'CANCELLED'
		WHERE id = $1 AND "employeeId" = $2 AND status = 'PENDING'
		RETURNING id`, id, employeeID)
	if err != nil {
		return fmt.Errorf("ما تكدر تلغي هذا الطلب")
	}
	return nil
}

// OwnerOf منو مقدّم الطلب — نحتاجها حتى نمنعه يبت بطلبه هو.
func (r *LeaveRequestRepository) OwnerOf(id string) (string, error) {
	var employeeID string
	err := r.db.Get(&employeeID, `SELECT "employeeId" FROM "LeaveRequest" WHERE id = $1`, id)
	return employeeID, err
}

// RouteOf مسار طلب معيّن — نحتاجه حتى نتأكد إن الي يبت بيه مخوّل لهذا المسار.
func (r *LeaveRequestRepository) RouteOf(id string) (string, error) {
	var route string
	err := r.db.Get(&route, `SELECT route FROM "LeaveRequest" WHERE id = $1`, id)
	return route, err
}

// PendingCountFor عدد الطلبات المنتظرة قرار هذا الشخص — للشارة بالقائمة.
func (r *LeaveRequestRepository) PendingCountFor(routes []string) (int, error) {
	if len(routes) == 0 {
		return 0, nil
	}
	var n int
	// الموافقة الأولية لسّه تحتاج قرار — تنعد بالشارة مثل الجديدة
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "LeaveRequest"
		WHERE status = ANY($2) AND route = ANY($1)`,
		pq.Array(routes), pq.Array(model.LeaveOpenStatuses()))
	return n, err
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
