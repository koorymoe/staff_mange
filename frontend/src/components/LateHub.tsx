import { lazy, Suspense, useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { askJustification } from '../utils/justify'

// ═══ «⏰ المتأخر» — خانة وحدة لكل تأخير (ع 10-16) ═══
// «عروض أسعار المشاريع، وتأخير المشاريع، والمتأخرات — كلهن بخانة وحدة اسمها
// المتأخر. يدخل ويختار وين يريد: المشاريع، الحجوزات، الفواتير، المحاسب».
// وعلى كل متأخر: «📝 اطلب تفسير» يوصل للموظف نفسه، وجوابه يرجع للمراقب
// بتبويب «التبريرات» بالصندوق، ويحكم عليه (يتسجّل تقصير لو انرفض).
const ProjectQuoteWatchPanel = lazy(() => import('./ProjectQuoteWatchPanel'))
const ProjectDelaysPanel = lazy(() => import('./ProjectDelaysPanel'))
const LateAuditPage = lazy(() => import('../pages/LateAuditPage'))

type Area = 'projects' | 'bookings' | 'invoices'
const AREAS: { key: Area; label: string; hint: string }[] = [
  { key: 'projects', label: '🏗️ المشاريع', hint: 'منو استلم وما تواصل، وعروض أسعار متأخرة، ومراحل تجاوزت حدها' },
  { key: 'bookings', label: '📋 الحجوزات', hint: 'حجوزات تأخرت — ذنب منو وليش' },
  { key: 'invoices', label: '🧾 الفواتير والمحاسب', hint: 'فواتير ما تدققت خلال ٢٤ ساعة — المحاسب متأخر عليها' },
]


export default function LateHub() {
  const [area, setArea] = useState<Area>('projects')
  const cur = AREAS.find((a) => a.key === area)!
  return (
    <div className="space-y-3" dir="rtl">
      <div className="flex flex-wrap gap-2">
        {AREAS.map((a) => (
          <button key={a.key} onClick={() => setArea(a.key)}
            className={`rounded-xl border px-4 py-2 text-sm font-bold ${area === a.key ? 'border-transparent bg-[#0f2040] text-white' : 'border-slate-200 bg-white text-slate-600'}`}>
            {a.label}
          </button>
        ))}
      </div>
      <p className="text-[12px] text-slate-500">{cur.hint}</p>
      <Suspense fallback={<p className="py-10 text-center text-slate-400">جاري التحميل…</p>}>
        {area === 'projects' && (
          <div className="space-y-6">
            <ProjectQuoteWatchPanel />
            <div className="border-t-2 border-slate-200 pt-4">
              <h3 className="mb-2 text-sm font-extrabold text-[#0f2040]">⏱️ مراحل المشاريع الي تجاوزت حدها</h3>
              <ProjectDelaysPanel />
            </div>
          </div>
        )}
        {area === 'bookings' && <LateAuditPage />}
        {area === 'invoices' && <LateInvoices />}
      </Suspense>
    </div>
  )
}

interface LateInv { id: string; accountingCode: string; bookingCode: string | null; customerName: string | null; leaderName: string | null; netTotal: number; createdAt: string; waitHours: number }
function LateInvoices() {
  const [d, setD] = useState<{ hours: number; rows: LateInv[]; accountants: { id: string; name: string }[] } | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => { apiRequest<typeof d>('/late-invoices').then(setD).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر الجلب')) }, [])
  if (err) return <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>
  if (!d) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  const ask = async () => {
    if (d.accountants.length === 0) { alert('ماكو محاسب فعّال بالنظام'); return }
    const acc = d.accountants.length === 1 ? d.accountants[0]
      : d.accountants.find((a) => a.name === prompt(`لمنو؟ ${d.accountants.map((a) => a.name).join(' / ')}`)) 
    if (!acc) return
    try { await askJustification(acc.id, `عندك ${d.rows.length} فاتورة ما تدققت من أكثر من ${d.hours} ساعة — ليش؟`) }
    catch (e) { alert(e instanceof Error ? e.message : 'تعذر الطلب') }
  }
  return (
    <div className="space-y-3">
      <div className={`flex flex-wrap items-center justify-between gap-2 rounded-2xl border p-4 ${d.rows.length ? 'border-red-200 bg-red-50/60' : 'border-emerald-200 bg-emerald-50/50'}`}>
        <p className="text-sm font-bold">{d.rows.length ? `🔴 ${d.rows.length} فاتورة ما تدققت من أكثر من ${d.hours} ساعة — المحاسب متأخر عليها` : `🟢 المحاسب ماشي بالوقت — ماكو فاتورة تجاوزت ${d.hours} ساعة`}</p>
        {d.rows.length > 0 && <button onClick={ask} className="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-bold text-white">📝 اطلب تفسير من المحاسب</button>}
      </div>
      {d.rows.length > 0 && (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="w-full text-right text-[13px]">
            <thead className="bg-slate-50 text-[11px] text-slate-500"><tr><th className="p-2">الفاتورة</th><th className="p-2">الحجز</th><th className="p-2">الزبون</th><th className="p-2">الليدر</th><th className="p-2">الصافي</th><th className="p-2">صارلها</th></tr></thead>
            <tbody>{d.rows.map((r) => (
              <tr key={r.id} className="border-t border-slate-100">
                <td className="p-2 font-mono text-xs">{r.accountingCode}</td><td className="p-2">{r.bookingCode ?? '—'}</td>
                <td className="p-2">{r.customerName ?? '—'}</td><td className="p-2">{r.leaderName ?? '—'}</td>
                <td className="p-2 tabular-nums">{Math.round(r.netTotal).toLocaleString('en-US')} د.ع</td>
                <td className="p-2 font-bold text-red-700">{r.waitHours < 48 ? `${Math.round(r.waitHours)} ساعة` : `${(r.waitHours / 24).toFixed(1)} يوم`}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}
    </div>
  )
}
