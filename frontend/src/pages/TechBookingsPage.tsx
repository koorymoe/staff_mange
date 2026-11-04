import { useEffect, useMemo, useState } from 'react'
import { api, type Booking } from '../api'
import BookingsList from './BookingsList'

// ═══ حجوزات التقنيين — للمالك والمدير بوحدة التقنيين (قرار (ع) 10-09) ═══
// كل حجز رحّله الإداري لتقني: وين وصل (تواصل · كشف · قرار · زيارة · خلص)
// حتى نتابع شغل التقنيين. للاطلاع بس — الشغل نفسه يمّ التقني.

const STATUS: Record<Booking['status'], [string, string]> = {
  PENDING: ['بالانتظار', 'bg-slate-100 text-slate-700'],
  CONFIRMED: ['مؤكد', 'bg-sky-100 text-sky-800'],
  IN_PROGRESS: ['قيد العمل', 'bg-amber-100 text-amber-800'],
  WAITING: ['معلّق', 'bg-orange-100 text-orange-800'],
  PARTIAL: ['جزئي', 'bg-violet-100 text-violet-800'],
  COMPLETED: ['خلص', 'bg-emerald-100 text-emerald-800'],
  CANCELLED: ['ملغي', 'bg-red-100 text-red-700'],
}
const d = (s?: string | null) => (s ? new Date(s).toLocaleString('ar-IQ', { month: 'numeric', day: 'numeric', hour: 'numeric', minute: '2-digit' }) : '')

function stage(b: Booking): string {
  if (b.status === 'COMPLETED') return '✅ خلص'
  if (b.techVisitedAt) return '🚗 زار الزبون'
  if (b.techVisitAt) return `📅 زيارة ${d(b.techVisitAt)}`
  if (b.techDecision) return b.techDecision === 'PHONE' ? '📞 قرر: بالتلفون' : '🚗 قرر: زيارة'
  if (b.techDiagnosedAt) return '📝 كتب الكشف'
  if (b.techContactedAt) return '☎️ تواصل ويا الزبون'
  return '⏳ ما بدأ'
}

// قرار (ع) 10-09: «نفس ترتيب الحجوزات العادية» — طلب حذف، تفاصيل، تواصل ويا
// الزبون، والمعلومات تنحفظ. فالتبويب الأول هو قائمة الحجوزات نفسها على محطة
// «عند إدارة التقنيين»، والثاني متابعة كل تقني.
export default function TechBookingsPage() {
  const [view, setView] = useState<'list' | 'follow'>('list')
  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {([['list', '📋 الحجوزات'], ['follow', '📊 متابعة التقنيين']] as const).map(([k, l]) => (
          <button key={k} onClick={() => setView(k)}
            className={`rounded-xl border px-4 py-2 text-sm font-bold ${view === k ? 'border-[#2c5aad] bg-[#2c5aad] text-white' : 'border-slate-200 bg-white text-slate-600'}`}>{l}</button>
        ))}
      </div>
      {view === 'list' ? <BookingsList bucket="at_tech" /> : <TechFollow />}
    </div>
  )
}

function TechFollow() {
  const [rows, setRows] = useState<Booking[] | null>(null)
  const [err, setErr] = useState('')
  const [tech, setTech] = useState('')
  const [open, setOpen] = useState<'open' | 'done' | 'all'>('open')

  useEffect(() => {
    let alive = true
    api.getTechHandovers().then((r) => { if (alive) setRows(r) }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر التحميل') })
    return () => { alive = false }
  }, [])

  const techs = useMemo(() => {
    const m = new Map<string, { name: string; open: number; done: number }>()
    for (const b of rows ?? []) {
      const id = b.handoverToId ?? ''
      const t = m.get(id) ?? { name: b.handoverTo?.name ?? '—', open: 0, done: 0 }
      if (b.status === 'COMPLETED') t.done++
      else if (b.status !== 'CANCELLED') t.open++
      m.set(id, t)
    }
    return [...m.entries()]
  }, [rows])

  const shown = (rows ?? []).filter((b) => (!tech || b.handoverToId === tech)
    && (open === 'all' || (open === 'done' ? b.status === 'COMPLETED' : b.status !== 'COMPLETED' && b.status !== 'CANCELLED')))

  return (
    <div dir="rtl" className="space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-extrabold">🛠️ حجوزات التقنيين</h1>
        <p className="mt-1 text-sm text-blue-100">كل حجز رحّله الإداري لتقني (آخر ٩٠ يوم) — وين وصل بيه.</p>
      </div>
      {err && <p className="rounded-xl bg-red-50 px-4 py-2 text-sm font-bold text-red-700">{err}</p>}
      {!rows && !err && <p className="text-slate-400">جاري التحميل…</p>}
      {rows && (
        <>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {techs.map(([id, t]) => (
              <button key={id} onClick={() => setTech(tech === id ? '' : id)}
                className={`rounded-2xl border p-4 text-right shadow-sm ${tech === id ? 'border-[#2c5aad] bg-blue-50' : 'border-slate-200 bg-white'}`}>
                <p className="font-extrabold text-[#0f2040]">👤 {t.name}</p>
                <p className="mt-1 text-xs text-slate-600"><b className="text-amber-700">{t.open}</b> مفتوح · <b className="text-emerald-700">{t.done}</b> خلص</p>
              </button>
            ))}
          </div>
          <div className="flex flex-wrap gap-2">
            {([['open', 'المفتوحة'], ['done', 'الي خلصت'], ['all', 'الكل']] as const).map(([k, l]) => (
              <button key={k} onClick={() => setOpen(k)}
                className={`rounded-xl border px-4 py-2 text-sm font-bold ${open === k ? 'border-[#2c5aad] bg-[#2c5aad] text-white' : 'border-slate-200 bg-white text-slate-600'}`}>{l}</button>
            ))}
          </div>
          <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-sm">
            <table className="w-full text-right text-sm">
              <thead className="bg-slate-50 text-xs text-slate-500">
                <tr>
                  <th className="px-3 py-2.5">الحجز</th><th className="px-3 py-2.5">الزبون</th><th className="px-3 py-2.5">التقني</th>
                  <th className="px-3 py-2.5">رحّله</th><th className="px-3 py-2.5">وين وصل</th><th className="px-3 py-2.5">الحالة</th><th className="px-3 py-2.5">الكشف</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {shown.length === 0 && <tr><td colSpan={7} className="p-6 text-center text-slate-400">ماكو حجوزات.</td></tr>}
                {shown.map((b) => {
                  const [label, cls] = STATUS[b.status] ?? [b.status, 'bg-slate-100']
                  return (
                    <tr key={b.id}>
                      <td className="whitespace-nowrap px-3 py-2.5 font-bold">{b.code}</td>
                      <td className="px-3 py-2.5">{b.customer?.name ?? ''}</td>
                      <td className="whitespace-nowrap px-3 py-2.5">{b.handoverTo?.name ?? '—'}</td>
                      <td className="whitespace-nowrap px-3 py-2.5 text-xs text-slate-500">{b.handoverBy?.name ?? ''} · {d(b.handoverAt)}{b.handoverReason ? <span className="block text-slate-400">{b.handoverReason}</span> : null}</td>
                      <td className="whitespace-nowrap px-3 py-2.5 text-xs font-bold">{stage(b)}</td>
                      <td className="px-3 py-2.5"><span className={`whitespace-nowrap rounded-lg px-2 py-0.5 text-xs font-bold ${cls}`}>{label}</span></td>
                      <td className="max-w-[18rem] px-3 py-2.5 text-xs text-slate-600">{b.techDiagnosis ?? ''}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
