import { useCallback, useEffect, useRef, useState } from 'react'
import { api } from '../api'

// 🔗 ربط حساب الموظف ببوت بلاغات النظام (دازاي 11-05).
// الربط مرة وحدة، والفك بس من المالك (دازاي 11-05).
// الزر يفتح التلغرام على البوت بضغطة وحدة، والربط يصير تلقائياً (رمز مرة وحدة، ١٠ دقايق).
// بلا يوزر وباسورد بالتلغرام: الباسورد يبقى مكتوب بالمحادثة ويطّلع عليه أي أحد يمسك التلفون.
export default function TelegramLinkCard() {
  const [st, setSt] = useState<Awaited<ReturnType<typeof api.getMyTelegram>> | null>(null)
  const [link, setLink] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const poll = useRef<number | null>(null)

  const load = useCallback(() => api.getMyTelegram().then((s) => {
    setSt(s)
    if (s.linked && poll.current) { window.clearInterval(poll.current); poll.current = null; setLink('') }
  }).catch(() => setErr('تعذر جلب حالة الربط')), [])
  useEffect(() => {
    load()
    return () => { if (poll.current) window.clearInterval(poll.current) }
  }, [load])

  const start = async () => {
    setBusy(true); setErr('')
    try {
      const r = await api.createTelegramLink()
      setLink(r.url)
      window.open(r.url, '_blank', 'noopener')
      // ننتظر الربط دقيقتين (كل ٣ ثواني) حتى تتحدّث الحالة لحالها
      if (poll.current) window.clearInterval(poll.current)
      let n = 0
      poll.current = window.setInterval(() => { if (++n > 40 && poll.current) { window.clearInterval(poll.current); poll.current = null } load() }, 3000)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر إنشاء الرابط')
    } finally { setBusy(false) }
  }

  if (!st) return <p className="text-xs text-slate-400">جاري التحميل…</p>
  if (!st.enabled) return <p className="text-sm text-slate-500">بوت التلغرام مو مفعّل بعد.</p>
  if (st.linked) {
    return (
      <div className="space-y-2 text-sm">
        <p className="font-bold text-emerald-700">✅ حسابك مربوط بالتلغرام{st.tgUsername ? ` (@${st.tgUsername})` : ''}</p>
        <p className="text-xs text-slate-500">إذا واجهتك مشكلة بالنظام، دزها للبوت (صورة + شرح واضح)، أو من زر «🆘 بلّغ» فوق.</p>
        <p className="text-xs text-slate-400">الربط مرة وحدة. إذا تحتاج تغيّر التلغرام، المالك يفك الربط.</p>
      </div>
    )
  }
  return (
    <div className="space-y-2 text-sm">
      <p className="text-slate-600">اربط حسابك ببوت التلغرام حتى تگدر تبلّغ عن أي مشكلة بالنظام: صورة + شرح، ويوصلك ردّنا هناك.</p>
      <button onClick={start} disabled={busy} className="w-full rounded-lg bg-sky-500 py-2 font-bold text-white hover:bg-sky-600 disabled:opacity-50">
        {busy ? 'لحظة…' : '🔗 اربط حسابي بالتلغرام'}
      </button>
      {link && (
        <p className="text-xs text-slate-500">
          إذا ما انفتح التلغرام، <a href={link} target="_blank" rel="noreferrer" className="font-bold text-sky-600 underline">اضغط هنا</a>، وبعدين «ابدأ». الرابط يصلح ١٠ دقايق ولمرة وحدة.
        </p>
      )}
      {err && <p className="text-xs text-red-600">{err}</p>}
    </div>
  )
}
