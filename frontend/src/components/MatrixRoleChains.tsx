import { Fragment, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type RoleChainReport } from '../api'
import MatrixNote from './MatrixNote'
import { CHAIN_TONE } from './chainTone'

// ═══ ماتركس ٢٠٥٠ — تقارير الأدوار من «سلسلة الحجز» ═══
// لكل دور: كل موظف ومحطاته، مقارنته بفريقه، اتجاهه عن الفترة السابقة،
// ونمط تأخيره (يوم/ساعة)، وحجوزاته الي انكسرت بيها السلسلة.
// للمدير والمالك بس — بيها «ماتركس على المراقب».

const fm = (m: number | null) => {
  if (m == null) return '—'
  if (m < 60) return `${m}د`
  if (m < 1440) return `${Math.floor(m / 60)}س${m % 60 ? ` ${m % 60}د` : ''}`
  return `${Math.floor(m / 1440)}ي${Math.floor((m % 1440) / 60) ? ` ${Math.floor((m % 1440) / 60)}س` : ''}`
}
const pctTone = (p: number) => (p >= 80 ? 'text-emerald-600' : p >= 60 ? 'text-amber-600' : 'text-red-600')

export default function MatrixRoleChains() {
  const [roles, setRoles] = useState<{ key: string; title: string }[]>([])
  const [role, setRole] = useState('COORDINATOR')
  const [days, setDays] = useState(30)
  const [rep, setRep] = useState<{ k: string; r: RoleChainReport | null } | null>(null)
  const [open, setOpen] = useState<string | null>(null)
  const key = `${role}:${days}`
  const loading = rep?.k !== key
  const r = loading ? null : rep.r

  useEffect(() => { void api.getRoleChainRoles().then(setRoles).catch(() => {}) }, [])
  useEffect(() => {
    let alive = true
    void api.getRoleChain(role, days).catch(() => null).then((d) => { if (alive) setRep({ k: key, r: d }) })
    return () => { alive = false }
  }, [role, days, key])

  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-slate-200 bg-white p-3 text-slate-800">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm font-extrabold text-[#0f2040]">🔗 سلسلة الحجز — تقارير الأدوار <span className="text-[11px] font-normal text-slate-500">كل دور يتحاسب على محطاته هو</span></p>
        <select value={days} onChange={(e) => setDays(Number(e.target.value))} className="rounded-lg border border-slate-200 px-2 py-1 text-xs">
          {[7, 30, 60, 90].map((d) => <option key={d} value={d}>آخر {d} يوم</option>)}
        </select>
      </div>
      <div className="flex flex-wrap gap-1">
        {roles.map((x) => (
          <button key={x.key} type="button" onClick={() => { setRole(x.key); setOpen(null) }}
            className={`rounded-full px-3 py-1 text-xs font-bold ${role === x.key ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'}`}>{x.title}</button>
        ))}
      </div>
      {loading ? <p className="text-xs text-slate-400">ماتركس يحلّل…</p> : !r ? <p className="text-xs text-red-600">تعذر جلب التقرير</p> : (
        <>
          <MatrixNote>{r.insights.join(' ')}</MatrixNote>
          {r.stations.length > 0 && (
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
              {r.stations.map((s) => (
                <div key={s.key} className="rounded-xl border border-slate-200 bg-slate-50 p-2 text-xs">
                  <b className="block text-slate-800">{s.title}</b>
                  <span className="text-slate-500">{s.count} قياس · الوسيط {fm(s.medianMin)}</span>
                  <div className="mt-1 flex gap-2"><span>✅ {s.ok}</span><span>🟠 {s.late}</span><span>🔴 {s.missed}</span>{s.issue > 0 && <span>⚠️ {s.issue}</span>}</div>
                </div>
              ))}
            </div>
          )}
          <div className="overflow-x-auto">
            <table className="w-full min-w-[640px] text-right text-xs">
              <thead className="bg-slate-100 text-slate-600">
                <tr><th className="p-2">الموظف</th><th className="p-2">قياسات</th><th className="p-2">بوقتها</th><th className="p-2">قبلها</th><th className="p-2">🟠</th><th className="p-2">🔴</th><th className="p-2">⚠️</th><th className="p-2">أضعف محطة</th>{role === 'TECH' && <th className="p-2">تقييم الليدرية</th>}</tr>
              </thead>
              <tbody>
                {r.employees.map((e) => (
                  <Fragment key={e.id || 'none'}>
                    <tr onClick={() => setOpen(open === e.id ? null : e.id)} className="cursor-pointer border-t border-slate-100 hover:bg-sky-50">
                      <td className="p-2 font-bold">{open === e.id ? '▾' : '▸'} {e.name}</td>
                      <td className="p-2">{e.total}</td>
                      <td className={`p-2 font-black ${pctTone(e.onTimePct)}`}>{e.total ? `${e.onTimePct}%` : '—'}</td>
                      <td className="p-2 text-slate-500">{e.prevPct != null ? `${e.prevPct}%` : '—'}</td>
                      <td className="p-2">{e.late}</td><td className="p-2">{e.missed}</td><td className="p-2">{e.issues}</td>
                      <td className="p-2">{e.worst || '—'}</td>
                      {role === 'TECH' && <td className="p-2">{e.ratingAvg != null ? `${e.ratingAvg}/5 (${e.ratingCount})` : '—'}</td>}
                    </tr>
                    {open === e.id && (
                      <tr className="bg-slate-50">
                        <td colSpan={role === 'TECH' ? 9 : 8} className="space-y-2 p-3">
                          {e.insights.length > 0 && <MatrixNote>{e.insights.join(' ')}</MatrixNote>}
                          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                            {e.stations.map((s) => (
                              <div key={s.key} className="rounded-lg border border-slate-200 bg-white p-2">
                                <b>{s.title}</b>
                                <p className="text-slate-500">وسيطه {fm(s.medianMin)} · معدله {fm(s.avgMin)} · الفريق {fm(s.teamMedianMin)}</p>
                                <p>✅ {s.ok} · 🟠 {s.late} · 🔴 {s.missed}{s.issue ? ` · ⚠️ ${s.issue}` : ''}</p>
                              </div>
                            ))}
                          </div>
                          {e.ratingNotes.length > 0 && <div><b>ملاحظات الليدرية:</b>{e.ratingNotes.map((n, i) => <p key={i} className="text-slate-600">• {n}</p>)}</div>}
                          {e.bad && e.bad.length > 0 && (
                            <div>
                              <b>وين انكسرت السلسلة:</b>
                              {e.bad.map((b, i) => (
                                <p key={i} className="text-slate-700">
                                  {CHAIN_TONE[b.status].icon} <Link to={`/bookings?focus=${b.id}`} className="font-bold text-sky-700 underline">{b.code}</Link> — {b.station}: {b.note}
                                </p>
                              ))}
                            </div>
                          )}
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))}
                {r.employees.length === 0 && <tr><td colSpan={9} className="p-4 text-center text-slate-400">ماكو قياسات لهذا الدور بهالفترة.</td></tr>}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
