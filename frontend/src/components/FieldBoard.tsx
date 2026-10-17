import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type Booking } from '../api'
import { pollVisible } from '../utils/poll'

// ═══ «بالميدان والمعلّق» — طلب (ع) 10-13 ═══
// من يضغط «الكوادر بالميدان» بلوحة المتابعة: الي طالعين للزبائن هسه، والمتوقفة،
// والإنجاز الجزئي، والمنجزة الناقصها فاتورة/تقرير. نفس محطات الخادم
// (`bucketCondition`) — فالأرقام هنا وبلوحة المتابعة تتفق.

const GROUPS: { bucket: string; title: string; tone: string; empty: string }[] = [
  { bucket: 'in_field', title: '🚗 بالميدان هسه', tone: 'border-sky-200 bg-sky-50/60', empty: 'محد طالع لزبون هسه.' },
  { bucket: 'stopped', title: '⏸️ متوقفة', tone: 'border-red-200 bg-red-50/60', empty: 'ماكو حجز متوقف.' },
  { bucket: 'partial', title: '🔄 إنجاز جزئي — تحتاج موعد إكمال', tone: 'border-amber-200 bg-amber-50/60', empty: 'ماكو إنجاز جزئي.' },
  { bucket: 'needs_paper', title: '🧾 منجزة وناقصها فاتورة أو تقرير', tone: 'border-violet-200 bg-violet-50/60', empty: 'كل المنجز ورقه كامل ✓' },
]

const t = (s: string | null | undefined) => s
  ? new Date(s).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
  : '—'

function Card({ b, bucket }: { b: Booking; bucket: string }) {
  const crew = (b.assignments ?? []).map((a) => a.employee?.name).filter(Boolean).join('، ')
  const missing = [!b.hasInvoice && 'الفاتورة', !b.hasReport && 'التقرير'].filter(Boolean).join(' و')
  return (
    <Link to={`/bookings?q=${encodeURIComponent(b.code)}`} className="block rounded-xl border border-slate-200 bg-white p-3 text-sm transition hover:-translate-y-0.5 hover:shadow-md">
      <div className="flex items-center justify-between gap-2">
        <p className="truncate font-extrabold text-[#0f2040]">{b.customer?.name || '—'}</p>
        <span className="shrink-0 rounded-md bg-slate-100 px-2 py-0.5 text-[11px] text-slate-500">{b.code}</span>
      </div>
      <p className="mt-0.5 truncate text-xs text-slate-500">{b.service?.name || '—'}</p>
      <p className="mt-1 truncate text-xs text-slate-700">👷 {crew || 'بلا كادر مسجّل'}</p>
      <p className="mt-1 text-[11px] text-slate-500">
        {bucket === 'in_field' && <>📍 طلع: {t(b.startedAt || b.arrivedAt)}</>}
        {bucket === 'stopped' && <>⏸️ توقف: {t(b.workStoppedAt)}</>}
        {bucket === 'partial' && <>🔄 آخر إنجاز جزئي: {t(b.lastPartialAt)} · {b.partialCount} مرة</>}
        {bucket === 'needs_paper' && <>🏁 خلص: {t(b.completedAt)} · <b className="text-violet-700">ناقص {missing}</b></>}
      </p>
    </Link>
  )
}

export default function FieldBoard() {
  const [data, setData] = useState<Record<string, { items: Booking[]; total: number } | null>>({})
  useEffect(() => {
    let alive = true
    const load = () => GROUPS.forEach((g) => {
      api.getBookingsPaged({ bucket: g.bucket, pageSize: 60 })
        .then((r) => { if (alive) setData((d) => ({ ...d, [g.bucket]: { items: r.items, total: r.total } })) })
        .catch(() => { if (alive) setData((d) => ({ ...d, [g.bucket]: null })) })
    })
    load()
    const stop = pollVisible(load, 60_000)
    return () => { alive = false; stop() }
  }, [])

  return (
    <div className="space-y-4">
      {GROUPS.map((g) => {
        const d = data[g.bucket]
        return (
          <section key={g.bucket} className={`rounded-2xl border p-4 ${g.tone}`}>
            <h3 className="mb-3 flex items-center justify-between text-base font-extrabold text-[#0f2040]">
              <span>{g.title}</span>
              <span className="rounded-full bg-white px-3 py-0.5 text-sm tabular-nums shadow-sm">{d ? d.total : '…'}</span>
            </h3>
            {d === undefined ? <div className="h-16 animate-pulse rounded-xl bg-white/60" />
              : d === null ? <p className="text-xs text-slate-500">ما وصلت البيانات.</p>
              : d.items.length === 0 ? <p className="text-sm text-slate-500">{g.empty}</p>
              : (
                <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                  {d.items.map((b) => <Card key={b.id} b={b} bucket={g.bucket} />)}
                </div>
              )}
            {d && d.total > d.items.length && <p className="mt-2 text-[11px] text-slate-500">ينعرض أول {d.items.length} من {d.total}.</p>}
          </section>
        )
      })}
    </div>
  )
}
