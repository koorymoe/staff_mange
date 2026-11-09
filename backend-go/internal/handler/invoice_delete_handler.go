package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// InvoiceDeleteHandler «طلب حذف فاتورة» (دازاي 10-25): المحاسب يطلب ويكتب السبب،
// والمالك وحده يوافق. قبل الحذف تنحفظ نسخة الفاتورة ومواردها كاملة بالطلب نفسه.
type InvoiceDeleteHandler struct {
	db     *sqlx.DB
	notify *repository.NotificationRepository
}

func NewInvoiceDeleteHandler(db *sqlx.DB, notify *repository.NotificationRepository) *InvoiceDeleteHandler {
	return &InvoiceDeleteHandler{db: db, notify: notify}
}

type invoiceDeleteRequest struct {
	ID              string     `db:"id" json:"id"`
	InvoiceID       string     `db:"invoiceId" json:"invoiceId"`
	AccountingCode  *string    `db:"accountingCode" json:"accountingCode"`
	BookingCode     *string    `db:"bookingCode" json:"bookingCode"`
	NetTotal        *float64   `db:"netTotal" json:"netTotal"`
	RequestedByName *string    `db:"requestedByName" json:"requestedByName"`
	Reason          string     `db:"reason" json:"reason"`
	Status          string     `db:"status" json:"status"`
	DecidedByName   *string    `db:"decidedByName" json:"decidedByName"`
	DecidedAt       *time.Time `db:"decidedAt" json:"decidedAt"`
	DecisionNote    *string    `db:"decisionNote" json:"decisionNote"`
	CreatedAt       time.Time  `db:"createdAt" json:"createdAt"`
}

const invoiceDeleteSelect = `SELECT d.id, d."invoiceId", d."accountingCode", d."bookingCode", d."netTotal", rq.name AS "requestedByName",
	d.reason, d.status, dc.name AS "decidedByName", d."decidedAt", d."decisionNote", d."createdAt"
	FROM "LeaderInvoiceDeleteRequest" d
	LEFT JOIN "Employee" rq ON rq.id = d."requestedById"
	LEFT JOIN "Employee" dc ON dc.id = d."decidedById"`

// POST /api/leader-invoices/{id}/delete-request — {reason}
func (h *InvoiceDeleteHandler) Request(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Reason string `json:"reason"`
	}
	_ = DecodeJSON(r, &b)
	reason := strings.TrimSpace(b.Reason)
	if len([]rune(reason)) < 5 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الحذف (٥ حروف على الأقل)")
		return
	}
	var inv struct {
		Status string   `db:"status"`
		Code   *string  `db:"accountingCode"`
		Net    *float64 `db:"netTotal"`
		BCode  *string  `db:"bookingCode"`
	}
	err := h.db.Get(&inv, `SELECT li.status, li."accountingCode", li."netTotal", b.code AS "bookingCode"
		FROM "LeaderInvoice" li LEFT JOIN "Booking" b ON b.id = li."bookingId" WHERE li.id = $1`, r.PathValue("id"))
	if errors.Is(err, sql.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "الفاتورة مو موجودة")
		return
	} else if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	if inv.Status == "APPROVED" {
		WriteError(w, http.StatusBadRequest, "الفاتورة معتمدة — اسحب اعتمادها أول، وبعدين اطلب حذفها")
		return
	}
	var id string
	err = h.db.Get(&id, `INSERT INTO "LeaderInvoiceDeleteRequest" ("invoiceId", "accountingCode", "bookingCode", "netTotal", "requestedById", reason)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		r.PathValue("id"), inv.Code, inv.BCode, inv.Net, middleware.EmployeeIDFromContext(r), reason)
	if err != nil {
		if strings.Contains(err.Error(), "LeaderInvoiceDeleteRequest_pending") {
			WriteError(w, http.StatusConflict, "اكو طلب حذف لهاي الفاتورة بانتظار المالك")
			return
		}
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الطلب")
		return
	}
	if h.notify != nil {
		code := ""
		if inv.Code != nil {
			code = *inv.Code
		}
		_ = h.notify.CreateForRole("OWNER", "INVOICE_DELETE_REQUEST", "🗑️ طلب حذف فاتورة "+code+" — السبب: "+reason+". افتح «فواتير الليدر» ← طلبات الحذف.")
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// GET /api/leader-invoices/delete-requests?status=PENDING — المالك والمحاسب
func (h *InvoiceDeleteHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []invoiceDeleteRequest{}
	q := invoiceDeleteSelect
	args := []any{}
	if st := r.URL.Query().Get("status"); st != "" {
		q += ` WHERE d.status = $1`
		args = append(args, st)
	}
	if err := h.db.Select(&rows, q+` ORDER BY d."createdAt" DESC LIMIT 200`, args...); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الطلبات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/leader-invoices/delete-requests/{id}/decide — {approve, note} — المالك وحده
func (h *InvoiceDeleteHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	_ = DecodeJSON(r, &b)
	me := middleware.EmployeeIDFromContext(r)
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	defer tx.Rollback()
	var req struct {
		InvoiceID string `db:"invoiceId"`
		Status    string `db:"status"`
	}
	if err := tx.Get(&req, `SELECT "invoiceId", status FROM "LeaderInvoiceDeleteRequest" WHERE id = $1 FOR UPDATE`, r.PathValue("id")); err != nil {
		WriteError(w, http.StatusNotFound, "الطلب مو موجود")
		return
	}
	if req.Status != "PENDING" {
		WriteError(w, http.StatusConflict, "الطلب انحسم قبل")
		return
	}
	if !b.Approve {
		_, err = tx.Exec(`UPDATE "LeaderInvoiceDeleteRequest" SET status = 'REJECTED', "decidedById" = $2, "decidedAt" = now(), "decisionNote" = NULLIF($3, '') WHERE id = $1`,
			r.PathValue("id"), me, strings.TrimSpace(b.Note))
	} else {
		var status string
		if err := tx.Get(&status, `SELECT status FROM "LeaderInvoice" WHERE id = $1 FOR UPDATE`, req.InvoiceID); err != nil {
			WriteError(w, http.StatusNotFound, "الفاتورة ما عادت موجودة")
			return
		}
		if status == "APPROVED" {
			WriteError(w, http.StatusBadRequest, "الفاتورة انعتمدت بعد الطلب — اسحب اعتمادها أول")
			return
		}
		// نسخة كاملة قبل الحذف: الفاتورة + موادها + تعديلاتها
		_, err = tx.Exec(`UPDATE "LeaderInvoiceDeleteRequest" SET status = 'APPROVED', "decidedById" = $2, "decidedAt" = now(),
			"decisionNote" = NULLIF($3, ''),
			snapshot = jsonb_build_object(
				'invoice', (SELECT to_jsonb(li) FROM "LeaderInvoice" li WHERE li.id = $4),
				'materials', COALESCE((SELECT jsonb_agg(to_jsonb(m)) FROM "LeaderInvoiceMaterialItem" m WHERE m."leaderInvoiceId" = $4), '[]'::jsonb),
				'adjustments', COALESCE((SELECT jsonb_agg(to_jsonb(a)) FROM "LeaderInvoiceAdjustment" a WHERE a."invoiceId" = $4), '[]'::jsonb))
			WHERE id = $1`, r.PathValue("id"), me, strings.TrimSpace(b.Note), req.InvoiceID)
		if err == nil {
			_, err = tx.Exec(`DELETE FROM "LeaderInvoice" WHERE id = $1`, req.InvoiceID)
		}
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر: "+err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
