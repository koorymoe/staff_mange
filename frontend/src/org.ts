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
  people: OrgPerson[]
}
export type OrgDivision = {
  id: string
  name: string
  parentId: string | null
  summary: string
  permissions: string[]
  people: OrgPerson[]
  units: OrgUnit[]
}
export const orgPositions: Record<OrgPosition, string> = {
  DIRECTOR: 'مدير',
  DEPUTY: 'نائب',
  MEMBER: 'عضو',
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
