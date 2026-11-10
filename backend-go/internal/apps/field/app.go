// Package field — شغل الميدان والمشاريع على «ون بيس». المسار لأي موظف داخل،
// والفحص على الحجز/المشروع نفسه بالمعالج (مكتوب بوصف كل فعل) — الخطوة الجاية
// تنقل هالفحوص لـCheck بالنواة.
package field

import (
	"fmt"
	"net/http"

	"staffmange-api/internal/core"
)

const suffix = " — الفحص على الشي نفسه بالمعالج"

var labels = map[core.Action]string{
	"field.booking_create":         "يسجّل حجز (الحجز الداخلي بصلاحية — authz.InternalBookingCreate)",
	"field.booking_tech_step":      "خطوة الفني على حجزه (الخدمة تفحص إنه الفني المحوّل إله)",
	"field.customer_signature":     "ياخذ توقيع الزبون (كادر الحجز بس — CanCollect)",
	"field.survey_report":          "يرفع تقرير كشف (يفحص إنه مكلّف)",
	"field.survey_photo_upload":    "يرفع صور كشف (يفحص الحجز)",
	"field.survey_photo_delete":    "يحذف صورة كشف (الي رفعها أو المدير)",
	"field.work_report":            "يرفع تقرير عمل (يفحص إنه بالكادر)",
	"field.crew_rating":            "يقيّم كادر حجز اشتغل بيه",
	"field.after_check":            "جرد بعد الحجز (يفحص إنه بالكادر)",
	"field.leave_preliminary":      "رأي أولي بإجازة (المسؤول المباشر بس)",
	"field.project_contract_value": "قيمة عقد المشروع (authz.ProjectContractValue أو مشرفه)",
	"field.project_delay_reason":   "سبب تأخير مشروع (مشرفه أو صلاحية إدارة المشاريع)",
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
		panic(fmt.Sprintf("field: فعل مو معرّف %q", act))
	}
	a.routes = append(a.routes, route{pattern, act, h})
}

func (*App) Name() string { return "field" }

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
