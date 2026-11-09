import { ApiError } from '../../api'
import { type AccessAgent } from '../../permissionAccess'

export const sourceLabel = {
  TEMPLATE: 'قالب المنصب',
  MANUAL: 'منح يدوي',
  EXCEPTION: 'استثناء معتمد',
  ORG: 'منصب بالهيكل',
}

export const tiers = ['quotation_create', 'quotation_edit_own', 'quotation_manage_all']

export function exceptionDeadline(value: string) {
  if (!value) return null
  const at = new Date(`${value}:00+03:00`)
  if (!Number.isFinite(at.getTime()) || at.getTime() <= Date.now()) {
    throw new Error('موعد الانتهاء لازم بالمستقبل، بتوقيت بغداد.')
  }
  return at
}

export function baghdadInput(at: string) {
  const parts = Object.fromEntries(
    new Intl.DateTimeFormat('en-GB', {
      timeZone: 'Asia/Baghdad',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hourCycle: 'h23',
    })
      .formatToParts(new Date(at))
      .map((part) => [part.type, part.value]),
  )
  return `${parts.year}-${parts.month}-${parts.day}T${parts.hour}:${parts.minute}`
}

export const agentPermission = {
  MATRIX: 'agent_matrix',
  QUALITY: 'agent_quality',
  BOOKING: 'agent_booking',
}

export const defaults: AccessAgent[] = [
  { agent: 'MATRIX', name: 'ماتركس', editable: false },
  { agent: 'IRIS', name: 'ايريس', editable: false },
  { agent: 'QUALITY', name: 'وكيل الجودة', editable: false },
  { agent: 'BOOKING', name: 'وكيل الحجوزات', editable: false },
]

export const time = (at: string) =>
  new Intl.DateTimeFormat('ar-IQ', {
    timeZone: 'Asia/Baghdad',
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(at))

export const conflict = (e: unknown) => e instanceof ApiError && e.status === 409

export const message = (e: unknown) =>
  e instanceof Error ? e.message : 'تعذر إكمال العملية. جرّب مرة ثانية.'
