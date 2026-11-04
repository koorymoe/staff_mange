package authz

import "testing"

// المرحلة ١: نفس السلوك القديم بالضبط — جدول لكل دور (مالك، أدمن، مراقب، محاسب،
// مدير مشاريع، كوادر، فني، مبيعات) ولكل نوع بيانات، منقول من الـ handlers القديمة.
func TestSeeAllMatchesOldRules(t *testing.T) {
	type c struct {
		res      Resource
		eff, raw string
		perms    []string
		want     bool
	}
	cases := []c{
		// الحجوزات (canSeeAllBookings)
		{Bookings, "OWNER", "OWNER", nil, true},
		{Bookings, "MONITOR", "SALES", nil, true}, // مراقب من منصبه
		{Bookings, "FINANCE", "FINANCE", nil, true},
		{Bookings, "PROCUREMENT_ADMIN", "TECHNICIAN", nil, true},
		{Bookings, "TECHNICIAN", "TECHNICIAN", nil, false},
		{Bookings, "SALES", "SALES", []string{"sales_booking"}, false}, // تسجيل مو قراءة
		{Bookings, "SALES", "SALES", []string{"view_bookings"}, true},
		{Bookings, "DESIGNER", "DESIGNER", []string{"finance"}, true},
		// فواتير الليدرية (canReviewInvoices — دور بس)
		{LeaderInvoices, "FINANCE", "FINANCE", nil, true},
		{LeaderInvoices, "HR_COORDINATOR", "HR_COORDINATOR", nil, false},
		{LeaderInvoices, "TECHNICIAN", "TECHNICIAN", []string{"finance"}, false},
		// المصاريف
		{Expenses, "MONITOR", "MONITOR", nil, true},
		{Expenses, "SALES", "SALES", []string{"expenses_manage"}, true},
		{Expenses, "PROJECT_MANAGER", "PROJECT_MANAGER", nil, false},
		// تقارير العمل
		{WorkReports, "SERVICE_MANAGER", "SERVICE_MANAGER", nil, true},
		{WorkReports, "FINANCE", "FINANCE", nil, false},
		{WorkReports, "QUALITY_ENGINEER", "QUALITY_ENGINEER", []string{"quality_control"}, true},
		// تقييم الموظفين (الدور الأصلي)
		{StaffScores, "ADMIN", "OWNER", nil, true},
		{StaffScores, "MONITOR", "MONITOR", nil, false},
		{StaffScores, "ADMIN", "SALES", nil, false}, // دور فعلي ما يكفي — الأصلي بس
		// نوع مو معرّف يضيق
		{Resource("x"), "OWNER", "OWNER", nil, false},
	}
	for _, x := range cases {
		if got := Decide(x.res, x.eff, x.raw, x.perms); got != x.want {
			t.Errorf("%s eff=%s raw=%s perms=%v: got %v want %v", x.res, x.eff, x.raw, x.perms, got, x.want)
		}
	}
}
