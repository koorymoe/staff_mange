// Package sections — باقي أقسام النظام على «ون بيس» (منقولة دفعة وحدة 11-10 بـ
// scripts/onepiece-bulk.py). كل قسم = تطبيق باسم أول جزء من المسار بعد /api.
//
// معنى كل قاعدة نفس الحارس القديم بالضبط:
//
//	requireAdmin / RequireRole(... "ADMIN" ...)  ← roles: الدور (أو المنصب)، والمالك والمدير يعبرون
//	RequireRole بلا "ADMIN"                       ← strictRoles: المدير ما يعبر (مثل قبل)، والمالك يعبر
//	guards.Permission/AnyPermission               ← perm
//	guards.RoleOr(Any)Permission                  ← rolePerm
//	guards.LeaderOr(Any)Permission                ← leaderPerm
//	authz.Guard(…, authz.X)                       ← fromAuthz (القاعدة من authz.SeeAllRules نفسها)
//	RequireOwner / RequireOwnerOnly               ← owner
//	حارس ما ينترجم بأمان                          ← يبقى طبقة ورا القاعدة بـmain.go
package sections

import (
	"sort"

	"staffmange-api/internal/authz"
	"staffmange-api/internal/core"
)

func any_(l string) core.Rule  { return core.Rule{Label: l, Anyone: true} }
func admin(l string) core.Rule { return core.AdminRule(l) }
func owner(l string) core.Rule { return core.Rule{Label: l, OwnerOnly: true} }
func perm(l string, p ...string) core.Rule {
	return core.PermRule(l, p...)
}
func rolePerm(l string, roles []string, p ...string) core.Rule {
	return core.RolePermRule(l, roles, p...)
}
func leaderPerm(l string, p ...string) core.Rule { return core.LeaderPermRule(l, p...) }
func leaderRolePerm(l string, roles []string, p ...string) core.Rule {
	return core.Rule{Label: l, Leader: true, Roles: roles, Perms: p}
}
func roles(l string, r ...string) core.Rule { return core.Rule{Label: l, Roles: r} }

// strictRoles نفس RequireRole القديم لمن ما بيه "ADMIN": المالك يعبر دائماً، والمدير لا.
func strictRoles(l string, r ...string) core.Rule {
	return core.Rule{Label: l, Strict: true, Roles: append([]string{"OWNER"}, r...)}
}

// fromAuthz نفس authz.Guard: أدوار القاعدة أو صلاحياتها، والمالك والمدير يعبرون.
func fromAuthz(l string, res authz.Resource) core.Rule {
	r := authz.SeeAllRules[res]
	return core.RolePermRule(l, r.Roles, r.Perms...)
}

// New تطبيق لكل قسم، وقواعده جاهزة — المعالجات تنربط بـmain (Set).
func New() map[string]*core.SetApp {
	out := map[string]*core.SetApp{}
	for name, rs := range rules() {
		out[name] = core.NewSetApp(name, rs)
	}
	return out
}

// Apps للتسجيل (Mount) بترتيب ثابت.
func Apps(m map[string]*core.SetApp) []core.App {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]core.App, 0, len(keys))
	for _, k := range keys {
		out = append(out, m[k])
	}
	return out
}
