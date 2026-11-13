import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixOrder, type MatrixOrderAction } from '../api'
import MatrixMissions from './MatrixMissions'

// ═══ 🎯 أوامر ماتركس (دازاي 11-13) ═══
// «مهام ماتركس اكدر اوجه لماتركس مهمه — اكله روح سويلي عرض سعر، اكله روح اخصم هذا الموظف نقطه».
// تكتب الأمر ← ماتركس يفهمه ويطلّع بطاقة بالضبط شنو راح يسوي ← ما يتنفّذ شي إلا تضغط «نفّذ».
// تكدر تشيل أي إجراء من البطاقة قبل التنفيذ.

const EXAMPLES = [
  'اخصم من فني الحجز B120 نقطتين لأن تأخر ساعة',
  'سويلي عرض سعر لزبون جديد: ٤ كاميرات وجهاز تسجيل',
  'انطي أكثر واحد متأخر هالأسبوع مهمة يكتب تبرير خلال يومين',
  'راقب موظف المخزن أسبوع وكلي شنو يصير',
]

const KIND: Record<MatrixOrderAction['kind'], { icon: string; label: string; tone: string }> = {
  points: { icon: '➖', label: 'خصم نقاط', tone: 'bg-red-50 text-red-800 ring-red-200' },
  task: { icon: '📋', label: 'مهمة', tone: 'bg-sky-50 text-sky-800 ring-sky-200' },
  notify: { icon: '🔔', label: 'تنبيه', tone: 'bg-amber-50 text-amber-800 ring-amber-200' },
  watch: { icon: '👁️', label: 'مراقبة', tone: 'bg-violet-50 text-violet-800 ring-violet-200' },
  quotation: { icon: '🧾', label: 'عرض سعر', tone: 'bg-emerald-50 text-emerald-800 ring-emerald-200' },
}

const STATUS: Record<MatrixOrder['status'], string> = {
  PROPOSED: '⏳ ينتظر موافقتك', RUNNING: '⚙️ يتنفّذ', DONE: '✅ انفّذ', CANCELLED: '✖️ ملغي', FAILED: '⚠️ ما انفّذ',
}

function describe(a: MatrixOrderAction): string {
  switch (a.kind) {
    case 'points': return `يخصم ${a.points} نقطة من ${a.employeeName} — السبب: ${a.reason}`
    case 'task': return `مهمة لـ${a.employeeName}: «${a.title}»${a.message ? ` — ${a.message}` : ''} · موعدها بعد ${a.days} يوم`
    case 'notify': return `تنبيه لـ${a.employeeName}: «${a.message}»`
    case 'watch': return `ماتركس يراقب ${a.employeeName} ${a.days} يوم${a.reason ? ` — ${a.reason}` : ''}`
    case 'quotation': {
      const items = (a.items ?? []).map((i) => `${i.quantity}× ${i.productName}${i.unitPrice ? ` (${i.unitPrice.toLocaleString('en-US')})` : ''}`).join('، ')
      return `عرض سعر (مسودة) لـ${a.customerName}${a.projectName ? ` · ${a.projectName}` : ''}${a.template ? ` · من نموذج «${a.template}»` : ''}${items ? ` — ${items}` : ' — بلا بنود، تكمّلها من الشاشة'}`
    }
  }
}

function OrderCard({ o, onChange }: { o: MatrixOrder; onChange: (o: MatrixOrder) => void }) {
  const [skip, setSkip] = useState<number[]>([])
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const pending = o.status === 'PROPOSED'
  const left = o.actions.length - skip.length
  const run = async () => {
    setBusy(true); setErr('')
    try {
      window.dispatchEvent(new Event('matrix-mangekyo'))
      onChange(await api.runMatrixOrder(o.id, skip))
    } catch (e) { setErr(e instanceof Error ? e.message : 'ما انفّذ') } finally { setBusy(false) }
  }
  const cancel = async () => {
    setBusy(true)
    try { await api.cancelMatrixOrder(o.id); onChange({ ...o, status: 'CANCELLED' }) } catch { setErr('ما انلغى') } finally { setBusy(false) }
  }
  return (
    <div className={`rounded-2xl border bg-white p-3 shadow-sm ${pending ? 'border-violet-300 ring-2 ring-violet-100' : 'border-slate-200'}`}>
      <div className="flex flex-wrap items-start justify-between gap-2">
        <p className="text-sm font-bold text-slate-800">🗣️ «{o.text}»</p>
        <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-bold text-slate-600">{STATUS[o.status]}</span>
      </div>
      {o.reply && <p className="mt-1 text-xs leading-5 text-violet-800">🤖 {o.reply}</p>}
      <ul className="mt-2 space-y-1.5">
        {o.actions.map((a, i) => {
          const off = skip.includes(i) || a.skip
          return (
            <li key={i} className={`flex items-start gap-2 rounded-xl px-2.5 py-2 text-xs ring-1 ${KIND[a.kind].tone} ${off ? 'opacity-40 line-through' : ''}`}>
              <span className="font-extrabold">{KIND[a.kind].icon} {KIND[a.kind].label}</span>
              <span className="flex-1 leading-5">
                {describe(a)}
                {a.done && <span className="mt-0.5 block font-bold text-emerald-700">✅ {a.result}{a.link && <> · <Link to={a.link} className="underline">افتح</Link></>}</span>}
                {a.error && <span className="mt-0.5 block font-bold text-red-700">⚠️ {a.error}</span>}
              </span>
              {pending && (
                <button type="button" onClick={() => setSkip((s) => s.includes(i) ? s.filter((x) => x !== i) : [...s, i])}
                  className="shrink-0 rounded-lg bg-white/80 px-2 py-0.5 text-[11px] font-bold text-slate-600 no-underline ring-1 ring-slate-200">
                  {skip.includes(i) ? 'رجّعه' : 'شيله'}
                </button>
              )}
            </li>
          )
        })}
      </ul>
      {pending && (
        <div className="mt-3 flex flex-wrap items-center gap-2">
          <button type="button" disabled={busy || left === 0} onClick={() => void run()}
            className="rounded-xl bg-violet-700 px-4 py-2 text-sm font-extrabold text-white disabled:opacity-50">
            {busy ? 'ينفّذ…' : `✅ نفّذ (${left})`}
          </button>
          <button type="button" disabled={busy} onClick={() => void cancel()} className="rounded-xl bg-white px-4 py-2 text-sm font-bold text-slate-600 ring-1 ring-slate-200">إلغاء</button>
          <span className="text-[11px] text-slate-500">ماكو شي يتنفّذ قبل ما تضغط «نفّذ».</span>
        </div>
      )}
      {err && <p className="mt-2 text-xs font-bold text-red-700">{err}</p>}
    </div>
  )
}

export default function MatrixOrders() {
  const [rows, setRows] = useState<MatrixOrder[]>([])
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [showWatch, setShowWatch] = useState(false)
  const load = useCallback(() => { api.getMatrixOrders().then(setRows).catch(() => setRows([])) }, [])
  useEffect(load, [load])

  const send = async () => {
    if (!text.trim()) return
    setBusy(true); setErr('')
    try {
      const o = await api.proposeMatrixOrder(text.trim())
      setRows((r) => [o, ...r]); setText('')
    } catch (e) { setErr(e instanceof Error ? e.message : 'ماتركس ما فهم') } finally { setBusy(false) }
  }
  const replace = (o: MatrixOrder) => setRows((r) => r.map((x) => (x.id === o.id ? o : x)))
  const pending = rows.filter((o) => o.status === 'PROPOSED')
  const history = rows.filter((o) => o.status !== 'PROPOSED')

  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-violet-200 bg-gradient-to-b from-violet-50 to-white p-3 sm:p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <p className="text-lg font-extrabold text-[#0f2040]">🎯 أوامر ماتركس</p>
          <p className="text-xs text-slate-500">كله شنو يسوي بكلامك. يفهمه ويوريك بالضبط شنو راح يصير — وما ينفّذ إلا من تضغط «نفّذ». أسماء الموظفين والزبائن ما تطلع للذكاء الاصطناعي.</p>
        </div>
        <button type="button" onClick={() => setShowWatch((v) => !v)} className="rounded-lg bg-white px-3 py-1.5 text-xs font-bold text-violet-800 ring-1 ring-violet-200">
          👁️ المراقبة {showWatch ? '▲' : '▼'}
        </button>
      </div>
      <div className="space-y-2">
        <textarea value={text} onChange={(e) => setText(e.target.value)} rows={2} maxLength={1500}
          onKeyDown={(e) => { if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) void send() }}
          placeholder="مثلاً: اخصم من فلان نقطة لأن ما سلّم الفاتورة · سويلي عرض سعر لشركة … · انطي فلان مهمة …"
          className="w-full rounded-xl border border-violet-200 bg-white p-3 text-sm focus:border-violet-500 focus:outline-none" />
        <div className="flex flex-wrap items-center gap-2">
          <button type="button" disabled={busy || !text.trim()} onClick={() => void send()}
            className="rounded-xl bg-[#0f2040] px-4 py-2 text-sm font-extrabold text-white disabled:opacity-50">{busy ? 'ماتركس يفكر…' : '🎯 وجّه الأمر'}</button>
          {EXAMPLES.map((x) => (
            <button key={x} type="button" onClick={() => setText(x)} className="rounded-full bg-white px-2.5 py-1 text-[11px] text-slate-600 ring-1 ring-slate-200 hover:bg-violet-50">{x}</button>
          ))}
        </div>
        {err && <p className="text-xs font-bold text-red-700">{err}</p>}
      </div>
      {pending.map((o) => <OrderCard key={o.id} o={o} onChange={replace} />)}
      {history.length > 0 && (
        <details className="space-y-2">
          <summary className="cursor-pointer text-xs font-bold text-slate-500">📜 الأوامر السابقة ({history.length})</summary>
          <div className="mt-2 space-y-2">{history.map((o) => <OrderCard key={o.id} o={o} onChange={replace} />)}</div>
        </details>
      )}
      {showWatch && <MatrixMissions />}
    </div>
  )
}
