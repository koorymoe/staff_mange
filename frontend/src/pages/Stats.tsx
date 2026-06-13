import { useEffect, useState } from 'react'
import { api, type Stats } from '../api'

export default function StatsPage() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api
      .getStats()
      .then(setStats)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">الإحصائيات العامة</h2>
      <p className="mt-1 text-slate-500">نظرة شاملة على أداء الشركة والموظفين.</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>
      )}

      {stats && (
        <div className="mt-6 flex flex-col gap-6">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
            {[
              { label: 'إجمالي الزبائن', value: stats.totalCustomers },
              { label: 'إجمالي الحجوزات', value: stats.totalBookings },
              { label: 'حجوزات مثبتة', value: stats.confirmedBookings },
              { label: 'حجوزات منجزة', value: stats.completedBookings },
              { label: 'إجمالي الإيرادات', value: stats.totalRevenue.toLocaleString() },
            ].map((card) => (
              <div
                key={card.label}
                className="rounded-xl border border-white bg-white p-5 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
              >
                <p className="text-sm text-slate-500">{card.label}</p>
                <p className="mt-2 text-2xl font-extrabold text-brand-800">{card.value}</p>
              </div>
            ))}
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
                إحصائيات موظفي المبيعات
              </h3>
              <table className="w-full text-right">
                <thead>
                  <tr className="bg-slate-50 text-sm text-slate-500">
                    <th className="px-4 py-2">الموظف</th>
                    <th className="px-4 py-2">عدد التحويلات</th>
                    <th className="px-4 py-2">المثبتة</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {stats.salesStats.map((s) => (
                    <tr key={s.employeeId}>
                      <td className="px-4 py-2 font-medium">{s.name}</td>
                      <td className="px-4 py-2">{s.totalTransferred}</td>
                      <td className="px-4 py-2">{s.confirmed}</td>
                    </tr>
                  ))}
                  {stats.salesStats.length === 0 && (
                    <tr>
                      <td colSpan={3} className="px-4 py-4 text-center text-slate-400">
                        لا توجد بيانات
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
                إحصائيات الفنيين
              </h3>
              <table className="w-full text-right">
                <thead>
                  <tr className="bg-slate-50 text-sm text-slate-500">
                    <th className="px-4 py-2">الموظف</th>
                    <th className="px-4 py-2">المهام المسندة</th>
                    <th className="px-4 py-2">المنجزة</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {stats.technicianStats.map((s) => (
                    <tr key={s.employeeId}>
                      <td className="px-4 py-2 font-medium">{s.name}</td>
                      <td className="px-4 py-2">{s.totalAssigned}</td>
                      <td className="px-4 py-2">{s.completed}</td>
                    </tr>
                  ))}
                  {stats.technicianStats.length === 0 && (
                    <tr>
                      <td colSpan={3} className="px-4 py-4 text-center text-slate-400">
                        لا توجد بيانات
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
