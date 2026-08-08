package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type QualityFollowUpRepository struct {
	db *sqlx.DB
}

func NewQualityFollowUpRepository(db *sqlx.DB) *QualityFollowUpRepository {
	return &QualityFollowUpRepository{db: db}
}

func (r *QualityFollowUpRepository) hydrate(q *model.QualityFollowUp) {
	var booking model.Booking
	if err := r.db.Get(&booking, `SELECT * FROM "Booking" WHERE id = $1`, q.BookingID); err == nil {
		q.Booking = &booking
	}
	var customer model.Customer
	if err := r.db.Get(&customer, `SELECT * FROM "Customer" WHERE id = $1`, q.CustomerID); err == nil {
		q.Customer = &customer
	}
	if q.ContactedByEmployeeID != nil {
		var brief model.EmployeeBrief
		if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, *q.ContactedByEmployeeID); err == nil {
			q.ContactedByEmployee = &brief
		}
	}
	q.Financials = r.loadFinancials(q.BookingID)
	// تفاصيل التنفيذ: منو طلع، ومتى بدا وخلّص. مهندس الجودة يحتاجها
	// قبل ما يتصل حتى يسأل سؤال محدد بدل سؤال عام.
	q.Execution = r.ExecutionOf(q.BookingID)
	if q.InspectedByID != nil {
		var b model.EmployeeBrief
		if err := r.db.Get(&b, `SELECT id, name FROM "Employee" WHERE id = $1`, *q.InspectedByID); err == nil {
			q.InspectedBy = &b
		}
	}
	if q.PenalizedEmployeeID != nil {
		var b model.EmployeeBrief
		if err := r.db.Get(&b, `SELECT id, name FROM "Employee" WHERE id = $1`, *q.PenalizedEmployeeID); err == nil {
			q.PenalizedEmployee = &b
		}
	}
}

// loadFinancials يجمع تفاصيل الحجز ومشروعه المرتبط وأمواله باستعلام واحد.
//
// الفارق يُحسب هنا بالسيرفر (مو بالواجهة) حتى يبقى تعريف واحد للفارق بكل
// الشاشات — الفارق = المتفق عليه ناقص (العربون + المستلم).
func (r *QualityFollowUpRepository) loadFinancials(bookingID string) *model.QualityFollowUpFinancials {
	var f model.QualityFollowUpFinancials
	err := r.db.Get(&f, `
		SELECT
			b.code                       AS "bookingCode",
			s.name                       AS "serviceName",
			b.address                    AS "location",
			b.notes                      AS "workDetails",
			p.id                         AS "projectId",
			p.code                       AS "projectCode",
			p.name                       AS "projectName",
			p.stage                      AS "projectStage",
			b."quotedPrice"              AS "quotedPrice",
			p.price                      AS "projectPrice",
			b."advancePaid"              AS "advancePaid",
			b."amountCollected"          AS "amountCollected"
		FROM "Booking" b
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		LEFT JOIN "Project" p ON p."bookingId" = b.id
		WHERE b.id = $1`, bookingID)
	if err != nil {
		return nil
	}
	// السعر المعتمد: سعر المشروع إذا الحجز انرحّل لمشروع، وإلا السعر المتفق بالحجز
	if f.ProjectPrice != nil && *f.ProjectPrice > 0 {
		f.AgreedTotal = *f.ProjectPrice
	} else if f.QuotedPrice != nil {
		f.AgreedTotal = *f.QuotedPrice
	}
	if f.AdvancePaid != nil {
		f.ReceivedTotal += *f.AdvancePaid
	}
	if f.AmountCollected != nil {
		f.ReceivedTotal += *f.AmountCollected
	}
	f.Difference = f.AgreedTotal - f.ReceivedTotal
	return &f
}

// CreateForBooking تنشئ سطر متابعة جودة تلقائياً لحجز اكتمل (idempotent — ما تكرر لو
// تكرر استدعاء Complete على نفس الحجز بالخطأ).
func (r *QualityFollowUpRepository) CreateForBooking(bookingID, customerID string) error {
	_, err := r.db.Exec(`
		INSERT INTO "QualityFollowUp" (id, "bookingId", "customerId")
		VALUES (gen_random_uuid()::text, $1, $2)
		ON CONFLICT ("bookingId") DO NOTHING
	`, bookingID, customerID)
	return err
}

func (r *QualityFollowUpRepository) List() ([]model.QualityFollowUp, error) {
	items := []model.QualityFollowUp{}
	if err := r.db.Select(&items, `SELECT * FROM "QualityFollowUp" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range items {
		r.hydrate(&items[i])
	}
	return items, nil
}

func (r *QualityFollowUpRepository) Update(id, status string, contactedByEmployeeID string, contactNotes *string) (*model.QualityFollowUp, error) {
	var q model.QualityFollowUp
	err := r.db.Get(&q, `
		UPDATE "QualityFollowUp" SET
			status = $2,
			"contactNotes" = COALESCE($3, "contactNotes"),
			"contactedByEmployeeId" = $4,
			"contactedAt" = now()
		WHERE id = $1
		RETURNING *
	`, id, status, contactNotes, contactedByEmployeeID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&q)
	return &q, nil
}

// ══════════════════════════════════════════════════════════════════
// حكم الجودة والكشف
// ══════════════════════════════════════════════════════════════════

// qualityKpiPoints كم نقطة تنخصم على شكوى زبون مثبتة.
//
// نقطة وحدة: هي معيار «شكوى الزبائن» من معايير الكي بي اي الثمانية.
// ما نخصم أكثر — الشكوى الوحدة ما تلغي تقييم الموظف كله.
const qualityKpiPoints = 1

// LeaderOf يرجّع **الليدر** المسؤول عن حجز — ولا أحد غيره.
//
// ⚠️ ماكو بديل: إذا الكادر ما بيه ليدر، نرجّع فاضي وما ينغرم أحد.
// كان عندنا بديل ياخذ TECH_1، وهذا غلط — الفني ما يتحمّل مسؤولية
// شغل الفريق قدام الزبون، وهاي مسؤولية الليدر حصراً بقرار صاحب
// العمل. الأفضل ما ينغرم أحد على إن ينغرم واحد مو مسؤول.
//
// والحجز الي بلا ليدر ما تروح شكواه بالهوا: الإدارة تنبّه بإشعار
// حتى تعرف إن أكو شكوى بلا مسؤول وتحدد المسؤول بنفسها.
func (r *QualityFollowUpRepository) LeaderOf(bookingID string) (string, error) {
	var id string
	err := r.db.Get(&id, `
		SELECT a."employeeId"
		FROM "BookingAssignment" a
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."bookingId" = $1 AND e."isLeader" = true
		ORDER BY a.role ASC
		LIMIT 1`, bookingID)
	return id, err
}

// ExecutionOf تفاصيل تنفيذ الحجز — منو طلع، ومتى بدا وخلّص، وشكد استغرق.
func (r *QualityFollowUpRepository) ExecutionOf(bookingID string) *model.BookingExecutionDetail {
	var row struct {
		StartedAt       *time.Time `db:"startedAt"`
		CompletedAt     *time.Time `db:"completedAt"`
		CompletionNotes *string    `db:"completionNotes"`
		WorkStoppedAt   *time.Time `db:"workStoppedAt"`
		WorkStopReason  *string    `db:"workStopReason"`
	}
	if err := r.db.Get(&row, `
		SELECT "startedAt", "completedAt", "completionNotes", "workStoppedAt", "workStopReason"
		FROM "Booking" WHERE id = $1`, bookingID); err != nil {
		return nil
	}

	d := &model.BookingExecutionDetail{
		StartedAt:       row.StartedAt,
		CompletedAt:     row.CompletedAt,
		CompletionNotes: row.CompletionNotes,
		WorkStoppedAt:   row.WorkStoppedAt,
		WorkStopReason:  row.WorkStopReason,
		Crew:            []model.BookingCrewMember{},
		ProgressReports: []model.BookingProgressReport{},
	}
	// المدة تنحسب هنا بالسيرفر حتى تكون وحدة بكل مكان بدل ما كل شاشة
	// تحسبها بطريقتها.
	if row.StartedAt != nil && row.CompletedAt != nil {
		m := int(row.CompletedAt.Sub(*row.StartedAt).Minutes())
		if m >= 0 {
			d.DurationMinutes = &m
		}
	}

	_ = r.db.Select(&d.Crew, `
		SELECT a."employeeId", e.name, a.role::text AS role, e."isLeader"
		FROM "BookingAssignment" a
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."bookingId" = $1
		ORDER BY e."isLeader" DESC, a.role ASC`, bookingID)

	_ = r.db.Select(&d.ProgressReports, `
		SELECT * FROM "BookingProgressReport" WHERE "bookingId" = $1 ORDER BY "dayNumber" ASC`, bookingID)

	return d
}

// Verdict يسجّل حكم مهندس الجودة بعد ما يتصل بالزبون.
//
// كل شي بمعاملة وحدة: لو انسجّل الحكم وما انخصمت النقطة، الليدر يبقى
// نظيف بالسجل والشكوى بلا أثر. ولو انعكست، تنزل غرامة بلا سبب مكتوب.
//
// يرجّع معرّف الموظف الي انغرم (فاضي إذا ما انغرم أحد).
func (r *QualityFollowUpRepository) Verdict(id, byEmployeeID string, req model.QualityVerdictRequest) (penalizedID string, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	var cur struct {
		BookingID  string  `db:"bookingId"`
		ReportType *string `db:"reportType"`
	}
	if err = tx.Get(&cur, `SELECT "bookingId", "reportType" FROM "QualityFollowUp" WHERE id = $1`, id); err != nil {
		return "", errors.New("متابعة الجودة مو موجودة")
	}
	// ⚠️ الحكم ينكتب مرة وحدة. بدون هذا الفحص، ضغطتين على «تقرير سلبي»
	// تخصمن نقطتين من الليدر — ولا وحدة منهن غلط بنظر النظام. انكشفت
	// بفحص حي (انخصمت -٢ بدل -١).
	if cur.ReportType != nil && *cur.ReportType != "" {
		return "", errors.New("انبتّ بهاي المتابعة من قبل — ما تنعاد")
	}
	bookingID := cur.BookingID

	status := "CONTACTED_OK"
	inspection := "NONE"
	if req.ReportType == "NEGATIVE" {
		status = "CONTACTED_ISSUE"
		if req.NeedsInspection {
			inspection = "PENDING"
		}
	}

	// ── الخصم ──
	// ينزل فوراً بالتقرير السلبي، إلا إذا المهندس طلب كشف — وقتها
	// ينتظر نتيجة الكشف، لأن الزبون ممكن يكون يجذب.
	var kpiID *string
	var penalized *string
	if req.ReportType == "NEGATIVE" && !req.NeedsInspection {
		leaderID, lerr := r.leaderOfTx(tx, bookingID)
		if lerr == nil && leaderID != "" {
			newID, derr := r.deductQualityPointTx(tx, leaderID, byEmployeeID, req.Notes)
			if derr != nil {
				return "", derr
			}
			kpiID = &newID
			penalized = &leaderID
		}
	}

	if _, err = tx.Exec(`
		UPDATE "QualityFollowUp" SET
			status = $2,
			"reportType" = $3,
			"inspectionStatus" = $4,
			"contactNotes" = NULLIF($5,''),
			"contactedByEmployeeId" = $6,
			"contactedAt" = now(),
			"kpiEvaluationId" = $7,
			"penalizedEmployeeId" = $8
		WHERE id = $1`, id, status, req.ReportType, inspection, req.Notes, byEmployeeID, kpiID, penalized); err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}
	if penalized != nil {
		return *penalized, nil
	}
	return "", nil
}

// Inspect نتيجة الكشف الميداني على شكوى موقوفة.
//
//	كلام الزبون صح  → تنزل الغرامة الحين
//	كلام الزبون كذب → ما ينغرم أحد، وينزاد عداد الشكاوى الكاذبة للزبون
func (r *QualityFollowUpRepository) Inspect(id, byEmployeeID string, req model.QualityInspectionRequest) (penalizedID string, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	var row struct {
		BookingID  string `db:"bookingId"`
		CustomerID string `db:"customerId"`
		Inspection string `db:"inspectionStatus"`
	}
	if err = tx.Get(&row, `
		SELECT "bookingId", "customerId", "inspectionStatus"
		FROM "QualityFollowUp" WHERE id = $1`, id); err != nil {
		return "", errors.New("متابعة الجودة مو موجودة")
	}
	if row.Inspection != "PENDING" {
		return "", errors.New("هذي المتابعة ما تنتظر كشف")
	}

	var kpiID *string
	var penalized *string

	if req.Result == "CUSTOMER_RIGHT" {
		leaderID, lerr := r.leaderOfTx(tx, row.BookingID)
		if lerr == nil && leaderID != "" {
			newID, derr := r.deductQualityPointTx(tx, leaderID, byEmployeeID,
				"شكوى زبون تأكدت بالكشف: "+req.Notes)
			if derr != nil {
				return "", derr
			}
			kpiID = &newID
			penalized = &leaderID
		}
	} else {
		// الزبون كان يجذب — العلامة تروح عليه هو، والكادر يبقى نظيف.
		if _, err = tx.Exec(`
			UPDATE "Customer" SET
				"falseClaimCount" = "falseClaimCount" + 1,
				"lastFalseClaimAt" = now(),
				"falseClaimNote" = NULLIF($2,'')
			WHERE id = $1`, row.CustomerID, req.Notes); err != nil {
			return "", err
		}
	}

	if _, err = tx.Exec(`
		UPDATE "QualityFollowUp" SET
			"inspectionStatus" = 'DONE',
			"inspectionResult" = $2,
			"inspectionNotes" = NULLIF($3,''),
			"inspectedById" = $4,
			"inspectedAt" = now(),
			"kpiEvaluationId" = COALESCE($5, "kpiEvaluationId"),
			"penalizedEmployeeId" = COALESCE($6, "penalizedEmployeeId")
		WHERE id = $1`, id, req.Result, req.Notes, byEmployeeID, kpiID, penalized); err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}
	if penalized != nil {
		return *penalized, nil
	}
	return "", nil
}

// leaderOfTx نفس LeaderOf بس داخل المعاملة — الليدر حصراً، بلا بديل.
func (r *QualityFollowUpRepository) leaderOfTx(tx *sqlx.Tx, bookingID string) (string, error) {
	var id string
	err := tx.Get(&id, `
		SELECT a."employeeId"
		FROM "BookingAssignment" a
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."bookingId" = $1 AND e."isLeader" = true
		ORDER BY a.role ASC
		LIMIT 1`, bookingID)
	return id, err
}

// deductQualityPointTx يخصم نقطة معيار «شكوى الزبائن» من الليدر.
//
// نستعمل نفس جدول التقييمات الموجود (KpiEvaluation) مو جدول جديد —
// حتى الخصم يطلع بتقييم الموظف الشهري وبترتيبه متل أي تقييم ثاني،
// وينلغى بنفس الطريقة لو انكشف إنه غلط.
func (r *QualityFollowUpRepository) deductQualityPointTx(tx *sqlx.Tx, employeeID, evaluatorID, notes string) (string, error) {
	reason := model.QualityKpiCriterion
	if strings.TrimSpace(notes) != "" {
		reason += ": " + strings.TrimSpace(notes)
	}
	var id string
	err := tx.Get(&id, `
		INSERT INTO "KpiEvaluation" (id, "employeeId", "evaluatorId", points, reason, "deductionAmount")
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		uuid.NewString(), employeeID, evaluatorID, -qualityKpiPoints, reason,
		float64(qualityKpiPoints)*10000)
	return id, err
}
