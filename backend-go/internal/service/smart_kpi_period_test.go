package service

import (
	"testing"
	"time"
)

// (11-28) صفحة الـKPI على شهر الشركة: «2026-12» = من 26/11 لحد 25/12.
func TestSmartKpiPeriod(t *testing.T) {
	start, end, _, _ := getMonthRange("2026-12")
	if start.Format("2006-01-02") != "2026-11-26" || end.Format("2006-01-02") != "2026-12-26" {
		t.Fatalf("%v %v", start, end)
	}
	if n := getWorkingDays(start, end); n != 30 {
		t.Fatalf("أيام الفترة %d", n)
	}
	if n := getWorkingDaysSoFar(start, end, time.Date(2026, 11, 28, 15, 0, 0, 0, time.UTC)); n != 3 {
		t.Fatalf("لحد ٢٨: %d", n)
	}
	if n := getWorkingDaysSoFar(start, end, time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC)); n != 30 {
		t.Fatalf("بعد الفترة: %d", n)
	}
}
