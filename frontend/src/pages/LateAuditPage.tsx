import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import PctWhy from '../ui/PctWhy'
import PageHeader from '../components/PageHeader'
import { dayBaghdad } from '../utils/format'
import { LineIcon } from '../components/LineIcon'

import { notify } from '../ui/notify'
// ═══ فحص المتأخرات — قرار (ع) 10-15 ═══
// «ليش متأخرات؟ هل بيهن مشكلة؟ وشكد مدة التأخير؟» — نفس تعريف ماتركس، ولكل حجز
// سبب مرجّح من الأدلة، والحجوزات الي بياناتها مشكوكة معلّمة.

interface Row {
  id: string; code: string; customer: string | null; service: string | null; division: string; status: string
  scheduledAt: string; completedAt: string | null; daysLate: number; cause: string; causeDetail: string
  dataIssues: string[]; leader: string | null; team: string | null; postpones: number; reschedules: number
  waitingSince?: string | null; waitingKind?: string | null; waitingNote?: string | null; contactAttempts?: number
}
const day = (s: string | null) => dayBaghdad(s) || '—'
const ago = (n: number) => new Date(Date.now() - n * 86400_000).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

export default function LateAuditPage() {
  const [from, setFrom] = useState(ago(30))
  const [to, setTo] = useState(ago(0))
  const [data, setData] = useState<{ total: number; rows: Row[]; waiting?: Row[]; minDays?: number; causes: Record<string, string> } | null>(null)
  const [err, setErr] = useState('')
  const [cause, setCause] = useState('')
  const [div, setDiv] = useState('')
  const [state, setState] = useState('')
  const [minDays, setMinDays] = useState(0)
  const [onlyIssues, setOnlyIssues] = useState(false)
  const [leader, setLeader] = useState('')
  const [q, setQ] = useState('')
  const [view, setView] = useState<'list' | 'grid'>('list')
  const [adv, setAdv] = useState(false)

  useEffect(() => {
    let alive = true
    api.lateAudit(from, to).then((d) => { if (alive) { setData(d as never); setErr('') } }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب') })
    return () => { alive = false }
  }, [from, to])

  const rows = useMemo(() => (data?.rows ?? []).filter((r) =>
    (!cause || r.cause === cause) && (!div || r.division === div) && r.daysLate >= minDays &&
    (!state || (state === 'open' ? r.status !== 'COMPLETED' : r.status === 'COMPLETED')) &&
    (!onlyIssues || r.dataIssues.length > 0) && (!leader || (r.leader ?? 'بلا ليدر') === leader) &&
    (!q.trim() || `${r.code} ${r.customer ?? ''} ${data?.causes[r.cause] ?? r.cause} ${r.causeDetail}`.includes(q.trim()))), [data, cause, div, minDays, state, onlyIssues, leader, q])

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

  const excel = () => api.lateAuditExcel(from, to).catch((e) => notify(e instanceof Error ? e.message : 'تعذر'))

  return (
    <div dir="rtl" className="space-y-4">
      <PageHeader title="🔍 فحص المتأخرات" subtitle={`المتأخر = عدّى ${data?.minDays ?? 2} يوم أو أكثر على يوم موعده (يوم واحد مو تأخير). الي حاطينه «زبون ما رد» أو «الزبون يقرر» بقسم منفصل تحت.`} />
      <div className="flex flex-wrap items-center gap-2 rounded-2xl bg-white p-3 shadow-sm">
        <span className="text-sm text-slate-500">موعدها من</span>
        <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm" />
        <span className="text-sm text-slate-500">إلى</span>
        <input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm" />
        <button onClick={excel} className="mr-auto rounded-lg bg-emerald-700 px-3 py-1.5 text-sm font-bold text-white">⬇️ إكسل</button>
      </div>
      {err && <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>}

      {data && (
        <div className="grid grid-cols-1 gap-3 lg:grid-cols-4">
          <div className="rounded-2xl bg-white p-4 shadow-sm">
            <p className="text-sm text-slate-500">المتأخرات</p>
            <p className="text-3xl font-black text-red-600">{all.length}<span className="text-base font-bold text-slate-400"> من {data.total}</span></p>
            <p className="text-xs text-slate-500"><PctWhy title="نسبة المتأخرات" suffix="٪" parts={{ numLabel: 'حجز متأخر', num: all.length, denLabel: 'حجز موعده بالفترة', den: data.total }}
              meaning={<>شكد من حجوزات الفترة تأخرت. المتأخر = خلص بعد موعده بيومين أو أكثر، أو بعده مفتوح وعدّى على موعده يومين أو أكثر. كل ما تنزل أحسن. الحجوزات الي بالانتظار (الزبون ما رد / يقرر) ما تنحسب متأخرة.</>}
              source={<>جدول الحجوزات: كل حجز موعده بين {from} و{to}، بلا الملغية والمؤرشفة والداخلية. معدل التأخير {avg} يوم = مجموع أيام التأخير ÷ عدد المتأخرات.</>} /> من حجوزات الفترة · معدل التأخير {avg} يوم</p>
          </div>
          <div className="rounded-2xl bg-white p-4 shadow-sm lg:col-span-2">
            <p className="mb-1 text-sm font-bold text-slate-600">ليش تأخرت؟</p>
            <div className="flex flex-wrap gap-1">
              {byCause.map(([k, n]) => (
                <button key={k} onClick={() => setCause(cause === k ? '' : k)} className={`rounded-lg px-2 py-1 text-xs font-bold ${cause === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-700'}`}>{data.causes[k] ?? k}: {n}</button>
              ))}
            </div>
          </div>
          <div className="rounded-2xl bg-white p-4 shadow-sm">
            <p className="mb-1 text-sm font-bold text-slate-600">حسب الليدر</p>
            <div className="flex max-h-24 flex-wrap gap-1 overflow-auto">
              {byLeader.map(([k, n]) => (
                <button key={k} onClick={() => setLeader(leader === k ? '' : k)} className={`rounded-lg px-2 py-1 text-xs ${leader === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100'}`}>{k}: {n}</button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* (ك) 11-23 زينا: الجدول مثل الصورة — بحث، قائمة/بطاقات، تصفية متقدمة، وأعمدة بأيقونات. الألوان بـindex.css (.la-*) */}
      <div className="la-panel space-y-3 rounded-2xl border p-3">
        <div className="flex flex-wrap items-center gap-3">
          <label className="la-field flex min-w-[260px] max-w-xl flex-1 items-center gap-3 rounded-xl border px-4 py-2">
            <LineIcon n="search" className="h-5 w-5 shrink-0" />
            <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="ابحث برقم الحجز، الاسم، أو السبب..." className="w-full bg-transparent text-[15px] outline-none" />
          </label>
          <div className="la-field ms-auto flex overflow-hidden rounded-xl border">
            <button type="button" onClick={() => setView('list')} aria-label="قائمة" className={`la-tog px-3 py-2 ${view === 'list' ? 'la-tog-on' : ''}`}><LineIcon n="list" className="h-5 w-5" /></button>
            <button type="button" onClick={() => setView('grid')} aria-label="بطاقات" className={`la-tog px-3 py-2 ${view === 'grid' ? 'la-tog-on' : ''}`}><LineIcon n="grid" className="h-5 w-5" /></button>
          </div>
          <button type="button" onClick={() => setAdv((v) => !v)} className={`la-adv flex items-center gap-2 rounded-xl border px-4 py-2 text-[15px] font-bold ${adv ? 'la-tog-on' : ''}`}><LineIcon n="filter" className="h-5 w-5" />تصفية متقدمة</button>
        </div>
        {adv && (
          <div className="flex flex-wrap items-center gap-2">
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
          </div>
        )}

        {view === 'list' ? (
          <div className="la-table overflow-x-auto rounded-2xl border">
            <table className="w-full min-w-[1100px] table-fixed border-separate border-spacing-y-0 text-[15px]">
              <colgroup>{['8%', '17%', '14%', '8%', '15%', '14%', '13%', '11%'].map((w, i) => <col key={i} style={{ width: w }} />)}</colgroup>
              <thead>
                <tr className="la-th h-12 font-bold">
                  {([['', 'رقم الحجز'], ['user', 'الإسم'], ['cal', 'الموعد'], ['clock', 'التأخير'], ['list', 'السبب المرجّح'], ['user', 'المدير / الكادر'], ['file', 'ملاحظة'], ['', 'الإجراءات']] as const).map(([ic, h]) => (
                    <th key={h} className="px-2 text-center"><span className="inline-flex items-center gap-1.5">{ic && <LineIcon n={ic} className="h-5 w-5" />}{h}</span></th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows.map((r) => {
                  const open = r.status !== 'COMPLETED'
                  return (
                    <tr key={r.id} className={`la-row h-14 ${r.dataIssues.length ? 'la-row-warn' : ''}`}>
                      <td className="px-2 text-center"><Link to={`/bookings?focus=${r.id}`} className="la-code inline-block rounded-lg border px-3 py-1 text-sm font-extrabold" dir="ltr">{r.code}</Link></td>
                      <td className="px-3 py-1.5">
                        <p className="font-extrabold">{r.customer ?? '—'}</p>
                        <p className="la-sub flex items-center gap-1 text-xs"><LineIcon n="wrench" className="h-3.5 w-3.5 shrink-0" />{[r.service, r.division === 'DECOR' ? 'ديكور' : 'هندسي'].filter(Boolean).join(' · ')}</p>
                      </td>
                      <td className="px-2 text-center text-sm font-bold">
                        <span dir="ltr">{day(r.scheduledAt)}</span> ← {open ? <b className="text-red-500">بعده</b> : <span dir="ltr">{day(r.completedAt)}</span>}
                        <span className={`block ${open ? 'font-extrabold text-emerald-400' : ''}`}>{open ? 'مفتوح' : r.completedAt ? new Date(r.completedAt).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'Asia/Baghdad' }) : ''}</span>
                        {r.reschedules > 0 && <span className="la-sub block text-xs font-normal">انعاد جدولته {r.reschedules} مرة</span>}
                      </td>
                      <td className="px-2 text-center"><span className={`la-late ${r.daysLate >= 30 ? 'la-d30' : r.daysLate >= 14 ? 'la-d14' : r.daysLate >= 7 ? 'la-d7' : 'la-d0'} inline-block min-w-[4.5rem] rounded-xl border px-3 py-1 text-sm font-extrabold`}>{r.daysLate} يوم</span></td>
                      <td className="px-2 text-center"><b>{data?.causes[r.cause] ?? r.cause}</b>{r.causeDetail && <span className="la-sub block text-xs">{r.causeDetail}</span>}</td>
                      <td className="px-2">
                        {r.leader || r.team ? (
                          <div className="flex items-center justify-center gap-2">
                            <span className="text-center"><b className="block text-sm">{r.leader ?? '—'}</b>{r.team && <span className="la-sub block text-xs">{r.team}</span>}</span>
                            <LineIcon n="user" className="h-6 w-6 shrink-0 text-sky-400" />
                          </div>
                        ) : <p className="text-center">-</p>}
                      </td>
                      <td className="px-2 text-center">{r.dataIssues.length ? r.dataIssues.map((d) => <span key={d} className="la-note mb-0.5 flex items-start gap-1.5 rounded-lg border px-2 py-1 text-xs font-bold"><LineIcon n="warn" className="mt-0.5 h-4 w-4 shrink-0 text-amber-400" />{d}</span>) : '-'}</td>
                      <td className="px-2">
                        <div className="flex items-center justify-center gap-2">
                          <Link to={`/bookings?focus=${r.id}`} title="عرض الحجز" className="la-act la-act-eye flex h-8 w-9 items-center justify-center rounded-lg border"><LineIcon n="eye" className="h-5 w-5" /></Link>
                          <Link to={`/bookings?focus=${r.id}`} title="تعديل بالحجوزات" className="la-act flex h-8 w-9 items-center justify-center rounded-lg border"><LineIcon n="edit" className="h-4 w-4" /></Link>
                          <button type="button" title="نسخ رقم الحجز" onClick={() => { void navigator.clipboard?.writeText(r.code).then(() => notify(`انسخ ${r.code}`)) }} className="la-act flex h-8 w-9 items-center justify-center rounded-lg border"><LineIcon n="dots" className="h-5 w-5" /></button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
                {data && rows.length === 0 && <tr><td colSpan={8} className="p-4 text-center opacity-70">ماكو حجوزات متأخرة بهذا الفلتر ✓</td></tr>}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
            {rows.map((r) => (
              <Link key={r.id} to={`/bookings?focus=${r.id}`} className={`la-row la-card block space-y-2 rounded-2xl border p-4 ${r.dataIssues.length ? 'la-row-warn' : ''}`}>
                <div className="flex items-center justify-between gap-2">
                  <span className="la-code rounded-lg border px-3 py-1 text-sm font-extrabold" dir="ltr">{r.code}</span>
                  <span className={`la-late ${r.daysLate >= 30 ? 'la-d30' : r.daysLate >= 14 ? 'la-d14' : r.daysLate >= 7 ? 'la-d7' : 'la-d0'} rounded-xl border px-3 py-1 text-sm font-extrabold`}>{r.daysLate} يوم</span>
                </div>
                <p className="font-extrabold">{r.customer ?? '—'}</p>
                <p className="text-sm"><b>{data?.causes[r.cause] ?? r.cause}</b>{r.causeDetail && <span className="la-sub"> · {r.causeDetail}</span>}</p>
                <p className="la-sub text-xs">{day(r.scheduledAt)} ← {r.status === 'COMPLETED' ? day(r.completedAt) : 'بعده مفتوح'} · {r.leader ?? 'بلا ليدر'}</p>
                {r.dataIssues.map((d) => <span key={d} className="la-note flex items-start gap-1.5 rounded-lg border px-2 py-1 text-xs font-bold"><LineIcon n="warn" className="mt-0.5 h-4 w-4 shrink-0 text-amber-400" />{d}</span>)}
              </Link>
            ))}
            {data && rows.length === 0 && <p className="p-4 text-center opacity-70">ماكو حجوزات متأخرة بهذا الفلتر ✓</p>}
          </div>
        )}
      </div>

      {/* (دازاي 11-06) بالانتظار: الموظف سجّل «زبون ما رد» أو «الزبون يقرر» — مو تأخير، بس تبين المحاولات */}
      {data?.waiting && data.waiting.length > 0 && (
        <div className="rounded-2xl bg-white p-3 shadow-sm">
          <p className="mb-2 text-sm font-bold text-slate-700">⏸️ بالانتظار — ما تنحسب متأخرة ({data.waiting.length})</p>
          <table className="min-w-full text-right text-sm">
            <thead className="bg-slate-50 text-slate-600">
              <tr><th className="px-3 py-1.5">الحجز</th><th className="px-3 py-1.5">الموعد</th><th className="px-3 py-1.5">السبب</th><th className="px-3 py-1.5">المحاولات</th><th className="px-3 py-1.5">من</th><th className="px-3 py-1.5">ملاحظة</th></tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {data.waiting.map((r) => (
                <tr key={r.id}>
                  <td className="px-3 py-1.5"><Link to={`/bookings?focus=${r.id}`} className="font-bold text-brand-700 underline">{r.code}</Link><span className="block text-xs text-slate-500">{r.customer}</span></td>
                  <td className="px-3 py-1.5 text-xs">{day(r.scheduledAt)}</td>
                  <td className="px-3 py-1.5 text-xs font-bold">{r.waitingKind === 'CUSTOMER_DECISION' ? '🤔 الزبون يقرر' : '📵 زبون ما رد'}</td>
                  <td className="px-3 py-1.5 text-xs"><span className={`rounded-full px-2 py-0.5 font-black ${(r.contactAttempts ?? 0) >= 3 ? 'bg-red-100 text-red-700' : 'bg-slate-100 text-slate-700'}`}>{r.contactAttempts ?? 0}</span></td>
                  <td className="px-3 py-1.5 text-xs">{day(r.waitingSince ?? null)}</td>
                  <td className="px-3 py-1.5 text-xs text-slate-500">{r.waitingNote ?? ''}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
