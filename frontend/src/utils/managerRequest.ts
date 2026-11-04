import type { ManagerRequest } from '../api'

// (ع) 11-04: تسميات طلبات المدير — مشتركة بين نافذة الرفع وصفحة المدير
export const PRIORITY: Record<ManagerRequest['priority'], { label: string; cls: string }> = {
  NORMAL: { label: 'عادي', cls: 'bg-slate-100 text-slate-700' },
  HIGH: { label: 'مهم', cls: 'bg-amber-100 text-amber-800' },
  URGENT: { label: '🚨 عاجل', cls: 'bg-red-100 text-red-700' },
}
export const STATUS: Record<ManagerRequest['status'], { label: string; cls: string }> = {
  OPEN: { label: 'ينتظر المدير', cls: 'bg-sky-100 text-sky-800' },
  DONE: { label: '✅ انعمل', cls: 'bg-emerald-100 text-emerald-800' },
  REJECTED: { label: '❌ انرفض', cls: 'bg-red-100 text-red-700' },
}
