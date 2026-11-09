package repository

import (
	"fmt"
	"strings"
	"time"

	"staffmange-api/internal/errlog"
	"staffmange-api/internal/model"
)

// بند ٧م: شاشة الزبائن چانت تنزّل ٣٠٠٠ زبون مرة وحدة والمتصفح يرشّح ويعدّ.
// بعد ما يعبر العدد الحد، الباقي يختفي بصمت. هنا نفس الشروط حرفياً بالخادم.
//
// ⚠️ كل شرط نسخة من Customers.tsx (filteredCustomers وأرقام الرأس) ومن
// matches() بـ utils/search.ts. أي تغيير بواحد لازم ينعكس بالثاني.

type CustomerPageQuery struct {
	Search   string
	Location string
	Kind     string // vip | normal
	Claim    string // false | clean
	GPS      bool
	Page     int
	PerPage  int
}

type CustomerPageRow struct {
	model.Customer
	GpsNumber       *string    `db:"gpsNumber"`
	DeviceID        *string    `db:"deviceId"`
	SubscriptionEnd *time.Time `db:"subscriptionEnd"`
}

type CustomerPageResult struct {
	Rows         []CustomerPageRow
	Total        int
	All          int
	NewThisMonth int
	VIP          int
	GPSCount     int
	Locations    []string
}

// custCodeSQL = FormatCode بالـ Go: CUST-%05d
const custCodeSQL = `('CUST-' || lpad(c."customerCode"::text, greatest(5, length(c."customerCode"::text)), '0'))`

func custSearchSQL(p string) string {
	ors := []string{}
	for _, f := range []string{custCodeSQL, `c.name`, `c.phone`} {
		ors = append(ors, `ar_norm(`+f+`) LIKE '%' || ar_norm(`+p+`) || '%'`)
		ors = append(ors, `(regexp_replace(ar_norm(`+p+`), '[^0-9]', '', 'g') <> '' AND regexp_replace(ar_norm(`+p+`), '[^0-9]', '', 'g') = replace(ar_norm(`+p+`), ' ', '')
			AND regexp_replace(ar_norm(`+f+`), '[^0-9]', '', 'g') LIKE '%' || regexp_replace(ar_norm(`+p+`), '[^0-9]', '', 'g') || '%')`)
	}
	return `(` + strings.Join(ors, " OR ") + `)`
}

func (r *CustomerRepository) Page(q CustomerPageQuery) (*CustomerPageResult, error) {
	args := []any{}
	arg := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	from := ` FROM "Customer" c LEFT JOIN "CustomerGpsInfo" gi ON gi."customerId" = c.id `
	if q.GPS {
		from = ` FROM "Customer" c JOIN "CustomerServiceTag" t ON t."customerId" = c.id AND t.service = 'GPS' LEFT JOIN "CustomerGpsInfo" gi ON gi."customerId" = c.id `
	}
	search := strings.TrimSpace(q.Search)
	base := `TRUE`
	sp := ""
	if search != "" {
		sp = arg(search)
		base += ` AND ` + custSearchSQL(sp)
	}
	where := base
	if q.Location != "" {
		where += ` AND btrim(COALESCE(c.location, '')) = ` + arg(q.Location)
	}
	switch q.Kind {
	case "vip":
		where += ` AND COALESCE(c.position, '') <> ''`
	case "normal":
		where += ` AND COALESCE(c.position, '') = ''`
	}
	switch q.Claim {
	case "false":
		where += ` AND c."falseClaimCount" > 0`
	case "clean":
		where += ` AND c."falseClaimCount" = 0`
	}
	out := &CustomerPageResult{Rows: []CustomerPageRow{}, Locations: []string{}}
	if err := r.db.Get(&out.Total, `SELECT count(*)`+from+`WHERE `+where, args...); err != nil {
		return nil, err
	}
	per := q.PerPage
	if per <= 0 || per > 200 {
		per = 10
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	order := ` ORDER BY c."customerCode" ASC`
	if search != "" && !q.GPS {
		order = ` ORDER BY (ar_norm(c.name) = ar_norm(` + sp + `)) DESC, (ar_norm(c.name) LIKE ar_norm(` + sp + `) || '%') DESC,
			position(ar_norm(` + sp + `) in ar_norm(c.name)), c."customerCode" ASC`
	}
	if err := r.db.Select(&out.Rows, `SELECT c.*, gi."gpsNumber", gi."deviceId", gi."subscriptionEnd"`+from+`WHERE `+where+order+
		fmt.Sprintf(` LIMIT %d OFFSET %d`, per, (page-1)*per), args...); err != nil {
		return nil, err
	}

	// أرقام الرأس: على كل الزبائن (أو نتائج البحث وقت البحث) — مثل الواجهة.
	// ⚠️ ما تتأثر بالتبويب ولا بالمرشّحات.
	sargs := []any{}
	sw := `TRUE`
	if len([]rune(search)) >= 2 {
		sargs = append(sargs, search)
		sw = custSearchSQL("$1")
	}
	var st struct {
		All int `db:"a"`
		New int `db:"n"`
		VIP int `db:"v"`
	}
	if err := r.db.Get(&st, `SELECT count(*) AS a,
		count(*) FILTER (WHERE c."createdAt" >= ((date_trunc('month', now() AT TIME ZONE 'Asia/Baghdad')) AT TIME ZONE 'Asia/Baghdad') AT TIME ZONE 'UTC') AS n,
		count(*) FILTER (WHERE COALESCE(c.position, '') <> '') AS v
		FROM "Customer" c WHERE `+sw, sargs...); err != nil {
		return nil, err
	}
	out.All, out.NewThisMonth, out.VIP = st.All, st.New, st.VIP
	// عدد تبويب الجي بي اس (كل زبائن الخدمة، بلا بحث ولا مرشّحات — مثل الواجهة القديمة)
	errlog.Err(r.db.Get(&out.GPSCount, `SELECT count(DISTINCT "customerId") FROM "CustomerServiceTag" WHERE service = 'GPS'`))
	lfrom := ` FROM "Customer" c `
	if q.GPS {
		lfrom = ` FROM "Customer" c JOIN "CustomerServiceTag" t ON t."customerId" = c.id AND t.service = 'GPS' `
	}
	errlog.Err(r.db.Select(&out.Locations, `SELECT DISTINCT btrim(c.location) COLLATE "C" AS l`+lfrom+
		`WHERE btrim(COALESCE(c.location, '')) <> '' ORDER BY 1 LIMIT 60`))
	return out, nil
}
