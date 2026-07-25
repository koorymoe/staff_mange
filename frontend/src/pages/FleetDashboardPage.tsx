import { useEffect, useState, useCallback } from 'react'
import { api, type FleetDashboardSummary } from '../api'

export default function FleetDashboardPage() {
  const [data, setData] = useState<FleetDashboardSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    api.getFleetDashboard()
      .then(d => setData(d))
      .catch(() => setError('تعذر جلب بيانات لوحة تحكم الأسطول'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(load, [load])

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">لوحة التحكم الشاملة للأسطول</h2>
          <p className="mt-1 text-slate-500">
            نظرة سريعة على حالة السيارات والمصاريف {data ? `لشهر ${data.period}` : ''}
          </p>
        </div>
        <button onClick={load} className="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-50">
          تحديث البيانات
        </button>
      </div>

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && error && <p className="text-red-500">{error}</p>}

      {!loading && !error && data && (
        <div className="space-y-6">
          {/* Stat Cards */}
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
            {[
              { label: 'سيارات عاملة', value: data.activeVehiclesCount, color: 'text-emerald-600', bg: 'bg-emerald-50' },
              { label: 'سيارات بالصيانة', value: data.inMaintenanceCount, color: 'text-red-600', bg: 'bg-red-50' },
              { label: 'بمهمة حالياً', value: data.onMissionCount, color: 'text-violet-600', bg: 'bg-violet-50' },
              { label: 'تحتاج صيانة/تنظيف', value: data.needsServiceCount, color: 'text-amber-600', bg: 'bg-amber-50' },
              { label: 'وثائق منتهية قريباً', value: data.expiringDocsCount, color: 'text-blue-600', bg: 'bg-blue-50' },
            ].map(c => (
              <div key={c.label} className={`rounded-2xl ${c.bg} p-4 text-center`}>
                <p className={`text-3xl font-black ${c.color}`}>{c.value}</p>
                <p className="mt-1 text-xs font-medium text-slate-500">{c.label}</p>
              </div>
            ))}
          </div>

          {/* Fuel / Fleet Cost Summary */}
          <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
            <h3 className="mb-4 text-lg font-bold text-brand-900">استهلاك الوقود والمصاريف — الشهر الحالي</h3>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-2">
              <div className="rounded-xl bg-slate-50 p-4 text-center">
                <p className="text-2xl font-black text-brand-700">{data.fleetFuelCostThisMonth.toLocaleString('ar-IQ')}</p>
                <p className="mt-1 text-xs font-medium text-slate-500">إجمالي تكلفة الوقود (دينار)</p>
              </div>
              <div className="rounded-xl bg-slate-50 p-4 text-center">
                <p className="text-2xl font-black text-brand-700">{data.fleetTotalCostThisMonth.toLocaleString('ar-IQ')}</p>
                <p className="mt-1 text-xs font-medium text-slate-500">إجمالي مصاريف الأسطول (دينار)</p>
              </div>
            </div>

            <h4 className="mb-2 mt-5 text-sm font-bold text-slate-700">مصاريف كل سيارة هذا الشهر</h4>
            {data.vehicleExpenses.length === 0 && <p className="text-sm text-slate-400">لا توجد مصاريف مسجلة هذا الشهر.</p>}
            {data.vehicleExpenses.length > 0 && (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-slate-100 text-slate-500">
                      <th className="py-2 text-right font-medium">السيارة</th>
                      <th className="py-2 text-right font-medium">رقم اللوحة</th>
                      <th className="py-2 text-right font-medium">تكلفة الوقود</th>
                      <th className="py-2 text-right font-medium">إجمالي التكلفة</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.vehicleExpenses.map(v => (
                      <tr key={v.vehicleId} className="border-b border-slate-50">
                        <td className="py-2">{v.vehicleName}</td>
                        <td className="py-2">{v.plateNumber}</td>
                        <td className="py-2">{v.fuelCost.toLocaleString('ar-IQ')}</td>
                        <td className="py-2 font-bold">{v.totalCost.toLocaleString('ar-IQ')}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* Top Usage / Cost */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
              <h3 className="mb-3 text-lg font-bold text-brand-900">الأكثر استخداماً (مهام/مسافة)</h3>
              {data.topByUsage.length === 0 && <p className="text-sm text-slate-400">لا توجد بيانات كافية.</p>}
              <ol className="space-y-2">
                {data.topByUsage.map((v, i) => (
                  <li key={v.vehicleId} className="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 text-sm">
                    <span className="font-bold text-slate-700">{i + 1}. {v.vehicleName} ({v.plateNumber})</span>
                    <span className="text-slate-500">{v.missionCount} مهمة · {v.distanceKm} كم</span>
                  </li>
                ))}
              </ol>
            </div>
            <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
              <h3 className="mb-3 text-lg font-bold text-brand-900">الأكثر تكلفة</h3>
              {data.topByCost.length === 0 && <p className="text-sm text-slate-400">لا توجد بيانات كافية.</p>}
              <ol className="space-y-2">
                {data.topByCost.map((v, i) => (
                  <li key={v.vehicleId} className="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 text-sm">
                    <span className="font-bold text-slate-700">{i + 1}. {v.vehicleName} ({v.plateNumber})</span>
                    <span className="text-slate-500">{v.totalCost.toLocaleString('ar-IQ')} دينار</span>
                  </li>
                ))}
              </ol>
            </div>
          </div>

          {/* Alerts */}
          <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
            <h3 className="mb-3 text-lg font-bold text-brand-900">تنبيهات الأسطول</h3>
            {data.alerts.length === 0 && <p className="text-sm text-slate-400">لا توجد تنبيهات حالياً.</p>}
            <ul className="space-y-2">
              {data.alerts.map((a, i) => (
                <li
                  key={i}
                  className={`rounded-lg px-3 py-2 text-sm ${a.severity === 'danger' ? 'bg-red-50 text-red-700' : 'bg-amber-50 text-amber-700'}`}
                >
                  <span className="font-bold">{a.vehicleName}:</span> {a.message}
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}
    </div>
  )
}
