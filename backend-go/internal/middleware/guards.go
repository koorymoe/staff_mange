package middleware

import (
	"net/http"

	"staffmange-api/internal/repository"
)

// Guards (DRY 11-02) — نفس دوال Require* بالضبط، بس المستودعات الثلاثة
// (الصلاحيات، الموظفين، الإشعارات) تنعطى مرة وحدة بدل ما تتكرر بكل مسار.
// ⚠️ تمرير مباشر بلا أي منطق إضافي — الحارس ما يتغير.
type Guards struct {
	p *repository.PermissionRepository
	e *repository.EmployeeRepository
	n *repository.NotificationRepository
}

func NewGuards(p *repository.PermissionRepository, e *repository.EmployeeRepository, n *repository.NotificationRepository) *Guards {
	return &Guards{p: p, e: e, n: n}
}

type mw = func(http.Handler) http.Handler

func (g *Guards) Permission(name string) mw { return RequirePermission(g.p, g.e, g.n, name) }
func (g *Guards) AnyPermission(names ...string) mw {
	return RequireAnyPermission(g.p, g.e, g.n, names...)
}
func (g *Guards) RoleOrPermission(roles []string, name string) mw {
	return RequireRoleOrPermission(g.p, g.e, g.n, roles, name)
}
func (g *Guards) RoleOrAnyPermission(roles []string, names ...string) mw {
	return RequireRoleOrAnyPermission(g.p, g.e, g.n, roles, names...)
}
func (g *Guards) LeaderOrPermission(name string) mw {
	return RequireLeaderOrPermission(g.p, g.e, g.n, name)
}
func (g *Guards) LeaderOrAnyPermission(names ...string) mw {
	return RequireLeaderOrAnyPermission(g.p, g.e, g.n, names...)
}
func (g *Guards) LeaderOrRoleOrAnyPermission(roles []string, names ...string) mw {
	return RequireLeaderOrRoleOrAnyPermission(g.p, g.e, g.n, roles, names...)
}
