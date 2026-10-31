import { useEffect, useState } from 'react'
import { api, type ProjectStageStats } from '../api'

// ═══ إحصائيات المشاريع — تصميم (ع) 10-13 ═══
// بطاقات لكل مرحلة، رسم توزيع، دائرة الإنجاز، أعلى المراحل، تطور ١٢ شهر،
// ملخص سريع، وجدول التفاصيل. كله من المشاريع نفسها — بلا أرقام مخترعة.

type Meta = { color: string; icon: string; tint: string }
function metaOf(stage: string): Meta {
  if (stage.includes('مكتمل')) return { color: '#22c55e', icon: '✅', tint: 'bg-emerald-50 border-emerald-100' }
  if (stage.includes('مرفوض')) return { color: '#f43f5e', icon: '❌', tint: 'bg-rose-50 border-rose-100' }
  if (stage.includes('عقد')) return { color: '#3b82f6', icon: '📄', tint: 'bg-sky-50 border-sky-100' }
  if (stage.includes('كشف')) return { color: '#f59e0b', icon: '🔍', tint: 'bg-amber-50 border-amber-100' }
  if (stage.includes('سعر')) return { color: '#a855f7', icon: '🧾', tint: 'bg-violet-50 border-violet-100' }
  if (stage.includes('اتصال')) return { color: '#0ea5e9', icon: '📞', tint: 'bg-cyan-50 border-cyan-100' }
  if (stage.includes('تجهيز')) return { color: '#f59e0b', icon: '📦', tint: 'bg-orange-50 border-orange-100' }
  if (stage.includes('التسليم')) return { color: '#14b8a6', icon: '🤝', tint: 'bg-teal-50 border-teal-100' }
  if (stage.includes('تنفيذ')) return { color: '#f97316', icon: '🛠️', tint: 'bg-orange-50 border-orange-100' }
  return { color: '#64748b', icon: '📁', tint: 'bg-slate-50 border-slate-200' }
}
const clean = (s: string) => s.replace(/^\d+\.\s*/, '')
const MONTHS = ['يناير', 'فبراير', 'مارس', 'أبريل', 'مايو', 'يونيو', 'يوليو', 'أغسطس', 'سبتمبر', 'أكتوبر', 'نوفمبر', 'ديسمبر']
const monthName = (m: string) => MONTHS[Number(m.slice(5, 7)) - 1] ?? m
const card = 'rounded-2xl border border-slate-200 bg-white p-4 shadow-sm'

export default function ProjectStatsBoard() {
  const [stats, setStats] = useState<ProjectStageStats[] | null>(null)
  const [months, setMonths] = useState<Awaited<ReturnType<typeof api.getProjectMonths>>>([])
  useEffect(() => {
    api.getProjectStageStats().then(setStats).catch(() => setStats([]))
    api.getProjectMonths().then(setMonths).catch(() => setMonths([]))
  }, [])
  if (!stats) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  if (stats.length === 0) return <p className={`${card} text-center text-slate-500`}>لا توجد مشاريع بعد</p>

  const total = stats.reduce((n, r) => n + r.count, 0)
  const pct = (n: number) => (total ? Math.round((n / total) * 100) : 0)
  const done = stats.filter((r) => r.stage.includes('مكتمل')).reduce((n, r) => n + r.count, 0)
  const rejected = stats.filter((r) => r.stage.includes('مرفوض')).reduce((n, r) => n + r.count, 0)
  const open = total - done - rejected
  const sorted = [...stats].sort((a, b) => b.count - a.count)
  const max = Math.max(1, ...stats.map((r) => r.count))
  const withNew = months.filter((m) => m.new > 0)
  const best = withNew.length ? withNew.reduce((a, b) => (b.new > a.new ? b : a)) : null
  const worst = withNew.length ? withNew.reduce((a, b) => (b.new < a.new ? b : a)) : null
  const avg = months.length ? months.reduce((n, m) => n + m.new, 0) / months.length : 0

  return (
    <div dir="rtl" className="space-y-4">
      {/* البطاقات */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 xl:grid-cols-6">
        <div className={`${card} flex items-center justify-between`}>
          <div><p className="text-xs font-bold text-[#0f2040]">إجمالي المشاريع</p><p className="text-3xl font-black text-[#0f2040]">{total}</p><p className="text-[11px] text-slate-400">100%</p></div>
          <span className="grid h-12 w-12 place-items-center rounded-2xl bg-indigo-50 text-2xl">🗂️</span>
        </div>
        {sorted.slice(0, 5).map((r) => {
          const m = metaOf(r.stage)
          return (
            <div key={r.stage} className={`flex items-center justify-between rounded-2xl border p-4 shadow-sm ${m.tint}`}>
              <div><p className="text-xs font-bold" style={{ color: m.color }}>{clean(r.stage)}</p><p className="text-3xl font-black text-[#0f2040]">{r.count}</p><p className="text-[11px] text-slate-500">{pct(r.count)}% من الإجمالي</p></div>
              <span className="grid h-12 w-12 place-items-center rounded-2xl bg-white text-2xl">{m.icon}</span>
            </div>
          )
        })}
      </div>

      <div className="grid gap-4 xl:grid-cols-[1.4fr_1fr_1fr]">
        {/* توزيع حسب المرحلة */}
        <div className={card}>
          <h3 className="mb-3 font-extrabold text-[#0f2040]">📊 توزيع المشاريع حسب المرحلة</h3>
          <div className="flex h-56 items-end gap-3 border-b border-slate-200 px-2">
            {stats.map((r) => (
              <div key={r.stage} className="flex flex-1 flex-col items-center justify-end gap-1">
                <b className="text-sm text-[#0f2040]">{r.count}</b>
                <div className="w-full max-w-[56px] rounded-t-md" style={{ height: `${Math.max(2, (r.count / max) * 180)}px`, background: metaOf(r.stage).color }} />
              </div>
            ))}
          </div>
          <div className="mt-1 flex gap-3 px-2">{stats.map((r) => <p key={r.stage} className="flex-1 text-center text-[11px] text-slate-600">{clean(r.stage)}</p>)}</div>
        </div>

        {/* نسبة الإنجاز */}
        <div className={card}>
          <h3 className="mb-3 font-extrabold text-[#0f2040]">🎯 نسبة الإنجاز العامة</h3>
          <div className="flex items-center gap-4">
            <Donut parts={[{ v: done, c: '#22c55e' }, { v: open, c: '#3b82f6' }, { v: rejected, c: '#f43f5e' }]} total={total} />
            <ul className="flex-1 space-y-3 text-sm">
              {[['المشاريع المكتملة', done, '#22c55e'], ['قيد التنفيذ', open, '#3b82f6'], ['المشاريع المرفوضة', rejected, '#f43f5e']].map(([l, v, c]) => (
                <li key={l as string} className="flex items-center justify-between gap-2">
                  <span className="flex items-center gap-1.5"><i className="h-2.5 w-2.5 rounded-full" style={{ background: c as string }} />{l}</span>
                  <span className="tabular-nums"><b style={{ color: c as string }}>{pct(v as number)}%</b> · {v}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* الأعلى عدداً */}
        <div className={card}>
          <h3 className="mb-3 font-extrabold text-[#0f2040]">🏆 المراحل الأعلى عدداً</h3>
          <ol className="space-y-2">
            {sorted.slice(0, 5).map((r, i) => {
              const m = metaOf(r.stage)
              return (
                <li key={r.stage} className="flex items-center gap-2 text-sm">
                  <span className="grid h-6 w-6 place-items-center rounded-full bg-slate-100 text-xs font-bold">{i + 1}</span>
                  <span>{m.icon}</span><span className="flex-1 font-bold text-slate-700">{clean(r.stage)}</span>
                  <span className="rounded-lg px-3 py-0.5 text-xs font-black text-white" style={{ background: m.color }}>{r.count}</span>
                </li>
              )
            })}
          </ol>
        </div>
      </div>

      <div className="grid gap-4 xl:grid-cols-[2fr_1fr]">
        {/* تطور ١٢ شهر */}
        <div className={card}>
          <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
            <h3 className="font-extrabold text-[#0f2040]">📈 تطور المشاريع عبر الأشهر</h3>
            <div className="flex gap-3 text-[11px]">{[['المشاريع الجديدة', '#f43f5e'], ['قيد التنفيذ', '#3b82f6'], ['المكتملة', '#22c55e']].map(([l, c]) => <span key={l} className="flex items-center gap-1"><i className="h-2 w-2 rounded-full" style={{ background: c }} />{l}</span>)}</div>
          </div>
          {months.length > 0 ? <Lines months={months} /> : <p className="py-8 text-center text-sm text-slate-400">ماكو حركة بعد.</p>}
        </div>

        {/* ملخص سريع */}
        <div className={card}>
          <h3 className="mb-3 font-extrabold text-[#0f2040]">🧾 ملخص سريع</h3>
          <div className="grid grid-cols-2 gap-2">
            {[['📈', 'أعلى شهر بمشاريع جديدة', best ? `${monthName(best.month)} (${best.new})` : '—'], ['📊', 'متوسط المشاريع الجديدة شهرياً', avg.toFixed(1)],
              ['🎯', 'نسبة الإنجاز', `${pct(done)}%`], ['📉', 'أقل شهر بمشاريع جديدة', worst ? `${monthName(worst.month)} (${worst.new})` : '—']].map(([i, l, v]) => (
              <div key={l} className="rounded-xl border border-slate-100 bg-slate-50 p-3">
                <p className="text-[11px] text-slate-500">{i} {l}</p><p className="mt-1 text-lg font-black text-[#0f2040]">{v}</p>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* الجدول */}
      <div className={`${card} overflow-x-auto`}>
        <h3 className="mb-2 font-extrabold text-[#0f2040]">📋 تفاصيل إحصائيات المشاريع</h3>
        <table className="w-full text-sm">
          <thead><tr className="bg-[#0f2040] text-white">{['#', 'مرحلة المشروع', 'عدد المشاريع', 'النسبة من الإجمالي', 'مؤشر مرئي'].map((h) => <th key={h} className="px-3 py-2 text-right text-xs">{h}</th>)}</tr></thead>
          <tbody>
            {stats.map((r, i) => {
              const m = metaOf(r.stage)
              return (
                <tr key={r.stage} className="border-b border-slate-100">
                  <td className="px-3 py-2 text-slate-400">{i + 1}</td>
                  <td className="px-3 font-bold">{m.icon} {clean(r.stage)}</td>
                  <td className="px-3 tabular-nums">{r.count}</td>
                  <td className="px-3 tabular-nums">{pct(r.count)}%</td>
                  <td className="px-3"><div className="h-2 w-40 rounded-full bg-slate-100"><div className="h-2 rounded-full" style={{ width: `${pct(r.count)}%`, background: m.color }} /></div></td>
                </tr>
              )
            })}
            <tr className="bg-slate-50 font-black"><td /><td className="px-3 py-2">الإجمالي</td><td className="px-3">{total}</td><td className="px-3">100%</td><td /></tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}

function Donut({ parts, total }: { parts: { v: number; c: string }[]; total: number }) {
  const R = 52, C = 2 * Math.PI * R
  const lens = parts.map((p) => (total ? (p.v / total) * C : 0))
  const offs = lens.map((_, i) => lens.slice(0, i).reduce((a, b) => a + b, 0))
  return (
    <svg width="140" height="140" viewBox="0 0 140 140" className="shrink-0">
      <g transform="rotate(-90 70 70)">
        {parts.map((p, i) => (
          <circle key={i} cx="70" cy="70" r={R} fill="none" stroke={p.c} strokeWidth="20" strokeDasharray={`${lens[i]} ${C - lens[i]}`} strokeDashoffset={-offs[i]} />
        ))}
      </g>
      <text x="70" y="68" textAnchor="middle" className="fill-[#0f2040] text-2xl font-black">{total}</text>
      <text x="70" y="88" textAnchor="middle" className="fill-slate-500 text-[10px]">إجمالي المشاريع</text>
    </svg>
  )
}

function Lines({ months }: { months: { month: string; new: number; completed: number; inProgress: number }[] }) {
  const W = 700, H = 200, P = 24
  const max = Math.max(1, ...months.flatMap((m) => [m.new, m.completed, m.inProgress]))
  const x = (i: number) => P + (i * (W - 2 * P)) / Math.max(1, months.length - 1)
  const y = (v: number) => H - P - (v / max) * (H - 2 * P)
  const path = (k: 'new' | 'completed' | 'inProgress') => months.map((m, i) => `${i ? 'L' : 'M'}${x(i)},${y(m[k])}`).join(' ')
  return (
    <svg viewBox={`0 0 ${W} ${H + 20}`} className="w-full">
      {[0, 0.25, 0.5, 0.75, 1].map((f) => <line key={f} x1={P} x2={W - P} y1={y(max * f)} y2={y(max * f)} stroke="#e2e8f0" strokeDasharray="3 3" />)}
      {([['inProgress', '#3b82f6'], ['completed', '#22c55e'], ['new', '#f43f5e']] as const).map(([k, c]) => (
        <g key={k}>
          <path d={path(k)} fill="none" stroke={c} strokeWidth="2.5" />
          {months.map((m, i) => <circle key={i} cx={x(i)} cy={y(m[k])} r="3.5" fill={c}><title>{`${monthName(m.month)}: ${m[k]}`}</title></circle>)}
        </g>
      ))}
      {months.map((m, i) => <text key={m.month} x={x(i)} y={H + 12} textAnchor="middle" className="fill-slate-500 text-[10px]">{monthName(m.month)}</text>)}
    </svg>
  )
}
