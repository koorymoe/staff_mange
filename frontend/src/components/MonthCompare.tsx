import { useEffect, useState } from 'react'
import { api, type MonthStats } from '../api'

// ═══ مقارنة الأشهر — قرار (ع) 10-10 ═══
// «شهر التاسع جان بي ٢٠ مشروع و٢٠٠ حجز والمبالغ ٤٠٠ مليون، وشهر العاشر…».
// الأرقام من الخادم بنفس تعريفها بباقي النظام، والفرق والنسبة تنحسب هنا.

const ym = (d: Date) => d.toISOString().slice(0, 7)
const shift = (m: string, n: number) => { const [y, mo] = m.split('-').map(Number); return ym(new Date(Date.UTC(y, mo - 1 + n, 1))) }
const AR_MONTH = (m: string) => new Date(m + '-01').toLocaleDateString('ar-IQ', { month: 'long', year: 'numeric' })
const fmt = (n: number) => Math.round(n).toLocaleString('en-US')
const thisMonth = () => ym(new Date(Date.now() + 3 * 3600_000))
const money = (n: number) => `${fmt(Math.round(n / 1000) * 1000)} د.ع`

// good: هل الزيادة شي زين؟ (الإلغاء والشكاوى والمصاريف زيادتها مو زينة)
const ROWS: { key: keyof MonthStats; label: string; money?: boolean; good: boolean }[] = [
  { key: 'bookingsCreated', label: '📥 حجوزات انفتحت', good: true },
  { key: 'bookingsCompleted', label: '✅ حجوزات انجزت', good: true },
  { key: 'bookingsCancelled', label: '✖ حجوزات انلغت', good: false },
  { key: 'projectsCreated', label: '🏗️ مشاريع جديدة', good: true },
  { key: 'internalWorks', label: '🏭 أعمال داخل الشركة', good: true },
  { key: 'newCustomers', label: '👤 زبائن جدد', good: true },
  { key: 'complaints', label: '⚠️ شكاوى', good: false },
  { key: 'bookingsRevenue', label: '💳 فلوس الحجوزات', money: true, good: true },
  { key: 'projectsRevenue', label: '🏗️ فلوس المشاريع', money: true, good: true },
  { key: 'expenses', label: '🧾 المصاريف', money: true, good: false },
]

export default function MonthCompare() {
  const [now] = useState(thisMonth)
  const [a, setA] = useState(() => shift(thisMonth(), -1))
  const [b, setB] = useState(now)
  const [rows, setRows] = useState<MonthStats[] | null>(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    let alive = true
    api.getMonthCompare([a, b]).then((r) => { if (alive) { setRows(r); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [a, b])

  const A = rows?.find((r) => r.month === a)
  const B = rows?.find((r) => r.month === b)
  const totalA = A ? A.bookingsRevenue + A.projectsRevenue : 0
  const totalB = B ? B.bookingsRevenue + B.projectsRevenue : 0

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center gap-2 rounded-2xl bg-white p-4 shadow-sm">
        <span className="text-sm font-bold text-slate-600">قارن</span>
        <input type="month" value={a} onChange={(e) => e.target.value && setA(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
        <span className="text-sm font-bold text-slate-600">ويا</span>
        <input type="month" value={b} onChange={(e) => e.target.value && setB(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
      </div>
      {err && <p className="text-sm font-bold text-red-600">{err}</p>}
      {!rows && !err && <p className="text-slate-400">جاري الحساب…</p>}
      {A && B && (
        <>
          <div className="rounded-2xl border border-violet-200 bg-violet-50/60 p-4 text-sm leading-relaxed text-violet-950">
            <b>🤖 الخلاصة:</b> {AR_MONTH(a)} جان بي {fmt(A.projectsCreated)} مشروع و{fmt(A.bookingsCreated)} حجز وفلوس {money(totalA)}.
            {' '}أما {AR_MONTH(b)} فبي {fmt(B.projectsCreated)} مشروع و{fmt(B.bookingsCreated)} حجز وفلوس {money(totalB)}
            {totalA > 0 && <> — الفلوس {totalB >= totalA ? 'زادت' : 'نزلت'} {Math.abs(Math.round(((totalB - totalA) / totalA) * 100))}%</>}.
            {b === now && <span className="block text-xs text-violet-700">⏳ {AR_MONTH(b)} بعده ما خلص، فأرقامه لحد اليوم.</span>}
          </div>
          <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-sm">
            <table className="w-full min-w-[560px] text-right text-sm">
              <thead className="bg-slate-50 text-xs text-slate-500">
                <tr><th className="p-3">البند</th><th className="p-3">{AR_MONTH(a)}</th><th className="p-3">{AR_MONTH(b)}</th><th className="p-3">الفرق</th></tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {ROWS.map((r) => {
                  const va = A[r.key] as number
                  const vb = B[r.key] as number
                  const diff = vb - va
                  const up = diff > 0
                  const better = diff === 0 ? null : up === r.good
                  const pctTxt = va > 0 ? ` (${up ? '+' : ''}${Math.round((diff / va) * 100)}%)` : ''
                  return (
                    <tr key={r.key}>
                      <td className="p-3 text-slate-700">{r.label}</td>
                      <td className="p-3 font-bold">{r.money ? money(va) : fmt(va)}</td>
                      <td className="p-3 font-bold">{r.money ? money(vb) : fmt(vb)}</td>
                      <td className={`p-3 text-xs font-extrabold ${better == null ? 'text-slate-400' : better ? 'text-emerald-700' : 'text-red-600'}`}>
                        {diff === 0 ? 'نفسه' : `${up ? '▲' : '▼'} ${r.money ? money(Math.abs(diff)) : fmt(Math.abs(diff))}${pctTxt}`}
                      </td>
                    </tr>
                  )
                })}
                <tr className="bg-slate-50">
                  <td className="p-3 font-extrabold">💰 مجموع الفلوس</td>
                  <td className="p-3 font-extrabold">{money(totalA)}</td>
                  <td className="p-3 font-extrabold">{money(totalB)}</td>
                  <td className={`p-3 text-xs font-extrabold ${totalB >= totalA ? 'text-emerald-700' : 'text-red-600'}`}>
                    {totalB === totalA ? 'نفسه' : `${totalB > totalA ? '▲' : '▼'} ${money(Math.abs(totalB - totalA))}`}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
