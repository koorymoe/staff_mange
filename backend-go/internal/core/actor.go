// Package core — نواة «ون بيس» (دازاي 11-10).
//
// كل شي بالنظام يمر على سؤال واحد: **منو هذا، شنو دوره، شنو منصبه، شنو
// صلاحياته** (Actor)، وكل مسار يذكر **شنو يسوي** (Action) — والقرار يصير بمكان
// واحد (Policy). التطبيقات (apps) تسجّل سياساتها ومساراتها، والنواة ترفض
// تشغيل السيرفر إذا مسار بلا سياسة.
//
// ⚠️ النواة تبني على الموجود (middleware.RequireAuth، الهيكل التنظيمي، authz)
// — ما تعيد كتابته. فالمسار القديم والجديد يقرون نفس الحقيقة.
package core

import (
	"context"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

// PermLister مصدر صلاحيات الموظف الفعلية (PermissionRepository.Names).
type PermLister interface {
	Names(employeeID string) ([]string, error)
}

// Actor صاحب الطلب — يتحسب مرة وحدة بالطلب.
type Actor struct {
	ID            string
	Role          string   // الدور الأصلي بالحساب
	EffectiveRole string   // الدور الفعلي (يشمل المنصب بالهيكل)
	OrgRoles      []string // أدوار المنصب
	Perms         map[string]bool
	Stealth       bool // المالك داخل بحساب موظف (قراءة بس)
}

// Owner المالك الأصلي.
func (a *Actor) Owner() bool { return a != nil && a.Role == "OWNER" }

// Admin المالك أو مدير النظام (الدور الأصلي).
func (a *Actor) Admin() bool { return a != nil && model.IsAdminRole(a.Role) }

// HasRole الدور الأصلي أو الفعلي أو أي دور من المنصب.
func (a *Actor) HasRole(roles ...string) bool {
	if a == nil {
		return false
	}
	for _, r := range roles {
		if r == a.Role || r == a.EffectiveRole {
			return true
		}
		for _, o := range a.OrgRoles {
			if o == r {
				return true
			}
		}
	}
	return false
}

// Can يملك وحدة من الصلاحيات؟
func (a *Actor) Can(perms ...string) bool {
	if a == nil {
		return false
	}
	for _, p := range perms {
		if a.Perms[p] {
			return true
		}
	}
	return false
}

type actorKey struct{}

// ActorOf الـActor المحسوب (nil إذا الطلب بلا دخول).
func ActorOf(r *http.Request) *Actor {
	a, _ := r.Context().Value(actorKey{}).(*Actor)
	return a
}

// resolveActor يبني الـActor من سياق RequireAuth. خطأ بقراءة الصلاحيات = بلا
// صلاحيات (يضيّق ما يوسّع).
func resolveActor(r *http.Request, perms PermLister) *Actor {
	id := middleware.EmployeeIDFromContext(r)
	if id == "" {
		return nil
	}
	a := &Actor{ID: id, Role: middleware.RoleFromContext(r), EffectiveRole: middleware.EffectiveRole(r),
		OrgRoles: middleware.OrgRoles(r), Perms: map[string]bool{}, Stealth: middleware.IsStealth(r)}
	if perms != nil {
		if names, err := perms.Names(id); err == nil {
			for _, n := range names {
				a.Perms[n] = true
			}
		}
	}
	return a
}

func withActor(r *http.Request, a *Actor) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), actorKey{}, a))
}
