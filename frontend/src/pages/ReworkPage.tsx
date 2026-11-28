import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { notify } from '../ui/notify'
import { fmtDateTime } from '../utils/date'

// ═══ 🔧 الخطأ الفني والتصحيح (ل) 11-24 دازاي ═══
// (دازاي 11-28) «ابو الجوده من يكتشف الخطا من خلال الاتصال او الشكوى هوه يسجله… والنظام مباشرة يطلع المعلومات»:
// يبحث الحجز برقم الزبون أو اسمه أو كوده، يختاره ومعلوماته تطلع لحالها (الزبون، الخدمة، الكادر، الشكاوى).
// حجز التصحيح اختياري وقت التسجيل — ينربط بعدين، والمراقب يوافق لمن ينربط.
// بالموافقة: ساعات حجز التصحيح × سعر ساعة الي صلّح ← ينضاف للي صلّح وينخصم من الكادر الأول.

type Item = {
  id: string; fixCode: string | null; origCode: string; note: string; status: 'PENDING' | 'APPROVED' | 'REJECTED'
  hours: number | null; amount: number | null; by: string; createdAt: string; decidedBy: string | null; decisionNote: string | null; fixCrew: string; origCrew: string
}
type Resp = { items: Item[]; canDecide: boolean; canReport?: boolean }
type Bk = { code: string; customer: string; phone: string; service: string; completedAt: string | null; status: string; crew: string; complaints: number }
const ST = { PENDING: ['⏳ ينتظر المراقب', 'bg-amber-100 text-amber-800'], APPROVED: ['✅ انوافق', 'bg-emerald-100 text-emerald-800'], REJECTED: ['🚫 انرفض', 'bg-slate-200 text-slate-600'] } as const
const inp = 'w-full rounded-lg border border-slate-300 p-2 text-sm'

/** بحث حجز: تكتب رقم الزبون أو اسمه أو كود الحجز، وتختار من القائمة. */
function BookingPicker({ label, value, onPick }: { label: string; value: Bk | null; onPick: (b: Bk | null) => void }) {
  const [q, setQ] = useState('')
  const [res, setRes] = useState<Bk[]>([])
  useEffect(() => {
    if (q.trim().length < 2) return
    let alive = true
    const t = setTimeout(() => {
      apiRequest<{ items: Bk[] }>(`/rework/bookings?q=${encodeURIComponent(q.trim())}`)
        .then((r) => { if (alive) setRes(r.items) }).catch(() => { if (alive) setRes([]) })
    }, 300)
    return () => { alive = false; clearTimeout(t) }
  }, [q])
  if (value) {
    return (
      <div className="rounded-xl border border-emerald-300 bg-emerald-50 p-3 text-sm">
        <div className="flex items-center justify-between"><b className="text-[#0f2040]">{label}: <bdi dir="ltr">{value.code}</bdi></b>
          <button type="button" className="text-xs text-slate-500 underline" onClick={() => onPick(null)}>غيّر</button></div>
        <p className="mt-1 text-slate-700">👤 {value.customer} · <bdi dir="ltr">{value.phone}</bdi> · 🔧 {value.service || '—'}</p>
        <p className="text-xs text-slate-500">الكادر: {value.crew || '—'}{value.completedAt ? ` · انجز ${fmtDateTime(value.completedAt)}` : ''}{value.complaints ? ` · ⚠️ ${value.complaints} شكوى` : ''}</p>
      </div>
    )
  }
  return (
    <div className="relative">
      <label className="mb-1 block text-xs font-bold text-slate-600">{label}</label>
      <input value={q} onChange={(e) => { setQ(e.target.value); if (e.target.value.trim().length < 2) setRes([]) }} placeholder="رقم الزبون أو اسمه أو كود الحجز…" className={inp} />
      {res.length > 0 && q.trim().length >= 2 && (
        <div className="absolute z-20 mt-1 max-h-72 w-full overflow-y-auto rounded-xl border border-slate-200 bg-white shadow-lg">
          {res.map((b) => (
            <button key={b.code} type="button" onClick={() => { onPick(b); setQ(''); setRes([]) }} className="block w-full border-b border-slate-100 p-2 text-right text-sm hover:bg-slate-50">
              <b dir="ltr">{b.code}</b> — {b.customer} <span className="text-xs text-slate-500">({b.service || '—'}{b.complaints ? ` · ⚠️ ${b.complaints} شكوى` : ''})</span>
              <span className="block text-xs text-slate-400">الكادر: {b.crew || '—'}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

function LinkFix({ id, onDone }: { id: string; onDone: (r: Resp) => void }) {
  const [b, setB] = useState<Bk | null>(null)
  const [busy, setBusy] = useState(false)
  const save = async () => {
    if (!b) return
    setBusy(true)
    try { onDone(await apiRequest<Resp>(`/rework/${id}/fix`, { method: 'POST', body: JSON.stringify({ fixCode: b.code }) })); notify('انربط حجز التصحيح ✅', 'success') }
    catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  return (
    <div className="mt-2 space-y-2 rounded-xl bg-slate-50 p-3">
      <BookingPicker label="حجز التصحيح (الي رجعوا بيه للزبون)" value={b} onPick={setB} />
      {b && <button type="button" disabled={busy} onClick={() => void save()} className="rounded-lg bg-[#0f2040] px-4 py-1.5 text-sm font-bold text-white disabled:opacity-50">اربط</button>}
    </div>
  )
}

export default function ReworkPage() {
  const [d, setD] = useState<Resp | null>(null)
  const [err, setErr] = useState('')
  const [orig, setOrig] = useState<Bk | null>(null)
  const [fix, setFix] = useState<Bk | null>(null)
  const [note, setNote] = useState('')
  const [busy, setBusy] = useState('')
  useEffect(() => { apiRequest<Resp>('/rework').then(setD).catch((e: Error) => setErr(e.message)) }, [])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!d) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  const reporter = !!d.canReport
  const submit = async () => {
    if (!orig) return
    setBusy('new')
    try {
      setD(await apiRequest<Resp>('/rework', { method: 'POST', body: JSON.stringify({ origCode: orig.code, fixCode: fix?.code ?? '', note }) }))
      setOrig(null); setFix(null); setNote('')
      notify('انسجّل — ينتظر موافقة المراقب ✅', 'success')
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy('') }
  }
  const decide = async (id: string, approve: boolean) => {
    setBusy(id)
    try { setD(await apiRequest<Resp>(`/rework/${id}/decide`, { method: 'POST', body: JSON.stringify({ approve }) })); notify(approve ? 'انوافق وانتقل المبلغ ✅' : 'انرفض', 'success') } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy('') }
  }
  return (
    <div className="mx-auto max-w-5xl space-y-4" data-testid="rework">
      <header>
        <h1 className="text-2xl font-extrabold text-[#0f2040]">🔧 الخطأ الفني والتصحيح</h1>
        <p className="text-sm text-slate-500">
          {reporter
            ? 'لمن تكتشف خطأ فني بالاتصال أو الشكوى: ابحث الحجز برقم الزبون أو اسمه، ومعلوماته تطلع لحالها. حجز التصحيح تربطه هسه أو بعدين لمن يرجعون للزبون.'
            : 'إذا رجعتوا لزبون تصلّحون غلط كادر سابق: اختاروا الحجزين واكتبوا الخطأ.'}
          {' '}المراقب يوافق، ووقت التصحيح × سعر ساعة الي صلّح ينضاف إله وينخصم من الكادر الأول.
        </p>
      </header>
      <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
        <div className="grid gap-3 md:grid-cols-2">
          <BookingPicker label="الحجز الي بيه الخطأ" value={orig} onPick={setOrig} />
          <BookingPicker label={reporter ? 'حجز التصحيح (اختياري — تكدر تربطه بعدين)' : 'حجز التصحيح'} value={fix} onPick={setFix} />
        </div>
        <textarea value={note} onChange={(e) => setNote(e.target.value)} rows={2} placeholder="شنو چان الخطأ الفني (وشلون انكشف: اتصال، شكوى…)" className={inp} />
        <button type="button" disabled={busy === 'new' || !orig || (!reporter && !fix) || note.trim().length < 5} onClick={() => void submit()}
          className="rounded-lg bg-[#0f2040] px-6 py-2 text-sm font-bold text-white disabled:opacity-40">سجّل</button>
      </section>
      <section className="space-y-2">
        {d.items.length === 0 && <p className="rounded-2xl bg-white p-6 text-center text-slate-500">ماكو أخطاء مسجّلة بعد.</p>}
        {d.items.map((x) => (
          <div key={x.id} className="space-y-1 rounded-2xl bg-white p-4 shadow-sm">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <b className="text-[#0f2040]">
                {x.fixCode ? <><bdi dir="ltr">{x.fixCode}</bdi> صلّح <bdi dir="ltr">{x.origCode}</bdi></> : <>خطأ بالحجز <bdi dir="ltr">{x.origCode}</bdi> — لسه ما انربط التصحيح</>}
              </b>
              <span className={`rounded-full px-3 py-0.5 text-xs font-bold ${ST[x.status][1]}`}>{ST[x.status][0]}</span>
            </div>
            <p className="text-sm text-slate-700">{x.note}</p>
            <p className="text-xs text-slate-500">الي صلّحوا: {x.fixCrew || '—'} · الكادر الأول: {x.origCrew || '—'} · سجّله {x.by} · {fmtDateTime(x.createdAt)}</p>
            {x.status === 'APPROVED' && <p className="text-sm">⏱️ {x.hours} ساعة · 💰 <b dir="ltr">{Math.round(x.amount ?? 0).toLocaleString('en-US')}</b> د.ع انتقلت{x.decidedBy ? ` — وافق ${x.decidedBy}` : ''}</p>}
            {x.decisionNote && <p className="text-xs text-slate-500">ملاحظة: {x.decisionNote}</p>}
            {x.status === 'PENDING' && !x.fixCode && <LinkFix id={x.id} onDone={setD} />}
            {d.canDecide && x.status === 'PENDING' && (
              <div className="flex gap-2 pt-1">
                <button type="button" disabled={busy === x.id || !x.fixCode} title={x.fixCode ? '' : 'اربط حجز التصحيح أول'} onClick={() => void decide(x.id, true)} className="rounded-lg bg-emerald-600 px-4 py-1.5 text-sm font-bold text-white disabled:opacity-50">✅ موافق</button>
                <button type="button" disabled={busy === x.id} onClick={() => void decide(x.id, false)} className="rounded-lg bg-slate-200 px-4 py-1.5 text-sm font-bold text-slate-700 disabled:opacity-50">🚫 رفض</button>
              </div>
            )}
          </div>
        ))}
      </section>
    </div>
  )
}
