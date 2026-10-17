import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type CourierMessage } from '../api'
import { useSession } from '../session'
import MatrixEyeGraphic from './MatrixEyeGraphic'
import { pollVisible } from '../utils/poll'
import { COURIER_TRIP_MS, tweenLid } from '../utils/courier'

// ═══ رسول ماتركس — قرار (ع) 10-13 ═══
//
// المالك/المدير يوجّه شغل (مهمة، إجازة، مشروع، حجز) → شارنغانه يتحول
// مانغكيو ويغمض ويروح (`matrix-courier-go` لـMatrixEye). بعد ثواني الطريق
// يطلع يم عين الموظف — **بس لو فاتح النظام** (الطلب ما ينرسل والتبويب
// مخفي) — يسلّمه الرسالة ويرجع. الرحلة كلها ~٢٠ ثانية.
//
// والأرقام على البنود حقيقية: عدد الي توجّهله وما فتح بنده بعد (من الخادم).

const TRAVEL_MS = 6_000
const SHOW_MS = 9_000

export default function MatrixCourier() {
  const { employee } = useSession()
  const isBoss = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
  const [visit, setVisit] = useState<CourierMessage | null>(null)
  const [lid, setLid] = useState(0)
  const busy = useRef(false)
  const sent = useRef<Set<string>>(new Set())

  useEffect(() => {
    if (!employee) return
    let alive = true
    const load = () => api.getCourierState().then((st) => {
      if (!alive) return
      window.dispatchEvent(new CustomEvent('matrix-courier-badges', { detail: st.badges }))
      const now = new Date(st.now).getTime()
      // المرسل: كل توجيه جديد = رحلة وحدة للعين
      for (const m of st.outgoing) {
        const age = now - new Date(m.createdAt).getTime()
        if (sent.current.has(m.id) || age > COURIER_TRIP_MS - 4000) continue
        sent.current.add(m.id)
        window.dispatchEvent(new CustomEvent('matrix-courier-go', { detail: { age, name: m.recipientName } }))
        break
      }
      // المستلم: العين توصل بعد وقت الطريق، وحدة وحدة
      if (!busy.current) {
        const m = st.incoming.find((x) => now - new Date(x.createdAt).getTime() >= TRAVEL_MS)
        if (m) {
          busy.current = true
          setVisit(m)
          setLid(0)
          window.setTimeout(() => tweenLid(0, 0.95, 900, setLid), 200)
          window.setTimeout(() => tweenLid(0.95, 0, 1200, setLid), SHOW_MS - 1300)
          window.setTimeout(() => {
            setVisit(null)
            busy.current = false
            void api.courierDelivered(m.id).catch(() => {})
          }, SHOW_MS)
        }
      }
    }).catch(() => {})
    load()
    // المدير يحتاج يعرف بسرعة إنه وجّه (حتى العين تروح بوقتها)؛ الموظف أبطأ شوية
    const stop = pollVisible(load, isBoss ? 3000 : 5000)
    window.addEventListener('matrix-courier-refresh', load)
    return () => { alive = false; stop(); window.removeEventListener('matrix-courier-refresh', load) }
  }, [employee, isBoss])

  if (!visit) return null
  return (
    <div dir="rtl" className="matrix-courier">
      <span className="matrix-courier-eye" style={{ ['--eye' as string]: '#b91c1c' }}>
        <MatrixEyeGraphic group={'ADMINS'} mood={'CALM'} lid={lid} flash={false} width={64} sharingan={1} />
      </span>
      <div className="matrix-courier-say">
        <b>ماتركس:</b> {visit.message}
        <Link to={visit.link} className="ms-2 font-bold text-red-700 underline">افتحه ←</Link>
      </div>
    </div>
  )
}

// رسائل ماتركس الي وصلت — يرجع يقراها من عينه.
export function CourierInbox({ onGo }: { onGo: () => void }) {
  const [rows, setRows] = useState<CourierMessage[] | null>(null)
  useEffect(() => { api.getCourierInbox().then(setRows).catch(() => setRows([])) }, [])
  if (!rows || rows.length === 0) return null
  return (
    <div className="mt-2">
      <p className="mb-1 text-xs font-bold text-slate-500">📨 رسائل ماتركس من المدير</p>
      <ul className="max-h-52 space-y-1 overflow-y-auto">
        {rows.map((m) => (
          <li key={m.id} className="rounded-lg bg-red-50/60 px-2 py-1.5">
            <Link to={m.link} onClick={onGo} className="block text-slate-800 hover:underline">{m.message}</Link>
            <p className="text-[11px] text-slate-500">{new Date(m.createdAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'short', timeStyle: 'short' })}</p>
          </li>
        ))}
      </ul>
    </div>
  )
}
