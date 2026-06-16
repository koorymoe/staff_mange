import { Fragment, useEffect, useState } from 'react'
import { api, type Booking } from '../api'
import { useSession } from '../session'

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
  const { employee } = useSession()
  const canSeeStats = employee?.role === 'ADMIN' || employee?.role === 'MONITOR'
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [expandedId, setExpandedId] = useState<string | null>(null)

  useEffect(() => {
    api
      .getBookings()
      .then(setBookings)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const isMonitor = employee?.role === 'MONITOR'
  const pendingReschedules = bookings.filter((b) => b.pendingScheduledAt)

  const handleApproveReschedule = async (b: Booking) => {
    const updated = await api.approveReschedule(b.id)
    setBookings((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
  }
  const handleRejectReschedule = async (b: Booking) => {
    const updated = await api.rejectReschedule(b.id)
    setBookings((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
  }

  const filtered = bookings.filter((b) => {
    const q = search.trim().toLowerCase()
    if (!q) return true
    return (
      b.code.toLowerCase().includes(q) ||
      b.customer.name.toLowerCase().includes(q) ||
      b.customer.code.toLowerCase().includes(q)
    )
  })

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

      {!loading && !error && isMonitor && pendingReschedules.length > 0 && (
        <div className="mt-6 overflow-hidden rounded-xl border border-amber-200 bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="bg-gradient-to-l from-amber-500 to-amber-700 px-4 py-3 font-bold text-white">
            ⏳ طلبات تعديل المواعيد بانتظار الموافقة
          </h3>
          <div className="flex flex-col divide-y divide-slate-100">
            {pendingReschedules.map((b) => (
              <div key={b.id} className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-sm">
                <span className="font-mono font-semibold text-brand-600">{b.code}</span>
                <span className="text-slate-600">{b.customer.name}</span>
                <span className="text-slate-500">
                  من: {b.scheduledAt ? new Date(b.scheduledAt).toLocaleString('ar-IQ') : '-'}
                </span>
                <span className="font-bold text-brand-800">
                  إلى: {new Date(b.pendingScheduledAt!).toLocaleString('ar-IQ')}
                </span>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleApproveReschedule(b)}
                    className="rounded-lg bg-emerald-600 px-3 py-1 text-xs font-medium text-white hover:bg-emerald-700"
                  >
                    موافقة
                  </button>
                  <button
                    onClick={() => handleRejectReschedule(b)}
                    className="rounded-lg bg-red-100 px-3 py-1 text-xs font-medium text-red-700 hover:bg-red-200"
                  >
                    رفض
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

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

      <div className="mt-6">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="بحث برقم الحجز، اسم الزبون، أو كود الزبون..."
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 sm:w-96"
        />
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">رقم الحجز</th>
                <th className="px-4 py-3 text-sm font-semibold">الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">كود الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">الخدمة</th>
                <th className="px-4 py-3 text-sm font-semibold">الموظف الذي سجل</th>
                <th className="px-4 py-3 text-sm font-semibold">السيارة</th>
                <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
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
                    <td className="px-4 py-3">{b.customer.name}</td>
                    <td className="px-4 py-3 font-mono text-sm text-slate-500">{b.customer.code}</td>
                    <td className="px-4 py-3 text-slate-600">{b.service?.name || '-'}</td>
                    <td className="px-4 py-3 text-slate-600">{b.transferEmployee?.name || '-'}</td>
                    <td className="px-4 py-3 text-slate-500">{b.assignedVehicle || '-'}</td>
                    <td className="px-4 py-3 text-slate-500">
                      {new Date(b.createdAt).toLocaleDateString('ar-IQ')}
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
                            <p className="mt-1 text-slate-700">{b.confirmedByName || '-'}</p>
                          </div>
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
      )}
    </div>
  )
}
