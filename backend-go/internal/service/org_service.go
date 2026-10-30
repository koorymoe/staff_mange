package service

import (
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// ═══ 🏛️ الهيكل الجديد (دازاي 10-29) — model/org.go ═══
// المنصب ينطي حزمة صلاحيات مكانه بنظام الصلاحيات الحالي (EmployeePermission)،
// ونسجّل بـOrgPermissionGrant بس الي انطت بسببه — فسحب المنصب يشيلها وحدها
// والصلاحيات اليدوية القديمة ما تنمس.

type OrgService struct{ db *sqlx.DB }

func NewOrgService(db *sqlx.DB) *OrgService { return &OrgService{db: db} }

// Seed يزرع الإدارات والوحدات وحزمها إذا ناقصة. ما يطمس الي تغيّر بعدين من الشاشة.
func (s *OrgService) Seed() error {
	changed := false
	for i, d := range model.OrgSeed {
		var parent any
		if d.Parent != "" {
			parent = d.Parent
		}
		if _, err := s.db.Exec(`INSERT INTO "OrgDivision" (id, name, "parentId", sort, summary) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING`, d.ID, d.Name, parent, i, d.Summary); err != nil {
			return err
		}
		if err := s.seedBundle(d.ID, d.Perms); err != nil {
			return err
		}
		if d.HeadAll {
			if err := s.markHead(d.ID, d.Perms, &changed); err != nil {
				return err
			}
		}
		for j, u := range d.Units {
			if _, err := s.db.Exec(`INSERT INTO "OrgUnit" (id, "divisionId", name, sort) VALUES ($1, $2, $3, $4)
				ON CONFLICT (id) DO NOTHING`, u.ID, d.ID, u.Name, j); err != nil {
				return err
			}
			if err := s.seedBundle(u.ID, u.Perms); err != nil {
				return err
			}
			if err := s.markHead(u.ID, u.Head, &changed); err != nil {
				return err
			}
		}
	}
	if changed {
		return s.resyncAll("صلاحيات حسب المنصب")
	}
	return nil
}

// markHead يعلّم صلاحيات الرئيس بالحزمة — مرة وحدة لكل حزمة (إذا ماكو ولا وحدة معلّمة).
func (s *OrgService) markHead(scope string, head []string, changed *bool) error {
	if len(head) == 0 {
		return nil
	}
	var n int
	if err := s.db.Get(&n, `SELECT count(*) FROM "OrgBundle" WHERE scope = $1 AND "headOnly"`, scope); err != nil || n > 0 {
		return err
	}
	res, err := s.db.Exec(`UPDATE "OrgBundle" SET "headOnly" = true WHERE scope = $1 AND permission = ANY($2)`, scope, pq.Array(head))
	if err != nil {
		return err
	}
	if k, _ := res.RowsAffected(); k > 0 {
		*changed = true
	}
	return nil
}

// resyncAll يعيد حساب صلاحيات الهيكل لكل صاحب منصب فعّال (بعد تغيير الحزم).
func (s *OrgService) resyncAll(reason string) error {
	var emps []string
	if err := s.db.Select(&emps, `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "endedAt" IS NULL`); err != nil {
		return err
	}
	for _, e := range emps {
		tx, err := s.db.Beginx()
		if err != nil {
			return err
		}
		if err := syncOrgPermissions(tx, e, "", reason); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (s *OrgService) seedBundle(scope string, perms []string) error {
	var n int
	if err := s.db.Get(&n, `SELECT count(*) FROM "OrgBundle" WHERE scope = $1`, scope); err != nil || n > 0 {
		return err
	}
	for _, p := range perms {
		if _, ownerOnly := model.PermissionOwnerOnly[p]; ownerOnly {
			continue
		}
		// مفتاح مو موجود بالكتالوج يتجاهل (ما نكسر التعيين على اسم غلط)
		if _, err := s.db.Exec(`INSERT INTO "OrgBundle" (scope, permission)
			SELECT $1, name FROM "Permission" WHERE name = $2 ON CONFLICT DO NOTHING`, scope, p); err != nil {
			return err
		}
	}
	return nil
}

// ── الشجرة ──

type OrgChartUnit struct {
	model.OrgUnit
	People []model.OrgAssignment `json:"people"`
}

type OrgChartDivision struct {
	model.OrgDivision
	Permissions []string `json:"permissions"`
	// HeadPermissions الجزء من Permissions الي للمدير/النائب بس (headOnly)
	HeadPermissions []string              `json:"headPermissions"`
	People          []model.OrgAssignment `json:"people"`
	Units           []OrgChartUnit        `json:"units"`
}

const orgAssignCols = `a.id, a."employeeId", e.name AS "employeeName", a."divisionId", a."unitId", a.position,
	a."allowanceIQD", a.acting, a."actingUntil", a."replacesId", a."startedAt", a."endedAt", a."endedById", a."endReason", a.note`

// bundles حزم الصلاحيات: الكل، والجزء الي للرئيس/النائب بس.
func (s *OrgService) bundles() (all, head map[string][]string, err error) {
	rows := []struct {
		Scope string `db:"scope"`
		Perm  string `db:"permission"`
		Head  bool   `db:"headOnly"`
	}{}
	if err := s.db.Select(&rows, `SELECT scope, permission, "headOnly" FROM "OrgBundle" ORDER BY permission`); err != nil {
		return nil, nil, err
	}
	all, head = map[string][]string{}, map[string][]string{}
	for _, r := range rows {
		all[r.Scope] = append(all[r.Scope], r.Perm)
		if r.Head {
			head[r.Scope] = append(head[r.Scope], r.Perm)
		}
	}
	return all, head, nil
}

func (s *OrgService) Chart() ([]OrgChartDivision, error) {
	divs := []model.OrgDivision{}
	if err := s.db.Select(&divs, `SELECT id, name, "parentId", sort, summary FROM "OrgDivision" ORDER BY sort`); err != nil {
		return nil, err
	}
	units := []model.OrgUnit{}
	if err := s.db.Select(&units, `SELECT id, "divisionId", name, sort FROM "OrgUnit" ORDER BY sort`); err != nil {
		return nil, err
	}
	people := []model.OrgAssignment{}
	if err := s.db.Select(&people, `SELECT `+orgAssignCols+` FROM "OrgAssignment" a JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a."endedAt" IS NULL ORDER BY CASE a.position WHEN 'DIRECTOR' THEN 0 WHEN 'DEPUTY' THEN 1 ELSE 2 END, e.name`); err != nil {
		return nil, err
	}
	b, hb, err := s.bundles()
	if err != nil {
		return nil, err
	}
	out := make([]OrgChartDivision, 0, len(divs))
	for _, d := range divs {
		cd := OrgChartDivision{OrgDivision: d, Permissions: nz(b[d.ID]), HeadPermissions: nz(hb[d.ID]), People: []model.OrgAssignment{}, Units: []OrgChartUnit{}}
		for _, p := range people {
			if p.DivisionID == d.ID && p.UnitID == nil {
				cd.People = append(cd.People, p)
			}
		}
		for _, u := range units {
			if u.DivisionID != d.ID {
				continue
			}
			u.Permissions = nz(b[u.ID])
			u.HeadPermissions = nz(hb[u.ID])
			cu := OrgChartUnit{OrgUnit: u, People: []model.OrgAssignment{}}
			for _, p := range people {
				if p.UnitID != nil && *p.UnitID == u.ID {
					cu.People = append(cu.People, p)
				}
			}
			cd.Units = append(cd.Units, cu)
		}
		out = append(out, cd)
	}
	return out, nil
}

func nz(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// Mine مناصب موظف (الفعّالة).
func (s *OrgService) Mine(empID string) ([]model.OrgAssignment, error) {
	out := []model.OrgAssignment{}
	err := s.db.Select(&out, `SELECT `+orgAssignCols+`, d.name AS "divisionName", u.name AS "unitName"
		FROM "OrgAssignment" a JOIN "Employee" e ON e.id = a."employeeId"
		JOIN "OrgDivision" d ON d.id = a."divisionId" LEFT JOIN "OrgUnit" u ON u.id = a."unitId"
		WHERE a."employeeId" = $1 AND a."endedAt" IS NULL ORDER BY a."startedAt"`, empID)
	return out, err
}

// ── التعيين والإنهاء ──

type OrgAssignInput struct {
	EmployeeID   string  `json:"employeeId"`
	DivisionID   string  `json:"divisionId"`
	UnitID       string  `json:"unitId"`
	Position     string  `json:"position"`
	AllowanceIQD float64 `json:"allowanceIQD"`
	Note         string  `json:"note"`
}

var ErrOrgSeatTaken = errors.New("هالمكان بي مدير/نائب هسه — أنهِ منصبه أول أو عيّن بمكان ثاني")

// Assign يعيّن موظف بمنصب ويمنحه حزمة مكانه.
func (s *OrgService) Assign(actorID string, in OrgAssignInput) (*model.OrgAssignment, error) {
	in.Position = strings.ToUpper(strings.TrimSpace(in.Position))
	if in.Position != model.OrgDirector && in.Position != model.OrgDeputy && in.Position != model.OrgMember {
		return nil, errors.New("المنصب لازم مدير أو نائب أو عضو")
	}
	if in.AllowanceIQD < 0 {
		return nil, errors.New("المخصص ما يكون سالب")
	}
	if err := model.CheckIQD("المخصص", in.AllowanceIQD); err != nil {
		return nil, err
	}
	var unit any
	if in.UnitID != "" {
		var div string
		if err := s.db.Get(&div, `SELECT "divisionId" FROM "OrgUnit" WHERE id = $1`, in.UnitID); err != nil {
			return nil, errors.New("الوحدة مو موجودة")
		}
		if in.DivisionID == "" {
			in.DivisionID = div
		}
		if div != in.DivisionID {
			return nil, errors.New("الوحدة مو تابعة لهالإدارة")
		}
		unit = in.UnitID
	} else {
		var n int
		if err := s.db.Get(&n, `SELECT count(*) FROM "OrgDivision" WHERE id = $1`, in.DivisionID); err != nil || n == 0 {
			return nil, errors.New("الإدارة مو موجودة")
		}
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// قفل الموظف أول — تعيينين بنفس الوكت لنفس الموظف ما يتسابقون على الصلاحيات
	var active string
	if err := tx.Get(&active, `SELECT status FROM "Employee" WHERE id = $1 FOR UPDATE`, in.EmployeeID); err != nil {
		return nil, errors.New("الموظف مو موجود")
	}
	if active != "ACTIVE" {
		return nil, errors.New("الموظف مو فعّال")
	}
	var id string
	err = tx.Get(&id, `INSERT INTO "OrgAssignment" ("employeeId", "divisionId", "unitId", position, "allowanceIQD", note, "createdById")
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, '')) RETURNING id`,
		in.EmployeeID, in.DivisionID, unit, in.Position, in.AllowanceIQD, strings.TrimSpace(in.Note), actorID)
	if err != nil {
		if strings.Contains(err.Error(), "OrgAssignment_head_uq") {
			return nil, ErrOrgSeatTaken
		}
		if strings.Contains(err.Error(), "OrgAssignment_emp_uq") {
			return nil, errors.New("الموظف معيّن بهالمكان أصلاً")
		}
		return nil, err
	}
	if err := syncOrgPermissions(tx, in.EmployeeID, actorID, "تعيين بمنصب"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.byID(id)
}

// End ينهي منصب (ما يحذفه) ويسحب صلاحياته الي جت بسببه بس.
func (s *OrgService) End(actorID, id, reason string) (*model.OrgAssignment, error) {
	reason = strings.TrimSpace(reason)
	if len([]rune(reason)) < 3 {
		return nil, errors.New("اكتب سبب الإنهاء")
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var emp string
	err = tx.Get(&emp, `UPDATE "OrgAssignment" SET "endedAt" = now(), "endedById" = NULLIF($2, ''), "endReason" = $3
		WHERE id = $1 AND "endedAt" IS NULL RETURNING "employeeId"`, id, actorID, reason)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("المنصب مو موجود أو منتهي")
	}
	if err != nil {
		return nil, err
	}
	if err := syncOrgPermissions(tx, emp, actorID, "إنهاء منصب: "+reason); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.byID(id)
}

// Restore يرجّع منصب منتهي (تراجع عن قرار سحب). يفشل إذا مكانه انشغل.
func (s *OrgService) Restore(actorID, id string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var emp string
	err = tx.Get(&emp, `UPDATE "OrgAssignment" SET "endedAt" = NULL, "endedById" = NULL, "endReason" = NULL
		WHERE id = $1 AND "endedAt" IS NOT NULL RETURNING "employeeId"`, id)
	if err != nil {
		if strings.Contains(err.Error(), "OrgAssignment_head_uq") || strings.Contains(err.Error(), "OrgAssignment_emp_uq") {
			return ErrOrgSeatTaken
		}
		return err
	}
	if err := syncOrgPermissions(tx, emp, actorID, "رجوع منصب"); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OrgService) byID(id string) (*model.OrgAssignment, error) {
	var a model.OrgAssignment
	err := s.db.Get(&a, `SELECT `+orgAssignCols+` FROM "OrgAssignment" a JOIN "Employee" e ON e.id = a."employeeId" WHERE a.id = $1`, id)
	return &a, err
}

// ── مزامنة الصلاحيات ──

type orgQuerier interface {
	Select(dest any, q string, args ...any) error
	Exec(q string, args ...any) (sql.Result, error)
}

// OrgDesiredPermissions الصلاحيات الي يستحقها الموظف من مناصبه الفعّالة.
//   - منصب بوحدة: حزمة الوحدة (صلاحيات headOnly لرئيسها/نائبها بس).
//   - مدير/نائب إدارة: حزمة الإدارة + حزم كل وحداتها.
//   - عضو إدارة (بلا وحدة): حزمة الإدارة بس.
func orgDesired(q orgQuerier, empID string) (map[string]bool, error) {
	rows := []struct {
		Perm string `db:"permission"`
	}{}
	err := q.Select(&rows, `
		SELECT DISTINCT b.permission FROM "OrgAssignment" a
		JOIN "OrgBundle" b ON b.scope = a."unitId"
		WHERE a."employeeId" = $1 AND a."endedAt" IS NULL AND a."unitId" IS NOT NULL
		  AND (NOT b."headOnly" OR a.position IN ('DIRECTOR', 'DEPUTY'))
		UNION
		SELECT DISTINCT b.permission FROM "OrgAssignment" a
		JOIN "OrgBundle" b ON b.scope = a."divisionId"
		WHERE a."employeeId" = $1 AND a."endedAt" IS NULL AND a."unitId" IS NULL
		  AND (NOT b."headOnly" OR a.position IN ('DIRECTOR', 'DEPUTY'))
		UNION
		SELECT DISTINCT b.permission FROM "OrgAssignment" a
		JOIN "OrgUnit" u ON u."divisionId" = a."divisionId"
		JOIN "OrgBundle" b ON b.scope = u.id
		WHERE a."employeeId" = $1 AND a."endedAt" IS NULL AND a."unitId" IS NULL AND a.position IN ('DIRECTOR', 'DEPUTY')`, empID)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, r := range rows {
		if _, ownerOnly := model.PermissionOwnerOnly[r.Perm]; !ownerOnly {
			out[r.Perm] = true
		}
	}
	return out, nil
}

// OrgWants هل مناصب الموظف الفعّالة تستحق هالصلاحية؟ (لانتهاء الاستثناءات المؤقتة)
func OrgWants(tx *sqlx.Tx, empID, perm string) bool {
	d, err := orgDesired(tx, empID)
	return err == nil && d[perm]
}

func syncOrgPermissions(tx *sqlx.Tx, empID, actorID, reason string) error {
	if _, err := tx.Exec(`SELECT 1 FROM "Employee" WHERE id = $1 FOR UPDATE`, empID); err != nil {
		return err
	}
	var emp struct {
		Role      string         `db:"role"`
		Secondary pq.StringArray `db:"secondaryRoles"`
	}
	if err := tx.Get(&emp, `SELECT role::text AS role, COALESCE("secondaryRoles", '{}') AS "secondaryRoles" FROM "Employee" WHERE id = $1`, empID); err != nil {
		return err
	}
	tmpl := map[string]bool{}
	for _, rl := range append([]string{emp.Role}, emp.Secondary...) {
		for _, p := range model.RoleDefaultPermissions[rl] {
			tmpl[p] = true
		}
	}
	desired, err := orgDesired(tx, empID)
	if err != nil {
		return err
	}
	type held struct {
		Name    string `db:"name"`
		ViaOrg  bool   `db:"viaOrg"`
		Granted bool   `db:"granted"`
		Pinned  bool   `db:"pinned"`
	}
	cur := []held{}
	if err := tx.Select(&cur, `
		SELECT p.name,
		       EXISTS (SELECT 1 FROM "OrgPermissionGrant" g WHERE g."employeeId" = $1 AND g.permission = p.name) AS "viaOrg",
		       EXISTS (SELECT 1 FROM "EmployeePermission" ep WHERE ep."employeeId" = $1 AND ep."permissionId" = p.id) AS granted,
		       EXISTS (SELECT 1 FROM "PermissionException" x WHERE x."employeeId" = $1 AND x.permission = p.name) AS pinned
		FROM "Permission" p`, empID); err != nil {
		return err
	}
	changed := false
	names := make([]string, 0, len(cur))
	idx := map[string]held{}
	for _, h := range cur {
		idx[h.Name] = h
		names = append(names, h.Name)
	}
	sort.Strings(names)
	audit := func(action, perm string) {
		_, _ = tx.Exec(`INSERT INTO "PermissionAuditLog" ("employeeId", "actorId", action, permission, reason) VALUES ($1, NULLIF($2, ''), $3, $4, $5)`,
			empID, actorID, action, perm, reason)
	}
	for _, n := range names {
		h := idx[n]
		switch {
		case desired[n] && !h.Granted && !h.ViaOrg:
			// يستحقها ومو عنده — ننطيها ونسجّل إنها من المنصب.
			// (ViaOrg ويا مو عنده = الأدمن شالها بيده: نحترم قراره وما نرجّعها)
			if _, err := tx.Exec(`INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
				SELECT gen_random_uuid()::text, $1, id FROM "Permission" WHERE name = $2 ON CONFLICT DO NOTHING`, empID, n); err != nil {
				return err
			}
			if _, err := tx.Exec(`INSERT INTO "OrgPermissionGrant" ("employeeId", permission) VALUES ($1, $2) ON CONFLICT DO NOTHING`, empID, n); err != nil {
				return err
			}
			audit("ORG_GRANT", n)
			changed = true
		case !desired[n] && h.ViaOrg:
			// انطت بسبب منصب وهسه ما يستحقها — تنسحب (اليدوي ما ينمس لأنه مو مسجّل ORG).
			// إلا إذا دوره يستحقها (قالب) أو عليها استثناء مثبّت — نشيل علامة المنصب بس.
			if h.Granted && !tmpl[n] && !h.Pinned {
				if _, err := tx.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1
					AND "permissionId" = (SELECT id FROM "Permission" WHERE name = $2)`, empID, n); err != nil {
					return err
				}
				audit("ORG_REVOKE", n)
				changed = true
			}
			if _, err := tx.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = $1 AND permission = $2`, empID, n); err != nil {
				return err
			}
		}
	}
	if changed {
		if _, err := tx.Exec(`UPDATE "Employee" SET "accessVersion" = "accessVersion" + 1 WHERE id = $1`, empID); err != nil {
			return err
		}
	}
	return nil
}

// TargetRole دور موظف (لحارس «مناصب المالك والمدير يعيّنها المالك بس»).
func (s *OrgService) TargetRole(empID string) string {
	var role string
	_ = s.db.Get(&role, `SELECT role::text FROM "Employee" WHERE id = $1`, empID)
	return role
}

// AssignmentEmployee صاحب المنصب.
func (s *OrgService) AssignmentEmployee(id string) string {
	var emp string
	_ = s.db.Get(&emp, `SELECT "employeeId" FROM "OrgAssignment" WHERE id = $1`, id)
	return emp
}

// DB للأدوات المشتركة (رأي كمال التنفيذي).
func (s *OrgService) DB() *sqlx.DB { return s.db }
