package repository

import (
	"fmt"
	"strings"

	"staffmange-api/internal/errlog"

	"staffmange-api/internal/model"
)

// بند ٧م: قائمة فواتير الليدرية چانت تنزّل **كل** الفواتير من أول يوم
// والمتصفح يرشّح ويعدّ ويجمع. هنا نفس المنطق حرفياً بالخادم: صفحة وحدة
// تنزل، والعدّادات والمجموع ينحسبون على الكل — فالأرقام تضل صحيحة.
//
// ⚠️ كل شرط هنا نسخة من دوال LeaderInvoicesListPage.tsx (inTab وغيرها).
// أي تغيير بواحد لازم ينعكس بالثاني.

type LeaderInvoicePageQuery struct {
	EmployeeID string
	Tab        string
	Search     string
	Leader     string
	Customer   string
	System     string
	Month      string
	Page       int
	PerPage    int
}

type LeaderInvoicePage struct {
	Rows      []model.LeaderInvoice `json:"rows"`
	Total     int                   `json:"total"`
	Sum       float64               `json:"sum"`
	Counts    map[string]int        `json:"counts"`
	Leaders   []string              `json:"leaders"`
	Customers []string              `json:"customers"`
	Systems   []string              `json:"systems"`
}

const (
	liFrom       = ` FROM "LeaderInvoice" LEFT JOIN "Booking" b ON b.id = "LeaderInvoice"."bookingId" LEFT JOIN "Employee" e ON e.id = "LeaderInvoice"."employeeId" `
	liInternal   = `COALESCE(b."bookingType"::text, '') = 'INTERNAL'`
	liAudited    = `("LeaderInvoice"."auditVerdict" IS NOT NULL AND btrim("LeaderInvoice"."auditVerdict") <> '')`
	liAtMonitor  = `("LeaderInvoice"."monitorRequestedAt" IS NOT NULL AND "LeaderInvoice"."monitorDecidedAt" IS NULL)`
	liReviewed   = `"LeaderInvoice"."monitorDecidedAt" IS NOT NULL`
	liApproved   = `"LeaderInvoice".status = 'APPROVED'`
	liInvInt     = `(COALESCE("LeaderInvoice".systems, '[]'::jsonb) ? 'شغل داخل الشركة')`
	liNeedNumber = `(` + liApproved + ` AND btrim(COALESCE("LeaderInvoice"."externalInvoiceNumber", '')) = ''
		AND NOT "LeaderInvoice"."isFree" AND COALESCE("LeaderInvoice"."auditVerdict", '') <> 'FREE'
		AND NOT ` + liInternal + ` AND NOT ` + liInvInt + `)`
)

func liTabSQL(tab string) string {
	if tab == "INTERNAL" {
		return liInternal
	}
	ext := `NOT ` + liInternal
	switch tab {
	case "ALL":
		return ext
	case "AUDIT":
		return ext + ` AND NOT ` + liApproved + ` AND NOT ` + liAudited
	case "MONITOR":
		return ext + ` AND NOT ` + liApproved + ` AND ` + liAtMonitor
	case "REVIEWED":
		return ext + ` AND ` + liReviewed
	case "PENDING":
		return ext + ` AND NOT ` + liApproved + ` AND ` + liAudited + ` AND NOT ` + liAtMonitor
	case "APPROVED":
		return ext + ` AND ` + liApproved
	default: // NO_NUMBER
		return ext + ` AND ` + liNeedNumber
	}
}

func (r *LeaderInvoiceRepository) Page(q LeaderInvoicePageQuery) (*LeaderInvoicePage, error) {
	args := []any{}
	arg := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	base := `TRUE`
	if q.EmployeeID != "" {
		base += ` AND "LeaderInvoice"."employeeId" = ` + arg(q.EmployeeID)
	}
	// البحث والمرشّحات يقيّدون الصفوف والمجموع بس. العدّادات (تحت) على الكل، مثل الواجهة.
	filt := ``
	if s := strings.TrimSpace(q.Search); s != "" {
		p := arg(s)
		fields := []string{`"LeaderInvoice"."externalInvoiceNumber"`, `"LeaderInvoice"."accountingCode"`, `"LeaderInvoice"."customerName"`, `e.name`, `b.code`}
		ors := []string{}
		for _, f := range fields {
			ors = append(ors, `ar_norm(`+f+`) LIKE '%' || ar_norm(`+p+`) || '%'`)
			ors = append(ors, `(ar_norm(`+p+`) ~ '^[0-9 ]+$' AND regexp_replace(ar_norm(`+f+`), '[^0-9]', '', 'g') LIKE '%' || regexp_replace(ar_norm(`+p+`), '[^0-9]', '', 'g') || '%')`)
		}
		filt += ` AND (` + strings.Join(ors, " OR ") + `)`
	}
	if q.Leader != "" {
		filt += ` AND e.name = ` + arg(q.Leader)
	}
	if q.Customer != "" {
		filt += ` AND "LeaderInvoice"."customerName" = ` + arg(q.Customer)
	}
	if q.System != "" {
		filt += ` AND COALESCE("LeaderInvoice".systems, '[]'::jsonb) ? ` + arg(q.System)
	}
	if q.Month != "" {
		filt += ` AND to_char("LeaderInvoice"."createdAt", 'YYYY-MM') = ` + arg(q.Month)
	}
	where := base + filt + ` AND ` + liTabSQL(q.Tab)

	out := &LeaderInvoicePage{Counts: map[string]int{}, Rows: []model.LeaderInvoice{}}
	var agg struct {
		N   int     `db:"n"`
		Sum float64 `db:"s"`
	}
	if err := r.db.Get(&agg, `SELECT count(*) AS n, COALESCE(SUM("LeaderInvoice"."netTotal"), 0) AS s`+liFrom+`WHERE `+where, args...); err != nil {
		return nil, err
	}
	out.Total, out.Sum = agg.N, agg.Sum

	per := q.PerPage
	if per <= 0 || per > 200 {
		per = 50
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	if err := r.db.Select(&out.Rows, `SELECT "LeaderInvoice".*`+liFrom+`WHERE `+where+
		fmt.Sprintf(` ORDER BY "LeaderInvoice"."createdAt" DESC LIMIT %d OFFSET %d`, per, (page-1)*per), args...); err != nil {
		return nil, err
	}
	r.attachFreeReasonLabels(out.Rows)
	if err := r.hydrateAll(out.Rows); err != nil {
		return nil, err
	}

	// العدّادات والخيارات: على الكل (بس مقيّدة بالموظف) — مثل الواجهة بالضبط.
	baseArgs := []any{}
	cb := `TRUE`
	if q.EmployeeID != "" {
		baseArgs = append(baseArgs, q.EmployeeID)
		cb += ` AND "LeaderInvoice"."employeeId" = $1`
	}
	cols := []string{}
	tabs := []string{"INTERNAL", "AUDIT", "MONITOR", "REVIEWED", "PENDING", "NO_NUMBER", "APPROVED", "ALL"}
	for _, t := range tabs {
		cols = append(cols, `count(*) FILTER (WHERE `+liTabSQL(t)+`)`)
	}
	row := r.db.QueryRowx(`SELECT `+strings.Join(cols, ", ")+liFrom+`WHERE `+cb, baseArgs...)
	vals := make([]int, len(tabs))
	ptrs := make([]any, len(tabs))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := row.Scan(ptrs...); err != nil {
		return nil, err
	}
	for i, t := range tabs {
		out.Counts[t] = vals[i]
	}
	out.Leaders, out.Customers, out.Systems = []string{}, []string{}, []string{}
	errlog.Err(r.db.Select(&out.Leaders, `SELECT DISTINCT e.name`+liFrom+`WHERE `+cb+` AND COALESCE(e.name, '') <> '' ORDER BY 1`, baseArgs...))
	errlog.Err(r.db.Select(&out.Customers, `SELECT DISTINCT "LeaderInvoice"."customerName"`+liFrom+`WHERE `+cb+` AND COALESCE("LeaderInvoice"."customerName", '') <> '' ORDER BY 1`, baseArgs...))
	errlog.Err(r.db.Select(&out.Systems, `SELECT DISTINCT s`+liFrom+`, jsonb_array_elements_text(CASE WHEN jsonb_typeof("LeaderInvoice".systems) = 'array' THEN "LeaderInvoice".systems ELSE '[]'::jsonb END) s WHERE `+cb+` ORDER BY 1`, baseArgs...))
	return out, nil
}
