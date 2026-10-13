import { useEffect, useMemo, useState } from 'react'
import { api, type QualityFollowUp } from '../api'
import BookingCodeChip from './BookingCodeChip'

// ═══ «شنو سوّى مهندس الجودة» — (ع) 10-13 ═══
// بمكتب المراقب: مو شاشة شغل الجودة (أزرار التقرير)، تدقيق بس — منو تواصل
// ويا الزبون ووكتيش، وشنو التقرير الي أخذه منه، والكشف إذا صار.
const REPORT: Record<string, { t: string; c: string }> = {
  POSITIVE: { t: '✅ تقرير إيجابي', c: 'bg-emerald-100 text-emerald-800' },
  NEGATIVE: { t: '⛔ تقرير سلبي', c: 'bg-red-100 text-red-800' },
  STUBBORN: { t: '🗿 الزبون مارد', c: 'bg-slate-200 text-slate-700' },
}
const STATUS: Record<string, string> = {
  PENDING: '⏳ ينتظر التواصل', CONTACTED_OK: 'تواصل — تمام', CONTACTED_ISSUE: 'تواصل — أكو مشكلة',
  CONVERTED: 'انحوّلت صيانة', RECONTACT: 'يعيد التواصل', CLOSED: 'مقفولة',
}
const dt = (s: string | null) => s ? new Date(s).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }) : '—'

export default function QualityAuditView() {
  const [items, setItems] = useState<QualityFollowUp[] | null>(null)
  const [err, setErr] = useState('')
  const [tab, setTab] = useState<'done' | 'pending'>('done')
  const [q, setQ] = useState('')
  useEffect(() => { api.getQualityFollowUps().then(setItems).catch((e) => { setErr(e instanceof Error ? e.message : 'تعذر'); setItems([]) }) }, [])

  const done = useMemo(() => (items ?? []).filter((i) => i.status !== 'PENDING')
    .sort((a, b) => new Date(b.contactedAt || b.createdAt).getTime() - new Date(a.contactedAt || a.createdAt).getTime()), [items])
  const pending = useMemo(() => (items ?? []).filter((i) => i.status === 'PENDING'), [items])
  const list = (tab === 'done' ? done : pending).filter((i) => !q || `${i.customer?.name} ${i.customer?.phone} ${i.booking?.code} ${i.contactedByEmployee?.name}`.includes(q))
  const byQe = useMemo(() => {
    const m = new Map<string, number>()
    for (const i of done) { const n = i.contactedByEmployee?.name || '—'; m.set(n, (m.get(n) ?? 0) + 1) }
    return [...m.entries()]
  }, [done])

  if (!items) return <p className="py-6 text-center text-slate-400">جاري التحميل…</p>
  return (
    <section className="space-y-3 rounded-2xl border border-slate-200 bg-white p-4" dir="rtl">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-lg font-extrabold text-[#0f2040]">⭐ متابعة الجودة — شنو سوّى مهندس الجودة</h3>
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="ابحث بالزبون، الهاتف، الحجز، المهندس…" className="w-64 rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      <div className="flex flex-wrap items-center gap-2">
        <button onClick={() => setTab('done')} className={`rounded-lg px-4 py-2 text-sm font-bold ${tab === 'done' ? 'bg-[#0f2040] text-white' : 'border border-slate-300 text-slate-600'}`}>📞 تواصل وأخذ تقرير ({done.length})</button>
        <button onClick={() => setTab('pending')} className={`rounded-lg px-4 py-2 text-sm font-bold ${tab === 'pending' ? 'bg-amber-600 text-white' : 'border border-slate-300 text-slate-600'}`}>⏳ لسه ما تواصل ({pending.length})</button>
        {byQe.length > 0 && <span className="text-xs text-slate-500">· {byQe.map(([n, c]) => `${n}: ${c}`).join(' · ')}</span>}
      </div>
      {list.length === 0 ? <p className="py-6 text-center text-sm text-slate-500">ماكو شي هنا.</p> : (
        <div className="space-y-2">
          {list.map((i) => (
            <div key={i.id} className="rounded-xl border border-slate-200 p-3 text-sm">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="font-extrabold text-[#0f2040]">{i.customer?.name || '—'} <span className="text-xs font-normal text-slate-500">· {i.customer?.phone}</span></p>
                <div className="flex items-center gap-2">
                  {i.booking?.code && <BookingCodeChip code={i.booking.code} />}
                  <span className="rounded-full bg-slate-100 px-2.5 py-0.5 text-[11px] font-bold text-slate-600">{STATUS[i.status] || i.status}</span>
                  {i.reportType && <span className={`rounded-full px-2.5 py-0.5 text-[11px] font-bold ${REPORT[i.reportType]?.c}`}>{REPORT[i.reportType]?.t}</span>}
                </div>
              </div>
              {i.status === 'PENDING' ? (
                <p className="mt-1 text-xs text-amber-700">⏳ ينتظر من {dt(i.createdAt)} — محد تواصل بعد</p>
              ) : (
                <>
                  <p className="mt-1 text-xs text-slate-600">👷 مهندس الجودة: <b>{i.contactedByEmployee?.name || '—'}</b> · تواصل: <b>{dt(i.contactedAt)}</b></p>
                  <p className="mt-1.5 rounded-lg bg-sky-50 px-3 py-2 text-slate-800">💬 تقرير الزبون: {i.contactNotes?.trim() ? `«${i.contactNotes}»` : <span className="text-slate-400">ما كتب ملاحظات</span>}</p>
                </>
              )}
              {i.inspectionStatus !== 'NONE' && (
                <p className="mt-1.5 text-xs text-slate-600">🔍 الكشف: {i.inspectionStatus === 'PENDING' ? 'ينتظر' : i.inspectionResult === 'CUSTOMER_RIGHT' ? 'الزبون على حق' : 'الزبون مو على حق'}
                  {i.inspectedBy ? ` · ${i.inspectedBy.name}` : ''}{i.inspectionNotes ? ` — «${i.inspectionNotes}»` : ''}</p>
              )}
              {i.penalizedEmployee && <p className="mt-1 text-xs font-bold text-red-700">📉 انخصمت نقطة من: {i.penalizedEmployee.name}</p>}
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
