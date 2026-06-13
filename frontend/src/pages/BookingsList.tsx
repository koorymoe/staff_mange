import { useEffect, useState } from 'react'
import { api, type Booking } from '../api'

const statusLabels: Record<string, string> = {
  PENDING: 'بانتظار التثبيت',
  CONFIRMED: 'مثبت',
  COMPLETED: 'منجز',
  CANCELLED: 'ملغى',
}

const statusColors: Record<string, string> = {
  PENDING: 'bg-amber-100 text-amber-700',
  CONFIRMED: 'bg-blue-100 text-blue-700',
  COMPLETED: 'bg-emerald-100 text-emerald-700',
  CANCELLED: 'bg-red-100 text-red-700',
}

export default function BookingsList() {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')

  useEffect(() => {
    api
      .getBookings()
      .then(setBookings)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

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

      {!loading && !error && topServices.length > 0 && (
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
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filtered.map((b) => (
                <tr key={b.id}>
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
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={8} className="px-4 py-6 text-center text-slate-400">
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
