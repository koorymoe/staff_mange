package main

import (
	"net/http"
	"testing"
)

// ═══ 🛡️ المسارات الحساسة (11-18) ═══
// فني عادي (بلا صلاحيات) يضرب مسارات تغيّر فلوس أو تقييم أو صلاحيات أو بلاغات:
// لازم ياخذ 404 (قاعدة النظام: الرفض = «مو موجود»)، مو 200 ولا 403 ولا 500.
func TestTechnicianBlockedFromSensitiveRoutes(t *testing.T) {
	srv, db := setupTestServer(t)
	_, username := createTestEmployee(t, db, "TECHNICIAN")
	victim, _ := createTestEmployee(t, db, "SALES")
	token := loginAndGetToken(t, srv.URL, username)

	cases := []struct {
		name, method, path string
		body               any
	}{
		{"تأكيد تقييم ليدر سيء", http.MethodPost, "/api/score-reviews/ratings/x", map[string]any{"confirm": true}},
		{"تحديد الغلطان بمشكلة", http.MethodPost, "/api/score-reviews/issues/x", map[string]any{"faultEmployeeId": victim}},
		{"مبرّر نقص العدّة", http.MethodPost, "/api/inventory/after-shortages/justify", map[string]any{"bookingId": "x", "employeeId": victim}},
		{"قائمة نواقص العدّة", http.MethodGet, "/api/inventory/after-shortages", nil},
		{"تقييم ماتركس بس", http.MethodPut, "/api/employees/" + victim + "/ai-only-rating", map[string]any{"aiOnly": true}},
		{"تغيير حالة بلاغ", http.MethodPatch, "/api/bot-reports/x", map[string]any{"status": "DONE"}},
		{"توضيح بلاغ مو إله", http.MethodPost, "/api/me/bot-reports/x/clarify", map[string]any{"text": "توضيح"}},
		{"تذكير المدير بطلب شراء", http.MethodPost, "/api/procurement/x/remind", nil},
		{"شرح التقييم لكل الموظفين", http.MethodGet, "/api/staff-score/explain", nil},
		{"تعديل سيارة", http.MethodPut, "/api/vehicles/x", map[string]any{"name": "x"}},
		{"حذف سيارة", http.MethodDelete, "/api/vehicles/x", nil},
		{"صوت كمال", http.MethodPost, "/api/secretary/speak", map[string]any{"text": "هلا"}},
	}
	for _, c := range cases {
		resp := doRequest(t, c.method, srv.URL+c.path, token, c.body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s (%s %s): نريد 404، طلع %d", c.name, c.method, c.path, resp.StatusCode)
		}
	}
	// الفني ما يگدر يرفع نفسه ولا غيره لمدير. (بعد هالكثر محاولات مرفوضة النظام يقفل جلسته
	// — authzViolations — فـ401 هم رفض صحيح.)
	resp := doRequest(t, http.MethodPut, srv.URL+"/api/employees/"+victim, token, map[string]any{"role": "ADMIN"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("فني غيّر دور موظف: %d", resp.StatusCode)
	}
	var role string
	db.Get(&role, `SELECT role::text FROM "Employee" WHERE id = $1`, victim)
	if role != "SALES" {
		t.Fatalf("الدور تغيّر: %s", role)
	}
}
