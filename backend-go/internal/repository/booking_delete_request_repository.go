package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type BookingDeleteRequestRepository struct {
	db *sqlx.DB
}

func NewBookingDeleteRequestRepository(db *sqlx.DB) *BookingDeleteRequestRepository {
	return &BookingDeleteRequestRepository{db: db}
}

const bookingDeleteSelect = `SELECT r.*, b.code AS "bookingCode", b.status::text AS "bookingStatus",
		COALESCE(c.name, '') AS "customerName",
		e.name AS "requestedByName", d.name AS "decidedByName", n.name AS "needsInfoByName"
	FROM "BookingDeleteRequest" r
	JOIN "Booking" b ON b.id = r."bookingId"
	LEFT JOIN "Customer" c ON c.id = b."customerId"
	JOIN "Employee" e ON e.id = r."requestedById"
	LEFT JOIN "Employee" d ON d.id = r."decidedById"
	LEFT JOIN "Employee" n ON n.id = r."needsInfoById"`

func decorateDeleteRequests(rows []model.BookingDeleteRequest) []model.BookingDeleteRequest {
	for i := range rows {
		rows[i].StatusLabel = model.BookingDeleteStatusLabels[rows[i].Status]
		if rows[i].Channel != nil {
			rows[i].ChannelLabel = model.BookingDeleteChannelLabels[*rows[i].Channel]
		}
		if rows[i].RequestType != nil {
			rows[i].RequestTypeLabel = model.BookingDeleteTypeLabels[*rows[i].RequestType]
		}
	}
	return rows
}

func (r *BookingDeleteRequestRepository) Create(bookingID, requestedByID, reason, channel, requestType string) (*model.BookingDeleteRequest, error) {
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "BookingDeleteRequest" (id, "bookingId", "requestedById", reason, channel, "requestType")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING id`, bookingID, requestedByID, reason, channel, requestType)
	if err != nil {
		// الفهرس الفريد يمنع طلبين معلقين لنفس الحجز
		return nil, fmt.Errorf("أكو طلب حذف معلّق على هذا الحجز أصلاً")
	}
	return r.Get(id)
}

// SetNeedsInfo تعلّم الطلب «معلقة» — ناقصه معلومات ولسه ما انبتّ فيه.
// شرط PENDING داخل التحديث يمنع تعليم طلب انبتّ فيه أصلاً.
func (r *BookingDeleteRequestRepository) SetNeedsInfo(id, note, byID string) (*model.BookingDeleteRequest, error) {
	var bookingID string
	if err := r.db.Get(&bookingID, `
		UPDATE "BookingDeleteRequest"
		SET "needsInfo" = true, "needsInfoNote" = $2, "needsInfoAt" = now(), "needsInfoById" = $3
		WHERE id = $1 AND status = 'PENDING'
		RETURNING "bookingId"`, id, note, byID); err != nil {
		return nil, fmt.Errorf("الطلب مو موجود أو انبتّ بيه من قبل")
	}
	return r.Get(id)
}

// Counts عدّ مجمَّع واحد للبطاقات الخمس — يبقى مطابقاً للقائمة الكاملة
// حتى لو الواجهة تعرض صفحة واحدة منها.
func (r *BookingDeleteRequestRepository) Counts() (model.BookingDeleteRequestCounts, error) {
	var c model.BookingDeleteRequestCounts
	err := r.db.Get(&c, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'APPROVED')                       AS approved,
			COUNT(*) FILTER (WHERE status = 'PENDING' AND "needsInfo")        AS "needsInfo",
			COUNT(*) FILTER (WHERE status = 'PENDING' AND NOT "needsInfo")    AS "awaitingReview",
			COUNT(*) FILTER (WHERE status = 'REJECTED')                       AS rejected,
			COUNT(*)                                                          AS total
		FROM "BookingDeleteRequest"`)
	return c, err
}

func (r *BookingDeleteRequestRepository) Get(id string) (*model.BookingDeleteRequest, error) {
	var row model.BookingDeleteRequest
	if err := r.db.Get(&row, bookingDeleteSelect+` WHERE r.id = $1`, id); err != nil {
		return nil, err
	}
	out := decorateDeleteRequests([]model.BookingDeleteRequest{row})
	return &out[0], nil
}

func (r *BookingDeleteRequestRepository) List(status string) ([]model.BookingDeleteRequest, error) {
	rows := []model.BookingDeleteRequest{}
	q := bookingDeleteSelect
	args := []any{}
	if status != "" {
		args = append(args, status)
		q += ` WHERE r.status = $1`
	}
	q += ` ORDER BY (r.status = 'PENDING') DESC, r."createdAt" DESC LIMIT 500`
	err := r.db.Select(&rows, q, args...)
	return decorateDeleteRequests(rows), err
}

// Decide الموافقة تحذف الحجز فعلياً، والرفض يخلي الحجز مكانه.
//
// الاثنين بمعاملة وحدة: لو الحذف فشل ما نأشر الطلب "انحذف" ونخلي
// المستخدم يظن إن الحجز راح وهو موجود.
func (r *BookingDeleteRequestRepository) Decide(id string, approve bool, note *string, byID string) (*model.BookingDeleteRequest, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	status := model.BookingDeleteStatusRejected
	if approve {
		status = model.BookingDeleteStatusApproved
	}

	var bookingID string
	// شرط PENDING داخل التحديث يمنع البت مرتين
	if err := tx.Get(&bookingID, `
		UPDATE "BookingDeleteRequest"
		SET status = $2, "decidedById" = $3, "decidedAt" = now(), "decisionNote" = NULLIF($4,'')
		WHERE id = $1 AND status = 'PENDING'
		RETURNING "bookingId"`, id, status, byID, derefStr(note)); err != nil {
		return nil, fmt.Errorf("الطلب مو موجود أو انبتّ بيه من قبل")
	}

	if approve {
		// أرشفة مو محو. المحو كان يشيل الحجز وكل تاريخه للأبد — وحتى
		// طلب الحذف نفسه (منو طلبه وليش ومنو وافق) ينمحي معاه
		// بالـCASCADE، فما يبقى ولا دليل على القرار.
		//
		// هسه الحجز يختفي من الحجوزات ومن تنسيق الحجوزات، ويضل بالأرشيف
		// بسبب حذفه ومنو حذفه — نقدر نجاوب «شكد حجز انلغى الشهر هذا
		// وليش» بدل ما نخمّن.
		if _, err := tx.Exec(`
			UPDATE "Booking"
			SET "archivedAt" = now(), "archivedById" = $2,
			    "archiveReason" = COALESCE(
			        (SELECT reason FROM "BookingDeleteRequest" WHERE id = $3), 'حذف بموافقة الإدارة')
			WHERE id = $1 AND "archivedAt" IS NULL`, bookingID, byID, id); err != nil {
			return nil, fmt.Errorf("تعذر أرشفة الحجز")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(id)
}

func (r *BookingDeleteRequestRepository) PendingCount() (int, error) {
	var n int
	err := r.db.Get(&n, `SELECT COUNT(*) FROM "BookingDeleteRequest" WHERE status = 'PENDING'`)
	return n, err
}
