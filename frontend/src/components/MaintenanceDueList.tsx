import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { notify } from '../ui/notify'
import { fmtDateTime } from '../utils/date'

// 🛠️ صيانة وقائية (دازاي 11-06، فكرة ٧): زبائن فات ٦ أشهر على آخر شغلهم، أو عندهم شكاوى متكررة،
// وماكو حجز مفتوح. الجودة تتصل وتعرض صيانة.
// (بلاغ #55، 11-20) «ماكو قرار اتخذه كذالك بالعدد»: أزرار قرار لكل زبون (يختفي لمدة)،
// العدد الحقيقي مقسوم (متأخر صيانة / شكاوى)، والقائمة ٥٠ ٥٠.
type Row = { customerId: string; name: string; phone?: string | null; lastDone: string; service: string | null; months: number; complaints: number; lastComplaint?: string | null; reason: string; kind: 'LATE' | 'COMPLAINTS' }

const DECISIONS: { key: string; label: string; cls: string }[] = [
  { key: 'BOOKED', label: '✅ حجز صيانة', cls: 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100' },
  { key: 'DECLINED', label: '🚫 ما يريد', cls: 'bg-slate-100 text-slate-700 hover:bg-slate-200' },
  { key: 'NO_ANSWER', label: '📵 ما رد', cls: 'bg-amber-50 text-amber-800 hover:bg-amber-100' },
  { key: 'POSTPONED', label: '⏳ أجّل شهر', cls: 'bg-sky-50 text-sky-700 hover:bg-sky-100' },
]
const PAGE = 20

export default function MaintenanceDueList() {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [err, setErr] = useState('')
  const [tab, setTab] = useState<'ALL' | 'LATE' | 'COMPLAINTS'>('LATE')
  const [page, setPage] = useState(1)
  const [busy, setBusy] = useState<string | null>(null)
  useEffect(() => {
    apiRequest<Row[]>('/maintenance-due').then(setRows).catch((e) => setErr(e.message))
  }, [])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!rows) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  if (rows.length === 0) return <p className="rounded-2xl bg-white p-6 text-center text-slate-500">ماكو زبون يستاهل اتصال صيانة هسه ✅</p>
  const late = rows.filter((r) => r.kind === 'LATE').length
  const list = tab === 'ALL' ? rows : rows.filter((r) => r.kind === tab)
  const pages = Math.max(1, Math.ceil(list.length / PAGE))
  const cur = Math.min(page, pages)
  const pageRows = list.slice((cur - 1) * PAGE, cur * PAGE)
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
    <button type="button" onClick={() => { setTab(k); setPage(1) }}
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
      {/* (ك) 11-21 زينا: جدول مثل شاشة المستخدمين — أعمدة، وأزرار القرار بآخر كل صف، وترقيم صفحات */}
      <div className="overflow-x-auto rounded-xl border border-slate-200" data-testid="maint-table">
        <table className="w-full min-w-[900px] text-sm">
          <thead className="bg-slate-50 text-slate-600">
            <tr>
              {['ت', 'الاسم الكامل', 'رقم الموبايل', 'نوع الخدمة', 'آخر شغل', 'الشكاوى', 'تاريخ الشكوى', 'الإجراءات'].map((h) => (
                <th key={h} className="whitespace-nowrap border-s border-slate-200 px-1.5 py-3.5 text-center text-sm font-bold first:border-s-0">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {pageRows.map((r, i) => (
              <tr key={r.customerId} className="align-middle hover:bg-slate-50 [&>td]:whitespace-nowrap [&>td]:border-s [&>td]:border-slate-200 [&>td:first-child]:border-s-0">
                <td className="px-1.5 py-3.5 text-center text-slate-500">{(cur - 1) * PAGE + i + 1}</td>
                <td className="px-1.5 py-3.5 text-start font-bold text-slate-800">{r.name}</td>
                <td className="px-1.5 py-3.5 text-center">{r.phone ? <a href={`tel:${r.phone}`} className="text-brand-700" dir="ltr">{r.phone}</a> : <span className="text-slate-400">—</span>}</td>
                <td className="px-1.5 py-3.5 text-center text-slate-600">{r.service ?? '—'}</td>
                <td className="px-1.5 py-3.5 text-center text-slate-600">آخر شغل قبل {r.months} شهر</td>
                <td className="px-1.5 py-3.5 text-center">
                  {r.complaints > 0
                    ? <span className="whitespace-nowrap rounded-full bg-rose-50 px-3 py-1 text-xs font-bold text-rose-700">{r.complaints} شكاوى بالسنة</span>
                    : <span className="text-slate-400">—</span>}
</td>
                {/* (ك) 11-21 زينا: «تاريخ الشكوى» — يوم وساعة آخر شكوى للزبون، بعمود واضح */}
                <td className="px-1.5 py-3.5 text-center text-slate-600">{r.complaints > 0 && r.lastComplaint ? fmtDateTime(r.lastComplaint) : <span className="text-slate-400">—</span>}</td>
                <td className="px-1.5 py-3.5 text-center">
                  <div className="flex flex-nowrap justify-center gap-1">
                    {[...DECISIONS].reverse().map((d) => (
                      <button key={d.key} type="button" disabled={busy === r.customerId} onClick={() => void decide(r, d.key)}
                        className={`whitespace-nowrap rounded-full px-2.5 py-1 text-[11px] font-bold disabled:opacity-50 ${d.cls}`}>{d.label}</button>
                    ))}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-2 px-1 pt-2">
        <span className="text-sm text-slate-500">عرض {pageRows.length} من {list.length} نتيجة</span>
        {pages > 1 && (
          <div className="flex items-center gap-1" dir="ltr">
            <button type="button" aria-label="السابقة" disabled={cur === 1} onClick={() => setPage(cur - 1)} className="h-10 w-12 rounded-xl bg-slate-100 text-lg text-slate-700 disabled:opacity-40">‹</button>
            {Array.from({ length: pages }, (_, i) => i + 1).filter((n) => n === 1 || n === pages || Math.abs(n - cur) <= 1).map((n, i, arr) => (
              <span key={n} className="flex items-center gap-1">
                {i > 0 && n - arr[i - 1] > 1 && <span className="text-slate-400">…</span>}
                <button type="button" onClick={() => setPage(n)} aria-current={n === cur ? 'page' : undefined}
                  className={`h-10 min-w-10 rounded-xl px-3 text-base font-bold ${n === cur ? 'bg-[#1e40af] text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'}`}>{n}</button>
              </span>
            ))}
            <button type="button" aria-label="التالية" disabled={cur === pages} onClick={() => setPage(cur + 1)} className="h-10 w-12 rounded-xl bg-slate-100 text-lg text-slate-700 disabled:opacity-40">›</button>
          </div>
        )}
      </div>
    </section>
  )
}
