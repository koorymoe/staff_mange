import { useState } from 'react'
import { apiRequest } from '../api'

// 🧭 «رتّب مسار اليوم» (دازاي 11-06، فكرة ٢): حجوزاتي اليوم بأقصر ترتيب من الشركة.
// اقتراح بس — موعد الزبون أهم من المسافة، والليدر يقرر.
type Plan = {
  date: string
  plan: { stops: { id: string; label: string }[]; km: number; originalKm: number; savedKm: number; mapsUrl: string }
  noLocation: { id: string; label: string }[]
}

export default function RoutePlanCard() {
  const [p, setP] = useState<Plan | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const load = async () => {
    setBusy(true)
    setErr('')
    try {
      setP(await apiRequest<Plan>('/route-plan'))
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر')
    } finally {
      setBusy(false)
    }
  }
  if (!p)
    return (
      <div className="flex flex-wrap items-center gap-2">
        <button type="button" disabled={busy} onClick={() => void load()} className="rounded-xl border border-brand-200 bg-white px-3 py-1.5 text-sm font-bold text-brand-700 disabled:opacity-50">
          {busy ? '…' : '🧭 رتّب مسار اليوم'}
        </button>
        {err && <span role="alert" className="text-xs text-rose-700">{err}</span>}
      </div>
    )
  const { plan } = p
  return (
    <section className="space-y-2 rounded-2xl border border-brand-100 bg-white p-3 text-sm" aria-label="مسار اليوم">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <b>🧭 مسار اليوم المقترح</b>
        <button type="button" onClick={() => setP(null)} className="text-xs text-slate-500">سد</button>
      </div>
      {plan.stops.length === 0 ? (
        <p className="text-slate-500">ماكو حجوزات بيها موقع اليوم.</p>
      ) : (
        <>
          <ol className="list-inside list-decimal space-y-0.5">
            {plan.stops.map((s) => <li key={s.id}>{s.label}</li>)}
          </ol>
          <p className="text-xs text-slate-600">
            {plan.km} كم بخط مستقيم
            {plan.savedKm > 0 ? ` — أقصر بـ${plan.savedKm} كم من ترتيب المواعيد (${plan.originalKm} كم)` : ' — ترتيب المواعيد أصلاً هو الأقصر'}
          </p>
          <p className="text-[11px] text-amber-700">⚠️ اقتراح بالمسافة بس — إذا متفق ويا زبون على ساعة، الموعد أهم.</p>
          {plan.mapsUrl && (
            <a href={plan.mapsUrl} target="_blank" rel="noopener noreferrer" className="inline-block rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white">
              🗺️ افتح بخرائط Google
            </a>
          )}
        </>
      )}
      {p.noLocation.length > 0 && (
        <p className="text-xs text-slate-500">بلا موقع على الخريطة: {p.noLocation.map((x) => x.label).join('، ')}</p>
      )}
    </section>
  )
}
