package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// BookingProgressRepository الإنجاز الجزئي للحجوزات الي تاخذ أكثر من يوم.
type BookingProgressRepository struct {
	db *sqlx.DB
}

func NewBookingProgressRepository(db *sqlx.DB) *BookingProgressRepository {
	return &BookingProgressRepository{db: db}
}

// PartialComplete يقفل يوم شغل على حجز ما انخلص، ويرجّعه للإداري.
//
// كل شي بمعاملة وحدة: لو انسجّل التقرير وما انتغيرت الحالة، الحجز يبقى
// «قيد التنفيذ» وما يوصل للإداري — يعني الكادر راح والحجز معلّق ومحد
// يدري. ولو انعكست، تنضيع معلومة وين وصلوا.
//
// ⚠️ ما نمسح التكليفات: هي الي تصير «الكادر المقترح» لليوم الجاي.
func (r *BookingProgressRepository) PartialComplete(bookingID, reportedByID string, req model.PartialCompleteRequest) (*model.BookingProgressReport, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// رقم اليوم = الي قبله + ١. نحسبه من التقارير مو من partialCount حتى
	// لو انحذف تقرير يدوياً ما ينكسر الترقيم.
	var dayNumber int
	if err = tx.Get(&dayNumber, `
		SELECT COALESCE(MAX("dayNumber"), 0) + 1 FROM "BookingProgressReport" WHERE "bookingId" = $1
	`, bookingID); err != nil {
		return nil, err
	}

	// لقطة الكادر الي اشتغل اليوم — قبل ما يبدّلهم الإداري باچر.
	crew := []string{}
	if err = tx.Select(&crew, `
		SELECT e.name FROM "BookingAssignment" a
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."bookingId" = $1 ORDER BY a.role
	`, bookingID); err != nil {
		return nil, err
	}
	var crewSnapshot *string
	if len(crew) > 0 {
		s := strings.Join(crew, "، ")
		crewSnapshot = &s
	}

	percent := req.PercentDone
	if percent < 0 {
		percent = 0
	}
	if percent > 99 {
		// ١٠٠٪ يعني خلص — لازم يأشّر «تم الإنجاز» مو «إنجاز جزئي»،
		// وإلا يبقى حجز مفتوح للأبد على شغل منتهي.
		percent = 99
	}

	var report model.BookingProgressReport
	if err = tx.Get(&report, `
		INSERT INTO "BookingProgressReport"
			(id, "bookingId", "dayNumber", "reportedById", "workDone", "remainingWork",
			 "percentDone", blockers, "materialsUsed", "crewSnapshot")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING *
	`, uuid.NewString(), bookingID, dayNumber, reportedByID, req.WorkDone, req.RemainingWork,
		percent, req.Blockers, req.MaterialsUsed, crewSnapshot); err != nil {
		return nil, err
	}

	// الحجز يرجع للإداري: PARTIAL مو COMPLETED ومو CANCELLED. والموعد
	// ينمسح — الإداري هو الي يحدد اليوم الجاي بعد ما يتفق مع الزبون.
	if _, err = tx.Exec(`
		UPDATE "Booking" SET
			status = 'PARTIAL',
			"partialCount" = "partialCount" + 1,
			"lastPartialAt" = now(),
			"amountCollected" = COALESCE($2, "amountCollected"),
			"updatedAt" = now()
		WHERE id = $1
	`, bookingID, req.AmountCollected); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	report.ReportedBy = r.loadBrief(reportedByID)
	return &report, nil
}

func (r *BookingProgressRepository) loadBrief(id string) *model.EmployeeBrief {
	if id == "" {
		return nil
	}
	var b model.EmployeeBrief
	if err := r.db.Get(&b, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &b
}

// Reports كل تقارير حجز مرتبة باليوم — هذا الي يقراه الكادر الجاي حتى
// يعرف وين وصل الي قبله.
func (r *BookingProgressRepository) Reports(bookingID string) ([]model.BookingProgressReport, error) {
	rows := []model.BookingProgressReport{}
	if err := r.db.Select(&rows, `
		SELECT * FROM "BookingProgressReport" WHERE "bookingId" = $1 ORDER BY "dayNumber" ASC
	`, bookingID); err != nil {
		return nil, err
	}
	// جلب الأسماء بدفعة وحدة مو استعلام لكل صف
	ids := make([]string, 0, len(rows))
	for _, x := range rows {
		if x.ReportedByID != nil {
			ids = append(ids, *x.ReportedByID)
		}
	}
	if len(ids) > 0 {
		briefs := []model.EmployeeBrief{}
		q, args, err := sqlx.In(`SELECT id, name FROM "Employee" WHERE id IN (?)`, ids)
		if err == nil {
			if err := r.db.Select(&briefs, r.db.Rebind(q), args...); err == nil {
				byID := map[string]model.EmployeeBrief{}
				for _, b := range briefs {
					byID[b.ID] = b
				}
				for i := range rows {
					if rows[i].ReportedByID != nil {
						if b, ok := byID[*rows[i].ReportedByID]; ok {
							c := b
							rows[i].ReportedBy = &c
						}
					}
				}
			}
		}
	}
	return rows, nil
}

// SuggestedCrew الكادر الي اشتغل على الحجز بالأيام الفائتة — النظام
// يقترحهم لأنهم يعرفون الشغل والزبون والطريق، والإداري إله الحق يبدّل.
//
// نجيبهم من التكليفات الحالية ومن لقطات التقارير سوه؟ لا — من التكليفات
// بس. اللقطة نص أسماء للعرض، والتكليف هو الربط الحقيقي بالموظف.
func (r *BookingProgressRepository) SuggestedCrew(bookingID string) ([]model.SuggestedCrewMember, error) {
	rows := []model.SuggestedCrewMember{}
	err := r.db.Select(&rows, `
		SELECT a."employeeId", e.name, a.role::text AS role,
		       (SELECT COUNT(*) FROM "BookingProgressReport" p WHERE p."bookingId" = a."bookingId") AS "daysWorked"
		FROM "BookingAssignment" a
		JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."bookingId" = $1
		ORDER BY a.role
	`, bookingID)
	if err != nil {
		return nil, err
	}
	// حالة الموظف الحين — الإداري لازم يعرف قبل ما يعتمد الاقتراح إن
	// واحد منهم مأرشف أو موقوف، بدل ما يكتشفها يوم الموعد.
	for i := range rows {
		var status string
		if err := r.db.Get(&status, `SELECT status::text FROM "Employee" WHERE id = $1`, rows[i].EmployeeID); err != nil {
			continue
		}
		rows[i].Available = status == "ACTIVE"
		if !rows[i].Available {
			rows[i].Note = "غير متاح حالياً (" + status + ")"
		}
	}
	return rows, nil
}

// ScheduleContinuation يحدد موعد إكمال حجز منجز جزئياً ويرجّعه للطابور.
//
// ⚠️ ما نستعمل «التأجيل» (Postpone) لهذا: التأجيل يعني الزبون طلب
// تغيير الموعد، وينعد بعداد التأجيلات وينبّه على الزبون الي يأجّل
// كثير. الإكمال شي ثاني تماماً — شغلنا احنا مو تأجيل الزبون، وخلطهم
// يخلي إحصاءات التأجيل كذب.
//
// الحالة ترجع CONFIRMED حتى الحجز يدخل طابور اليوم الجديد عادي،
// والتكليفات تبقى مثل ما هي (الكادر المقترح).
func (r *BookingProgressRepository) ScheduleContinuation(bookingID, scheduledAt, byEmployeeID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	if err := tx.Get(&status, `SELECT status::text FROM "Booking" WHERE id = $1`, bookingID); err != nil {
		return errors.New("الحجز مو موجود")
	}
	if status != "PARTIAL" {
		return errors.New("هذا الحجز مو بحالة إنجاز جزئي")
	}

	if _, err := tx.Exec(`
		UPDATE "Booking" SET
			status = 'CONFIRMED',
			"scheduledAt" = $2::timestamp,
			"scheduledEndAt" = $2::timestamp + interval '1 hour',
			"awaitingReschedule" = false,
			"updatedAt" = now()
		WHERE id = $1
	`, bookingID, scheduledAt); err != nil {
		return err
	}
	// نسجّلها بسجل تغييرات الموعد حتى يبقى واضح ليش انتغير
	if _, err := tx.Exec(`
		INSERT INTO "ScheduleChangeLog" (id, "bookingId", "changedById", "newTime", reason, kind)
		VALUES ($1, $2, $3, $4::timestamp, $5, $6)
	`, uuid.NewString(), bookingID, byEmployeeID, scheduledAt, "جدولة إكمال حجز منجز جزئياً", "CONTINUATION"); err != nil {
		return err
	}
	return tx.Commit()
}

// ListPartial الحجوزات الواقفة بإنجاز جزئي — شاشة الإداري.
func (r *BookingProgressRepository) ListPartial() ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `
		SELECT id FROM "Booking"
		WHERE status = 'PARTIAL' AND "archivedAt" IS NULL
		ORDER BY "lastPartialAt" DESC NULLS LAST
	`)
	return ids, err
}
