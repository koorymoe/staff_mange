package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// (دازاي 11-16) تغيّر دوره بنص الشهر: فترتين — القديم لحد التغيير، والحالي من بعده.
func TestRoleViews(t *testing.T) {
	from := time.Date(2026, 10, 31, 21, 0, 0, 0, time.UTC)
	to := time.Date(2026, 11, 30, 21, 0, 0, 0, time.UTC)
	at := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	prev, yes := "HR_COORDINATOR", false
	p := repository.Scorable{ID: "x", Role: "TECHNICIAN", IsLeader: true, RoleChangedAt: &at, PrevRole: &prev, PrevIsLeader: &yes}

	v := RoleViews(p, from, to)
	if len(v) != 2 || v[0].Role != "HR_COORDINATOR" || v[0].IsLeader || !v[0].To.Equal(at) ||
		v[1].Role != "TECHNICIAN" || !v[1].IsLeader || !v[1].From.Equal(at) || !v[1].To.Equal(to) {
		t.Fatalf("الفترات غلط: %+v", v)
	}
	// الشهر الجاي: التغيير قبله — تقييم عادي
	if v := RoleViews(p, to, to.AddDate(0, 1, 0)); len(v) != 0 {
		t.Fatalf("الشهر الجاي لازم بلا فترات: %+v", v)
	}
	// بلا تغيير
	if v := RoleViews(repository.Scorable{ID: "y", Role: "SALES"}, from, to); len(v) != 0 {
		t.Fatalf("بلا تغيير لازم فاضي: %+v", v)
	}
}
