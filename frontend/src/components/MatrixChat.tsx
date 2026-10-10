import { useCallback, useEffect, useRef, useState } from 'react'
import { api, type MatrixChatMessage, type MatrixChatRow } from '../api'

// ═══ 💬 دردشة ماتركس — المالك والمدير (قرار (ع) 10-10) ═══
// «نسولف ويا ويرد مثل أي ذكاء اصطناعي». ماتركس يتذكر المحادثة، ويفتح أدواته
// (قراءة بس) لما يحتاج أرقام الشركة. الأسماء ما تطلع للنموذج.
const TOOL_AR: Record<string, string> = {
  money_month: 'فلوس الشهر', invoices_status: 'الفواتير', late_bookings: 'الحجوزات المتأخرة', team_performance: 'أداء الفرق',
  employee_profile: 'ملف موظف', attendance_today: 'دوام اليوم', project_delays: 'تأخير المشاريع', procurement_watch: 'شغل المخازن',
  staff_scores: 'التقييم',
}
const SUGGEST = ['شلون الشركة اليوم؟', 'سويلي فحص فوري للأموال هالشهر', 'منو الموظف الي يحتاج متابعة هالأسبوع وليش؟', 'شنو المشاريع المتأخرة؟']

export default function MatrixChat() {
  const [chats, setChats] = useState<MatrixChatRow[]>([])
  const [enabled, setEnabled] = useState(true)
  const [cur, setCur] = useState<string | null>(null)
  const [msgs, setMsgs] = useState<MatrixChatMessage[]>([])
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const end = useRef<HTMLDivElement>(null)

  const loadChats = useCallback(() => {
    api.getMatrixChats().then((r) => { setChats(r.chats); setEnabled(r.enabled) }).catch(() => {})
  }, [])
  useEffect(loadChats, [loadChats])
  useEffect(() => {
    if (!cur) return
    let alive = true
    api.getMatrixChat(cur).then((r) => { if (alive) setMsgs(r) }).catch(() => {})
    return () => { alive = false }
  }, [cur])
  useEffect(() => { end.current?.scrollIntoView({ behavior: 'smooth' }) }, [msgs, busy])

  const send = async (t: string) => {
    const body = t.trim()
    if (!body || busy) return
    setErr(''); setBusy(true)
    let id = cur
    try {
      if (!id) {
        const c = await api.createMatrixChat()
        id = c.id
        setCur(c.id)
      }
      setMsgs((m) => [...m, { id: 'tmp', role: 'USER', text: body, steps: null, createdAt: new Date().toISOString() }])
      setText('')
      const reply = await api.sendMatrixChat(id, body)
      setMsgs((m) => [...m.map((x) => (x.id === 'tmp' ? { ...x, id: 'u' + reply.id } : x)), reply])
      loadChats()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الإرسال')
      setMsgs((m) => m.filter((x) => x.id !== 'tmp'))
      setText(body)
    } finally { setBusy(false) }
  }

  return (
    <div dir="rtl" className="grid h-[calc(100vh-180px)] min-h-[520px] gap-3 lg:grid-cols-[260px_1fr]">
      <aside className="hidden flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white lg:flex">
        <button type="button" onClick={() => { setCur(null); setMsgs([]) }}
          className="m-3 rounded-xl bg-gradient-to-l from-violet-600 to-fuchsia-600 px-3 py-2 text-sm font-bold text-white">＋ محادثة جديدة</button>
        <div className="flex-1 space-y-1 overflow-y-auto px-2 pb-2">
          {chats.map((c) => (
            <div key={c.id} className={`group flex items-center gap-1 rounded-lg px-2 py-2 text-sm ${cur === c.id ? 'bg-violet-50 font-bold text-violet-900' : 'text-slate-700 hover:bg-slate-50'}`}>
              <button type="button" onClick={() => setCur(c.id)} className="min-w-0 flex-1 truncate text-right">{c.title}</button>
              <button type="button" title="حذف" onClick={() => { if (window.confirm('تحذف المحادثة؟')) api.deleteMatrixChat(c.id).then(() => { if (cur === c.id) { setCur(null); setMsgs([]) } loadChats() }) }}
                className="hidden text-xs text-slate-400 hover:text-red-600 group-hover:block">✕</button>
            </div>
          ))}
          {chats.length === 0 && <p className="p-3 text-xs text-slate-400">ماكو محادثات بعد.</p>}
        </div>
      </aside>

      <section className="flex min-h-0 flex-col overflow-hidden rounded-2xl border border-slate-200 bg-gradient-to-b from-white to-violet-50/40">
        <header className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <div className="flex items-center gap-2">
            <span className="grid h-9 w-9 place-items-center rounded-full bg-gradient-to-br from-violet-600 to-fuchsia-600 text-lg text-white">🤖</span>
            <div>
              <b className="text-[#0f2040]">ماتركس</b>
              <p className="text-[11px] text-slate-500">{enabled ? 'يسولف ويفتش بالنظام بنفسه — قراءة بس، والأسماء ما تطلع' : 'يحتاج مفتاح هايكو حتى يسولف'}</p>
            </div>
          </div>
          <button type="button" onClick={() => { setCur(null); setMsgs([]) }} className="rounded-lg px-2 py-1 text-xs font-bold text-violet-700 lg:hidden">＋ جديدة</button>
        </header>

        <div className="flex-1 space-y-3 overflow-y-auto px-4 py-4">
          {msgs.length === 0 && !busy && (
            <div className="mx-auto mt-6 max-w-lg text-center">
              <p className="text-2xl">👋</p>
              <p className="mt-2 font-bold text-[#0f2040]">هلا، شتحب نحچي اليوم؟</p>
              <div className="mt-4 flex flex-wrap justify-center gap-2">
                {SUGGEST.map((s) => (
                  <button key={s} type="button" disabled={!enabled} onClick={() => void send(s)}
                    className="rounded-full border border-violet-200 bg-white px-3 py-1.5 text-xs font-bold text-violet-800 hover:bg-violet-50 disabled:opacity-40">{s}</button>
                ))}
              </div>
            </div>
          )}
          {msgs.map((m) => (
            <div key={m.id} className={`flex ${m.role === 'USER' ? 'justify-start' : 'justify-end'}`}>
              <div className={`max-w-[85%] whitespace-pre-wrap rounded-2xl px-4 py-2.5 text-sm leading-relaxed shadow-sm ${m.role === 'USER' ? 'rounded-br-md bg-[#0f2040] text-white' : 'rounded-bl-md border border-violet-100 bg-white text-slate-800'}`}>
                {m.text}
                {m.role === 'ASSISTANT' && m.steps && (
                  <p className="mt-1.5 border-t border-slate-100 pt-1 text-[10px] text-slate-400">🔎 فتّش: {[...new Set(m.steps.split(','))].map((s) => TOOL_AR[s] ?? s).join('، ')}</p>
                )}
              </div>
            </div>
          ))}
          {busy && (
            <div className="flex justify-end">
              <div className="rounded-2xl rounded-bl-md border border-violet-100 bg-white px-4 py-3 text-sm text-violet-700 shadow-sm">
                <span className="inline-flex gap-1"><span className="animate-bounce">●</span><span className="animate-bounce [animation-delay:120ms]">●</span><span className="animate-bounce [animation-delay:240ms]">●</span></span>
                <span className="mr-2 text-xs">ماتركس يفكّر ويفتش…</span>
              </div>
            </div>
          )}
          <div ref={end} />
        </div>

        {err && <p className="mx-4 mb-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{err}</p>}
        <form onSubmit={(e) => { e.preventDefault(); void send(text) }} className="flex items-end gap-2 border-t border-slate-100 bg-white p-3">
          <textarea value={text} onChange={(e) => setText(e.target.value)} rows={1} disabled={!enabled}
            onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); void send(text) } }}
            placeholder={enabled ? 'اكتب لماتركس… (Enter للإرسال)' : 'ماتركس يحتاج مفتاح هايكو'}
            className="max-h-40 min-h-[44px] flex-1 resize-y rounded-xl border border-slate-300 px-3 py-2.5 text-sm outline-none focus:border-violet-500 disabled:bg-slate-50" />
          <button type="submit" disabled={busy || !text.trim() || !enabled}
            className="h-11 rounded-xl bg-gradient-to-l from-violet-600 to-fuchsia-600 px-5 text-sm font-bold text-white disabled:opacity-40">إرسال</button>
        </form>
      </section>
    </div>
  )
}
