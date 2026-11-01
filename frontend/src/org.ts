import { apiRequest } from './api'

export type OrgPosition = 'DIRECTOR' | 'DEPUTY' | 'MEMBER'
export type OrgPerson = {
  id: string
  employeeId: string
  employeeName: string
  divisionId: string
  unitId: string | null
  position: OrgPosition
  allowanceIQD: number
  acting: boolean
  startedAt: string
  note: string | null
  divisionName?: string
  unitName?: string | null
}
export type OrgUnit = {
  id: string
  divisionId: string
  name: string
  permissions: string[]
  headPermissions: string[]
  people: OrgPerson[]
}
export type OrgDivision = {
  id: string
  name: string
  parentId: string | null
  summary: string
  permissions: string[]
  headPermissions: string[]
  people: OrgPerson[]
  units: OrgUnit[]
}
export const orgPositions: Record<OrgPosition, string> = {
  DIRECTOR: 'مدير',
  DEPUTY: 'نائب',
  MEMBER: 'عضو',
}

/** Preview this assignment only, not the employee's effective/manual grants.
 * Missing head metadata must never silently grant the full bundle to a member.
 */
export function orgAssignmentPermissions(
  division: OrgDivision | undefined,
  unitId: string,
  position: OrgPosition,
): string[] | null {
  if (!division) return null
  const unit = unitId ? division.units?.find((u) => u.id === unitId) : undefined
  if (unitId && !unit) return null
  const scope = unit ?? division
  if (!Array.isArray(scope.permissions)) return null
  // (دازاي 11-01) صلاحيات الرئيس (headOnly) للمدير بس؛ النائب مثل العضو لحد الوكالة
  const memberOf = (sc: { permissions: string[]; headPermissions?: string[] }) => {
    if (!Array.isArray(sc.headPermissions)) return null
    const headOnly = new Set(sc.headPermissions)
    return sc.permissions.filter((name) => !headOnly.has(name))
  }
  let permissions: string[] | null = position === 'DIRECTOR' ? scope.permissions : memberOf(scope)
  if (!permissions) return null
  if (!unit && position !== 'MEMBER') {
    if (
      !Array.isArray(division.units) ||
      division.units.some((u) => !Array.isArray(u.permissions))
    )
      return null
    for (const u of division.units) {
      const p = position === 'DIRECTOR' ? u.permissions : memberOf(u)
      if (!p) return null
      permissions = [...permissions, ...p]
    }
  }
  return [...new Set(permissions)].sort()
}
export const orgPeople = (chart: OrgDivision[]) =>
  chart.flatMap((d) => [
    ...(d.people ?? []).filter((p) => !p.unitId),
    ...(d.units ?? []).flatMap((u) => u.people ?? []),
  ])
export const orgApi = {
  chart: () => apiRequest<OrgDivision[]>('/org/chart'),
  me: () => apiRequest<OrgPerson[]>('/org/me'),
  assign: (body: {
    employeeId: string
    divisionId: string
    unitId?: string
    position: OrgPosition
    allowanceIQD: number
    note?: string
  }) =>
    apiRequest('/org/assign', { method: 'POST', body: JSON.stringify(body) }),
  end: (id: string, reason: string) =>
    apiRequest(`/org/assignments/${encodeURIComponent(id)}/end`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    }),
}
