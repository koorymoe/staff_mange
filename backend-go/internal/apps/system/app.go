// Package system — أول تطبيق على «ون بيس»: شاشة تدقيق الصلاحيات.
// (دازاي 11-10: «هاي الأشياء مهمة جداً، إذا نريد ندقق بيهن»)
package system

import (
	"encoding/json"
	"net/http"

	"staffmange-api/internal/core"
)

type App struct{ router *core.Router }

func New() *App { return &App{} }

func (*App) Name() string { return "system" }

func (*App) Policies() map[core.Action]core.Rule {
	return map[core.Action]core.Rule{
		"system.matrix": {Label: "يشوف مصفوفة الصلاحيات (كل فعل، منو يگدر عليه، ومساراته)", OwnerOnly: true},
	}
}

func (a *App) Routes(r *core.Router) {
	a.router = r
	r.Handle("GET /api/core/matrix", "system.matrix", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"actions": a.router.Matrix(), "routes": a.router.Routes()})
	})
}
