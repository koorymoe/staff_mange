package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type DisciplineRepository struct {
	db *sqlx.DB
}

func NewDisciplineRepository(db *sqlx.DB) *DisciplineRepository {
	return &DisciplineRepository{db: db}
}

// Penalize يخصم نقاط من موظف ويسجّل سبب الخصم — كلها بمعاملة وحدة حتى
// ما يصير خصم بلا سجل ولا سجل بلا خصم.
//
// لو نفس الغرامة انسجّلت قبل لنفس الحجز ونفس السبب، ما تنعاد: الفهرس
// الفريد يمنعها، وهاي مقصودة — المهمة تشتغل كل شوية وما يصير تخصم كل
// مرة تمر على نفس الحجز المتأخر.
//
// يرجّع applied=false لو الغرامة مكررة (مو خطأ).
func (r *DisciplineRepository) Penalize(employeeID, kind, reason string, bookingID *string, points int) (applied bool, remaining int, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return false, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		INSERT INTO "DisciplineEvent" (id, "employeeId", "bookingId", kind, delta, reason)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`, uuid.NewString(), employeeID, bookingID, kind, -points, reason)
	if err != nil {
		return false, 0, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return false, 0, nil // انسجّلت قبل — ما نخصم مرة ثانية
	}

	// الرصيد ما ينزل تحت الصفر — الغرامة عقوبة مو دين
	var left int
	err = tx.Get(&left, `
		INSERT INTO "DisciplinePoints" ("employeeId", points, "updatedAt")
		VALUES ($1, GREATEST($2::int - $3::int, 0), now())
		ON CONFLICT ("employeeId") DO UPDATE
			SET points = GREATEST("DisciplinePoints".points - $3::int, 0), "updatedAt" = now()
		RETURNING points
	`, employeeID, model.DisciplineStartingPoints, points)
	if err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return true, left, nil
}

// Adjust تعديل يدوي على رصيد موظف — زيادة أو نقصان — من المالك أو مدير
// النظام. الحركة والرصيد ينتحدثون بمعاملة وحدة، ونسجّل منو عدّل.
//
// نفس حدود التلقائي بالضبط: ما ينزل تحت الصفر ولا يفوت ١٠٠. التعديل
// اليدوي ما يعطي صلاحيات أوسع من النظام، يعطي بس القدرة على التصحيح.
//
// يرجّع الرصيد بعد التعديل والفرق الي انطبّق فعلاً (ممكن يقل عن
// المطلوب إذا الرصيد وصل الحد).
func (r *DisciplineRepository) Adjust(employeeID string, delta int, reason, byEmployeeID string) (remaining int, applied int, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	// نقفل الصف ونقرا الرصيد الحالي — بدون القفل، تعديلين بنفس اللحظة
	// يقرون نفس الرصيد وواحد منهم يضيع.
	var before int
	if err = tx.Get(&before, `
		INSERT INTO "DisciplinePoints" ("employeeId", points, "updatedAt")
		VALUES ($1, $2, now())
		ON CONFLICT ("employeeId") DO UPDATE SET "updatedAt" = "DisciplinePoints"."updatedAt"
		RETURNING points
	`, employeeID, model.DisciplineStartingPoints); err != nil {
		return 0, 0, err
	}

	after := before + delta
	if after < 0 {
		after = 0
	}
	if after > model.DisciplineStartingPoints {
		after = model.DisciplineStartingPoints
	}
	applied = after - before
	if applied == 0 {
		return before, 0, nil // الرصيد أصلاً على الحد — ما نسجّل حركة فاضية
	}

	if _, err = tx.Exec(`
		INSERT INTO "DisciplineEvent" (id, "employeeId", "bookingId", kind, delta, reason, "byEmployeeId")
		VALUES ($1, $2, NULL, $3, $4, $5, $6)
	`, uuid.NewString(), employeeID, model.DisciplineManual, applied, reason, byEmployeeID); err != nil {
		return 0, 0, err
	}
	if _, err = tx.Exec(`
		UPDATE "DisciplinePoints" SET points = $2, "updatedAt" = now() WHERE "employeeId" = $1
	`, employeeID, after); err != nil {
		return 0, 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return after, applied, nil
}

// RestoreOne يرجّع نقطة وحدة لموظف اشتغل نظيف. ما يتجاوز الرصيد الأصلي.
func (r *DisciplineRepository) RestoreOne(employeeID, reason string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`
		INSERT INTO "DisciplineEvent" (id, "employeeId", "bookingId", kind, delta, reason)
		VALUES ($1, $2, NULL, $3, 1, $4)
	`, uuid.NewString(), employeeID, model.DisciplineRestore, reason); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE "DisciplinePoints"
		SET points = LEAST(points + 1, $2::int), "lastRestoredAt" = now(), "updatedAt" = now()
		WHERE "employeeId" = $1
	`, employeeID, model.DisciplineStartingPoints); err != nil {
		return err
	}
	return tx.Commit()
}

// List يرجّع أرصدة كل الموظفين الي عندهم سجل — الناقصين أول.
func (r *DisciplineRepository) List() ([]model.DisciplinePoints, error) {
	rows := []struct {
		model.DisciplinePoints
		Name string `db:"name"`
	}{}
	if err := r.db.Select(&rows, `
		SELECT dp.*, e.name
		FROM "DisciplinePoints" dp
		JOIN "Employee" e ON e.id = dp."employeeId"
		ORDER BY dp.points ASC, e.name
	`); err != nil {
		return nil, err
	}
	out := make([]model.DisciplinePoints, 0, len(rows))
	for _, r0 := range rows {
		p := r0.DisciplinePoints
		p.EmployeeName = r0.Name
		p.DeductedDinar = (model.DisciplineStartingPoints - p.Points) * model.DisciplineDinarPerPoint
		out = append(out, p)
	}
	return out, nil
}

// Events يرجّع آخر الحركات — لموظف معيّن أو للكل.
func (r *DisciplineRepository) Events(employeeID string, limit int) ([]model.DisciplineEvent, error) {
	rows := []struct {
		model.DisciplineEvent
		Name string `db:"name"`
	}{}
	query := `
		SELECT de.*, e.name FROM "DisciplineEvent" de
		JOIN "Employee" e ON e.id = de."employeeId"
		WHERE ($1 = '' OR de."employeeId" = $1)
		ORDER BY de."createdAt" DESC LIMIT $2`
	if err := r.db.Select(&rows, query, employeeID, limit); err != nil {
		return nil, err
	}
	out := make([]model.DisciplineEvent, 0, len(rows))
	for _, r0 := range rows {
		e := r0.DisciplineEvent
		e.EmployeeName = r0.Name
		out = append(out, e)
	}
	return out, nil
}

// OverduePaperwork يرجّع الحجوزات المنجزة الي عدّى عليها الوقت المسموح
// بلا فاتورة أو تقرير، مع الإداري الي كلّف الكادر عليها.
//
// «الإداري الي كلّف» ينجاب من BookingAssignment.assignedById — منو ضغط
// زر التعيين فعلاً. لو التعيين قديم (قبل ما نسجّل هذا العمود) ننزل على
// منو ثبّت الحجز، وإذا ما اكو ولا واحد نتجاهل الحجز بدل ما نغرّم واحد
// بالغلط.
type OverduePaperwork struct {
	BookingID   string `db:"bookingId"`
	BookingCode string `db:"code"`
	AdminID     string `db:"adminId"`
	AdminName   string `db:"adminName"`
	HasInvoice  bool   `db:"hasInvoice"`
	HasReport   bool   `db:"hasReport"`
}

func (r *DisciplineRepository) OverduePaperwork(hours int) ([]OverduePaperwork, error) {
	rows := []OverduePaperwork{}
	err := r.db.Select(&rows, `
		SELECT b.id AS "bookingId", b.code,
		       adm.id AS "adminId", adm.name AS "adminName",
		       EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id) AS "hasInvoice",
		       EXISTS (SELECT 1 FROM "WorkReport" wr WHERE wr."bookingId" = b.id) AS "hasReport"
		FROM "Booking" b
		JOIN LATERAL (
			SELECT e.id, e.name
			FROM "Employee" e
			WHERE e.id = COALESCE(
				(SELECT ba."assignedById" FROM "BookingAssignment" ba
				 WHERE ba."bookingId" = b.id AND ba."assignedById" IS NOT NULL
				 ORDER BY ba."createdAt" LIMIT 1),
				b."confirmedByEmployeeId"
			)
			AND e.status = 'ACTIVE'
		) adm ON true
		WHERE b.status = 'COMPLETED'
		  AND b."completedAt" IS NOT NULL
		  AND b."completedAt" < now() - ($1::text || ' hours')::interval
		  -- الغرامات تبدي من تاريخ تشغيل النظام: ما نحاسب أحد على شغل
		  -- قديم ما جان النظام يطالبه بيه أصلاً
		  AND b."completedAt" > (SELECT "startsAt" FROM "DisciplineConfig" WHERE id = 1)
		  AND (
		    NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)
		    OR NOT EXISTS (SELECT 1 FROM "WorkReport" wr WHERE wr."bookingId" = b.id)
		  )
	`, hours)
	return rows, err
}

// CleanSince يرجّع الموظفين الي رصيدهم ناقص وما انغرموا من مدة كافية —
// يستحقون رجوع نقطة.
//
// «نظيف» يعني: ما اكو ولا غرامة عليه من آخر رجوع نقطة (أو من أول
// غرامة إذا ما رجعت له نقطة بعد) لهسه بمدة الأيام المطلوبة.
func (r *DisciplineRepository) EligibleForRestore(days int) ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `
		SELECT dp."employeeId"
		FROM "DisciplinePoints" dp
		WHERE dp.points < $1::int
		  AND COALESCE(dp."lastRestoredAt", '-infinity'::timestamptz) < now() - ($2::text || ' days')::interval
		  AND NOT EXISTS (
		    SELECT 1 FROM "DisciplineEvent" de
		    WHERE de."employeeId" = dp."employeeId"
		      AND de.delta < 0
		      AND de."createdAt" > now() - ($2::text || ' days')::interval
		  )
	`, model.DisciplineStartingPoints, days)
	return ids, err
}

// SystemAuthorID يرجّع معرّف حساب يصلح كـ«كاتب» لإعلانات النظام.
// لوحة الإعلانات تشترط كاتب موجود بجدول الموظفين، والنظام ما عنده
// حساب — فنستعمل حساب المالك (أو مدير النظام إذا ماكو مالك).
func (r *DisciplineRepository) SystemAuthorID() (string, error) {
	var id string
	err := r.db.Get(&id, `
		SELECT id FROM "Employee"
		WHERE role IN ('OWNER', 'ADMIN') AND status = 'ACTIVE'
		ORDER BY CASE role WHEN 'OWNER' THEN 0 ELSE 1 END
		LIMIT 1
	`)
	return id, err
}

// StartsAt يرجّع لحظة تشغيل نظام الغرامات — كل شي انتهى قبلها ما
// ينحاسب عليه.
func (r *DisciplineRepository) StartsAt() (string, error) {
	var t string
	err := r.db.Get(&t, `SELECT "startsAt"::text FROM "DisciplineConfig" WHERE id = 1`)
	return t, err
}

// OverdueAudit الحجوزات المنجزة الي مبلغها ما انتدقّق بعد المهلة.
//
// الغرامة تروح للمحاسب — أي موظف دوره FINANCE أو عنده صلاحية
// «finance». مو لواحد محدد لأن المحاسبة ممكن تكون أكثر من شخص، وكل
// واحد منهم مسؤول عن الطابور نفسه.
//
// ⚠️ ما نغرّم إلا إذا **الفاتورة موجودة**: التدقيق يحتاج فاتورة، ولو
// غرّمناه على حجز بلا فاتورة نغرّمه على تقصير غيره (الإداري ينغرم
// عليها أصلاً بغرامة الورق).
func (r *DisciplineRepository) OverdueAudit(hours int) ([]OverduePaperwork, error) {
	rows := []OverduePaperwork{}
	err := r.db.Select(&rows, `
		SELECT b.id AS "bookingId", b.code,
		       acc.id AS "adminId", acc.name AS "adminName",
		       true AS "hasInvoice", true AS "hasReport"
		FROM "Booking" b
		CROSS JOIN LATERAL (
			SELECT e.id, e.name
			FROM "Employee" e
			WHERE e.status = 'ACTIVE'
			  AND (
			    e.role = 'FINANCE'
			    OR EXISTS (
			      SELECT 1 FROM "EmployeePermission" ep
			      JOIN "Permission" p ON p.id = ep."permissionId"
			      WHERE ep."employeeId" = e.id AND p.name = 'finance'
			    )
			  )
		) acc
		WHERE b.status = 'COMPLETED'
		  AND b."completedAt" IS NOT NULL
		  AND b."amountVerified" = false
		  AND b."completedAt" < now() - ($1::text || ' hours')::interval
		  AND b."completedAt" > (SELECT "startsAt" FROM "DisciplineConfig" WHERE id = 1)
		  -- التدقيق يحتاج فاتورة. بلا فاتورة، التقصير مو تقصير المحاسب.
		  AND EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)
	`, hours)
	return rows, err
}
