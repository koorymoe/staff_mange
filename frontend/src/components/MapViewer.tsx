import { useEffect, useRef } from 'react'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

function createPinIcon() {
  return L.divIcon({
    className: '',
    html: `<div style="
      width:16px;height:16px;border-radius:50% 50% 50% 0;
      transform:rotate(-45deg);
      background:#e11d48;border:3px solid white;
      box-shadow:0 2px 6px rgba(0,0,0,0.4);
    "></div>`,
    iconSize: [22, 22],
    iconAnchor: [11, 22],
  })
}

interface Props {
  lat: number
  lng: number
  height?: number
}

// عرض نقطة موقع بس (بدون تعديل) — تُستخدم لعرض موقع حجز/مهمة داخل النظام
// نفسه بدل ما نفتحه برابط خارجي بتاب جديد.
export default function MapViewer({ lat, lng, height = 260 }: Props) {
  const mapRef = useRef<L.Map | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!containerRef.current) return
    const map = L.map(containerRef.current, { center: [lat, lng], zoom: 15 })
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap',
      maxZoom: 19,
    }).addTo(map)
    L.marker([lat, lng], { icon: createPinIcon() }).addTo(map)
    mapRef.current = map
    return () => { map.remove(); mapRef.current = null }
  }, [lat, lng])

  return <div ref={containerRef} className="w-full rounded-xl border border-slate-200" style={{ height }} />
}
