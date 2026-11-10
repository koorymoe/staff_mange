// Package system — أول تطبيق على «ون بيس»: شاشة تدقيق الصلاحيات.
// (دازاي 11-10: «هاي الأشياء مهمة جداً، إذا نريد ندقق بيهن»)
package system

import (
	"encoding/json"
	"net/http"

	"staffmange-api/internal/core"
	"staffmange-api/internal/model"
)

// Deps معالجات قديمة تنتقل تحت قواعد النواة (المنطق نفسه، الحارس صار هنا).
type Deps struct {
	SetSwitch http.HandlerFunc
}

type App struct {
	router *core.Router
	deps   Deps
}

func New(d Deps) *App { return &App{deps: d} }

func (*App) Name() string { return "system" }

func (*App) Policies() map[core.Action]core.Rule {
	return map[core.Action]core.Rule{
		"system.matrix": {Label: "يشوف مصفوفة الصلاحيات (كل فعل، منو يگدر عليه، ومساراته)", OwnerOnly: true},
		// المالك دايماً، ومدير النظام بس بالمفاتيح المسموحة له
		"system.switch_set": {Label: "يشغّل/يطفي مفتاح بالنظام", Strict: true, Check: func(a *core.Actor, r *http.Request) bool {
			return a.Owner() || (a.Role == "ADMIN" && model.SystemSwitchAdminAllowed(r.PathValue("key")))
		}},
	}
}

func (a *App) Routes(r *core.Router) {
	a.router = r
	r.Handle("PUT /api/system/switches/{key}", "system.switch_set", a.deps.SetSwitch)
	r.Handle("GET /api/core/matrix", "system.matrix", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"actions": a.router.Matrix(), "routes": a.router.Routes()})
	})
}
