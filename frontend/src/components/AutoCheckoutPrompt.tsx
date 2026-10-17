import { useEffect, useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

// ═══ «شنو صار بعد ما خلص دوامك؟» — قرار (ع) 10-06 ═══
// الي انسكّر انصرافه تلقائياً، أول ما يرجع يفتح النظام ماتركس يسأله:
// خلّصت وطلعت؟ چنت تشتغل لحد ساعة كذا؟ لو رجعت تشتغل هسه؟
// «چنت أشتغل» يتأكد منها ماتركس بالدليل (حجز، فاتورة، تقرير، مهمة)،
// والي بلا دليل يروح للمراقب يقرر.

export default function AutoCheckoutPrompt() {
  const { employee } = useSession()
  const [info, setInfo] = useState<{ id: string; at: string; label: string } | null>(null)
  const [step, setStep] = useState<'ask' | 'worked' | 'done'>('ask')
  const [until, setUntil] = useState('')
  const [note, setNote] = useState('')
  const [msg, setMsg] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const exempt = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
  useEffect(() => { if (!exempt) void api.getAttendanceGate().then((g) => setInfo(g.autoClosed ?? null)).catch(() => {}) }, [exempt])
  if (!info) return null

  const answer = async (kind: 'ACK' | 'WORKED' | 'BACK') => {
    setBusy(true); setErr('')
    try {
      let iso: string | undefined
      if (kind === 'WORKED') {
        const [h, m] = until.split(':').map(Number)
        const d = new Date()
        d.setHours(h, m, 0, 0)
        if (d.getTime() > Date.now()) d.setDate(d.getDate() - 1)
        iso = d.toISOString()
      }
      const r = await api.answerAutoCheckout(info.id, kind, iso, note)
      setMsg(r.message); setStep('done')
      window.dispatchEvent(new Event('matrix-refresh'))
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }

  return (
    <div dir="rtl" className="fixed inset-0 z-[85] flex items-end justify-center bg-black/40 p-3 sm:items-center">
      <div className="w-full max-w-md rounded-3xl bg-white p-5 shadow-2xl">
        <div className="text-center text-4xl">🤖</div>
        {step === 'ask' && (
          <>
            <p className="mt-2 text-center text-lg font-extrabold text-[#0f2040]">ما سجّلت انصراف، فسجّلتك تلقائي الساعة {info.label}</p>
            <p className="mt-1 text-center text-sm text-slate-600">شنو صار بعدها؟</p>
            <div className="mt-4 space-y-2">
              <button type="button" disabled={busy} onClick={() => void answer('ACK')} className="w-full rounded-xl border border-slate-300 py-2.5 font-bold">✅ خلّصت وطلعت — الوقت صحيح</button>
              <button type="button" disabled={busy} onClick={() => setStep('worked')} className="w-full rounded-xl border border-amber-300 bg-amber-50 py-2.5 font-bold text-amber-900">🛠️ چنت أشتغل بعدها</button>
              <button type="button" disabled={busy} onClick={() => void answer('BACK')} className="w-full rounded-xl bg-[#0f2040] py-2.5 font-bold text-white">🔁 رجعت أشتغل هسه — سجّل رجوعي</button>
            </div>
          </>
        )}
        {step === 'worked' && (
          <div className="mt-2 space-y-2">
            <p className="text-center font-extrabold text-[#0f2040]">لحد يمته چنت تشتغل؟</p>
            <input type="time" value={until} onChange={(e) => setUntil(e.target.value)} className="w-full rounded-xl border border-slate-300 p-2.5 text-center text-lg" />
            <textarea value={note} onChange={(e) => setNote(e.target.value)} rows={2} placeholder="شنو چنت تشتغل؟ (مثلاً: كمّلت حجز B12 بالموقع)" className="w-full rounded-xl border border-slate-300 p-2 text-sm" />
            <p className="text-[11px] text-slate-500">ماتركس يتأكد من النظام (حجز خلّصته، فاتورة، تقرير، مهمة). إذا لگى دليل يصحّح وقتك لحاله، وإذا ما لگى يروح للمراقب يقرر.</p>
            {err && <p className="text-sm text-red-600">{err}</p>}
            <div className="grid grid-cols-2 gap-2">
              <button type="button" onClick={() => setStep('ask')} className="rounded-xl border border-slate-300 py-2.5 font-bold">رجوع</button>
              <button type="button" disabled={busy || !until || note.trim().length < 3} onClick={() => void answer('WORKED')} className="rounded-xl bg-amber-600 py-2.5 font-bold text-white disabled:opacity-50">أرسل</button>
            </div>
          </div>
        )}
        {step === 'done' && (
          <>
            <p className="mt-2 text-center text-sm font-bold text-slate-800">{msg}</p>
            <button type="button" onClick={() => setInfo(null)} className="mt-4 w-full rounded-xl border border-slate-300 py-2.5 font-bold">تمام</button>
          </>
        )}
        {err && step === 'ask' && <p className="mt-2 text-center text-sm text-red-600">{err}</p>}
      </div>
    </div>
  )
}
