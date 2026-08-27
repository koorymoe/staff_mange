package handler

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
)

// ═══ وين هذا الحجز؟ ═══
//
// «من نريد نبحث بأي خانة عن رقم الحجز، إذا جان الموظف يبحث بالمكان
// الخطأ يطلعله من النظام: هذا الحجز مو بهذا المكان، بالمكان الفلاني».
//
// ⚠️⚠️ **المشكلة الي يحلّها**: الموظف يدوّر على حجز بشاشة «الحجوزات»
// وما يلگاه، فيستنتج **إنه مو موجود** — والحقيقة إنه بالأرشيف أو
// انلغى. فيسأل زميله، أو يسوي حجزاً ثانياً بنفس الشغل. والنظام يعرف
// وين هو بالضبط وساكت.
//
// ⚠️ ويرجّع **مكاناً واحداً** مو قائمة احتمالات: «موجود بالأرشيف»
// جواب، و«ممكن بالأرشيف أو بالملغاة أو…» مو جواب — يرجّع الموظف
// يدوّر بيده، وهذا الي نلغيه.
//
// ⚠️ وما ينطي بيانات الحجز: يقول **وين** بس. الوصول للتفاصيل يبقى
// محكوماً بصلاحية الشاشة نفسها — البحث ما يصير باباً خلفياً.

type LocateHandler struct{ db *sqlx.DB }

func NewLocateHandler(db *sqlx.DB) *LocateHandler { return &LocateHandler{db: db} }

type locateResult struct {
	Found bool   `json:"found"`
	Code  string `json:"code,omitempty"`
	// Where اسم المكان بالعربي مثل ما يشوفه الموظف بالقائمة
	Where string `json:"where,omitempty"`
	Route string `json:"route,omitempty"`
	// Hint سطر يشرح الحالة
	Hint string `json:"hint,omitempty"`
	// InvoiceStage مرحلة فاتورة الحجز إذا عنده فاتورة
	InvoiceStage string `json:"invoiceStage,omitempty"`
	InvoiceRoute string `json:"invoiceRoute,omitempty"`
}

// GET /api/locate?q=CODE
func (h *LocateHandler) Locate(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		WriteJSON(w, http.StatusOK, locateResult{})
		return
	}

	var row struct {
		ID         string  `db:"id"`
		Code       string  `db:"code"`
		Status     string  `db:"status"`
		Archived   *string `db:"archivedAt"`
		Cancelled  *string `db:"cancelledAt"`
		InvStatus  *string `db:"invStatus"`
		InvVerdict *string `db:"invVerdict"`
		InvMonitor *string `db:"invMonitor"`
		InvDecided *string `db:"invDecided"`
	}
	// ⚠️ البحث **بكود الحجز** — جدول الزبون ماكو عنده عمود كود
	// (فحصته). إضافة عمود ما موجود تخلّي الاستعلام يفشل كاملاً
	// ويرجّع «ماكو» لكل شي — وهذا بالضبط الي صار بأول محاولة، والفحص
	// هو الي مسكه.
	err := h.db.Get(&row, `
		SELECT b.id, b.code, b.status,
		       b."archivedAt"::text AS "archivedAt",
		       b."cancelledAt"::text AS "cancelledAt",
		       i.status  AS "invStatus",
		       i."auditVerdict" AS "invVerdict",
		       i."monitorRequestedAt"::text AS "invMonitor",
		       i."monitorDecidedAt"::text  AS "invDecided"
		FROM "Booking" b
		LEFT JOIN LATERAL (
			SELECT status, "auditVerdict", "monitorRequestedAt", "monitorDecidedAt"
			FROM "LeaderInvoice" WHERE "bookingId" = b.id
			ORDER BY "createdAt" DESC LIMIT 1
		) i ON TRUE
		WHERE lower(btrim(b.code)) = lower($1)
		ORDER BY b."createdAt" DESC
		LIMIT 1`, q)
	if err != nil {
		WriteJSON(w, http.StatusOK, locateResult{})
		return
	}

	res := locateResult{Found: true, Code: row.Code}
	switch {
	case row.Archived != nil:
		res.Where, res.Route = "أرشيف الحجوزات", "/bookings-archive"
		res.Hint = "الحجز انأرشف — تلگاه بالأرشيف مو بالحجوزات."
	case row.Cancelled != nil || row.Status == "CANCELLED":
		res.Where, res.Route = "الحجوزات الملغاة", "/bookings"
		res.Hint = "الحجز انلغى — يظهر بالحجوزات بحالة «ملغى»."
	case row.Status == "COMPLETED":
		res.Where, res.Route = "الحجوزات المنجزة", "/bookings"
		res.Hint = "الحجز انخلص — رشّح الحجوزات على «منجز»."
	default:
		res.Where, res.Route = "الحجوزات", "/bookings"
	}

	// ⚠️ وإذا عنده فاتورة نقول **بأي طابور** هي: أكثر بحث عن حجز سببه
	// فاتورته، والجواب «موجود بالحجوزات» ما يكفي.
	if row.InvStatus != nil {
		res.InvoiceRoute = "/leader-invoices"
		switch {
		case *row.InvStatus == "APPROVED":
			res.InvoiceStage = "فاتورته معتمدة"
		case row.InvMonitor != nil && row.InvDecided == nil:
			res.InvoiceStage = "فاتورته عند المراقب"
		case row.InvVerdict != nil && strings.TrimSpace(*row.InvVerdict) != "":
			res.InvoiceStage = "فاتورته بانتظار الاعتماد"
		default:
			res.InvoiceStage = "فاتورته بانتظار التدقيق"
		}
	}
	WriteJSON(w, http.StatusOK, res)
}
