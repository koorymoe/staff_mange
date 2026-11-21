package service

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ 🧪 محاكاة التقييم الجديد لكل الأدوار (دازاي 11-21) ═══
// «سوي موظفين لكل الأدوار، وجرب تسوي حجز وترتيب ومشروع وشكوى وكلشي… وانطيني نتيجة».
// لكل دور موظفين: «زين» و«مقصّر» بحقيقة معروفة، ٥ أيام شغل، حجوزات، مشاريع، شكاوى،
// اتصالات جودة، مهام، حضور، تقارير، تقييم مراقب. نشغّل ماتركس ونقارن حكمه بالحقيقة.
//
// ⚠️ يشتغل بس على قاعدة اسمها ينتهي بـ _sim:
//   SIM_DATABASE_URL=postgres://…/x_sim SIM_REPORT=../docs/محاكاة-التقييم-الجديد.md go test ./internal/service -run SimulationRoles -v

type rolePersona struct {
	id, name, role string
	leader         bool
	division       string
	good           bool
}

var rolePeople = []rolePersona{
	{id: "rs_t_good", name: "فني زين", role: "TECHNICIAN", good: true},
	{id: "rs_t_bad", name: "فني مقصّر", role: "TECHNICIAN"},
	{id: "rs_l_good", name: "ليدر زين", role: "TECHNICIAN", leader: true, good: true},
	{id: "rs_l_bad", name: "ليدر مقصّر", role: "TECHNICIAN", leader: true},
	{id: "rs_q_good", name: "جودة زين", role: "QUALITY_ENGINEER", good: true},
	{id: "rs_q_bad", name: "جودة مقصّر", role: "QUALITY_ENGINEER"},
	{id: "rs_s_good", name: "مبيعات زين", role: "SALES", good: true},
	{id: "rs_s_bad", name: "مبيعات مقصّر", role: "SALES"},
	{id: "rs_f_good", name: "حسابات زين", role: "FINANCE", good: true},
	{id: "rs_f_bad", name: "حسابات مقصّر", role: "FINANCE"},
	{id: "rs_m_good", name: "مراقب زين", role: "MONITOR", good: true},
	{id: "rs_m_bad", name: "مراقب مقصّر", role: "MONITOR"},
	{id: "rs_p_good", name: "كميات زين", role: "PROCUREMENT_ADMIN", good: true},
	{id: "rs_p_bad", name: "كميات مقصّر", role: "PROCUREMENT_ADMIN"},
	{id: "rs_d_good", name: "تصميم زين", role: "DESIGNER", good: true},
	{id: "rs_d_bad", name: "تصميم مقصّر", role: "DESIGNER"},
	{id: "rs_pm_good", name: "مدير مشاريع زين", role: "PROJECT_MANAGER", good: true},
	{id: "rs_pm_bad", name: "مدير مشاريع مقصّر", role: "PROJECT_MANAGER"},
	{id: "rs_it_good", name: "دعم تقني زين", role: "IT_SUPPORT", good: true},
	{id: "rs_it_bad", name: "دعم تقني مقصّر", role: "IT_SUPPORT"},
	{id: "rs_e_good", name: "تقني زين", role: "ENGINEER", good: true},
	{id: "rs_e_bad", name: "تقني مقصّر", role: "ENGINEER"},
	{id: "rs_dec_good", name: "ديكور زين", role: "TECHNICIAN", division: "DECOR", good: true},
	{id: "rs_dec_bad", name: "ديكور مقصّر", role: "TECHNICIAN", division: "DECOR"},
}

func TestSimulationRoles(t *testing.T) {
	url := os.Getenv("SIM_DATABASE_URL")
	if url == "" {
		t.Skip("SIM_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	var dbname string
	_ = db.Get(&dbname, `SELECT current_database()`)
	if !strings.HasSuffix(dbname, "_sim") {
		t.Fatalf("رفض: القاعدة %q مو نسخة محاكاة (لازم تنتهي بـ _sim)", dbname)
	}
	ex := func(q string, a ...any) {
		t.Helper()
		if _, err := db.Exec(q, a...); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}
	for _, q := range []string{
		`DELETE FROM "MatrixScore" WHERE "employeeId" LIKE 'rs\_%'`,
		`DELETE FROM "StaffRating" WHERE "raterId" LIKE 'rs\_%' OR "rateeId" LIKE 'rs\_%'`,
		`DELETE FROM "CrewRating" WHERE "leaderId" LIKE 'rs\_%'`,
		`DELETE FROM "QualityFollowUp" WHERE "bookingId" LIKE 'rs\_%'`,
		`DELETE FROM "Complaint" WHERE id LIKE 'rs\_%'`,
		`DELETE FROM "BookingAssignment" WHERE "bookingId" LIKE 'rs\_%'`,
		`DELETE FROM "ProjectStageLog" WHERE "projectId" LIKE 'rs\_%'`,
		`DELETE FROM "Project" WHERE id = 'rs_prj_from_booking'`,
		`DELETE FROM "Project" WHERE id LIKE 'rs\_%'`,
		`DELETE FROM "Booking" WHERE id LIKE 'rs\_%'`,
		`DELETE FROM "ExtraTask" WHERE id LIKE 'rs\_%'`,
		`DELETE FROM "DailyWorkReport" WHERE "employeeId" LIKE 'rs\_%'`,
		`DELETE FROM "Attendance" WHERE "employeeId" LIKE 'rs\_%'`,
		`DELETE FROM "Employee" WHERE id LIKE 'rs\_%'`,
		`DELETE FROM "Customer" WHERE id = 'rs_cust'`,
	} {
		ex(q)
	}
	created := time.Now().AddDate(0, -2, 0)
	for _, p := range rolePeople {
		div := "ENGINEERING"
		if p.division != "" {
			div = p.division
		}
		ex(`INSERT INTO "Employee" (id, name, role, "isLeader", status, "createdAt", division) VALUES ($1, $2, $3::"EmployeeRole", $4, 'ACTIVE', $5, $6)`,
			p.id, "🧪 "+p.name, p.role, p.leader, created, div)
	}
	ex(`INSERT INTO "Customer" (id, name, phone) VALUES ('rs_cust', '🧪 زبون محاكاة الأدوار', '07000000011')`)
	var svc string
	if err := db.Get(&svc, `SELECT id FROM "Service" ORDER BY "createdAt" LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	loc := debriefLoc
	at := func(day string, hh, mm int) time.Time {
		d, _ := time.ParseInLocation("2006-01-02", day, loc)
		return d.Add(time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute).UTC()
	}
	days := []string{companyDay(-5), companyDay(-4), companyDay(-3), companyDay(-2), companyDay(-1)}
	monitor := "rs_m_good"
	for di, day := range days {
		for _, p := range rolePeople {
			// المقصّر يغيب يومين من ٥
			if !p.good && (di == 1 || di == 3) {
				continue
			}
			in := at(day, 8, 0)
			if !p.good {
				in = at(day, 9, 10) // يتأخر
			}
			ex(`INSERT INTO "Attendance" (id, "employeeId", "checkIn", "checkOut", date) VALUES ($1, $2, $3, $4, $3)`,
				fmt.Sprintf("rs_att_%s_%d", p.id, di), p.id, in, at(day, 16, 0))
			if p.good || di%2 == 0 {
				ex(`INSERT INTO "DailyWorkReport" ("employeeId", day, body, late) VALUES ($1, $2::date, $3, false)`,
					p.id, day, "اشتغلت اليوم على شغلي اليومي وخلصت المطلوب مني وسلّمته")
			}
			// تقييم المراقب اليومي لموظفين المكاتب
			score := 5
			if !p.good {
				score = 2
			}
			if p.id != monitor {
				ex(`INSERT INTO "StaffRating" ("raterId", "rateeId", stage, period, score) VALUES ($1, $2, 'MONITOR_DAILY', $3, $4)`, monitor, p.id, day, score)
			}
		}
		// مهام إضافية: الزين يخلص بموعده، المقصّر يفوّتها
		for _, p := range rolePeople {
			due := at(day, 14, 0)
			var done any = due.Add(-time.Hour)
			if !p.good {
				done = nil
			}
			ex(`INSERT INTO "ExtraTask" (id, title, "assignedToId", "dueAt", "doneAt", status) VALUES ($1, 'مهمة محاكاة', $2, $3, $4, 'OPEN')`,
				fmt.Sprintf("rs_task_%s_%d", p.id, di), p.id, due, done)
		}
		// حجزين باليوم: كادر زين (مبيعات زين، جودة زين يتصل بنفس اليوم) وكادر مقصّر
		for _, b := range []struct {
			code, leader, tech, sales, quality string
			late, hours                        int
			qualityAfter                       time.Duration
			crewScore                          int
		}{
			{fmt.Sprintf("RS%dG", di+1), "rs_l_good", "rs_t_good", "rs_s_good", "rs_q_good", 0, 2, 3 * time.Hour, 5},
			{fmt.Sprintf("RS%dB", di+1), "rs_l_bad", "rs_t_bad", "rs_s_bad", "rs_q_bad", 60, 6, 40 * time.Hour, 2},
		} {
			id := "rs_" + b.code
			sched := at(day, 9, 0)
			arr := sched.Add(time.Duration(b.late) * time.Minute)
			done := arr.Add(time.Duration(b.hours) * time.Hour)
			ex(`INSERT INTO "Booking" (id, code, "customerId", "serviceId", status, "scheduledAt", "departedAt", "arrivedAt", "startedAt", "completedAt",
				"createdById", "confirmedByEmployeeId", "confirmedAt", "createdAt", "updatedAt")
				VALUES ($1, $2, 'rs_cust', $3, 'COMPLETED', $4, $5, $6, $6, $7, $8, $8, $9, $9, $7)`,
				id, b.code, svc, sched, sched.Add(-40*time.Minute), arr, done, b.sales, at(day, 7, 30))
			decor := "rs_dec_good"
			if b.leader == "rs_l_bad" {
				decor = "rs_dec_bad"
			}
			for i, e := range []string{b.leader, b.tech, decor} {
				ex(`INSERT INTO "BookingAssignment" (id, "bookingId", "employeeId", role) VALUES ($1, $2, $3, $4::"TechnicianRole")`,
					fmt.Sprintf("%s_%d", id, i), id, e, []string{"TECH_1", "TECH_2", "TECH_3"}[i])
			}
			// الديكور: الليدر يقيّمه (الزين ٥، المقصّر ٢ مؤكَّد)
			ex(`INSERT INTO "CrewRating" (id, "bookingId", "leaderId", "technicianId", score, "reviewStatus") VALUES ($1, $2, 'rs_l_good', $3, $4, 'CONFIRMED')`, id+"_crd", id, decor, b.crewScore)
			if b.leader == "rs_l_good" { // الليدر الزين يقيّم فنيّه، المقصّر لا
				ex(`INSERT INTO "CrewRating" (id, "bookingId", "leaderId", "technicianId", score) VALUES ($1, $2, $3, $4, $5)`, id+"_cr", id, b.leader, b.tech, b.crewScore)
			} else {
				// ليدر ثاني (زين) قيّم الفني المقصّر سيء — مؤكَّد من مدير الإدارة
				ex(`INSERT INTO "CrewRating" (id, "bookingId", "leaderId", "technicianId", score, "reviewStatus") VALUES ($1, $2, 'rs_l_good', $3, $4, 'CONFIRMED')`, id+"_cr", id, b.tech, b.crewScore)
			}
			ex(`INSERT INTO "QualityFollowUp" (id, "bookingId", "customerId", status, "contactedAt", "contactedByEmployeeId", "createdAt")
				VALUES ($1, $2, 'rs_cust', 'DONE', $3, $4, $5)`, id+"_q", id, done.Add(b.qualityAfter), b.quality, done)
		}
	}
	// ── شكاوى: الجودة الزين يتصل خلال ساعة ويحلها بيوم؛ المقصّر يتصل بعد يوم ونص وبعدها مفتوحة ٨ أيام ──
	cAt := time.Now().UTC().Add(-8 * 24 * time.Hour)
	ex(`INSERT INTO "Complaint" (id, "customerId", description, status, "createdByEmployeeId", "assignedToEmployeeId", "createdAt", "contactedAt", "contactedById", "resolvedAt")
		VALUES ('rs_c_good', 'rs_cust', 'الكاميرا الأمامية تطفي كل ليلة بعد الساعة عشرة', 'RESOLVED', 'rs_s_good', 'rs_q_good', $1, $2, 'rs_q_good', $3)`,
		cAt, cAt.Add(time.Hour), cAt.Add(24*time.Hour))
	ex(`INSERT INTO "Complaint" (id, "customerId", description, status, "createdByEmployeeId", "assignedToEmployeeId", "createdAt", "contactedAt", "contactedById")
		VALUES ('rs_c_bad', 'rs_cust', 'الجهاز ما يشتغل', 'NEW', 'rs_s_bad', 'rs_q_bad', $1, $2, 'rs_q_bad')`, cAt, cAt.Add(36*time.Hour))
	ex(`UPDATE "Complaint" SET "bookingId" = 'rs_RS1G' WHERE id = 'rs_c_good'`)
	// ── المبيعات: حجز الزين تحول لمشروع، وحجزين للمقصّر انلغوا ──
	ex(`INSERT INTO "Project" (id, code, name, stage, "createdAt", "updatedAt", "bookingId") VALUES ('rs_prj_from_booking', 'RSPFB', '🧪 مشروع من حجز', '4. العقد', now(), now(), 'rs_RS5G')`)
	for i := 1; i <= 2; i++ {
		ex(`INSERT INTO "Booking" (id, code, "customerId", "serviceId", status, "scheduledAt", "createdById", "createdAt", "updatedAt")
			VALUES ($1, $2, 'rs_cust', $3, 'CANCELLED', now() + interval '1 day', 'rs_s_bad', now() - interval '2 day', now())`,
			fmt.Sprintf("rs_RSX%d", i), fmt.Sprintf("RSX%d", i), svc)
	}
	// ── مشاريع: كل «زين» عنده مشروع يتحرك، كل «مقصّر» من التصميم والتقني ومدير المشاريع والجودة عنده مشروع واكف ٦٠ يوم ──
	old := time.Now().UTC().AddDate(0, 0, -60)
	for _, owner := range []string{"rs_d_good", "rs_d_bad", "rs_pm_good", "rs_pm_bad", "rs_e_good", "rs_e_bad", "rs_q_good", "rs_q_bad"} {
		id := "rs_prj_" + strings.TrimPrefix(owner, "rs_")
		ex(`INSERT INTO "Project" (id, code, name, stage, "createdAt", "updatedAt", "responsibleEmployeeId", "delegatedToEmployeeId")
			VALUES ($1, $2, $3, '2. مرحلة الكشف', $4, $4, $5, $5)`, id, strings.ToUpper(id), "🧪 مشروع "+owner, old, owner)
		if strings.HasSuffix(owner, "_good") {
			ex(`INSERT INTO "ProjectStageLog" ("projectId", "toStage", "createdAt") VALUES ($1, '2. مرحلة الكشف', now() - interval '1 day')`, id)
			ex(`UPDATE "Project" SET "updatedAt" = now() WHERE id = $1`, id)
		}
	}

	// ── تشغيل ماتركس ──
	chain := NewMatrixChainService(repository.NewMatrixChainRepository(db), func(string) string { return "" })
	projects := NewMatrixProjectService(repository.NewMatrixProjectRepository(db))
	svcScore := NewMatrixScoreService(repository.NewMatrixScoreRepository(db), chain, nil, projects.Report)
	if _, err := svcScore.Run(time.Now()); err != nil {
		t.Logf("⚠️ Run رجّع أخطاء (غير قاتلة): %v", err)
	}
	people, err := svcScore.repo.Scorables()
	if err != nil {
		t.Fatal(err)
	}
	sim := []repository.Scorable{}
	for _, p := range people {
		if strings.HasPrefix(p.ID, "rs_") {
			sim = append(sim, p)
		}
	}
	scores, err := svcScore.compute(sim, time.Now().AddDate(0, 0, -10), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]StaffScore{}
	for _, s := range scores {
		by[s.ID] = s
	}
	final := func(id string) float64 {
		if f := by[id].Final; f != nil {
			return *f
		}
		return -1
	}
	part := func(id, key string) (float64, bool) {
		for _, p := range by[id].TechParts {
			if p.Key == key {
				return p.Pct, true
			}
		}
		return 0, false
	}
	rule := func(emp, r string) (int, int) {
		var p, m int
		_ = db.QueryRow(`SELECT COALESCE(sum(points),0), COALESCE(sum("maxPoints"),0) FROM "MatrixScore" WHERE "employeeId" = $1 AND rule = $2 AND "cancelledAt" IS NULL`, emp, r).Scan(&p, &m)
		return p, m
	}

	type check struct {
		name string
		ok   bool
		note string
	}
	checks := []check{}
	add := func(name string, ok bool, note string) { checks = append(checks, check{name, ok, note}) }

	// ١) لكل دور: الزين أعلى من المقصّر، والرقمين بين ٠ و١٠٠
	groups := map[string][2]string{}
	for _, p := range rolePeople {
		base := strings.TrimSuffix(strings.TrimSuffix(p.id, "_good"), "_bad")
		g := groups[base]
		if p.good {
			g[0] = p.id
		} else {
			g[1] = p.id
		}
		groups[base] = g
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		g := groups[k]
		a, b := final(g[0]), final(g[1])
		label := strings.TrimPrefix(by[g[0]].GroupLabel, "")
		add(fmt.Sprintf("%s: الزين أعلى من المقصّر", label), a > b && a >= 0 && b >= 0 && a <= 100,
			fmt.Sprintf("زين %.0f · مقصّر %.0f (الفئة %s)", a, b, by[g[0]].Group))
	}
	// ٢) الجودة
	p, m := rule("rs_q_good", "QUALITY")
	p2, m2 := rule("rs_q_bad", "QUALITY")
	add("الجودة: الي يتصل بنفس اليوم ياخذ الاتصال كامل، والي يتصل بعد يوم ونص نص", m > 0 && p == m && m2 > 0 && p2 < m2,
		fmt.Sprintf("زين %d/%d · مقصّر %d/%d", p, m, p2, m2))
	p, m = rule("rs_q_good", "COMPLAINT_CONTACT")
	p2, m2 = rule("rs_q_bad", "COMPLAINT_CONTACT")
	add("الشكوى: الاتصال خلال ساعة كامل، وبعد يوم ونص صفر", p == 2 && m == 2 && p2 == 0 && m2 == 2, fmt.Sprintf("زين %d/%d · مقصّر %d/%d", p, m, p2, m2))
	p, m = rule("rs_q_good", "COMPLAINT_RESOLVE")
	p2, m2 = rule("rs_q_bad", "COMPLAINT_RESOLVE")
	add("الشكوى: الحل بيوم كامل، والمفتوحة ٨ أيام صفر", p == 2 && m == 2 && p2 == 0 && m2 == 2, fmt.Sprintf("زين %d/%d · مقصّر %d/%d", p, m, p2, m2))
	p, m = rule("rs_s_good", "COMPLAINT_LOG")
	p2, m2 = rule("rs_s_bad", "COMPLAINT_LOG")
	add("المبيعات: الشكوى المربوطة بالحجز وشرحها واضح كاملة، والناقصة صفر", p == 2 && m == 2 && p2 == 0 && m2 == 2, fmt.Sprintf("زين %d/%d · مقصّر %d/%d", p, m, p2, m2))
	v1, _ := part("rs_s_good", "CONVERSION")
	v2, _ := part("rs_s_bad", "CONVERSION")
	add("المبيعات: الحجز الي صار مشروع ينحسب «تم»، والملغي ينزّل", v1 == 100 && v2 < 100, fmt.Sprintf("زين %.0f٪ · مقصّر %.0f٪", v1, v2))
	_, okQ := part("rs_q_good", "COMPLAINT_RESOLVE")
	add("الجودة: بند «حل الشكوى» يطلع بتفاصيل التقييم", okQ, "")
	// ٣) المشاريع لكل الأدوار
	for _, id := range []string{"rs_d_good", "rs_e_good", "rs_q_good", "rs_pm_good"} {
		v, ok := part(id, "PROJECTS")
		add("المشاريع: "+by[id].GroupLabel+" الي مشروعه يتحرك ياخذ البند كامل", ok && v >= 99, fmt.Sprintf("%.0f٪", v))
	}
	for _, id := range []string{"rs_d_bad", "rs_e_bad", "rs_q_bad"} {
		v, ok := part(id, "PROJECTS")
		add("المشاريع: "+by[id].GroupLabel+" الي مشروعه واكف ينحسب عليه (مو مكافأة)", ok && v < 50 && by[id].ExtraJobs == 0, fmt.Sprintf("%.0f٪", v))
	}
	_, okDec := part("rs_dec_good", "LEADER")
	add("الديكور: صار ينقيّم بقواعد الفني (تقييم الليدر، الحضور، الجرد…) مو المعادلة القديمة", okDec && by["rs_dec_good"].Group == GroupDecor,
		fmt.Sprintf("زين %.0f · مقصّر %.0f", final("rs_dec_good"), final("rs_dec_bad")))
	if v, ok := part("rs_t_good", "PROJECTS"); ok {
		add("الي ما عنده مشروع ما يطلعله بند المشاريع", false, fmt.Sprintf("طلع %.0f٪", v))
	} else {
		add("الي ما عنده مشروع ما يطلعله بند المشاريع (وزنه يتوزع)", true, "")
	}
	// ٤) سلامة الأرقام
	bad := []string{}
	for _, s := range scores {
		if s.Final == nil || math.IsNaN(*s.Final) || *s.Final < 0 || *s.Final > 100 {
			bad = append(bad, s.Name)
		}
		if s.TechParts != nil {
			w := 0.0
			for _, p := range s.TechParts {
				w += p.Weight
			}
			if math.Abs(w-80) > 0.5 && math.Abs(w-100) > 0.5 {
				bad = append(bad, fmt.Sprintf("%s (مجموع الأوزان %.1f)", s.Name, w))
			}
		}
	}
	add("كل موظف إله رقم بين ٠ و١٠٠ وأوزانه مجموعها صحيح", len(bad) == 0, strings.Join(bad, "، "))

	// ── التقرير ──
	var b strings.Builder
	w := func(f string, a ...any) { fmt.Fprintf(&b, f+"\n", a...) }
	pass := 0
	for _, c := range checks {
		if c.ok {
			pass++
		}
	}
	w("# 🧪 محاكاة التقييم الجديد — كل الأدوار")
	w("")
	w("> انولّدت %s على نسخة منفصلة (`%s`). %d موظف وهمي (زين ومقصّر لكل دور)، ٥ أيام شغل، %d حجز، %d مشاريع، شكوتين.",
		time.Now().In(loc).Format("2006-01-02 15:04"), dbname, len(rolePeople), len(days)*2, 8)
	w("")
	w("## النتيجة: %d من %d حكم صح", pass, len(checks))
	w("")
	w("| | الحكم | التفصيل |")
	w("|---|---|---|")
	for _, c := range checks {
		mark := "✅"
		if !c.ok {
			mark = "❌"
		}
		w("| %s | %s | %s |", mark, c.name, strings.ReplaceAll(c.note, "|", "/"))
	}
	w("")
	w("## كل موظف وبنوده")
	w("")
	w("| الموظف | الفئة | النهائي | البنود |")
	w("|---|---|---|---|")
	sort.SliceStable(rolePeople, func(i, j int) bool { return rolePeople[i].id < rolePeople[j].id })
	for _, rp := range rolePeople {
		s := by[rp.id]
		parts := []string{}
		for _, p := range s.TechParts {
			parts = append(parts, fmt.Sprintf("%s %.0f٪×%.0f", p.Label, p.Pct, p.Weight))
		}
		if len(parts) == 0 {
			parts = append(parts, "(المعادلة القديمة)")
		}
		w("| %s | %s | %.0f | %s |", rp.name, s.GroupLabel, final(rp.id), strings.Join(parts, " · "))
	}
	report := b.String()
	t.Log("\n" + report)
	if p := os.Getenv("SIM_REPORT"); p != "" {
		if err := os.WriteFile(p, []byte(report), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range checks {
		if !c.ok {
			t.Errorf("❌ %s — %s", c.name, c.note)
		}
	}
}
