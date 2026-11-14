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
	"eng.design":         "DESIGNER",
	"contracts.media":    "MEDIA",
	"it.main":            "IT_SUPPORT",
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
			(om.unit = oa."unitId" AND (NOT om.head OR oa.position = 'DIRECTOR'))
			OR (oa."unitId" IS NULL AND oa.position IN ('DIRECTOR', 'DEPUTY') AND (NOT om.head OR oa.position = 'DIRECTOR')
				AND om.unit IN (SELECT ou.id FROM "OrgUnit" ou WHERE ou."divisionId" = oa."divisionId"))
		WHERE oa."endedAt" IS NULL AND NOT EXISTS (SELECT 1 FROM "PermissionRevoke" rv WHERE rv."employeeId" = oa."employeeId" AND rv.permission = 'role:' || om.role) AND om.role = ANY(` + rolesExpr + `))`
}

// OrgRolesOfEmployeeSQL الأدوار الفعلية لموظف واحد ($1).
func OrgRolesOfEmployeeSQL() string {
	return `SELECT DISTINCT om.role FROM "OrgAssignment" oa
		JOIN (VALUES ` + orgRoleValues() + `) AS om(unit, role, head) ON
			(om.unit = oa."unitId" AND (NOT om.head OR oa.position = 'DIRECTOR'))
			OR (oa."unitId" IS NULL AND oa.position IN ('DIRECTOR', 'DEPUTY') AND (NOT om.head OR oa.position = 'DIRECTOR')
				AND om.unit IN (SELECT ou.id FROM "OrgUnit" ou WHERE ou."divisionId" = oa."divisionId"))
		WHERE oa."endedAt" IS NULL AND NOT EXISTS (SELECT 1 FROM "PermissionRevoke" rv WHERE rv."employeeId" = oa."employeeId" AND rv.permission = 'role:' || om.role) AND oa."employeeId" = $1 ORDER BY om.role`
}

// RoleRevokeKey (دازاي 11-15) «شلت منها الصلاحيات وبعدهن يرجعن»: الدور الجاي من المنصب (مراقب،
// محاسب…) چان ما ينشال أبداً. هسه المدير يكدر يشيله بـPermissionRevoke(permission = "role:<الدور>")
// وكل الاستعلامات فوك وتحت تتجاوزه.
func RoleRevokeKey(role string) string { return "role:" + role }

// OrgRolePriority ترتيب الدور الفعلي الواحد (EffectiveRole) لمن عنده أكثر من دور.
var OrgRolePriority = []string{"MONITOR", "FINANCE", "HR_COORDINATOR", "PROJECT_MANAGER", "PROCUREMENT_ADMIN", "QUALITY_ENGINEER", "GPS_ADMIN", "MEDIA", "IT_SUPPORT", "DESIGNER"}

// EffectiveRoleSQL تعبير SQL للدور الفعلي: دور المنصب حسب OrgRolePriority، وإلا الدور
// الأصلي (ADMIN/OWNER ما يتغيرون) — لماتركس (التقييم والتذكيرات والواجبات) حتى يشوف
// الموظف بمنصبه (دازاي 11-01).
func EffectiveRoleSQL(idExpr, roleExpr string) string {
	prio := "ARRAY['" + strings.Join(OrgRolePriority, "','") + "']"
	return `CASE WHEN ` + roleExpr + `::text IN ('ADMIN', 'OWNER') THEN ` + roleExpr + `::text ELSE COALESCE((
		SELECT om.role FROM "OrgAssignment" oa
		JOIN (VALUES ` + orgRoleValues() + `) AS om(unit, role, head) ON
			(om.unit = oa."unitId" AND (NOT om.head OR oa.position = 'DIRECTOR'))
			OR (oa."unitId" IS NULL AND oa.position IN ('DIRECTOR', 'DEPUTY') AND (NOT om.head OR oa.position = 'DIRECTOR')
				AND om.unit IN (SELECT ou.id FROM "OrgUnit" ou WHERE ou."divisionId" = oa."divisionId"))
		WHERE oa."endedAt" IS NULL AND NOT EXISTS (SELECT 1 FROM "PermissionRevoke" rv WHERE rv."employeeId" = oa."employeeId" AND rv.permission = 'role:' || om.role) AND oa."employeeId" = ` + idExpr + `
		ORDER BY array_position(` + prio + `, om.role) LIMIT 1), ` + roleExpr + `::text) END`
}

// CoordRaterSQL شرط «هذا المقيّم إداري، مو فني» لواجب تقييم الليدر (COORD_LEADER).
// (دازاي 11-12/11-13) الفني الي سجّل حجز لا يتطالب بتقييم الليدر ولا ينحسب عليه
// إهمال — تقييم الكادر شغل الليدر، والليدر يقيّمه المثبّت الإداري بس.
func CoordRaterSQL(idExpr string) string {
	return `EXISTS (SELECT 1 FROM "Employee" ce WHERE ce.id = ` + idExpr + ` AND NOT ce."isLeader"
		AND (` + EffectiveRoleSQL("ce.id", "ce.role") + `) <> 'TECHNICIAN')`
}

// OrgRoleSourcesSQL (دازاي 11-15 «منين جاية») كل دور جاي من منصب لموظف ($1)، ويا المنصب
// الي جابه، وهل المدير شاله (revoked) — بدون فلتر الشيل حتى يبين بالشاشة.
func OrgRoleSourcesSQL() string {
	return `SELECT DISTINCT om.role, oa.id AS "assignmentId", d.name AS "divisionName", COALESCE(u.name, '') AS "unitName",
			oa.position, oa.acting, oa."startedAt",
			EXISTS (SELECT 1 FROM "PermissionRevoke" rv WHERE rv."employeeId" = oa."employeeId" AND rv.permission = 'role:' || om.role) AS revoked
		FROM "OrgAssignment" oa
		JOIN (VALUES ` + orgRoleValues() + `) AS om(unit, role, head) ON
			(om.unit = oa."unitId" AND (NOT om.head OR oa.position = 'DIRECTOR'))
			OR (oa."unitId" IS NULL AND oa.position IN ('DIRECTOR', 'DEPUTY') AND (NOT om.head OR oa.position = 'DIRECTOR')
				AND om.unit IN (SELECT ou.id FROM "OrgUnit" ou WHERE ou."divisionId" = oa."divisionId"))
		JOIN "OrgDivision" d ON d.id = oa."divisionId"
		LEFT JOIN "OrgUnit" u ON u.id = oa."unitId"
		WHERE oa."endedAt" IS NULL AND oa."employeeId" = $1 ORDER BY om.role`
}
