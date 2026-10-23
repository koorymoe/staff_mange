import { useEffect, useRef, useState } from 'react'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { api } from '../api'
import { TopBar } from '../components/ui'

// ═══ موقعي — (ع) 10-14 ═══
// الزبون يحدد موقعه بنفسه، ويصير هو الموقع المعتمد بالنظام لكل حجوزاته:
// الفني يوصل له بالضبط، والنظام يتعلّم المسافة والوقت لكل زبون.

const KARBALA: [number, number] = [32.616, 44.0249]
const pinIcon = () => L.divIcon({
  className: '',
  html: '<div style="width:18px;height:18px;border-radius:50% 50% 50% 0;transform:rotate(-45deg);background:#e11d48;border:3px solid #fff;box-shadow:0 2px 6px rgba(0,0,0,.4)"></div>',
  iconSize: [24, 24], iconAnchor: [12, 24],
})

export default function MyLocationPage() {
  const el = useRef<HTMLDivElement>(null)
  const mapRef = useRef<L.Map | null>(null)
  const markerRef = useRef<L.Marker | null>(null)
  const [pt, setPt] = useState<{ lat: number; lng: number } | null>(null)
  const [saved, setSaved] = useState<{ lat: number; lng: number } | null>(null)
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!el.current || mapRef.current) return
    const m = L.map(el.current, { center: KARBALA, zoom: 12, zoomControl: false })
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution: '&copy; OpenStreetMap', maxZoom: 19 }).addTo(m)
    m.on('click', (e: L.LeafletMouseEvent) => setPt({ lat: e.latlng.lat, lng: e.latlng.lng }))
    mapRef.current = m
    let alive = true
    api.myLocation().then((l) => {
      if (alive && l.lat != null && l.lng != null) { setSaved({ lat: l.lat, lng: l.lng }); setPt({ lat: l.lat, lng: l.lng }) }
    }).catch(() => {})
    return () => { alive = false; m.remove(); mapRef.current = null }
  }, [])

  useEffect(() => {
    const m = mapRef.current
    if (!m || !pt) return
    if (!markerRef.current) {
      markerRef.current = L.marker([pt.lat, pt.lng], { icon: pinIcon(), draggable: true }).addTo(m)
      markerRef.current.on('dragend', () => { const p = markerRef.current!.getLatLng(); setPt({ lat: p.lat, lng: p.lng }) })
    } else markerRef.current.setLatLng([pt.lat, pt.lng])
    m.setView([pt.lat, pt.lng], Math.max(m.getZoom(), 16))
  }, [pt])

  const locate = () => {
    if (!navigator.geolocation) { setMsg({ ok: false, text: 'جهازك ما يدعم تحديد الموقع' }); return }
    setBusy(true); setMsg(null)
    navigator.geolocation.getCurrentPosition(
      (p) => {
        setBusy(false)
        setPt({ lat: p.coords.latitude, lng: p.coords.longitude })
        const acc = Math.round(p.coords.accuracy)
        setMsg(acc > 300
          ? { ok: false, text: `الموقع تقريبي (±${acc} م) — اسحب الدبوس الأحمر لباب بيتك بالضبط` }
          : { ok: true, text: `تم تحديد موقعك (±${acc} م) — تأكد إن الدبوس على بيتك واضغط حفظ` })
      },
      (e) => { setBusy(false); setMsg({ ok: false, text: e.code === e.PERMISSION_DENIED ? 'اسمح للتطبيق ياخذ موقعك من إعدادات المتصفح' : 'تعذر تحديد موقعك — حدده بالضغط على الخريطة' }) },
      { enableHighAccuracy: true, timeout: 15000, maximumAge: 0 },
    )
  }

  const save = async () => {
    if (!pt) return
    setBusy(true)
    try { await api.setMyLocation(pt.lat, pt.lng); setSaved(pt); setMsg({ ok: true, text: '✓ انحفظ موقعك — الفني راح يوصلك عليه' }) }
    catch (e) { setMsg({ ok: false, text: e instanceof Error ? e.message : 'تعذر الحفظ' }) }
    finally { setBusy(false) }
  }

  return (
    <div className="page-in pb-4">
      <TopBar title="موقعي" />
      <div className="space-y-3 px-4 pt-1">
        <p className="text-[13px] text-mute">حدد موقع بيتك أو محلك مرة وحدة، ويصير هو المعتمد لكل طلباتك حتى يوصلك الفني بالضبط.</p>
        <div className="card overflow-hidden"><div ref={el} className="h-[380px] w-full" /></div>
        {msg && <p className={`rounded-xl px-3 py-2 text-[13px] font-bold ${msg.ok ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-800'}`}>{msg.text}</p>}
        <div className="grid grid-cols-2 gap-2">
          <button disabled={busy} onClick={locate} className="h-[50px] rounded-xl bg-white font-extrabold text-navy ring-1 ring-line disabled:opacity-50">📍 موقعي الحالي</button>
          <button disabled={busy || !pt || (saved?.lat === pt.lat && saved?.lng === pt.lng)} onClick={save} className="h-[50px] rounded-xl bg-navy-2 font-extrabold text-white disabled:opacity-50">حفظ الموقع</button>
        </div>
        {!pt && <p className="text-center text-[12px] text-mute">اضغط على الخريطة لتحديد الموقع</p>}
      </div>
    </div>
  )
}
