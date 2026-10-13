import { useEffect, useState } from 'react'
import { api, type Justification } from '../api'
import { useSession } from '../session'
import { pollVisible } from '../utils/poll'

// ═══ «المراقب يطلب تبرير» — (ع) 10-13 ═══
// يطلع للموظف نفسه أول ما يفتح النظام: يقرا شنو المطلوب ويكتب عذره، والرد
// يرجع للمراقب. وبعد ما يحكم، النتيجة توصله كإشعار.
export default function JustificationPrompt() {
  const { employee } = useSession()
  const [pending, setPending] = useState<Justification[]>([])
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [later, setLater] = useState<string | null>(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    if (!employee) return
    let alive = true
    const load = () => api.getMyJustifications().then((r) => { if (alive) setPending(r.filter((j) => !j.answeredAt)) }).catch(() => {})
    load()
    const stop = pollVisible(load, 60_000)
    return () => { alive = false; stop() }
  }, [employee])

  const cur = pending.find((j) => j.id !== later)
  if (!cur) return null
  const send = async () => {
    setBusy(true); setErr('')
    try { await api.answerJustification(cur.id, text); setText(''); setPending((p) => p.filter((j) => j.id !== cur.id)) }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذّر الإرسال') }
    finally { setBusy(false) }
  }
  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/40 p-4" dir="rtl">
      <div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-2xl">
        <h3 className="text-lg font-extrabold text-[#0f2040]">📝 المراقب يطلب تبرير</h3>
        <p className="mt-1 text-[11px] text-slate-500">من {cur.monitorName || 'المراقب'} · {new Date(cur.createdAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad' })}{pending.length > 1 ? ` · عندك ${pending.length} طلبات` : ''}</p>
        <p className="mt-3 rounded-xl bg-amber-50 px-3 py-2 text-sm text-amber-900">{cur.question}</p>
        <textarea value={text} onChange={(e) => setText(e.target.value)} rows={4} placeholder="اكتب عذرك أو شنو صار بالضبط…"
          className="mt-3 w-full rounded-xl border border-slate-300 p-3 text-sm outline-none focus:border-brand-500" />
        {err && <p className="mt-1 text-xs text-red-600">{err}</p>}
        <div className="mt-3 flex gap-2">
          <button disabled={busy || text.trim().length < 3} onClick={() => void send()} className="flex-1 rounded-xl bg-[#0f2040] py-2.5 text-sm font-bold text-white disabled:opacity-40">إرسال التبرير للمراقب</button>
          <button onClick={() => setLater(cur.id)} className="rounded-xl border border-slate-300 px-4 text-sm text-slate-600">بعدين</button>
        </div>
        <p className="mt-2 text-[10px] text-slate-400">هذا توثيق ومتابعة — ماكو غرامة.</p>
      </div>
    </div>
  )
}
