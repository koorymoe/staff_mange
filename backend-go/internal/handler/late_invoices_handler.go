package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ فواتير متأخرة على المحاسب — قرار (ع) 10-16 ═══
// «المحاسب يعتبر متأخر إذا تأخر على التدقيق لمدة ٢٤ ساعة فقط — لأن هوه
// محاسب واحد يجي نهاية كل يوم يسويهن ويروح». الفاتورة الي انرفعت وما
// انتدققت خلال ٢٤ ساعة = متأخرة. المجانية والداخلي مستثنين (ما يحتاجن تدقيق مبلغ).
const lateInvoiceHours = 24

type LateInvoicesHandler struct{ db *sqlx.DB }

func NewLateInvoicesHandler(db *sqlx.DB) *LateInvoicesHandler { return &LateInvoicesHandler{db: db} }

type lateInvoiceRow struct {
	ID           string    `db:"id" json:"id"`
	AccountingNo string    `db:"accountingCode" json:"accountingCode"`
	BookingCode  *string   `db:"bookingCode" json:"bookingCode"`
	Customer     *string   `db:"customerName" json:"customerName"`
	LeaderName   *string   `db:"leaderName" json:"leaderName"`
	NetTotal     float64   `db:"netTotal" json:"netTotal"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
	WaitHours    float64   `db:"waitHours" json:"waitHours"`
}

// GET /api/late-invoices
func (h *LateInvoicesHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []lateInvoiceRow{}
	if err := h.db.Select(&rows, `
		SELECT li.id, li."accountingCode", b.code AS "bookingCode", li."customerName", e.name AS "leaderName",
		       li."netTotal"::float8 AS "netTotal", li."createdAt",
		       round((EXTRACT(EPOCH FROM ((now() AT TIME ZONE 'UTC') - li."createdAt")) / 3600)::numeric, 1)::float8 AS "waitHours"
		FROM "LeaderInvoice" li
		LEFT JOIN "Booking" b ON b.id = li."bookingId"
		LEFT JOIN "Employee" e ON e.id = li."employeeId"
		WHERE li."auditedAt" IS NULL AND li.status::text <> 'APPROVED'
		  AND NOT li."isFree" AND li.systems::text NOT LIKE '%شغل داخل الشركة%'
		  AND li."createdAt" < (now() AT TIME ZONE 'UTC') - make_interval(hours => $1)
		ORDER BY li."createdAt" ASC LIMIT 300`, lateInvoiceHours); err != nil {
		log.Printf("late invoices: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الفواتير المتأخرة")
		return
	}
	// المحاسبين — حتى المراقب يطلب التفسير من صاحب الشغل مباشرة
	acc := []struct {
		ID   string `db:"id" json:"id"`
		Name string `db:"name" json:"name"`
	}{}
	if err := h.db.Select(&acc, `SELECT id, name FROM "Employee" WHERE status = 'ACTIVE' AND role::text = 'FINANCE' ORDER BY name`); err != nil {
		log.Printf("late invoices accountants: %v", err)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"hours": lateInvoiceHours, "rows": rows, "accountants": acc})
}
