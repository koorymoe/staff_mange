package service

import (
	"regexp"
	"testing"

	"staffmange-api/internal/repository"
)

// (ع) 10-16: «المدير رجال بسيط» — ولا سطر بالنشاط يطلع بيه إنكليزي.
func TestActivityDescribeArabicOnly(t *testing.T) {
	s := &EmployeeActivityService{}
	latin := regexp.MustCompile(`[a-zA-Z]`)
	for _, c := range []struct{ m, p string }{
		{"POST", "/api/duplicate-candidates"},
		{"PUT", "/api/duplicate-candidates/{id}/dismiss"},
		{"POST", "/api/staff-ratings"},
		{"POST", "/api/justifications"},
		{"PUT", "/api/bookings/{id}/progress/review"},
		{"PUT", "/api/work-reports/{id}/review"},
		{"POST", "/api/some-new-thing"},
	} {
		text, area := s.describe(repository.ActivityRow{Method: c.m, Pattern: c.m + " " + c.p, Path: "/api/x"}) // بلا معرّف حتى ما يدوّر الاسم
		if latin.MatchString(text) || latin.MatchString(area) {
			t.Errorf("%s %s → %q / %q فيه إنكليزي", c.m, c.p, text, area)
		}
		t.Logf("%s %s → %s (%s)", c.m, c.p, text, area)
	}
}
