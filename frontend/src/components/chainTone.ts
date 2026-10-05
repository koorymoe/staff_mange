import type { ChainStatus } from '../api'

export const CHAIN_TONE: Record<ChainStatus, { icon: string; cls: string; label: string }> = {
  OK: { icon: '✅', cls: 'border-emerald-200 bg-emerald-50', label: 'بوقتها' },
  ISSUE: { icon: '⚠️', cls: 'border-amber-200 bg-amber-50', label: 'بيها نقص' },
  LATE: { icon: '🟠', cls: 'border-orange-300 bg-orange-50', label: 'متأخرة' },
  MISSED: { icon: '🔴', cls: 'border-red-300 bg-red-50', label: 'ما انسوّت' },
  WAITING: { icon: '⏳', cls: 'border-sky-200 bg-sky-50', label: 'تنتظر' },
  NA: { icon: '➖', cls: 'border-slate-200 bg-slate-50 opacity-70', label: 'ما تنطبق' },
}
