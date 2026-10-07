package handler

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// ═══ تسعير الأعمال الداخلية القديمة (مؤقت — قرار (ع) 10-07) ═══
// ٩٧ حجز داخل الشركة قديمة بلا فاتورة: المحاسب أو المراقب يكتب شنو انعمل
// والسعر، فتنسوّى فاتورة داخلية. يخص الحجوزات المسجّلة **قبل** internalBackfillCutoff
// بس — والجديد مسؤولية الإداري. ينشال بعد ما تخلص القائمة.
const internalBackfillCutoff = "2026-10-08"

// internalWorkPaperSQL حجز داخلي مخلّص ناقصه ورق — نفس شروط «ناقصه ورق».
const internalNeedsInvoiceSQL = `b."bookingType" = 'INTERNAL' AND b.status = 'COMPLETED' AND b."archivedAt" IS NULL
	AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id AND li."revokedAt" IS NULL)`

type InternalBackfillHandler struct {
	db       *sqlx.DB
	invoices *service.LeaderInvoiceService
}

func NewInternalBackfillHandler(db *sqlx.DB, invoices *service.LeaderInvoiceService) *InternalBackfillHandler {
	return &InternalBackfillHandler{db: db, invoices: invoices}
}

type internalBackfillRow struct {
	ID         string  `db:"id" json:"id"`
	Code       string  `db:"code" json:"code"`
	Department *string `db:"department" json:"department"`
	Notes      *string `db:"notes" json:"notes"`
}

// GET /api/internal-backfill — الحجوزات القديمة الي تنتظر سعر.
func (h *InternalBackfillHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []internalBackfillRow{}
	if err := h.db.Select(&rows, `SELECT b.id, b.code, b."internalDepartment" AS department, b.notes
		FROM "Booking" b WHERE `+internalNeedsInvoiceSQL+` AND b."createdAt" < $1::date
		ORDER BY b."createdAt"`, internalBackfillCutoff); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب القائمة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/internal-backfill/{id} {work, price} — يسوي الفاتورة الداخلية.
func (h *InternalBackfillHandler) Price(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Work  string  `json:"work"`
		Price float64 `json:"price"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	id := r.PathValue("id")
	var ok bool
	_ = h.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Booking" b WHERE b.id = $1 AND `+internalNeedsInvoiceSQL+`
		AND b."createdAt" < $2::date)`, id, internalBackfillCutoff)
	if !ok {
		WriteError(w, http.StatusBadRequest, "هالزر للأعمال الداخلية القديمة الي بلا فاتورة بس — الجديدة يسوي فاتورتها الإداري")
		return
	}
	inv, err := h.invoices.CreateInternalInvoice(middleware.EmployeeIDFromContext(r), model.CreateInternalInvoiceRequest{
		BookingID: id, Work: strings.TrimSpace(body.Work), Price: body.Price,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// GET /api/bookings/internal-paperwork — طابور الإداري: أعمال داخلية خلصت وناقصها
// فاتورة أو تقرير (قرار (ع) 10-07: الورق مسؤولية إداري الحجوزات).
func (h *InternalBackfillHandler) CoordinatorQueue(w http.ResponseWriter, r *http.Request) {
	type row struct {
		ID         string  `db:"id" json:"id"`
		Code       string  `db:"code" json:"code"`
		Department *string `db:"department" json:"department"`
		HasInvoice bool    `db:"hasInvoice" json:"hasInvoice"`
		HasReport  bool    `db:"hasReport" json:"hasReport"`
		DoneAt     *string `db:"doneAt" json:"doneAt"`
	}
	rows := []row{}
	if err := h.db.Select(&rows, `SELECT * FROM (
		SELECT b.id, b.code, b."internalDepartment" AS department,
		       EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id AND li."revokedAt" IS NULL) AS "hasInvoice",
		       EXISTS (SELECT 1 FROM "WorkReport" wr WHERE wr."bookingId" = b.id) AS "hasReport",
		       to_char(b."completedAt" + interval '3 hours', 'YYYY-MM-DD HH24:MI') AS "doneAt"
		FROM "Booking" b
		WHERE b."bookingType" = 'INTERNAL' AND b.status = 'COMPLETED' AND b."archivedAt" IS NULL
		  AND b."createdAt" >= $1::date) x
		WHERE NOT x."hasInvoice" OR NOT x."hasReport"
		ORDER BY x."doneAt" DESC LIMIT 100`, internalBackfillCutoff); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الطابور")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
