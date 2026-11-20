import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { notify } from '../ui/notify'

// 🛠️ صيانة وقائية (دازاي 11-06، فكرة ٧): زبائن فات ٦ أشهر على آخر شغلهم، أو عندهم شكاوى متكررة،
// وماكو حجز مفتوح. الجودة تتصل وتعرض صيانة.
// (بلاغ #55، 11-20) «ماكو قرار اتخذه كذالك بالعدد»: أزرار قرار لكل زبون (يختفي لمدة)،
// العدد الحقيقي مقسوم (متأخر صيانة / شكاوى)، والقائمة ٥٠ ٥٠.
type Row = { customerId: string; name: string; phone?: string | null; lastDone: string; service: string | null; months: number; complaints: number; reason: string; kind: 'LATE' | 'COMPLAINTS' }

const DECISIONS: { key: string; label: string; cls: string }[] = [
  { key: 'BOOKED', label: '✅ حجز صيانة', cls: 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100' },
  { key: 'DECLINED', label: '🚫 ما يريد', cls: 'bg-slate-100 text-slate-700 hover:bg-slate-200' },
  { key: 'NO_ANSWER', label: '📵 ما رد', cls: 'bg-amber-50 text-amber-800 hover:bg-amber-100' },
  { key: 'POSTPONED', label: '⏳ أجّل شهر', cls: 'bg-sky-50 text-sky-700 hover:bg-sky-100' },
]
const PAGE = 50

export default function MaintenanceDueList() {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [err, setErr] = useState('')
  const [tab, setTab] = useState<'ALL' | 'LATE' | 'COMPLAINTS'>('LATE')
  const [shown, setShown] = useState(PAGE)
  const [busy, setBusy] = useState<string | null>(null)
  useEffect(() => {
    apiRequest<Row[]>('/maintenance-due').then(setRows).catch((e) => setErr(e.message))
  }, [])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!rows) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  if (rows.length === 0) return <p className="rounded-2xl bg-white p-6 text-center text-slate-500">ماكو زبون يستاهل اتصال صيانة هسه ✅</p>
  const late = rows.filter((r) => r.kind === 'LATE').length
  const list = tab === 'ALL' ? rows : rows.filter((r) => r.kind === tab)
  const decide = async (r: Row, decision: string) => {
    setBusy(r.customerId)
    try {
      await apiRequest(`/maintenance-due/${r.customerId}/decision`, { method: 'POST', body: JSON.stringify({ decision }) })
      setRows((x) => (x ?? []).filter((y) => y.customerId !== r.customerId))
      notify('انحفظ القرار ✅', 'success')
    } catch (e) {
      notify(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setBusy(null)
    }
  }
  const chip = (k: typeof tab, label: string) => (
    <button type="button" onClick={() => { setTab(k); setShown(PAGE) }}
      className={`rounded-full px-3 py-1 text-xs font-bold ${tab === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-600'}`}>{label}</button>
  )
  return (
    <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm" aria-label="صيانة وقائية">
      <p className="text-sm text-slate-600">
        <b>{rows.length}</b> زبون يستاهل اتصال: <b>{late}</b> متأخر صيانة (فات ٦ أشهر) و<b>{rows.length - late}</b> بس شكاوى متكررة.
        اتصل وسجّل القرار — الزبون يختفي من القائمة (ما رد: أسبوع، أجّل: شهر، حجز أو ما يريد: ٦ أشهر).
      </p>
      <div className="flex flex-wrap gap-2">
        {chip('LATE', `متأخر صيانة (${late})`)}
        {chip('COMPLAINTS', `شكاوى متكررة (${rows.length - late})`)}
        {chip('ALL', `الكل (${rows.length})`)}
      </div>
      <ul className="divide-y divide-slate-100 text-sm">
        {list.slice(0, shown).map((r) => (
          <li key={r.customerId} className="space-y-1 py-2">
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
              <b className="text-slate-800">{r.name}</b>
              {r.phone && <a href={`tel:${r.phone}`} className="text-brand-700" dir="ltr">{r.phone}</a>}
              <span className="text-slate-500">{r.service ?? '—'}</span>
              <span className="text-slate-500">آخر شغل قبل {r.months} شهر</span>
              {r.complaints >= 2 && <span className="rounded bg-rose-50 px-2 py-0.5 text-xs font-bold text-rose-700">{r.complaints} شكاوى بالسنة</span>}
            </div>
            <div className="flex flex-wrap gap-1">
              {DECISIONS.map((d) => (
                <button key={d.key} type="button" disabled={busy === r.customerId} onClick={() => void decide(r, d.key)}
                  className={`rounded-lg px-2 py-1 text-[11px] font-bold disabled:opacity-50 ${d.cls}`}>{d.label}</button>
              ))}
            </div>
          </li>
        ))}
      </ul>
      {list.length > shown && (
        <button type="button" onClick={() => setShown((n) => n + PAGE)} className="w-full rounded-lg bg-slate-100 py-2 text-xs font-bold text-slate-700 hover:bg-slate-200">
          المزيد ({list.length - shown} باقي)
        </button>
      )}
    </section>
  )
}
