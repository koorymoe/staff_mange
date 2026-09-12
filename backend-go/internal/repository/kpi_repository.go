package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type KpiRepository struct {
	db *sqlx.DB
}

func NewKpiRepository(db *sqlx.DB) *KpiRepository {
	return &KpiRepository{db: db}
}

func (r *KpiRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *KpiRepository) hydrate(e *model.KpiEvaluation) {
	e.Employee = r.loadEmployeeBrief(e.EmployeeID)
	e.Evaluator = r.loadEmployeeBrief(e.EvaluatorID)
	if e.CancelledByEmployeeID != nil {
		e.CancelledByEmployee = r.loadEmployeeBrief(*e.CancelledByEmployeeID)
	}
}

func (r *KpiRepository) List() ([]model.KpiEvaluation, error) {
	evals := []model.KpiEvaluation{}
	if err := r.db.Select(&evals, `SELECT * FROM "KpiEvaluation" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range evals {
		r.hydrate(&evals[i])
	}
	return evals, nil
}

func (r *KpiRepository) ListForEmployee(employeeID string) ([]model.KpiEvaluation, error) {
	evals := []model.KpiEvaluation{}
	if err := r.db.Select(&evals, `SELECT * FROM "KpiEvaluation" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID); err != nil {
		return nil, err
	}
	for i := range evals {
		r.hydrate(&evals[i])
	}
	return evals, nil
}

func (r *KpiRepository) Create(employeeID, evaluatorID string, points int, reason string, deductionAmount float64) (*model.KpiEvaluation, error) {
	var e model.KpiEvaluation
	err := r.db.Get(&e, `
		INSERT INTO "KpiEvaluation" (id, "employeeId", "evaluatorId", points, reason, "deductionAmount")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, employeeID, evaluatorID, points, reason, deductionAmount)
	if err != nil {
		return nil, err
	}
	r.hydrate(&e)
	return &e, nil
}

func (r *KpiRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "KpiEvaluation" WHERE id = $1`, id)
	return err
}

// Cancel "يرجّع" نقطة كي بي اي — ما تنحذف نهائياً، تنعلّم "ملغاة" حتى يضل
// تاريخها موجود ويشوفه المراقب، بس تأثيرها المالي يوقف يحسب بالمجاميع.
func (r *KpiRepository) Cancel(id, cancelledByEmployeeID string) (*model.KpiEvaluation, error) {
	var e model.KpiEvaluation
	err := r.db.Get(&e, `
		UPDATE "KpiEvaluation" SET
			cancelled = true,
			"cancelledAt" = now(),
			"cancelledByEmployeeId" = $2
		WHERE id = $1
		RETURNING *
	`, id, cancelledByEmployeeID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&e)
	return &e, nil
}

// RoleLeaderboard يرجع ترتيب موظفي دور معيّن حسب مجموع نقاط الـKPI ضمن فترة زمنية،
// حتى يشوف كل موظف ترتيبه بين نظرائه بنفس الدور (فني مع الفنيين، إداري مع الإداريين).
// SumPointsForEmployeeMonth يرجّع مجموع نقاط الكي بي اي (غير الملغاة) لموظف
// معيّن خلال شهر معيّن (monthPrefix بصيغة "YYYY-MM") — يُستخدم بصفحة إحصائيات
// الموظفين الشهرية لإعادة استخدام نفس آلية تسجيل النقاط الموجودة أصلاً.
func (r *KpiRepository) SumPointsForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var total int
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM(points), 0) FROM "KpiEvaluation"
		WHERE "employeeId" = $1 AND cancelled = false AND to_char("createdAt", 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return total, err
}

// SumPointsForEmployeeRange نفس SumPointsForEmployeeMonth لكن لمدى تاريخ حر
// (from/to بصيغة "YYYY-MM-DD") — تُستخدم بالإحصائية الأسبوعية بفلتر التاريخ.
func (r *KpiRepository) SumPointsForEmployeeRange(employeeID, from, to string) (int, error) {
	var total int
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM(points), 0) FROM "KpiEvaluation"
		WHERE "employeeId" = $1 AND cancelled = false AND baghdad_date("createdAt") BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return total, err
}

// MonthlyPointsSeriesForEmployee يرجّع مجموع نقاط الكي بي اي (غير الملغاة) لكل
// شهر من monthsCount شهر تنتهي بشهر endMonth (بصيغة "YYYY-MM-01"، بالترتيب من
// الأقدم للأحدث) — يُستخدم بمنحنى الأداء المتحرك بصفحة إحصائيات الموظفين،
// يقدر المستخدم يتصفح أي شهر بالفلتر بدل ما يبقى محصور بآخر 6 أشهر ثابتة.
func (r *KpiRepository) MonthlyPointsSeriesForEmployee(employeeID string, monthsCount int, endMonth string) ([]model.MonthlyPointsBucket, error) {
	buckets := []model.MonthlyPointsBucket{}
	err := r.db.Select(&buckets, `
		SELECT to_char(m, 'YYYY-MM') AS month, COALESCE(SUM(k.points), 0) AS points
		FROM generate_series(
			$3::date - ($2 - 1) * interval '1 month',
			$3::date,
			interval '1 month'
		) m
		LEFT JOIN "KpiEvaluation" k ON k."employeeId" = $1 AND k.cancelled = false
			AND to_char(k."createdAt", 'YYYY-MM') = to_char(m, 'YYYY-MM')
		GROUP BY m
		ORDER BY m
	`, employeeID, monthsCount, endMonth)
	return buckets, err
}

// ═══ الترتيب حسب الشغل مو حسب المسمّى ═══
//
// المشكلة الي كانت: الترتيب ينبني على e.role. يعني الموظف الي دوره
// «محاسب» بس ينطوه صلاحيات إدارة الكوادر وتنسيق الحجوزات — يشتغل
// شغل المنسّقين كل يوم — **ما يظهر بتصنيفهم أبداً**. ينقارن بمحاسبين
// ما يشتغلون شغله.
//
// النظام يحكم بالاسم المكتوب بملفه، مو بالشغل الي يسويه فعلاً. وهاي
// تخلي التصنيف كله يكذب: منسّق شاطر ما يطلع بالقائمة، ومحاسب ما
// يمسّ التنسيق يتصدّرها.
//
// الحل: نجمّع بالصلاحية. منو عنده صلاحية «تنسيق الحجوزات» ينقارن
// بمنسّقي الحجوزات — مهما كان اسم دوره.
//
// ⚠️ والموظف يظهر بأكثر من تصنيف إذا يشتغل أكثر من شغلة. هذا مو خلل
// — هذا بالضبط واقعه، والتصنيف الواحد كان يخفيه.
func (r *KpiRepository) PermissionLeaderboard(permission string, since, until string) ([]model.KpiLeaderboardEntry, error) {
	entries := []model.KpiLeaderboardEntry{}
	err := r.db.Select(&entries, `
		WITH holders AS (
			-- ⚠️ OWNER و ADMIN مستثنون: عندهم كل الصلاحيات بحكم
			-- موقعهم، فيطلعون بكل تصنيف ويزاحمون الي يشتغل الشغل
			-- فعلاً. التصنيف للي ينفّذ، مو للي يملك الوصول.
			SELECT DISTINCT e.id, e.name
			FROM "Employee" e
			JOIN "EmployeePermission" ep ON ep."employeeId" = e.id
			JOIN "Permission" p          ON p.id = ep."permissionId"
			WHERE e.status = 'ACTIVE'
			  AND p.name = $1
			  AND e.role NOT IN ('OWNER', 'ADMIN')
		)
		SELECT
			h.id AS "employeeId",
			h.name AS "employeeName",
			COALESCE(SUM(k.points), 0) AS points,
			COUNT(k.id) AS "evaluationCount",
			COALESCE((
				SELECT COUNT(*) FROM "BookingAssignment" ba
				JOIN "Booking" b ON b.id = ba."bookingId"
				WHERE ba."employeeId" = h.id AND b.status = 'COMPLETED'
				  AND b."completedAt" >= $2::timestamp
				  AND ($3 = '' OR b."completedAt" < $3::timestamp)
			), 0) AS "completedBookings",
			COALESCE((
				SELECT COUNT(*) FROM "BookingAssignment" ba
				JOIN "Booking" b ON b.id = ba."bookingId"
				WHERE ba."employeeId" = h.id
				  AND b."createdAt" >= $2::timestamp
				  AND ($3 = '' OR b."createdAt" < $3::timestamp)
				  AND b.status <> 'CANCELLED'
			), 0) AS "assignedBookings",
			COALESCE((
				SELECT COUNT(DISTINCT a.date) FROM "Attendance" a
				WHERE a."employeeId" = h.id
				  AND a.date >= $2::date
				  AND ($3 = '' OR a.date < $3::date)
			), 0) AS "attendedDays"
		FROM holders h
		LEFT JOIN "KpiEvaluation" k
		       ON k."employeeId" = h.id
		      AND k."createdAt" >= $2::timestamp
		      AND ($3 = '' OR k."createdAt" < $3::timestamp)
		      AND k.cancelled = false
		GROUP BY h.id, h.name
		ORDER BY points DESC, "completedBookings" DESC
	`, permission, since, until)
	return entries, err
}

// RoleLeaderboard ترتيب أصحاب دور واحد بفترة محددة.
//
// ⚠️ until مطلوب مو بس since: بدونه ما نكدر نجيب **الفترة السابقة**
// (الأسبوع الي قبله) للمقارنة. والرقم بلا مقارنة ما يگول شي — «٨٦
// نقطة» زين لو خبل؟ السهم الي يگول «+٧ عن الأسبوع الماضي» هو الي
// يخلي الموظف يعرف هل هو يتحسّن لو ينزل.
func (r *KpiRepository) RoleLeaderboard(role string, since, until string) ([]model.KpiLeaderboardEntry, error) {
	entries := []model.KpiLeaderboardEntry{}
	err := r.db.Select(&entries, `
		SELECT
			e.id AS "employeeId",
			e.name AS "employeeName",
			COALESCE(SUM(k.points), 0) AS points,
			COUNT(k.id) AS "evaluationCount",
			COALESCE((
				SELECT COUNT(*) FROM "BookingAssignment" ba
				JOIN "Booking" b ON b.id = ba."bookingId"
				WHERE ba."employeeId" = e.id AND b.status = 'COMPLETED'
				  AND b."completedAt" >= $2::timestamp
				  AND ($3 = '' OR b."completedAt" < $3::timestamp)
			), 0) AS "completedBookings",
			-- كل الحجوزات الي انكلّف بيها بالفترة (مو المنجزة بس) —
			-- بدونها ما نكدر نحسب معدل الإنجاز، والمعدل هو الي يميّز
			-- الي خلّص ٨ من ٨ عن الي خلّص ٨ من ٢٠.
			COALESCE((
				SELECT COUNT(*) FROM "BookingAssignment" ba
				JOIN "Booking" b ON b.id = ba."bookingId"
				WHERE ba."employeeId" = e.id
				  AND b."createdAt" >= $2::timestamp
				  AND ($3 = '' OR b."createdAt" < $3::timestamp)
				  AND b.status <> 'CANCELLED'
			), 0) AS "assignedBookings",
			-- أيام حضور بالفترة: أساس «الالتزام بالدوام»
			COALESCE((
				SELECT COUNT(DISTINCT a.date) FROM "Attendance" a
				WHERE a."employeeId" = e.id
				  AND a.date >= $2::date
				  AND ($3 = '' OR a.date < $3::date)
			), 0) AS "attendedDays"
		FROM "Employee" e
		LEFT JOIN "KpiEvaluation" k
		       ON k."employeeId" = e.id
		      AND k."createdAt" >= $2::timestamp
		      AND ($3 = '' OR k."createdAt" < $3::timestamp)
		      AND k.cancelled = false
		WHERE e.role = $1 AND e.status = 'ACTIVE'
		GROUP BY e.id, e.name
		ORDER BY points DESC, "completedBookings" DESC
	`, role, since, until)
	return entries, err
}
