import { useEffect, useRef, useState, useCallback } from 'react'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

const KARBALA: [number, number] = [32.6160, 44.0249]

interface SearchResult {
  display_name: string
  lat: string
  lon: string
}

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

function createMyLocationIcon() {
  return L.divIcon({
    className: '',
    html: `<div style="
      width:14px;height:14px;border-radius:50%;
      background:#2c5aad;border:3px solid white;
      box-shadow:0 0 0 3px rgba(44,90,173,0.3), 0 2px 6px rgba(0,0,0,0.3);
    "></div>`,
    iconSize: [20, 20],
    iconAnchor: [10, 10],
  })
}

interface Props {
  value: { lat: number; lng: number } | null
  onChange: (point: { lat: number; lng: number } | null) => void
}

export default function LocationPicker({ value, onChange }: Props) {
  const mapRef = useRef<L.Map | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const markerRef = useRef<L.Marker | null>(null)
  const myMarkerRef = useRef<L.Marker | null>(null)

  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const [searchError, setSearchError] = useState('')

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return

    const map = L.map(containerRef.current, {
      center: value ? [value.lat, value.lng] : KARBALA,
      zoom: value ? 15 : 12,
    })

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap',
      maxZoom: 19,
    }).addTo(map)

    map.on('click', (e: L.LeafletMouseEvent) => {
      onChange({ lat: e.latlng.lat, lng: e.latlng.lng })
    })

    mapRef.current = map

    // مركز الخريطة أول ما تفتح على موقع الموظف الحالي، حتى ما يحتاج يدور يدوياً
    if (!value && navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          const m = mapRef.current
          if (!m) return
          const here: [number, number] = [pos.coords.latitude, pos.coords.longitude]
          m.setView(here, 13)
          myMarkerRef.current = L.marker(here, { icon: createMyLocationIcon(), interactive: false }).addTo(m)
        },
        () => {},
        { timeout: 5000 }
      )
    }

    return () => { map.remove(); mapRef.current = null }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const map = mapRef.current
    if (!map) return

    if (value) {
      if (myMarkerRef.current) { myMarkerRef.current.remove(); myMarkerRef.current = null }
      if (markerRef.current) {
        markerRef.current.setLatLng([value.lat, value.lng])
      } else {
        markerRef.current = L.marker([value.lat, value.lng], { icon: createPinIcon(), draggable: true }).addTo(map)
        markerRef.current.on('dragend', () => {
          const pos = markerRef.current!.getLatLng()
          onChange({ lat: pos.lat, lng: pos.lng })
        })
      }
      map.setView([value.lat, value.lng], Math.max(map.getZoom(), 15))
    } else if (markerRef.current) {
      markerRef.current.remove()
      markerRef.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value])

  const handleSearch = useCallback(async () => {
    if (!query.trim()) return
    setSearching(true)
    setSearchError('')
    try {
      const res = await fetch(
        `https://nominatim.openstreetmap.org/search?format=json&limit=5&countrycodes=iq&q=${encodeURIComponent(query.trim())}`
      )
      if (!res.ok) throw new Error('تعذر البحث')
      const data: SearchResult[] = await res.json()
      setResults(data)
      if (data.length === 0) setSearchError('لم يتم العثور على نتائج')
    } catch {
      setSearchError('تعذر البحث عن المنطقة، حدد الموقع يدوياً من الخريطة')
    } finally {
      setSearching(false)
    }
  }, [query])

  const pickResult = (r: SearchResult) => {
    const lat = parseFloat(r.lat)
    const lng = parseFloat(r.lon)
    onChange({ lat, lng })
    setResults([])
    setQuery(r.display_name)
  }

  const useMyLocation = () => {
    if (!navigator.geolocation) return
    navigator.geolocation.getCurrentPosition(
      (pos) => onChange({ lat: pos.coords.latitude, lng: pos.coords.longitude }),
      () => setSearchError('تعذر تحديد موقعك الحالي')
    )
  }

  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-2 sm:flex-row">
        <div className="relative flex-1">
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleSearch() } }}
            placeholder="دور عن اسم المنطقة (مثال: الكرادة، الحسينية...)"
            className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20"
          />
          {results.length > 0 && (
            <div className="absolute z-[1000] mt-1 w-full overflow-hidden rounded-xl border border-slate-200 bg-white shadow-lg">
              {results.map((r, i) => (
                <button
                  key={i}
                  type="button"
                  onClick={() => pickResult(r)}
                  className="block w-full border-b border-slate-100 px-4 py-2.5 text-right text-sm text-slate-700 last:border-b-0 hover:bg-slate-50"
                >
                  {r.display_name}
                </button>
              ))}
            </div>
          )}
        </div>
        <button
          type="button"
          onClick={handleSearch}
          disabled={searching}
          className="rounded-xl bg-brand-600 px-5 py-2.5 text-sm font-bold text-white hover:bg-brand-700 disabled:opacity-50"
        >
          {searching ? 'جارٍ البحث...' : 'بحث'}
        </button>
        <button
          type="button"
          onClick={useMyLocation}
          className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-medium text-slate-600 hover:border-brand-400"
        >
          📍 موقعي الحالي
        </button>
      </div>

      {searchError && <p className="text-xs text-red-600">{searchError}</p>}

      <div ref={containerRef} className="h-72 w-full overflow-hidden rounded-xl border border-slate-200" />

      <p className="text-xs text-slate-500">
        اضغط على الخريطة لتحديد الموقع، أو دور بالاسم واختر من النتائج. بعد التحديد تكدر تسحب العلامة الحمراء لتثبيت الموقع بدقة أكثر.
        {value && (
          <span className="mr-2 font-semibold text-brand-700">
            تم تحديد نقطة ✓ ({value.lat.toFixed(5)}, {value.lng.toFixed(5)})
          </span>
        )}
      </p>
    </div>
  )
}
