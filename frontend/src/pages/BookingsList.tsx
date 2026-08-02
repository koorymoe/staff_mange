import { Fragment, useEffect, useRef, useState } from 'react'
import { api, type Booking, type Employee, type Vehicle } from '../api'
import { useSession } from '../session'
import MapViewer from '../components/MapViewer'

// أسماء كل خدمات الحجز (الزبون ممكن يطلب أكثر من منظومة بنفس الحجز)
function serviceNames(b: { service?: { name: string } | null; services?: { name: string }[] }): string {
  if (b.services && b.services.length > 0) return b.services.map((s) => s.name).join(' + ')
  return b.service?.name || 'بدون خدمة محددة'
}

const techRoles: { key: 'TECH_1' | 'TECH_2' | 'TECH_3'; label: string }[] = [
  { key: 'TECH_1', label: 'الفني الأول' },
  { key: 'TECH_2', label: 'الفني الثاني' },
  { key: 'TECH_3', label: 'الفني الثالث' },
]

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function addDays(dateStr: string, delta: number) {
  const d = new Date(`${dateStr}T00:00:00`)
  d.setDate(d.getDate() + delta)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function formatDateArabic(dateStr: string) {
  return new Date(`${dateStr}T00:00:00`).toLocaleDateString('ar-IQ', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
}

// موعد الحجز الفعلي (الموعد المحدد للزبون) هو المعيار الصح للفلترة بالتاريخ —
// لو ما محدد موعد بعد (لسا ما تنسّق)، نرجع لتاريخ التسجيل حتى الحجز يبقى قابل
// للعثور عليه بدل ما يضيع من كل الفلاتر.
function relevantDate(b: Booking): string {
  return b.scheduledAt || b.createdAt
}

const statusLabels: Record<string, string> = {
  PENDING: 'بانتظار التثبيت',
  CONFIRMED: 'مثبت',
  IN_PROGRESS: 'جاري التنفيذ',
  COMPLETED: 'منجز',
  CANCELLED: 'ملغى',
}

const statusColors: Record<string, string> = {
  PENDING: 'bg-amber-100 text-amber-700',
  CONFIRMED: 'bg-blue-100 text-blue-700',
  IN_PROGRESS: 'bg-orange-100 text-orange-700',
  COMPLETED: 'bg-emerald-100 text-emerald-700',
  CANCELLED: 'bg-red-100 text-red-700',
}

const techRoleLabels: Record<string, string> = {
  TECH_1: 'الفني الأول',
  TECH_2: 'الفني الثاني',
  TECH_3: 'الفني الثالث',
}

export default function BookingsList() {
  const { employee, permissions } = useSession()
  const canSeeStats = employee?.role === 'ADMIN' || employee?.role === 'MONITOR' || permissions.includes('monitoring')
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [selectedDate, setSelectedDate] = useState<string | null>(todayStr())
  const [selectedMonth, setSelectedMonth] = useState<string | null>(null)
  const dateInputRef = useRef<HTMLInputElement>(null)
  const monthInputRef = useRef<HTMLInputElement>(null)
  const isAdmin = employee?.role === 'ADMIN'
  const [technicians, setTechnicians] = useState<Employee[]>([])
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [mapBooking, setMapBooking] = useState<Booking | null>(null)
  // الشخصيات المهمة: أي موظف يقدر يعلّم زبون بضغطة زر. نجيب المعرّفات بس
  // (بدون تفاصيل) حتى نعرف أي زر يكون مضغوط.
  const [vipIds, setVipIds] = useState<string[]>([])
  const [vipBusy, setVipBusy] = useState<string | null>(null)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [assigning, setAssigning] = useState(false)

  useEffect(() => {
    if (isAdmin) {
      api.getEmployees().then((all) => setTechnicians(all.filter((e) => e.role === 'TECHNICIAN')))
      // المركبات بيانات مساعدة هنا — مو كل من يشوف الحجوزات عنده صلاحية
      // المركبات، فالرفض ما يجوز يكسر الصفحة (شوف Coordinator.tsx)
      api.getVehicles().then(setVehicles).catch(() => setVehicles([]))
    }
  }, [isAdmin])

  useEffect(() => { api.getVipCustomerIds().then(setVipIds).catch(() => {}) }, [])

  // تعليم/إزالة "شخصية مهمة" — التعليم متاح لأي موظف، أما إزالة التعليم
  // فمحصورة بمدير النظام (السيرفر يفرضها كمان).
  const toggleVip = async (booking: Booking) => {
    const customerId = booking.customer?.id
    if (!customerId) return
    const already = vipIds.includes(customerId)
    if (already && !isAdmin) {
      alert('إزالة التعليم متاحة لمدير النظام فقط')
      return
    }
    setVipBusy(customerId)
    try {
      if (already) {
        await api.unmarkVipCustomer(customerId)
        setVipIds((prev) => prev.filter((id) => id !== customerId))
      } else {
        await api.markVipCustomer({
          customerId,
          bookingId: booking.id,
          requestSummary: booking.service?.name || undefined,
        })
        setVipIds((prev) => [...prev, customerId])
      }
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تنفيذ العملية')
    } finally {
      setVipBusy(null)
    }
  }

  const handleVehicleChange = async (booking: Booking, assignedVehicle: string) => {
    setAssigning(true)
    try {
      const updated = await api.updateBookingDetails(booking.id, { assignedVehicle })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تعديل السيارة')
    } finally {
      setAssigning(false)
    }
  }

  const handleReassign = async (booking: Booking, role: 'TECH_1' | 'TECH_2' | 'TECH_3', employeeId: string) => {
    if (!employeeId) return
    setAssigning(true)
    try {
      const updated = await api.assignTechnician(booking.id, { employeeId, role })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تعديل التكليف')
    } finally {
      setAssigning(false)
    }
  }

  useEffect(() => {
    // نطلب من السيرفر يوم واحد بس (لو محدد) بدل ما نجيب كل أرشيف الحجوزات التاريخي
    // ونفلتره بالواجهة — نفس المشكلة اللي صلحناها بصفحة تنسيق الحجوزات، وهذي الصفحة
    // كانت أسوأ لأنها ما تفلتر بالحالة إطلاقاً (كل الحجوزات من كل الأزمنة).
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoading(true)
    // فلتر الشهر يحتاج كل الحجوزات (نفلترها بالواجهة)، أما فلتر يوم واحد فنطلبه
    // من السيرفر مباشرة (أخف على قاعدة البيانات) — الاثنين ما ينفعون بنفس الوقت.
    api
      .getBookings(selectedMonth ? undefined : selectedDate ? { date: selectedDate } : undefined)
      .then(setBookings)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [selectedDate, selectedMonth])

  const filtered = bookings
    .filter((b) => {
      const q = search.trim().toLowerCase()
      if (q && !(
        b.code.toLowerCase().includes(q) ||
        (b.customer?.name || '').toLowerCase().includes(q) ||
        (b.customer?.code || '').toLowerCase().includes(q) ||
        (b.customer?.phone || '').includes(search.trim())
      )) return false
      if (selectedMonth && !relevantDate(b).startsWith(selectedMonth)) return false
      return true
    })
    // الأحدث للأقدم دايماً — هذا الترتيب الافتراضي المطلوب.
    .sort((a, b) => new Date(relevantDate(b)).getTime() - new Date(relevantDate(a)).getTime())

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">الحجوزات</h2>
      <p className="mt-1 text-slate-500">
        سجل كامل بجميع الحجوزات السابقة مع تفاصيلها.
        {canSeeStats && (
          <> إحصائية "أكثر الخدمات طلباً" لكل الخدمات صارت بصفحة <a href="/stats" className="text-brand-600 hover:underline">إحصائيات الموظفين</a>.</>
        )}
      </p>

      <div className="mt-6 flex flex-wrap items-center gap-3">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="بحث برقم الحجز، اسم الزبون، كود الزبون، أو رقم الهاتف..."
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 sm:w-96"
        />

        <div className="flex items-center gap-1 rounded-lg border border-slate-300 bg-white px-1 py-1">
          <button
            onClick={() => { setSelectedMonth(null); setSelectedDate((d) => addDays(d || todayStr(), 1)) }}
            className="rounded-md px-2 py-1 text-slate-500 hover:bg-slate-100"
            title="اليوم التالي"
          >
            ▶
          </button>
          <div className="relative">
            <button
              onClick={() => dateInputRef.current?.showPicker?.()}
              className="min-w-[180px] rounded-md px-2 py-1 text-sm font-medium text-brand-800 hover:bg-slate-100"
            >
              📅 {selectedMonth ? 'فلتر شهر مفعّل' : selectedDate ? formatDateArabic(selectedDate) : 'كل الحجوزات'}
            </button>
            <input
              ref={dateInputRef}
              type="date"
              value={selectedDate || ''}
              onChange={(e) => { if (e.target.value) { setSelectedMonth(null); setSelectedDate(e.target.value) } }}
              className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
            />
          </div>
          <button
            onClick={() => { setSelectedMonth(null); setSelectedDate((d) => addDays(d || todayStr(), -1)) }}
            className="rounded-md px-2 py-1 text-slate-500 hover:bg-slate-100"
            title="اليوم السابق"
          >
            ◀
          </button>
        </div>

        {selectedDate !== todayStr() && !selectedMonth && (
          <button
            onClick={() => setSelectedDate(todayStr())}
            className="rounded-lg border border-brand-200 px-3 py-2 text-sm font-medium text-brand-700 hover:bg-brand-50"
          >
            اليوم
          </button>
        )}

        <button
          onClick={() => { setSelectedMonth(null); setSelectedDate((d) => (d === null ? todayStr() : null)) }}
          className={`rounded-lg border px-3 py-2 text-sm font-medium ${
            selectedDate === null && !selectedMonth
              ? 'border-brand-500 bg-brand-500 text-white'
              : 'border-slate-300 text-slate-600 hover:bg-slate-50'
          }`}
          title="كل الحجوزات (بدون فلتر تاريخ)، مرتبة من الأحدث للأقدم"
        >
          📋 كل الحجوزات
        </button>

        {/* فلتر بحث متقدم: اختيار شهر كامل بدل يوم وحد */}
        <div className="relative">
          <button
            onClick={() => monthInputRef.current?.showPicker?.()}
            className={`rounded-lg border px-3 py-2 text-sm font-medium ${
              selectedMonth
                ? 'border-brand-500 bg-brand-500 text-white'
                : 'border-slate-300 text-slate-600 hover:bg-slate-50'
            }`}
          >
            🗓️ {selectedMonth
              ? new Date(`${selectedMonth}-01T00:00:00`).toLocaleDateString('ar-IQ', { year: 'numeric', month: 'long' })
              : 'فلتر متقدم: اختر شهر'}
          </button>
          <input
            ref={monthInputRef}
            type="month"
            value={selectedMonth || ''}
            onChange={(e) => { if (e.target.value) { setSelectedDate(null); setSelectedMonth(e.target.value) } }}
            className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
          />
        </div>
        {selectedMonth && (
          <button
            onClick={() => { setSelectedMonth(null); setSelectedDate(todayStr()) }}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50"
          >
            إلغاء فلتر الشهر
          </button>
        )}
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="overflow-x-auto">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">رقم الحجز</th>
                <th className="px-4 py-3 text-sm font-semibold">الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">كود الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">الخدمة</th>
                <th className="px-4 py-3 text-sm font-semibold">الموظف الذي سجل</th>
                <th className="px-4 py-3 text-sm font-semibold">السيارة</th>
                <th className="px-4 py-3 text-sm font-semibold">موعد التنفيذ</th>
                <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                <th className="px-4 py-3 text-sm font-semibold"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filtered.map((b) => (
                <Fragment key={b.id}>
                  <tr>
                    <td className="px-4 py-3 font-mono text-sm font-semibold text-brand-600">
                      {b.code}
                    </td>
                    <td className="px-4 py-3">{b.customer?.name || 'زبون غير معروف'}</td>
                    <td className="px-4 py-3 font-mono text-sm text-slate-500">{b.customer?.code || '-'}</td>
                    <td className="px-4 py-3 text-slate-600">{serviceNames(b)}</td>
                    <td className="px-4 py-3 text-slate-600">{b.transferEmployee?.name || '-'}</td>
                    <td className="px-4 py-3 text-slate-500">{b.assignedVehicle || '-'}</td>
                    <td className="px-4 py-3 text-slate-500">
                      {b.scheduledAt
                        ? new Date(b.scheduledAt).toLocaleString('ar-IQ', { day: 'numeric', month: 'numeric', hour: '2-digit', minute: '2-digit' })
                        : <span className="text-amber-600">لم يُنسَّق بعد ({new Date(b.createdAt).toLocaleDateString('ar-IQ')})</span>}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-2 py-1 text-xs font-bold ${statusColors[b.status]}`}>
                        {statusLabels[b.status] || b.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => setExpandedId(expandedId === b.id ? null : b.id)}
                        className="rounded-lg border border-brand-200 px-3 py-1 text-xs font-medium text-brand-700 transition-colors hover:bg-brand-50"
                      >
                        {expandedId === b.id ? 'إخفاء' : 'التفاصيل'}
                      </button>
                    </td>
                  </tr>
                  {expandedId === b.id && (
                    <tr>
                      <td colSpan={9} className="bg-slate-50 px-4 py-4">
                        <div className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
                          <div>
                            <p className="text-slate-400">اسم الزبون</p>
                            <p className="mt-1 font-bold text-slate-700">{b.customer?.name || 'زبون غير معروف'}</p>
                            {b.customer?.id && (
                              <button
                                onClick={() => toggleVip(b)}
                                disabled={vipBusy === b.customer?.id}
                                className={`mt-1 rounded-full px-3 py-1 text-xs font-bold transition-colors disabled:opacity-50 ${
                                  vipIds.includes(b.customer.id)
                                    ? 'bg-amber-400 text-amber-950'
                                    : 'bg-slate-100 text-slate-500 hover:bg-amber-100 hover:text-amber-700'
                                }`}
                              >
                                {vipIds.includes(b.customer.id) ? '⭐ شخصية مهمة' : '☆ تعليم كشخصية مهمة'}
                              </button>
                            )}
                          </div>
                          <div>
                            <p className="text-slate-400">رقم هاتف الزبون</p>
                            <p className="mt-1 font-bold text-brand-700" dir="ltr">{b.customer?.phone || '-'}</p>
                          </div>
                          <div>
                            <div className="flex items-center justify-between">
                              <p className="text-slate-400">الكادر الذي تم تكليفه</p>
                              {isAdmin && (
                                <button
                                  type="button"
                                  onClick={() => setEditingId(editingId === b.id ? null : b.id)}
                                  className="rounded-md bg-brand-50 px-2 py-0.5 text-xs font-bold text-brand-700 hover:bg-brand-100"
                                >
                                  {editingId === b.id ? 'إغلاق التعديل' : 'تعديل'}
                                </button>
                              )}
                            </div>
                            {editingId === b.id ? (
                              <div className="mt-2 flex flex-col gap-2">
                                {techRoles.map((tr) => {
                                  const current = b.assignments.find((a) => a.role === tr.key)
                                  return (
                                    <div key={tr.key}>
                                      <label className="mb-0.5 block text-xs text-slate-400">{tr.label}</label>
                                      <select
                                        value={current?.employee.id || ''}
                                        disabled={assigning}
                                        onChange={(e) => handleReassign(b, tr.key, e.target.value)}
                                        className="w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm outline-none focus:border-brand-500"
                                      >
                                        <option value="">-- اختر --</option>
                                        {technicians.map((t) => (
                                          <option key={t.id} value={t.id}>{t.name}{t.isLeader ? ' (ليدر)' : ''}</option>
                                        ))}
                                      </select>
                                    </div>
                                  )
                                })}
                              </div>
                            ) : b.assignments.length > 0 ? (
                              <ul className="mt-1 list-inside list-disc text-slate-700">
                                {b.assignments.map((a) => (
                                  <li key={a.id}>
                                    {a.employee.name}
                                    <span className="text-slate-400"> ({techRoleLabels[a.role] || a.role})</span>
                                  </li>
                                ))}
                              </ul>
                            ) : (
                              <p className="mt-1 text-slate-400">لم يتم تكليف أحد بعد</p>
                            )}
                          </div>
                          <div>
                            <p className="text-slate-400">السيارة المخصصة</p>
                            {editingId === b.id ? (
                              <select
                                value={b.assignedVehicle || ''}
                                disabled={assigning}
                                onChange={(e) => handleVehicleChange(b, e.target.value)}
                                className="mt-1 w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm outline-none focus:border-brand-500"
                              >
                                <option value="">-- اختر سيارة --</option>
                                {vehicles.map((v) => (
                                  <option key={v.id} value={`${v.name} - ${v.plateNumber}`}>
                                    {v.name} ({v.plateNumber}){v.color ? ` - ${v.color}` : ''}
                                  </option>
                                ))}
                              </select>
                            ) : (
                              <p className="mt-1 text-slate-700">{b.assignedVehicle || '-'}</p>
                            )}
                          </div>
                          <div>
                            <p className="text-slate-400">عنوان تنفيذ المهمة</p>
                            <p className="mt-1 text-slate-700">{b.address || '-'}</p>
                            {b.mapLatitude != null && b.mapLongitude != null ? (
                              <button
                                onClick={() => setMapBooking(b)}
                                className="text-xs text-brand-500 hover:underline"
                              >
                                عرض الموقع على الخريطة 🗺️
                              </button>
                            ) : b.mapLocation && (
                              <a href={b.mapLocation} target="_blank" rel="noreferrer" className="text-xs text-brand-500 hover:underline">
                                فتح على الخريطة
                              </a>
                            )}
                          </div>
                          <div>
                            <p className="text-slate-400">التكلفة المقدرة</p>
                            <p className="mt-1 text-slate-700">
                              {b.quotedPrice != null ? b.quotedPrice.toLocaleString() : '-'}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">الدفعة المقدمة</p>
                            <p className="mt-1 text-slate-700">
                              {b.advancePaid != null ? b.advancePaid.toLocaleString() : '-'}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">وقت تسجيل الحجز</p>
                            <p className="mt-1 text-slate-700">
                              {new Date(b.createdAt).toLocaleString('ar-IQ')}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">الموعد المحدد للزبون</p>
                            <p className="mt-1 text-slate-700">
                              {b.scheduledAt ? new Date(b.scheduledAt).toLocaleString('ar-IQ') : 'غير محدد'}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">وقت إنجاز الحجز</p>
                            <p className="mt-1 text-slate-700">
                              {b.completedAt ? new Date(b.completedAt).toLocaleString('ar-IQ') : 'لم يُنجز بعد'}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">المبلغ المستلم</p>
                            <p className="mt-1 text-slate-700">
                              {b.amountCollected != null ? b.amountCollected.toLocaleString() : '-'}
                              {b.amountCollected != null && (
                                <span className="mr-2 text-xs text-slate-400">
                                  ({b.amountVerified ? 'مدقق' : 'بانتظار التدقيق'})
                                </span>
                              )}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">ملاحظات الإنجاز</p>
                            <p className="mt-1 text-slate-700">{b.completionNotes || '-'}</p>
                          </div>
                          <div>
                            <p className="text-slate-400">ملاحظات الحجز</p>
                            <p className="mt-1 text-slate-700">{b.notes || '-'}</p>
                          </div>
                          <div>
                            <p className="text-slate-400">من أكد الحجز</p>
                            <p className="mt-1 text-slate-700">{b.confirmedByEmployee?.name || b.confirmedByName || '-'}</p>
                          </div>
                          <div>
                            <p className="text-slate-400">من عدّل الحجز</p>
                            <p className="mt-1 text-slate-700">
                              {b.lastEditedBy?.name || '-'}
                              {b.lastEditedBy && b.lastEditedAt && (
                                <span className="text-xs text-slate-400"> ({new Date(b.lastEditedAt).toLocaleString('ar-IQ')})</span>
                              )}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">موظف المبيعات (مصدر الزبون)</p>
                            <p className="mt-1 text-slate-700">{b.transferEmployee?.name || '-'}</p>
                          </div>
                          {b.adminNotes && (
                            <div className="col-span-full">
                              <p className="text-slate-400">ملاحظات الإدارة</p>
                              <p className="mt-1 whitespace-pre-line rounded-lg bg-amber-50 border border-amber-200 p-2 text-sm text-amber-800">{b.adminNotes}</p>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={9} className="px-4 py-6 text-center text-slate-400">
                    لا توجد حجوزات
                  </td>
                </tr>
              )}
            </tbody>
          </table>
          </div>
        </div>
      )}

      {mapBooking && mapBooking.mapLatitude != null && mapBooking.mapLongitude != null && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={() => setMapBooking(null)}>
          <div className="w-full max-w-2xl rounded-2xl bg-white p-4 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="mb-3 flex items-center justify-between">
              <h3 className="font-bold text-brand-900">موقع الحجز {mapBooking.code}</h3>
              <button onClick={() => setMapBooking(null)} className="rounded-lg px-3 py-1 text-sm text-slate-500 hover:bg-slate-100">✕ إغلاق</button>
            </div>
            <MapViewer lat={mapBooking.mapLatitude} lng={mapBooking.mapLongitude} height={380} />
          </div>
        </div>
      )}
    </div>
  )
}
