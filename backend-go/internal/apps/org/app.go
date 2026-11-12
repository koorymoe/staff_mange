// Package org — مكتب الإدارات (المشاكل، الخطط، التوجيهات) على «ون بيس».
//
// المسار مفتوح لأي موظف داخل، بس **الشي نفسه** (هالمشكلة، هالخطة، هالتوجيه)
// ينفحص بالخدمة حسب منصب الشخص بالهيكل — والي ما إله يرجع ErrOrgHidden (404).
// فالقاعدة هنا «أي موظف» عن قصد، والوصف يگول وين الفحص الحقيقي.
package org

import (
	"net/http"

	"staffmange-api/internal/core"
)

type Deps struct {
	AddIssue, UpdateIssue, SetSolution                  http.HandlerFunc
	CreatePlan, AddPlanItem, UpdatePlanItem, SubmitPlan http.HandlerFunc
	ReviewPlan, DraftPlanAI, AddComment                 http.HandlerFunc
	IssueDirective, AckDirective, ArchiveDirective      http.HandlerFunc
}

type App struct{ d Deps }

func New(d Deps) *App { return &App{d: d} }

func (*App) Name() string { return "org" }

const byPosition = " — الفحص على الشي نفسه حسب المنصب بالخدمة (ErrOrgHidden ← 404)"

func (*App) Policies() map[core.Action]core.Rule {
	rule := func(label string) core.Rule { return core.Rule{Label: label + byPosition, Anyone: true} }
	return map[core.Action]core.Rule{
		"org.issue_add":         rule("يرفع مشكلة بإدارته"),
		"org.issue_update":      rule("يحدّث حالة مشكلة"),
		"org.issue_solution":    rule("يكتب خطة حل لمشكلة"),
		"org.plan_create":       rule("ينشئ خطة شهرية لإدارته"),
		"org.plan_item_add":     rule("يضيف بند للخطة"),
		"org.plan_item_update":  rule("يحدّث تقدّم بند"),
		"org.plan_submit":       rule("يرفع الخطة للاعتماد"),
		"org.plan_review":       rule("يعتمد أو يرجّع خطة"),
		"org.plan_draft_ai":     rule("يطلب مسودة خطة من الذكاء"),
		"org.comment_add":       rule("يعلّق على مشكلة أو خطة"),
		"org.directive_issue":   rule("يصدر توجيه لمرؤوسيه"),
		"org.directive_ack":     rule("يأكد استلام توجيه إله"),
		"org.directive_archive": rule("يأرشف توجيه أصدره"),
	}
}

func (a *App) Routes(r *core.Router) {
	r.Handle("POST /api/org/issues", "org.issue_add", a.d.AddIssue)
	r.Handle("PATCH /api/org/issues/{id}", "org.issue_update", a.d.UpdateIssue)
	r.Handle("PUT /api/org/issues/{id}/solution", "org.issue_solution", a.d.SetSolution)
	r.Handle("POST /api/org/plans", "org.plan_create", a.d.CreatePlan)
	r.Handle("POST /api/org/plans/{id}/items", "org.plan_item_add", a.d.AddPlanItem)
	r.Handle("PATCH /api/org/plan-items/{id}", "org.plan_item_update", a.d.UpdatePlanItem)
	r.Handle("POST /api/org/plans/{id}/submit", "org.plan_submit", a.d.SubmitPlan)
	r.Handle("POST /api/org/plans/{id}/review", "org.plan_review", a.d.ReviewPlan)
	r.Handle("POST /api/org/plans/{id}/draft-ai", "org.plan_draft_ai", a.d.DraftPlanAI)
	r.Handle("POST /api/org/comments", "org.comment_add", a.d.AddComment)
	r.Handle("POST /api/org/directives", "org.directive_issue", a.d.IssueDirective)
	r.Handle("POST /api/org/directives/{id}/ack", "org.directive_ack", a.d.AckDirective)
	r.Handle("POST /api/org/directives/{id}/archive", "org.directive_archive", a.d.ArchiveDirective)
}
