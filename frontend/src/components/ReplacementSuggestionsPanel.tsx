import { useEffect, useState } from 'react'
import { api, type ReplacementSuggestions } from '../api'
import { useSession } from '../session'

// ═══ ماتركس — اقتراح استبدال ═══
// نفس حارس GET /api/ai/replacement-suggestions: ADMIN (والمالك) أو
// it_assets أو vehicle_management. اللي ما يعبر الحارس ما نطلب له أصلاً.
// الخادم يفلتر: صاحب it_assets يشوف الأجهزة، وصاحب vehicle_management المركبات.
// اقتراح بس — القرار للإدارة.

function canSeeReplacements(role: string | undefined, permissions: string[], part: 'it' | 'vehicles' | 'any') {
  if (role === 'ADMIN') return true
  if (part === 'it') return permissions.includes('it_assets')
  if (part === 'vehicles') return permissions.includes('vehicle_management')
  return permissions.includes('it_assets') || permissions.includes('vehicle_management')
}

const fmt = (n: number) => Math.round(n).toLocaleString('en-US')

export default function ReplacementSuggestionsPanel({ part, compact = false }: { part: 'it' | 'vehicles' | 'any'; compact?: boolean }) {
  const { employee, permissions } = useSession()
  const allowed = canSeeReplacements(employee?.role, permissions, part)
  const [data, setData] = useState<ReplacementSuggestions | null | undefined>(undefined)

  useEffect(() => {
    if (!allowed) return
    let alive = true
    api.getReplacementSuggestions()
      .then((r) => { if (alive) setData(r) })
      .catch(() => { if (alive) setData(null) })
    return () => { alive = false }
  }, [allowed])

  if (!allowed) return null
  if (data === undefined) return <div className="h-12 animate-pulse rounded-xl bg-slate-100" />
  if (data === null) return <p className="text-xs text-slate-400">ما وصلت اقتراحات الاستبدال.</p>

  const showIt = part !== 'vehicles' && data.itIncluded
  const showVeh = part !== 'it' && data.vehiclesIncluded
  const empty = (!showIt || data.it.length === 0) && (!showVeh || data.vehicles.length === 0)

  return (
    <div className="space-y-2">
      {empty && <p className="text-sm text-slate-500">ماكو شي يحتاج استبدال حسب القواعد ✅</p>}
      {showIt && data.it.length > 0 && (
        <ul className="space-y-1.5">
          {data.it.slice(0, compact ? 4 : 50).map((a) => (
            <li key={a.id} className="rounded-xl bg-amber-50 px-3 py-2 text-sm">
              <b className="text-amber-900">🖥️ {a.name}</b>
              <span className="mr-2 text-xs text-amber-700">{a.reason}</span>
            </li>
          ))}
        </ul>
      )}
      {showVeh && data.vehicles.length > 0 && (
        <ul className="space-y-1.5">
          {data.vehicles.slice(0, compact ? 4 : 50).map((v) => (
            <li key={v.id} className="rounded-xl bg-amber-50 px-3 py-2 text-sm">
              <b className="text-amber-900">🚗 {v.name}{v.plateNumber ? ` (${v.plateNumber})` : ''}</b>
              <span className="mr-2 text-xs text-amber-700">{v.reasons.join(' · ')}</span>
            </li>
          ))}
        </ul>
      )}
      {showVeh && (data.fleetMedian != null
        ? <p className="text-[11px] text-slate-400">وسيط كلفة الأسطول (١٢ شهر): {fmt(data.fleetMedian)} د.ع</p>
        : data.fleetNote && <p className="text-[11px] text-slate-400">{data.fleetNote}</p>)}
      <p className="text-[11px] text-slate-400">🧠 اقتراح ماتركس — القرار إلكم.</p>
    </div>
  )
}
