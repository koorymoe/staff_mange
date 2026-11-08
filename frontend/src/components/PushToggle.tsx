import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// ═══ 🔔 (ع) 11-08: تفعيل إشعارات الموبايل ═══
// بعد التفعيل، كل إشعار يوصلك على الجهاز حتى والموقع مسدود. كل جهاز يتفعّل لحاله.
// ⚠️ بالآيفون لازم الموقع يكون مضاف للشاشة الرئيسية (Add to Home Screen) أول.
type State = 'checking' | 'unsupported' | 'denied' | 'off' | 'on' | 'busy'

const keyBytes = (b64: string) => {
  const pad = '='.repeat((4 - (b64.length % 4)) % 4)
  const raw = atob((b64 + pad).replace(/-/g, '+').replace(/_/g, '/'))
  return Uint8Array.from(raw, (c) => c.charCodeAt(0))
}

const supported = () =>
  typeof window !== 'undefined' && 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window

async function subscribe() {
  const reg = await navigator.serviceWorker.ready
  const { publicKey } = await apiRequest<{ publicKey: string }>('/push/key')
  const sub = (await reg.pushManager.getSubscription())
    ?? (await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: keyBytes(publicKey) }))
  await apiRequest('/push/subscribe', { method: 'POST', body: JSON.stringify(sub.toJSON()) })
}

export default function PushToggle() {
  const [state, setState] = useState<State>('checking')
  const [err, setErr] = useState('')

  useEffect(() => {
    let alive = true
    const done = (s: State) => { if (alive) setState(s) }
    if (!supported()) { done('unsupported'); return }
    if (Notification.permission === 'denied') { done('denied'); return }
    navigator.serviceWorker.getRegistration().then(async (reg) => {
      const sub = reg ? await reg.pushManager.getSubscription() : null
      // جهاز مفعّل من قبل: نجدد الاشتراك بصمت (حتى لو تبدّل صاحب الجلسة)
      if (sub && Notification.permission === 'granted') {
        await apiRequest('/push/subscribe', { method: 'POST', body: JSON.stringify(sub.toJSON()) }).catch(() => {})
        done('on')
      } else done('off')
    }).catch(() => done('off'))
    return () => { alive = false }
  }, [])

  const enable = async () => {
    setErr(''); setState('busy')
    try {
      const p = await Notification.requestPermission()
      if (p !== 'granted') { setState(p === 'denied' ? 'denied' : 'off'); return }
      await subscribe()
      setState('on')
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر التفعيل')
      setState('off')
    }
  }
  const disable = async () => {
    setState('busy')
    try {
      const reg = await navigator.serviceWorker.getRegistration()
      const sub = reg ? await reg.pushManager.getSubscription() : null
      if (sub) {
        await apiRequest('/push/unsubscribe', { method: 'POST', body: JSON.stringify({ endpoint: sub.endpoint }) }).catch(() => {})
        await sub.unsubscribe()
      }
    } finally { setState('off') }
  }

  if (state === 'checking') return null
  return (
    <div className="border-b border-slate-100 bg-slate-50 px-4 py-2 text-xs">
      {state === 'unsupported' && <p className="text-slate-500">📵 هالمتصفح ما يدعم إشعارات الموبايل. بالآيفون: ضيف الموقع للشاشة الرئيسية أول.</p>}
      {state === 'denied' && <p className="text-slate-500">🔕 الإشعارات مقفولة من إعدادات المتصفح — فعّلها من هناك.</p>}
      {(state === 'off' || state === 'busy') && (
        <button type="button" disabled={state === 'busy'} onClick={() => void enable()} className="w-full rounded-lg bg-brand-600 px-3 py-1.5 font-bold text-white disabled:opacity-50">
          🔔 فعّل الإشعارات على هالجهاز
        </button>
      )}
      {state === 'on' && (
        <div className="flex items-center justify-between">
          <span className="font-bold text-emerald-700">🔔 الإشعارات مفعّلة على هالجهاز</span>
          <button type="button" onClick={() => void disable()} className="text-slate-500 underline">طفّيها</button>
        </div>
      )}
      {err && <p className="mt-1 text-red-600">{err}</p>}
    </div>
  )
}
