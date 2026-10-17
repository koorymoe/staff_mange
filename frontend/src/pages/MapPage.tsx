import { useState, useEffect, useRef, useCallback } from 'react'
import { useSession } from '../session'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'
import { api } from '../api'
import type { Booking } from '../api'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import BookingCodeChip from '../components/BookingCodeChip'

const KARBALA: [number, number] = [32.6160, 44.0249]

const MARKER_COLORS: Record<string, string> = {
  PENDING: '#f59e0b',
  CONFIRMED: '#3b82f6',
  IN_PROGRESS: '#8b5cf6',
  COMPLETED: '#10b981',
  CANCELLED: '#ef4444',
  // اتصلنا بالزبون وما رد — الحجز محفوظ بس مو بطابور الشغل
  WAITING: '#64748b',
}

const STATUS_LABELS: Record<string, string> = {
  PENDING: 'بانتظار التأكيد',
  CONFIRMED: 'مؤكد',
  IN_PROGRESS: 'جاري التنفيذ',
  COMPLETED: 'مكتمل',
  CANCELLED: 'ملغي',
  WAITING: 'في الانتظار',
}

function createMarkerIcon(color: string, isSelected = false) {
  const size = isSelected ? 16 : 12
  return L.divIcon({
    className: '',
    html: `<div style="
      width:${size}px;height:${size}px;border-radius:50%;
      background:${color};border:3px solid white;
      box-shadow:0 2px 6px rgba(0,0,0,0.4);
      ${isSelected ? 'transform:scale(1.3);' : ''}
    "></div>`,
    iconSize: [size + 6, size + 6],
    iconAnchor: [(size + 6) / 2, (size + 6) / 2],
  })
}

function createMyLocationIcon() {
  return L.divIcon({
    className: '',
    html: `<div style="
      width:18px;height:18px;border-radius:50%;
      background:#2c5aad;border:4px solid white;
      box-shadow:0 0 0 3px rgba(44,90,173,0.3), 0 2px 8px rgba(0,0,0,0.3);
      animation: pulse 2s infinite;
    "></div>
    <style>@keyframes pulse { 0%,100%{box-shadow:0 0 0 3px rgba(44,90,173,0.3)} 50%{box-shadow:0 0 0 8px rgba(44,90,173,0.1)} }</style>`,
    iconSize: [26, 26],
    iconAnchor: [13, 13],
  })
}

export default function MapPage() {
  // كل حفظ بهاي الشاشة يمر من هنا — الفشل ينعرض بدل ما ينبلع
  const guard = useSaveGuard()
  const { employee } = useSession()
  const [bookings, setBookings] = useState<Booking[]>([])
  const [selectedBooking, setSelectedBooking] = useState<Booking | null>(null)
  const [statusFilter, setStatusFilter] = useState<string>('active')
  const [myLocation, setMyLocation] = useState<[number, number] | null>(null)
  const [routeInfo, setRouteInfo] = useState<{ distance: string; duration: string } | null>(null)
  const [loading, setLoading] = useState(true)
  const [locatingMe, setLocatingMe] = useState(false)
  const [showLinkInput, setShowLinkInput] = useState<string | null>(null)
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [linkValue, setLinkValue] = useState('')

  const mapRef = useRef<L.Map | null>(null)
  const mapContainerRef = useRef<HTMLDivElement>(null)
  const markersRef = useRef<L.LayerGroup | null>(null)
  const myMarkerRef = useRef<L.Marker | null>(null)
  const routeLayerRef = useRef<L.Polyline | null>(null)

  useEffect(() => {
    // نجيب الي يطلبه الفلتر بس. جان يسحب أرشيف الشركة كله بكل فتحة
    // للخريطة — والفلتر الافتراضي «النشطة» أصلاً، يعني ٩٥٪ من الي
    // انسحب ما ينعرض.
    //
    // لاحظ: ما نقدر نقصّه على النشطة ونخلص، لأن بالفلتر خيارات «الكل»
    // و«مكتملة» — ولو جبنا النشطة بس جان المستخدم يختار «مكتملة»
    // ويشوف خريطة فارغة ويظن ماكو حجوزات منجزة.
    const wanted: Parameters<typeof api.getBookings>[0] =
      statusFilter === 'active'
        ? { status: ['PENDING', 'CONFIRMED', 'IN_PROGRESS'] }
        : statusFilter === 'all'
          ? { limit: 500 }
          : { status: statusFilter as Booking['status'], limit: 500 }
    api.getBookings(wanted).then(b => {
      // الفني يشوف بس الحجوزات المسندة له تحديداً من قبل الإداري/المنسق — مو كل حجوزات الشركة
      const visible = employee?.role === 'TECHNICIAN'
        ? b.filter((booking) => booking.assignments?.some((a) => a.employee.id === employee.id))
        : b
      setBookings(visible)
      setLoading(false)
    }).catch(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [employee?.id, statusFilter])

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) return

    const map = L.map(mapContainerRef.current, {
      center: KARBALA,
      zoom: 12,
      zoomControl: false,
    })

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap',
      maxZoom: 19,
    }).addTo(map)

    L.control.zoom({ position: 'topright' }).addTo(map)

    markersRef.current = L.layerGroup().addTo(map)
    mapRef.current = map

    return () => { map.remove(); mapRef.current = null }
  }, [])

  // خرائط Leaflet لازم تتحدّث حجمها يدوياً كل ما يتغير حجم/ظهور الحاوية
  // الي بيها (فتح/سكر قائمة الحجوزات بالموبايل، أو تغيير حجم الشاشة)
  useEffect(() => {
    const timer = setTimeout(() => mapRef.current?.invalidateSize(), 320)
    return () => clearTimeout(timer)
  }, [sidebarOpen])

  useEffect(() => {
    const onResize = () => mapRef.current?.invalidateSize()
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  const filteredBookings = bookings.filter(b => {
    if (!b.mapLatitude || !b.mapLongitude) return false
    if (statusFilter === 'active') return ['PENDING', 'CONFIRMED', 'IN_PROGRESS'].includes(b.status)
    if (statusFilter === 'all') return true
    return b.status === statusFilter
  })

  const allWithLocation = bookings.filter(b => b.mapLatitude && b.mapLongitude)

  useEffect(() => {
    if (!markersRef.current || !mapRef.current) return
    markersRef.current.clearLayers()

    filteredBookings.forEach(booking => {
      const color = MARKER_COLORS[booking.status] || '#6b7280'
      const isSelected = selectedBooking?.id === booking.id
      const marker = L.marker(
        [booking.mapLatitude!, booking.mapLongitude!],
        { icon: createMarkerIcon(color, isSelected) }
      )

      marker.bindPopup(`
        <div dir="rtl" style="font-family:inherit;min-width:200px">
          <strong style="color:${color}">${booking.code}</strong> - ${STATUS_LABELS[booking.status]}<br/>
          <b>${booking.customer?.name}</b><br/>
          ${booking.customer?.phone}<br/>
          ${booking.address || ''}<br/>
          ${booking.service ? `<small>${booking.service.name}</small>` : ''}
          ${booking.scheduledAt ? `<br/><small>📅 ${new Date(booking.scheduledAt).toLocaleDateString('ar-IQ')}</small>` : ''}
        </div>
      `)

      marker.on('click', () => setSelectedBooking(booking))
      marker.addTo(markersRef.current!)
    })
  }, [filteredBookings, selectedBooking])

  const locateMe = useCallback(() => {
    // متصفحات الموبايل ما تسمح بتحديد الموقع الجغرافي إلا على اتصال آمن
    // (HTTPS) — النظام حالياً شغال على HTTP عادي، فهذا السبب الرئيسي يخلي
    // الزر "ميتفاعل" بصمت بدون أي رسالة توضح ليش. لازم دومين + شهادة SSL
    // حتى يشتغل تحديد الموقع.
    if (!window.isSecureContext) {
      alert('تحديد الموقع الجغرافي يحتاج اتصال آمن (HTTPS) — هذا النظام لسه شغال بدون HTTPS. لازم نربط دومين ونضيف شهادة أمان حتى تشتغل هذي الميزة.')
      return
    }
    if (!navigator.geolocation) {
      alert('المتصفح ما يدعم تحديد الموقع الجغرافي')
      return
    }
    setLocatingMe(true)
    navigator.geolocation.getCurrentPosition(
      pos => {
        const coords: [number, number] = [pos.coords.latitude, pos.coords.longitude]
        setMyLocation(coords)
        setLocatingMe(false)

        if (mapRef.current) {
          if (myMarkerRef.current) myMarkerRef.current.remove()
          myMarkerRef.current = L.marker(coords, { icon: createMyLocationIcon() })
            .addTo(mapRef.current)
            .bindPopup('<div dir="rtl"><b>موقعي الحالي</b></div>')
          mapRef.current.setView(coords, 14)
        }
      },
      (err) => {
        setLocatingMe(false)
        const reasons: Record<number, string> = {
          1: 'رفضت السماح بالوصول لموقعك — فعّل صلاحية الموقع للمتصفح من إعدادات الجهاز',
          2: 'تعذر تحديد موقعك حالياً — تأكد خدمة تحديد الموقع مفعّلة بجهازك',
          3: 'انتهت مهلة تحديد الموقع — جرب مرة ثانية',
        }
        alert(reasons[err.code] || 'تعذر تحديد موقعك')
      },
      { enableHighAccuracy: true, timeout: 10000 }
    )
  }, [])

  const showRoute = useCallback(async (booking: Booking) => {
    if (!myLocation || !booking.mapLatitude || !booking.mapLongitude || !mapRef.current) return

    if (routeLayerRef.current) {
      routeLayerRef.current.remove()
      routeLayerRef.current = null
    }
    setRouteInfo(null)

    try {
      const [fromLng, fromLat] = [myLocation[1], myLocation[0]]
      const [toLng, toLat] = [booking.mapLongitude!, booking.mapLatitude!]
      const res = await fetch(
        `https://router.project-osrm.org/route/v1/driving/${fromLng},${fromLat};${toLng},${toLat}?overview=full&geometries=geojson`
      )
      const data = await res.json()

      if (data.routes?.[0]) {
        const route = data.routes[0]
        const coords = route.geometry.coordinates.map((c: number[]) => [c[1], c[0]] as [number, number])

        routeLayerRef.current = L.polyline(coords, {
          color: 'var(--t-info)',
          weight: 5,
          opacity: 0.8,
        }).addTo(mapRef.current!)

        mapRef.current!.fitBounds(routeLayerRef.current.getBounds(), { padding: [50, 50] })

        const distKm = (route.distance / 1000).toFixed(1)
        const durMin = Math.ceil(route.duration / 60)
        setRouteInfo({
          distance: `${distKm} كم`,
          duration: durMin >= 60 ? `${Math.floor(durMin / 60)} ساعة ${durMin % 60} دقيقة` : `${durMin} دقيقة`,
        })
      }
    } catch { /* OSRM might be unavailable */ }
  }, [myLocation])

  const parseLocationLink = (link: string): [number, number] | null => {
    // Google Maps: https://maps.google.com/?q=32.6,44.0 or @32.6,44.0 or /place/32.6,44.0
    let match = link.match(/@(-?\d+\.?\d*),(-?\d+\.?\d*)/)
    if (match) return [parseFloat(match[1]), parseFloat(match[2])]

    match = link.match(/[?&]q=(-?\d+\.?\d*),(-?\d+\.?\d*)/)
    if (match) return [parseFloat(match[1]), parseFloat(match[2])]

    match = link.match(/place\/(-?\d+\.?\d*),(-?\d+\.?\d*)/)
    if (match) return [parseFloat(match[1]), parseFloat(match[2])]

    // maps.app.goo.gl short links - can't resolve without server
    // Direct coordinates: 32.6,44.0
    match = link.match(/^(-?\d+\.?\d*)\s*,\s*(-?\d+\.?\d*)$/)
    if (match) return [parseFloat(match[1]), parseFloat(match[2])]

    return null
  }

  const saveLocation = async (bookingId: string) => {
    const coords = parseLocationLink(linkValue.trim())
    if (!coords) return

    const updated = await guard.run('حفظ موقع الحجز', () => api.updateBookingDetails(bookingId, {
      mapLatitude: coords[0],
      mapLongitude: coords[1],
      mapLocation: linkValue.trim(),
    }))
    if (!updated) return
    setBookings(prev => prev.map(b => b.id === updated.id ? updated : b))
    setShowLinkInput(null)
    setLinkValue('')

    if (mapRef.current) {
      mapRef.current.setView(coords, 15)
    }
  }

  const isCoordinator = employee?.role === 'ADMIN' || employee?.role === 'HR_COORDINATOR'
  const bookingsWithoutLocation = bookings.filter(b =>
    !b.mapLatitude && !b.mapLongitude &&
    ['PENDING', 'CONFIRMED', 'IN_PROGRESS'].includes(b.status)
  )

  return (
    <>
      <SaveError message={guard.error} onClose={guard.clear} />
    <div className="-m-3 flex h-[calc(100%+1.5rem)] flex-col sm:-m-5 sm:h-[calc(100%+2.5rem)] lg:-m-8 lg:h-[calc(100%+4rem)]">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200 bg-white px-3 py-2.5 sm:px-4 sm:py-3">
        <div className="flex items-center gap-2 sm:gap-3">
          {/* زر إظهار/إخفاء قائمة الحجوزات — يطلع بس بالموبايل */}
          <button
            onClick={() => setSidebarOpen(o => !o)}
            className="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 text-slate-600 lg:hidden"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="3" y1="6" x2="21" y2="6" /><line x1="3" y1="12" x2="15" y2="12" /><line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          </button>
          <h2 className="text-lg font-bold text-brand-900 sm:text-xl">خريطة المواقع</h2>
          <span className="rounded-full bg-brand-100 px-2.5 py-0.5 text-xs font-medium text-brand-700">
            {allWithLocation.length} موقع
          </span>
        </div>

        <div className="flex items-center gap-2">
          {/* Filter */}
          <select
            value={statusFilter}
            onChange={e => setStatusFilter(e.target.value)}
            className="rounded-lg border border-slate-300 px-2 py-1.5 text-xs sm:px-3 sm:text-sm"
          >
            <option value="active">النشطة</option>
            <option value="all">الكل</option>
            <option value="PENDING">بانتظار التأكيد</option>
            <option value="CONFIRMED">مؤكدة</option>
            <option value="IN_PROGRESS">جاري التنفيذ</option>
            <option value="COMPLETED">مكتملة</option>
          </select>

          {/* Locate me */}
          <button
            onClick={locateMe}
            disabled={locatingMe}
            className="flex items-center gap-1.5 rounded-lg bg-brand-600 px-2.5 py-1.5 text-xs font-medium text-white hover:bg-brand-700 disabled:opacity-50 sm:px-3 sm:text-sm"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="3" /><path d="M12 2v4M12 18v4M2 12h4M18 12h4" />
            </svg>
            <span className="hidden sm:inline">{locatingMe ? 'جاري التحديد...' : 'موقعي'}</span>
          </button>
        </div>
      </div>

      <div className="relative flex flex-1 overflow-hidden">
        {/* خلفية شفافة تسكر قائمة الحجوزات بالموبايل لما تكون مفتوحة */}
        {sidebarOpen && (
          <div onClick={() => setSidebarOpen(false)} className="absolute inset-0 z-30 bg-black/40 lg:hidden" />
        )}

        {/* Sidebar — بالموبايل: لوحة منزلقة فوق الخريطة. بسطح المكتب: عمود ثابت بجنب الخريطة */}
        <div
          className={`absolute inset-y-0 right-0 z-40 w-[85%] max-w-sm transform overflow-y-auto border-l border-slate-200 bg-white transition-transform duration-300 ease-in-out sm:w-80 lg:static lg:z-auto lg:w-80 lg:translate-x-0 lg:transform-none ${
            sidebarOpen ? 'translate-x-0' : 'translate-x-full'
          }`}
        >
          {/* Route info */}
          {routeInfo && selectedBooking && (
            <div className="border-b border-blue-200 bg-blue-50 p-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-bold text-blue-800">المسار إلى <BookingCodeChip code={selectedBooking.code} /></span>
                <button onClick={() => { routeLayerRef.current?.remove(); setRouteInfo(null) }} className="text-blue-500 text-xs">إغلاق</button>
              </div>
              <div className="mt-2 flex gap-4">
                <div className="text-center">
                  <div className="text-lg font-bold text-blue-900">{routeInfo.distance}</div>
                  <div className="text-xs text-blue-600">المسافة</div>
                </div>
                <div className="text-center">
                  <div className="text-lg font-bold text-blue-900">{routeInfo.duration}</div>
                  <div className="text-xs text-blue-600">الوقت المتوقع</div>
                </div>
              </div>
            </div>
          )}

          {/* Bookings without location */}
          {isCoordinator && bookingsWithoutLocation.length > 0 && (
            <div className="border-b border-amber-200 bg-amber-50 p-3">
              <h3 className="text-sm font-bold text-amber-800 mb-2">
                بدون موقع ({bookingsWithoutLocation.length})
              </h3>
              <div className="space-y-2 max-h-40 overflow-y-auto">
                {bookingsWithoutLocation.slice(0, 5).map(b => (
                  <div key={b.id} className="flex items-center justify-between bg-white rounded px-2 py-1.5 text-xs">
                    <div>
                      <span className="font-bold text-amber-700"><BookingCodeChip code={b.code} /></span>
                      <span className="text-slate-600 mr-1">{b.customer?.name}</span>
                    </div>
                    <button
                      onClick={() => { setShowLinkInput(b.id); setLinkValue('') }}
                      className="text-brand-600 font-medium hover:underline"
                    >
                      + إضافة
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Location link input modal */}
          {showLinkInput && (
            <div className="border-b border-brand-200 bg-brand-50 p-3">
              <h3 className="text-sm font-bold text-brand-800 mb-2">إضافة رابط الموقع</h3>
              <p className="text-xs text-brand-600 mb-2">الصق رابط Google Maps أو أدخل الإحداثيات (مثال: 32.616,44.025)</p>
              <input
                value={linkValue}
                onChange={e => setLinkValue(e.target.value)}
                placeholder="الصق رابط الموقع هنا..."
                className="w-full rounded-lg border border-brand-300 px-3 py-2 text-sm mb-2"
                dir="ltr"
              />
              {linkValue && !parseLocationLink(linkValue.trim()) && (
                <p className="text-xs text-red-500 mb-2">تعذر استخراج الإحداثيات من الرابط</p>
              )}
              <div className="flex gap-2">
                <button
                  onClick={() => saveLocation(showLinkInput)}
                  disabled={!linkValue || !parseLocationLink(linkValue.trim())}
                  className="flex-1 rounded-lg bg-brand-600 px-3 py-1.5 text-sm text-white hover:bg-brand-700 disabled:opacity-50"
                >
                  حفظ الموقع
                </button>
                <button
                  onClick={() => setShowLinkInput(null)}
                  className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm hover:bg-slate-50"
                >
                  إلغاء
                </button>
              </div>
            </div>
          )}

          {/* Booking list */}
          <div className="p-2">
            <h3 className="px-2 py-1 text-xs font-medium text-slate-500">
              الحجوزات على الخريطة ({filteredBookings.length})
            </h3>
            {filteredBookings.length === 0 && !loading && (
              <p className="px-2 py-4 text-center text-sm text-slate-400">لا توجد حجوزات بمواقع</p>
            )}
            {filteredBookings.map(booking => (
              <button
                key={booking.id}
                onClick={() => {
                  setSelectedBooking(booking)
                  if (mapRef.current && booking.mapLatitude && booking.mapLongitude) {
                    mapRef.current.setView([booking.mapLatitude, booking.mapLongitude], 15)
                  }
                }}
                className={`w-full text-right rounded-lg p-2.5 mb-1 transition ${
                  selectedBooking?.id === booking.id
                    ? 'bg-brand-50 border border-brand-300'
                    : 'hover:bg-slate-50 border border-transparent'
                }`}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="font-bold text-sm" style={{ color: MARKER_COLORS[booking.status] }}>
                    <BookingCodeChip code={booking.code} />
                  </span>
                  <span className="text-[10px] rounded-full px-2 py-0.5" style={{
                    background: MARKER_COLORS[booking.status] + '20',
                    color: MARKER_COLORS[booking.status],
                  }}>
                    {STATUS_LABELS[booking.status]}
                  </span>
                </div>
                <div className="text-sm font-medium text-slate-800">{booking.customer?.name}</div>
                <div className="text-xs text-slate-500">{booking.address || booking.customer?.phone}</div>
                {booking.service && (
                  <div className="text-xs text-slate-400 mt-0.5">{booking.service.name}</div>
                )}
                {booking.scheduledAt && (
                  <div className="text-xs text-slate-400 mt-0.5">
                    {new Date(booking.scheduledAt).toLocaleDateString('ar-IQ')} {new Date(booking.scheduledAt).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })}
                  </div>
                )}

                {/* Route button */}
                {myLocation && booking.mapLatitude && booking.mapLongitude && (
                  <button
                    onClick={e => { e.stopPropagation(); showRoute(booking) }}
                    className="mt-1.5 w-full flex items-center justify-center gap-1 rounded bg-brand-600 px-2 py-1 text-xs text-white hover:bg-brand-700"
                  >
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                      <path d="M3 11l19-9-9 19-2-8-8-2z" />
                    </svg>
                    عرض الطريق
                  </button>
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Map */}
        {/* isolate تحصر طبقات Leaflet الداخلية (z-index عالي جداً افتراضياً بمكتبة
            الخرائط) جوة هذا العنصر بس، حتى ما تطلع فوق القائمة الجانبية الرئيسية
            ولا فوق قائمة الحجوزات المنزلقة بالموبايل. */}
        <div className="isolate relative flex-1">
          <div ref={mapContainerRef} className="absolute inset-0" />

          {/* Legend */}
          <div className="absolute bottom-4 left-4 z-[1000] rounded-lg bg-white/95 p-3 shadow-lg backdrop-blur">
            <div className="text-xs font-medium text-slate-600 mb-1.5">دليل الألوان</div>
            <div className="space-y-1">
              {Object.entries(MARKER_COLORS).filter(([k]) => k !== 'CANCELLED').map(([status, color]) => (
                <div key={status} className="flex items-center gap-2 text-xs">
                  <div className="w-3 h-3 rounded-full" style={{ background: color }} />
                  <span>{STATUS_LABELS[status]}</span>
                </div>
              ))}
              <div className="flex items-center gap-2 text-xs border-t border-slate-200 pt-1 mt-1">
                <div className="w-3 h-3 rounded-full bg-brand-600 border-2 border-white shadow" />
                <span>موقعي</span>
              </div>
            </div>
          </div>

          {/* My location hint */}
          {!myLocation && (
            <div className="absolute top-4 left-4 z-[1000] rounded-lg bg-white/95 px-3 py-2 shadow text-xs text-slate-600">
              اضغط "موقعي" لعرض موقعك الحالي والطرق
            </div>
          )}
        </div>
      </div>
    </div>
    </>
  )
}
