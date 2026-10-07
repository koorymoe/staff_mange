package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

func TestStationPoints(t *testing.T) {
	cases := map[string]int{ChainOK: 2, ChainLate: 1, ChainIssue: 1, ChainMissed: 0}
	for st, want := range cases {
		if got, ok := StationPoints(st); !ok || got != want {
			t.Fatalf("%s → %d %v", st, got, ok)
		}
	}
	for _, st := range []string{ChainWaiting, ChainNA} {
		if _, ok := StationPoints(st); ok {
			t.Fatalf("%s لازم ما تنحسب", st)
		}
	}
}

func TestTaskPoints(t *testing.T) {
	due := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	at := func(h int) *time.Time { x := due.Add(time.Duration(h) * time.Hour); return &x }
	if p, _ := TaskPoints(due, at(-1), due); p != 2 {
		t.Fatal("بموعدها = ٢")
	}
	if p, _ := TaskPoints(due, at(20), due); p != 1 {
		t.Fatal("خلال يوم = ١")
	}
	if p, _ := TaskPoints(due, at(30), due); p != 0 {
		t.Fatal("بعد يوم = ٠")
	}
	if _, ok := TaskPoints(due, nil, due.Add(2*time.Hour)); ok {
		t.Fatal("لسه بمهلة اليوم: ما تنحسب")
	}
	if p, ok := TaskPoints(due, nil, due.Add(30*time.Hour)); !ok || p != 0 {
		t.Fatal("ما خلصت = ٠")
	}
}

func TestFinalScore(t *testing.T) {
	if f, _ := FinalScore(80, true, 60, true); f < 71.99 || f > 72.01 {
		t.Fatalf("0.6*80+0.4*60 = 72, got %v", f)
	}
	if f, _ := FinalScore(80, true, 0, false); f != 80 {
		t.Fatal("بلا بشر = ماتركس")
	}
	if _, ok := FinalScore(0, false, 0, false); ok {
		t.Fatal("بلا شي = ماكو")
	}
}

func TestBuildScores(t *testing.T) {
	now := time.Now()
	people := []repository.Scorable{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}}
	pts := []repository.MatrixScoreRow{
		{EmployeeID: "a", Source: "BOOKING", Rule: "PAPER", Points: 1, MaxPoints: 2},
		{EmployeeID: "a", Source: "DAY", Rule: "ATTENDANCE", Points: 2, MaxPoints: 2},
		{EmployeeID: "a", Source: "DAY", Rule: "ATTENDANCE", Points: 0, MaxPoints: 2, CancelledAt: &now},
	}
	hum := []repository.HumanRow{{RateeID: "a", Score: 4, Stage: "MONITOR_PERIODIC"}, {RateeID: "a", Score: 2, Stage: "MONITOR_PERIODIC"}, {RateeID: "a", Score: 1, Stage: "LEADER_CREW"}}
	out := BuildScores(people, pts, hum, nil)
	a := out[0]
	// الحضور يدخل بالاعتمادية بس، مو بالتقييم (قرار (ع) 10-06).
	if a.Earned != 1 || a.Max != 2 || *a.MatrixPct != 50 || *a.HumanPct != 60 {
		t.Fatalf("%+v", a)
	}
	if a.Reliability == nil || len(a.RelParts) == 0 || a.RelParts[0].Key != "ATTENDANCE" || a.RelParts[0].Pct != 100 {
		t.Fatalf("reliability %+v", a.RelParts)
	}
	if *a.Final < 53.99 || *a.Final > 54.01 { // 0.6*50 + 0.4*60
		t.Fatalf("final %v", *a.Final)
	}
	if len(a.TopLosses) != 1 || a.TopLosses[0].Rule != "PAPER" {
		t.Fatalf("losses %+v", a.TopLosses)
	}
	if out[1].Final != nil {
		t.Fatal("B بلا شي")
	}
}

func TestPeriodKey(t *testing.T) {
	if k := PeriodKey(time.Date(2026, 10, 3, 12, 0, 0, 0, debriefLoc)); k != "2026-10-A" {
		t.Fatal(k)
	}
	if k := PeriodKey(time.Date(2026, 10, 20, 12, 0, 0, 0, debriefLoc)); k != "2026-10-B" {
		t.Fatal(k)
	}
}

func TestProjectBookingChain(t *testing.T) {
	conf := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	exec := conf.AddDate(0, 0, 20)
	f := &repository.ChainFacts{Status: "CONFIRMED", ToProjects: true, ConfirmedAt: &conf}
	if !AtProjects(f) {
		t.Fatal("محوّل وما وصل التنفيذ = عند المشاريع")
	}
	if !crewStart(f).Equal(conf) {
		t.Fatal("قبل التنفيذ يبقى التثبيت")
	}
	f.ProjectExecAt = &exec
	if AtProjects(f) {
		t.Fatal("وصل التنفيذ = رجع للإداري")
	}
	if !crewStart(f).Equal(exec) {
		t.Fatal("الكادر ينحسب من رجوعه، مو من التثبيت")
	}
	plain := &repository.ChainFacts{Status: "CONFIRMED", ConfirmedAt: &conf}
	if AtProjects(plain) || !crewStart(plain).Equal(conf) {
		t.Fatal("حجز عادي ما يتأثر")
	}
}

func TestProjectCompletionBlock(t *testing.T) {
	v := 1000.0
	if ProjectCompletionBlock(nil, 0, 0, "") == "" {
		t.Fatal("بلا قيمة = رفض")
	}
	if ProjectCompletionBlock(&v, 0, 0, "") == "" {
		t.Fatal("بلا دفعات = رفض")
	}
	if ProjectCompletionBlock(&v, 400, 1, "") == "" {
		t.Fatal("باقي بلا سبب = رفض")
	}
	if ProjectCompletionBlock(&v, 400, 1, "يدفع بعد شهر") != "" {
		t.Fatal("باقي بسبب = يمشي")
	}
	if ProjectCompletionBlock(&v, 1000, 2, "") != "" {
		t.Fatal("مدفوع كامل = يمشي")
	}
}

func TestScoreGroups(t *testing.T) {
	cases := []struct {
		p    repository.Scorable
		want string
	}{
		{repository.Scorable{Role: "SALES"}, GroupSales},
		{repository.Scorable{Role: "TECHNICIAN"}, GroupTechs},
		{repository.Scorable{Role: "TECHNICIAN", IsLeader: true}, GroupLeaders},
		{repository.Scorable{Role: "TECHNICIAN", IsServiceManager: true}, GroupTechnical},
		{repository.Scorable{Role: "MONITOR"}, GroupMonitors},
		{repository.Scorable{Role: "FINANCE"}, GroupFinance},
		{repository.Scorable{Role: "ENGINEER"}, GroupTechnical},
	}
	for _, c := range cases {
		if g := ScoreGroup(c.p); g != c.want {
			t.Fatalf("%+v → %s want %s", c.p, g, c.want)
		}
	}
	// الفني ما ينحسبله تقييم المراقب الدوري — بس الليدر.
	people := []repository.Scorable{{ID: "t", Role: "TECHNICIAN"}}
	hum := []repository.HumanRow{{RateeID: "t", Score: 1, Stage: "MONITOR_PERIODIC"}, {RateeID: "t", Score: 5, Stage: "LEADER_CREW"}}
	out := BuildScores(people, nil, hum, nil)
	if out[0].HumanCount != 1 || *out[0].HumanAvg != 5 {
		t.Fatalf("tech human %+v", out[0])
	}
	if periodicAllowed("MONITOR_PERIODIC", repository.Scorable{Role: "MONITOR"}) || !periodicAllowed("ADMIN_PERIODIC", repository.Scorable{Role: "MONITOR"}) {
		t.Fatal("المراقب يقيّمه المدير بس")
	}
}

func TestReliability(t *testing.T) {
	rules := map[string][2]int{"ATTENDANCE": {16, 20}, "INVENTORY": {4, 4}}
	parts, v := BuildReliability(rules, repository.ReliabilityFact{Sessions: 10, Auto: 2, Complaints: 1})
	if v == nil || len(parts) != 4 {
		t.Fatalf("parts %d", len(parts))
	}
	// (80 + 80 + 100 + 75) / 4
	if *v < 83.74 || *v > 83.76 {
		t.Fatalf("got %v", *v)
	}
}
