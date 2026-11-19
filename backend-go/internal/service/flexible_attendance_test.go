package service

import (
	"testing"

	"staffmange-api/internal/repository"
)

// (دازاي 11-19) الدوام المرن: «ما سجّل حضور» ما تنحسب غياب ولا تنزّل الاعتمادية.
func TestFlexibleNoAbsence(t *testing.T) {
	people := []repository.Scorable{{ID: "f", Role: "SALES", Flexible: true}, {ID: "n", Role: "SALES"}}
	pts := []repository.MatrixScoreRow{}
	for _, id := range []string{"f", "n"} {
		pts = append(pts,
			repository.MatrixScoreRow{EmployeeID: id, Rule: "ATTENDANCE", Points: 2, MaxPoints: 2},
			repository.MatrixScoreRow{EmployeeID: id, Rule: "ATTENDANCE", Points: 0, MaxPoints: 2})
	}
	rows := BuildScores(people, pts, nil, nil)
	pct := func(s StaffScore, key string) float64 {
		for _, p := range s.RelParts {
			if p.Key == key {
				return p.Pct
			}
		}
		return -1
	}
	if v := pct(rows[0], "ATTENDANCE"); v != 100 {
		t.Fatalf("المرن الحضور = %v", v)
	}
	if v := pct(rows[1], "ATTENDANCE"); v != 50 {
		t.Fatalf("العادي الحضور = %v", v)
	}
}
