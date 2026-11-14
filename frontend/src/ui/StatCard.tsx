import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import PctWhy, { type PctParts } from './PctWhy'

// ═══ (دازاي 11-15) بطاقة الرقم الموحّدة ═══
// منسوخة من `Stat` بالرئيسية الإدارية (أغنى نسخة) — چانت ١٤ نسخة بالنظام.
// لون خفيف + أيقونة بدائرة + الرقم كبير، وإذا بيها `to`/`onClick` تصير قابلة للضغط.
// ⚠️ النسبة (pct) دائماً تنلف بـPctWhy إذا انطى `why` — قاعدة البيت.
export type StatTone = 'slate' | 'green' | 'amber' | 'red' | 'blue' | 'violet'
export type StatWhy = { title: string; parts: PctParts; meaning: ReactNode; source: ReactNode }

const TONE: Record<StatTone, { card: string; ink: string; chip: string }> = {
  slate: { card: 'border-slate-200 bg-slate-50/70', ink: 'text-slate-800', chip: 'bg-white text-slate-500' },
  blue: { card: 'border-sky-100 bg-sky-50/70', ink: 'text-sky-800', chip: 'bg-sky-100 text-sky-600' },
  green: { card: 'border-emerald-100 bg-emerald-50/70', ink: 'text-emerald-700', chip: 'bg-emerald-100 text-emerald-600' },
  amber: { card: 'border-amber-100 bg-amber-50/70', ink: 'text-amber-700', chip: 'bg-amber-100 text-amber-600' },
  red: { card: 'border-red-100 bg-red-50/70', ink: 'text-red-700', chip: 'bg-red-100 text-red-600' },
  violet: { card: 'border-violet-100 bg-violet-50/70', ink: 'text-violet-700', chip: 'bg-violet-100 text-violet-600' },
}

export default function StatCard({
  label, value, hint, tone = 'slate', to, icon, big, pct, why, pctLabel, cta, onClick, active, className = '',
}: {
  label: ReactNode
  value: ReactNode
  hint?: ReactNode
  tone?: StatTone
  to?: string
  icon?: ReactNode
  big?: boolean
  pct?: number | null
  why?: StatWhy
  // (دازاي 11-15) نسبة بشرح مخصّص (PctWhy بـformula مثلاً) — تنحط بمكان الرقم %
  pctLabel?: ReactNode
  // (دازاي 11-15) سطر دعوة تحت الرقم («← عرض التفاصيل») للبطاقات الي چانت تكتبه
  cta?: ReactNode
  onClick?: () => void
  active?: boolean
  className?: string
}) {
  const t = TONE[tone]
  const p = pct == null || !isFinite(pct) ? null : Math.max(0, Math.min(100, Math.round(pct)))
  const body = (
    <>
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className={`text-[13px] font-extrabold leading-snug ${tone === 'slate' ? 'text-slate-700' : t.ink}`}>{label}</p>
          <p className={`mt-1 font-extrabold tabular-nums text-[#0f2040] ${big ? 'text-3xl' : 'text-2xl'}`}>{value}</p>
          {hint && <p className="mt-0.5 text-[11px] leading-snug text-slate-500">{hint}</p>}
          {cta && <p className="mt-2 text-xs font-bold opacity-80">{cta}</p>}
        </div>
        <div className="flex shrink-0 flex-col items-center gap-2">
          {icon && <span className={`grid place-items-center rounded-2xl ${big ? 'h-14 w-14 text-2xl' : 'h-12 w-12 text-xl'} ${t.chip}`}>{icon}</span>}
          {(to || onClick) && <span className="grid h-6 w-6 place-items-center rounded-full bg-white text-xs text-slate-500 shadow-sm">‹</span>}
        </div>
      </div>
      {p != null && (
        <div className="mt-2 flex items-center gap-2">
          <div className="h-2 flex-1 rounded-full bg-white/80"><div className={`h-2 rounded-full ${tone === 'blue' ? 'bg-sky-500' : tone === 'amber' ? 'bg-amber-500' : 'bg-emerald-500'}`} style={{ width: `${p}%` }} /></div>
          {pctLabel ?? (why
            ? <PctWhy title={why.title} value={p} parts={why.parts} meaning={why.meaning} source={why.source} className={`text-[11px] font-extrabold tabular-nums ${t.ink}`} />
            : <span className={`text-[11px] font-extrabold tabular-nums ${t.ink}`}>{p}%</span>)}
        </div>
      )}
    </>
  )
  const cls = `block rounded-2xl border p-3.5 transition ${t.card}${active ? ' ring-2 ring-brand-500' : ''} ${className}`
  if (onClick) return <button type="button" onClick={onClick} className={`${cls} w-full text-right hover:-translate-y-0.5 hover:shadow-md`}>{body}</button>
  return to
    ? <Link to={to} className={`${cls} hover:-translate-y-0.5 hover:shadow-md`}>{body}</Link>
    : <div className={cls}>{body}</div>
}
