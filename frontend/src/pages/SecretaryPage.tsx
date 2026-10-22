import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { useSession } from '../session'

type SecretaryCard = {
  kind: 'booking' | 'project' | 'customer' | 'employee' | 'attendance' | 'task' | 'summary' | 'info'
  title: string; status: string; owner: string | null; updatedAt: string | null; lines: string[]; link: string | null
}
type Proposal = {
  actionId: string; kind: 'task'; summary: string
  fields: { employeeId: string; employeeName: string; title: string; description: string; priority: 'NORMAL' | 'URGENT'; dueAt: string }
  candidates: { id: string; name: string }[]
}
type Chat = { id: string; at: string; role: 'user' | 'assistant'; text: string; transcript: string | null; cards: SecretaryCard[]; action?: { id: string; status: 'PROPOSED' | 'SENT' | 'CANCELLED' | 'FAILED'; result: string } }

const mockToday = {
  greeting: 'صباح الخير دازاي',
  summary: { lines: ['٢ حجز يحتاج متابعة اليوم', 'شكوى وحدة تنتظر الرد', '٤ مهام موعدها اليوم'], cards: [
    { kind: 'summary', title: 'ملخص الشغل اليوم', status: 'يحتاج متابعة', owner: null, updatedAt: new Date().toISOString(), lines: ['أهم الأولويات مرتبة حتى ما تضيع منك'], link: null },
  ] as SecretaryCard[] },
  pending: [
    { id: 'leave:mock-1', kind: 'leave', question: 'إجازة أحمد خليل يوم الخميس — تحتاج مراجعة', why: 'الطلب وصل للإدارة', impact: 'يبقى جدول الدوام يحتاج ترتيب قبل الخميس', affects: 'أحمد خليل وفريقه', link: '/leaves', createdAt: new Date().toISOString() },
    { id: 'justification:mock-2', kind: 'justification', question: 'تبرير تأخر تقرير العمل ينتظر قراءتك', why: 'الموظف أرسل التبرير', impact: 'يبقى البند مفتوحاً لحين المراجعة', affects: 'فريق التنفيذ', link: '/matrix/decisions', createdAt: new Date().toISOString() },
  ],
}

const mockReply = (text: string): { reply: string; cards: SecretaryCard[]; proposal: Proposal | null } => {
  if (/مهمة|ودّ|ودي|كلّف/.test(text)) return {
    reply: 'حضّرت المهمة، راجعها قبل الإرسال.', cards: [], proposal: {
      actionId: 'task-mock-1', kind: 'task', summary: 'فهمت منك: مهمة عاجلة لأحمد خليل — متابعة الحجز 8850 اليوم',
      fields: { employeeId: 'emp-1', employeeName: 'أحمد خليل', title: 'متابعة الحجز 8850', description: 'راجع حالة الحجز وتواصل مع الفريق اليوم.', priority: 'URGENT', dueAt: '2026-10-22T18:00' },
      candidates: [{ id: 'emp-1', name: 'أحمد خليل' }, { id: 'emp-2', name: 'أحمد حسين' }],
    },
  }
  if (/دوام|مداوم|بصمة/.test(text)) return { reply: 'هذا ملخص دوام الصبح حتى الآن.', proposal: null, cards: [{ kind: 'attendance', title: 'دوام الصباح', status: 'محدّث', owner: null, updatedAt: new Date().toISOString(), lines: ['٨ مداومين', '٢ متأخرين', 'موظف واحد بإجازة موثقة'], link: '/attendance' }] }
  return { reply: 'لقيت آخر حالة مرتبطة بطلبك. هذا المختصر، وتكدر تفتح التفاصيل.', proposal: null, cards: [{ kind: 'booking', title: 'حجز 8850 — كرار جعفر', status: 'قيد العمل', owner: 'فريق التنفيذ', updatedAt: new Date().toISOString(), lines: ['آخر تحديث: بانتظار تنفيذ الزيارة', 'الموعد المطلوب: اليوم'], link: '/bookings?focus=8850' }] }
}

const time = (iso: string | null) => iso ? new Date(iso).toLocaleTimeString('ar-IQ', { hour: 'numeric', minute: '2-digit' }) : ''

export default function SecretaryPage() {
  const { employee } = useSession()
  const allowed = employee?.actualRole === 'OWNER' || employee?.role === 'ADMIN'
  const [input, setInput] = useState('')
  const [chats, setChats] = useState<Chat[]>([{ id: 'welcome', at: new Date().toISOString(), role: 'assistant', text: 'آني سكرتيرك. گلي شتريد أعرف أو أتابع؟', transcript: null, cards: [] }])
  const [proposal, setProposal] = useState<Proposal | null>(null)
  const [recording, setRecording] = useState(false)
  const [deferred, setDeferred] = useState<string[]>([])
  const holdTimer = useRef<number | null>(null)

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

  const send = (raw = input, transcript: string | null = null) => {
    const text = raw.trim(); if (!text) return
    const answer = mockReply(text)
    setChats((v) => [...v, { id: `u-${Date.now()}`, at: new Date().toISOString(), role: 'user', text, transcript, cards: [] }, { id: `a-${Date.now()}`, at: new Date().toISOString(), role: 'assistant', text: answer.reply, transcript: null, cards: answer.cards }])
    setProposal(answer.proposal); setInput('')
  }
  const startVoice = () => { holdTimer.current = window.setTimeout(() => setRecording(true), 180) }
  const stopVoice = () => { if (holdTimer.current) window.clearTimeout(holdTimer.current); if (!recording) return; setRecording(false); const transcript = 'شوفلي منو مداوم الصبح'; send(transcript, transcript) }
  const sendProposal = () => { if (!proposal) return; setChats((v) => [...v, { id: `sent-${Date.now()}`, at: new Date().toISOString(), role: 'assistant', text: 'تم إرسال المهمة ومتابعتها من السكرتير.', transcript: null, cards: [{ kind: 'task', title: proposal.fields.title, status: 'انرسلت', owner: proposal.fields.employeeName, updatedAt: new Date().toISOString(), lines: ['الموعد: ' + (proposal.fields.dueAt || 'بدون موعد')], link: '/my-tasks' }], action: { id: proposal.actionId, status: 'SENT', result: 'تم الإرسال' } }]); setProposal(null) }

  return <main dir="rtl" className="min-h-dvh bg-[radial-gradient(circle_at_top,#dbeafe,transparent_42%),#f8fafc] text-slate-800">
    <div className="mx-auto min-h-dvh max-w-2xl px-4 pb-36 pt-5 sm:px-6">
      <header className="flex items-start justify-between"><div><p className="text-sm font-bold text-brand-600">السكرتير</p><h1 className="mt-1 text-2xl font-black text-slate-900">{mockToday.greeting}</h1><p className="mt-1 text-sm text-slate-500">كل المهم قدامك، بلا ما تدوّر بالنظام.</p></div><div className="rounded-2xl bg-amber-100 px-3 py-2 text-center text-xs font-bold text-amber-900"><b className="block text-lg leading-none">{mockToday.pending.length}</b>ينتظر قرارك</div></header>
      <section className="mt-5 rounded-3xl border border-white/70 bg-white/85 p-4 shadow-[0_18px_50px_-30px_rgba(15,32,64,.5)] backdrop-blur"><div className="flex items-center justify-between"><h2 className="font-extrabold">ملخص اليوم</h2><span className="text-xs text-slate-400">محدّث هسه</span></div><ul className="mt-3 space-y-2">{mockToday.summary.lines.map((line) => <li key={line} className="flex gap-2 text-sm"><span className="text-brand-500">●</span>{line}</li>)}</ul></section>
      <section className="mt-5"><div className="mb-3 flex items-center justify-between"><h2 className="font-extrabold">ينتظر قرارك</h2><span className="text-xs text-slate-400">{deferred.length ? `${deferred.length} مؤجل` : 'افتحه من مكانه'}</span></div><div className="space-y-3">{mockToday.pending.map((item) => <article key={item.id} className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm"><p className="font-bold">{item.question}</p><p className="mt-2 text-sm text-slate-500">{item.why}</p><p className="mt-1 text-xs text-slate-400">الأثر: {item.impact}</p><div className="mt-3 flex gap-2"><Link to={item.link} className="rounded-xl bg-brand-600 px-3 py-2 text-xs font-bold text-white">افتح التفاصيل</Link><button onClick={() => setDeferred((v) => [...v, item.id])} className="rounded-xl border border-slate-200 px-3 py-2 text-xs font-bold text-slate-600">أجّل</button></div></article>)}</div></section>
      <section className="mt-6"><h2 className="mb-3 font-extrabold">آخر المحادثات</h2><div className="space-y-3">{chats.map((chat) => <div key={chat.id} className={chat.role === 'user' ? 'mr-auto max-w-[85%] rounded-2xl rounded-tr-sm bg-brand-600 p-3 text-sm text-white' : 'max-w-[92%] rounded-2xl rounded-tl-sm border border-slate-200 bg-white p-3 text-sm shadow-sm'}>{chat.transcript && <p className="mb-2 border-b border-white/30 pb-2 text-xs opacity-80">🎙️ {chat.transcript}</p>}<p>{chat.text}</p>{chat.cards.map((card) => <article key={card.title} className="mt-3 rounded-xl border border-slate-200 bg-slate-50 p-3 text-slate-800"><div className="flex items-center justify-between gap-2"><b>{card.title}</b><span className="text-xs text-brand-600">{card.status}</span></div>{card.owner && <p className="mt-1 text-xs text-slate-500">المسؤول: {card.owner}</p>}<ul className="mt-2 space-y-1 text-xs text-slate-600">{card.lines.map((line) => <li key={line}>• {line}</li>)}</ul>{card.link && <Link className="mt-3 inline-block text-xs font-bold text-brand-600" to={card.link}>افتح التفاصيل ←</Link>}</article>)}<span className="mt-2 block text-[10px] opacity-60">{time(chat.at)}</span></div>)}</div></section>
    </div>
    {proposal && <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-950/45 p-4"><div className="mx-auto mt-8 max-w-lg rounded-3xl bg-white p-5 shadow-2xl"><p className="text-xs font-bold text-brand-600">مسودة مهمة — ما تنرسل بدون تأكيدك</p><h2 className="mt-2 text-lg font-black">{proposal.summary}</h2><div className="mt-4 space-y-3"><label className="block text-sm font-bold">الموظف<select value={proposal.fields.employeeId} onChange={(e) => { const x = proposal.candidates.find((c) => c.id === e.target.value); if (x) setProposal({ ...proposal, fields: { ...proposal.fields, employeeId: x.id, employeeName: x.name } }) }} className="mt-1 w-full rounded-xl border border-slate-200 p-2 font-normal">{proposal.candidates.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select></label><label className="block text-sm font-bold">عنوان المهمة<input value={proposal.fields.title} onChange={(e) => setProposal({ ...proposal, fields: { ...proposal.fields, title: e.target.value } })} className="mt-1 w-full rounded-xl border border-slate-200 p-2 font-normal" /></label><label className="block text-sm font-bold">التفاصيل<textarea value={proposal.fields.description} onChange={(e) => setProposal({ ...proposal, fields: { ...proposal.fields, description: e.target.value } })} className="mt-1 min-h-20 w-full rounded-xl border border-slate-200 p-2 font-normal" /></label></div><div className="mt-5 flex gap-2"><button onClick={sendProposal} className="flex-1 rounded-xl bg-brand-600 py-3 text-sm font-bold text-white">إرسال المهمة</button><button onClick={() => setProposal(null)} className="rounded-xl border border-slate-200 px-4 text-sm font-bold">إلغاء</button></div></div></div>}
    <div className="fixed inset-x-0 bottom-0 z-30 border-t border-slate-200 bg-white/95 p-3 backdrop-blur"><div className="mx-auto flex max-w-2xl items-end gap-2"><button onPointerDown={startVoice} onPointerUp={stopVoice} onPointerLeave={stopVoice} className={`grid h-12 w-12 shrink-0 place-items-center rounded-full text-xl text-white shadow-lg ${recording ? 'animate-pulse bg-red-600' : 'bg-brand-600'}`} aria-label="اضغط مطولاً وتكلم">🎙️</button><form onSubmit={(e) => { e.preventDefault(); send() }} className="flex flex-1 gap-2 rounded-2xl border border-slate-200 bg-slate-50 p-1"><input value={input} onChange={(e) => setInput(e.target.value)} placeholder={recording ? 'دا أسمعك…' : 'اكتب طلبك للسكرتير…'} className="min-w-0 flex-1 bg-transparent px-3 text-sm outline-none" /><button className="rounded-xl bg-slate-900 px-4 text-sm font-bold text-white">إرسال</button></form></div></div>
  </main>
}
