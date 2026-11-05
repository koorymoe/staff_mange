import { type OrgDivision } from '../../org'
import { type DeskPlan, type DeskIssue, type DeskItem } from '../../orgDesk'

export type Tab = 'mine' | 'issues' | 'plan' | 'units' | 'authority' | 'team' | 'directives' | 'outputs'

export const canAddItems = (plan: DeskPlan) =>
  plan.status === 'DRAFT' || plan.status === 'RETURNED'

export const dueOrder = (a: DeskItem, b: DeskItem) =>
  !a.dueDate
    ? b.dueDate
      ? 1
      : 0
    : !b.dueDate
      ? -1
      : a.dueDate.localeCompare(b.dueDate)

export const teamPeople = (division: OrgDivision) => [
  ...new Map(
    [
      ...(division.people ?? []),
      ...(division.units ?? []).flatMap((u) => u.people ?? []),
    ].map((p) => [p.employeeId, p]),
  ).values(),
]

export const number = (n: number | null | undefined) =>
  n == null
    ? 'غير متوفر'
    : new Intl.NumberFormat('ar-IQ', { maximumFractionDigits: 2 }).format(n)

export const percent = (n: number) =>
  Math.min(100, Math.max(0, Number.isFinite(n) ? n : 0))

export const statuses = {
  OPEN: 'مفتوحة',
  IN_PROGRESS: 'قيد المعالجة',
  RESOLVED: 'انحلت',
}

export const date = (value: string) => {
  const d = new Date(value)
  return Number.isNaN(d.getTime())
    ? 'التاريخ مو متوفر'
    : new Intl.DateTimeFormat('ar-IQ', {
        timeZone: 'Asia/Baghdad',
        dateStyle: 'medium',
        timeStyle: 'short',
      }).format(d)
}

export type Modal =
  | { kind: 'issue' }
  | { kind: 'status'; issue: DeskIssue }
  | { kind: 'plan' }
  | { kind: 'item' }
  | { kind: 'authority' }
  | { kind: 'solution'; issue: DeskIssue }

export const planStatusNames = {
  DRAFT: 'مسودة',
  SUBMITTED: 'مرفوعة',
  APPROVED: 'معتمدة',
  RETURNED: 'مرجعة للتعديل',
}

export const field = (data: FormData, key: string) =>
  String(data.get(key) ?? '').trim()
