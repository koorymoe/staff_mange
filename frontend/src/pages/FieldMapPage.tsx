import { useEffect, useRef, useState } from 'react'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet.heat'
import { api, type BookingTrack, type CompanyLocation, type FieldMap } from '../api'
import PageHeader from '../components/PageHeader'
import LocationPicker from '../components/LocationPicker'
import { useSession } from '../session'

// ═══ خريطة الميدان — قرار (ع) 10-14 ═══
// للمالك والمدير والمراقب: الشركة، حجوزات اليوم بمواقعها، آخر موقع لكل موظف،
// منين انطلق كل حجز، والخريطة الحرارية (وين تتجمع الحركة والتوقفات).
// واختيار حجز يطلع مساره وتوقفاته ويّا مقارنة ضغطات الموظف بحساب النظام.

const today = () => new Date().toLocaleDateString('en-CA')
const hm = (v?: string | null) => (v ? new Date(v).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }) : '—')
const ORIGIN: Record<string, string> = { COMPANY: '🏢 من الشركة', PREV_BOOKING: '🔁 من حجز سابق', OUTSIDE: '⚠️ من خارج الشركة', UNKNOWN: '❔ بلا موقع' }
const STAGE: Record<string, string> = { ACCEPTED: 'استلام', MATERIALS: 'المواد', DEPARTED: 'انطلاق', ARRIVED: 'وصول', STARTED: 'بدء', COMPLETED: 'إنجاز' }
const VERDICT: Record<string, string> = { OK: '', FAR: '📏 بعيد', BATCH: '📚 تسجيل جماعي', AUTO: '🤖 النظام' }

type HeatLayerFactory = (pts: [number, number, number][], o: Record<string, unknown>) => L.Layer
const heatLayer = (L as unknown as { heatLayer: HeatLayerFactory }).heatLayer

function diffMin(a?: string | null, b?: string | null) {
  if (!a || !b) return null
  return Math.round((new Date(a).getTime() - new Date(b).getTime()) / 60000)
}

export default function FieldMapPage() {
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'
  const [from, setFrom] = useState(today())
  const [to, setTo] = useState(today())
  const [mode, setMode] = useState<'map' | 'heat'>('map')
  const [data, setData] = useState<FieldMap | null>(null)
  const [err, setErr] = useState('')
  const [reload, setReload] = useState(0)
  const [sel, setSel] = useState<string | null>(null)
  const [trackState, setTrackState] = useState<{ id: string; t: BookingTrack } | null>(null)
  const track = sel && trackState?.id === sel ? trackState.t : null
  const [editCompany, setEditCompany] = useState(false)
  const mapEl = useRef<HTMLDivElement>(null)
  const mapRef = useRef<L.Map | null>(null)
  const layerRef = useRef<L.LayerGroup | null>(null)

  useEffect(() => {
    let alive = true
    api.getFieldMap(from, to).then((d) => { if (alive) { setData(d); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب') })
    return () => { alive = false }
  }, [from, to, reload])

  useEffect(() => {
    let alive = true
    if (!sel) return
    api.getBookingTrack(sel).then((t) => { if (alive) setTrackState({ id: sel, t }) }).catch(() => { /* يبقى بلا مسار */ })
    return () => { alive = false }
  }, [sel])

  useEffect(() => {
    if (!mapEl.current || mapRef.current) return
    const m = L.map(mapEl.current, { center: [32.616, 44.0249], zoom: 12 })
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution: '&copy; OpenStreetMap', maxZoom: 19 }).addTo(m)
    layerRef.current = L.layerGroup().addTo(m)
    mapRef.current = m
    return () => { m.remove(); mapRef.current = null }
  }, [])

  useEffect(() => {
    const m = mapRef.current, g = layerRef.current
    if (!m || !g || !data) return
    g.clearLayers()
    const bounds: L.LatLngExpression[] = []
    if (data.company) {
      const c: [number, number] = [data.company.lat, data.company.lng]
      L.circle(c, { radius: data.company.radiusM, color: '#0d3b8c', weight: 2, fillOpacity: 0.08 }).addTo(g)
      L.marker(c, { icon: L.divIcon({ className: '', html: '<div style="font-size:24px">🏢</div>', iconSize: [28, 28], iconAnchor: [14, 14] }) })
        .bindPopup('<div dir="rtl"><b>شركة الأماني</b></div>').addTo(g)
      bounds.push(c)
    }
    if (mode === 'heat') {
      if (data.heat.length && heatLayer) {
        heatLayer(data.heat.map(([a, b, w]) => [a, b, Math.min(1, 0.3 + w / 30)] as [number, number, number]), { radius: 22, blur: 18, maxZoom: 16 }).addTo(g)
        data.heat.forEach(([a, b]) => bounds.push([a, b]))
      }
    } else {
      for (const b of data.bookings) {
        if (b.lat == null || b.lng == null) continue
        const color = b.completedAt ? '#16a34a' : b.arrivedAt ? '#0284c7' : b.departedAt ? '#dc2626' : '#94a3b8'
        L.circleMarker([b.lat, b.lng], { radius: 9, color: '#fff', weight: 2, fillColor: color, fillOpacity: 1 })
          .bindPopup(`<div dir="rtl"><b>${b.code}</b> · ${b.customer ?? ''}<br/>${b.team ?? ''}<br/>انطلاق ${hm(b.departedAt)} ${b.departOrigin ? ORIGIN[b.departOrigin] : ''}<br/>وصول ${hm(b.arrivedAt)} · إنجاز ${hm(b.completedAt)}</div>`)
          .on('click', () => setSel(b.id)).addTo(g)
        bounds.push([b.lat, b.lng])
      }
      for (const e of data.employees) {
        L.marker([e.lat, e.lng], { icon: L.divIcon({ className: '', html: `<div style="background:#0d3b8c;color:#fff;border-radius:10px;padding:2px 6px;font-size:11px;font-weight:700;white-space:nowrap;border:2px solid #fff">🚗 ${e.name}</div>`, iconAnchor: [10, 10] }) })
          .bindPopup(`<div dir="rtl"><b>${e.name}</b><br/>آخر موقع ${hm(e.at)}</div>`).addTo(g)
        bounds.push([e.lat, e.lng])
      }
      if (track) {
        if (track.path.length > 1) L.polyline(track.path, { color: '#7c3aed', weight: 4, opacity: 0.8 }).addTo(g)
        for (const s of track.stops) {
          L.circleMarker([s.lat, s.lng], { radius: s.kind === 'ARRIVAL' ? 11 : 7, color: '#fff', weight: 2, fillColor: s.kind === 'ARRIVAL' ? '#0284c7' : '#f59e0b', fillOpacity: 1 })
            .bindPopup(`<div dir="rtl"><b>${s.kind === 'ARRIVAL' ? 'عند الزبون' : 'توقف بالطريق'}</b><br/>${hm(s.startAt)} ← ${hm(s.endAt)} (${s.minutes} دقيقة)</div>`).addTo(g)
        }
        track.path.forEach((p) => bounds.push(p))
      }
    }
    if (bounds.length) m.fitBounds(L.latLngBounds(bounds as L.LatLngTuple[]).pad(0.15), { maxZoom: 15 })
  }, [data, mode, track])

  const saveCompany = async (c: CompanyLocation) => {
    try { await api.setCompanyLocation(c); setEditCompany(false); setReload((n) => n + 1) } catch (e) { alert(e instanceof Error ? e.message : 'تعذر الحفظ') }
  }

  const t = track?.times
  const rows: [string, string | null | undefined, string | null | undefined][] = t ? [
    ['الانطلاق', t.departedAt, t.autoDepartedAt], ['الوصول', t.arrivedAt, t.autoArrivedAt], ['الإنجاز', t.completedAt, t.autoCompletedAt],
  ] : []
  const selBooking = data?.bookings.find((b) => b.id === sel)

  return (
    <div dir="rtl" className="space-y-4">
      <PageHeader title="🗺️ خريطة الميدان" subtitle="منين طلع كل فريق · وين كل حجز · المسار والتوقفات · ضغطات الموظف مقابل حساب النظام" />

      <div className="flex flex-wrap items-center gap-2 rounded-2xl bg-white p-3 shadow-sm">
        <label className="text-sm text-slate-500">من</label>
        <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
        <label className="text-sm text-slate-500">إلى</label>
        <input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
        <div className="mr-auto flex gap-1">
          <button onClick={() => setMode('map')} className={`rounded-lg px-3 py-1.5 text-sm font-bold ${mode === 'map' ? 'bg-brand-600 text-white' : 'bg-slate-100 text-slate-600'}`}>📍 خريطة المواقع</button>
          <button onClick={() => setMode('heat')} className={`rounded-lg px-3 py-1.5 text-sm font-bold ${mode === 'heat' ? 'bg-red-600 text-white' : 'bg-slate-100 text-slate-600'}`}>🔥 الخريطة الحرارية</button>
          <button onClick={() => setReload((n) => n + 1)} className="rounded-lg bg-slate-100 px-3 py-1.5 text-sm">↻</button>
          {isAdmin && <button onClick={() => setEditCompany((v) => !v)} className="rounded-lg border border-brand-300 px-3 py-1.5 text-sm font-bold text-brand-700">🏢 موقع الشركة</button>}
        </div>
      </div>

      {err && <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>}
      {data && !data.company && <p className="rounded-xl border border-amber-300 bg-amber-50 p-3 text-sm font-bold text-amber-800">⚠️ موقع الشركة ما محدد — بدونه النظام ما يعرف الانطلاق من الشركة. {isAdmin ? 'حدده من زر «🏢 موقع الشركة».' : 'اطلب من المدير يحدده.'}</p>}
      {editCompany && <CompanyEditor initial={data?.company ?? null} onSave={saveCompany} />}

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white xl:col-span-2">
          <div ref={mapEl} className="h-[560px] w-full" />
          <div className="flex flex-wrap gap-3 p-2 text-xs text-slate-500">
            <span>⚪ ما انطلق</span><span className="text-red-600">● بالطريق</span><span className="text-sky-600">● عند الزبون</span><span className="text-green-600">● خلص</span>
            <span className="text-violet-600">— المسار</span><span className="text-amber-600">● توقف بالطريق</span>
          </div>
        </div>

        <div className="space-y-3">
          {sel && selBooking && (
            <div className="rounded-2xl bg-white p-4 shadow-sm">
              <div className="flex items-center justify-between">
                <p className="font-bold text-slate-800">{selBooking.code} · {selBooking.customer}</p>
                <button onClick={() => setSel(null)} className="text-xs text-slate-400">✕</button>
              </div>
              {t?.departOrigin && <p className="mt-1 text-sm">{ORIGIN[t.departOrigin]}</p>}
              <table className="mt-2 w-full text-sm">
                <thead><tr className="text-xs text-slate-400"><th className="text-right">المرحلة</th><th>الموظف</th><th>النظام</th><th>الفرق</th></tr></thead>
                <tbody>
                  {rows.map(([label, manual, auto]) => {
                    const d = diffMin(manual, auto)
                    return (
                      <tr key={label} className="border-t border-slate-100">
                        <td className="py-1">{label}</td>
                        <td className="text-center">{hm(manual)}</td>
                        <td className="text-center">{hm(auto)}</td>
                        <td className={`text-center font-bold ${d == null ? 'text-slate-300' : Math.abs(d) > 20 ? 'text-red-600' : 'text-emerald-600'}`}>
                          {d == null ? (auto && !manual ? 'نسى يضغط' : '—') : `${d > 0 ? '+' : ''}${d} د`}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
              {track && track.stops.length > 0 && (
                <div className="mt-3">
                  <p className="text-xs font-bold text-slate-500">التوقفات</p>
                  {track.stops.map((s, i) => (
                    <p key={i} className="text-sm">{s.kind === 'ARRIVAL' ? '📍 عند الزبون' : '⏸️ بالطريق'} {hm(s.startAt)} ← {hm(s.endAt)} · <b>{s.minutes} د</b></p>
                  ))}
                </div>
              )}
              {track && track.proofs.length > 0 && (
                <div className="mt-3">
                  <p className="text-xs font-bold text-slate-500">الضغطات</p>
                  {track.proofs.map((p, i) => (
                    <p key={i} className="text-sm">{STAGE[p.stage] ?? p.stage} {hm(p.pressedAt)} {p.offline ? '📴' : ''} {p.distanceM != null ? `· ${Math.round(p.distanceM)} م` : ''} <span className="text-amber-700">{VERDICT[p.verdict] ?? p.verdict}</span></p>
                  ))}
                </div>
              )}
              {track && track.path.length === 0 && <p className="mt-2 text-xs text-slate-400">ماكو نقاط مسار لهذا الحجز.</p>}
            </div>
          )}
          <div className="rounded-2xl bg-white p-3 shadow-sm">
            <p className="mb-2 text-sm font-bold text-slate-700">الحجوزات ({data?.bookings.length ?? 0})</p>
            <div className="max-h-[440px] space-y-1 overflow-auto">
              {data?.bookings.map((b) => (
                <button key={b.id} onClick={() => setSel(b.id)} className={`block w-full rounded-lg px-2 py-1.5 text-right text-sm ${sel === b.id ? 'bg-brand-50 ring-1 ring-brand-300' : 'hover:bg-slate-50'}`}>
                  <span className="font-bold">{b.code}</span> · {b.customer} {b.lat == null && <span className="text-xs text-amber-600">(بلا موقع)</span>}
                  <span className="block text-xs text-slate-500">
                    {b.departedAt ? `انطلق ${hm(b.departedAt)} ${b.departOrigin ? ORIGIN[b.departOrigin] : ''}` : 'ما انطلق'}
                    {b.arrivedAt ? ` · وصل ${hm(b.arrivedAt)}` : ''}{b.autoArrivedAt ? ` (النظام ${hm(b.autoArrivedAt)})` : ''}
                  </span>
                </button>
              ))}
              {data && data.bookings.length === 0 && <p className="text-sm text-slate-400">ماكو حجوزات بهذي الفترة.</p>}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function CompanyEditor({ initial, onSave }: { initial: CompanyLocation | null; onSave: (c: CompanyLocation) => void }) {
  const [pt, setPt] = useState<{ lat: number; lng: number } | null>(initial ? { lat: initial.lat, lng: initial.lng } : null)
  const [radius, setRadius] = useState(initial?.radiusM ?? 300)
  return (
    <div className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
      <p className="font-bold text-slate-800">🏢 موقع شركة الأماني — كل انطلاق ينقاس عليه</p>
      <LocationPicker value={pt} onChange={setPt} />
      <div className="flex items-center gap-2">
        <label className="text-sm">نصف القطر (متر)</label>
        <input type="number" min={50} max={3000} value={radius} onChange={(e) => setRadius(Number(e.target.value))} className="w-28 rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
        <button disabled={!pt} onClick={() => pt && onSave({ lat: pt.lat, lng: pt.lng, radiusM: radius })} className="mr-auto rounded-lg bg-brand-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">حفظ</button>
      </div>
    </div>
  )
}
