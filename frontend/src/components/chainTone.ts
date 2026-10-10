import type { ChainStatus } from '../api'

export const CHAIN_TONE: Record<ChainStatus, { icon: string; cls: string; label: string }> = {
  OK: { icon: '✅', cls: 'border-emerald-200 bg-emerald-50', label: 'بوقتها' },
  ISSUE: { icon: '⚠️', cls: 'border-amber-200 bg-amber-50', label: 'بيها نقص' },
  LATE: { icon: '🟠', cls: 'border-orange-300 bg-orange-50', label: 'متأخرة' },
  MISSED: { icon: '🔴', cls: 'border-red-300 bg-red-50', label: 'ما صارت' },
  WAITING: { icon: '⏳', cls: 'border-sky-200 bg-sky-50', label: 'تنتظر' },
  NA: { icon: '➖', cls: 'border-slate-200 bg-slate-50 opacity-70', label: 'ما تنطبق' },
}

// ألوان الشريط المقسّم (بوقتها / بيها نقص / متأخرة / ما صارت).
export const CHAIN_BAR = { ok: '#10b981', issue: '#f59e0b', late: '#fb923c', missed: '#ef4444' }

// المدة بلغة مفهومة: «فوراً»، «٤٠ دقيقة»، «٣ ساعات و١٠ دقايق»، «يومين».
export function chainDuration(m: number | null | undefined): string {
  if (m == null) return 'ماكو بيانات'
  if (m <= 0) return 'فوراً'
  if (m < 60) return `${m} دقيقة`
  if (m < 1440) {
    const h = Math.floor(m / 60), r = m % 60
    return r ? `${h} ساعة و${r} دقيقة` : `${h} ساعة`
  }
  const d = Math.floor(m / 1440), h = Math.floor((m % 1440) / 60)
  return h ? `${d} يوم و${h} ساعة` : `${d} يوم`
}
