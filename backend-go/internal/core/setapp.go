package core

import (
	"fmt"
	"net/http"
	"strings"
)

// SetApp تطبيق بقواعد ثابتة، ومعالجاته تنربط بمكانها بـmain (Set) — لنقل قسم
// قديم بلا ما نغيّر وين تنخلق معالجاته. Set لفعل مو معرّف = خطأ برمجي فوري.
type SetApp struct {
	name   string
	rules  map[Action]Rule
	routes []struct {
		pattern string
		act     Action
		h       http.HandlerFunc
		mws     []func(http.Handler) http.Handler
	}
}

func NewSetApp(name string, rules map[Action]Rule) *SetApp {
	return &SetApp{name: name, rules: rules}
}

func (a *SetApp) Set(pattern string, act Action, h http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	if _, ok := a.rules[act]; !ok {
		panic(fmt.Sprintf("%s: فعل مو معرّف %q", a.name, act))
	}
	a.routes = append(a.routes, struct {
		pattern string
		act     Action
		h       http.HandlerFunc
		mws     []func(http.Handler) http.Handler
	}{pattern, act, h, mws})
}

func (a *SetApp) Name() string              { return a.name }
func (a *SetApp) Policies() map[Action]Rule { return a.rules }
func (a *SetApp) Routes(r *Router) {
	for _, rt := range a.routes {
		r.Handle(rt.pattern, rt.act, rt.h, rt.mws...)
	}
}

// AdminRule «مدير النظام» — نفس RequireRole("ADMIN") القديم: المالك والمدير ومن
// منصبه يعطيه دور ADMIN.
func AdminRule(label string) Rule { return Rule{Label: label, Roles: []string{"ADMIN"}} }

// PermRule «صلاحية» — نفس guards.Permission القديم: المالك والمدير يعبرون، أو الصلاحية.
func PermRule(label string, perms ...string) Rule { return Rule{Label: label, Perms: perms} }

// RolePermRule نفس guards.RoleOrAnyPermission القديم: الدور (الأصلي أو المنصب أو
// المراقب) أو وحدة من الصلاحيات — والمالك والمدير يعبرون.
func RolePermRule(label string, roles []string, perms ...string) Rule {
	return Rule{Label: label, Roles: roles, Perms: perms}
}

// LeaderPermRule نفس guards.LeaderOr(Any)Permission: الليدر أو الصلاحية، والمالك والمدير يعبرون.
func LeaderPermRule(label string, perms ...string) Rule {
	return Rule{Label: label, Leader: true, Perms: perms}
}

// ActionFor اسم فعل ثابت من المسار ("POST /api/x/{id}/y" ← app.post_x_by_id_y) — للمسارات
// الي تتسجّل بحلقة.
func ActionFor(app, pattern string) Action {
	method, path, _ := strings.Cut(pattern, " ")
	parts := []string{}
	for _, seg := range strings.Split(strings.TrimPrefix(path, "/api/"), "/") {
		if strings.HasPrefix(seg, "{") {
			seg = "by_" + strings.Trim(seg, "{}.")
		}
		parts = append(parts, strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' {
				return r
			}
			return '_'
		}, strings.ToLower(seg)))
	}
	return Action(app + "." + strings.ToLower(method) + "_" + strings.Join(parts, "_"))
}

// Add يعرّف فعل ويربط مساره بنفس الخطوة (للحلقات).
func (a *SetApp) Add(pattern string, rule Rule, h http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	act := ActionFor(a.name, pattern)
	if _, dup := a.rules[act]; dup {
		panic(fmt.Sprintf("%s: فعل مكرر %q", a.name, act))
	}
	a.rules[act] = rule
	a.Set(pattern, act, h, mws...)
}
