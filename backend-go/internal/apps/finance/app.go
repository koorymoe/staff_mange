// Package finance — الفلوس على «ون بيس»: الكريدت، المصاريف، الصندوق الدوّار.
//
// قواعد الكريدت والصندوق تنقرا من محرك الصلاحيات (authz.SeeAllRules) — مكان
// واحد، فأي تعديل هناك يوصل هنا. ومنع التكرار (idem) يبقى على كل كتابة فلوس.
package finance

import (
	"staffmange-api/internal/authz"
	"staffmange-api/internal/core"
)

// fromAuthz نفس authz.Guard: أدوار القاعدة أو صلاحياتها، والمالك والمدير يعبرون.
func fromAuthz(label string, res authz.Resource) core.Rule {
	r := authz.SeeAllRules[res]
	return core.RolePermRule(label, r.Roles, r.Perms...)
}

func New() *core.SetApp {
	own := func(l string) core.Rule {
		return core.Rule{Label: l + " — على نفسه/حسب صلاحيته بالمعالج", Anyone: true}
	}
	fin := []string{"ADMIN", "FINANCE"}
	return core.NewSetApp("finance", map[core.Action]core.Rule{
		"finance.credits":          fromAuthz("يشوف الكريدت (البيع بالآجل)", authz.Credits),
		"finance.credits_export":   fromAuthz("يصدّر الكريدت", authz.Credits),
		"finance.credit_add":       fromAuthz("يسجّل كريدت", authz.Credits),
		"finance.credit_payments":  fromAuthz("يشوف تسديدات كريدت", authz.Credits),
		"finance.credit_pay":       fromAuthz("يسجّل تسديد كريدت", authz.Credits),
		"finance.daily_audit":      fromAuthz("يشوف تدقيق اليوم المالي", authz.Credits),
		"finance.expenses":         own("يشوف المصاريف (الكل لصاحب الصلاحية — authz.Expenses)"),
		"finance.expense_add":      own("يسجّل مصروف"),
		"finance.expense_status":   core.RolePermRule("يعتمد/يرفض مصروف", fin, "finance_audit", "expenses_manage"),
		"finance.gps_costs":        core.RolePermRule("يشوف كلف تنصيب الـGPS", fin, "finance_audit", "gps_install_costs"),
		"finance.funds":            fromAuthz("يشوف الصناديق", authz.Funds),
		"finance.fund_update":      fromAuthz("يعدّل مبلغ صندوق", authz.FundAmount),
		"finance.fund_topup":       fromAuthz("يعبّي صندوق", authz.FundAmount),
		"finance.fund_disburse":    fromAuthz("يصرف من الصندوق لموظف", authz.Funds),
		"finance.fund_balances":    fromAuthz("يشوف أرصدة الموظفين", authz.Funds),
		"finance.fund_txns":        fromAuthz("يشوف حركات الصندوق", authz.Funds),
		"finance.fund_receipt":     fromAuthz("يشوف وصل حركة", authz.Funds),
		"finance.settle_for":       fromAuthz("يسجّل تسوية بدل موظف", authz.Funds),
		"finance.settle_review":    fromAuthz("يراجع تسوية (مو تسويته هو — ReviewSettlement)", authz.Funds),
		"finance.settle_discharge": fromAuthz("يبرّي ذمة تسوية", authz.FundDischarge),
		"finance.discharge_accts":  fromAuthz("يشوف حسابات الإبراء", authz.Funds),
		"finance.my_balance":       own("يشوف رصيده هو"),
		"finance.my_txns":          own("يشوف حركاته هو"),
		"finance.my_settle":        own("يسجّل تسوية إله"),
	})
}
