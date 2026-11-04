package service

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ 🧪 ٩.٦ البيئة التجريبية الكاملة (دازاي 11-02) ═══
// موظفين بكل الأدوار بالهيكلين، كل واحد إله «شخصية» معروفة (ممتاز، بطيء، يكذب
// بالتقرير، ما يكتب تقرير، ما يقيّم، موظف مكتب شغله ما يتسجّل...)، ونحاكي ٣ أيام
// شغل كاملة، ونشغّل كل محرّكات ماتركس، ونقارن حكمه بالحقيقة المعروفة.
//
// ⚠️ يشتغل بس على قاعدة اسمها ينتهي بـ _sim (نسخة) — أبداً مو على الحقيقية:
//   createdb -T staffmange staffmange_sim
//   SIM_DATABASE_URL=postgres://…/staffmange_sim SIM_REPORT=../docs/محاكاة-ماتركس.md go test ./internal/service -run Simulation -v

type simPerson struct {
	id, name, role, persona, expect string
	leader                          bool
	org                             []string // divisionId, unitId(""=NULL), position
}

var simPeople = []simPerson{
	{id: "sim_l_good", name: "ليدر ممتاز", role: "TECHNICIAN", leader: true, persona: "يشتغل زين، يقيّم كادره، يكتب تقريره بالأكواد", expect: "أعلى الليدرية"},
	{id: "sim_l_neg", name: "ليدر ما يقيّم", role: "TECHNICIAN", leader: true, persona: "شغله بطيء، وما يقيّم كادره أبد", expect: "ينكتب بسجل الإهمال وينزل تقييمه"},
	{id: "sim_t_good", name: "فني ممتاز", role: "TECHNICIAN", persona: "سريع وبالموعد وتقريره صادق", expect: "أعلى الفنيين"},
	{id: "sim_t_slow", name: "فني بطيء", role: "TECHNICIAN", persona: "يوصل متأخر ويطوّل ٣ أضعاف، تقريره صادق", expect: "يحتاج تدريب (سرعة)"},
	{id: "sim_t_liar", name: "فني يكذب بالتقرير", role: "TECHNICIAN", persona: "يكتب إنه خلّص حجز مو حجزه", expect: "تقاريره ما تطابق"},
	{id: "sim_t_lazy", name: "فني ما يكتب تقرير", role: "TECHNICIAN", persona: "يداوم بس ما يكتب تقرير يومين من ٣", expect: "ينزل بتقرير الإنجاز"},
	{id: "sim_sales", name: "موظف حجوزات", role: "SALES", persona: "يسجّل ويثبّت الحجوزات ويقيّم الليدر", expect: "تقييم زين"},
	{id: "sim_fin", name: "محاسب مكتب", role: "FINANCE", persona: "يشتغل ورق ومكالمات — ما يتسجّل بالنظام، تقريره صادق", expect: "ما لازم ينظلم"},
	{id: "sim_mon_old", name: "مراقب (هيكل قديم)", role: "MONITOR", persona: "يقيّم موظفين المكاتب يومياً", expect: "بلا إهمال"},
	{id: "sim_ops_dir", name: "مدير العمليات (هيكل جديد)", role: "ENGINEER", persona: "منصبه مدير إدارة العمليات، يقيّم الكوادر يومياً", expect: "ينعرف كمقيّم عمليات", org: []string{"ops", "", "DIRECTOR"}},
	{id: "sim_mon_new", name: "مراقب (هيكل جديد)", role: "SALES", persona: "دوره الأصلي مبيعات، منصبه بوحدة المراقبة", expect: "ينعرف كمراقب من منصبه", org: []string{"admin", "admin.monitoring", "MEMBER"}},
	{id: "sim_hr_new", name: "إداري كوادر (هيكل جديد)", role: "TECHNICAL", persona: "منصبه بوحدة إدارة الكوادر، يعتمد على زميله بالتقييم", expect: "يكفي واحد يقيّم — ما ينظلم", org: []string{"ops", "ops.staff", "MEMBER"}},
}

func TestSimulation(t *testing.T) {
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

	// ── تنظيف محاكاة سابقة ──
	for _, q := range []string{
		`DELETE FROM "TrainingNeed" WHERE "employeeId" LIKE 'sim\_%'`,
		`DELETE FROM "DailyReportMatch" WHERE "employeeId" LIKE 'sim\_%'`,
		`DELETE FROM "MatrixScore" WHERE "employeeId" LIKE 'sim\_%'`,
		`DELETE FROM "EmployeeNegligence" WHERE "employeeId" LIKE 'sim\_%'`,
		`DELETE FROM "DailyWorkReport" WHERE "employeeId" LIKE 'sim\_%'`,
		`DELETE FROM "StaffRating" WHERE "raterId" LIKE 'sim\_%' OR "rateeId" LIKE 'sim\_%'`,
		`DELETE FROM "CrewRating" WHERE "leaderId" LIKE 'sim\_%'`,
		`DELETE FROM "BookingAssignment" WHERE "bookingId" LIKE 'sim\_%'`,
		`DELETE FROM "Booking" WHERE id LIKE 'sim\_%'`,
		`DELETE FROM "Attendance" WHERE "employeeId" LIKE 'sim\_%'`,
		`DELETE FROM "OrgAssignment" WHERE "employeeId" LIKE 'sim\_%'`,
		`DELETE FROM "Employee" WHERE id LIKE 'sim\_%'`,
		`DELETE FROM "Customer" WHERE id = 'sim_cust'`,
	} {
		ex(q)
	}

	// بالنسخة: مدير العمليات الحالي ينسحب حتى مدير المحاكاة ياخذ مكانه (فريد)
	ex(`UPDATE "OrgAssignment" SET "endedAt" = now() WHERE "divisionId" = 'ops' AND "unitId" IS NULL AND position = 'DIRECTOR' AND "endedAt" IS NULL`)
	// ── الموظفين ──
	created := time.Now().AddDate(0, 0, -30)
	for _, p := range simPeople {
		ex(`INSERT INTO "Employee" (id, name, role, "isLeader", status, "createdAt") VALUES ($1, $2, $3::"EmployeeRole", $4, 'ACTIVE', $5)`,
			p.id, "🧪 "+p.name, p.role, p.leader, created)
		if p.org != nil {
			var unit any
			if p.org[1] != "" {
				unit = p.org[1]
			}
			ex(`INSERT INTO "OrgAssignment" ("employeeId", "divisionId", "unitId", position) VALUES ($1, $2, $3, $4)`, p.id, p.org[0], unit, p.org[2])
		}
	}
	ex(`INSERT INTO "Customer" (id, name, phone) VALUES ('sim_cust', '🧪 زبون المحاكاة', '07000000000')`)
	var svc string
	if err := db.Get(&svc, `SELECT id FROM "Service" ORDER BY "createdAt" LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	// أيام الإلزام تبدي قبل المحاكاة
	start := companyDay(-5)
	ex(`UPDATE "OrgGovernorState" SET value = LEAST(value, $1) WHERE key IN ('daily_report_since', 'rating_duty_since')`, start)

	loc := debriefLoc
	at := func(day string, hh, mm int) time.Time { // وقت بغداد ← UTC بلا منطقة (مثل الجداول)
		d, _ := time.ParseInLocation("2006-01-02", day, loc)
		return d.Add(time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute).UTC()
	}
	days := []string{companyDay(-3), companyDay(-2), companyDay(-1)}
	for di, day := range days {
		for _, p := range simPeople {
			ex(`INSERT INTO "Attendance" (id, "employeeId", "checkIn", "checkOut", date) VALUES ($1, $2, $3, $4, $3)`,
				fmt.Sprintf("sim_att_%s_%d", p.id, di), p.id, at(day, 8, 0), at(day, 16, 0))
		}
		// حجزين باليوم: A للكادر الزين، B للكادر البطيء
		type bk struct {
			code, leader      string
			techs             []string
			late, wait, hours int // دقائق تأخير، انتظار قبل البدء، ساعات شغل
		}
		for _, b := range []bk{
			{code: fmt.Sprintf("SIM%dA", di+1), leader: "sim_l_good", techs: []string{"sim_t_good", "sim_t_liar"}, late: 0, wait: 5, hours: 2},
			{code: fmt.Sprintf("SIM%dB", di+1), leader: "sim_l_neg", techs: []string{"sim_t_slow", "sim_t_lazy"}, late: 50, wait: 40, hours: 6},
		} {
			id := "sim_" + b.code
			sched := at(day, 9, 0)
			arr := sched.Add(time.Duration(b.late) * time.Minute)
			st := arr.Add(time.Duration(b.wait) * time.Minute)
			ex(`INSERT INTO "Booking" (id, code, "customerId", "serviceId", status, "scheduledAt", "departedAt", "arrivedAt", "startedAt", "completedAt",
				"createdById", "confirmedByEmployeeId", "confirmedAt", "createdAt", "updatedAt")
				VALUES ($1, $2, 'sim_cust', $3, 'COMPLETED', $4, $5, $6, $7, $8, 'sim_sales', 'sim_sales', $9, $9, $8)`,
				id, b.code, svc, sched, sched.Add(-40*time.Minute), arr, st, st.Add(time.Duration(b.hours)*time.Hour), at(day, 7, 30))
			roles := []string{"TECH_1", "TECH_2", "TECH_3"}
			for i, e := range append([]string{b.leader}, b.techs...) {
				ex(`INSERT INTO "BookingAssignment" (id, "bookingId", "employeeId", role) VALUES ($1, $2, $3, $4::"TechnicianRole")`,
					fmt.Sprintf("%s_%d", id, i), id, e, roles[i])
			}
			// الليدر الزين يقيّم كادره؛ الثاني لا
			if b.leader == "sim_l_good" {
				for _, tch := range b.techs {
					score := 5
					if tch == "sim_t_liar" {
						score = 3
					}
					ex(`INSERT INTO "CrewRating" (id, "bookingId", "leaderId", "technicianId", score) VALUES ($1, $2, $3, $4, $5)`,
						id+"_cr_"+tch, id, b.leader, tch, score)
				}
			}
			// موظف الحجوزات يقيّم الليدر
			ldScore := 5
			if b.leader == "sim_l_neg" {
				ldScore = 3
			}
			ex(`INSERT INTO "StaffRating" ("raterId", "rateeId", stage, "bookingId", score) VALUES ('sim_sales', $1, 'COORD_LEADER', $2, $3)`, b.leader, id, ldScore)
		}
		// مدير العمليات (هيكل جديد) يقيّم كل الكوادر يومياً
		for _, r := range [][2]any{{"sim_l_good", 5}, {"sim_l_neg", 3}, {"sim_t_good", 5}, {"sim_t_slow", 2}, {"sim_t_liar", 3}, {"sim_t_lazy", 3}} {
			ex(`INSERT INTO "StaffRating" ("raterId", "rateeId", stage, period, score) VALUES ('sim_ops_dir', $1, 'OPS_DAILY', $2, $3)`, r[0], day, r[1])
		}
		// تقارير الإنجاز
		reports := map[string]string{
			"sim_l_good":  fmt.Sprintf("قدت الكادر على الحجز SIM%dA وخلصناه بالموعد وسلّمنا للزبون", di+1),
			"sim_l_neg":   fmt.Sprintf("اشتغلنا على الحجز SIM%dB وطوّل بسبب الموقع", di+1),
			"sim_t_good":  fmt.Sprintf("ركبت شغل الحجز SIM%dA كامل وجربته ويا الزبون", di+1),
			"sim_t_slow":  fmt.Sprintf("اشتغلت على الحجز SIM%dB طول اليوم وتعبنا بيه", di+1),
			"sim_t_liar":  fmt.Sprintf("خلصت الحجز SIM%dB لحالي وكان صعب جداً", di+1), // مو حجزه
			"sim_sales":   fmt.Sprintf("سجّلت وثبّت الحجزين SIM%dA و SIM%dB ويا الزبائن", di+1, di+1),
			"sim_fin":     "راجعت حسابات الشهر وطابقت الوصولات ويا الصندوق وسويت مكالمات للموردين",
			"sim_mon_old": "راجعت تقارير الموظفين اليوم وقيّمت موظفين المكاتب",
			"sim_ops_dir": "قيّمت الكوادر اليوم وتابعت التأخير بحجز الكادر الثاني",
			"sim_mon_new": "تابعت الشكاوي وراجعت ملفات المكتب",
			"sim_hr_new":  "تابعت دوام الكوادر وتوزيعهم على الحجوزات",
		}
		if di == 2 { // اليوم الأخير بس
			reports["sim_t_lazy"] = fmt.Sprintf("اشتغلت ويا الكادر على الحجز SIM%dB", di+1)
		}
		for emp, body := range reports {
			late := emp == "sim_l_neg" && di == 0
			ex(`INSERT INTO "DailyWorkReport" ("employeeId", day, body, late) VALUES ($1, $2::date, $3, $4)`, emp, day, body, late)
		}
		// المراقب القديم يقيّم موظفين المكاتب؛ المراقب الجديد يعتمد عليه («يكفي واحد»)
		for _, r := range [][2]any{{"sim_fin", 4}, {"sim_sales", 5}, {"sim_mon_new", 4}, {"sim_hr_new", 4}, {"sim_ops_dir", 5}} {
			ex(`INSERT INTO "StaffRating" ("raterId", "rateeId", stage, period, score) VALUES ('sim_mon_old', $1, 'MONITOR_DAILY', $2, $3)`, r[0], day, r[1])
		}
	}

	// ── تشغيل كل محرّكات ماتركس ──
	duty := NewRatingDuty(db)
	if _, err := duty.Sweep(); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	m := NewReportMatcher(db)
	for _, d := range days {
		if _, err := m.SweepDay(d); err != nil {
			t.Fatalf("match %s: %v", d, err)
		}
	}
	chain := NewMatrixChainService(repository.NewMatrixChainRepository(db), func(string) string { return "" })
	svcScore := NewMatrixScoreService(repository.NewMatrixScoreRepository(db), chain, nil, nil)
	if _, err := svcScore.Run(time.Now()); err != nil {
		t.Logf("⚠️ Run رجّع أخطاء (غير قاتلة): %v", err)
	}
	people, err := svcScore.repo.Scorables()
	if err != nil {
		t.Fatal(err)
	}
	sim := []repository.Scorable{}
	for _, p := range people {
		if strings.HasPrefix(p.ID, "sim_") {
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
	rule := func(emp, r string) (int, int) {
		var x struct{ P, M int }
		_ = db.QueryRow(`SELECT COALESCE(sum(points),0), COALESCE(sum("maxPoints"),0) FROM "MatrixScore" WHERE "employeeId" = $1 AND rule = $2 AND "cancelledAt" IS NULL`, emp, r).Scan(&x.P, &x.M)
		return x.P, x.M
	}
	final := func(id string) float64 {
		if f := by[id].Final; f != nil {
			return *f
		}
		return -1
	}

	// ── الأحكام مقابل الحقيقة ──
	type check struct {
		name string
		ok   bool
		note string
	}
	checks := []check{}
	add := func(name string, ok bool, note string) { checks = append(checks, check{name, ok, note}) }
	pct := func(p, mx int) string {
		if mx == 0 {
			return "—"
		}
		return fmt.Sprintf("%d٪ (%d/%d)", p*100/mx, p, mx)
	}

	add("الفني الممتاز أعلى من البطيء والكذاب والي ما يكتب", final("sim_t_good") > final("sim_t_slow") && final("sim_t_good") > final("sim_t_liar") && final("sim_t_good") > final("sim_t_lazy"),
		fmt.Sprintf("ممتاز %.0f · بطيء %.0f · كذاب %.0f · ما يكتب %.0f", final("sim_t_good"), final("sim_t_slow"), final("sim_t_liar"), final("sim_t_lazy")))
	add("الليدر الممتاز أعلى من الليدر الي ما يقيّم", final("sim_l_good") > final("sim_l_neg"), fmt.Sprintf("%.0f مقابل %.0f", final("sim_l_good"), final("sim_l_neg")))
	var negN int
	_ = db.Get(&negN, `SELECT count(*) FROM "EmployeeNegligence" WHERE "employeeId" = 'sim_l_neg'`)
	p, mx := rule("sim_l_neg", "RATING_DUTY")
	add("الليدر الي ما يقيّم انكتب بسجل الإهمال ونزل تقييمه", negN > 0 && mx > 0 && p < mx, fmt.Sprintf("سجلات إهمال %d · واجب التقييم %s", negN, pct(p, mx)))
	var goodNeg int
	_ = db.Get(&goodNeg, `SELECT count(*) FROM "EmployeeNegligence" WHERE "employeeId" IN ('sim_l_good','sim_ops_dir','sim_mon_old','sim_sales')`)
	add("الي قيّموا بوقتهم ما انكتبوا مهملين", goodNeg == 0, fmt.Sprintf("%d سجل", goodNeg))
	var hrNeg, monNewNeg int
	_ = db.Get(&hrNeg, `SELECT count(*) FROM "EmployeeNegligence" WHERE "employeeId" = 'sim_hr_new'`)
	_ = db.Get(&monNewNeg, `SELECT count(*) FROM "EmployeeNegligence" WHERE "employeeId" = 'sim_mon_new'`)
	add("الهيكل الجديد: إداري الكوادر والمراقب ما ينظلمون لأن زميلهم قيّم («يكفي واحد»)", hrNeg == 0 && monNewNeg == 0, fmt.Sprintf("إداري الكوادر %d · المراقب الجديد %d", hrNeg, monNewNeg))
	add("الهيكل الجديد: مدير العمليات ينعرف كمقيّم عمليات من منصبه", duty.is(opsRaterSQL, "sim_ops_dir"), "")
	add("الهيكل الجديد: المراقب ينعرف من منصبه (دوره الأصلي مبيعات)", duty.is(monitorRaterSQL, "sim_mon_new"), "")
	p, mx = rule("sim_t_liar", "REPORT_MATCH")
	add("الكذاب: تقاريره ما تطابق الشغل المسجّل", mx > 0 && p < mx, "صدق التقارير "+pct(p, mx))
	p, mx = rule("sim_t_good", "REPORT_MATCH")
	add("الصادق: تقاريره تطابق", mx > 0 && p == mx, "صدق التقارير "+pct(p, mx))
	p, mx = rule("sim_t_lazy", "DAILY_REPORT")
	add("الي ما يكتب تقرير نزل بتقرير الإنجاز", mx > 0 && p < mx, "تقرير الإنجاز "+pct(p, mx))
	slow, err := StudyTraining(db, "sim_t_slow")
	if err != nil {
		t.Fatal(err)
	}
	add("البطيء: ماتركس يحسم يحتاج تدريب", slow.Verdict == "NEEDS", trainingVerdictAr(slow.Verdict)+" — "+strings.Join(slow.Areas, "، ")+" · "+slow.Evidence)
	good, _ := StudyTraining(db, "sim_t_good")
	add("الممتاز: ماتركس ما يطلب إله تدريب", good != nil && good.Verdict != "NEEDS", trainingVerdictAr(good.Verdict)+" · "+good.Evidence)
	p, mx = rule("sim_fin", "REPORT_MATCH")
	add("موظف المكتب الصادق ما ينظلم بمطابقة التقرير (بلا أثر ينتظر حكم المراقب)", mx == 0 || p == mx, "صدق التقارير "+pct(p, mx)+" — شغله ورق ومكالمات ما تتسجّل بالنظام")
	add("موظف المكتب عنده تقييم (مو «ماكو شغل كافي»)", by["sim_fin"].Max > 0, fmt.Sprintf("نقاط ماتركس %d/%d", by["sim_fin"].Earned, by["sim_fin"].Max))

	// ── التقرير ──
	var b strings.Builder
	w := func(f string, a ...any) { fmt.Fprintf(&b, f+"\n", a...) }
	pass := 0
	for _, c := range checks {
		if c.ok {
			pass++
		}
	}
	w("# 🧪 محاكاة ماتركس — البيئة التجريبية الكاملة (٩.٦)")
	w("")
	w("> انولّدت تلقائياً %s على نسخة منفصلة من القاعدة (`%s`) — ولا شي انلمس بالقاعدة الحقيقية.", time.Now().In(loc).Format("2006-01-02 15:04"), dbname)
	w("> ١٢ موظف وهمي بكل الأدوار بالهيكلين، ٣ أيام شغل (%s ← %s)، ٦ حجوزات. بعدين شغّلنا: واجب التقييم، مطابقة التقارير، تقييم ماتركس، ودراسة التدريب.", days[0], days[2])
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
	w("## الموظفين: الشخصية مقابل حكم ماتركس")
	w("")
	w("| الموظف | الشخصية (الحقيقة) | المتوقع | ماتركس ٪ | البشر ٪ | النهائي | أكثر شي نزّله |")
	w("|---|---|---|---|---|---|---|")
	sort.SliceStable(simPeople, func(i, j int) bool { return final(simPeople[i].id) > final(simPeople[j].id) })
	f1 := func(x *float64) string {
		if x == nil {
			return "—"
		}
		return fmt.Sprintf("%.0f", *x)
	}
	for _, sp := range simPeople {
		s := by[sp.id]
		loss := "—"
		if len(s.TopLosses) > 0 {
			loss = s.TopLosses[0].Title
		}
		w("| %s | %s | %s | %s | %s | %s | %s |", sp.name, sp.persona, sp.expect, f1(s.MatrixPct), f1(s.HumanPct), f1(s.Final), loss)
	}
	w("")
	w("## الثغرات الي طلعت")
	w("")
	gaps := 0
	for _, c := range checks {
		if !c.ok {
			gaps++
			w("- ❌ **%s** — %s", c.name, c.note)
		}
	}
	if gaps == 0 {
		w("- ماكو — كل الأحكام طابقت الحقيقة.")
	}
	report := b.String()
	t.Log("\n" + report)
	if p := os.Getenv("SIM_REPORT"); p != "" {
		if err := os.WriteFile(p, []byte(report), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
