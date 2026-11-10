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

// Deps فحوص «الشي نفسه» — تعريف واحد يستعمله كل فعل يحتاجه.
type Deps struct {
	// BookingParty الموظف طرف بالحجز (repository.BookingPartySQL — نفس تعريف القوائم)
	BookingParty func(bookingID, employeeID string) bool
	// ProjectParty مسؤول/مفوّض/كاشف/منشئ المشروع أو مشرف حجزه
	ProjectParty func(projectID, employeeID string) bool
	// BookingLeader ليدر الحجز، BookingTech فني مكلّف بيه (matrix_chain_repository)
	BookingLeader func(bookingID, employeeID string) bool
	BookingTech   func(bookingID, employeeID string) bool
	// SurveyPhotoUploader منو رفع صورة الكشف ("" إذا مو موجودة)
	SurveyPhotoUploader func(photoID string) string
	// قواعد معقدة تبقى بمكانها بالمعالج — النواة تناديها (تعريف واحد، مو نسختين)
	CanCreateInternal func(r *http.Request) bool
	CanLeaveDecide    func(r *http.Request, leaveID string) bool
	CanContractValue  func(r *http.Request) bool
}

// App المعالجات تنحط بمكانها بـmain (Set)، والتسجيل مرة وحدة بالأخير.
type App struct {
	routes []route
	d      Deps
}

// Use يكمّل فحص يتعرّف بعدين بـmain (معالجه ينخلق بعد التطبيق).
func (a *App) Use(f func(d *Deps)) { f(&a.d) }

func New(d Deps) *App { return &App{d: d} }

func (a *App) Set(pattern string, act core.Action, h http.HandlerFunc) {
	if _, ok := labels[act]; !ok {
		panic(fmt.Sprintf("field: فعل مو معرّف %q", act))
	}
	a.routes = append(a.routes, route{pattern, act, h})
}

func (*App) Name() string { return "field" }

// (11-10) فحوص انتقلت من المعالج للنواة — الي ما إله علاقة ياخذ 404 قبل ما يوصل.
// المعالج يبقى يفحص (دفاع ثاني)، وأدق: التوقيع مثلاً للكادر بس.
func (a *App) checks() map[core.Action]core.Rule {
	party := func(x *core.Actor, r *http.Request) bool { return a.d.BookingParty(r.PathValue("id"), x.ID) }
	return map[core.Action]core.Rule{
		"field.booking_tech_step":  {Label: "خطوة الفني على حجز هو طرف بيه", Check: party},
		"field.customer_signature": {Label: "ياخذ توقيع الزبون على حجز هو طرف بيه (والمعالج: الكادر بس)", Check: party},
		"field.crew_rating": {Label: "يقيّم فنيّي حجز هو ليدره (bookingId من الطلب)", Check: func(x *core.Actor, r *http.Request) bool {
			return a.d.BookingLeader(core.BodyField(r, "bookingId"), x.ID)
		}},
		"field.after_check": {Label: "جرد بعد حجز هو مكلّف بيه (bookingId من الطلب)", Check: func(x *core.Actor, r *http.Request) bool {
			return a.d.BookingTech(core.BodyField(r, "bookingId"), x.ID)
		}},
		"field.survey_photo_delete": {Label: "يحذف صورة كشف هو رفعها (والمدير)", Check: func(x *core.Actor, r *http.Request) bool {
			up := a.d.SurveyPhotoUploader(r.PathValue("id"))
			return up != "" && up == x.ID
		}},
		"field.booking_create": {Label: "يسجّل حجز — الحجز الداخلي (bookingType=INTERNAL) بصلاحية الحجز الداخلي", Anyone: true, Check: func(_ *core.Actor, r *http.Request) bool {
			return core.BodyField(r, "bookingType") != "INTERNAL" || a.d.CanCreateInternal(r)
		}},
		"field.leave_preliminary": {Label: "رأي أولي بإجازة — بس المسؤول عن مسار الطلب", Check: func(_ *core.Actor, r *http.Request) bool {
			return a.d.CanLeaveDecide(r, r.PathValue("id"))
		}},
		"field.project_contract_value": {Label: "قيمة عقد المشروع — المالية/مدير المشاريع (authz.ProjectContractValue)", Check: func(_ *core.Actor, r *http.Request) bool {
			return a.d.CanContractValue(r)
		}},
		"field.project_delay_reason": {Label: "سبب تأخير مشروع — طرف بالمشروع أو صلاحية إدارة المشاريع", Check: func(x *core.Actor, r *http.Request) bool {
			return x.Can("project_management") || a.d.ProjectParty(r.PathValue("id"), x.ID)
		}},
	}
}

func (a *App) Policies() map[core.Action]core.Rule {
	out := map[core.Action]core.Rule{}
	checked := a.checks()
	for act, l := range labels {
		if c, ok := checked[act]; ok {
			out[act] = c
			continue
		}
		out[act] = core.Rule{Label: l + suffix, Anyone: true}
	}
	return out
}

func (a *App) Routes(r *core.Router) {
	// فحص ناقص = السيرفر ما يشتغل (أحسن من 500 بأول طلب)
	if a.d.BookingParty == nil || a.d.ProjectParty == nil || a.d.BookingLeader == nil || a.d.BookingTech == nil ||
		a.d.SurveyPhotoUploader == nil || a.d.CanCreateInternal == nil || a.d.CanLeaveDecide == nil || a.d.CanContractValue == nil {
		panic("field: فحص ناقص بـDeps")
	}
	for _, rt := range a.routes {
		r.Handle(rt.pattern, rt.act, rt.h)
	}
}
