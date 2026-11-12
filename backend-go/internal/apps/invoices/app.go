// Package invoices — فواتير الليدرية والخدمات ودفعات المشاريع على «ون بيس».
// كل حارس قديم صار قاعدة بنفس المعنى، ومنع التكرار (idem) باقي على كل كتابة فلوس.
package invoices

import (
	"net/http"

	"staffmange-api/internal/core"
)

type Deps struct {
	// IsBookingManager مسؤول خدمة الحجز (PaperworkGuard.IsBookingManager)
	IsBookingManager func(employeeID, bookingID string) (bool, error)
}

// basketPerms سلة الليدر — نفس requireLeaderBasket القديم
var basketPerms = []string{"leader_basket", "finance", "finance_audit", "invoice_gps", "invoice_dashcam", "invoice_internal", "leader_invoices_view", "monitoring"}

func New(d Deps) *core.SetApp {
	basket := func(l string) core.Rule { return core.LeaderPermRule(l, basketPerms...) }
	finance := func(l string) core.Rule { return core.RolePermRule(l, []string{"ADMIN", "FINANCE"}, "finance_audit") }
	monInbox := func(l string) core.Rule {
		return core.RolePermRule(l, []string{"ADMIN", "MONITOR"}, "monitoring", "auditing", "mon_inbox")
	}
	cost := func(l string) core.Rule { return core.LeaderPermRule(l, "execution_cost") }
	byHandler := func(l string) core.Rule {
		return core.Rule{Label: l + " — القرار بالمعالج (authz.ProjectPayments…)", Anyone: true}
	}
	return core.NewSetApp("invoices", map[core.Action]core.Rule{
		"invoices.list":      basket("يشوف فواتير الليدرية"),
		"invoices.page":      basket("يشوف صفحة فواتير"),
		"invoices.view":      basket("يشوف فاتورة"),
		"invoices.by_number": basket("يدوّر فاتورة برقمها"),
		"invoices.create": {Label: "يسجّل فاتورة حجز — الليدر أو مسؤول خدمة الحجز (bookingId من الطلب)", Check: func(a *core.Actor, r *http.Request) bool {
			if a.IsLeader() {
				return true
			}
			ok, err := d.IsBookingManager(a.ID, core.BodyField(r, "bookingId"))
			return err == nil && ok
		}},
		"invoices.create_service":       core.RolePermRule("يسجّل فاتورة GPS/داش كام", []string{"ADMIN", "OWNER"}, "invoice_gps", "invoice_dashcam"),
		"invoices.create_manual":        core.RolePermRule("يسجّل فاتورة بكلفة يدوية", []string{"ADMIN", "OWNER"}, "invoice_manual"),
		"invoices.create_internal":      core.RolePermRule("يسجّل فاتورة شغل داخلي", []string{"ADMIN", "HR_COORDINATOR"}, "coordinator", "invoice_internal"),
		"invoices.estimate":             cost("يحسب كلفة تقديرية (ويا شرط التسليم الناقص)"),
		"invoices.camera_cost":          cost("يحسب كلفة كاميرات"),
		"invoices.camera_options":       cost("يشوف خيارات كلفة الكاميرات"),
		"invoices.late":                 monInbox("يشوف الفواتير المتأخرة"),
		"invoices.monitor_decide":       monInbox("يقرّر على فاتورة بصندوق المراقب"),
		"invoices.approve":              finance("يعتمد فاتورة"),
		"invoices.audit":                finance("يدقق فاتورة"),
		"invoices.monitor_request":      finance("يحوّل فاتورة للمراقب"),
		"invoices.revoke":               finance("يسحب اعتماد فاتورة"),
		"invoices.delete_request":       finance("يطلب حذف فاتورة"),
		"invoices.delete_requests":      finance("يشوف طلبات حذف الفواتير"),
		"invoices.no_number":            finance("يشوف المعتمدة بلا رقم"),
		"invoices.set_number":           finance("يحط رقم الفاتورة الخارجي"),
		"invoices.adjust":               finance("يعدّل مبلغ فاتورة"),
		"invoices.adjustments":          core.RolePermRule("يشوف تعديلات فاتورة", []string{"ADMIN", "OWNER", "FINANCE", "MONITOR"}, "leader_invoices_view", "finance", "monitoring", "auditing"),
		"invoices.return":               {Label: "يرجّع فاتورة للمحاسب", OwnerOnly: true},
		"invoices.delete_decide":        {Label: "يقرّر على طلب حذف فاتورة", OwnerOnly: true},
		"invoices.project_payments":     byHandler("يشوف دفعات مشروع"),
		"invoices.project_payment_add":  byHandler("يسجّل دفعة مشروع"),
		"invoices.project_payment_ok":   byHandler("يأكد دفعة"),
		"invoices.project_payment_void": byHandler("يلغي دفعة"),
		"invoices.payments_overview":    byHandler("يشوف فلوس كل المشاريع"),
		"invoices.payments_audit":       byHandler("يشوف سجل الدفعات (المالك)"),
		"invoices.payment_amount":       byHandler("يصحّح مبلغ دفعة (المالك)"),
	})
}
