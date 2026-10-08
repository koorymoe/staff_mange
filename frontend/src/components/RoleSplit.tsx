import { useMemo, useState, type ReactNode } from 'react'
import { roleChipColor, roleGradient, roleLabel } from '../roleLabels'

// ═══ «كل دور ينفصل عن الثاني» — قرار (ع) 10-07 ═══
// تبويب لكل دور موجود بالبيانات (الأكثر موظفين أول)، ويعرض موظفين الدور
// المختار بس. نفس الشكل بالإحصائيات اليومية والأسبوعية والشهرية.
export default function RoleSplit<T extends { role: string }>({ items, children }: { items: T[]; children: (list: T[], role: string) => ReactNode }) {
  const [picked, setPicked] = useState('')
  const { role, list, tabs } = useRoleTabs(items, picked, setPicked)
  if (items.length === 0) return null
  return (
    <div className="space-y-4">
      {tabs}
      {children(list, role)}
    </div>
  )
}

/** نفس تبويبات الأدوار، للصفحات الي عندها جدول جاهز وتريد بس تفلتره. */
// eslint-disable-next-line react-refresh/only-export-components
export function useRoleTabs<T extends { role: string }>(items: T[], picked: string, setPicked: (r: string) => void) {
  const roles = useMemo(() => {
    const m = new Map<string, number>()
    items.forEach((r) => m.set(r.role, (m.get(r.role) ?? 0) + 1))
    return [...m.entries()].sort((a, b) => b[1] - a[1]).map(([role, n]) => ({ role, n }))
  }, [items])
  const role = roles.some((r) => r.role === picked) ? picked : (roles[0]?.role ?? '')
  const list = useMemo(() => items.filter((r) => r.role === role), [items, role])
  const tabs = (
    <div className="flex flex-wrap gap-2">
      {roles.map(({ role: r, n }) => {
        const on = r === role
        const c = roleChipColor(r)
        return (
          <button key={r} type="button" onClick={() => setPicked(r)}
            className={`flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-bold shadow-sm transition ${on ? `bg-gradient-to-l ${roleGradient(r)} text-white` : `${c.bg} ${c.text} hover:brightness-95`}`}>
            {roleLabel(r)}
            <span className={`rounded-full px-2 text-xs tabular-nums ${on ? 'bg-white/25' : 'bg-white'}`}>{n}</span>
          </button>
        )
      })}
    </div>
  )
  return { role, list, tabs }
}

/** رقم صغير بعنوان — نفس شكل بطاقات الإحصائيات. */
export function MiniMetric({ label, value, tone, hint }: { label: string; value: ReactNode; tone?: 'good' | 'mid' | 'bad'; hint?: string }) {
  const t = tone === 'good' ? 'text-emerald-700' : tone === 'mid' ? 'text-amber-700' : tone === 'bad' ? 'text-red-700' : 'text-slate-800'
  return (
    <div className="rounded-lg bg-slate-50 px-3 py-2" title={hint}>
      <p className="text-[11px] text-slate-500">{label}</p>
      <p className={`text-sm font-bold tabular-nums ${t}`}>{value}</p>
    </div>
  )
}

/** بطاقة رقم كبير للملخص. */
export function SummaryCard({ label, value, tone }: { label: string; value: ReactNode; tone?: 'good' | 'mid' | 'bad' }) {
  const t = tone === 'good' ? 'text-emerald-700' : tone === 'mid' ? 'text-amber-700' : tone === 'bad' ? 'text-red-700' : 'text-slate-800'
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3 shadow-sm">
      <p className="text-xs text-slate-500">{label}</p>
      <p className={`mt-1 text-lg font-extrabold tabular-nums ${t}`}>{value}</p>
    </div>
  )
}
