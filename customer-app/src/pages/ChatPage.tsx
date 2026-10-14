import { useEffect, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { api, type Account, type ChatMsg } from '../api'
import MatrixEye from '../components/MatrixEye'

const SUGGEST = ['شنو خدماتكم؟', 'أريد أحجز', 'شنو العروض الحالية؟', 'شوكت يوصلني الفريق؟']

// معرّف مؤقت للرسالة — randomUUID ما يشتغل على http (المعاينة بالـIP)
let seq = 0
const tmpId = () => `tmp-${++seq}`

const time = (s: string) => new Date(s).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })

// ═══ ماتركس — المحادثة ═══
// ماتركس يجاوب، يستفسر، ويسجّل الحجز بنفسه — والحجز يطلع ببطاقة وياها رقمه.
export default function ChatPage({ me }: { me: Account }) {
  const nav = useNavigate()
  const loc = useLocation()
  const [msgs, setMsgs] = useState<ChatMsg[] | null>(null)
  const [text, setText] = useState(() => (loc.state as { prefill?: string } | null)?.prefill ?? '')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const end = useRef<HTMLDivElement>(null)

  useEffect(() => { api.chat().then((r) => setMsgs(r.messages)).catch((e) => { setErr(e.message); setMsgs([]) }) }, [])
  useEffect(() => { end.current?.scrollIntoView({ behavior: 'smooth' }) }, [msgs, busy])

  const send = async (t = text) => {
    const body = t.trim()
    if (!body || busy) return
    setText(''); setErr(''); setBusy(true)
    const mine: ChatMsg = { id: tmpId(), role: 'user', content: body, createdAt: new Date().toISOString() }
    setMsgs((m) => [...(m ?? []), mine])
    try {
      const reply = await api.send(body)
      setMsgs((m) => [...(m ?? []), reply])
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الإرسال')
      setMsgs((m) => (m ?? []).filter((x) => x.id !== mine.id))
      setText(body)
    } finally { setBusy(false) }
  }

  return (
    <div className="flex h-[100dvh] flex-col pb-[72px]">
      <header className="flex items-center gap-3 border-b border-slate-200 bg-white px-4 py-3">
        <MatrixEye size={42} busy={busy} />
        <div>
          <p className="font-black text-ink">ماتركس</p>
          <p className="text-[11px] text-emerald-600">{busy ? 'يكتب…' : 'مساعد الأماني الذكي • متصل'}</p>
        </div>
      </header>

      <div className="flex-1 space-y-3 overflow-y-auto px-4 py-4">
        {msgs === null && <div className="h-20 animate-pulse rounded-2xl bg-slate-200" />}
        {msgs?.length === 0 && (
          <div className="rise rounded-3xl bg-white p-5 text-center shadow-sm">
            <div className="flex justify-center"><MatrixEye size={64} /></div>
            <p className="mt-3 text-lg font-black text-ink">هلا {me.name.split(' ')[0]}! أنا ماتركس 👁️</p>
            <p className="mt-1 text-sm text-slate-500">أجاوبك عن خدماتنا وعروضنا، وأسجّلك حجز بدقيقة. شتحتاج اليوم؟</p>
          </div>
        )}
        {msgs?.map((m) => (
          <div key={m.id} className={`rise flex ${m.role === 'user' ? 'justify-start' : 'justify-end'}`}>
            <div className={`max-w-[85%] rounded-3xl px-4 py-2.5 text-[15px] leading-relaxed whitespace-pre-wrap shadow-sm ${
              m.role === 'user' ? 'rounded-tr-md bg-navy text-white' : 'rounded-tl-md bg-white text-ink'}`}>
              {m.content}
              {m.bookingCode && (
                <button onClick={() => nav('/bookings')} className="mt-2 flex w-full items-center gap-2 rounded-2xl bg-emerald-50 p-3 text-right">
                  <span className="text-2xl">✅</span>
                  <span className="flex-1"><b className="block text-sm text-emerald-800">تم تسجيل حجزك</b><span className="text-xs text-emerald-700">رقم الحجز {m.bookingCode}</span></span>
                  <span className="text-xs font-bold text-emerald-700">حجوزاتي ←</span>
                </button>
              )}
              <span className={`mt-1 block text-[10px] ${m.role === 'user' ? 'text-blue-200' : 'text-slate-400'}`}>{time(m.createdAt)}</span>
            </div>
          </div>
        ))}
        {busy && (
          <div className="flex justify-end">
            <div className="flex gap-1 rounded-3xl rounded-tl-md bg-white px-4 py-3 shadow-sm">
              {[0, 1, 2].map((i) => <span key={i} className="dot h-2 w-2 rounded-full bg-slate-400" style={{ animationDelay: `${i * 0.2}s` }} />)}
            </div>
          </div>
        )}
        <div ref={end} />
      </div>

      {err && <p className="mx-4 mb-2 rounded-xl bg-rose-50 px-3 py-2 text-sm text-rose-700">{err}</p>}
      {msgs && msgs.length < 2 && !busy && (
        <div className="no-scrollbar flex gap-2 overflow-x-auto px-4 pb-2">
          {SUGGEST.map((s) => <button key={s} onClick={() => send(s)} className="shrink-0 rounded-full border border-navy/20 bg-white px-3.5 py-1.5 text-sm font-bold text-navy">{s}</button>)}
        </div>
      )}
      <form onSubmit={(e) => { e.preventDefault(); void send() }} className="flex items-end gap-2 border-t border-slate-200 bg-white px-3 py-2.5">
        <textarea value={text} onChange={(e) => setText(e.target.value)} rows={1} maxLength={1500} placeholder="اكتب رسالتك…"
          onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); void send() } }}
          className="max-h-32 min-h-11 flex-1 resize-none rounded-3xl bg-slate-100 px-4 py-2.5 text-base outline-none focus:ring-2 focus:ring-navy/20" />
        <button disabled={busy || !text.trim()} aria-label="إرسال" className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-sun text-white shadow-md disabled:opacity-40">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" className="-scale-x-100"><path d="M3 20l18-8L3 4v6l12 2-12 2z" /></svg>
        </button>
      </form>
    </div>
  )
}
