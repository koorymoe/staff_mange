import { useEffect, useId, useState } from 'react'
import { pendingPhotos, uploadOrQueue } from '../utils/offlinePhotos'

// 📸 صورة «بعد الشغل» من «مهامي» (دازاي 11-06): قبل چانت بشاشة الديكور بس، فالخدمة الي
// تطلب صور قبل «تم» ما چان الفني الهندسي يگدر يلبّيها. وبلا نت تنحفظ وترفع وحدها.
export default function WorkPhotoButton({ bookingId }: { bookingId: string }) {
  const inputId = useId()
  const [sent, setSent] = useState(0)
  const [queued, setQueued] = useState(0)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  useEffect(() => {
    const upd = () => void pendingPhotos(bookingId).then((q) => setQueued(q.length))
    upd()
    window.addEventListener('photo-queue', upd)
    return () => window.removeEventListener('photo-queue', upd)
  }, [bookingId])
  const onFile = async (input: HTMLInputElement) => {
    const f = input.files?.[0]
    if (!f) return
    setBusy(true)
    setErr('')
    try {
      const r = await uploadOrQueue(bookingId, 'AFTER', f)
      if (!r.queued) setSent((n) => n + 1)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر رفع الصورة')
    } finally {
      setBusy(false)
      input.value = ''
    }
  }
  return (
    <div className="flex flex-wrap items-center gap-2 text-xs">
      <label htmlFor={inputId} className={`cursor-pointer rounded-lg border border-slate-300 bg-white px-3 py-2 font-bold text-slate-700 ${busy ? 'opacity-50' : ''}`}>
        {busy ? '…' : '📸 صورة بعد الشغل'}
      </label>
      <input id={inputId} type="file" accept="image/*" capture="environment" className="hidden" disabled={busy} onChange={(e) => void onFile(e.currentTarget)} />
      {sent > 0 && <span className="text-emerald-700">✅ انرفع {sent}</span>}
      {queued > 0 && <span className="text-amber-700">⏳ {queued} محفوظة بالموبايل — ترفع من يرجع النت</span>}
      {err && <span role="alert" className="text-rose-700">{err}</span>}
    </div>
  )
}
