package handler

import (
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ رقابة مشرفي المشاريع على عرض السعر — قرار (ع) 10-16 ═══
// «اريد تشدد بالمراقبة على مشرفين المشاريع… شكد تأخروا بعرض السعر: شوكت
// استلم المشروع، شوكت تواصل وي الزبون، شوكت سواله عرض سعر». ثلاث محطات
// لكل مشروع، والمراقب يشوف الفجوة بين كل وحدة والثانية. بلا غرامات.

const (
	quoteWatchContactHours = 24 // ما تواصل خلال يوم من الاستلام = متأخر
	quoteWatchQuoteHours   = 72 // ما سوّى عرض سعر خلال ٣ أيام من الاستلام = متأخر
)

type ProjectQuoteWatchHandler struct {
	db   *sqlx.DB
	deps *projectMonitorDeps
}

func NewProjectQuoteWatchHandler(db *sqlx.DB) *ProjectQuoteWatchHandler {
	return &ProjectQuoteWatchHandler{db: db}
}

// POST /api/projects/{id}/customer-contacted — المشرف يأشّر «تواصلت ويا الزبون»
func (h *ProjectQuoteWatchHandler) MarkContacted(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Note string `json:"note"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	note := strings.TrimSpace(req.Note)
	if utf8.RuneCountInString(note) < 5 {
		WriteError(w, http.StatusBadRequest, "اكتب شنو حجيت ويا الزبون — جملة قصيرة تكفي")
		return
	}
	res, err := h.db.Exec(`UPDATE "Project" SET "customerContactedAt" = now() AT TIME ZONE 'UTC',
		"customerContactedById" = $2, "customerContactNote" = $3, "updatedAt" = now()
		WHERE id = $1 AND "customerContactedAt" IS NULL`, r.PathValue("id"), middleware.EmployeeIDFromContext(r), note)
	if err != nil {
		log.Printf("project contacted: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر التسجيل")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		WriteError(w, http.StatusBadRequest, "التواصل مسجّل من قبل لهذا المشروع")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type quoteWatchRow struct {
	ProjectID      string     `db:"projectId" json:"projectId"`
	Code           string     `db:"code" json:"code"`
	Name           string     `db:"name" json:"name"`
	Stage          string     `db:"stage" json:"stage"`
	SupervisorID   *string    `db:"supervisorId" json:"supervisorId"`
	SupervisorName *string    `db:"supervisorName" json:"supervisorName"`
	ReceivedAt     time.Time  `db:"receivedAt" json:"receivedAt"`
	ContactedAt    *time.Time `db:"contactedAt" json:"contactedAt"`
	ContactNote    *string    `db:"contactNote" json:"contactNote"`
	QuotedAt       *time.Time `db:"quotedAt" json:"quotedAt"`
	ContactHours   *float64   `db:"contactHours" json:"contactHours"`
	QuoteHours     *float64   `db:"quoteHours" json:"quoteHours"`
	WaitingHours   float64    `db:"waitingHours" json:"waitingHours"`
	Late           bool       `db:"late" json:"late"`
	LateWhy        *string    `db:"lateWhy" json:"lateWhy"`
	OffSystem      bool       `db:"offSystem" json:"offSystem"`
	// Watch إجراءات المراقب الفعّالة على المشروع (👁️ المراقب متابع)
	Watch []projectMonitorAction `db:"-" json:"watch"`
}

// GET /api/project-quote-watch — المراقب: كل مشروع ومحطاته الثلاث
func (h *ProjectQuoteWatchHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []quoteWatchRow{}
	// 🔴 (ع 10-16) «حساب صحيح؟ المشرفين هلكد متأخرين؟» — چان يحسب المكتمل
	// والي بالتنفيذ متأخر لأن ما بيه زر تواصل مسجّل. هسه **المرحلة تحكم**:
	// المشروع الي عبر «اتصال بالزبون» معناه تواصل، والي عبر «عرض السعر»
	// معناه انطى سعر — حتى لو برّا النظام. المتأخر بس الي واكف بمرحلته وفات وقته.
	// والمكتمل/المرفوض ما يطلع أصلاً.
	err := h.db.Select(&rows, `
		WITH base AS (
		  SELECT p.id, p.code, p.name, p.stage,
		         COALESCE(p."delegatedToEmployeeId", p."responsibleEmployeeId") AS sup,
		         COALESCE(p."delegatedAt" AT TIME ZONE 'UTC', p."createdAt") AS recv,
		         p."customerContactedAt" AS contacted, p."customerContactNote" AS note,
		         (SELECT min(q."createdAt") FROM "Quotation" q WHERE q."projectId" = p.id) AS quoted,
		         -- عبر مرحلة الاتصال = تواصل · عبر مرحلة عرض السعر = انطى سعر
		         (p.stage NOT LIKE '1.%') AS "pastContact",
		         (p.stage LIKE '4.%' OR p.stage LIKE '5.%' OR p.stage LIKE '6.%' OR p.stage LIKE '7.%' OR p.stage LIKE '%الإعلام%' OR p.stage LIKE '%إعادة التسعير%') AS "pastQuote"
		  FROM "Project" p
		  WHERE p.stage NOT LIKE '%مرفوض%' AND p.stage NOT LIKE '%مكتمل%'
		)
		SELECT b.id AS "projectId", b.code, b.name, b.stage, b.sup AS "supervisorId", e.name AS "supervisorName",
		       b.recv AS "receivedAt", b.contacted AS "contactedAt", b.note AS "contactNote", b.quoted AS "quotedAt",
		       round((EXTRACT(EPOCH FROM (b.contacted - b.recv)) / 3600)::numeric, 1)::float8 AS "contactHours",
		       round((EXTRACT(EPOCH FROM (b.quoted - b.recv)) / 3600)::numeric, 1)::float8 AS "quoteHours",
		       round((EXTRACT(EPOCH FROM ((now() AT TIME ZONE 'UTC') - b.recv)) / 3600)::numeric, 1)::float8 AS "waitingHours",
		       CASE
		         WHEN b.contacted IS NULL AND NOT b."pastContact" AND (now() AT TIME ZONE 'UTC') - b.recv > make_interval(hours => $1) THEN 'ما تواصل ويا الزبون لهسه'
		         WHEN b.quoted IS NULL AND NOT b."pastQuote" AND (now() AT TIME ZONE 'UTC') - b.recv > make_interval(hours => $2) THEN 'ما سوّى عرض سعر لهسه'
		       END AS "lateWhy",
		       -- ماتركس: سوّى عرض سعر (أو عبر المرحلة) بلا ما يسجّل تواصل — ما ماشي بالنظام.
		       -- بس للمشاريع الي استلمها بعد ما انضاف زر التواصل (10-16)، حتى ما نظلم القديم.
		       (b.contacted IS NULL AND (b.quoted IS NOT NULL OR b."pastContact")
		        AND b.recv >= '2026-10-16'::timestamp) AS "offSystem"
		FROM base b LEFT JOIN "Employee" e ON e.id = b.sup
		ORDER BY (b.quoted IS NULL AND NOT b."pastQuote") DESC, b.recv ASC
		LIMIT 400`, quoteWatchContactHours, quoteWatchQuoteHours)
	if err != nil {
		log.Printf("project quote watch: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب متابعة عروض الأسعار")
		return
	}
	watch := h.watchActions()
	for i := range rows {
		rows[i].Late = rows[i].LateWhy != nil
		rows[i].Watch = watch[rows[i].ProjectID]
		if rows[i].Watch == nil {
			rows[i].Watch = []projectMonitorAction{}
		}
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"rows": rows, "contactHours": quoteWatchContactHours, "quoteHours": quoteWatchQuoteHours,
	})
}
