import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { useSession } from '../session'
import { api, type SecretaryChat, type SecretaryProposal as Proposal, type SecretaryReply, type SecretaryToday } from '../api'

type Chat = SecretaryChat

const time = (iso: string | null) => iso ? new Date(iso).toLocaleTimeString('ar-IQ', { hour: 'numeric', minute: '2-digit' }) : ''

export default function SecretaryPage() {
  const { employee } = useSession()
  const allowed = employee?.actualRole === 'OWNER' || employee?.role === 'ADMIN'
  const [input, setInput] = useState('')
  const [chats, setChats] = useState<Chat[]>([])
  const [today, setToday] = useState<SecretaryToday | null>(null)
  const [proposal, setProposal] = useState<Proposal | null>(null)
  const [recording, setRecording] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [deferred, setDeferred] = useState<string[]>([])
  const [reload, setReload] = useState(0)
  const holdTimer = useRef<number | null>(null)
  const recorder = useRef<MediaRecorder | null>(null)

  // الجلب بمكان واحد ويا alive (دروس AGENTS.md). ما ننادي المسارات إلا للمالك والمدير (403 = مخالفة).
  useEffect(() => {
    if (!allowed) return
    let alive = true
    api.secretaryHistory(50)
      .then((h) => { if (alive) setChats(h) })
      .catch((e: Error) => { if (alive) setError(e.message || 'تعذر جلب المحادثات') })
    return () => { alive = false }
  }, [allowed])
  // ملخص اليوم ينعاد بعد كل إرسال (رفع العدّاد) — المحادثة تبقى مثل ما هي
  useEffect(() => {
    if (!allowed) return
    let alive = true
    api.secretaryToday()
      .then((t) => { if (alive) setToday(t) })
      .catch((e: Error) => { if (alive) setError(e.message || 'تعذر جلب ملخص اليوم') })
    return () => { alive = false }
  }, [allowed, reload])

  useEffect(() => {
    let manifest = document.querySelector<HTMLLinkElement>('link[rel="manifest"]')
    const created = !manifest
    if (!manifest) {
      manifest = document.createElement('link')
      manifest.rel = 'manifest'
      document.head.append(manifest)
    }
    const previous = manifest?.href
    if (manifest) manifest.href = `${import.meta.env.BASE_URL}secretary.webmanifest`
    document.title = 'السكرتير'
    return () => { if (created) manifest?.remove(); else if (manifest && previous) manifest.href = previous; document.title = 'نظام شركة الأماني' }
  }, [])

  if (!allowed) return <main dir="rtl" className="grid min-h-dvh place-items-center bg-slate-100 p-5"><div className="max-w-sm rounded-3xl bg-white p-8 text-center shadow-xl"><p className="text-4xl">🔒</p><h1 className="mt-3 text-xl font-black text-slate-800">السكرتير خاص بالإدارة</h1><p className="mt-2 text-sm text-slate-500">هاي المساحة للمالك ومدير النظام فقط.</p></div></main>

  const now = () => new Date().toISOString()
  const pushReply = (userText: string, r: SecretaryReply) => {
    setChats((v) => [...v,
      { id: `u-${r.messageId}`, at: now(), role: 'user', text: userText, transcript: r.transcript, cards: [], action: null },
      { id: r.messageId, at: now(), role: 'assistant', text: r.reply, transcript: null, cards: r.cards, action: r.proposal ? { id: r.proposal.actionId, status: 'PROPOSED', result: null } : null }])
    setProposal(r.proposal)
  }
  const send = (raw = input) => {
    const text = raw.trim(); if (!text || busy) return
    setBusy(true); setError(''); setInput('')
    api.secretaryMessage(text).then((r) => pushReply(text, r)).catch((e: Error) => { setError(e.message); setInput(text) }).finally(() => setBusy(false))
  }
  const startVoice = () => {
    if (busy) return
    holdTimer.current = window.setTimeout(async () => {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
        const r = new MediaRecorder(stream); const chunks: Blob[] = []
        r.ondataavailable = (e) => { if (e.data.size) chunks.push(e.data) }
        r.onstop = () => {
          stream.getTracks().forEach((t) => t.stop())
          const blob = new Blob(chunks, { type: r.mimeType || 'audio/webm' })
          if (!blob.size) return
          setBusy(true); setError('')
          api.secretaryVoice(blob).then((res) => pushReply(res.transcript || '🎙️', res)).catch((e: Error) => setError(e.message)).finally(() => setBusy(false))
        }
        recorder.current = r; r.start(); setRecording(true)
      } catch { setError('ما گدرت أفتح المايك — اسمح بيه من المتصفح أو اكتب') }
    }, 180)
  }
  const stopVoice = () => {
    if (holdTimer.current) window.clearTimeout(holdTimer.current)
    if (recorder.current && recorder.current.state === 'recording') recorder.current.stop()
    recorder.current = null; setRecording(false)
  }
  const sendProposal = () => {
    if (!proposal || busy) return
    if (!proposal.fields.employeeId) { setError('اختار الموظف أول'); return }
    setBusy(true); setError('')
    api.secretaryConfirm(proposal.actionId, proposal.fields)
      .then((r) => { setChats((v) => [...v, { id: `sent-${proposal.actionId}`, at: now(), role: 'assistant', text: 'انرسلت المهمة.', transcript: null, cards: [r.card], action: { id: proposal.actionId, status: 'SENT', result: null } }]); setProposal(null); setReload((n) => n + 1) })
      .catch((e: Error) => setError(e.message)).finally(() => setBusy(false))
  }
  const cancelProposal = () => {
    if (!proposal) return
    const id = proposal.actionId; setProposal(null)
    api.secretaryCancel(id).catch(() => {})
  }
  const pending = (today?.pending ?? []).filter((p) => !deferred.includes(p.id))

  return <main dir="rtl" className="min-h-dvh bg-[radial-gradient(circle_at_top,#dbeafe,transparent_42%),#f8fafc] text-slate-800">
    <div className="mx-auto min-h-dvh max-w-2xl px-4 pb-36 pt-5 sm:px-6">
      <header className="flex items-start justify-between"><div><p className="text-sm font-bold text-brand-600">السكرتير</p><h1 className="mt-1 text-2xl font-black text-slate-900">{today?.greeting ?? 'السكرتير'}</h1><p className="mt-1 text-sm text-slate-500">كل المهم قدامك، بلا ما تدوّر بالنظام.</p></div><div className="rounded-2xl bg-amber-100 px-3 py-2 text-center text-xs font-bold text-amber-900"><b className="block text-lg leading-none">{pending.length}</b>ينتظر قرارك</div></header>
      <section className="mt-5 rounded-3xl border border-white/70 bg-white/85 p-4 shadow-[0_18px_50px_-30px_rgba(15,32,64,.5)] backdrop-blur"><div className="flex items-center justify-between"><h2 className="font-extrabold">ملخص اليوم</h2><span className="text-xs text-slate-400">محدّث هسه</span></div><ul className="mt-3 space-y-2">{(today?.summary.lines ?? []).map((line) => <li key={line} className="flex gap-2 text-sm"><span className="text-brand-500">●</span>{line}</li>)}</ul></section>
      <section className="mt-5"><div className="mb-3 flex items-center justify-between"><h2 className="font-extrabold">ينتظر قرارك</h2><span className="text-xs text-slate-400">{deferred.length ? `${deferred.length} مؤجل` : 'افتحه من مكانه'}</span></div><div className="space-y-3">{pending.map((item) => <article key={item.id} className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm"><p className="font-bold">{item.question}</p><p className="mt-2 text-sm text-slate-500">{item.why}</p><p className="mt-1 text-xs text-slate-400">الأثر: {item.impact}</p><div className="mt-3 flex gap-2"><Link to={item.link} className="rounded-xl bg-brand-600 px-3 py-2 text-xs font-bold text-white">افتح التفاصيل</Link><button onClick={() => setDeferred((v) => [...v, item.id])} className="rounded-xl border border-slate-200 px-3 py-2 text-xs font-bold text-slate-600">أجّل</button></div></article>)}</div></section>
      {error && <p role="alert" className="mt-4 rounded-2xl bg-red-50 p-3 text-sm font-bold text-red-700">{error}</p>}
      <section className="mt-6"><h2 className="mb-3 font-extrabold">آخر المحادثات</h2>{chats.length === 0 && <p className="text-sm text-slate-500">آني سكرتيرك. گلي شتريد أعرف أو أتابع؟</p>}<div className="space-y-3">{chats.map((chat) => <div key={chat.id} className={chat.role === 'user' ? 'mr-auto max-w-[85%] rounded-2xl rounded-tr-sm bg-brand-600 p-3 text-sm text-white' : 'max-w-[92%] rounded-2xl rounded-tl-sm border border-slate-200 bg-white p-3 text-sm shadow-sm'}>{chat.transcript && <p className="mb-2 border-b border-white/30 pb-2 text-xs opacity-80">🎙️ {chat.transcript}</p>}<p>{chat.text}</p>{chat.action && chat.action.status !== 'PROPOSED' && <p className="mt-2 text-xs font-bold text-slate-500">{chat.action.status === 'SENT' ? '✅ المهمة انرسلت' : chat.action.status === 'CANCELLED' ? '✖️ المسودة انلغت' : '⚠️ ما انرسلت'}</p>}{chat.cards.map((card) => <article key={card.title} className="mt-3 rounded-xl border border-slate-200 bg-slate-50 p-3 text-slate-800"><div className="flex items-center justify-between gap-2"><b>{card.title}</b><span className="text-xs text-brand-600">{card.status}</span></div>{card.owner && <p className="mt-1 text-xs text-slate-500">المسؤول: {card.owner}</p>}<ul className="mt-2 space-y-1 text-xs text-slate-600">{card.lines.map((line) => <li key={line}>• {line}</li>)}</ul>{card.link && <Link className="mt-3 inline-block text-xs font-bold text-brand-600" to={card.link}>افتح التفاصيل ←</Link>}</article>)}<span className="mt-2 block text-[10px] opacity-60">{time(chat.at)}</span></div>)}</div></section>
    </div>
    {proposal && <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-950/45 p-4"><div className="mx-auto mt-8 max-w-lg rounded-3xl bg-white p-5 shadow-2xl"><p className="text-xs font-bold text-brand-600">مسودة مهمة — ما تنرسل بدون تأكيدك</p><h2 className="mt-2 text-lg font-black">{proposal.summary}</h2><div className="mt-4 space-y-3"><label className="block text-sm font-bold">الموظف<select value={proposal.fields.employeeId} onChange={(e) => { const x = proposal.candidates.find((c) => c.id === e.target.value); if (x) setProposal({ ...proposal, fields: { ...proposal.fields, employeeId: x.id, employeeName: x.name } }) }} className="mt-1 w-full rounded-xl border border-slate-200 p-2 font-normal">{!proposal.fields.employeeId && <option value="">اختار الموظف…</option>}{(proposal.candidates.length ? proposal.candidates : [{ id: proposal.fields.employeeId, name: proposal.fields.employeeName }]).filter((c) => c.id).map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select></label><label className="block text-sm font-bold">عنوان المهمة<input value={proposal.fields.title} onChange={(e) => setProposal({ ...proposal, fields: { ...proposal.fields, title: e.target.value } })} className="mt-1 w-full rounded-xl border border-slate-200 p-2 font-normal" /></label><p className="text-sm text-slate-600">{proposal.fields.priority === 'URGENT' ? '⚡ عاجلة' : 'عادية'}</p><label className="block text-sm font-bold">⏳ آخر موعد للإنجاز <span className="text-red-600">*</span><input type="datetime-local" required value={proposal.fields.dueAt || ''} onChange={(e) => setProposal({ ...proposal, fields: { ...proposal.fields, dueAt: e.target.value } })} className="mt-1 w-full rounded-xl border border-slate-200 p-2 font-normal" /></label><label className="block text-sm font-bold">التفاصيل<textarea value={proposal.fields.description} onChange={(e) => setProposal({ ...proposal, fields: { ...proposal.fields, description: e.target.value } })} className="mt-1 min-h-20 w-full rounded-xl border border-slate-200 p-2 font-normal" /></label></div><div className="mt-5 flex gap-2"><button onClick={sendProposal} disabled={busy} className="flex-1 disabled:opacity-60 rounded-xl bg-brand-600 py-3 text-sm font-bold text-white">إرسال المهمة</button><button onClick={cancelProposal} className="rounded-xl border border-slate-200 px-4 text-sm font-bold">إلغاء</button></div></div></div>}
    <div className="fixed inset-x-0 bottom-0 z-30 border-t border-slate-200 bg-white/95 p-3 backdrop-blur"><div className="mx-auto flex max-w-2xl items-end gap-2"><button onPointerDown={startVoice} onPointerUp={stopVoice} onPointerLeave={stopVoice} className={`grid h-12 w-12 shrink-0 place-items-center rounded-full text-xl text-white shadow-lg ${recording ? 'animate-pulse bg-red-600' : 'bg-brand-600'}`} aria-label="اضغط مطولاً وتكلم">🎙️</button><form onSubmit={(e) => { e.preventDefault(); send() }} className="flex flex-1 gap-2 rounded-2xl border border-slate-200 bg-slate-50 p-1"><input value={input} onChange={(e) => setInput(e.target.value)} placeholder={recording ? 'دا أسمعك…' : busy ? 'دا أفكر…' : 'اكتب طلبك للسكرتير…'} className="min-w-0 flex-1 bg-transparent px-3 text-sm outline-none" /><button className="rounded-xl bg-slate-900 px-4 text-sm font-bold text-white">إرسال</button></form></div></div>
  </main>
}
