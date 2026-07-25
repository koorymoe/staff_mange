import { useEffect, useState } from 'react'
import { api, type Booking, type Employee, type CartItem, type Product } from '../api'
import { useSession } from '../session'
import LocationPicker from '../components/LocationPicker'

// Convert an ISO date string to the local "YYYY-MM-DDTHH:mm" format expected by datetime-local inputs
const toLocalInput = (iso: string) => {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// ساعات العمل وطول كل موعد (بالساعات) - تستخدم لاقتراح أقرب موعد فاضي
const WORK_START_HOUR = 8
const LAST_SLOT_HOUR = 22
const SLOT_HOURS = 2


const techRoles: { key: 'TECH_1' | 'TECH_2' | 'TECH_3'; label: string }[] = [
  { key: 'TECH_1', label: 'الفني الأول' },
  { key: 'TECH_2', label: 'الفني الثاني' },
  { key: 'TECH_3', label: 'الفني الثالث' },
]

export default function Coordinator() {
  const { employee: currentUser } = useSession()
  const [bookings, setBookings] = useState<Booking[]>([])
  const [matches, setMatches] = useState<Record<string, Employee[]>>({})
  const [supervisors, setSupervisors] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // قيم التثبيت (التكلفة المقدرة + العنوان + الموعد) قبل الضغط على تثبيت
  const [priceDrafts, setPriceDrafts] = useState<Record<string, string>>({})
  const [addressDrafts, setAddressDrafts] = useState<Record<string, string>>({})
  const [scheduleDrafts, setScheduleDrafts] = useState<Record<string, string>>({})

  // سلة المنتجات
  const [cartItems, setCartItems] = useState<Record<string, CartItem[]>>({})
  const [cartOpen, setCartOpen] = useState<Record<string, boolean>>({})
  const [cartForm, setCartForm] = useState<Record<string, { productName: string; quantity: string; unitPrice: string; notes: string }>>({})
  const [products, setProducts] = useState<Product[]>([])
  const [scheduleMode, setScheduleMode] = useState<Record<string, 'slots' | 'manual'>>({})
  const [editingLocationId, setEditingLocationId] = useState<string | null>(null)

  const getAvailableSlots = (excludeId?: string) => {
    const takenKeys = new Set<string>()
    const toKey = (d: Date) => `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}-${d.getHours()}`
    bookings
      .filter(b => b.scheduledAt && (b.status === 'CONFIRMED' || b.status === 'PENDING' || b.status === 'IN_PROGRESS'))
      .forEach(b => takenKeys.add(toKey(new Date(b.scheduledAt!))))
    Object.entries(scheduleDrafts).forEach(([id, val]) => {
      if (id !== excludeId && val) takenKeys.add(toKey(new Date(val)))
    })

    const slots: { value: string; label: string }[] = []
    const slot = new Date()
    slot.setMinutes(0, 0, 0)
    if (slot.getHours() < WORK_START_HOUR) slot.setHours(WORK_START_HOUR)
    else if (slot.getHours() > LAST_SLOT_HOUR) { slot.setDate(slot.getDate() + 1); slot.setHours(WORK_START_HOUR) }
    else { slot.setHours(slot.getHours() + 1); if (slot.getHours() > LAST_SLOT_HOUR) { slot.setDate(slot.getDate() + 1); slot.setHours(WORK_START_HOUR) } }

    for (let i = 0; i < 200 && slots.length < 30; i++) {
      const key = toKey(slot)
      if (!takenKeys.has(key)) {
        const pad = (n: number) => String(n).padStart(2, '0')
        const value = `${slot.getFullYear()}-${pad(slot.getMonth() + 1)}-${pad(slot.getDate())}T${pad(slot.getHours())}:00`
        const label = slot.toLocaleString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'numeric', hour: '2-digit', minute: '2-digit' })
        slots.push({ value, label })
      }
      slot.setHours(slot.getHours() + SLOT_HOURS)
      if (slot.getHours() > LAST_SLOT_HOUR || slot.getHours() < WORK_START_HOUR) { slot.setDate(slot.getDate() + 1); slot.setHours(WORK_START_HOUR) }
    }
    return slots
  }

  const loadCart = async (bookingId: string) => {
    const items = await api.getCartItems(bookingId)
    setCartItems(prev => ({ ...prev, [bookingId]: items }))
  }

  const addCartItem = async (bookingId: string) => {
    const form = cartForm[bookingId]
    if (!form?.productName || !form?.quantity || !form?.unitPrice) return
    await api.addCartItem(bookingId, {
      productName: form.productName,
      quantity: Number(form.quantity),
      unitPrice: Number(form.unitPrice),
      notes: form.notes || undefined,
    })
    setCartForm(prev => ({ ...prev, [bookingId]: { productName: '', quantity: '', unitPrice: '', notes: '' } }))
    loadCart(bookingId)
  }

  const removeCartItem = async (bookingId: string, itemId: string) => {
    await api.deleteCartItem(itemId)
    loadCart(bookingId)
  }

  const toggleCart = (bookingId: string) => {
    const isOpen = !cartOpen[bookingId]
    setCartOpen(prev => ({ ...prev, [bookingId]: isOpen }))
    if (isOpen && !cartItems[bookingId]) loadCart(bookingId)
  }

  const load = () => {
    api
      .getBookings()
      .then((data) => {
        setBookings(data)
        data
          .filter((b) => b.status === 'CONFIRMED' && !b.transferToProjects && b.service)
          .forEach((b) => loadMatches(b))
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])
  useEffect(() => {
    api.getSupervisors().then(setSupervisors)
    api.getProducts().then(setProducts)
  }, [])

  const handleSupervisorChange = async (booking: Booking, employeeId: string) => {
    const updated = await api.assignSupervisor(booking.id, employeeId || null)
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const loadMatches = async (booking: Booking) => {
    if (!booking.service) return
    const employees = await api.matchEmployees(booking.service.id)
    setMatches((prev) => ({ ...prev, [booking.id]: employees }))
  }

  const handleConfirm = async (booking: Booking, transferToProjects: boolean) => {
    const priceValue = priceDrafts[booking.id]
    const addressValue = addressDrafts[booking.id]
    const scheduleValue = scheduleDrafts[booking.id]
    const updated = await api.confirmBooking(booking.id, {
      confirmedByName: currentUser?.name || 'الإداري',
      confirmedByEmployeeId: currentUser?.id,
      transferToProjects,
      quotedPrice: priceValue ? Number(priceValue) : undefined,
      address: addressValue || undefined,
      scheduledAt: scheduleValue || undefined,
    })
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    if (!transferToProjects) loadMatches(updated)
  }

  const handleScheduleChange = async (booking: Booking, value: string) => {
    if (!value) return
    const updated = await api.scheduleBooking(booking.id, value, currentUser?.id)
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const handleDetailsBlur = async (
    booking: Booking,
    field: 'quotedPrice' | 'address' | 'mapLocation',
    value: string,
  ) => {
    if (field === 'quotedPrice') {
      const num = value === '' ? null : Number(value)
      if (num === booking.quotedPrice) return
      const updated = await api.updateBookingDetails(booking.id, { quotedPrice: num })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } else if (field === 'mapLocation') {
      if (value === (booking.mapLocation || '')) return
      let lat: number | null = null, lng: number | null = null
      let m = value.match(/@(-?\d+\.?\d*),(-?\d+\.?\d*)/)
      if (!m) m = value.match(/[?&]q=(-?\d+\.?\d*),(-?\d+\.?\d*)/)
      if (!m) m = value.match(/place\/(-?\d+\.?\d*),(-?\d+\.?\d*)/)
      if (!m) m = value.match(/^(-?\d+\.?\d*)\s*,\s*(-?\d+\.?\d*)$/)
      if (m) { lat = parseFloat(m[1]); lng = parseFloat(m[2]) }
      const updated = await api.updateBookingDetails(booking.id, {
        mapLocation: value,
        mapLatitude: lat,
        mapLongitude: lng,
      })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } else {
      if (value === (booking.address || '')) return
      const updated = await api.updateBookingDetails(booking.id, { address: value })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    }
  }

  // تحديد الموقع مباشرة من الخريطة بدل الاعتماد على تحليل رابط ملصوق (اللي يفشل بكثير من روابط
  // خرائط كوكل المختصرة لأنها ما تحمل الإحداثيات أصلاً بالرابط نفسه)
  const handleMapPointChange = async (booking: Booking, point: { lat: number; lng: number } | null) => {
    if (!point) return
    const updated = await api.updateBookingDetails(booking.id, {
      mapLatitude: point.lat,
      mapLongitude: point.lng,
    })
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const handleAssign = async (
    booking: Booking,
    role: 'TECH_1' | 'TECH_2' | 'TECH_3',
    employeeId: string,
  ) => {
    if (!employeeId) return
    const updated = await api.assignTechnician(booking.id, { employeeId, role })
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const handleVehicleChange = async (booking: Booking, assignedVehicle: string) => {
    if (assignedVehicle === (booking.assignedVehicle || '')) return
    const updated = await api.updateBookingDetails(booking.id, { assignedVehicle })
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const pendingBookings = bookings.filter((b) => b.status === 'PENDING')
  const confirmedBookings = bookings.filter((b) => b.status === 'CONFIRMED')
  const upcomingAppointments = bookings
    .filter((b) => b.scheduledAt && (b.status === 'CONFIRMED' || b.status === 'PENDING') && new Date(b.scheduledAt) > new Date())
    .sort((a, b) => new Date(a.scheduledAt!).getTime() - new Date(b.scheduledAt!).getTime())

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">تنسيق الحجوزات (الإداري)</h2>
      <p className="mt-1 text-slate-500">
        ثبّت الحجز مع الزبون مع تحديد التكلفة والعنوان، ثم وجّهه لكادر الشد أو لإدارة المشاريع وحدد
        الفنيين المتاحين (اختياري الآن، يمكن تحديدهم لاحقاً).
      </p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>
      )}

      {!loading && !error && (
        <>
          {/* المواعيد المحجوزة مسبقاً - حتى لا يتم تحديد موعد متعارض */}
          {upcomingAppointments.length > 0 && (
            <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
                🗓️ المواعيد المحجوزة (تجنب تحديد نفس الوقت لزبون آخر)
              </h3>
              <div className="flex flex-col divide-y divide-slate-100">
                {upcomingAppointments.map((b) => (
                  <div key={b.id} className="flex flex-wrap items-center justify-between gap-2 px-4 py-2 text-sm">
                    <span className="font-mono font-semibold text-brand-600">{b.code}</span>
                    <span className="text-slate-600">{b.customer.name}</span>
                    <span className="text-slate-600">{b.service?.name || 'بدون خدمة محددة'}</span>
                    <span className="font-bold text-brand-800">
                      {new Date(b.scheduledAt!).toLocaleString('ar-IQ', {
                        weekday: 'long',
                        hour: '2-digit',
                        minute: '2-digit',
                        day: 'numeric',
                        month: 'numeric',
                      })}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* بانتظار التثبيت */}
          <h3 className="mt-6 mb-3 text-lg font-bold text-brand-800">
            بانتظار التثبيت ({pendingBookings.length})
          </h3>
          <div className="flex flex-col gap-4">
            {pendingBookings.map((booking) => (
              <div
                key={booking.id}
                className="rounded-xl border border-amber-200 bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <span className="font-mono text-sm font-semibold text-brand-600">
                      {booking.code}
                    </span>
                    <span className="mr-3 rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-700">
                      بانتظار التثبيت
                    </span>
                    {booking.priority === 'URGENT' && (
                      <span className="mr-2 rounded-full bg-red-100 px-3 py-1 text-xs font-bold text-red-700">
                        عاجل
                      </span>
                    )}
                  </div>
                  <div className="text-sm text-slate-500">
                    {booking.service?.name || 'بدون خدمة محددة'}
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-1 gap-2 text-sm sm:grid-cols-3">
                  <div>
                    <span className="text-slate-400">الزبون: </span>
                    <span className="font-medium text-brand-800">{booking.customer.name}</span>
                  </div>
                  <div>
                    <span className="text-slate-400">الهاتف: </span>
                    {booking.customer.phone}
                  </div>
                  <div>
                    <span className="text-slate-400">الموقع المسجل: </span>
                    {booking.customer.location || '-'}
                  </div>
                </div>
                {booking.notes && (
                  <p className="mt-2 text-sm text-slate-500">
                    <span className="text-slate-400">ملاحظات الزبون: </span>
                    {booking.notes}
                  </p>
                )}

                <div className="mt-4 rounded-lg bg-slate-50 p-4">
                  <h4 className="text-sm font-bold text-brand-800">
                    بعد الاتفاق مع الزبون، حدد التفاصيل قبل التثبيت
                  </h4>
                  <div className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-3">
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">
                        التكلفة المقدرة (اختياري)
                      </label>
                      <input
                        type="number"
                        placeholder="مثال: 150000"
                        value={priceDrafts[booking.id] || ''}
                        onChange={(e) =>
                          setPriceDrafts((prev) => ({ ...prev, [booking.id]: e.target.value }))
                        }
                        className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                      />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">
                        عنوان تنفيذ المهمة (اختياري)
                      </label>
                      <input
                        placeholder="عنوان مفصل لموقع تنفيذ المهمة"
                        value={addressDrafts[booking.id] || ''}
                        onChange={(e) =>
                          setAddressDrafts((prev) => ({ ...prev, [booking.id]: e.target.value }))
                        }
                        className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                      />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">
                        الموعد المحدد للزبون
                      </label>
                      <div className="flex gap-2">
                        {scheduleMode[booking.id] === 'manual' ? (
                          <input
                            type="datetime-local"
                            value={scheduleDrafts[booking.id] || ''}
                            onChange={(e) =>
                              setScheduleDrafts((prev) => ({ ...prev, [booking.id]: e.target.value }))
                            }
                            className="flex-1 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                          />
                        ) : (
                          <select
                            value={scheduleDrafts[booking.id] || ''}
                            onChange={(e) =>
                              setScheduleDrafts((prev) => ({ ...prev, [booking.id]: e.target.value }))
                            }
                            className="flex-1 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                          >
                            <option value="">-- اختر موعد --</option>
                            {getAvailableSlots(booking.id).map(s => (
                              <option key={s.value} value={s.value} disabled={s.label.includes('محجوز')}>
                                {s.label}
                              </option>
                            ))}
                          </select>
                        )}
                        <button
                          type="button"
                          onClick={() => setScheduleMode(prev => ({ ...prev, [booking.id]: prev[booking.id] === 'manual' ? 'slots' : 'manual' }))}
                          className="rounded-lg border border-slate-300 px-3 py-2 text-xs text-slate-600 transition-colors hover:bg-slate-50"
                        >
                          {scheduleMode[booking.id] === 'manual' ? 'المواعيد' : 'يدوي'}
                        </button>
                      </div>
                      <p className="mt-1 text-xs text-slate-400">
                        اختر من المواعيد المتاحة أو اكتب يدوياً
                      </p>
                    </div>
                  </div>
                </div>

                <div className="mt-4 flex flex-wrap gap-2">
                  <button
                    onClick={() => handleConfirm(booking, false)}
                    className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg"
                  >
                    تثبيت وترحيل لكادر الشد
                  </button>
                  <button
                    onClick={() => handleConfirm(booking, true)}
                    className="rounded-lg border border-brand-300 px-4 py-2 text-sm font-medium text-brand-700 transition-colors hover:bg-brand-50"
                  >
                    تثبيت وتحويل لإدارة المشاريع
                  </button>
                </div>
              </div>
            ))}
            {pendingBookings.length === 0 && (
              <p className="text-slate-400">لا توجد حجوزات بانتظار التثبيت.</p>
            )}
          </div>

          {/* تم تثبيتها */}
          <h3 className="mt-8 mb-3 text-lg font-bold text-brand-800">
            تم تثبيتها ({confirmedBookings.length})
          </h3>
          <div className="flex flex-col gap-4">
            {confirmedBookings.map((booking) => (
              <div
                key={booking.id}
                className="rounded-xl border border-emerald-200 bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <span className="font-mono text-sm font-semibold text-brand-600">
                      {booking.code}
                    </span>
                    <span className="mr-3 rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-700">
                      تم التثبيت
                    </span>
                    {booking.priority === 'URGENT' && (
                      <span className="mr-2 rounded-full bg-red-100 px-3 py-1 text-xs font-bold text-red-700">
                        عاجل
                      </span>
                    )}
                    {booking.status === 'CONFIRMED' && booking.materialsReadyAt && (
                      <span className="mr-2 rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-700">
                        ⏰ المواد جاهزة، بانتظار انطلاق الفريق ({booking.materialsReadyBy?.name || '-'})
                      </span>
                    )}
                  </div>
                  <div className="text-sm text-slate-500">
                    {booking.service?.name || 'بدون خدمة محددة'}
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-1 gap-2 text-sm sm:grid-cols-3">
                  <div>
                    <span className="text-slate-400">الزبون: </span>
                    <span className="font-medium text-brand-800">{booking.customer.name}</span>
                  </div>
                  <div>
                    <span className="text-slate-400">الهاتف: </span>
                    {booking.customer.phone}
                  </div>
                  <div>
                    <span className="text-slate-400">الموقع المسجل: </span>
                    {booking.customer.location || '-'}
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-1 gap-3 rounded-lg bg-slate-50 p-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      كلفة العمل التقديرية (بدون المواد)
                    </label>
                    <input
                      type="number"
                      placeholder="غير محددة"
                      defaultValue={booking.quotedPrice ?? ''}
                      onBlur={(e) => handleDetailsBlur(booking, 'quotedPrice', e.target.value)}
                      className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                    {(cartItems[booking.id]?.length ?? 0) > 0 && (
                      <p className="mt-1 text-xs text-slate-400">
                        + كلفة المواد من السلة: {(cartItems[booking.id] || []).reduce((sum, i) => sum + i.totalPrice, 0).toLocaleString()} د.ع (تلقائي)
                      </p>
                    )}
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      عنوان تنفيذ المهمة
                    </label>
                    <input
                      placeholder="غير محدد"
                      defaultValue={booking.address || ''}
                      onBlur={(e) => handleDetailsBlur(booking, 'address', e.target.value)}
                      className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                  <div className="sm:col-span-2">
                    <div className="mb-1 flex items-center justify-between">
                      <label className="block text-sm font-medium text-slate-600">موقع تنفيذ المهمة على الخريطة</label>
                      <button
                        type="button"
                        onClick={() => setEditingLocationId(editingLocationId === booking.id ? null : booking.id)}
                        className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100"
                      >
                        {editingLocationId === booking.id ? 'إخفاء الخريطة' : 'تعديل الموقع على الخريطة'}
                      </button>
                    </div>
                    {booking.mapLatitude != null && booking.mapLongitude != null ? (
                      <a
                        href={`https://www.openstreetmap.org/?mlat=${booking.mapLatitude}&mlon=${booking.mapLongitude}#map=17/${booking.mapLatitude}/${booking.mapLongitude}`}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-block text-xs text-brand-500 hover:underline"
                      >
                        فتح الموقع المحدد حالياً على الخريطة
                      </a>
                    ) : (
                      <p className="text-xs text-slate-400">ما تحدد موقع لهذا الحجز بعد</p>
                    )}
                    {editingLocationId === booking.id && (
                      <div className="mt-2">
                        <LocationPicker
                          value={booking.mapLatitude != null && booking.mapLongitude != null
                            ? { lat: booking.mapLatitude, lng: booking.mapLongitude }
                            : null}
                          onChange={(point) => handleMapPointChange(booking, point)}
                        />
                      </div>
                    )}
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      الموعد المحدد
                    </label>
                    <div className="flex gap-2">
                      {scheduleMode[`c_${booking.id}`] === 'manual' ? (
                        <input
                          type="datetime-local"
                          defaultValue={booking.scheduledAt ? toLocalInput(booking.scheduledAt) : ''}
                          onBlur={(e) => handleScheduleChange(booking, e.target.value)}
                          className="flex-1 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                        />
                      ) : (
                        <select
                          defaultValue=""
                          onChange={(e) => { if (e.target.value) handleScheduleChange(booking, e.target.value) }}
                          className="flex-1 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                        >
                          <option value="">
                            {booking.scheduledAt
                              ? `الحالي: ${new Date(booking.scheduledAt).toLocaleString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'numeric', hour: '2-digit', minute: '2-digit' })}`
                              : '-- اختر موعد --'}
                          </option>
                          {getAvailableSlots(booking.id).map(s => (
                            <option key={s.value} value={s.value} disabled={s.label.includes('محجوز')}>
                              {s.label}
                            </option>
                          ))}
                        </select>
                      )}
                      <button
                        type="button"
                        onClick={() => setScheduleMode(prev => ({ ...prev, [`c_${booking.id}`]: prev[`c_${booking.id}`] === 'manual' ? 'slots' : 'manual' }))}
                        className="rounded-lg border border-slate-300 px-3 py-2 text-xs text-slate-600 transition-colors hover:bg-slate-50"
                      >
                        {scheduleMode[`c_${booking.id}`] === 'manual' ? 'المواعيد' : 'يدوي'}
                      </button>
                    </div>
                    {booking.scheduleLogs?.length > 0 && (
                      <details className="mt-1">
                        <summary className="cursor-pointer text-xs text-slate-400 hover:text-slate-600">سجل التعديلات ({booking.scheduleLogs.length})</summary>
                        <div className="mt-1 max-h-32 space-y-1 overflow-y-auto rounded-lg bg-slate-50 p-2">
                          {booking.scheduleLogs.map(log => (
                            <p key={log.id} className="text-[11px] text-slate-500">
                              <span className="font-bold text-slate-700">{log.changedBy.name}</span>
                              {' غيّر من '}
                              <span className="text-red-500">{log.oldTime ? new Date(log.oldTime).toLocaleString('ar-IQ') : 'بدون موعد'}</span>
                              {' إلى '}
                              <span className="text-emerald-600">{new Date(log.newTime).toLocaleString('ar-IQ')}</span>
                              {' — '}
                              <span className="text-slate-400">{new Date(log.createdAt).toLocaleString('ar-IQ')}</span>
                            </p>
                          ))}
                        </div>
                      </details>
                    )}
                  </div>
                </div>

                {booking.transferToProjects ? (
                  <p className="mt-4 rounded-lg bg-brand-50 px-4 py-2 text-sm text-brand-700">
                    تم تحويل هذا الطلب إلى إدارة المشاريع.
                  </p>
                ) : (
                  <div className="mt-4">
                    <h4 className="text-sm font-bold text-brand-800">توجيه كادر الشد</h4>
                    <p className="mt-1 text-xs text-slate-400">
                      تحديد الكادر والسيارة اختياري - يمكن تثبيت الحجز وتحديدهم لاحقاً.
                    </p>

                    <div className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-2">
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          تيم ليدر (اختياري)
                        </label>
                        <select
                          value={booking.projectSupervisor?.id || ''}
                          onChange={(e) => handleSupervisorChange(booking, e.target.value)}
                          className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                        >
                          <option value="">-- بدون تيم ليدر --</option>
                          {supervisors.map((s) => (
                            <option key={s.id} value={s.id}>
                              {s.name}
                            </option>
                          ))}
                        </select>
                      </div>
                      {booking.expenseResponsible && (
                        <div className="flex items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 self-end">
                          <span className="text-xs text-emerald-700">المسؤول عن المصاريف:</span>
                          <span className="text-xs font-bold text-emerald-800">{booking.expenseResponsible.name}</span>
                        </div>
                      )}
                    </div>

                    {matches[booking.id] && (
                      <div className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-3">
                        {techRoles.map((tr) => {
                          const assigned = booking.assignments.find((a) => a.role === tr.key)
                          const candidates = matches[booking.id]
                          return (
                            <div key={tr.key}>
                              <label className="mb-1 block text-sm font-medium text-slate-600">
                                {tr.label} (اختياري)
                              </label>
                              <select
                                value={assigned?.employee.id || ''}
                                onChange={(e) => handleAssign(booking, tr.key, e.target.value)}
                                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                              >
                                <option value="">-- اختر فني --</option>
                                {candidates.length === 0 && (
                                  <option value="" disabled>
                                    لا يوجد موظف متاح
                                  </option>
                                )}
                                {candidates.map((c) => (
                                  <option key={c.id} value={c.id}>
                                    {c.hasRequiredSkill === false ? `⚠️ ${c.name} (لا يمتلك المهارة المطلوبة)` : `${c.name} (${c.position || 'فني'})`}
                                  </option>
                                ))}
                              </select>
                            </div>
                          )
                        })}
                      </div>
                    )}
                    <div className="mt-3 sm:w-1/3">
                      <label className="mb-1 block text-sm font-medium text-slate-600">
                        السيارة المخصصة للمهمة (اختياري)
                      </label>
                      <input
                        defaultValue={booking.assignedVehicle || ''}
                        onBlur={(e) => handleVehicleChange(booking, e.target.value)}
                        placeholder="مثال: تويوتا هايلوكس - أبيض - 12345"
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                      />
                    </div>
                    {booking.assignments.some((a) => {
                      const candidate = matches[booking.id]?.find((c) => c.id === a.employee.id)
                      return candidate && candidate.hasRequiredSkill === false
                    }) && (
                      <p className="mt-2 rounded-lg bg-amber-50 border border-amber-200 px-4 py-2 text-sm text-amber-700">
                        ⚠️ تحذير: الفني المختار لا يمتلك المهارة المطلوبة لهذه الخدمة
                      </p>
                    )}
                  </div>
                )}

                {/* سلة المنتجات */}
                <div className="mt-4 border-t border-slate-100 pt-4">
                  <button
                    onClick={() => toggleCart(booking.id)}
                    className="flex items-center gap-2 rounded-lg bg-gradient-to-l from-violet-500 to-violet-700 px-4 py-2 text-sm font-medium text-white shadow-md transition-all hover:shadow-lg"
                  >
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
                    سلة الزبون
                    {cartItems[booking.id]?.length ? (
                      <span className="rounded-full bg-white/30 px-2 py-0.5 text-xs">{cartItems[booking.id].length}</span>
                    ) : null}
                  </button>

                  {cartOpen[booking.id] && (
                    <div className="mt-3 rounded-xl border border-violet-100 bg-violet-50/50 p-4">
                      {/* قائمة العناصر */}
                      {(cartItems[booking.id] || []).length > 0 && (
                        <div className="mb-4 overflow-x-auto">
                          <table className="w-full text-sm">
                            <thead>
                              <tr className="border-b border-violet-200 text-right text-xs text-violet-600">
                                <th className="pb-2 pr-2">المنتج</th>
                                <th className="pb-2 pr-2">الكمية</th>
                                <th className="pb-2 pr-2">سعر الوحدة</th>
                                <th className="pb-2 pr-2">المجموع</th>
                                <th className="pb-2 pr-2">ملاحظات</th>
                                <th className="pb-2"></th>
                              </tr>
                            </thead>
                            <tbody>
                              {(cartItems[booking.id] || []).map(item => (
                                <tr key={item.id} className="border-b border-violet-100">
                                  <td className="py-2 pr-2 font-medium text-brand-800">{item.productName}</td>
                                  <td className="py-2 pr-2">{item.quantity}</td>
                                  <td className="py-2 pr-2">{item.unitPrice.toLocaleString()}</td>
                                  <td className="py-2 pr-2 font-bold text-violet-700">{item.totalPrice.toLocaleString()}</td>
                                  <td className="py-2 pr-2 text-xs text-slate-500">{item.notes || '-'}</td>
                                  <td className="py-2">
                                    <button
                                      onClick={() => removeCartItem(booking.id, item.id)}
                                      className="rounded p-1 text-red-400 transition-colors hover:bg-red-50 hover:text-red-600"
                                    >
                                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                                    </button>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                            <tfoot>
                              <tr className="font-bold text-violet-800">
                                <td colSpan={3} className="pt-3 pr-2">التكلفة الكلية</td>
                                <td className="pt-3 pr-2">{(cartItems[booking.id] || []).reduce((sum, i) => sum + i.totalPrice, 0).toLocaleString()}</td>
                                <td colSpan={2}></td>
                              </tr>
                            </tfoot>
                          </table>
                        </div>
                      )}

                      {/* إضافة عنصر جديد */}
                      <div className="grid grid-cols-2 gap-2 sm:grid-cols-5">
                        <div className="relative">
                          <select
                            value={cartForm[booking.id]?.productName || ''}
                            onChange={e => {
                              const prod = products.find(p => p.name === e.target.value)
                              setCartForm(prev => ({
                                ...prev,
                                [booking.id]: {
                                  ...prev[booking.id] || { productName: '', quantity: '', unitPrice: '', notes: '' },
                                  productName: e.target.value,
                                  unitPrice: prod?.defaultPrice ? String(prod.defaultPrice) : prev[booking.id]?.unitPrice || '',
                                },
                              }))
                            }}
                            className="w-full rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-500"
                          >
                            <option value="">-- اختر منتج --</option>
                            {products.map(p => (
                              <option key={p.id} value={p.name}>{p.name} {p.defaultPrice ? `(${p.defaultPrice.toLocaleString()})` : ''}</option>
                            ))}
                          </select>
                          {(() => {
                            const selProd = products.find(p => p.name === cartForm[booking.id]?.productName)
                            return selProd?.imageBase64 ? (
                              <img src={selProd.imageBase64} className="mt-1 h-12 w-12 rounded border border-violet-200 object-contain bg-white" />
                            ) : null
                          })()}
                        </div>
                        <input
                          type="number"
                          placeholder="الكمية"
                          value={cartForm[booking.id]?.quantity || ''}
                          onChange={e => setCartForm(prev => ({ ...prev, [booking.id]: { ...prev[booking.id] || { productName: '', quantity: '', unitPrice: '', notes: '' }, quantity: e.target.value } }))}
                          className="rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-500"
                        />
                        <input
                          type="number"
                          placeholder="سعر الوحدة"
                          value={cartForm[booking.id]?.unitPrice || ''}
                          onChange={e => setCartForm(prev => ({ ...prev, [booking.id]: { ...prev[booking.id] || { productName: '', quantity: '', unitPrice: '', notes: '' }, unitPrice: e.target.value } }))}
                          className="rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-500"
                        />
                        <input
                          placeholder="ملاحظات (اختياري)"
                          value={cartForm[booking.id]?.notes || ''}
                          onChange={e => setCartForm(prev => ({ ...prev, [booking.id]: { ...prev[booking.id] || { productName: '', quantity: '', unitPrice: '', notes: '' }, notes: e.target.value } }))}
                          className="rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-500"
                        />
                        <button
                          onClick={() => addCartItem(booking.id)}
                          className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-violet-700"
                        >
                          إضافة
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ))}
            {confirmedBookings.length === 0 && (
              <p className="text-slate-400">لا توجد حجوزات مثبتة بعد.</p>
            )}
          </div>
        </>
      )}
    </div>
  )
}
