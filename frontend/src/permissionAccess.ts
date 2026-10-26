import { apiRequest } from './api'

export type AccessPermission = {
  name: string
  label: string
  grantable: boolean
  lockReason?: string
}
export type AccessGroup = {
  key: string
  title: string
  bundle?: string
  perms: AccessPermission[]
}
export type AccessItem = {
  name: string
  source: 'TEMPLATE' | 'MANUAL' | 'EXCEPTION'
  expiresAt?: string
  reason?: string
}
export type AccessState = {
  employee: {
    id: string
    name: string
    role: string
    secondaryRoles: string[]
    isLeader: boolean
  }
  perms: AccessItem[]
  missingFromTemplate: string[]
  version: number
}
export type RoleDrift = {
  employeeId: string
  name: string
  role: string
  status: 'MATCH' | 'DIFF' | 'NO_TEMPLATE'
  extra: string[]
  missing: string[]
  exceptions: string[]
  version: number
}
export type AccessAudit = {
  id: string
  action: 'ADD' | 'REMOVE' | 'EXPIRE'
  permission: string
  actorName: string | null
  reason: string | null
  createdAt: string
}
export type AccessAgent = {
  agent: 'MATRIX' | 'IRIS' | 'QUALITY' | 'BOOKING'
  name: string
  editable: boolean
}
const employeePath = (id: string) =>
  `/employees/${encodeURIComponent(id)}/access`
export const accessApi = {
  catalog: () => apiRequest<AccessGroup[]>('/permissions/catalog'),
  get: (id: string) => apiRequest<AccessState>(employeePath(id)),
  save: (
    id: string,
    body: { add: string[]; remove: string[]; reason: string; version: number },
  ) =>
    apiRequest<AccessState>(employeePath(id), {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  drift: () => apiRequest<RoleDrift[]>('/permissions/role-drift'),
  align: (id: string, body: { reason: string; version: number }) =>
    apiRequest<RoleDrift>(
      `/permissions/role-drift/${encodeURIComponent(id)}/align`,
      { method: 'POST', body: JSON.stringify(body) },
    ),
  exception: (body: {
    employeeId: string
    permission: string
    reason: string
    expiresAt?: string
    version: number
  }) =>
    apiRequest('/permissions/exceptions', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  removeException: (employeeId: string, permission: string) =>
    apiRequest(
      `/permissions/exceptions?${new URLSearchParams({ employeeId, permission })}`,
      { method: 'DELETE' },
    ),
  audit: (employeeId: string) =>
    apiRequest<AccessAudit[]>(
      `/permissions/audit-log?${new URLSearchParams({ employeeId })}`,
    ),
  agents: () => apiRequest<AccessAgent[]>('/ai/agents'),
  renameAgent: (agent: string, name: string) =>
    apiRequest(`/ai/agents/${encodeURIComponent(agent)}`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    }),
}
