package service

import (
	"math"
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

func near(a, b float64) bool { return math.Abs(a-b) < 0.01 }

// (دازاي 11-18) الفني: ليدر 30 · حضور 20 · عدّة 15 · مهام 10 · زملاء 5 = 80، وماتركس 20.
func TestTechScoreWeights(t *testing.T) {
	rows := []StaffScore{{ID: "t", Group: GroupTechs, RelParts: []ReliabilityPart{
		{Key: "ATTENDANCE", Pct: 100}, {Key: "INVENTORY", Pct: 100}, {Key: "TASK_DUE", Pct: 50}, {Key: "PEERS", Pct: 75},
		{Key: "COMPLAINTS", Pct: 0}, // ما يدخل برقم الفني
	}}}
	hum := []repository.HumanRow{{RateeID: "t", Stage: "LEADER_CREW", Score: 4}, {RateeID: "t", Stage: "LEADER_CREW", Score: 5},
		{RateeID: "t", Stage: "OPS_DAILY", Score: 1}} // إدارة العمليات ما تدخل
	applyTechScores(rows, hum)
	s := rows[0]
	// ليدر 90%×30=27 · حضور 20 · عدّة 15 · مهام 50%×10=5 · زملاء 75%×5=3.75 = 70.75
	// ماتركس: المهام 50 < 80 ← 80 ⇒ 80×20% = 16 ⇒ 86.75
	if s.Final == nil || !near(*s.Final, 86.75) {
		t.Fatalf("final = %v", s.Final)
	}
	if s.LegacyBase != nil {
		t.Fatal("legacy without old final")
	}
	if len(s.MatrixNotes) != 1 || !s.TechScore {
		t.Fatalf("notes = %v", s.MatrixNotes)
	}
}

func TestTechScoreRedistributesMissing(t *testing.T) {
	// ماكو ليدر ولا مهام: الحضور 20 والعدّة 15 والزملاء 5 تتوزع على ٨٠
	rows := []StaffScore{{ID: "t", Group: GroupTechs, RelParts: []ReliabilityPart{
		{Key: "ATTENDANCE", Pct: 50}, {Key: "INVENTORY", Pct: 100}, {Key: "PEERS", Pct: 100}}}}
	applyTechScores(rows, nil)
	// حضور 50%×40 = 20 · عدّة 30 · زملاء 10 ⇒ 60 · ماتركس 100−30 (حضور) = 70 ⇒ 14 ⇒ 74
	if f := rows[0].Final; f == nil || !near(*f, 74) {
		t.Fatalf("final = %v", f)
	}
	// غير الفنيين ما يتأثرون
	other := []StaffScore{{ID: "x", Group: GroupLeaders}}
	applyTechScores(other, nil)
	if other[0].TechScore {
		t.Fatal("leader got tech score")
	}
}

func TestRevenueTargetTech130(t *testing.T) {
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	now := to.Add(-time.Hour)
	f1, f2 := 80.0, 80.0
	rows := []StaffScore{{ID: "t", Group: GroupTechs, Final: &f1}, {ID: "l", Group: GroupTechnical, Final: &f2}}
	rev := []repository.RevenueRow{{EmployeeID: "t", Kind: "WORK", Amount: 1_000_000}, {EmployeeID: "l", Kind: "WORK", Amount: 1_000_000}}
	sal := map[string]float64{"t": 1_000_000, "l": 1_000_000}
	ApplyRevenue(rows, rev, sal, from, to, now)
	// الفني غطّى ١٠٠٪ من راتبه = ٧٧٪ من هدفه (١٣٠) فينزل؛ التقني هدفه ١٠٠ فما ينزل
	if *rows[0].Final >= 80 || rows[0].RevenueTarget != 130 {
		t.Fatalf("tech final = %v target %v", *rows[0].Final, rows[0].RevenueTarget)
	}
	if *rows[1].Final < 80 {
		t.Fatalf("leader dropped: %v", *rows[1].Final)
	}
}

// (دازاي 11-18) الوقت: +١٥ دقيقة عادي، +٦٠ نص، أكثر صفر.
func TestJobTimePoints(t *testing.T) {
	for _, c := range []struct {
		got, want float64
		pts       int
	}{{100, 90, 2}, {105, 90, 2}, {140, 90, 1}, {150, 90, 1}, {210, 90, 0}, {60, 90, 2}} {
		if p := JobTimePoints(c.got, c.want); p != c.pts {
			t.Errorf("%v مقابل %v: %d نريد %d", c.got, c.want, p, c.pts)
		}
	}
	s1, s2, l := "cam", "net", "L"
	jobs := []repository.LeaderJob{{LeaderID: &l, ServiceID: &s1, Minutes: 95}, {LeaderID: &l, ServiceID: &s1, Minutes: 220}, {LeaderID: &l, ServiceID: &s2, Minutes: 999}}
	got := LeaderTimes(jobs, map[string]float64{"cam": 90}) // net ماله وقت ← ما ينحسب
	if got["L"] != [3]int{2, 4, 1} {
		t.Fatalf("got %v", got["L"])
	}
}

// الليدر: الوقت 20 · مسؤوله 20 · ورق 15 · فنيينه 5 · حضور 10 · عدّة 5 · مهام 3 · زملاء 2
func TestLeaderScore(t *testing.T) {
	rows := []StaffScore{{ID: "L", Group: GroupLeaders,
		RelParts:   []ReliabilityPart{{Key: "ATTENDANCE", Pct: 100}, {Key: "INVENTORY", Pct: 100}, {Key: "TASK_DUE", Pct: 100}, {Key: "PEERS", Pct: 100}},
		ruleTotals: map[string][2]int{"PAPER": {2, 4}, "REPORT_MATCH": {2, 4}, "RATING": {2, 2}}}}
	hum := []repository.HumanRow{{RateeID: "L", Stage: "COORD_LEADER", Score: 5}, {RateeID: "L", Stage: "AUDIT", Score: 1}}
	applyRoleScores(rows, hum, map[string][3]int{"L": {2, 4, 1}})
	s := rows[0]
	// وقت 50%×20=10 · مسؤول 100%×20=20 · ورق 50%×15=7.5 · فنيين 5 · حضور 10 · عدّة 5 · مهام 3 · زملاء 2 = 62.5
	// ماتركس: وقت <70 (−30) وورق <80 (−25) ⇒ 45 ⇒ 9 ⇒ 71.5
	if s.Final == nil || !near(*s.Final, 71.5) || len(s.MatrixNotes) != 2 {
		t.Fatalf("final=%v notes=%v", s.Final, s.MatrixNotes)
	}
}

// (دازاي 11-19) المنسّق: الإداري ٨٠ شغل حجز، المدير ٤٠ + ٤٠ خطة، النائب ٥٠ + ٣٠، وماتركس ٢٠.
func TestCoordinatorShares(t *testing.T) {
	mk := func(id string) StaffScore {
		return StaffScore{ID: id, Group: GroupCoordinator, RelParts: []ReliabilityPart{{Key: "ATTENDANCE", Pct: 100}},
			ruleTotals: map[string][2]int{"CONTACT": {2, 2}, "CONFIRM": {2, 2}, "CREW": {2, 2}}}
	}
	rows := []StaffScore{mk("adm"), mk("dir"), mk("dep")}
	dir, dep, plan := "DIRECTOR", "DEPUTY", 50.0
	coord := map[string]repository.CoordFact{
		"adm": {Confirmed: 10, Postpones: 5},
		"dir": {Position: &dir, Plan: &plan, PlanItems: 4, Confirmed: 10},
		"dep": {Position: &dep, Plan: &plan, PlanItems: 2, Confirmed: 10},
	}
	applyRoleScores(rows, nil, nil, coord)
	// الإداري: تثبيت 25 + كادر 20 + تأجيل 50%×10 + حضور 5 = 55 من 60 ⇒ ×80/60 = 73.33 · ماتركس 100−20 (تأجيل) = 16 ⇒ 89.33
	// المدير: شغل 100%×40 = 40 + خطة 50%×40 = 20 · ماتركس 100−25 (خطة) = 15 ⇒ 75
	// النائب: 50 + 15 + 15 = 80
	want := map[string]float64{"adm": 89.33, "dir": 75, "dep": 80}
	for _, s := range rows {
		if s.Final == nil || !near(*s.Final, want[s.ID]) {
			t.Errorf("%s final = %v نريد %v", s.ID, s.Final, want[s.ID])
		}
	}
	if v, _ := PostponePct(3, 2); v != 0 {
		t.Fatal("postpone below zero")
	}
}
