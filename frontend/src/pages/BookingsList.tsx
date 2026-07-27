import { Fragment, useEffect, useRef, useState } from 'react'
import { api, type Booking } from '../api'
import { useSession } from '../session'

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

function toDateKey(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
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
  const dateInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    api
      .getBookings()
      .then(setBookings)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const filtered = bookings
    .filter((b) => {
      if (selectedDate && toDateKey(relevantDate(b)) !== selectedDate) return false
      const q = search.trim().toLowerCase()
      if (!q) return true
      return (
        b.code.toLowerCase().includes(q) ||
        (b.customer?.name || '').toLowerCase().includes(q) ||
        (b.customer?.code || '').toLowerCase().includes(q)
      )
    })
    .sort((a, b) => (selectedDate ? 0 : new Date(relevantDate(a)).getTime() - new Date(relevantDate(b)).getTime()))

  // Service popularity stats
  const serviceCounts = new Map<string, number>()
  bookings.forEach((b) => {
    const name = b.service?.name || 'بدون خدمة محددة'
    serviceCounts.set(name, (serviceCounts.get(name) || 0) + 1)
  })
  const topServices = [...serviceCounts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 5)
  const maxCount = Math.max(1, ...topServices.map(([, c]) => c))

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">الحجوزات</h2>
      <p className="mt-1 text-slate-500">
        سجل كامل بجميع الحجوزات السابقة مع تفاصيلها، وأكثر الخدمات التي يطلبها الزبائن.
      </p>


      {!loading && !error && canSeeStats && topServices.length > 0 && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
            📊 أكثر الخدمات التي طلبها الزبائن
          </h3>
          <div className="flex flex-col gap-2 p-4">
            {topServices.map(([name, count]) => (
              <div key={name}>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-600">{name}</span>
                  <span className="font-bold text-brand-800">{count}</span>
                </div>
                <div className="mt-1 h-2 w-full rounded-full bg-slate-100">
                  <div
                    className="h-2 rounded-full bg-gradient-to-l from-brand-400 to-brand-700"
                    style={{ width: `${(count / maxCount) * 100}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="mt-6 flex flex-wrap items-center gap-3">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="بحث برقم الحجز، اسم الزبون، أو كود الزبون..."
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 sm:w-96"
        />

        <div className="flex items-center gap-1 rounded-lg border border-slate-300 bg-white px-1 py-1">
          <button
            onClick={() => setSelectedDate((d) => addDays(d || todayStr(), 1))}
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
              📅 {selectedDate ? formatDateArabic(selectedDate) : 'كل الحجوزات'}
            </button>
            <input
              ref={dateInputRef}
              type="date"
              value={selectedDate || ''}
              onChange={(e) => e.target.value && setSelectedDate(e.target.value)}
              className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
            />
          </div>
          <button
            onClick={() => setSelectedDate((d) => addDays(d || todayStr(), -1))}
            className="rounded-md px-2 py-1 text-slate-500 hover:bg-slate-100"
            title="اليوم السابق"
          >
            ◀
          </button>
        </div>

        {selectedDate !== todayStr() && (
          <button
            onClick={() => setSelectedDate(todayStr())}
            className="rounded-lg border border-brand-200 px-3 py-2 text-sm font-medium text-brand-700 hover:bg-brand-50"
          >
            اليوم
          </button>
        )}

        <button
          onClick={() => setSelectedDate((d) => (d === null ? todayStr() : null))}
          className={`rounded-lg border px-3 py-2 text-sm font-medium ${
            selectedDate === null
              ? 'border-brand-500 bg-brand-500 text-white'
              : 'border-slate-300 text-slate-600 hover:bg-slate-50'
          }`}
          title="كل الحجوزات القادمة (بدون فلتر تاريخ)، مرتبة حسب موعد التنفيذ"
        >
          📋 كل الحجوزات
        </button>
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
                    <td className="px-4 py-3 text-slate-600">{b.service?.name || '-'}</td>
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
                            <p className="text-slate-400">الكادر الذي تم تكليفه</p>
                            {b.assignments.length > 0 ? (
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
                            <p className="mt-1 text-slate-700">{b.assignedVehicle || '-'}</p>
                          </div>
                          <div>
                            <p className="text-slate-400">عنوان تنفيذ المهمة</p>
                            <p className="mt-1 text-slate-700">{b.address || '-'}</p>
                            {b.mapLocation && (
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
    </div>
  )
}
