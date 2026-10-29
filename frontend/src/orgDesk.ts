import { apiRequest } from './api'
import type { OrgDivision } from './org'
export const orgPrinciples = [
  'دقة العمل',
  'استغلال الكوادر',
  'الحفاظ على العُدد',
  'جلب الحجوزات',
  'الإنجاز الآني',
  'السرعة',
  'نظافة الأدوات والمركبات',
  'عدم التأخر بالحجوزات',
  'متابعة الحجوزات والاتصالات',
  'المعايير الهندسية',
  'الاحترام المتبادل',
  'نظام المؤسسات',
  'تكلفة الكوادر',
  'الهيكلية التنفيذية',
  'الالتزام بالدوام',
  'المسؤولية والمحاسبة',
  'توجيهات المدير',
  'سرية العمل',
  'خلية نحل',
  'المناصب المؤقتة بالإجازات',
]
export type DeskIssue = {
  id: string
  title: string
  problem: string
  impact: string
  scope: 'SPECIFIC' | 'GENERAL'
  principle: number | null
  severity: 'LOW' | 'MEDIUM' | 'HIGH'
  status: 'OPEN' | 'IN_PROGRESS' | 'RESOLVED'
  resolution: string | null
  authorName: string | null
  unitId: string | null
  createdAt: string
}
export type DeskItem = {
  id: string
  kind: 'PROBLEM' | 'DEVELOPMENT'
  title: string
  issueId: string | null
  metric: string | null
  baseline: number | null
  target: number | null
  current: number | null
  progress: number
  byMatrix: boolean
  manualProgress: number
  doneAt: string | null
}
export type DeskComment = {
  id?: string
  targetType: 'ISSUE' | 'PLAN_ITEM'
  targetId: string
  authorName: string | null
  text: string
  createdAt: string
}
export type AuthorityRow = {
  decision: string
  mode: 'DECIDE' | 'ESCALATE'
  limitIQD: number | null
}
export type DeskMetric = {
  key: string
  label: string
  lowerBetter: boolean
  unit: string
}
export type OrgDesk = {
  division: OrgDivision
  period: string
  canEdit: boolean
  issues: DeskIssue[]
  plan: null | {
    id: string
    title: string
    period: string
    progress: number
    items: DeskItem[]
  }
  comments: DeskComment[]
  authority: AuthorityRow[]
  metrics: DeskMetric[]
}
export const orgDeskGet = (id: string, period: string) =>
  apiRequest<OrgDesk>(
    `/org/desk/${encodeURIComponent(id)}?${new URLSearchParams({ period })}`,
  )
export const orgDeskWrite = (path: string, body: unknown, method = 'POST') =>
  apiRequest(`/org/${path}`, { method, body: JSON.stringify(body) })
export const orgMonth = () => {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Baghdad',
    year: 'numeric',
    month: '2-digit',
  }).formatToParts(new Date())
  return `${parts.find((p) => p.type === 'year')!.value}-${parts.find((p) => p.type === 'month')!.value}`
}
export const orgIQD = (raw: string) => {
  const text = raw
    .replace(/[٠-٩]/g, (d) => String(d.charCodeAt(0) - 1632))
    .replace(/[۰-۹]/g, (d) => String(d.charCodeAt(0) - 1776))
    .replace(/[,٬\s]/g, '')
  const value = Number(text)
  return {
    value,
    valid: /^\d+$/.test(text) && Number.isSafeInteger(value),
    small: value > 0 && value < 1000,
  }
}
