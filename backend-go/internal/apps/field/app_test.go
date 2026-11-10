package field

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"staffmange-api/internal/core"
)

// (11-10) الي مو طرف ياخذ 404 من النواة قبل ما يوصل المعالج.
func TestFieldChecks(t *testing.T) {
	app := New(Deps{
		BookingParty: func(b, e string) bool { return b == "b1" && e == "crew" },
		ProjectParty: func(p, e string) bool { return p == "p1" && e == "sup" },
	})
	reached := map[string]bool{}
	ok := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) { reached[name] = true }
	}
	app.Set("PUT /api/bookings/{id}/tech/{step}", "field.booking_tech_step", ok("tech"))
	app.Set("POST /api/bookings/{id}/customer-signature", "field.customer_signature", ok("sig"))
	app.Set("POST /api/projects/{id}/delay-reason", "field.project_delay_reason", ok("delay"))

	pol := core.NewPolicies()
	for act, r := range app.Policies() {
		pol.Register(act, r)
	}
	req := func(method, path string, a *core.Actor, act core.Action) bool {
		mux := http.NewServeMux()
		var got bool
		mux.HandleFunc(method+" "+map[core.Action]string{
			"field.booking_tech_step":    "/api/bookings/{id}/tech/{step}",
			"field.customer_signature":   "/api/bookings/{id}/customer-signature",
			"field.project_delay_reason": "/api/projects/{id}/delay-reason",
		}[act], func(w http.ResponseWriter, r *http.Request) { got = pol.Allowed(a, act, r) })
		mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, path, nil))
		return got
	}
	crew := &core.Actor{ID: "crew", Role: "TECHNICIAN", Perms: map[string]bool{}}
	other := &core.Actor{ID: "x", Role: "TECHNICIAN", Perms: map[string]bool{}}
	pm := &core.Actor{ID: "pm", Role: "TECHNICIAN", Perms: map[string]bool{"project_management": true}}
	sup := &core.Actor{ID: "sup", Role: "TECHNICIAN", Perms: map[string]bool{}}
	admin := &core.Actor{ID: "a", Role: "ADMIN", Perms: map[string]bool{}}
	cases := []struct {
		name         string
		method, path string
		a            *core.Actor
		act          core.Action
		want         bool
	}{
		{"الكادر يسوي خطوة", "PUT", "/api/bookings/b1/tech/visited", crew, "field.booking_tech_step", true},
		{"غريب ما يسوي خطوة", "PUT", "/api/bookings/b1/tech/visited", other, "field.booking_tech_step", false},
		{"الكادر بحجز ثاني", "PUT", "/api/bookings/b2/tech/visited", crew, "field.booking_tech_step", false},
		{"توقيع الكادر", "POST", "/api/bookings/b1/customer-signature", crew, "field.customer_signature", true},
		{"توقيع غريب", "POST", "/api/bookings/b1/customer-signature", other, "field.customer_signature", false},
		{"مشرف المشروع", "POST", "/api/projects/p1/delay-reason", sup, "field.project_delay_reason", true},
		{"صلاحية إدارة المشاريع", "POST", "/api/projects/p9/delay-reason", pm, "field.project_delay_reason", true},
		{"غريب عن المشروع", "POST", "/api/projects/p1/delay-reason", other, "field.project_delay_reason", false},
		{"المدير يعبر (المعالج يفحص بعدين)", "POST", "/api/projects/p1/delay-reason", admin, "field.project_delay_reason", true},
	}
	for _, c := range cases {
		if got := req(c.method, c.path, c.a, c.act); got != c.want {
			t.Errorf("%s: %v، المطلوب %v", c.name, got, c.want)
		}
	}
}

func TestFieldBodyAndPhotoChecks(t *testing.T) {
	app := New(Deps{
		BookingParty:        func(string, string) bool { return false },
		ProjectParty:        func(string, string) bool { return false },
		BookingLeader:       func(b, e string) bool { return b == "b1" && e == "lead" },
		BookingTech:         func(b, e string) bool { return b == "b1" && e == "tech" },
		SurveyPhotoUploader: func(id string) string { return map[string]string{"ph1": "tech"}[id] },
	})
	pol := core.NewPolicies()
	for act, r := range app.Policies() {
		pol.Register(act, r)
	}
	body := func(s string) *http.Request {
		return httptest.NewRequest("POST", "/x", strings.NewReader(s))
	}
	lead := &core.Actor{ID: "lead", Role: "TECHNICIAN", Perms: map[string]bool{}}
	tech := &core.Actor{ID: "tech", Role: "TECHNICIAN", Perms: map[string]bool{}}
	if !pol.Allowed(lead, "field.crew_rating", body(`{"bookingId":"b1"}`)) || pol.Allowed(tech, "field.crew_rating", body(`{"bookingId":"b1"}`)) {
		t.Error("تقييم الكادر لليدر بس")
	}
	if pol.Allowed(lead, "field.crew_rating", body(`{"bookingId":"b2"}`)) || pol.Allowed(lead, "field.crew_rating", body(`bad`)) {
		t.Error("ليدر بحجز ثاني أو جسم خربان")
	}
	if !pol.Allowed(tech, "field.after_check", body(`{"bookingId":"b1"}`)) || pol.Allowed(lead, "field.after_check", body(`{"bookingId":"b1"}`)) {
		t.Error("الجرد بعد الحجز للفني المكلّف")
	}
	del := func(id string) *http.Request {
		mux := http.NewServeMux()
		var got *http.Request
		mux.HandleFunc("DELETE /api/survey-photos/{id}", func(_ http.ResponseWriter, r *http.Request) { got = r })
		mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("DELETE", "/api/survey-photos/"+id, nil))
		return got
	}
	if !pol.Allowed(tech, "field.survey_photo_delete", del("ph1")) || pol.Allowed(lead, "field.survey_photo_delete", del("ph1")) || pol.Allowed(tech, "field.survey_photo_delete", del("nope")) {
		t.Error("حذف صورة الكشف للي رفعها بس")
	}
}
