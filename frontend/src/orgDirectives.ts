import { apiRequest } from './api'
export type DirectiveScope = 'COMPANY' | 'DIVISION' | 'UNIT'
export type DirectiveTarget = {
  scope: DirectiveScope
  divisionId?: string
  unitId?: string
  label: string
}
export type OrgDirective = {
  id: string
  scope: DirectiveScope
  divisionId: string | null
  unitId: string | null
  place: string
  title: string
  body: string
  clause: number | null
  clauseCode: string | null
  issuedBy: string | null
  issuedAt: string
  archivedAt: string | null
  ackAt: string | null
  acks: number
  audience: number
}
export const directivesApi = {
  mine: () => apiRequest<OrgDirective[]>('/org/directives/mine'),
  managed: () => apiRequest<OrgDirective[]>('/org/directives'),
  issue: (body: {
    scope: DirectiveScope
    divisionId?: string
    unitId?: string
    title: string
    body: string
    clause: number | null
  }) =>
    apiRequest<{ id: string; audience: number }>('/org/directives', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  ack: (id: string) =>
    apiRequest(`/org/directives/${encodeURIComponent(id)}/ack`, {
      method: 'POST',
    }),
  archive: (id: string) =>
    apiRequest(`/org/directives/${encodeURIComponent(id)}/archive`, {
      method: 'POST',
    }),
}
export const directiveDate = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime())
    ? 'التاريخ مو متوفر'
    : new Intl.DateTimeFormat('ar-IQ', {
        timeZone: 'Asia/Baghdad',
        dateStyle: 'medium',
        timeStyle: 'short',
      }).format(date)
}
export const directiveTargetKey = (target: DirectiveTarget) =>
  `${target.scope}:${target.divisionId || ''}:${target.unitId || ''}`
