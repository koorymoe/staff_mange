import { useState, useEffect } from 'react'
import { api } from '../../api'

export default function GpsDashboard() {
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.getGpsStats()
      .then(setStats)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const cards = stats ? [
    { label: 'إجمالي الزبائن', value: stats.totalCustomers ?? 0, color: 'from-blue-500 to-blue-700' },
    { label: 'إجمالي الأجهزة', value: stats.totalDevices ?? 0, color: 'from-emerald-500 to-emerald-700' },
    { label: 'الاشتراكات الفعالة', value: stats.activeSubscriptions ?? 0, color: 'from-purple-500 to-purple-700' },
    { label: 'الطلبات المعلقة', value: stats.pendingRequests ?? 0, color: 'from-yellow-500 to-yellow-700' },
    { label: 'طلبات الصيانة المعلقة', value: stats.pendingMaintenance ?? 0, color: 'from-red-500 to-red-700' },
  ] : []

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">لوحة تحكم GPS</h2>
      <p className="mt-1 text-slate-500">نظرة عامة على نظام تتبع المركبات</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
          {cards.map((card) => (
            <div key={card.label} className="overflow-hidden rounded-xl bg-white shadow-md">
              <div className={`bg-gradient-to-l ${card.color} px-4 py-3`}>
                <p className="text-sm font-semibold text-white">{card.label}</p>
              </div>
              <div className="px-4 py-6 text-center">
                <p className="text-3xl font-extrabold text-slate-800">{card.value}</p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
