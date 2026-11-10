package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// App تطبيق مستقل (مثل جانغو): يعرّف أفعاله وقواعدها، وبعدين مساراته.
type App interface {
	Name() string
	Policies() map[Action]Rule
	Routes(r *Router)
}

// RouteInfo مسار مسجّل — للمصفوفة والفحص.
type RouteInfo struct {
	App     string `json:"app"`
	Pattern string `json:"pattern"`
	Action  Action `json:"action"`
}

// Router يسجّل مسارات التطبيقات على نفس الـmux الحالي (فالقديم والجديد يعيشون
// سوه). كل مسار لازم يذكر فعلاً مسجّلاً، وإلا السيرفر ما يشتغل.
type Router struct {
	mux      *http.ServeMux
	auth     func(http.Handler) http.Handler
	perms    PermLister
	policies *Policies
	app      string
	routes   []RouteInfo
	onDeny   func(r *http.Request, employeeID string)
	leader   func(employeeID string) bool
	grants   func(employeeID string) map[string]bool
}

// SetOwnerGrants يربط «صلاحيات المالك الي نطاها لمدير النظام» (جدول OwnerGrant).
func (rt *Router) SetOwnerGrants(f func(employeeID string) map[string]bool) { rt.grants = f }

// Policies القواعد المسجّلة (لشاشة الإنطاء).
func (rt *Router) Policies() *Policies { return rt.policies }

// SetLeader يربط فحص «ليدر هسه؟» (EmployeeRepository.IsLeaderFreshByID).
func (rt *Router) SetLeader(f func(employeeID string) bool) { rt.leader = f }

// OnDeny يربط تسجيل المخالفات (نفس حراس النظام القديمة: العدّ، الحظر على
// الكتابات المتكررة، تنبيه الإدارة) — بلاه النقل للنواة يخسّرنا هالحماية.
func (rt *Router) OnDeny(f func(r *http.Request, employeeID string)) { rt.onDeny = f }

func NewRouter(mux *http.ServeMux, auth func(http.Handler) http.Handler, perms PermLister, p *Policies) *Router {
	return &Router{mux: mux, auth: auth, perms: perms, policies: p}
}

// Mount يسجّل التطبيقات: قواعد الكل أولاً، بعدين المسارات.
func (rt *Router) Mount(apps ...App) {
	seen := map[string]bool{}
	for _, a := range apps {
		if seen[a.Name()] {
			panic(fmt.Sprintf("core: التطبيق %q مسجّل مرتين", a.Name()))
		}
		seen[a.Name()] = true
		for act, rule := range a.Policies() {
			if !strings.HasPrefix(string(act), a.Name()+".") {
				panic(fmt.Sprintf("core: الفعل %q لازم يبدي باسم تطبيقه %q", act, a.Name()))
			}
			rt.policies.Register(act, rule)
		}
	}
	for _, a := range apps {
		rt.app = a.Name()
		a.Routes(rt)
	}
	rt.app = ""
}

// Handle مسار بفعل. الرفض دائماً 404 (قاعدة دازاي): الي ما مسموحله ما يعرف إن الشي موجود.
//
// mws طبقات بعد القرار (مثل منع التكرار idem) — بنفس ترتيب السلسلة القديمة:
// الدخول ← القاعدة ← الطبقات ← المعالج.
func (rt *Router) Handle(pattern string, act Action, h http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	rule, ok := rt.policies.Rule(act)
	if !ok {
		panic(fmt.Sprintf("core: المسار %q يستعمل فعلاً مو مسجّل %q", pattern, act))
	}
	if rt.app != "" && !strings.HasPrefix(string(act), rt.app+".") {
		panic(fmt.Sprintf("core: التطبيق %q يستعمل فعل تطبيق ثاني %q بـ%q", rt.app, act, pattern))
	}
	var inner http.Handler = h
	for i := len(mws) - 1; i >= 0; i-- {
		inner = mws[i](inner)
	}
	gate := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := resolveActor(r, rt.perms, rt.leader, rt.grants)
		if !rt.policies.Allowed(a, act, r) {
			// المالك-بس ما ينحسب مخالفة (مثل RequireOwner القديم): المدير يفتح صفحة المالك بالغلط مو اعتداء
			if rt.onDeny != nil && a != nil && !rule.OwnerOnly {
				rt.onDeny(r, a.ID)
			}
			NotFound(w)
			return
		}
		inner.ServeHTTP(w, withActor(r, a))
	})
	var final http.Handler = gate
	if !rule.Public {
		final = rt.auth(gate)
	}
	rt.mux.Handle(pattern, final)
	rt.routes = append(rt.routes, RouteInfo{App: rt.app, Pattern: pattern, Action: act})
}

// Routes كل المسارات المسجّلة عبر النواة.
func (rt *Router) Routes() []RouteInfo {
	out := append([]RouteInfo(nil), rt.routes...)
	sort.Slice(out, func(i, j int) bool { return out[i].Pattern < out[j].Pattern })
	return out
}

// NotFound رد الرفض الموحّد.
func NotFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": "مو موجود"})
}
