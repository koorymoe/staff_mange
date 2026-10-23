import { useEffect, useState } from 'react'
import { api } from '../api'

// ═══ «⏱️ وقت العمل» — (ع) 10-16 ═══
// ثلاث مصادر للوقت (GPS، ضغطات الموظف، ورقة التنفيذ) وماتركس ياخذ الأوثق
// الموجود. هذا السطر يبيّن شنو اعتمد ومن وين — حتى المراقب يشوف نفس الرقم.
const SRC: Record<string, string> = { GPS: 'GPS', TAP: 'ضغطة الموظف', SHEET: 'التقرير', DEPART_30: 'الانطلاق + نص ساعة' }
const hm = (s: string | null) => (s ? new Date(s.endsWith('Z') ? s : s + 'Z').toLocaleTimeString('ar-IQ', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit' }) : '—')

export default function WorkWindowLine({ bookingId }: { bookingId: string }) {
  const [w, setW] = useState<{ startAt: string | null; endAt: string | null; startSrc: string | null; endSrc: string | null } | null>(null)
  useEffect(() => {
    let alive = true
    api.getBookingWorkWindow(bookingId).then((x) => { if (alive) setW(x) }).catch(() => {})
    return () => { alive = false }
  }, [bookingId])
  if (!w || (!w.startAt && !w.endAt)) return null
  const mins = w.startAt && w.endAt ? Math.round((new Date(w.endAt).getTime() - new Date(w.startAt).getTime()) / 60000) : null
  return (
    <p className="rounded-lg bg-indigo-50 px-3 py-1.5 text-xs text-indigo-900">
      ⏱️ وقت العمل الي يعتمده ماتركس: من <b>{hm(w.startAt)}</b> ({SRC[w.startSrc ?? ''] ?? '—'}) لـ <b>{hm(w.endAt)}</b> ({SRC[w.endSrc ?? ''] ?? '—'})
      {mins != null && mins > 0 && <> · <b>{Math.floor(mins / 60)}س {mins % 60}د</b></>}
    </p>
  )
}
