// Package requests — طلبات مفتوحة لأي موظف عن قصد (يطلب، يسأل، يرفع) على «ون بيس».
// الطلب ينسجل باسم صاحبه من الجلسة.
package requests

import (
	"fmt"
	"net/http"

	"staffmange-api/internal/core"
)

const suffix = " — مفتوح لأي موظف عن قصد، باسمه"

var labels = map[core.Action]string{
	"requests.assistant_ask":   "يسأل المساعد الذكي",
	"requests.file_upload":     "يرفع ملف (العرض برابط موقّع)",
	"requests.letter":          "يطلب كتاب رسمي باسمه",
	"requests.manager_request": "يرفع طلب للمدير",
	"requests.meeting":         "يطلب اجتماع (حد للطلبات المفتوحة)",
	"requests.procurement":     "يرفع طلب شراء (نوع الطلب ينفحص بالمعالج)",
	"requests.tech_supplier":   "يضيف مورد تقني (المدير يختار)",
	"requests.inventory_check": "يسجّل جرد أغراضه",
}

type route struct {
	pattern string
	act     core.Action
	h       http.HandlerFunc
}

// App المعالجات تنحط بمكانها بـmain (Set)، والتسجيل مرة وحدة بالأخير.
type App struct{ routes []route }

func New() *App { return &App{} }

func (a *App) Set(pattern string, act core.Action, h http.HandlerFunc) {
	if _, ok := labels[act]; !ok {
		panic(fmt.Sprintf("requests: فعل مو معرّف %q", act))
	}
	a.routes = append(a.routes, route{pattern, act, h})
}

func (*App) Name() string { return "requests" }

func (*App) Policies() map[core.Action]core.Rule {
	out := map[core.Action]core.Rule{}
	for act, l := range labels {
		out[act] = core.Rule{Label: l + suffix, Anyone: true}
	}
	return out
}

func (a *App) Routes(r *core.Router) {
	for _, rt := range a.routes {
		r.Handle(rt.pattern, rt.act, rt.h)
	}
}
