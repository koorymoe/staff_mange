package model

import (
	"fmt"
	"sort"
	"strings"
)

// ═══ الدور من المنصب — دازاي 11-01 ═══
// «الهيكلة الجديدة بيها نفس الشغل القديم؟ المراقب من أنطيه إدارة يطلع شغله؟».
// الصلاحيات تنعطي بالحزم، بس أكو شغل بالنظام مربوط باسم الدور (محاسب، إداري
// كوادر…). فكل وحدة تعطي صاحبها **دوراً فعلياً** فوك دوره الأصلي — نفس فكرة
// monitoring ← MONITOR القديمة. الدور الأصلي ما يتغيّر، وADMIN/OWNER ما ينعطون أبداً.
//
// مدير/نائب الإدارة ياخذ أدوار كل وحداتها.

// OrgUnitRole الوحدة ← الدور الفعلي لكل أعضائها.
var OrgUnitRole = map[string]string{
	"admin.monitoring":   "MONITOR",
	"admin.accounts":     "FINANCE",
	"ops.staff":          "HR_COORDINATOR",
	"quality.followup":   "QUALITY_ENGINEER",
	"supply.procurement": "PROCUREMENT_ADMIN",
	"eng.projects":       "PROJECT_MANAGER",
}

// OrgUnitHeadRole الوحدة ← دور فعلي لرئيسها/نائبها بس.
var OrgUnitHeadRole = map[string]string{
	"bizdev.tech": "GPS_ADMIN",
}

// orgRoleValues جدول VALUES (unit, role, head) للاستعلامات.
func orgRoleValues() string {
	rows := []string{}
	add := func(m map[string]string, head bool) {
		for u, r := range m {
			rows = append(rows, fmt.Sprintf("('%s', '%s', %t)", u, r, head))
		}
	}
	add(OrgUnitRole, false)
	add(OrgUnitHeadRole, true)
	sort.Strings(rows)
	return strings.Join(rows, ", ")
}

// OrgRoleEmployeesSQL استعلام فرعي: معرّفات الموظفين الي إلهم دور فعلي من منصبهم
// ضمن rolesExpr (تعبير SQL لمصفوفة نصوص، مثل `$1` أو `ARRAY['FINANCE']`).
func OrgRoleEmployeesSQL(rolesExpr string) string {
	return `(SELECT oa."employeeId" FROM "OrgAssignment" oa
		JOIN (VALUES ` + orgRoleValues() + `) AS om(unit, role, head) ON
			(om.unit = oa."unitId" AND (NOT om.head OR oa.position IN ('DIRECTOR', 'DEPUTY')))
			OR (oa."unitId" IS NULL AND oa.position IN ('DIRECTOR', 'DEPUTY')
				AND om.unit IN (SELECT ou.id FROM "OrgUnit" ou WHERE ou."divisionId" = oa."divisionId"))
		WHERE oa."endedAt" IS NULL AND om.role = ANY(` + rolesExpr + `))`
}

// OrgRolesOfEmployeeSQL الأدوار الفعلية لموظف واحد ($1).
func OrgRolesOfEmployeeSQL() string {
	return `SELECT DISTINCT om.role FROM "OrgAssignment" oa
		JOIN (VALUES ` + orgRoleValues() + `) AS om(unit, role, head) ON
			(om.unit = oa."unitId" AND (NOT om.head OR oa.position IN ('DIRECTOR', 'DEPUTY')))
			OR (oa."unitId" IS NULL AND oa.position IN ('DIRECTOR', 'DEPUTY')
				AND om.unit IN (SELECT ou.id FROM "OrgUnit" ou WHERE ou."divisionId" = oa."divisionId"))
		WHERE oa."endedAt" IS NULL AND oa."employeeId" = $1 ORDER BY om.role`
}

// OrgRolePriority ترتيب الدور الفعلي الواحد (EffectiveRole) لمن عنده أكثر من دور.
var OrgRolePriority = []string{"MONITOR", "FINANCE", "HR_COORDINATOR", "PROJECT_MANAGER", "PROCUREMENT_ADMIN", "QUALITY_ENGINEER", "GPS_ADMIN"}
