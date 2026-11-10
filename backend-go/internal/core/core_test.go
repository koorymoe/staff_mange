package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakePerms map[string][]string

func (f fakePerms) Names(id string) ([]string, error) { return f[id], nil }

type testApp struct{}

func (testApp) Name() string { return "demo" }
func (testApp) Policies() map[Action]Rule {
	return map[Action]Rule{
		"demo.view":   {Label: "يشوف", Anyone: true},
		"demo.edit":   {Label: "يعدّل", Roles: []string{"HR_COORDINATOR"}, Perms: []string{"demo_edit"}},
		"demo.secret": {Label: "سر المالك", OwnerOnly: true},
		"demo.strict": {Label: "صارم", Strict: true, Roles: []string{"OWNER"}, Perms: []string{"demo_create"}},
		"demo.own":    {Label: "شغله هو", Check: func(a *Actor, r *http.Request) bool { return r.PathValue("id") == a.ID }},
	}
}
func (testApp) Routes(r *Router) {
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }
	r.Handle("GET /demo", "demo.view", ok)
	r.Handle("PUT /demo", "demo.edit", ok)
	r.Handle("GET /demo/secret", "demo.secret", ok)
	r.Handle("GET /demo/own/{id}", "demo.own", ok)
	r.Handle("POST /demo/strict", "demo.strict", ok)
}

// fakeAuth يقلّد RequireAuth: الهوية من ترويسة X-Emp/X-Role.
func fakeAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Emp") == "" {
			w.WriteHeader(401)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setup(t *testing.T) (*Router, *http.ServeMux) {
	t.Helper()
	mux := http.NewServeMux()
	rt := NewRouter(mux, fakeAuth, fakePerms{"e2": {"demo_edit"}}, NewPolicies())
	rt.Mount(testApp{})
	return rt, mux
}

func TestPolicyDecisions(t *testing.T) {
	rt, _ := setup(t)
	p := rt.policies
	owner := &Actor{ID: "o", Role: "OWNER"}
	admin := &Actor{ID: "a", Role: "ADMIN"}
	tech := &Actor{ID: "t", Role: "TECHNICIAN"}
	coord := &Actor{ID: "c", Role: "TECHNICIAN", OrgRoles: []string{"HR_COORDINATOR"}}
	perm := &Actor{ID: "e2", Role: "TECHNICIAN", Perms: map[string]bool{"demo_edit": true}}
	cases := []struct {
		a    *Actor
		act  Action
		want bool
	}{
		{nil, "demo.view", false},
		{tech, "demo.view", true},
		{tech, "demo.edit", false},
		{coord, "demo.edit", true}, // الدور من المنصب
		{perm, "demo.edit", true},  // الصلاحية
		{admin, "demo.edit", true}, // المدير يعبر
		{admin, "demo.secret", false},
		{owner, "demo.secret", true},
		{tech, "demo.nope", false}, // فعل مو مسجّل = لا
		{admin, "demo.strict", false}, // Strict: المدير بلا الصلاحية ما يعبر
		{owner, "demo.strict", true},
		{&Actor{ID: "x", Role: "ADMIN", Perms: map[string]bool{"demo_create": true}}, "demo.strict", true},
	}
	for _, c := range cases {
		if got := p.Simulate(c.a, c.act); got != c.want {
			t.Errorf("%+v %s = %v، المطلوب %v", c.a, c.act, got, c.want)
		}
	}
}

func TestRouterDeniesWith404(t *testing.T) {
	_, mux := setup(t)
	do := func(method, path, emp string) int {
		req := httptest.NewRequest(method, path, nil)
		if emp != "" {
			req.Header.Set("X-Emp", emp)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w.Code
	}
	// ⚠️ fakeAuth ما يحط سياق، فالـActor nil — يكفي لفحص إن الرفض 404 مو 403
	if c := do("PUT", "/demo", "e9"); c != 404 {
		t.Fatalf("الرفض لازم 404، طلع %d", c)
	}
	if c := do("GET", "/demo", ""); c != 401 {
		t.Fatalf("بلا دخول: %d", c)
	}
}

func mustPanic(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), want) {
			t.Fatalf("المتوقع panic بيه %q، طلع %v", want, r)
		}
	}()
	f()
}

type badApp struct {
	pol    map[Action]Rule
	routes func(r *Router)
}

func (badApp) Name() string                { return "bad" }
func (b badApp) Policies() map[Action]Rule { return b.pol }
func (b badApp) Routes(r *Router)          { b.routes(r) }

func TestStartupRefusesUnsafeApps(t *testing.T) {
	noop := func(http.ResponseWriter, *http.Request) {}
	newRT := func() *Router { return NewRouter(http.NewServeMux(), fakeAuth, nil, NewPolicies()) }
	mustPanic(t, "مو مسجّل", func() {
		newRT().Mount(badApp{pol: map[Action]Rule{}, routes: func(r *Router) { r.Handle("GET /x", "bad.x", noop) }})
	})
	mustPanic(t, "بلا قاعدة", func() {
		newRT().Mount(badApp{pol: map[Action]Rule{"bad.x": {Label: "س"}}, routes: func(*Router) {}})
	})
	mustPanic(t, "بلا وصف", func() {
		newRT().Mount(badApp{pol: map[Action]Rule{"bad.x": {Anyone: true}}, routes: func(*Router) {}})
	})
	mustPanic(t, "يبدي باسم", func() {
		newRT().Mount(badApp{pol: map[Action]Rule{"other.x": {Label: "س", Anyone: true}}, routes: func(*Router) {}})
	})
}

func TestMatrixFlagsUnusedActions(t *testing.T) {
	rt, _ := setup(t)
	m := rt.Matrix()
	if len(m) != 5 {
		t.Fatalf("الأفعال: %d", len(m))
	}
	for _, row := range m {
		if row.Unused {
			t.Errorf("%s بلا مسار", row.Action)
		}
	}
}
