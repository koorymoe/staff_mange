package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

func TestRankCrew(t *testing.T) {
	rows := []repository.CrewCandidateRow{
		{EmployeeID: "a", Name: "أحمد", ServiceHasSkills: true, HasSkill: true, DoneCount: 10, ProblemCount: 5, DayLoad: 0},
		{EmployeeID: "b", Name: "باسم", ServiceHasSkills: true, HasSkill: true, DoneCount: 10, ProblemCount: 0, DayLoad: 1},
		{EmployeeID: "c", Name: "جاسم", ServiceHasSkills: true, HasSkill: false, DoneCount: 0, DayLoad: 0},
		{EmployeeID: "d", Name: "داود", ServiceHasSkills: true, HasSkill: true, DoneCount: 2, DayLoad: 0},
	}
	out, ok := rankCrew(rows, 3)
	if !ok || len(out) != 3 {
		t.Fatalf("ok=%v len=%d", ok, len(out))
	}
	// أحمد: 30+15-20=25، باسم: 30+15-10=35، داود: 30+3=33، جاسم: 0
	if out[0].EmployeeID != "b" || out[1].EmployeeID != "d" || out[2].EmployeeID != "a" {
		t.Fatalf("ترتيب غلط: %s %s %s", out[0].EmployeeID, out[1].EmployeeID, out[2].EmployeeID)
	}
	if out[2].ProblemRatePct == nil || *out[2].ProblemRatePct != 50 {
		t.Fatalf("نسبة المشاكل لازم ٥٠٪")
	}
	if out[1].ProblemRatePct != nil {
		t.Fatalf("أقل من ٥ منجز ما تنحسب نسبة")
	}
	if _, ok := rankCrew([]repository.CrewCandidateRow{{EmployeeID: "x", ServiceHasSkills: true}}, 3); ok {
		t.Fatalf("بلا مهارة ولا خبرة = ما نقترح")
	}
}

func TestCurveStatus(t *testing.T) {
	cases := []struct {
		s    []int
		want string
	}{
		{[]int{1, 2, 3}, CurveStatusTooEarly},
		{[]int{1, 2, 3, 4}, CurveStatusOK},
		{[]int{5, 4, 3, 2}, CurveStatusFollowUp},
		{[]int{0, 0, 0, 0, 0}, CurveStatusFollowUp},
		{[]int{2, 5, 5, 5}, CurveStatusOK}, // ثابت على القمة
		{[]int{6, 5, 5, 5}, CurveStatusFollowUp},
		{[]int{1, 3, 2, 2, 2}, CurveStatusFollowUp},
		{[]int{1, 3, 2, 2, 4}, CurveStatusOK},
	}
	for _, c := range cases {
		if got := curveStatus(c.s); got != c.want {
			t.Errorf("%v: got %s want %s", c.s, got, c.want)
		}
	}
}

func TestBucketCurve(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, debriefLoc)
	now := start.AddDate(0, 0, 20) // أسبوعين مكتملين + ثالث ناقص
	ev := []repository.EmployeeEvent{
		{Kind: "DONE", At: start.Add(time.Hour), PaperDue: true, PaperOK: true},
		{Kind: "DONE", At: start.Add(2 * time.Hour), PaperDue: true},
		{Kind: "STOP", At: start.AddDate(0, 0, 8)},
		{Kind: "COMPLAINT", At: start.AddDate(0, 0, 15)},
		{Kind: "DONE", At: start.AddDate(0, 0, -1)},
	}
	w := bucketCurve(start, now, ev)
	if len(w) != 3 || !w[0].Complete || !w[1].Complete || w[2].Complete {
		t.Fatalf("أسابيع غلط: %+v", w)
	}
	if w[0].Completed != 2 || w[0].PaperPct == nil || *w[0].PaperPct != 50 || w[0].Score != 2 {
		t.Fatalf("الأسبوع ١: %+v", w[0])
	}
	if w[1].Score != -2 || w[2].Score != -3 {
		t.Fatalf("نقاط: %d %d", w[1].Score, w[2].Score)
	}
}

func TestFlagVehicles(t *testing.T) {
	rows := []repository.VehicleCostRow{
		{ID: "1", MaintCost12m: 100},
		{ID: "2", MaintCost12m: 100},
		{ID: "3", MaintCost12m: 150, IncidentCost: 60}, // 210 ≥ 2×100
		{ID: "4", MaintCost12m: 50, Incidents180d: 3},
	}
	out, med, ok := flagVehicles(rows)
	if !ok || med != 100 {
		t.Fatalf("median=%v ok=%v", med, ok)
	}
	if len(out) != 2 || out[0].ID != "3" || out[1].ID != "4" {
		t.Fatalf("got %+v", out)
	}
	// أسطول صغير: قاعدة الكلفة تنطفي
	out, _, ok = flagVehicles(rows[:2])
	if ok || len(out) != 0 {
		t.Fatalf("أسطول صغير لازم بلا مقارنة")
	}
}

func TestParseISOWeek(t *testing.T) {
	m, err := parseISOWeek("2026-W01")
	if err != nil || m.Format("2006-01-02") != "2025-12-29" {
		t.Fatalf("%v %v", m, err)
	}
	m, _ = parseISOWeek("2026-W40")
	if m.Format("2006-01-02") != "2026-09-28" || isoWeekLabel(m) != "2026-W40" {
		t.Fatalf("W40 = %v", m)
	}
	if _, err := parseISOWeek("2026-40"); err == nil {
		t.Fatalf("صيغة غلط لازم ترفض")
	}
	if _, err := parseISOWeek("2025-W53"); err == nil {
		t.Fatalf("2025 ما بيها W53")
	}
}

func TestWeekWindowsAndMovers(t *testing.T) {
	mon := time.Date(2026, 9, 28, 0, 0, 0, 0, debriefLoc)
	now := mon.Add(50 * time.Hour)
	cf, ct, pf, pt, partial := weekWindows(mon, now)
	if !partial || ct != now || pt.Sub(pf) != ct.Sub(cf) {
		t.Fatalf("نافذة غلط")
	}
	ev := []repository.EmployeeEvent{
		{EmployeeID: "a", Name: "أ", Kind: "DONE", At: cf.Add(time.Hour)},
		{EmployeeID: "a", Name: "أ", Kind: "DONE", At: cf.Add(2 * time.Hour)},
		{EmployeeID: "b", Name: "ب", Kind: "DONE", At: pf.Add(time.Hour)},
		{EmployeeID: "b", Name: "ب", Kind: "COMPLAINT", At: cf.Add(time.Hour)},
		{EmployeeID: "c", Name: "ج", Kind: "DONE", At: pt.Add(time.Hour)}, // برا النافذة المقصوصة
	}
	up, down := weeklyMovers(ev, cf, ct, pf, pt, 3)
	if len(up) != 1 || up[0].EmployeeID != "a" || up[0].Delta != 2 {
		t.Fatalf("up=%+v", up)
	}
	if len(down) != 1 || down[0].EmployeeID != "b" || down[0].Delta != -4 {
		t.Fatalf("down=%+v", down)
	}
}
