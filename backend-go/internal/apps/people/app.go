// Package people — الموظفين على «ون بيس». يبدي بالأخطر: إضافة موظف.
package people

import (
	"net/http"

	"staffmange-api/internal/core"
)

type Deps struct {
	Create http.HandlerFunc
}

type App struct{ deps Deps }

func New(d Deps) *App { return &App{deps: d} }

func (*App) Name() string { return "people" }

func (*App) Policies() map[core.Action]core.Rule {
	return map[core.Action]core.Rule{
		// قبل: requireAuth بس، والفحص جوّا المعالج. المدير بلا الصلاحية ما يضيف (Strict).
		"people.create": {Label: "يضيف موظف جديد", Strict: true, Roles: []string{"OWNER"}, Perms: []string{"employee_create"}},
	}
}

func (a *App) Routes(r *core.Router) {
	r.Handle("POST /api/employees", "people.create", a.deps.Create)
}
