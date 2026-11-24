import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { notify } from '../ui/notify'
import { fmtDateTime } from '../utils/date'

// ═══ 🔧 الخطأ الفني والتصحيح (ل) 11-24 دازاي ═══
// الكادر الي صلّح غلط كادر سابق يسجّل هنا (حجز التصحيح + الحجز الأصلي + شنو الخطأ)، والمراقب يوافق.
// بالموافقة: ساعات حجز التصحيح × سعر ساعة الي صلّح ← ينضاف للي صلّح وينخصم من الكادر الأول.

type Item = {
  id: string; fixCode: string; origCode: string; note: string; status: 'PENDING' | 'APPROVED' | 'REJECTED'
  hours: number | null; amount: number | null; by: string; createdAt: string; decidedBy: string | null; decisionNote: string | null; fixCrew: string; origCrew: string
}
type Resp = { items: Item[]; canDecide: boolean }
const ST = { PENDING: ['⏳ ينتظر المراقب', 'bg-amber-100 text-amber-800'], APPROVED: ['✅ انوافق', 'bg-emerald-100 text-emerald-800'], REJECTED: ['🚫 انرفض', 'bg-slate-200 text-slate-600'] } as const
const inp = 'rounded-lg border border-slate-300 p-2 text-sm'

export default function ReworkPage() {
  const [d, setD] = useState<Resp | null>(null)
  const [err, setErr] = useState('')
  const [f, setF] = useState({ fixCode: '', origCode: '', note: '' })
  const [busy, setBusy] = useState('')
  useEffect(() => { apiRequest<Resp>('/rework').then(setD).catch((e: Error) => setErr(e.message)) }, [])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!d) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  const submit = async () => {
    setBusy('new')
    try { setD(await apiRequest<Resp>('/rework', { method: 'POST', body: JSON.stringify(f) })); setF({ fixCode: '', origCode: '', note: '' }); notify('انسجّل — ينتظر موافقة المراقب ✅', 'success') } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy('') }
  }
  const decide = async (id: string, approve: boolean) => {
    setBusy(id)
    try { setD(await apiRequest<Resp>(`/rework/${id}/decide`, { method: 'POST', body: JSON.stringify({ approve }) })); notify(approve ? 'انوافق وانتقل المبلغ ✅' : 'انرفض', 'success') } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy('') }
  }
  return (
    <div className="mx-auto max-w-5xl space-y-4" data-testid="rework">
      <header>
        <h1 className="text-2xl font-extrabold text-[#0f2040]">🔧 الخطأ الفني والتصحيح</h1>
        <p className="text-sm text-slate-500">إذا رجعتوا لزبون تصلّحون غلط كادر سابق: سجّلوه هنا. المراقب يوافق، ووقت التصحيح (من الحجز) × سعر ساعتكم ينضاف إلكم وينخصم من الكادر الأول.</p>
      </header>
      <section className="grid gap-2 rounded-2xl bg-white p-4 shadow-sm sm:grid-cols-[10rem_10rem_1fr_auto]">
        <input value={f.fixCode} onChange={(e) => setF({ ...f, fixCode: e.target.value })} placeholder="حجز التصحيح (B…)" className={inp} dir="ltr" />
        <input value={f.origCode} onChange={(e) => setF({ ...f, origCode: e.target.value })} placeholder="الحجز الأصلي (B…)" className={inp} dir="ltr" />
        <input value={f.note} onChange={(e) => setF({ ...f, note: e.target.value })} placeholder="شنو چان الخطأ الفني وشنو صلّحتوا" className={inp} />
        <button type="button" disabled={busy === 'new' || !f.fixCode || !f.origCode || f.note.trim().length < 5} onClick={() => void submit()} className="rounded-lg bg-[#0f2040] px-4 py-2 text-sm font-bold text-white disabled:opacity-40">سجّل</button>
      </section>
      <section className="space-y-2">
        {d.items.length === 0 && <p className="rounded-2xl bg-white p-6 text-center text-slate-500">ماكو تصحيحات بعد.</p>}
        {d.items.map((x) => (
          <div key={x.id} className="space-y-1 rounded-2xl bg-white p-4 shadow-sm">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <b className="text-[#0f2040]"><bdi dir="ltr">{x.fixCode}</bdi> صلّح <bdi dir="ltr">{x.origCode}</bdi></b>
              <span className={`rounded-full px-3 py-0.5 text-xs font-bold ${ST[x.status][1]}`}>{ST[x.status][0]}</span>
            </div>
            <p className="text-sm text-slate-700">{x.note}</p>
            <p className="text-xs text-slate-500">الي صلّحوا: {x.fixCrew || '—'} · الكادر الأول: {x.origCrew || '—'} · سجّله {x.by} · {fmtDateTime(x.createdAt)}</p>
            {x.status === 'APPROVED' && <p className="text-sm">⏱️ {x.hours} ساعة · 💰 <b dir="ltr">{Math.round(x.amount ?? 0).toLocaleString('en-US')}</b> د.ع انتقلت{x.decidedBy ? ` — وافق ${x.decidedBy}` : ''}</p>}
            {x.decisionNote && <p className="text-xs text-slate-500">ملاحظة: {x.decisionNote}</p>}
            {d.canDecide && x.status === 'PENDING' && (
              <div className="flex gap-2 pt-1">
                <button type="button" disabled={busy === x.id} onClick={() => void decide(x.id, true)} className="rounded-lg bg-emerald-600 px-4 py-1.5 text-sm font-bold text-white disabled:opacity-50">✅ موافق</button>
                <button type="button" disabled={busy === x.id} onClick={() => void decide(x.id, false)} className="rounded-lg bg-slate-200 px-4 py-1.5 text-sm font-bold text-slate-700 disabled:opacity-50">🚫 رفض</button>
              </div>
            )}
          </div>
        ))}
      </section>
    </div>
  )
}
