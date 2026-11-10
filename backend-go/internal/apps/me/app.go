// Package me — «شغلي أنا» على «ون بيس»: أفعال الموظف على نفسه (حضوره، إشعاراته،
// إجازته، تقريره...). أي موظف داخل، والخدمة تشتغل على **صاحب الطلب بس**
// (معرّفه من الجلسة، مو من الطلب) — فما يگدر يسويها باسم غيره.
package me

import (
	"fmt"
	"net/http"

	"staffmange-api/internal/core"
)

const onSelf = " — على نفسه بس (المعرّف من الجلسة)"

var labels = map[core.Action]string{
	"me.telegram_link":           "يربط حسابه بالتلغرام",
	"me.bot_report":              "يرفع بلاغ من النظام",
	"me.profile_survey":          "يملي استمارة ملفه",
	"me.daily_report":            "يرفع تقريره اليومي",
	"me.rating_duty":             "يسلّم واجب التقييم اليومي",
	"me.courier_seen":            "يعلّم رسائل ماتركس إله مقروءة",
	"me.courier_delivered":       "يأكد وصول رسالة ماتركس إله",
	"me.push_subscribe":          "يفعّل إشعارات جهازه",
	"me.push_unsubscribe":        "يطفي إشعارات جهازه",
	"me.checkin":                 "يسجّل حضوره",
	"me.checkout":                "يسجّل انصرافه",
	"me.attendance_answer":       "يجاوب سؤال الحضور التلقائي إله",
	"me.notification_read":       "يعلّم إشعاره مقروء",
	"me.notifications_read_all":  "يعلّم كل إشعاراته مقروءة",
	"me.privacy_accept":          "يوافق على سياسة الخصوصية",
	"me.attendance_icon_request": "يطلب أيقونة حضور",
	"me.leave_request":           "يقدّم طلب إجازة إله",
	"me.leave_cancel":            "يلغي طلب إجازته",
	"me.peer_note":               "يكتب رأيه بزملائه",
	"me.achievement":             "يسجّل إنجاز إله",
	"me.achievement_voice":       "يرفع فويس لإنجازه",
}

// App المعالجات تنحط بمكانها بـmain (Set)، والتسجيل مرة وحدة بالأخير (Mount).
type App struct {
	routes []route
}

type route struct {
	pattern string
	act     core.Action
	h       http.HandlerFunc
}

func New() *App { return &App{} }

// Set يربط مسار بفعل معروف — فعل مو بالقائمة خطأ برمجي.
func (a *App) Set(pattern string, act core.Action, h http.HandlerFunc) {
	if _, ok := labels[act]; !ok {
		panic(fmt.Sprintf("me: فعل مو معرّف %q", act))
	}
	a.routes = append(a.routes, route{pattern, act, h})
}

func (*App) Name() string { return "me" }

func (*App) Policies() map[core.Action]core.Rule {
	out := map[core.Action]core.Rule{}
	for act, l := range labels {
		out[act] = core.Rule{Label: l + onSelf, Anyone: true}
	}
	return out
}

func (a *App) Routes(r *core.Router) {
	for _, rt := range a.routes {
		r.Handle(rt.pattern, rt.act, rt.h)
	}
}
