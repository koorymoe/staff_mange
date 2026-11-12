package service

import (
	"strings"
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// (ل) 11-12 دازاي: «كل واحد يتقيّم حسب شغله» — المصمم على تسليم أقسام التصاميم بموعدها،
// وماتركس ينبّه على الي نقاطهم بس من الدوام والتقارير (ماكو شغل دور محسوب).
func TestDesignSectionPoints(t *testing.T) {
	due := time.Date(2026, 11, 10, 21, 0, 0, 0, time.UTC)
	now := due.Add(48 * time.Hour)
	early, late := due.Add(-time.Hour), due.Add(time.Hour)
	cases := []struct {
		status string
		sub    *time.Time
		now    time.Time
		pts    int
		ok     bool
	}{
		{"REVIEW", &early, now, 2, true},
		{"APPROVED", &late, now, 1, true},
		{"RETURNED", &early, now, 1, true},
		{"IN_PROGRESS", nil, now, 0, true},
		{"IN_PROGRESS", nil, due.Add(-time.Hour), 0, false}, // موعده بعده ما فات
	}
	for i, c := range cases {
		p, why, ok := DesignSectionPoints(c.status, due, c.sub, c.now)
		if p != c.pts || ok != c.ok || (ok && why == "") {
			t.Fatalf("حالة %d: %d %v %q", i, p, ok, why)
		}
	}
}

func TestRoleWorkWarning(t *testing.T) {
	people := []repository.Scorable{{ID: "d", Name: "مصممة", Role: "DESIGNER"}, {ID: "x", Name: "محاسب", Role: "FINANCE"}}
	at := time.Now()
	pts := []repository.MatrixScoreRow{
		{EmployeeID: "d", Rule: "ATTENDANCE", Points: 2, MaxPoints: 2, At: at},
		{EmployeeID: "d", Rule: "DAILY_REPORT", Points: 2, MaxPoints: 2, At: at},
		{EmployeeID: "x", Rule: "DAILY_REPORT", Points: 2, MaxPoints: 2, At: at},
		{EmployeeID: "x", Rule: "ACCOUNTANT", Points: 2, MaxPoints: 2, At: at},
	}
	rows := BuildScores(people, pts, nil, map[string]repository.ReliabilityFact{})
	work := map[string]bool{}
	for _, r := range rows {
		work[r.ID] = r.RoleWork
	}
	if work["d"] || !work["x"] {
		t.Fatalf("المصممة بلا شغل دور محسوب، والمحاسب عنده: %+v", work)
	}
	ins := strings.Join(boardInsights(rows), "\n")
	if !strings.Contains(ins, "⚠️ تقييم مو كامل: مصممة") || strings.Contains(ins, "محاسب (") {
		t.Fatalf("التنبيه لازم يذكر المصممة بس: %s", ins)
	}
	pts = append(pts, repository.MatrixScoreRow{EmployeeID: "d", Rule: "DESIGN_SECTION", Points: 2, MaxPoints: 2, At: at})
	for _, r := range BuildScores(people, pts, nil, map[string]repository.ReliabilityFact{}) {
		if r.ID == "d" && !r.RoleWork {
			t.Fatal("أقسام التصاميم = شغل دور")
		}
	}
}
