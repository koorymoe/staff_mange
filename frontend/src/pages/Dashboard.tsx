import { useEffect, useState } from 'react'
import { api, type Stats } from '../api'
import { useSession } from '../session'
import { roleLabels } from '../session'

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

const medals = ['🥇', '🥈', '🥉']

export default function Dashboard() {
  const { employee } = useSession()
  const [stats, setStats] = useState<Stats | null>(null)
  const [error, setError] = useState<string | null>(null)

  const isCommandCenter = employee?.role === 'ADMIN' || employee?.role === 'MONITOR'

  useEffect(() => {
    if (!isCommandCenter) return
    api.getStats().then(setStats).catch((e) => setError(e.message))
  }, [isCommandCenter])

  if (!isCommandCenter) {
    return (
      <div>
        <h2 className="text-2xl font-bold text-brand-900">لوحة التحكم</h2>
        <p className="mt-2 text-slate-500">
          مرحباً بك في نظام إدارة شركة الأماني المتكامل. اختر وحدة من القائمة الجانبية للبدء.
        </p>

        <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="font-bold text-brand-800">إدارة الكوادر</h3>
            <p className="mt-1 text-sm text-slate-500">
              إدارة الموظفين، الشهادات، والمهارات الفنية لكل خدمة.
            </p>
          </div>
          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="font-bold text-brand-800">الزبائن</h3>
            <p className="mt-1 text-sm text-slate-500">
              قاعدة بيانات الزبائن مع كود ثابت لكل زبون.
            </p>
          </div>
          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="font-bold text-brand-800">الخدمات</h3>
            <p className="mt-1 text-sm text-slate-500">
              قائمة الخدمات التي تقدمها الشركة وربطها بمهارات الموظفين.
            </p>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>
  }

  if (!stats) {
    return <p className="mt-6 text-slate-400">جاري تحميل بيانات الشركة...</p>
  }

  const t = stats.totals
  const maxServiceCount = Math.max(1, ...stats.serviceBreakdown.map((s) => s.count))

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-2">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">مركز القيادة - شركة الأماني</h2>
          <p className="mt-1 text-slate-500">نظرة شاملة ومباشرة على أداء كافة الأقسام والكوادر.</p>
        </div>
        {t.urgentPending > 0 && (
          <div className="animate-pulse rounded-full bg-red-100 px-4 py-2 text-sm font-bold text-red-700">
            ⚠ {t.urgentPending} حجز عاجل بانتظار التثبيت
          </div>
        )}
      </div>

      {/* Hero KPI cards */}
      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-xl bg-gradient-to-br from-brand-500 to-brand-800 p-5 text-white shadow-lg shadow-brand-900/20">
          <p className="text-sm text-brand-100">إجمالي الإيرادات المحصلة</p>
          <p className="mt-2 text-3xl font-extrabold">{t.totalRevenue.toLocaleString()}</p>
          {t.unverifiedRevenue > 0 && (
            <p className="mt-1 text-xs text-brand-100">
              منها {t.unverifiedRevenue.toLocaleString()} بانتظار التدقيق
            </p>
          )}
        </div>
        <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-sm text-slate-500">إجمالي الزبائن</p>
          <p className="mt-2 text-3xl font-extrabold text-brand-800">{t.totalCustomers}</p>
        </div>
        <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-sm text-slate-500">إجمالي الحجوزات</p>
          <p className="mt-2 text-3xl font-extrabold text-brand-800">{t.totalBookings}</p>
        </div>
        <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-sm text-slate-500">معدل الإنجاز</p>
          <p className="mt-2 text-3xl font-extrabold text-emerald-600">
            {t.totalBookings ? Math.round((t.completedBookings / t.totalBookings) * 100) : 0}%
          </p>
        </div>
      </div>

      {/* Pipeline funnel */}
      <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          { label: 'بانتظار التثبيت', value: t.pendingBookings, color: 'bg-amber-100 text-amber-700' },
          { label: 'مثبتة وقيد التنفيذ', value: t.confirmedBookings, color: 'bg-blue-100 text-blue-700' },
          { label: 'منجزة', value: t.completedBookings, color: 'bg-emerald-100 text-emerald-700' },
          { label: 'ملغاة', value: t.cancelledBookings, color: 'bg-red-100 text-red-700' },
        ].map((s) => (
          <div key={s.label} className={`rounded-xl p-4 text-center ${s.color}`}>
            <p className="text-2xl font-extrabold">{s.value}</p>
            <p className="mt-1 text-xs font-medium">{s.label}</p>
          </div>
        ))}
      </div>

      <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Sales leaderboard */}
        <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
            🏆 ترتيب موظفي المبيعات (حسب الحجوزات المثبتة)
          </h3>
          <div className="divide-y divide-slate-100">
            {stats.salesStats.map((s, i) => (
              <div key={s.employeeId} className="flex items-center justify-between px-4 py-3">
                <div className="flex items-center gap-3">
                  <span className="text-lg">{medals[i] || `#${i + 1}`}</span>
                  <span className="font-medium text-brand-800">{s.name}</span>
                </div>
                <div className="text-sm text-slate-500">
                  <span className="font-bold text-emerald-600">{s.confirmed}</span> مثبت من{' '}
                  {s.totalTransferred}
                </div>
              </div>
            ))}
            {stats.salesStats.length === 0 && (
              <p className="px-4 py-4 text-center text-slate-400">لا توجد بيانات</p>
            )}
          </div>
        </div>

        {/* Technician leaderboard */}
        <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
            🛠️ ترتيب الفنيين (حسب المهام المنجزة)
          </h3>
          <div className="divide-y divide-slate-100">
            {stats.technicianStats.map((s, i) => (
              <div key={s.employeeId} className="flex items-center justify-between px-4 py-3">
                <div className="flex items-center gap-3">
                  <span className="text-lg">{medals[i] || `#${i + 1}`}</span>
                  <div>
                    <span className="font-medium text-brand-800">{s.name}</span>
                    <span
                      className={`mr-2 inline-block h-2 w-2 rounded-full ${
                        s.onDuty ? 'bg-emerald-500' : 'bg-slate-300'
                      }`}
                      title={s.onDuty ? 'بالدوام' : 'خارج الدوام'}
                    />
                  </div>
                </div>
                <div className="text-sm text-slate-500">
                  <span className="font-bold text-emerald-600">{s.completed}</span> منجز من{' '}
                  {s.totalAssigned}
                  {s.revenueHandled > 0 && (
                    <span className="mr-2 text-xs text-slate-400">
                      ({s.revenueHandled.toLocaleString()})
                    </span>
                  )}
                </div>
              </div>
            ))}
            {stats.technicianStats.length === 0 && (
              <p className="px-4 py-4 text-center text-slate-400">لا توجد بيانات</p>
            )}
          </div>
        </div>

        {/* Service breakdown */}
        <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
            📊 الحجوزات حسب الخدمة
          </h3>
          <div className="flex flex-col gap-2 p-4">
            {stats.serviceBreakdown.map((s) => (
              <div key={s.serviceId || 'none'}>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-600">{s.name}</span>
                  <span className="font-bold text-brand-800">{s.count}</span>
                </div>
                <div className="mt-1 h-2 w-full rounded-full bg-slate-100">
                  <div
                    className="h-2 rounded-full bg-gradient-to-l from-brand-400 to-brand-700"
                    style={{ width: `${(s.count / maxServiceCount) * 100}%` }}
                  />
                </div>
              </div>
            ))}
            {stats.serviceBreakdown.length === 0 && (
              <p className="text-center text-slate-400">لا توجد بيانات</p>
            )}
          </div>
        </div>

        {/* Recent activity feed */}
        <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
            🕒 آخر الحجوزات
          </h3>
          <div className="divide-y divide-slate-100">
            {stats.recentBookings.map((b) => (
              <div key={b.id} className="flex items-center justify-between px-4 py-3 text-sm">
                <div>
                  <span className="font-mono font-semibold text-brand-600">{b.code}</span>
                  <span className="mr-2 text-slate-600">{b.customerName}</span>
                  {b.serviceName && <span className="mr-2 text-xs text-slate-400">{b.serviceName}</span>}
                  {b.priority === 'URGENT' && (
                    <span className="mr-2 rounded-full bg-red-100 px-2 py-0.5 text-xs font-bold text-red-700">
                      عاجل
                    </span>
                  )}
                </div>
                <span className={`rounded-full px-2 py-1 text-xs font-bold ${statusColors[b.status]}`}>
                  {statusLabels[b.status] || b.status}
                </span>
              </div>
            ))}
            {stats.recentBookings.length === 0 && (
              <p className="px-4 py-4 text-center text-slate-400">لا توجد حجوزات بعد</p>
            )}
          </div>
        </div>
      </div>

      {/* Team headcount */}
      <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <h3 className="bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 font-bold text-white">
          👥 توزيع الكادر حسب الصلاحية
        </h3>
        <div className="grid grid-cols-2 gap-3 p-4 sm:grid-cols-4 lg:grid-cols-7">
          {stats.roleCounts.map((r) => (
            <div key={r.role} className="rounded-lg bg-slate-50 p-3 text-center">
              <p className="text-2xl font-extrabold text-brand-800">{r.count}</p>
              <p className="mt-1 text-xs text-slate-500">{roleLabels[r.role] || r.role}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
