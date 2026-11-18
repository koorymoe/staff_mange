package service

import (
	"testing"
	"time"
)

func TestFleetDayRangeBaghdad(t *testing.T) {
	now := time.Date(2026, 11, 17, 22, 30, 0, 0, time.UTC) // 01:30 يوم 18 ببغداد
	from, to, err := fleetDayRange("", now)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 11, 17, 21, 0, 0, 0, time.UTC); !from.Equal(want) || !to.Equal(want.Add(24*time.Hour)) {
		t.Fatalf("got %v..%v", from, to)
	}
	if _, _, err := fleetDayRange("18-11", now); err == nil {
		t.Fatal("bad date accepted")
	}
}

func TestFleetLogsToolAccess(t *testing.T) {
	f := MatrixToolFilter("SALES", []string{"agent_matrix"})
	if f("fleet_logs_day") {
		t.Fatal("sales without vehicle perms got fleet tool")
	}
	if !MatrixToolFilter("QUANTITY_ADMIN", []string{"agent_matrix", "inventory"})("fleet_logs_day") {
		t.Fatal("inventory user denied")
	}
}
