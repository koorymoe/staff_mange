package core

import (
	"fmt"
	"net/http"
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
	}
}

func NewSetApp(name string, rules map[Action]Rule) *SetApp {
	return &SetApp{name: name, rules: rules}
}

func (a *SetApp) Set(pattern string, act Action, h http.HandlerFunc) {
	if _, ok := a.rules[act]; !ok {
		panic(fmt.Sprintf("%s: فعل مو معرّف %q", a.name, act))
	}
	a.routes = append(a.routes, struct {
		pattern string
		act     Action
		h       http.HandlerFunc
	}{pattern, act, h})
}

func (a *SetApp) Name() string              { return a.name }
func (a *SetApp) Policies() map[Action]Rule { return a.rules }
func (a *SetApp) Routes(r *Router) {
	for _, rt := range a.routes {
		r.Handle(rt.pattern, rt.act, rt.h)
	}
}

// AdminRule «مدير النظام» — نفس RequireRole("ADMIN") القديم: المالك والمدير ومن
// منصبه يعطيه دور ADMIN.
func AdminRule(label string) Rule { return Rule{Label: label, Roles: []string{"ADMIN"}} }
