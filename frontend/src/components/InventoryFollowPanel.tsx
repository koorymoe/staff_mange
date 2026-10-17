import { useEffect, useMemo, useState } from 'react'
import { apiRequest } from '../api'

// ═══ متابعة الجرد الحقيقي — (ع) 10-16 ═══
// الجرد الي يسوّيه الموظف من «جردي» أو قبل الطلعة من «مهامي». كل بطاقة
// تفلتر القائمة تحتها — الرقم = القائمة.
interface Check {
  id: string; employeeId: string; employeeName: string; complete: boolean; missingItems: string | null
  checkedAt: string; bookingId: string | null; bookingCode: string | null; today: boolean
}
interface NotChecked { employeeId: string; name: string; bookings: string }
interface Data { days: number; checks: Check[]; notCheckedToday: NotChecked[] }
type F = 'today' | 'missing' | 'incomplete' | 'all'

const fmt = (s: string) => new Date(s.endsWith('Z') ? s : s + 'Z').toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'short', timeStyle: 'short' })

export default function InventoryFollowPanel() {
  const [d, setD] = useState<Data | null>(null)
  const [err, setErr] = useState('')
  const [f, setF] = useState<F>('today')
  const [q, setQ] = useState('')
  useEffect(() => { apiRequest<Data>('/inventory/checks/follow?days=30').then(setD).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر الجلب')) }, [])

  const lists = useMemo(() => {
    const c = d?.checks ?? []
    return { today: c.filter((x) => x.today), incomplete: c.filter((x) => !x.complete), all: c }
  }, [d])
  if (err) return <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>
  if (!d) return <p className="py-8 text-center text-slate-400">جاري الجلب…</p>

  const tile = (k: F, label: string, n: number, tone: string, hint: string) => (
    <button onClick={() => setF(k)} className={`rounded-2xl border p-4 text-right transition ${tone} ${f === k ? 'ring-2 ring-brand-500' : 'hover:shadow'}`}>
      <p className="text-[12px] font-extrabold">{label}</p>
      <p className="mt-1 text-3xl font-black tabular-nums text-[#0f2040]">{n}</p>
      <p className="text-[11px] text-slate-500">{hint}</p>
    </button>
  )
  const match = (s: string) => !q.trim() || s.includes(q.trim())
  const rows = f === 'missing' ? [] : lists[f].filter((c) => match(`${c.employeeName} ${c.missingItems ?? ''} ${c.bookingCode ?? ''}`))
  const missing = d.notCheckedToday.filter((m) => match(`${m.name} ${m.bookings}`))

  return (
    <div className="space-y-3" dir="rtl">
      <p className="rounded-xl bg-sky-50 px-4 py-2 text-[12.5px] text-sky-900">
        الجرد الي يسوّيه الموظف من «جردي» أو قبل الطلعة من «مهامي». «ما جردوا اليوم» = فنّي أو ليدر عنده حجز موعده اليوم وما سجّل جرد اليوم.
      </p>
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {tile('today', 'جرود اليوم', lists.today.length, 'border-sky-200 bg-sky-50/60', 'سجّلوا جرد اليوم')}
        {tile('missing', 'ما جردوا اليوم', d.notCheckedToday.length, d.notCheckedToday.length ? 'border-red-200 bg-red-50/60' : 'border-emerald-200 bg-emerald-50/50', 'عندهم شغل اليوم وما جردوا')}
        {tile('incomplete', 'جرود ناقصة', lists.incomplete.length, lists.incomplete.length ? 'border-amber-200 bg-amber-50/60' : 'border-slate-200 bg-white', `آخر ${d.days} يوم — بيها أداة ناقصة`)}
        {tile('all', 'كل الجرود', lists.all.length, 'border-slate-200 bg-white', `آخر ${d.days} يوم`)}
      </div>
      <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="بحث بالاسم أو الأداة أو الحجز…" className="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" />
      <div className="rounded-2xl border border-slate-200 bg-white">
        {f === 'missing' ? (
          missing.length === 0 ? <p className="p-6 text-center text-emerald-700">كل الي عنده شغل اليوم جرد ✓</p> : (
            <ul className="divide-y divide-slate-100">
              {missing.map((m) => (
                <li key={m.employeeId} className="flex flex-wrap items-center justify-between gap-2 p-3">
                  <b className="text-red-700">❌ {m.name}</b>
                  <span className="text-xs text-slate-500">حجوزات اليوم: {m.bookings}</span>
                </li>
              ))}
            </ul>
          )
        ) : rows.length === 0 ? <p className="p-6 text-center text-slate-400">ماكو جرود</p> : (
          <ul className="divide-y divide-slate-100">
            {rows.map((c) => (
              <li key={c.id} className="flex flex-wrap items-start justify-between gap-2 p-3">
                <span>
                  <b className="text-slate-800">{c.employeeName}</b>
                  <span className="block text-xs text-slate-400">{fmt(c.checkedAt)}{c.bookingCode ? ` · قبل حجز ${c.bookingCode}` : ' · جرد عام'}</span>
                  {!c.complete && c.missingItems && <span className="mt-1 block text-xs text-amber-800">ناقص: {c.missingItems}</span>}
                </span>
                <span className={`rounded-full px-3 py-1 text-xs font-bold ${c.complete ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-100 text-amber-800'}`}>{c.complete ? '✓ مكتمل' : '⚠ ناقص'}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
