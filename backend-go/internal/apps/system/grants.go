package system

import (
	"encoding/json"
	"net/http"
	"sort"

	"staffmange-api/internal/middleware"
)

// ═══ صلاحيات المالك لمدير النظام (ون بيس 11-10) ═══
// «خلي صلاحية للمالك ينطي صلاحياته لمدير النظام — بعيداً عن صلاحيات الموظفين».
// المالك يختار فعل فعل من أفعاله (OwnerOnly)، وبس لحساب دوره ADMIN. الإنطاء نفسه
// (هالمسارين) NoDelegate — يبقى بيد المالك وحده.

// GrantsFor صلاحيات المالك الي عند الموظف — تستعملها النواة بكل طلب يحتاجها.
func GrantsFor(a *App) func(string) map[string]bool {
	return func(id string) map[string]bool {
		rows := []string{}
		if err := a.deps.DB.Select(&rows, `SELECT action FROM "OwnerGrant" WHERE "employeeId" = $1`, id); err != nil {
			return nil
		}
		out := make(map[string]bool, len(rows))
		for _, r := range rows {
			out[r] = true
		}
		return out
	}
}

type delegable struct {
	Action string `json:"action"`
	Label  string `json:"label"`
}

func (a *App) delegableActions() []delegable {
	out := []delegable{}
	for _, act := range a.router.Policies().Actions() {
		r, _ := a.router.Policies().Rule(act)
		if r.OwnerOnly && !r.NoDelegate {
			out = append(out, delegable{string(act), r.Label})
		}
	}
	return out
}

func (a *App) grantsView(w http.ResponseWriter, _ *http.Request) {
	admins := []struct {
		ID   string `db:"id" json:"id"`
		Name string `db:"name" json:"name"`
	}{}
	_ = a.deps.DB.Select(&admins, `SELECT id, name FROM "Employee" WHERE role::text = 'ADMIN' AND status::text = 'ACTIVE' ORDER BY name`)
	rows := []struct {
		Emp string `db:"employeeId"`
		Act string `db:"action"`
	}{}
	_ = a.deps.DB.Select(&rows, `SELECT "employeeId", action FROM "OwnerGrant"`)
	grants := map[string][]string{}
	for _, r := range rows {
		grants[r.Emp] = append(grants[r.Emp], r.Act)
	}
	log := []struct {
		Employee string `db:"employee" json:"employee"`
		Action   string `db:"action" json:"action"`
		Granted  bool   `db:"granted" json:"granted"`
		By       string `db:"by" json:"by"`
		At       string `db:"at" json:"at"`
	}{}
	_ = a.deps.DB.Select(&log, `SELECT COALESCE(e.name, l."employeeId") AS employee, l.action, l.granted,
		COALESCE(b.name, '') AS by, to_char(l."createdAt" AT TIME ZONE 'Asia/Baghdad', 'YYYY-MM-DD HH24:MI') AS at
		FROM "OwnerGrantLog" l LEFT JOIN "Employee" e ON e.id = l."employeeId" LEFT JOIN "Employee" b ON b.id = l."byId"
		ORDER BY l."createdAt" DESC LIMIT 50`)
	writeJSON(w, map[string]any{"admins": admins, "delegable": a.delegableActions(), "grants": grants, "log": log})
}

// grantsSet يبدّل مجموعة الأفعال لمدير واحد (الي ينضاف ينسجل «انعطى»، والي ينشال «انسحب»).
func (a *App) grantsSet(w http.ResponseWriter, r *http.Request) {
	emp := r.PathValue("employeeId")
	var in struct {
		Actions []string `json:"actions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	var isAdmin bool
	_ = a.deps.DB.Get(&isAdmin, `SELECT EXISTS (SELECT 1 FROM "Employee" WHERE id = $1 AND role::text = 'ADMIN' AND status::text = 'ACTIVE')`, emp)
	if !isAdmin {
		writeErr(w, http.StatusBadRequest, "صلاحيات المالك تنعطى لمدير النظام بس")
		return
	}
	allowed := map[string]bool{}
	for _, d := range a.delegableActions() {
		allowed[d.Action] = true
	}
	want := map[string]bool{}
	for _, act := range in.Actions {
		if !allowed[act] {
			writeErr(w, http.StatusBadRequest, "فعل ما ينعطى: "+act)
			return
		}
		want[act] = true
	}
	have := GrantsFor(a)(emp)
	me := middleware.EmployeeIDFromContext(r)
	tx, err := a.deps.DB.Beginx()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "تعذر")
		return
	}
	defer tx.Rollback() //nolint:errcheck
	changed := []string{}
	for act := range want {
		if !have[act] {
			tx.MustExec(`INSERT INTO "OwnerGrant" ("employeeId", action, "grantedById") VALUES ($1, $2, $3)`, emp, act, me)
			tx.MustExec(`INSERT INTO "OwnerGrantLog" ("employeeId", action, granted, "byId") VALUES ($1, $2, true, $3)`, emp, act, me)
			changed = append(changed, "+"+act)
		}
	}
	for act := range have {
		if !want[act] {
			tx.MustExec(`DELETE FROM "OwnerGrant" WHERE "employeeId" = $1 AND action = $2`, emp, act)
			tx.MustExec(`INSERT INTO "OwnerGrantLog" ("employeeId", action, granted, "byId") VALUES ($1, $2, false, $3)`, emp, act, me)
			changed = append(changed, "-"+act)
		}
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	sort.Strings(changed)
	writeJSON(w, map[string]any{"ok": true, "changed": changed})
}

// myGrants أفعال المالك الي عند صاحب الطلب — الواجهة تطلّع البنود والشاشات حسبها.
// (المالك نفسه ما يحتاجها — كلشي إله.)
func (a *App) myGrants(w http.ResponseWriter, r *http.Request) {
	out := []string{}
	if middleware.RoleFromContext(r) == "ADMIN" {
		for act := range GrantsFor(a)(middleware.EmployeeIDFromContext(r)) {
			out = append(out, act)
		}
		sort.Strings(out)
	}
	writeJSON(w, out)
}
