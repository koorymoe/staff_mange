package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// ═══ «الصلاحيات والوصول» — خادم المرحلة ١ (دازاي 10-26، العقد ويا ساسكي) ═══
// مجموعات، ومصدر كل صلاحية (قالب المنصب / منح يدوي / استثناء معتمد)، وحفظ بنسخة
// (٤٠٩ إذا زميل عدّل)، ومطابقة الصنف، وسجل كل تغيير، والمؤقتة تنشال لحالها.
type PermissionAccessHandler struct{ db *sqlx.DB }

func NewPermissionAccessHandler(db *sqlx.DB) *PermissionAccessHandler {
	return &PermissionAccessHandler{db: db}
}

type accessPerm struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Grantable  bool   `json:"grantable"`
	LockReason string `json:"lockReason,omitempty"`
}

type accessGroup struct {
	Key    string       `json:"key"`
	Title  string       `json:"title"`
	Bundle string       `json:"bundle,omitempty"`
	Perms  []accessPerm `json:"perms"`
}

func isOwnerReq(r *http.Request) bool { return middleware.RoleFromContext(r) == "OWNER" }

func lockFor(r *http.Request, name string) string {
	if why, ok := model.PermissionOwnerOnly[name]; ok && !isOwnerReq(r) {
		return why
	}
	return ""
}

// templateFor قالب المنصب = صلاحيات دوره + أدواره الثانوية.
func templateFor(role string, secondary []string) map[string]bool {
	out := map[string]bool{}
	for _, rl := range append([]string{role}, secondary...) {
		for _, p := range model.RoleDefaultPermissions[rl] {
			out[p] = true
		}
	}
	return out
}

// GET /api/permissions/catalog
func (h *PermissionAccessHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	var all []struct {
		Name  string `db:"name"`
		Label string `db:"label"`
	}
	if err := h.db.Select(&all, `SELECT name, label FROM "Permission" ORDER BY name`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الصلاحيات")
		return
	}
	label := map[string]string{}
	for _, p := range all {
		label[p.Name] = p.Label
	}
	used := map[string]bool{}
	mk := func(name string) accessPerm {
		used[name] = true
		lock := lockFor(r, name)
		return accessPerm{Name: name, Label: label[name], Grantable: lock == "", LockReason: lock}
	}
	out := []accessGroup{}
	for _, g := range model.PermissionGroups {
		ag := accessGroup{Key: g.Key, Title: g.Title, Bundle: g.Bundle, Perms: []accessPerm{}}
		for _, n := range g.Perms {
			if _, ok := label[n]; ok {
				ag.Perms = append(ag.Perms, mk(n))
			}
		}
		out = append(out, ag)
	}
	other := accessGroup{Key: "other", Title: "أخرى", Perms: []accessPerm{}}
	for _, p := range all {
		if !used[p.Name] {
			other.Perms = append(other.Perms, mk(p.Name))
		}
	}
	if len(other.Perms) > 0 {
		out = append(out, other)
	}
	WriteJSON(w, http.StatusOK, out)
}

type accessEmployee struct {
	ID             string         `db:"id" json:"id"`
	Name           string         `db:"name" json:"name"`
	Role           string         `db:"role" json:"role"`
	SecondaryRoles pq.StringArray `db:"secondaryRoles" json:"secondaryRoles"`
	IsLeader       bool           `db:"isLeader" json:"isLeader"`
	Version        int            `db:"accessVersion" json:"-"`
	// Org الصلاحيات الي جت ويا منصب بالهيكل (الهيكل الجديد 10-29)
	Org map[string]bool `db:"-" json:"-"`
}

type accessItem struct {
	Name      string     `json:"name"`
	Source    string     `json:"source"` // TEMPLATE | MANUAL | EXCEPTION | ORG
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Reason    string     `json:"reason,omitempty"`
}

type exceptionRow struct {
	Permission string     `db:"permission"`
	Reason     string     `db:"reason"`
	ExpiresAt  *time.Time `db:"expiresAt"`
}

func (h *PermissionAccessHandler) state(id string) (*accessEmployee, map[string]bool, map[string]exceptionRow, error) {
	var e accessEmployee
	if err := h.db.Get(&e, `SELECT id, name, role::text AS role, COALESCE("secondaryRoles", '{}') AS "secondaryRoles", "isLeader", "accessVersion"
		FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil, nil, nil, err
	}
	var names []string
	if err := h.db.Select(&names, `SELECT p.name FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1`, id); err != nil {
		return nil, nil, nil, err
	}
	have := map[string]bool{}
	for _, n := range names {
		have[n] = true
	}
	var ex []exceptionRow
	_ = h.db.Select(&ex, `SELECT permission, reason, "expiresAt" FROM "PermissionException" WHERE "employeeId" = $1`, id)
	exc := map[string]exceptionRow{}
	for _, x := range ex {
		exc[x.Permission] = x
	}
	var org []string
	_ = h.db.Select(&org, `SELECT permission FROM "OrgPermissionGrant" WHERE "employeeId" = $1`, id)
	e.Org = map[string]bool{}
	for _, n := range org {
		e.Org[n] = true
	}
	return &e, have, exc, nil
}

// GET /api/employees/{id}/access
func (h *PermissionAccessHandler) Access(w http.ResponseWriter, r *http.Request) {
	e, have, exc, err := h.state(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "الموظف مو موجود")
		return
	}
	tpl := templateFor(e.Role, e.SecondaryRoles)
	items := []accessItem{}
	for n := range have {
		it := accessItem{Name: n, Source: "MANUAL"}
		if x, ok := exc[n]; ok {
			it.Source, it.ExpiresAt, it.Reason = "EXCEPTION", x.ExpiresAt, x.Reason
		} else if tpl[n] {
			it.Source = "TEMPLATE"
		} else if e.Org[n] {
			it.Source = "ORG"
		}
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	missing := []string{}
	for n := range tpl {
		if !have[n] {
			missing = append(missing, n)
		}
	}
	sort.Strings(missing)
	WriteJSON(w, http.StatusOK, map[string]any{"employee": e, "perms": items, "missingFromTemplate": missing, "version": e.Version})
}

func (h *PermissionAccessHandler) audit(tx *sqlx.Tx, empID, actor, action, perm, reason string) {
	_, _ = tx.Exec(`INSERT INTO "PermissionAuditLog" ("employeeId", "actorId", action, permission, reason) VALUES ($1, $2, $3, $4, NULLIF($5, ''))`,
		empID, actor, action, perm, reason)
}

// apply يضيف ويشيل بمعاملة وحدة ويا التحقق من النسخة والقفل والسجل.
// extra (اختياري) ينفّذ جوّا نفس المعاملة — فشله يرجّع كلشي.
func (h *PermissionAccessHandler) apply(r *http.Request, empID string, version int, add, remove []string, reason string, extra func(tx *sqlx.Tx, actor string) error) (int, string) {
	if middleware.RoleFromContext(r) != "OWNER" {
		var role string
		_ = h.db.Get(&role, `SELECT role::text FROM "Employee" WHERE id = $1`, empID)
		if role == "OWNER" || role == "ADMIN" {
			return http.StatusForbidden, "صلاحيات المالك والمدير يعدّلها المالك بس"
		}
	}
	for _, n := range append(append([]string{}, add...), remove...) {
		if why := lockFor(r, n); why != "" {
			return http.StatusForbidden, "«" + n + "» مقفولة: " + why
		}
	}
	actor := middleware.EmployeeIDFromContext(r)
	tx, err := h.db.Beginx()
	if err != nil {
		return http.StatusInternalServerError, "تعذر"
	}
	defer tx.Rollback()
	var cur int
	if err := tx.Get(&cur, `SELECT "accessVersion" FROM "Employee" WHERE id = $1 FOR UPDATE`, empID); err != nil {
		return http.StatusNotFound, "الموظف مو موجود"
	}
	if cur != version {
		return http.StatusConflict, "زميل عدّل صلاحيات هذا الموظف هسه — حدّث الصفحة وشوف التغيير قبل ما تحفظ"
	}
	for _, n := range add {
		res, err := tx.Exec(`INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
			SELECT gen_random_uuid()::text, $1, p.id FROM "Permission" p WHERE p.name = $2
			AND NOT EXISTS (SELECT 1 FROM "EmployeePermission" ep WHERE ep."employeeId" = $1 AND ep."permissionId" = p.id)`, empID, n)
		if err != nil {
			return http.StatusInternalServerError, "تعذر الحفظ"
		}
		if k, _ := res.RowsAffected(); k > 0 {
			h.audit(tx, empID, actor, "ADD", n, reason)
		}
		// منح يدوي = الصلاحية صارت يدوية — إنهاء المنصب بعدين ما يسحبها
		if _, err := tx.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = $1 AND permission = $2`, empID, n); err != nil {
			return http.StatusInternalServerError, "تعذر الحفظ"
		}
	}
	// الشيل اليدوي يخلي علامة المنصب (إن وجدت) كـ«رفض» — المزامنة ما ترجّعها لحالها
	for _, n := range remove {
		res, err := tx.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1 AND "permissionId" = (SELECT id FROM "Permission" WHERE name = $2)`, empID, n)
		if err != nil {
			return http.StatusInternalServerError, "تعذر الحفظ"
		}
		if k, _ := res.RowsAffected(); k > 0 {
			h.audit(tx, empID, actor, "REMOVE", n, reason)
			_, _ = tx.Exec(`DELETE FROM "PermissionException" WHERE "employeeId" = $1 AND permission = $2`, empID, n)
		}
	}
	if extra != nil {
		if err := extra(tx, actor); err != nil {
			return http.StatusInternalServerError, "تعذر حفظ الاستثناء"
		}
	}
	if _, err := tx.Exec(`UPDATE "Employee" SET "accessVersion" = "accessVersion" + 1 WHERE id = $1`, empID); err != nil {
		return http.StatusInternalServerError, "تعذر"
	}
	if err := tx.Commit(); err != nil {
		return http.StatusInternalServerError, "تعذر"
	}
	return http.StatusOK, ""
}

// PUT /api/employees/{id}/access — {add, remove, reason, version}
func (h *PermissionAccessHandler) Save(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Add     []string `json:"add"`
		Remove  []string `json:"remove"`
		Reason  string   `json:"reason"`
		Version int      `json:"version"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if code, msg := h.apply(r, r.PathValue("id"), b.Version, b.Add, b.Remove, strings.TrimSpace(b.Reason), nil); code != http.StatusOK {
		WriteError(w, code, msg)
		return
	}
	h.Access(w, r)
}

type driftRow struct {
	EmployeeID string   `json:"employeeId"`
	Name       string   `json:"name"`
	Role       string   `json:"role"`
	Status     string   `json:"status"` // MATCH | DIFF | NO_TEMPLATE
	Extra      []string `json:"extra"`
	Missing    []string `json:"missing"`
	Exceptions []string `json:"exceptions"`
	// Org من المنصب بالهيكل — مو «زيادة» على القالب، و«مطابقة» ما تشيلها
	Org     []string `json:"org"`
	Version int      `json:"version"`
}

func (h *PermissionAccessHandler) drift(e *accessEmployee, have map[string]bool, exc map[string]exceptionRow) driftRow {
	tpl := templateFor(e.Role, e.SecondaryRoles)
	d := driftRow{EmployeeID: e.ID, Name: e.Name, Role: e.Role, Extra: []string{}, Missing: []string{}, Exceptions: []string{}, Org: []string{}, Version: e.Version}
	_, hasTpl := model.RoleDefaultPermissions[e.Role]
	for n := range have {
		switch {
		case exc[n].Permission != "":
			d.Exceptions = append(d.Exceptions, n)
		case e.Org[n] && !tpl[n]:
			d.Org = append(d.Org, n)
		case !tpl[n]:
			d.Extra = append(d.Extra, n)
		}
	}
	for n := range tpl {
		if !have[n] {
			d.Missing = append(d.Missing, n)
		}
	}
	sort.Strings(d.Extra)
	sort.Strings(d.Missing)
	sort.Strings(d.Exceptions)
	sort.Strings(d.Org)
	switch {
	case !hasTpl:
		d.Status = "NO_TEMPLATE"
	case len(d.Extra) == 0 && len(d.Missing) == 0:
		d.Status = "MATCH"
	default:
		d.Status = "DIFF"
	}
	return d
}

// GET /api/permissions/role-drift
func (h *PermissionAccessHandler) RoleDrift(w http.ResponseWriter, r *http.Request) {
	var ids []string
	if err := h.db.Select(&ids, `SELECT id FROM "Employee" WHERE status = 'ACTIVE' AND role NOT IN ('ADMIN', 'OWNER') ORDER BY name`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	out := []driftRow{}
	for _, id := range ids {
		e, have, exc, err := h.state(id)
		if err != nil {
			continue
		}
		out = append(out, h.drift(e, have, exc))
	}
	WriteJSON(w, http.StatusOK, out)
}

// POST /api/permissions/role-drift/{id}/align — {reason, version}: يطابقه ويا صنفه،
// والاستثناءات المعتمدة والمقفول يبقون.
func (h *PermissionAccessHandler) Align(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Reason  string `json:"reason"`
		Version int    `json:"version"`
	}
	_ = DecodeJSON(r, &b)
	if len([]rune(strings.TrimSpace(b.Reason))) < 3 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب المطابقة")
		return
	}
	e, have, exc, err := h.state(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "الموظف مو موجود")
		return
	}
	d := h.drift(e, have, exc)
	if d.Status == "NO_TEMPLATE" {
		WriteError(w, http.StatusBadRequest, "هذا الموظف ما إله قالب — حدد صنفه أول")
		return
	}
	keep := func(list []string) []string {
		out := []string{}
		for _, n := range list {
			if lockFor(r, n) == "" {
				out = append(out, n)
			}
		}
		return out
	}
	if code, msg := h.apply(r, e.ID, b.Version, keep(d.Missing), keep(d.Extra), "مطابقة الصنف: "+strings.TrimSpace(b.Reason), nil); code != http.StatusOK {
		WriteError(w, code, msg)
		return
	}
	e, have, exc, _ = h.state(e.ID)
	WriteJSON(w, http.StatusOK, h.drift(e, have, exc))
}

// POST /api/permissions/exceptions — {employeeId, permission, reason, expiresAt?}: يمنحها ويثبّتها استثناء
func (h *PermissionAccessHandler) AddException(w http.ResponseWriter, r *http.Request) {
	var b struct {
		EmployeeID string `json:"employeeId"`
		Permission string `json:"permission"`
		Reason     string `json:"reason"`
		ExpiresAt  string `json:"expiresAt"`
		Version    int    `json:"version"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.EmployeeID == "" || b.Permission == "" {
		WriteError(w, http.StatusBadRequest, "حدد الموظف والصلاحية")
		return
	}
	if len([]rune(strings.TrimSpace(b.Reason))) < 3 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الاستثناء")
		return
	}
	var exp *time.Time
	if b.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, b.ExpiresAt)
		if err != nil || t.Before(time.Now()) {
			WriteError(w, http.StatusBadRequest, "موعد الانتهاء لازم بالمستقبل")
			return
		}
		exp = &t
	}
	reason := strings.TrimSpace(b.Reason)
	if code, msg := h.apply(r, b.EmployeeID, b.Version, []string{b.Permission}, nil, "استثناء: "+reason, func(tx *sqlx.Tx, actor string) error {
		_, err := tx.Exec(`INSERT INTO "PermissionException" ("employeeId", permission, reason, "approvedById", "expiresAt")
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT ("employeeId", permission) DO UPDATE SET reason = EXCLUDED.reason,
			"approvedById" = EXCLUDED."approvedById", "expiresAt" = EXCLUDED."expiresAt"`,
			b.EmployeeID, b.Permission, reason, actor, exp)
		return err
	}); code != http.StatusOK {
		WriteError(w, code, msg)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DELETE /api/permissions/exceptions?employeeId=&permission=&version= — يشيل «الاستثناء» بس (الصلاحية تبقى).
// نفس حراسة الحفظ: النسخة، والأقفال، وحسابات الإدارة، والسجل.
func (h *PermissionAccessHandler) RemoveException(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	version, err := strconv.Atoi(q.Get("version"))
	if err != nil || q.Get("employeeId") == "" || q.Get("permission") == "" {
		WriteError(w, http.StatusBadRequest, "حدد الموظف والصلاحية والنسخة")
		return
	}
	perm := q.Get("permission")
	if why := lockFor(r, perm); why != "" {
		WriteError(w, http.StatusForbidden, "«"+perm+"» مقفولة: "+why)
		return
	}
	if code, msg := h.apply(r, q.Get("employeeId"), version, nil, nil, "", func(tx *sqlx.Tx, actor string) error {
		res, err := tx.Exec(`DELETE FROM "PermissionException" WHERE "employeeId" = $1 AND permission = $2`, q.Get("employeeId"), perm)
		if err != nil {
			return err
		}
		if k, _ := res.RowsAffected(); k > 0 {
			h.audit(tx, q.Get("employeeId"), actor, "UNPIN", perm, "شيل تثبيت الاستثناء — الصلاحية باقية")
		}
		return nil
	}); code != http.StatusOK {
		WriteError(w, code, msg)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/permissions/audit-log?employeeId=
func (h *PermissionAccessHandler) AuditLog(w http.ResponseWriter, r *http.Request) {
	rows := []struct {
		ID         string    `db:"id" json:"id"`
		EmployeeID string    `db:"employeeId" json:"employeeId"`
		Employee   *string   `db:"employeeName" json:"employeeName"`
		Actor      *string   `db:"actorName" json:"actorName"`
		Action     string    `db:"action" json:"action"`
		Permission string    `db:"permission" json:"permission"`
		Reason     *string   `db:"reason" json:"reason"`
		CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
	}{}
	q := `SELECT l.id, l."employeeId", e.name AS "employeeName", a.name AS "actorName", l.action, l.permission, l.reason, l."createdAt"
		FROM "PermissionAuditLog" l LEFT JOIN "Employee" e ON e.id = l."employeeId" LEFT JOIN "Employee" a ON a.id = l."actorId"`
	args := []any{}
	if id := r.URL.Query().Get("employeeId"); id != "" {
		q += ` WHERE l."employeeId" = $1`
		args = append(args, id)
	}
	if err := h.db.Select(&rows, q+` ORDER BY l."createdAt" DESC LIMIT 300`, args...); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// SweepExpired يشيل الاستثناءات المؤقتة الي خلص وقتها ويا صلاحيتها، ويرجّع شكد انشال.
func (h *PermissionAccessHandler) SweepExpired() int {
	var rows []struct {
		EmployeeID string `db:"employeeId"`
		Permission string `db:"permission"`
	}
	_ = h.db.Select(&rows, `SELECT "employeeId", permission FROM "PermissionException" WHERE "expiresAt" IS NOT NULL AND "expiresAt" < now()`)
	n := 0
	for _, x := range rows {
		tx, err := h.db.Beginx()
		if err != nil {
			continue
		}
		if service.OrgWants(tx, x.EmployeeID, x.Permission) {
			// المنصب يستحقها — الاستثناء ينتهي بس الصلاحية تبقى وتصير من المنصب
			_, _ = tx.Exec(`INSERT INTO "OrgPermissionGrant" ("employeeId", permission) VALUES ($1, $2) ON CONFLICT DO NOTHING`, x.EmployeeID, x.Permission)
		} else {
			_, _ = tx.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1 AND "permissionId" = (SELECT id FROM "Permission" WHERE name = $2)`, x.EmployeeID, x.Permission)
		}
		_, _ = tx.Exec(`DELETE FROM "PermissionException" WHERE "employeeId" = $1 AND permission = $2`, x.EmployeeID, x.Permission)
		_, _ = tx.Exec(`UPDATE "Employee" SET "accessVersion" = "accessVersion" + 1 WHERE id = $1`, x.EmployeeID)
		h.audit(tx, x.EmployeeID, "", "EXPIRE", x.Permission, "انتهى وقت الصلاحية المؤقتة")
		if tx.Commit() == nil {
			n++
		}
	}
	return n
}
