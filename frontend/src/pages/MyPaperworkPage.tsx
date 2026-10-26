import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type CarryoverItem } from '../api'

// ═══ 📄 «ورق حجوزاتي القديمة» (دازاي 10-26) ═══
// موظف انتقل من ليدر لمنصب ثاني: يكمّل فواتير وتقارير حجوزاته هو. الصلاحية
// مؤقتة وتنشال لحالها من يخلّص كلهن.
export default function MyPaperworkPage() {
  const [rows, setRows] = useState<CarryoverItem[] | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => { api.getCarryover().then(setRows).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])
  return (
    <div dir="rtl" className="mx-auto max-w-3xl space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">📄 ورق حجوزاتي القديمة</h1>
        <p className="mt-1 text-sm text-blue-100">حجوزات اشتغلت بيها وبعدها ناقصها فاتورة أو تقرير. من تخلّصها كلها، هالصفحة تنشال لحالها.</p>
      </div>
      {err && <p className="rounded-xl bg-red-50 p-3 text-red-700">{err}</p>}
      {!rows && !err && <p className="py-10 text-center text-slate-400">جاري التحميل…</p>}
      {rows && rows.length === 0 && (
        <p className="rounded-2xl bg-emerald-50 py-10 text-center text-lg font-bold text-emerald-700">✅ ماكو ورق ناقص — عاشت ايدك</p>
      )}
      {rows && rows.length > 0 && (
        <>
          <p className="text-sm font-bold text-slate-600">باقي {rows.length} حجز</p>
          <ul className="space-y-2">
            {rows.map((b) => (
              <li key={b.id} className="flex flex-wrap items-center gap-2 rounded-2xl bg-white p-4 shadow-sm ring-1 ring-slate-200">
                <b className="font-mono text-[#0f2040]">{b.code}</b>
                {b.completedAt && <span className="text-xs text-slate-400">انجز {b.completedAt}</span>}
                <span className="ms-auto flex flex-wrap gap-2">
                  {b.hasInvoice
                    ? <span className="rounded-lg bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-700">✔ الفاتورة</span>
                    : <Link to={`/leader-invoices/new?bookingId=${b.id}`} className="rounded-lg bg-[#0f2040] px-3 py-1.5 text-xs font-bold text-white">🧾 سوّي الفاتورة</Link>}
                  {b.hasReport
                    ? <span className="rounded-lg bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-700">✔ التقرير</span>
                    : <Link to={`/work-reports?bookingId=${b.id}`} className="rounded-lg bg-sky-600 px-3 py-1.5 text-xs font-bold text-white">📝 سوّي التقرير</Link>}
                </span>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  )
}
