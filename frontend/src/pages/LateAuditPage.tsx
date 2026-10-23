import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import PageHeader from '../components/PageHeader'

// ═══ فحص المتأخرات — قرار (ع) 10-15 ═══
// «ليش متأخرات؟ هل بيهن مشكلة؟ وشكد مدة التأخير؟» — نفس تعريف ماتركس، ولكل حجز
// سبب مرجّح من الأدلة، والحجوزات الي بياناتها مشكوكة معلّمة.

interface Row {
  id: string; code: string; customer: string | null; service: string | null; division: string; status: string
  scheduledAt: string; completedAt: string | null; daysLate: number; cause: string; causeDetail: string
  dataIssues: string[]; leader: string | null; team: string | null; postpones: number; reschedules: number
}
const day = (s: string | null) => (s ? new Date(s).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }) : '—')
const ago = (n: number) => new Date(Date.now() - n * 86400_000).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

export default function LateAuditPage() {
  const [from, setFrom] = useState(ago(30))
  const [to, setTo] = useState(ago(0))
  const [data, setData] = useState<{ total: number; rows: Row[]; causes: Record<string, string> } | null>(null)
  const [err, setErr] = useState('')
  const [cause, setCause] = useState('')
  const [div, setDiv] = useState('')
  const [state, setState] = useState('')
  const [minDays, setMinDays] = useState(0)
  const [onlyIssues, setOnlyIssues] = useState(false)
  const [leader, setLeader] = useState('')

  useEffect(() => {
    let alive = true
    api.lateAudit(from, to).then((d) => { if (alive) { setData(d as never); setErr('') } }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب') })
    return () => { alive = false }
  }, [from, to])

  const rows = useMemo(() => (data?.rows ?? []).filter((r) =>
    (!cause || r.cause === cause) && (!div || r.division === div) && r.daysLate >= minDays &&
    (!state || (state === 'open' ? r.status !== 'COMPLETED' : r.status === 'COMPLETED')) &&
    (!onlyIssues || r.dataIssues.length > 0) && (!leader || (r.leader ?? 'بلا ليدر') === leader)), [data, cause, div, minDays, state, onlyIssues, leader])

  const all = data?.rows ?? []
  const avg = all.length ? (all.reduce((a, r) => a + r.daysLate, 0) / all.length).toFixed(1) : '0'
  const count = (k: (r: Row) => string) => {
    const m = new Map<string, number>()
    for (const r of all) m.set(k(r), (m.get(k(r)) ?? 0) + 1)
    return [...m.entries()].sort((a, b) => b[1] - a[1])
  }
  const byCause = count((r) => r.cause)
  const byLeader = count((r) => r.leader ?? 'بلا ليدر')
  const issues = all.filter((r) => r.dataIssues.length > 0).length

  const excel = () => api.lateAuditExcel(from, to).catch((e) => alert(e instanceof Error ? e.message : 'تعذر'))

  return (
    <div dir="rtl" className="space-y-4">
      <PageHeader title="🔍 فحص المتأخرات" subtitle="كل حجز متأخر: شكد تأخّر، ليش، وهل بياناته صحيحة — نفس تعريف ماتركس بالضبط" />
      <div className="flex flex-wrap items-center gap-2 rounded-2xl bg-white p-3 shadow-sm">
        <span className="text-sm text-slate-500">موعدها من</span>
        <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm" />
        <span className="text-sm text-slate-500">إلى</span>
        <input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm" />
        <select value={div} onChange={(e) => setDiv(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
          <option value="">كل الأقسام</option><option value="ENGINEERING">هندسي</option><option value="DECOR">ديكور</option>
        </select>
        <select value={state} onChange={(e) => setState(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
          <option value="">مفتوح وخالص</option><option value="open">بعده مفتوح</option><option value="done">خلص متأخر</option>
        </select>
        <span className="text-sm text-slate-500">تأخير ≥</span>
        <input type="number" min={0} value={minDays} onChange={(e) => setMinDays(Number(e.target.value) || 0)} className="w-16 rounded-lg border border-slate-300 px-2 py-1 text-sm" />
        <span className="text-sm text-slate-500">يوم</span>
        <button onClick={() => setOnlyIssues((v) => !v)} className={`rounded-lg px-3 py-1.5 text-sm font-bold ${onlyIssues ? 'bg-amber-500 text-white' : 'bg-amber-50 text-amber-800'}`}>⚠️ بياناتها مشكوكة ({issues})</button>
        <button onClick={excel} className="mr-auto rounded-lg bg-emerald-700 px-3 py-1.5 text-sm font-bold text-white">⬇️ إكسل</button>
      </div>
      {err && <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>}

      {data && (
        <div className="grid grid-cols-1 gap-3 lg:grid-cols-4">
          <div className="rounded-2xl bg-white p-4 shadow-sm">
            <p className="text-sm text-slate-500">المتأخرات</p>
            <p className="text-3xl font-black text-red-600">{all.length}<span className="text-base font-bold text-slate-400"> من {data.total}</span></p>
            <p className="text-xs text-slate-500">{data.total ? Math.round((all.length / data.total) * 100) : 0}٪ من حجوزات الفترة · معدل التأخير {avg} يوم</p>
          </div>
          <div className="rounded-2xl bg-white p-4 shadow-sm lg:col-span-2">
            <p className="mb-1 text-sm font-bold text-slate-600">ليش تأخرت؟</p>
            <div className="flex flex-wrap gap-1">
              {byCause.map(([k, n]) => (
                <button key={k} onClick={() => setCause(cause === k ? '' : k)} className={`rounded-lg px-2 py-1 text-xs font-bold ${cause === k ? 'bg-brand-600 text-white' : 'bg-slate-100 text-slate-700'}`}>{data.causes[k] ?? k}: {n}</button>
              ))}
            </div>
          </div>
          <div className="rounded-2xl bg-white p-4 shadow-sm">
            <p className="mb-1 text-sm font-bold text-slate-600">حسب الليدر</p>
            <div className="flex max-h-24 flex-wrap gap-1 overflow-auto">
              {byLeader.map(([k, n]) => (
                <button key={k} onClick={() => setLeader(leader === k ? '' : k)} className={`rounded-lg px-2 py-1 text-xs ${leader === k ? 'bg-brand-600 text-white' : 'bg-slate-100'}`}>{k}: {n}</button>
              ))}
            </div>
          </div>
        </div>
      )}

      <div className="overflow-x-auto rounded-2xl bg-white shadow-sm">
        <table className="min-w-full text-right text-sm">
          <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
            <tr><th className="px-3 py-2">الحجز</th><th className="px-3 py-2">الموعد ← الإنجاز</th><th className="px-3 py-2">التأخير</th><th className="px-3 py-2">السبب المرجّح</th><th className="px-3 py-2">الليدر / الكادر</th><th className="px-3 py-2">ملاحظة</th></tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {rows.map((r) => (
              <tr key={r.id} className={r.dataIssues.length ? 'bg-amber-50/60' : ''}>
                <td className="px-3 py-2">
                  <Link to={`/bookings?focus=${r.id}`} className="font-bold text-brand-700 underline">{r.code}</Link>
                  <span className="block text-xs text-slate-500">{r.customer} · {r.service} · {r.division === 'DECOR' ? '🎨 ديكور' : '🔧 هندسي'}</span>
                </td>
                <td className="px-3 py-2 text-xs">{day(r.scheduledAt)} ← {r.status === 'COMPLETED' ? day(r.completedAt) : <b className="text-red-600">بعده مفتوح</b>}
                  {r.reschedules > 0 && <span className="block text-slate-400">انعاد جدولته {r.reschedules} مرة</span>}</td>
                <td className="px-3 py-2"><span className={`rounded-full px-2 py-0.5 text-xs font-black ${r.daysLate >= 7 ? 'bg-red-100 text-red-700' : r.daysLate >= 3 ? 'bg-amber-100 text-amber-800' : 'bg-slate-100 text-slate-700'}`}>{r.daysLate} يوم</span></td>
                <td className="px-3 py-2"><b>{data?.causes[r.cause] ?? r.cause}</b>{r.causeDetail && <span className="block text-xs text-slate-500">{r.causeDetail}</span>}</td>
                <td className="px-3 py-2 text-xs">{r.leader ?? '—'}{r.team && <span className="block text-slate-400">{r.team}</span>}</td>
                <td className="px-3 py-2 text-xs">{r.dataIssues.map((d) => <span key={d} className="mb-0.5 block rounded bg-amber-100 px-1.5 py-0.5 font-bold text-amber-800">⚠️ {d}</span>)}</td>
              </tr>
            ))}
            {data && rows.length === 0 && <tr><td colSpan={6} className="p-4 text-center text-slate-400">ماكو حجوزات متأخرة بهذا الفلتر ✓</td></tr>}
          </tbody>
        </table>
      </div>
    </div>
  )
}
