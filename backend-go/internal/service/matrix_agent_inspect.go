package service

import (
	"fmt"
	"strings"
	"time"
)

// ═══ ماتركس يفحص حجز أو زبون — طلب (ع) 10-14 ═══
//
// بالمحادثة: يدزله رقم حجز → يكله هل مكرر، متأخر لو عالموعد، ومشى بسلاسة
// لو صارت مشاكل. يدزله زبون → شنو طالب منا، آخر حركة، منو طلعله، شكاوى
// ومشاكل. (الموظف عنده employee_profile أصلاً.)
//
// 🔒 الخصوصية: اسم الزبون ورقمه ما يطلعون بمخرجات الأداة — «الزبون #كود»
// بس. أسماء الموظفين تنخفى مركزياً بـnameMask مثل باقي الأدوات.

func (a *MatrixAgent) registerInspectTools() {
	a.AddTool(agentTool{Name: "booking_check", Desc: "فحص حجز بكوده: هل مكرر (نفس الزبون ونفس العنوان بفارق يوم، أو مأشّر بتدقيق التكرار)، هل تأخر عن موعده أو كان عالموعد، هل مشى بسلاسة (تأجيلات، إنجاز جزئي، توقف، زبون ما يرد، شكاوى، إلغاء). استعمله لمن يدزلك رقم حجز ويسأل عن حاله.",
		Props: map[string]any{"code": map[string]any{"type": "string", "description": "كود الحجز مثل B823"}},
		Req:   []string{"code"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			code := strings.TrimPrefix(strings.TrimSpace(strArg(in, "code")), "#")
			if code == "" || a.db == nil {
				return "اكتب كود الحجز", nil
			}
			base, err := a.lookupBooking(code)
			if err != nil || strings.HasPrefix(base, "ما لگيت") {
				return base, err
			}
			return base + "\n\n" + a.bookingVerdict(code), nil
		},
	})
	a.AddTool(agentTool{Name: "customer_profile", Desc: "ملف زبون بالاسم أو الهاتف أو كود الزبون: شنو طلب منا (الخدمات وعدد الحجوزات)، آخر حركة ومتى، منو طلعله بآخر حجز، الشكاوى (عددها ومفتوحة لو لا ونصها)، والمشاكل (إلغاء، تأجيل، ما يرد، إنجاز جزئي، تكرار). إذا أكثر من زبون يطابق يرجّع قائمة تختار منها.",
		Props: map[string]any{"query": map[string]any{"type": "string", "description": "اسم الزبون أو رقم هاتفه أو كوده"}},
		Req:   []string{"query"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			return a.customerProfile(strings.TrimSpace(strArg(in, "query")))
		},
	})
}

// bookingVerdict الحكم على حجز: تكرار، الالتزام بالموعد، السلاسة.
func (a *MatrixAgent) bookingVerdict(code string) string {
	var b struct {
		ID          string     `db:"id"`
		CustomerID  string     `db:"customerId"`
		Address     *string    `db:"address"`
		Status      string     `db:"status"`
		CreatedAt   time.Time  `db:"createdAt"`
		Scheduled   *time.Time `db:"scheduledAt"`
		Arrived     *time.Time `db:"arrivedAt"`
		Started     *time.Time `db:"startedAt"`
		Completed   *time.Time `db:"completedAt"`
		Confirmed   *time.Time `db:"confirmedAt"`
		Postpones   int        `db:"postponeCount"`
		Partials    int        `db:"partialCount"`
		Stopped     *time.Time `db:"workStoppedAt"`
		StopWhy     *string    `db:"workStopReason"`
		Attempts    int        `db:"contactAttempts"`
		CancelWhy   *string    `db:"cancelReason"`
		PostponeWhy *string    `db:"postponeReason"`
	}
	if err := a.db.Get(&b, `SELECT id, "customerId", address, status::text AS status, "createdAt", "scheduledAt", "arrivedAt", "startedAt",
			"completedAt", "confirmedAt", COALESCE("postponeCount",0) AS "postponeCount", COALESCE("partialCount",0) AS "partialCount",
			"workStoppedAt", "workStopReason", COALESCE("contactAttempts",0) AS "contactAttempts", "cancelReason", "postponeReason"
		FROM "Booking" WHERE code ILIKE $1 ORDER BY "createdAt" DESC LIMIT 1`, code); err != nil {
		return "تعذر فحص الحجز"
	}
	var sb strings.Builder
	t := func(x *time.Time) string { return bagh(*x).Format("2006-01-02 15:04") }

	// ١. التكرار
	sb.WriteString("【التكرار】 ")
	var twins []struct {
		Code   string    `db:"code"`
		Status string    `db:"status"`
		At     time.Time `db:"at"`
	}
	_ = a.db.Select(&twins, `SELECT b2.code, b2.status::text AS status, COALESCE(b2."scheduledAt", b2."createdAt") AS at
		FROM "Booking" b2 WHERE b2."customerId" = $1 AND b2.id <> $2 AND b2."archivedAt" IS NULL
		  AND b2."partialJobBookingId" IS DISTINCT FROM $2 AND $2 IS DISTINCT FROM (SELECT "partialJobBookingId" FROM "Booking" WHERE id = $2)
		  AND $3::text IS NOT NULL AND ar_norm(b2.address) = ar_norm($3::text)
		  AND ABS(EXTRACT(EPOCH FROM (COALESCE(b2."scheduledAt", b2."createdAt") - $4::timestamp))) <= 86400`,
		b.CustomerID, b.ID, b.Address, func() time.Time {
			if b.Scheduled != nil {
				return *b.Scheduled
			}
			return b.CreatedAt
		}())
	var flagged []struct {
		Status     string  `db:"status"`
		Resolution *string `db:"resolution"`
	}
	_ = a.db.Select(&flagged, `SELECT status, resolution FROM "DuplicateCandidate" WHERE kind = 'BOOKING' AND ($1 IN ("entityAId", "entityBId"))`, b.ID)
	switch {
	case len(twins) > 0:
		codes := []string{}
		for _, x := range twins {
			codes = append(codes, fmt.Sprintf("%s (%s، %s)", x.Code, x.Status, bagh(x.At).Format("01-02")))
		}
		fmt.Fprintf(&sb, "مشتبه بيه تكرار ⚠️ — نفس الزبون ونفس العنوان بفارق يوم: %s.", strings.Join(codes, "، "))
	default:
		sb.WriteString("ماكو حجز ثاني لنفس الزبون بنفس العنوان بنفس اليوم — مو مكرر.")
	}
	for _, f := range flagged {
		res := "—"
		if f.Resolution != nil {
			res = *f.Resolution
		}
		fmt.Fprintf(&sb, " (بتدقيق التكرار: %s، القرار %s)", f.Status, res)
	}

	// ٢. الموعد
	sb.WriteString("\n【الموعد】 ")
	switch {
	case b.Scheduled == nil:
		fmt.Fprintf(&sb, "بعده ما انحط إله موعد (انفتح %s).", t(&b.CreatedAt))
	default:
		arrive := b.Arrived
		if arrive == nil {
			arrive = b.Started
		}
		if arrive != nil {
			d := int(arrive.Sub(*b.Scheduled).Minutes())
			switch {
			case d > 30:
				fmt.Fprintf(&sb, "تأخّر: الموعد %s ووصلوا/بدوا %s (متأخر %d دقيقة).", t(b.Scheduled), t(arrive), d)
			case d < -60:
				fmt.Fprintf(&sb, "وصلوا قبل الموعد بـ%d دقيقة (الموعد %s).", -d, t(b.Scheduled))
			default:
				fmt.Fprintf(&sb, "عالموعد: الموعد %s ووصلوا %s.", t(b.Scheduled), t(arrive))
			}
		} else if b.Completed == nil && b.Status != "CANCELLED" && time.Now().After(b.Scheduled.Add(2*time.Hour)) {
			fmt.Fprintf(&sb, "فات موعده (%s) وبعده ما مسجّل وصول ولا إنجاز ⚠️.", t(b.Scheduled))
		} else {
			fmt.Fprintf(&sb, "موعده %s.", t(b.Scheduled))
		}
		if b.Completed != nil {
			fmt.Fprintf(&sb, " خلص %s.", t(b.Completed))
		}
	}

	// ٣. السلاسة
	sb.WriteString("\n【السلاسة】 ")
	issues := []string{}
	if b.Postpones > 0 {
		s := fmt.Sprintf("تأجّل %d مرة", b.Postpones)
		if b.PostponeWhy != nil && *b.PostponeWhy != "" {
			s += " (آخر سبب: " + *b.PostponeWhy + ")"
		}
		issues = append(issues, s)
	}
	if b.Partials > 0 {
		issues = append(issues, fmt.Sprintf("إنجاز جزئي %d مرة", b.Partials))
	}
	if b.Stopped != nil {
		s := "توقف الشغل " + t(b.Stopped)
		if b.StopWhy != nil && *b.StopWhy != "" {
			s += " (" + *b.StopWhy + ")"
		}
		issues = append(issues, s)
	}
	if b.Attempts >= 2 {
		issues = append(issues, fmt.Sprintf("الزبون ما رد %d محاولات", b.Attempts))
	}
	if b.CancelWhy != nil && *b.CancelWhy != "" {
		issues = append(issues, "انلغى: "+*b.CancelWhy)
	}
	var complaints []struct {
		Status string `db:"status"`
		Desc   string `db:"description"`
	}
	_ = a.db.Select(&complaints, `SELECT status, description FROM "Complaint" WHERE "bookingId" = $1`, b.ID)
	for _, c := range complaints {
		issues = append(issues, fmt.Sprintf("شكوى (%s): %s", c.Status, cut(c.Desc, 120)))
	}
	if len(issues) == 0 {
		sb.WriteString("مشى بسلاسة — ماكو تأجيل ولا توقف ولا شكوى ✓")
	} else {
		sb.WriteString(strings.Join(issues, "؛ "))
	}
	return sb.String()
}

func cut(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "…"
}

// customerProfile ملف زبون — بلا اسمه ولا رقمه بالمخرجات.
func (a *MatrixAgent) customerProfile(q string) (string, error) {
	if q == "" || a.db == nil {
		return "اكتب اسم الزبون أو هاتفه أو كوده", nil
	}
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r >= '٠' && r <= '٩' {
			return '0' + (r - '٠')
		}
		return -1
	}, q)
	var matches []struct {
		ID   string    `db:"id"`
		Code int       `db:"customerCode"`
		Loc  *string   `db:"location"`
		N    int       `db:"n"`
		At   time.Time `db:"createdAt"`
	}
	err := a.db.Select(&matches, `SELECT c.id, c."customerCode", c.location, c."createdAt",
			(SELECT count(*) FROM "Booking" b WHERE b."customerId" = c.id) AS n
		FROM "Customer" c
		WHERE ($2 <> '' AND (regexp_replace(c.phone, '\D', '', 'g') LIKE '%' || right($2, 10) || '%' OR c."customerCode"::text = $2))
		   OR ar_norm(c.name) LIKE '%' || ar_norm($1) || '%'
		ORDER BY n DESC LIMIT 6`, q, digits)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "ما لگيت زبون يطابق «" + q + "».", nil
	}
	if len(matches) > 1 && !(len(digits) >= 7) {
		var sb strings.Builder
		fmt.Fprintf(&sb, "أكو %d زبائن يطابقون — اطلب من المستخدم يحدد بالكود أو الهاتف:\n", len(matches))
		for _, m := range matches {
			loc := "—"
			if m.Loc != nil && *m.Loc != "" {
				loc = *m.Loc
			}
			fmt.Fprintf(&sb, "- الزبون #%d — المنطقة %s، %d حجز، مسجّل من %s\n", m.Code, loc, m.N, bagh(m.At).Format("2006-01"))
		}
		return sb.String(), nil
	}
	m := matches[0]
	var sb strings.Builder
	fmt.Fprintf(&sb, "ملف الزبون #%d (مسجّل من %s)\n", m.Code, bagh(m.At).Format("2006-01-02"))

	// شنو طلب منا
	var svc []struct {
		Name string `db:"name"`
		N    int    `db:"n"`
	}
	_ = a.db.Select(&svc, `SELECT COALESCE(s.name, 'بلا خدمة') AS name, count(*) AS n FROM "Booking" b
		LEFT JOIN "Service" s ON s.id = b."serviceId" WHERE b."customerId" = $1 GROUP BY 1 ORDER BY 2 DESC`, m.ID)
	parts := []string{}
	for _, x := range svc {
		parts = append(parts, fmt.Sprintf("%s ×%d", x.Name, x.N))
	}
	fmt.Fprintf(&sb, "【شنو طلب منا】 %d حجز: %s\n", m.N, strings.Join(parts, "، "))
	var projects []string
	_ = a.db.Select(&projects, `SELECT p.code || ' «' || p.name || '» (' || p.stage || ')' FROM "Project" p WHERE right(regexp_replace(COALESCE(p.phone, ''), '\D', '', 'g'), 10) <> ''
		  AND right(regexp_replace(COALESCE(p.phone, ''), '\D', '', 'g'), 10) = right(regexp_replace((SELECT phone FROM "Customer" WHERE id = $1), '\D', '', 'g'), 10) ORDER BY p."createdAt" DESC LIMIT 5`, m.ID)
	if len(projects) > 0 {
		fmt.Fprintf(&sb, "مشاريع: %s\n", strings.Join(projects, "، "))
	}

	// آخر الحجوزات + آخر حركة + منو طلعله
	var recent []struct {
		Code    string     `db:"code"`
		Status  string     `db:"status"`
		Service *string    `db:"service"`
		Created time.Time  `db:"createdAt"`
		Last    time.Time  `db:"last"`
		Crew    *string    `db:"crew"`
		Sched   *time.Time `db:"scheduledAt"`
	}
	_ = a.db.Select(&recent, `SELECT b.code, b.status::text AS status, s.name AS service, b."createdAt", b."scheduledAt",
			GREATEST(b."createdAt", COALESCE(b."updatedAt", b."createdAt"), COALESCE(b."completedAt", b."createdAt"), COALESCE(b."startedAt", b."createdAt")) AS last,
			(SELECT string_agg(e.name, '، ') FROM "BookingAssignment" ba JOIN "Employee" e ON e.id = ba."employeeId" WHERE ba."bookingId" = b.id) AS crew
		FROM "Booking" b LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b."customerId" = $1 ORDER BY b."createdAt" DESC LIMIT 8`, m.ID)
	if len(recent) > 0 {
		r := recent[0]
		crew := "ماكو كادر مسجّل"
		if r.Crew != nil && *r.Crew != "" {
			crew = *r.Crew
		}
		fmt.Fprintf(&sb, "【آخر حركة】 %s على الحجز %s (%s)\n【منو طلعله آخر مرة】 %s\n", bagh(r.Last).Format("2006-01-02 15:04"), r.Code, r.Status, crew)
		sb.WriteString("【آخر الحجوزات】\n")
		for _, x := range recent {
			sv := "—"
			if x.Service != nil {
				sv = *x.Service
			}
			fmt.Fprintf(&sb, "- %s: %s — %s، انفتح %s\n", x.Code, sv, x.Status, bagh(x.Created).Format("2006-01-02"))
		}
	}

	// الشكاوى
	var comps []struct {
		Status string    `db:"status"`
		Type   *string   `db:"type"`
		Desc   string    `db:"description"`
		At     time.Time `db:"createdAt"`
		Rating *int      `db:"customerRating"`
	}
	_ = a.db.Select(&comps, `SELECT status, type::text AS type, description, "createdAt", "customerRating" FROM "Complaint" WHERE "customerId" = $1 ORDER BY "createdAt" DESC`, m.ID)
	if len(comps) == 0 {
		sb.WriteString("【الشكاوى】 ما سجّل ولا شكوى ✓\n")
	} else {
		open := 0
		for _, c := range comps {
			if c.Status == "NEW" || c.Status == "IN_PROGRESS" {
				open++
			}
		}
		fmt.Fprintf(&sb, "【الشكاوى】 %d شكوى (%d مفتوحة):\n", len(comps), open)
		for i, c := range comps {
			if i == 5 {
				break
			}
			fmt.Fprintf(&sb, "- %s [%s] %s", bagh(c.At).Format("2006-01-02"), c.Status, cut(c.Desc, 140))
			if c.Rating != nil {
				fmt.Fprintf(&sb, " (تقييمه %d/5)", *c.Rating)
			}
			sb.WriteString("\n")
		}
	}

	// المشاكل
	var p struct {
		Cancelled int `db:"cancelled"`
		Postponed int `db:"postponed"`
		NoAnswer  int `db:"noanswer"`
		Partial   int `db:"partial"`
		Stopped   int `db:"stopped"`
		Dups      int `db:"dups"`
	}
	_ = a.db.Get(&p, `SELECT
			count(*) FILTER (WHERE status::text = 'CANCELLED') AS cancelled,
			COALESCE(sum("postponeCount"), 0) AS postponed,
			count(*) FILTER (WHERE COALESCE("contactAttempts", 0) >= 2) AS noanswer,
			COALESCE(sum("partialCount"), 0) AS partial,
			count(*) FILTER (WHERE "workStoppedAt" IS NOT NULL) AS stopped,
			(SELECT count(*) FROM "DuplicateCandidate" d WHERE d.kind = 'BOOKING' AND d."entityAId" IN (SELECT id FROM "Booking" WHERE "customerId" = $1)) AS dups
		FROM "Booking" WHERE "customerId" = $1`, m.ID)
	probs := []string{}
	add := func(n int, s string) {
		if n > 0 {
			probs = append(probs, fmt.Sprintf(s, n))
		}
	}
	add(p.Cancelled, "%d حجز ملغي")
	add(p.Postponed, "%d تأجيل")
	add(p.NoAnswer, "%d حجز ما رد بيه")
	add(p.Partial, "%d إنجاز جزئي")
	add(p.Stopped, "%d توقف شغل")
	add(p.Dups, "%d حجز مشتبه بتكراره")
	if len(probs) == 0 {
		sb.WriteString("【المشاكل】 ماكو مشاكل مسجّلة وياه ✓")
	} else {
		sb.WriteString("【المشاكل】 " + strings.Join(probs, "، "))
	}
	return sb.String(), nil
}
