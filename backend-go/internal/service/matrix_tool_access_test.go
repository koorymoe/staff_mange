package service

import "testing"

// (دازاي 10-29) الموظف ياخذ أدوات ماتركس بحدود صلاحياته بس.
func TestMatrixToolFilter(t *testing.T) {
	if MatrixToolFilter("OWNER", nil) != nil || MatrixToolFilter("ADMIN", nil) != nil {
		t.Fatal("المالك والمدير ياخذون كل الأدوات")
	}
	fin := MatrixToolFilter("ACCOUNTANT", []string{"agent_matrix", "finance"})
	for name, want := range map[string]bool{
		"money_month": true, "revenue_by_responsible": true, "system_guide": true, "my_stats": true,
		"staff_scores": false, "employee_signals": false, "query_data": false, "describe_data": false, "unknown_tool": false,
	} {
		if got := fin(name); got != want {
			t.Errorf("%s: got %v want %v", name, got, want)
		}
	}
	plain := MatrixToolFilter("TECHNICIAN", []string{"agent_matrix"})
	if plain("money_month") || plain("late_bookings") || !plain("booking_check") {
		t.Error("الموظف بلا صلاحيات مجال ياخذ العامة بس")
	}
}
